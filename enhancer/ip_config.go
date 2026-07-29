package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

type ipConfigRequest struct {
	IP                string            `json:"ip"`
	Note              string            `json:"note"`
	Smart             bool              `json:"smart"`
	Routes            map[string]string `json:"routes"`
	ExistingDNSNodeID string            `json:"existing_dns_node_id,omitempty"`
}

type trafficReportRequest struct {
	Token          string `json:"token"`
	Scope          string `json:"scope"`
	RXBytes        uint64 `json:"rx_bytes"`
	TXBytes        uint64 `json:"tx_bytes"`
	Interface      string `json:"interface"`
	DNSReady       bool   `json:"dns_ready"`
	SystemDNSReady bool   `json:"system_dns_ready"`
	RoutesReady    bool   `json:"routes_ready"`
	HealthyRoutes  int    `json:"healthy_routes"`
	ExpectedRoutes int    `json:"expected_routes"`
	HealthMessage  string `json:"health_message"`
}

type serviceAuditRequest struct {
	Token   string            `json:"token"`
	Scope   string            `json:"scope"`
	Results map[string]string `json:"results"`
}

type routeApplication struct {
	TrafficPeers []string
	ProxyPeers   map[string][]string
	Routes       map[string]string
}

func (app *App) handleTrafficReport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload trafficReportRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if payload.Token == "" || payload.Scope != "unlock_peers" {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "仅接受解锁链路流量上报"})
		return
	}
	config, err := app.ipStore.UpdateClientReport(payload.Token, payload.RXBytes, payload.TXBytes, ClientHealth{
		DNSReady:       payload.DNSReady,
		SystemDNSReady: payload.SystemDNSReady,
		RoutesReady:    payload.RoutesReady,
		HealthyRoutes:  payload.HealthyRoutes,
		ExpectedRoutes: payload.ExpectedRoutes,
		Message:        payload.HealthMessage,
	})
	if errors.Is(err, os.ErrNotExist) {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "无效的配置令牌"})
		return
	} else if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, config)
}

func (app *App) handleTrafficClear(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		methodNotAllowed(writer, http.MethodDelete)
		return
	}
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	id := strings.TrimPrefix(request.URL.Path, "/enhancer/api/traffic/")
	config, err := app.ipStore.ClearTraffic(id)
	if errors.Is(err, os.ErrNotExist) {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "IP 配置不存在"})
		return
	} else if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, config)
}

func (app *App) handleServiceAuditReport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload serviceAuditRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if payload.Token == "" || payload.Scope != "unlock_services" || len(payload.Results) == 0 {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "无效的服务审计上报"})
		return
	}
	config, err := app.ipStore.UpdateServiceAudit(payload.Token, payload.Results)
	if errors.Is(err, os.ErrNotExist) {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "无效的配置令牌"})
		return
	} else if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, config)
}

func (app *App) handleIPConfigs(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	switch request.Method {
	case http.MethodGet:
		writeJSON(writer, http.StatusOK, app.ipStore.List())
	case http.MethodPost:
		app.createIPConfig(writer, request)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPost)
	}
}

func (app *App) handleIPConfig(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	path := strings.Trim(strings.TrimPrefix(request.URL.Path, "/enhancer/api/ip-configs/"), "/")
	auditRequest := strings.HasSuffix(path, "/audit")
	id := strings.TrimSuffix(path, "/audit")
	record, ok := app.ipStore.Record(id)
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "IP 配置不存在"})
		return
	}
	if auditRequest {
		if request.Method != http.MethodPost {
			methodNotAllowed(writer, http.MethodPost)
			return
		}
		config, err := app.ipStore.RequestServiceAudit(record.ID)
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusAccepted, config)
		return
	}
	switch request.Method {
	case http.MethodGet:
		writeJSON(writer, http.StatusOK, record.public())
	case http.MethodPut:
		app.updateIPConfig(writer, request, record)
	case http.MethodDelete:
		app.deleteIPConfig(writer, request, record)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (app *App) handleBootstrap(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	token := strings.TrimPrefix(request.URL.Path, "/enhancer/api/bootstrap/")
	record, ok := app.ipStore.GetByToken(token)
	if !ok || token == "" {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "无效的配置令牌"})
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writeJSON(writer, http.StatusOK, map[string]any{
		"id":                         record.ID,
		"expected_ip":                record.IP,
		"detected_ip":                clientIP(request),
		"master":                     publicBaseURL(request),
		"secret":                     record.NodeSecret,
		"smart":                      false,
		"dns":                        "127.0.0.1",
		"traffic_peers":              app.effectiveTrafficPeers(record),
		"health_probes":              app.healthProbes(request.Context(), record),
		"service_audit_requested_at": record.ServiceAuditRequestedAt,
		"agent_installer":            "https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh",
		"transport_installer":        "https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/prism_transport.sh",
	})
}

func (app *App) effectiveTrafficPeers(record ipConfigRecord) []string {
	seen := make(map[string]struct{})
	for _, peers := range app.effectiveProxyPeers(record) {
		for _, peer := range peers {
			if net.ParseIP(peer) != nil {
				seen[peer] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for peer := range seen {
		result = append(result, peer)
	}
	sort.Strings(result)
	return result
}

func (app *App) healthProbes(ctx context.Context, record ipConfigRecord) []map[string]any {
	services := make(map[string]Service)
	catalogServices := app.catalog.Snapshot(ctx, false).Services
	for _, service := range catalogServices {
		services[service.ID] = service
	}
	routes, _ := normalizeConflictingRoutes(nil, record.Routes, catalogServices)
	proxyPeers := app.effectiveProxyPeers(record)
	serviceIDs := make([]string, 0, len(routes))
	for serviceID := range routes {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)
	probes := make([]map[string]any, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		service, ok := services[serviceID]
		if !ok || len(service.Domains) == 0 {
			continue
		}
		probes = append(probes, map[string]any{
			"service_id":    service.ID,
			"name":          service.Name,
			"domain":        preferredProbeDomain(service),
			"unlock_test":   unlockTestProvider(service),
			"unlock_tests":  unlockTestProviders(service),
			"probe_domains": preferredProbeDomains(service),
			"route_domains": routingDomains(service.Domains),
			"traffic_peers": append([]string(nil), proxyPeers[routes[serviceID]]...),
		})
	}
	return probes
}

func unlockTestProvider(service Service) string {
	providers := unlockTestProviders(service)
	if len(providers) == 0 {
		return ""
	}
	return providers[0]
}

func unlockTestProviders(service Service) []string {
	providers := map[string][]string{
		"Apple TV+":                       {"Apple"},
		"Bilibili":                        {"Bilibili Anime"},
		"ChatGPT / OpenAI":                {"ChatGPT"},
		"Claude":                          {"Claude"},
		"Crunchyroll":                     {"Crunchyroll"},
		"DAZN":                            {"Dazn"},
		"Disney+":                         {"Disney+"},
		"Gemini":                          {"Gemini"},
		"HBO / Max":                       {"HBO Max"},
		"Microsoft Copilot Image Creator": {"Microsoft Copilot"},
		"Netflix":                         {"Netflix"},
		"Paramount+":                      {"ParamountPlus"},
		"Spotify":                         {"Spotify Registration"},
		"TikTok":                          {"TikTok"},
		"YouTube":                         {"YouTube Region", "YouTube CDN"},
	}
	return append([]string(nil), providers[service.Name]...)
}

func preferredProbeDomain(service Service) string {
	domains := preferredProbeDomains(service)
	if len(domains) > 0 {
		return domains[0]
	}
	return ""
}

func preferredProbeDomains(service Service) []string {
	preferred := map[string][]string{
		"Abema":                           {"abema.tv"},
		"Apple TV+":                       {"tv.apple.com"},
		"Bilibili":                        {"bilibili.com"},
		"ChatGPT / OpenAI":                {"chatgpt.com", "openai.com", "cdn.oaistatic.com", "files.oaiusercontent.com", "sora.com"},
		"Claude":                          {"claude.ai", "claude.com", "anthropic.com"},
		"Crunchyroll":                     {"crunchyroll.com"},
		"DAZN":                            {"dazn.com"},
		"Disney+":                         {"disneyplus.com", "bamgrid.com"},
		"FR:France.tv":                    {"france.tv"},
		"Gemini":                          {"gemini.google.com", "aisandbox-pa.googleapis.com", "alkaliminer-pa.googleapis.com", "proactivebackend-pa.googleapis.com", "robinfrontend-pa.googleapis.com"},
		"Google AI Studio":                {"aistudio.google.com", "alkalicore-pa.clients6.google.com", "alkalimakersuite-pa.clients6.google.com", "generativelanguage.googleapis.com", "waa-pa.clients6.google.com"},
		"HBO / Max":                       {"max.com", "hbomax.com"},
		"HOY TV":                          {"r.hoy.tv"},
		"iQIYI":                           {"iq.com"},
		"J:com On Demand":                 {"www.jcom.co.jp", "linkvod.myjcom.jp"},
		"Karaoke@DAM":                     {"www.clubdam.com"},
		"Microsoft Copilot Image Creator": {"copilot.microsoft.com"},
		"Netflix":                         {"netflix.com"},
		"NZ:Neon TV":                      {"www.neontv.co.nz"},
		"NZ:ThreeNow":                     {"www.threenow.co.nz"},
		"Paramount+":                      {"paramountplus.com"},
		"Spotify":                         {"spotify.com"},
		"Suno":                            {"suno.com", "suno.ai"},
		"TikTok":                          {"tiktok.com"},
		"Videomarket":                     {"www.videomarket.jp"},
		"Viu.TV":                          {"viu.com", "viu.tv"},
		"VN:Galaxy Play":                  {"galaxyplay.vn"},
		"VN:K+":                           {"www.kplus.vn"},
		"Wavve":                           {"www.wavve.com"},
		"YouTube":                         {"youtube.com", "redirector.googlevideo.com"},
	}
	serviceDomains := routingDomains(service.Domains)
	result := make([]string, 0, len(preferred[service.Name]))
	for _, candidate := range preferred[service.Name] {
		for _, domain := range serviceDomains {
			if domain == candidate || strings.HasSuffix(candidate, "."+domain) {
				result = append(result, candidate)
				break
			}
		}
	}
	if len(result) == 0 && len(serviceDomains) > 0 {
		candidates := append([]string(nil), serviceDomains...)
		sort.SliceStable(candidates, func(left, right int) bool {
			leftLabels := strings.Count(candidates[left], ".")
			rightLabels := strings.Count(candidates[right], ".")
			if leftLabels != rightLabels {
				return leftLabels < rightLabels
			}
			if len(candidates[left]) != len(candidates[right]) {
				return len(candidates[left]) < len(candidates[right])
			}
			return candidates[left] < candidates[right]
		})
		for _, candidate := range candidates {
			if !contains(result, candidate) {
				result = append(result, candidate)
			}
			if len(result) == 4 {
				break
			}
		}
	}
	return result
}

func (app *App) createIPConfig(writer http.ResponseWriter, request *http.Request) {
	var payload ipConfigRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	payload.IP = strings.TrimSpace(payload.IP)
	if net.ParseIP(payload.IP) == nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "请输入有效的 IPv4 或 IPv6 地址"})
		return
	}
	for _, existing := range app.ipStore.List() {
		if existing.IP == payload.IP {
			writeJSON(writer, http.StatusConflict, map[string]string{"error": "该 IP 已存在"})
			return
		}
	}
	dnsNodeID, nodeName, secret, externalNode, err := app.resolveIPConfigNode(request.Context(), request.Header.Get("Authorization"), payload)
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	application, err := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), dnsNodeID, nil, payload.Routes, publicBaseURL(request))
	if err != nil {
		if !externalNode {
			_ = app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+url.PathEscape(dnsNodeID), nil, nil)
		}
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	config, err := app.ipStore.Save(IPConfig{IP: payload.IP, Note: payload.Note, DNSNodeID: dnsNodeID, NodeName: nodeName, ExternalDNSNode: externalNode, Smart: payload.Smart, Routes: application.Routes, TrafficPeers: application.TrafficPeers}, secret, application.ProxyPeers)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusCreated, config)
}

func (app *App) resolveIPConfigNode(ctx context.Context, authorization string, payload ipConfigRequest) (string, string, string, bool, error) {
	if payload.ExistingDNSNodeID == "" {
		secret, err := randomToken()
		if err != nil {
			return "", "", "", false, err
		}
		nodeName := safeNodeName(payload.IP)
		nodePayload := map[string]any{
			"name": nodeName, "role": "dns", "public_ip": payload.IP, "country": "", "group": "ip clients", "priority": 1, "secret": secret,
		}
		var created map[string]any
		if err := app.upstreamJSON(ctx, authorization, http.MethodPost, "/api/nodes", nodePayload, &created); err != nil {
			return "", "", "", false, err
		}
		dnsNodeID := valueString(created["id"])
		if returnedSecret := valueString(created["secret"]); returnedSecret != "" {
			secret = returnedSecret
		}
		if dnsNodeID == "" {
			return "", "", "", false, errors.New("Controller did not return a DNS node ID")
		}
		return dnsNodeID, nodeName, secret, false, nil
	}

	var nodes []map[string]any
	if err := app.upstreamJSON(ctx, authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		return "", "", "", false, err
	}
	for _, node := range nodes {
		if valueString(node["id"]) != payload.ExistingDNSNodeID {
			continue
		}
		if valueString(node["role"]) != "dns" {
			return "", "", "", false, errors.New("only DNS client nodes can be adopted")
		}
		nodeIP := strings.TrimSpace(valueString(node["public_ip"]))
		if nodeIP == "" {
			nodeIP = strings.TrimSpace(strings.Split(valueString(node["address"]), ",")[0])
		}
		if nodeIP != payload.IP {
			return "", "", "", false, errors.New("DNS node address does not match the target IP")
		}
		secret := valueString(node["secret"])
		if secret == "" {
			return "", "", "", false, errors.New("DNS node has no installation secret; recreate the node")
		}
		for _, existing := range app.ipStore.List() {
			if existing.DNSNodeID == payload.ExistingDNSNodeID {
				return "", "", "", false, errors.New("DNS node is already managed by an IP configuration")
			}
		}
		nodeName := strings.TrimSpace(valueString(node["name"]))
		if nodeName == "" {
			nodeName = safeNodeName(payload.IP)
		}
		return payload.ExistingDNSNodeID, nodeName, secret, true, nil
	}
	return "", "", "", false, errors.New("DNS node to adopt was not found")
}

func (app *App) updateIPConfig(writer http.ResponseWriter, request *http.Request, record ipConfigRecord) {
	var payload ipConfigRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	application, err := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), record.DNSNodeID, record.Routes, payload.Routes, publicBaseURL(request))
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	record.Note = payload.Note
	record.Smart = payload.Smart
	record.Routes = application.Routes
	record.TrafficPeers = application.TrafficPeers
	saved, err := app.ipStore.Save(record.IPConfig, record.NodeSecret, application.ProxyPeers)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, saved)
}

func (app *App) deleteIPConfig(writer http.ResponseWriter, request *http.Request, record ipConfigRecord) {
	if _, err := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), record.DNSNodeID, record.Routes, nil, publicBaseURL(request)); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	if !record.ExternalDNSNode {
		if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+url.PathEscape(record.DNSNodeID), nil, nil); err != nil {
			writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		_ = app.nodeLabels.Delete(record.DNSNodeID)
	}
	if err := app.ipStore.Delete(record.ID); errors.Is(err, os.ErrNotExist) {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "IP 配置不存在"})
		return
	} else if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (app *App) applyIPRoutes(ctx context.Context, authorization, dnsNodeID string, previous, current map[string]string, baseURL string) (routeApplication, error) {
	const directGroup = "__prism_enhancer_direct__"

	var rules []map[string]any
	if err := app.upstreamJSON(ctx, authorization, http.MethodGet, "/api/rules", nil, &rules); err != nil {
		return routeApplication{}, err
	}
	var nodes []map[string]any
	if err := app.upstreamJSON(ctx, authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		return routeApplication{}, err
	}
	proxyNodes := make(map[string]map[string]any)
	for _, node := range nodes {
		if valueString(node["role"]) == "proxy" {
			proxyNodes[valueString(node["id"])] = node
		}
	}
	catalogServices := app.catalog.Snapshot(ctx, false).Services
	current, _ = normalizeConflictingRoutes(previous, current, catalogServices)
	services := make(map[string]Service)
	for _, service := range catalogServices {
		services[service.ID] = service
	}
	findRule := func(serviceID string) map[string]any {
		suffix := "/enhancer/rules/" + serviceID + ".list"
		for _, rule := range rules {
			if valueString(rule["source_type"]) == "all" && strings.HasSuffix(valueString(rule["value"]), suffix) {
				return rule
			}
		}
		return nil
	}
	prepareRule := func(service Service) (map[string]any, error) {
		rule := findRule(service.ID)
		if rule == nil {
			payload := map[string]any{
				"name": "Prism IP · " + service.Name, "type": "RULE-SET", "source_type": "all", "source_val": "",
				"target_type": "group", "target_val": directGroup,
				"value": strings.TrimRight(baseURL, "/") + "/enhancer/rules/" + service.ID + ".list", "enabled": true,
			}
			if err := app.upstreamJSON(ctx, authorization, http.MethodPost, "/api/rules", payload, &rule); err != nil {
				return nil, err
			}
			rules = append(rules, rule)
			return rule, nil
		}
		if valueString(rule["target_type"]) == "group" && valueString(rule["target_val"]) == directGroup {
			return rule, nil
		}
		ruleID := valueString(rule["id"])
		if ruleID == "" {
			return nil, errors.New("Controller 未返回规则 ID")
		}
		payload := make(map[string]any, len(rule))
		for key, value := range rule {
			payload[key] = value
		}
		payload["name"] = "Prism IP · " + service.Name
		payload["source_type"] = "all"
		payload["source_val"] = ""
		payload["target_type"] = "group"
		payload["target_val"] = directGroup
		payload["enabled"] = true
		if err := app.upstreamJSON(ctx, authorization, http.MethodPut, "/api/rules/"+url.PathEscape(ruleID), payload, nil); err != nil {
			return nil, err
		}
		return payload, nil
	}
	trafficPeerSet := make(map[string]struct{})
	proxyPeers := make(map[string][]string)
	for serviceID, proxyID := range current {
		proxyNode, exists := proxyNodes[proxyID]
		if !exists {
			return routeApplication{}, fmt.Errorf("解锁机不存在: %s", proxyID)
		}
		peers := nodePublicIPs(proxyNode)
		if len(peers) == 0 {
			return routeApplication{}, fmt.Errorf("解锁机没有可用公网 IP: %s", proxyID)
		}
		proxyPeers[proxyID] = append([]string(nil), peers...)
		for _, peer := range peers {
			trafficPeerSet[peer] = struct{}{}
		}
		service, ok := services[serviceID]
		if !ok {
			return routeApplication{}, fmt.Errorf("服务不存在: %s", serviceID)
		}
		rule, err := prepareRule(service)
		if err != nil {
			return routeApplication{}, err
		}
		ruleID := valueString(rule["id"])
		if ruleID == "" {
			return routeApplication{}, errors.New("Controller 未返回规则 ID")
		}
		payload := map[string]string{"dns_node_id": dnsNodeID, "proxy_node_id": proxyID}
		if err := app.upstreamJSON(ctx, authorization, http.MethodPost, "/api/rules/"+url.PathEscape(ruleID)+"/override", payload, nil); err != nil {
			return routeApplication{}, err
		}
	}
	for serviceID := range previous {
		if _, retained := current[serviceID]; retained {
			continue
		}
		rule := findRule(serviceID)
		if rule == nil {
			continue
		}
		if service, ok := services[serviceID]; ok {
			var err error
			rule, err = prepareRule(service)
			if err != nil {
				return routeApplication{}, err
			}
		}
		payload := map[string]string{"dns_node_id": dnsNodeID, "proxy_node_id": ""}
		if err := app.upstreamJSON(ctx, authorization, http.MethodPost, "/api/rules/"+url.PathEscape(valueString(rule["id"]))+"/override", payload, nil); err != nil {
			return routeApplication{}, err
		}
	}
	trafficPeers := make([]string, 0, len(trafficPeerSet))
	for peer := range trafficPeerSet {
		trafficPeers = append(trafficPeers, peer)
	}
	sort.Strings(trafficPeers)
	return routeApplication{TrafficPeers: trafficPeers, ProxyPeers: proxyPeers, Routes: current}, nil
}

func nodePublicIPs(node map[string]any) []string {
	seen := make(map[string]struct{})
	for _, field := range []string{"public_ip", "address", "ip", "ipv4", "ipv6"} {
		for _, value := range strings.FieldsFunc(valueString(node[field]), func(character rune) bool {
			return character == ',' || character == ';' || character == ' ' || character == '\t'
		}) {
			candidate := strings.Trim(value, "[]")
			if host, _, err := net.SplitHostPort(value); err == nil {
				candidate = strings.Trim(host, "[]")
			}
			if parsed := net.ParseIP(candidate); parsed != nil {
				if parsed.To4() != nil {
					seen[parsed.String()] = struct{}{}
				}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func (app *App) upstreamJSON(ctx context.Context, authorization, method, path string, payload, target any) error {
	targetURL := *app.upstream
	targetURL.Path = path
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, targetURL.String(), body)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", authorization)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := app.client.Do(request)
	if err != nil {
		return fmt.Errorf("Controller 请求失败: %w", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var message map[string]any
		_ = json.Unmarshal(data, &message)
		if value := valueString(message["error"]); value != "" {
			return errors.New(value)
		}
		return fmt.Errorf("Controller 返回 %s", response.Status)
	}
	if target != nil && len(data) > 0 {
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("解析 Controller 响应: %w", err)
		}
	}
	return nil
}

func safeNodeName(ip string) string {
	value := strings.NewReplacer(".", " ", ":", " ", "-", " ").Replace(ip)
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 58 {
		value = value[:58]
	}
	return "IP " + value
}

func valueString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func publicBaseURL(request *http.Request) string {
	scheme := request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := request.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = request.Host
	}
	return scheme + "://" + host
}

func clientIP(request *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}
