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
	IP     string            `json:"ip"`
	Note   string            `json:"note"`
	Smart  bool              `json:"smart"`
	Routes map[string]string `json:"routes"`
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
	id := strings.TrimPrefix(request.URL.Path, "/enhancer/api/ip-configs/")
	record, ok := app.ipStore.Record(id)
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "IP 配置不存在"})
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
		"id":              record.ID,
		"expected_ip":     record.IP,
		"detected_ip":     clientIP(request),
		"master":          publicBaseURL(request),
		"secret":          record.NodeSecret,
		"smart":           record.Smart,
		"dns":             "127.0.0.1",
		"traffic_peers":   record.TrafficPeers,
		"health_probes":   app.healthProbes(request.Context(), record),
		"agent_installer": "https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/main/agent_install.sh",
	})
}

func (app *App) healthProbes(ctx context.Context, record ipConfigRecord) []map[string]string {
	services := make(map[string]Service)
	for _, service := range app.catalog.Snapshot(ctx, false).Services {
		services[service.ID] = service
	}
	serviceIDs := make([]string, 0, len(record.Routes))
	for serviceID := range record.Routes {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)
	probes := make([]map[string]string, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		service, ok := services[serviceID]
		if !ok || len(service.Domains) == 0 {
			continue
		}
		probes = append(probes, map[string]string{
			"service_id":  service.ID,
			"name":        service.Name,
			"domain":      preferredProbeDomain(service),
			"unlock_test": unlockTestProvider(service),
		})
	}
	return probes
}

func unlockTestProvider(service Service) string {
	providers := map[string]string{
		"Apple TV+":                       "Apple",
		"Bilibili":                        "Bilibili Anime",
		"ChatGPT / OpenAI":                "ChatGPT",
		"Claude":                          "Claude",
		"Crunchyroll":                     "Crunchyroll",
		"DAZN":                            "Dazn",
		"Disney+":                         "Disney+",
		"Gemini":                          "Gemini",
		"HBO / Max":                       "HBO Max",
		"Microsoft Copilot Image Creator": "Microsoft Copilot",
		"Netflix":                         "Netflix",
		"Paramount+":                      "ParamountPlus",
		"Spotify":                         "Spotify Registration",
		"TikTok":                          "TikTok",
		"YouTube":                         "YouTube Region",
	}
	return providers[service.Name]
}

func preferredProbeDomain(service Service) string {
	preferred := map[string][]string{
		"Apple TV+":                       {"tv.apple.com"},
		"Bilibili":                        {"bilibili.com"},
		"ChatGPT / OpenAI":                {"chatgpt.com", "openai.com"},
		"Claude":                          {"claude.ai", "anthropic.com"},
		"Crunchyroll":                     {"crunchyroll.com"},
		"DAZN":                            {"dazn.com"},
		"Disney+":                         {"disneyplus.com", "bamgrid.com"},
		"Gemini":                          {"gemini.google.com", "bard.google.com"},
		"Google AI Studio":                {"aistudio.google.com"},
		"HBO / Max":                       {"max.com", "hbomax.com"},
		"Microsoft Copilot Image Creator": {"copilot.microsoft.com"},
		"Netflix":                         {"netflix.com"},
		"Paramount+":                      {"paramountplus.com"},
		"Spotify":                         {"spotify.com"},
		"Suno":                            {"suno.com", "suno.ai"},
		"TikTok":                          {"tiktok.com"},
		"YouTube":                         {"youtube.com"},
	}
	for _, candidate := range preferred[service.Name] {
		for _, domain := range service.Domains {
			if domain == candidate || strings.HasSuffix(candidate, "."+domain) {
				return candidate
			}
		}
	}
	return service.Domains[0]
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
	secret, err := randomToken()
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	nodeName := safeNodeName(payload.IP)
	nodePayload := map[string]any{
		"name": nodeName, "role": "dns", "public_ip": payload.IP, "country": "", "group": "ip clients", "priority": 1, "secret": secret,
	}
	var created map[string]any
	if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodPost, "/api/nodes", nodePayload, &created); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	dnsNodeID := valueString(created["id"])
	if returnedSecret := valueString(created["secret"]); returnedSecret != "" {
		secret = returnedSecret
	}
	if dnsNodeID == "" {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": "Controller 未返回 DNS 节点 ID"})
		return
	}
	trafficPeers, err := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), dnsNodeID, nil, payload.Routes, publicBaseURL(request))
	if err != nil {
		_ = app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+url.PathEscape(dnsNodeID), nil, nil)
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	config, err := app.ipStore.Save(IPConfig{IP: payload.IP, Note: payload.Note, DNSNodeID: dnsNodeID, NodeName: nodeName, Smart: payload.Smart, Routes: payload.Routes, TrafficPeers: trafficPeers}, secret)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusCreated, config)
}

func (app *App) updateIPConfig(writer http.ResponseWriter, request *http.Request, record ipConfigRecord) {
	var payload ipConfigRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	trafficPeers, err := app.applyIPRoutes(request.Context(), request.Header.Get("Authorization"), record.DNSNodeID, record.Routes, payload.Routes, publicBaseURL(request))
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	record.Note = payload.Note
	record.Smart = payload.Smart
	record.Routes = payload.Routes
	record.TrafficPeers = trafficPeers
	saved, err := app.ipStore.Save(record.IPConfig, record.NodeSecret)
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
	if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+url.PathEscape(record.DNSNodeID), nil, nil); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
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

func (app *App) applyIPRoutes(ctx context.Context, authorization, dnsNodeID string, previous, current map[string]string, baseURL string) ([]string, error) {
	const directGroup = "__prism_enhancer_direct__"

	var rules []map[string]any
	if err := app.upstreamJSON(ctx, authorization, http.MethodGet, "/api/rules", nil, &rules); err != nil {
		return nil, err
	}
	var nodes []map[string]any
	if err := app.upstreamJSON(ctx, authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		return nil, err
	}
	proxyNodes := make(map[string]map[string]any)
	for _, node := range nodes {
		if valueString(node["role"]) == "proxy" {
			proxyNodes[valueString(node["id"])] = node
		}
	}
	services := make(map[string]Service)
	for _, service := range app.catalog.Snapshot(ctx, false).Services {
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
	for serviceID, proxyID := range current {
		proxyNode, exists := proxyNodes[proxyID]
		if !exists {
			return nil, fmt.Errorf("解锁机不存在: %s", proxyID)
		}
		for _, peer := range nodePublicIPs(proxyNode) {
			trafficPeerSet[peer] = struct{}{}
		}
		service, ok := services[serviceID]
		if !ok {
			return nil, fmt.Errorf("服务不存在: %s", serviceID)
		}
		rule, err := prepareRule(service)
		if err != nil {
			return nil, err
		}
		ruleID := valueString(rule["id"])
		if ruleID == "" {
			return nil, errors.New("Controller 未返回规则 ID")
		}
		payload := map[string]string{"dns_node_id": dnsNodeID, "proxy_node_id": proxyID}
		if err := app.upstreamJSON(ctx, authorization, http.MethodPost, "/api/rules/"+url.PathEscape(ruleID)+"/override", payload, nil); err != nil {
			return nil, err
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
				return nil, err
			}
		}
		payload := map[string]string{"dns_node_id": dnsNodeID, "proxy_node_id": ""}
		if err := app.upstreamJSON(ctx, authorization, http.MethodPost, "/api/rules/"+url.PathEscape(valueString(rule["id"]))+"/override", payload, nil); err != nil {
			return nil, err
		}
	}
	trafficPeers := make([]string, 0, len(trafficPeerSet))
	for peer := range trafficPeerSet {
		trafficPeers = append(trafficPeers, peer)
	}
	sort.Strings(trafficPeers)
	return trafficPeers, nil
}

func nodePublicIPs(node map[string]any) []string {
	seen := make(map[string]struct{})
	for _, field := range []string{"public_ip", "address"} {
		for _, value := range strings.FieldsFunc(valueString(node[field]), func(character rune) bool {
			return character == ',' || character == ';' || character == ' ' || character == '\t'
		}) {
			candidate := strings.Trim(value, "[]")
			if host, _, err := net.SplitHostPort(value); err == nil {
				candidate = strings.Trim(host, "[]")
			}
			if parsed := net.ParseIP(candidate); parsed != nil {
				seen[parsed.String()] = struct{}{}
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
