package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (app *App) handleV2Dashboard(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	auth := request.Header.Get("Authorization")
	if !app.authorize(request.Context(), auth) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}

	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), auth, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	proxiesTotal, proxiesOnline, clientsTotal, warpReadyCount, warpConfiguredCount, outdatedCount := 0, 0, 0, 0, 0, 0
	latestVer := "v" + uiVersion
	for _, node := range nodes {
		app.nodeLabels.Overlay(node)
		app.nodeUpdates.overlay(node)
		role := valueString(node["role"])
		online := nodeRecentlyOnline(valueString(node["last_heartbeat"]))
		if role == "proxy" {
			proxiesTotal++
			if online {
				proxiesOnline++
			}
			nodeID := valueString(node["id"])
			if app.transport != nil && app.transport.ProxyWarpReady(nodeID) {
				warpReadyCount++
			}
			if app.warpPolicies != nil && len(app.warpPolicies.ForProxy(nodeID)) > 0 {
				warpConfiguredCount++
			}
		} else if role == "dns" {
			clientsTotal++
		}
		if mgmt, ok := node["management"].(map[string]any); ok {
			ver := strings.TrimSpace(valueString(mgmt["version"]))
			if ver != "" && ver != latestVer {
				outdatedCount++
			}
		}
	}

	var totalRX, totalTX uint64
	ipCount := 0
	if app.ipStore != nil {
		recs := app.ipStore.Records()
		ipCount = len(recs)
		for _, rec := range recs {
			totalRX += rec.TrafficRXBytes
			totalTX += rec.TrafficTXBytes
		}
	}

	catalogSnap := app.catalog.Snapshot(request.Context(), false)
	writeJSON(writer, http.StatusOK, map[string]any{
		"ui_version":            uiVersion,
		"latest_version":        latestVer,
		"proxies_total":         proxiesTotal,
		"proxies_online":        proxiesOnline,
		"clients_total":         clientsTotal,
		"ip_configs_total":      ipCount,
		"warp_ready_count":      warpReadyCount,
		"warp_configured_count": warpConfiguredCount,
		"outdated_nodes_count":  outdatedCount,
		"catalog_total":         len(catalogSnap.Services),
		"traffic_rx_bytes":      totalRX,
		"traffic_tx_bytes":      totalTX,
		"server_time":           time.Now().UTC(),
	})
}

func (app *App) handleV2Events(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	auth := strings.TrimSpace(request.Header.Get("Authorization"))
	if auth == "" {
		if tok := strings.TrimSpace(request.URL.Query().Get("token")); tok != "" {
			auth = "Bearer " + tok
		}
	}
	if !app.authorize(request.Context(), auth) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}

	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no")
	writer.WriteHeader(http.StatusOK)

	buildFingerprint := func() (string, map[string]any) {
		state := map[string]any{
			"ui_version": uiVersion,
			"ts":         time.Now().UTC().Unix(),
		}
		if app.transport != nil {
			state["warp_ready"] = app.transport.WarpReadyMap()
		}
		if app.warpPolicies != nil {
			state["warp_policies"] = app.warpPolicies.Snapshot()
		}
		if app.ipStore != nil {
			state["ip_count"] = len(app.ipStore.Records())
		}
		raw, _ := json.Marshal(map[string]any{
			"warp_ready":    state["warp_ready"],
			"warp_policies": state["warp_policies"],
			"ip_count":      state["ip_count"],
		})
		sum := sha256.Sum256(raw)
		return hex.EncodeToString(sum[:8]), state
	}

	lastHash, payload := buildFingerprint()
	data, _ := json.Marshal(payload)
	_, _ = fmt.Fprintf(writer, "event: sync\ndata: %s\n\n", data)
	flusher.Flush()

	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-request.Context().Done():
			return
		case <-ticker.C:
			hash, nextPayload := buildFingerprint()
			if hash != lastHash {
				lastHash = hash
				raw, _ := json.Marshal(nextPayload)
				_, _ = fmt.Fprintf(writer, "event: sync\ndata: %s\n\n", raw)
			} else {
				_, _ = fmt.Fprintf(writer, ": keepalive %d\n\n", time.Now().Unix())
			}
			flusher.Flush()
		}
	}
}
