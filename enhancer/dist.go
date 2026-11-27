package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var allowedDistFiles = map[string]string{
	"dns.sh":                        "text/x-shellscript; charset=utf-8",
	"bootstrap.sh":                  "text/x-shellscript; charset=utf-8",
	"enhanced_install.sh":           "text/x-shellscript; charset=utf-8",
	"agent_install.sh":              "text/x-shellscript; charset=utf-8",
	"prism_transport.sh":            "text/x-shellscript; charset=utf-8",
	"prism_media.sh":                "text/x-shellscript; charset=utf-8",
	"prismdns.sh":                   "text/x-shellscript; charset=utf-8",
	"prism_node_update.sh":          "text/x-shellscript; charset=utf-8",
	"uninstall.sh":                  "text/x-shellscript; charset=utf-8",
	"prism-egress-proxy_linux_amd64": "application/octet-stream",
	"prism-egress-proxy_linux_arm64": "application/octet-stream",
	"prism-agent_linux_amd64":       "application/octet-stream",
	"prism-agent_linux_arm64":       "application/octet-stream",
}

func (app *App) localDistFilePath(name string) (string, string, bool) {
	if app == nil || strings.TrimSpace(app.distDir) == "" {
		return "", "", false
	}
	contentType, allowed := allowedDistFiles[name]
	if !allowed {
		return "", "", false
	}
	cleanDir := filepath.Clean(app.distDir)
	fullPath := filepath.Join(cleanDir, name)
	info, err := os.Lstat(fullPath)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 {
		return "", "", false
	}
	return fullPath, contentType, true
}

func (app *App) hasSelfHostedRelease() bool {
	for _, required := range []string{"prism_transport.sh", "prism_media.sh", "prism_node_update.sh"} {
		if _, _, ok := app.localDistFilePath(required); !ok {
			return false
		}
	}
	return true
}

func (app *App) distOrRawURL(request *http.Request, file string) string {
	if _, _, ok := app.localDistFilePath(file); ok && request != nil {
		return strings.TrimRight(publicBaseURL(request), "/") + "/enhancer/dist/" + file
	}
	return prismRawURL(file)
}

func (app *App) handleDist(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		methodNotAllowed(writer, http.MethodGet, http.MethodHead)
		return
	}
	name := strings.TrimPrefix(request.URL.Path, "/enhancer/dist/")
	if name == "manifest.json" {
		app.handleDistManifest(writer, request)
		return
	}
	fullPath, contentType, ok := app.localDistFilePath(name)
	if !ok {
		http.NotFound(writer, request)
		return
	}
	data, err := os.ReadFile(fullPath)
	if err != nil || len(data) == 0 {
		http.NotFound(writer, request)
		return
	}
	sum := sha256.Sum256(data)
	setNoCacheHeaders(writer)
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("X-Prism-Version", prismSourceRef)
	writer.Header().Set("X-Content-SHA256", hex.EncodeToString(sum[:]))
	writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	writer.WriteHeader(http.StatusOK)
	if request.Method == http.MethodGet {
		_, _ = writer.Write(data)
	}
}

func (app *App) handleDistManifest(writer http.ResponseWriter, request *http.Request) {
	if !app.hasSelfHostedRelease() {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "self-hosted dist is not configured"})
		return
	}
	type distFileMeta struct {
		Size   int    `json:"size"`
		SHA256 string `json:"sha256"`
	}
	files := make(map[string]distFileMeta)
	names := make([]string, 0, len(allowedDistFiles))
	for name := range allowedDistFiles {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fullPath, _, ok := app.localDistFilePath(name)
		if !ok {
			continue
		}
		data, err := os.ReadFile(fullPath)
		if err != nil || len(data) == 0 {
			continue
		}
		sum := sha256.Sum256(data)
		files[name] = distFileMeta{Size: len(data), SHA256: hex.EncodeToString(sum[:])}
	}
	setNoCacheHeaders(writer)
	writeJSON(writer, http.StatusOK, map[string]any{
		"tag_name":    prismSourceRef,
		"version":     prismSourceRef,
		"immutable":   true,
		"draft":       false,
		"prerelease":  false,
		"self_hosted": true,
		"files":       files,
	})
}

func (app *App) handleBatchNodeUpdates(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	authorization := request.Header.Get("Authorization")
	if !app.authorize(request.Context(), authorization) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	if !app.nodeReleaseAvailable(request.Context()) {
		writeJSON(writer, http.StatusServiceUnavailable, map[string]string{"error": "可用更新版本暂未就绪"})
		return
	}
	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": "无法读取节点列表"})
		return
	}
	knownIDs := make(map[string]string, len(nodes))
	for _, node := range nodes {
		app.nodeLabels.Overlay(node)
		id := valueString(node["id"])
		if id != "" {
			knownIDs[id] = valueString(node["name"])
		}
	}

	store := app.nodeUpdates
	store.mu.Lock()
	defer store.mu.Unlock()
	now := time.Now().UTC()
	var queued []string
	var skipped []string

	ids := make([]string, 0, len(knownIDs))
	for id := range knownIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		info, exists := store.inventory[id]
		if !exists || now.Sub(info.LastSeen) >= nodeHeartbeatTTL || !nodeVersionBefore(info.Version, prismSourceRef) {
			skipped = append(skipped, id)
			continue
		}
		if old, ok := store.jobs[id]; ok && old.Binding == info.binding && old.Version == prismSourceRef {
			if current := old.current(now); current.Status == "queued" || current.Status == "running" {
				queued = append(queued, id)
				continue
			}
		}
		jobID, err := randomToken()
		if err != nil {
			continue
		}
		job := nodeUpdate{ID: jobID, Version: prismSourceRef, Status: "queued", RequestedAt: now, UpdatedAt: now}
		if err := store.save(id, savedNodeUpdate{nodeUpdate: job, Binding: info.binding}); err == nil {
			queued = append(queued, id)
		}
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{
		"version": prismSourceRef,
		"queued":  queued,
		"skipped": skipped,
	})
}
