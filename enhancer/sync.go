package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (app *App) modifyUpstreamResponse(response *http.Response) error {
	if response.StatusCode != http.StatusOK || response.Request == nil || response.Request.URL.Path != "/api/sync" {
		return nil
	}
	secret := strings.TrimSpace(strings.TrimPrefix(response.Request.Header.Get("Authorization"), "Bearer "))
	if secret == "" {
		secret = strings.TrimSpace(response.Request.URL.Query().Get("secret"))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}
	_ = response.Body.Close()
	response.Body = io.NopCloser(bytes.NewReader(body))

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	role, _ := payload["role"].(string)
	if role == "proxy" {
		return nil
	}
	if role != "dns" {
		return nil
	}
	record, ok := app.ipStore.GetByNodeSecret(secret)
	if !ok || len(record.Routes) == 0 {
		return nil
	}
	rules := objectMap(payload["rules"])
	overrides := objectMap(payload["rule_overrides"])
	services := app.catalog.Snapshot(response.Request.Context(), false).Services
	serviceByID := make(map[string]Service, len(services))
	managedDomains := make(map[string]struct{})
	for _, service := range services {
		serviceByID[service.ID] = service
		for _, domain := range routingDomains(service.Domains) {
			managedDomains[domain] = struct{}{}
		}
	}
	for key, value := range rules {
		rule, _ := value.(map[string]any)
		name, _ := rule["name"].(string)
		if strings.Contains(key, "/enhancer/rules/") || strings.HasPrefix(name, "enhancer:") || strings.Contains(name, "/enhancer/rules/") {
			delete(rules, key)
		}
	}
	for domain := range managedDomains {
		delete(overrides, domain)
	}
	payload["smart"] = false

	routes, _ := normalizeConflictingRoutes(nil, record.Routes, services)
	proxyPeers := app.effectiveProxyPeers(record)
	serviceIDs := make([]string, 0, len(routes))
	for serviceID := range routes {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)
	for _, serviceID := range serviceIDs {
		service, exists := serviceByID[serviceID]
		if !exists {
			continue
		}
		proxyID := routes[serviceID]
		peers := proxyPeers[proxyID]
		if len(peers) == 0 {
			continue
		}
		metaKey := "__meta__:enhancer:" + service.ID + ":" + proxyID
		rules[metaKey] = map[string]any{
			"pattern": metaKey, "ips": peers, "strategy": "", "name": "", "type": "meta",
			"priority": 100, "check": false, "node_id": proxyID,
		}
		for _, domain := range routingDomains(service.Domains) {
			name := "enhancer:" + service.ID
			rules[name+":"+domain] = map[string]any{
				"pattern": domain, "ips": peers, "strategy": "", "name": name, "type": "DOMAIN-SUFFIX",
				"priority": 100, "check": false, "node_id": "",
			}
			delete(overrides, domain)
		}
	}
	payload["rules"] = rules
	payload["rule_overrides"] = overrides
	return replaceSyncResponse(response, payload)
}

func replaceSyncResponse(response *http.Response, payload map[string]any) error {
	modified, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	response.Body = io.NopCloser(bytes.NewReader(modified))
	response.ContentLength = int64(len(modified))
	response.Header.Set("Content-Length", strconv.Itoa(len(modified)))
	response.Header.Del("Content-Encoding")
	response.Header.Del("ETag")
	return nil
}

func objectMap(value any) map[string]any {
	if result, ok := value.(map[string]any); ok {
		return result
	}
	return make(map[string]any)
}

func (app *App) effectiveProxyPeers(record ipConfigRecord) map[string][]string {
	result := effectiveProxyPeers(record)
	for proxyID, peers := range result {
		result[proxyID] = ipv4Peers(peers)
		if tunnelIP, ready := app.transport.EffectiveProxyIP(record.ID, proxyID); ready {
			result[proxyID] = []string{tunnelIP}
		}
	}
	return result
}

func effectiveProxyPeers(record ipConfigRecord) map[string][]string {
	result := cloneProxyPeers(record.ProxyPeers)
	if len(result) > 0 {
		return result
	}
	proxyIDs := make(map[string]struct{})
	for _, proxyID := range record.Routes {
		proxyIDs[proxyID] = struct{}{}
	}
	if len(proxyIDs) != 1 {
		return nil
	}
	peers := make([]string, 0, len(record.TrafficPeers))
	for _, peer := range record.TrafficPeers {
		if parsed := net.ParseIP(strings.TrimSpace(peer)); parsed != nil && parsed.To4() != nil {
			peers = append(peers, parsed.String())
		}
	}
	if len(peers) == 0 {
		return nil
	}
	for proxyID := range proxyIDs {
		return map[string][]string{proxyID: peers}
	}
	return nil
}

func ipv4Peers(peers []string) []string {
	result := make([]string, 0, len(peers))
	for _, peer := range peers {
		parsed := net.ParseIP(strings.TrimSpace(peer))
		if parsed == nil || parsed.To4() == nil {
			continue
		}
		result = append(result, parsed.String())
	}
	sort.Strings(result)
	return result
}
