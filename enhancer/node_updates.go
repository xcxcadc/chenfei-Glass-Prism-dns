package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const nodeHeartbeatTTL = 90 * time.Second
const nodeUpdateTTL = 15 * time.Minute

var nodeVersionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]{0,5})\.(0|[1-9][0-9]{0,5})\.(0|[1-9][0-9]{0,5})$`)
var agentVersionPattern = regexp.MustCompile(`^v?[0-9]{1,6}\.[0-9]{1,6}(\.[0-9]{1,6})?$`)

type nodeUpdate struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requested_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type nodeManagement struct {
	Version          string      `json:"version"`
	TransportVersion string      `json:"transport_version,omitempty"`
	AgentVersion     string      `json:"agent_version,omitempty"`
	IPv4             string      `json:"ipv4,omitempty"`
	LastSeen         time.Time   `json:"last_seen"`
	UpdateSupported  bool        `json:"update_supported"`
	Update           *nodeUpdate `json:"update,omitempty"`
	binding          [32]byte
}

type savedNodeUpdate struct {
	nodeUpdate
	Binding [32]byte `json:"binding"`
}

type nodeUpdateStore struct {
	mu           sync.Mutex
	path         string
	inventory    map[string]nodeManagement
	jobs         map[string]savedNodeUpdate
	releaseOK    bool
	releaseUntil time.Time
}

func newNodeUpdateStore(path string) (*nodeUpdateStore, error) {
	store := &nodeUpdateStore{path: path, inventory: make(map[string]nodeManagement), jobs: make(map[string]savedNodeUpdate)}
	if path == "" {
		return store, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &store.jobs); err != nil {
		return nil, err
	}
	if store.jobs == nil {
		store.jobs = make(map[string]savedNodeUpdate)
	}
	return store, nil
}

func (store *nodeUpdateStore) save(id string, job savedNodeUpdate) error {
	previous, existed := store.jobs[id]
	store.jobs[id] = job
	if store.path == "" {
		return nil
	}
	data, err := json.Marshal(store.jobs)
	if err == nil {
		err = os.WriteFile(store.path+".tmp", data, 0o600)
	}
	if err == nil {
		err = os.Rename(store.path+".tmp", store.path)
	}
	if err != nil {
		if existed {
			store.jobs[id] = previous
		} else {
			delete(store.jobs, id)
		}
	}
	return err
}

func nodeVersionBefore(current, target string) bool {
	if !nodeVersionPattern.MatchString(current) || !nodeVersionPattern.MatchString(target) {
		return false
	}
	left := strings.Split(strings.TrimPrefix(current, "v"), ".")
	right := strings.Split(strings.TrimPrefix(target, "v"), ".")
	for index, part := range left {
		from, _ := strconv.Atoi(part)
		to, _ := strconv.Atoi(right[index])
		if from != to {
			return from < to
		}
	}
	return false
}

func (job savedNodeUpdate) current(now time.Time) nodeUpdate {
	result := job.nodeUpdate
	if (result.Status == "queued" || result.Status == "running") && now.Sub(result.UpdatedAt) >= nodeUpdateTTL {
		result.Status = "failed"
	}
	return result
}

func (store *nodeUpdateStore) overlay(node map[string]any) {
	store.mu.Lock()
	defer store.mu.Unlock()
	id := valueString(node["id"])
	info, exists := store.inventory[id]
	if !exists {
		return
	}
	now := time.Now().UTC()
	info.UpdateSupported = now.Sub(info.LastSeen) < nodeHeartbeatTTL
	if job, ok := store.jobs[id]; ok && job.Binding == info.binding {
		result := job.current(now)
		info.Update = &result
	}
	node["management"] = info
}

func (app *App) handleNodeRelease(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	if !app.nodeReleaseAvailable(request.Context()) {
		writeJSON(writer, http.StatusServiceUnavailable, map[string]string{"error": "不可变正式版本暂不可用"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"version": prismSourceRef})
}

func (app *App) nodeReleaseAvailable(ctx context.Context) bool {
	store := app.nodeUpdates
	store.mu.Lock()
	if time.Now().Before(store.releaseUntil) {
		available := store.releaseOK
		store.mu.Unlock()
		return available
	}
	store.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/xcxcadc/chenfei-Glass-Prism-dns/releases/tags/"+prismSourceRef, nil)
	if err != nil {
		return false
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	var available bool
	if response, err := app.client.Do(request); err == nil {
		defer response.Body.Close()
		if response.StatusCode == http.StatusOK {
			var release struct {
				TagName    string `json:"tag_name"`
				Immutable  bool   `json:"immutable"`
				Draft      *bool  `json:"draft"`
				Prerelease *bool  `json:"prerelease"`
			}
			available = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&release) == nil &&
				release.TagName == prismSourceRef && release.Immutable && release.Draft != nil && !*release.Draft && release.Prerelease != nil && !*release.Prerelease
		}
	}
	if ctx.Err() != nil {
		return false
	}
	store.mu.Lock()
	store.releaseOK = available
	ttl := 30 * time.Second
	if available {
		ttl = 5 * time.Minute
	}
	store.releaseUntil = time.Now().Add(ttl)
	store.mu.Unlock()
	return available
}

func (app *App) handleNodeUpdates(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	authorization := request.Header.Get("Authorization")
	if !app.authorize(request.Context(), authorization) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
		return
	}
	var payload struct {
		NodeID  string `json:"node_id"`
		Version string `json:"version"`
	}
	if err := decodeJSON(request, &payload); err != nil || payload.NodeID == "" || payload.Version != prismSourceRef {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid node or release version"})
		return
	}
	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": "could not verify node"})
		return
	}
	for _, node := range nodes {
		if valueString(node["id"]) == payload.NodeID {
			if !app.nodeReleaseAvailable(request.Context()) {
				writeJSON(writer, http.StatusServiceUnavailable, map[string]string{"error": "不可变正式版本暂不可用"})
				return
			}
			app.queueNodeUpdate(writer, payload.NodeID, payload.Version)
			return
		}
	}
	writeJSON(writer, http.StatusNotFound, map[string]string{"error": "node not found"})
}

func (app *App) queueNodeUpdate(writer http.ResponseWriter, id, version string) {
	store := app.nodeUpdates
	store.mu.Lock()
	defer store.mu.Unlock()
	now := time.Now().UTC()
	info, exists := store.inventory[id]
	if old, ok := store.jobs[id]; ok && exists && old.Binding == info.binding && old.Version == version {
		if job := old.current(now); job.Status == "queued" || job.Status == "running" {
			writeJSON(writer, http.StatusAccepted, job)
			return
		}
	}
	if !exists || now.Sub(info.LastSeen) >= nodeHeartbeatTTL || !nodeVersionBefore(info.Version, version) {
		writeJSON(writer, http.StatusConflict, map[string]string{"error": "node update worker is offline, unavailable, or already current"})
		return
	}
	jobID, err := randomToken()
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "could not create update"})
		return
	}
	job := nodeUpdate{ID: jobID, Version: version, Status: "queued", RequestedAt: now, UpdatedAt: now}
	if err := store.save(id, savedNodeUpdate{nodeUpdate: job, Binding: info.binding}); err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "could not save update"})
		return
	}
	writeJSON(writer, http.StatusAccepted, job)
}

func (app *App) nodeWorkerID(role, credential string) string {
	if credential == "" || len(credential) > 256 {
		return ""
	}
	switch role {
	case "client":
		if record, ok := app.ipStore.GetByToken(credential); ok {
			var controllerRole string
			if app.controllerNodeRole(record.DNSNodeID, &controllerRole) == nil && controllerRole == "dns" {
				return record.DNSNodeID
			}
		}
	case "proxy":
		if node, err := controllerNodeBySecret(app.controllerDB, credential); err == nil && node.Role == "proxy" {
			return node.ID
		}
	}
	return ""
}

type nodeWorkerPayload struct {
	Role             string `json:"role"`
	Credential       string `json:"credential"`
	Action           string `json:"action"`
	ID               string `json:"id"`
	Version          string `json:"version"`
	TransportVersion string `json:"transport_version"`
	AgentVersion     string `json:"agent_version"`
	IPv4             string `json:"ipv4"`
	Status           string `json:"status"`
}

func (app *App) handleNodeUpdateWorker(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload nodeWorkerPayload
	if err := decodeJSON(request, &payload); err != nil || !nodeVersionPattern.MatchString(payload.Version) {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid worker payload"})
		return
	}
	id := app.nodeWorkerID(payload.Role, payload.Credential)
	if id == "" {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid node credential"})
		return
	}
	binding := sha256.Sum256([]byte(payload.Role + "\x00" + payload.Credential))
	writer.Header().Set("Cache-Control", "no-store")
	store := app.nodeUpdates
	store.mu.Lock()
	defer store.mu.Unlock()
	now := time.Now().UTC()
	job, exists := store.jobs[id]
	if payload.Action == "poll" {
		info := nodeManagement{Version: "v" + strings.TrimPrefix(payload.Version, "v"), LastSeen: now, binding: binding}
		if nodeVersionPattern.MatchString(payload.TransportVersion) {
			info.TransportVersion = payload.TransportVersion
		}
		if agentVersionPattern.MatchString(payload.AgentVersion) {
			info.AgentVersion = payload.AgentVersion
		}
		if address, err := netip.ParseAddr(payload.IPv4); err == nil && address.Is4() {
			info.IPv4 = address.String()
		}
		store.inventory[id] = info
		var assigned *nodeUpdate
		status := job.current(now).Status
		if exists && job.Binding == binding && (status == "queued" || status == "running") && job.Version == prismSourceRef && nodeVersionBefore(info.Version, job.Version) {
			if status == "queued" {
				job.Status, job.UpdatedAt = "running", now
				if err := store.save(id, job); err != nil {
					writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "could not claim update"})
					return
				}
			}
			assigned = &job.nodeUpdate
		}
		writeJSON(writer, http.StatusOK, map[string]any{"job": assigned})
		return
	}
	if payload.Action != "report" || (payload.Status != "succeeded" && payload.Status != "failed") {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid update report"})
		return
	}
	if !exists || job.ID != payload.ID || job.Binding != binding || job.Version != payload.Version || (job.Status != "running" && job.Status != payload.Status) {
		writeJSON(writer, http.StatusConflict, map[string]string{"error": "update job does not match"})
		return
	}
	if job.Status == "running" {
		job.Status, job.UpdatedAt = payload.Status, now
		if err := store.save(id, job); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "could not save update result"})
			return
		}
	}
	if info, ok := store.inventory[id]; ok && info.binding == binding && job.Status == "succeeded" && nodeVersionBefore(info.Version, job.Version) {
		info.Version, info.LastSeen = job.Version, now
		store.inventory[id] = info
	}
	writeJSON(writer, http.StatusOK, job.nodeUpdate)
}
