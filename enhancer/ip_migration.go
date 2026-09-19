package main

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

const ipConfigExportVersion = 1

type ipConfigExport struct {
	Version    int                   `json:"version"`
	ExportedAt time.Time             `json:"exported_at"`
	Configs    []ipConfigExportEntry `json:"configs"`
}

type ipConfigExportEntry struct {
	IP     string                         `json:"ip"`
	Note   string                         `json:"note,omitempty"`
	Smart  bool                           `json:"smart"`
	Routes map[string]ipConfigExportRoute `json:"routes"`
}

type ipConfigExportRoute struct {
	NodeID   string `json:"node_id,omitempty"`
	NodeName string `json:"node_name,omitempty"`
	PublicIP string `json:"public_ip,omitempty"`
	Country  string `json:"country,omitempty"`
}

type ipConfigImportResult struct {
	Imported int        `json:"imported"`
	Updated  int        `json:"updated"`
	Configs  []IPConfig `json:"configs"`
}

func (app *App) handleIPConfigExport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	proxyNodes := make(map[string]map[string]any)
	for _, node := range nodes {
		if valueString(node["role"]) == "proxy" {
			proxyNodes[valueString(node["id"])] = node
		}
	}
	configs := app.ipStore.List()
	entries := make([]ipConfigExportEntry, 0, len(configs))
	for _, config := range configs {
		routes := make(map[string]ipConfigExportRoute, len(config.Routes))
		for serviceID, proxyID := range config.Routes {
			route := ipConfigExportRoute{NodeID: proxyID}
			if node := proxyNodes[proxyID]; node != nil {
				route.NodeName = strings.TrimSpace(valueString(node["name"]))
				route.Country = strings.TrimSpace(valueString(node["country"]))
				peers := nodePublicIPs(node)
				if len(peers) > 0 {
					route.PublicIP = peers[0]
				}
			}
			routes[serviceID] = route
		}
		entries = append(entries, ipConfigExportEntry{IP: config.IP, Note: config.Note, Smart: config.Smart, Routes: routes})
	}
	payload := ipConfigExport{Version: ipConfigExportVersion, ExportedAt: time.Now().UTC(), Configs: entries}
	writer.Header().Set("Content-Disposition", `attachment; filename="prismdns-ip-configs-v1.json"`)
	writeJSON(writer, http.StatusOK, payload)
}

func (app *App) handleIPConfigImport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	var payload ipConfigExport
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if payload.Version != ipConfigExportVersion {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "不支持的 IP 配置备份版本"})
		return
	}
	if len(payload.Configs) == 0 || len(payload.Configs) > 256 {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "备份必须包含 1 到 256 个 IP 配置"})
		return
	}
	for index := range payload.Configs {
		entry := &payload.Configs[index]
		entry.IP = strings.TrimSpace(entry.IP)
		entry.Note = strings.TrimSpace(entry.Note)
		if net.ParseIP(entry.IP) == nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "备份包含无效 IP: " + entry.IP})
			return
		}
		if len(entry.Note) > 80 {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "备注不能超过 80 个字符"})
			return
		}
		if entry.Routes == nil {
			entry.Routes = map[string]ipConfigExportRoute{}
		}
		if len(entry.Routes) > 4096 {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "单个 IP 的服务路由数量过多"})
			return
		}
		for serviceID, route := range entry.Routes {
			if strings.TrimSpace(serviceID) == "" {
				writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "备份包含空服务 ID"})
				return
			}
			if strings.TrimSpace(route.NodeID) == "" && strings.TrimSpace(route.PublicIP) == "" && strings.TrimSpace(route.NodeName) == "" {
				writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "服务路由缺少解锁机标识: " + serviceID})
				return
			}
		}
	}

	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	resolvedRoutes := make([]map[string]string, len(payload.Configs))
	for index, entry := range payload.Configs {
		resolved, err := resolveImportedRoutes(entry.Routes, nodes)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": entry.IP + ": " + err.Error()})
			return
		}
		resolvedRoutes[index] = resolved
	}

	result := ipConfigImportResult{Configs: make([]IPConfig, 0, len(payload.Configs))}
	for index, entry := range payload.Configs {
		existingRecord, exists := app.ipStore.RecordByIP(entry.IP)
		var config IPConfig
		var err error
		if exists {
			existingRecord, err = app.reconcileIPConfigNode(request.Context(), request.Header.Get("Authorization"), existingRecord)
			if err == nil {
				application, applyErr := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), existingRecord.DNSNodeID, existingRecord.Routes, resolvedRoutes[index], publicBaseURL(request))
				if applyErr != nil {
					err = applyErr
				} else {
					existingRecord.Note = entry.Note
					existingRecord.Smart = entry.Smart
					existingRecord.Routes = application.Routes
					existingRecord.TrafficPeers = application.TrafficPeers
					config, err = app.ipStore.Save(existingRecord.IPConfig, existingRecord.NodeSecret, application.ProxyPeers)
				}
			}
		} else {
			payload := ipConfigRequest{IP: entry.IP, Note: entry.Note, Smart: entry.Smart, Routes: resolvedRoutes[index]}
			var dnsNodeID, nodeName, secret string
			var externalNode bool
			dnsNodeID, nodeName, secret, externalNode, err = app.resolveIPConfigNode(request.Context(), request.Header.Get("Authorization"), payload)
			if err == nil {
				application, applyErr := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), dnsNodeID, nil, resolvedRoutes[index], publicBaseURL(request))
				if applyErr != nil {
					err = applyErr
					if !externalNode {
						_ = app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+dnsNodeID, nil, nil)
					}
				} else {
					config, err = app.ipStore.Save(IPConfig{IP: entry.IP, Note: entry.Note, DNSNodeID: dnsNodeID, NodeName: nodeName, ExternalDNSNode: externalNode, Smart: entry.Smart, Routes: application.Routes, TrafficPeers: application.TrafficPeers}, secret, application.ProxyPeers)
				}
			}
		}
		if err != nil {
			writeJSON(writer, http.StatusBadGateway, map[string]any{"error": entry.IP + ": " + err.Error(), "imported": result.Imported, "updated": result.Updated})
			return
		}
		if exists {
			result.Updated++
		} else {
			result.Imported++
		}
		result.Configs = append(result.Configs, config)
	}
	writeJSON(writer, http.StatusOK, result)
}

func resolveImportedRoutes(routes map[string]ipConfigExportRoute, nodes []map[string]any) (map[string]string, error) {
	proxyNodes := make([]map[string]any, 0)
	byID := make(map[string]map[string]any)
	for _, node := range nodes {
		if valueString(node["role"]) != "proxy" {
			continue
		}
		proxyNodes = append(proxyNodes, node)
		byID[valueString(node["id"])] = node
	}
	result := make(map[string]string, len(routes))
	for serviceID, route := range routes {
		node := byID[strings.TrimSpace(route.NodeID)]
		if node == nil && strings.TrimSpace(route.PublicIP) != "" {
			for _, candidate := range proxyNodes {
				if contains(nodePublicIPs(candidate), route.PublicIP) {
					node = candidate
					break
				}
			}
		}
		if node == nil && strings.TrimSpace(route.NodeName) != "" {
			for _, candidate := range proxyNodes {
				if strings.EqualFold(strings.TrimSpace(valueString(candidate["name"])), strings.TrimSpace(route.NodeName)) {
					node = candidate
					break
				}
			}
		}
		if node == nil {
			return nil, errors.New("找不到解锁机，请先添加同一台解锁机或更新备份")
		}
		id := strings.TrimSpace(valueString(node["id"]))
		if id == "" {
			return nil, errors.New("解锁机缺少节点 ID")
		}
		result[serviceID] = id
	}
	return result, nil
}

func (store *IPConfigStore) RecordByIP(ip string) (ipConfigRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, record := range store.configs {
		if strings.EqualFold(strings.TrimSpace(record.IP), strings.TrimSpace(ip)) {
			return cloneIPConfigRecord(record), true
		}
	}
	return ipConfigRecord{}, false
}
