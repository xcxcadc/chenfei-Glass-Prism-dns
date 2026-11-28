package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	quickGeminiRegionRe = regexp.MustCompile(`,2,1,200,"([A-Z]{2,3})"`)
	quickGeminiTrueRe   = regexp.MustCompile(`(?:"45631641"|45631641\\"):true|45631641,null,true`)
	quickGeminiFalseRe  = regexp.MustCompile(`(?:"45631641"|45631641\\"):false|45631641,null,false`)
	quickYouTubeGLRe    = regexp.MustCompile(`"INNERTUBE_CONTEXT_GL"\s*:\s*"([A-Z]{2})"|"countryCode"\s*:\s*"([A-Z]{2})"`)
	quickPlayCountryRe  = regexp.MustCompile(`<div class="yVZQTb">([^<(]+)`)
	quickNetflixLocRe   = regexp.MustCompile(`netflix\.com/([a-z]{2})(?:-[a-z]+)?/title/`)
	quickTraceLocRe     = regexp.MustCompile(`(?m)^loc=([A-Z]{2})\s*$`)
)

type quickCheckRequest struct {
	ServiceID string   `json:"service_id"`
	ProxyIDs  []string `json:"proxy_ids,omitempty"`
}

type quickCheckItem struct {
	ProxyID   string `json:"proxy_id"`
	ProxyName string `json:"proxy_name"`
	ProxyIP   string `json:"proxy_ip"`
	Country   string `json:"country,omitempty"`
	Online    bool   `json:"online"`
	Status    string `json:"status"` // "yes", "no", "warn", "error"
	Summary   string `json:"summary"`
	Region    string `json:"region,omitempty"`
	LatencyMS int64  `json:"latency_ms"`
}

func (app *App) handleQuickCheck(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	authorization := request.Header.Get("Authorization")
	if !app.authorize(request.Context(), authorization) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	var payload quickCheckRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	service, ok := app.catalog.Service(request.Context(), strings.TrimSpace(payload.ServiceID))
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "服务不存在"})
		return
	}

	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	filter := make(map[string]struct{}, len(payload.ProxyIDs))
	for _, id := range payload.ProxyIDs {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			filter[trimmed] = struct{}{}
		}
	}

	type targetProxy struct {
		id      string
		name    string
		country string
		ip      string
		online  bool
	}
	var targets []targetProxy
	for _, node := range nodes {
		if valueString(node["role"]) != "proxy" {
			continue
		}
		app.nodeLabels.Overlay(node)
		id := valueString(node["id"])
		if len(filter) > 0 {
			if _, wanted := filter[id]; !wanted {
				continue
			}
		}
		ip := ""
		if app.transport != nil {
			ip = app.transport.ProxyEndpoint(id)
		}
		if ip == "" {
			peers := nodePublicIPv4s(node)
			if len(peers) > 0 {
				ip = peers[0]
			}
		}
		online := nodeRecentlyOnline(valueString(node["last_heartbeat"]))
		targets = append(targets, targetProxy{
			id:      id,
			name:    valueString(node["name"]),
			country: valueString(node["country"]),
			ip:      ip,
			online:  online,
		})
	}

	results := make([]quickCheckItem, len(targets))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for idx, target := range targets {
		wg.Add(1)
		go func(index int, item targetProxy) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if item.ip == "" {
				results[index] = quickCheckItem{
					ProxyID: item.id, ProxyName: item.name, Country: item.country, Online: item.online,
					Status: "error", Summary: "缺少可用公网 IPv4",
				}
				return
			}
			ctx, cancel := context.WithTimeout(request.Context(), 9*time.Second)
			defer cancel()
			res := runProxyQuickProbe(ctx, item.ip, service)
			res.ProxyID = item.id
			res.ProxyName = item.name
			res.ProxyIP = item.ip
			res.Country = item.country
			res.Online = item.online
			results[index] = res
		}(idx, target)
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		rank := func(s string) int {
			switch s {
			case "yes":
				return 0
			case "warn":
				return 1
			case "no":
				return 2
			default:
				return 3
			}
		}
		ri, rj := rank(results[i].Status), rank(results[j].Status)
		if ri != rj {
			return ri < rj
		}
		if results[i].LatencyMS > 0 && results[j].LatencyMS > 0 && results[i].LatencyMS != results[j].LatencyMS {
			return results[i].LatencyMS < results[j].LatencyMS
		}
		return results[i].ProxyName < results[j].ProxyName
	})

	writeJSON(writer, http.StatusOK, map[string]any{
		"service_id":   service.ID,
		"service_name": service.Name,
		"checked_at":   time.Now().UTC(),
		"results":      results,
	})
}

func nodeRecentlyOnline(lastHeartbeat string) bool {
	lastHeartbeat = strings.TrimSpace(lastHeartbeat)
	if lastHeartbeat == "" {
		return false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05-07:00"} {
		if parsed, err := time.Parse(layout, lastHeartbeat); err == nil {
			return time.Since(parsed) < 90*time.Second
		}
	}
	return false
}

func newProxySNIClient(proxyIP string, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy:               nil,
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
		DialTLSContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				host = addr
			}
			rawConn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp4", net.JoinHostPort(proxyIP, "443"))
			if err != nil {
				return nil, err
			}
			tlsConn := tls.Client(rawConn, &tls.Config{
				ServerName: host,
				MinVersion: tls.VersionTLS12,
				NextProtos: []string{"http/1.1"},
			})
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				_ = rawConn.Close()
				return nil, err
			}
			return tlsConn, nil
		},
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

func fetchViaProxySNI(ctx context.Context, client *http.Client, targetURL string) (int, string, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, "", nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1536*1024))
	finalURL := targetURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return resp.StatusCode, finalURL, body, nil
}

func runProxyQuickProbe(ctx context.Context, proxyIP string, service Service) quickCheckItem {
	started := time.Now()
	client := newProxySNIClient(proxyIP, 8*time.Second)
	key := strings.ToLower(strings.TrimSpace(service.Name))

	finish := func(status, summary, region string) quickCheckItem {
		return quickCheckItem{
			Status:    status,
			Summary:   summary,
			Region:    region,
			LatencyMS: time.Since(started).Milliseconds(),
		}
	}

	switch {
	case strings.Contains(key, "gemini") || strings.Contains(key, "ai studio"):
		code, finalURL, body, err := fetchViaProxySNI(ctx, client, "https://gemini.google.com/app")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		region := ""
		if m := quickGeminiRegionRe.FindSubmatch(body); len(m) == 2 {
			region = string(m[1])
		}
		if quickGeminiTrueRe.Match(body) && !strings.Contains(finalURL, "/faq") {
			if region != "" {
				return finish("yes", fmt.Sprintf("Yes (Region: %s)", region), region)
			}
			return finish("yes", "Yes", "")
		}
		if quickGeminiFalseRe.Match(body) || region == "CHN" || region == "HKG" || strings.Contains(finalURL, "/faq") {
			if region == "CHN" {
				return finish("no", "No (Region: CHN · Google送中)", "CHN")
			}
			if region != "" {
				return finish("no", fmt.Sprintf("No (Region: %s)", region), region)
			}
			return finish("no", "No (Region Blocked)", "")
		}
		return finish("warn", fmt.Sprintf("HTTP %d (Region: %s)", code, nonEmpty(region, "Unknown")), region)

	case strings.Contains(key, "chatgpt") || strings.Contains(key, "openai"):
		_, _, traceBody, err := fetchViaProxySNI(ctx, client, "https://chatgpt.com/cdn-cgi/trace")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		region := ""
		if m := quickTraceLocRe.FindSubmatch(traceBody); len(m) == 2 {
			region = string(m[1])
		}
		code, _, statusBody, err2 := fetchViaProxySNI(ctx, client, "https://ios.chat.openai.com/public-api/mobile/server_status/v1")
		bodyLower := strings.ToLower(string(statusBody))
		if err2 == nil && (strings.Contains(bodyLower, "unsupported_country") || region == "CN" || region == "HK" || region == "RU") {
			return finish("no", fmt.Sprintf("No (Region: %s)", nonEmpty(region, "Blocked")), region)
		}
		if err2 == nil && code >= 200 && code < 400 {
			if region != "" {
				return finish("yes", fmt.Sprintf("Yes (Region: %s)", region), region)
			}
			return finish("yes", "Yes", "")
		}
		if region != "" && region != "CN" && region != "HK" && region != "RU" {
			return finish("yes", fmt.Sprintf("Yes (Web · Region: %s)", region), region)
		}
		return finish("warn", fmt.Sprintf("HTTP %d (Region: %s)", code, nonEmpty(region, "?")), region)

	case strings.Contains(key, "claude"):
		code, finalURL, _, err := fetchViaProxySNI(ctx, client, "https://claude.ai/")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if strings.Contains(strings.ToLower(finalURL), "unavailable-in-region") {
			return finish("no", "No (Region Blocked)", "")
		}
		region := ""
		if _, _, traceBody, traceErr := fetchViaProxySNI(ctx, client, "https://claude.ai/cdn-cgi/trace"); traceErr == nil {
			if m := quickTraceLocRe.FindSubmatch(traceBody); len(m) == 2 {
				region = string(m[1])
			}
		}
		if code >= 200 && code < 405 {
			if region != "" {
				return finish("yes", fmt.Sprintf("Yes (Region: %s)", region), region)
			}
			return finish("yes", "Yes", "")
		}
		return finish("warn", fmt.Sprintf("HTTP %d", code), region)

	case strings.Contains(key, "youtube"):
		code, _, body, err := fetchViaProxySNI(ctx, client, "https://www.youtube.com/premium")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		text := string(body)
		region := ""
		if m := quickYouTubeGLRe.FindStringSubmatch(text); len(m) == 3 {
			region = nonEmpty(m[1], m[2])
		}
		if strings.Contains(strings.ToLower(text), "youtube premium is not available in your country") || region == "CN" {
			if region == "CN" {
				return finish("no", "No (CN · Google送中)", "CN")
			}
			return finish("no", fmt.Sprintf("No (Region: %s)", nonEmpty(region, "Blocked")), region)
		}
		if code >= 200 && code < 400 {
			if region != "" {
				return finish("yes", fmt.Sprintf("Yes (Region: %s)", region), region)
			}
			return finish("yes", "Yes", "")
		}
		return finish("warn", fmt.Sprintf("HTTP %d", code), region)

	case strings.Contains(key, "play store"):
		code, _, body, err := fetchViaProxySNI(ctx, client, "https://play.google.com/")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if m := quickPlayCountryRe.FindSubmatch(body); len(m) == 2 {
			country := strings.TrimSpace(string(m[1]))
			if strings.EqualFold(country, "China") {
				return finish("no", "China (Google送中)", "CN")
			}
			return finish("yes", fmt.Sprintf("Yes (%s)", country), country)
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Reachable (HTTP %d)", code), "")
		}
		return finish("warn", fmt.Sprintf("HTTP %d", code), "")

	case strings.Contains(key, "x (twitter)") || key == "x" || strings.Contains(key, "twitter"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://x.com/robots.txt")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "instagram"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://www.instagram.com/robots.txt")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "facebook"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://www.facebook.com/robots.txt")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "threads"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://www.threads.net/robots.txt")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "whatsapp"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://www.whatsapp.com/")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "telegram"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://telegram.org/")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "discord"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://discord.com/api/v10/gateway")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "reddit"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://www.reddit.com/robots.txt")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "pinterest"):
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://www.pinterest.com/robots.txt")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case key == "line":
		code, _, _, err := fetchViaProxySNI(ctx, client, "https://line.me/")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		if code >= 200 && code < 400 {
			return finish("yes", fmt.Sprintf("Yes (HTTP %d)", code), "")
		}
		return finish("no", fmt.Sprintf("No (HTTP %d)", code), "")

	case strings.Contains(key, "netflix"):
		code, finalURL, body, err := fetchViaProxySNI(ctx, client, "https://www.netflix.com/title/81280792")
		if err != nil {
			return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
		}
		region := ""
		if m := quickNetflixLocRe.FindStringSubmatch(strings.ToLower(finalURL)); len(m) == 2 {
			region = strings.ToUpper(m[1])
		}
		if code == 200 && !strings.Contains(strings.ToLower(string(body)), "oh no!") {
			if region != "" {
				return finish("yes", fmt.Sprintf("Yes (Region: %s)", region), region)
			}
			return finish("yes", "Yes (Full Unlock)", "")
		}
		if code == 404 {
			codeOrig, origURL, _, errOrig := fetchViaProxySNI(ctx, client, "https://www.netflix.com/title/80018499")
			if errOrig == nil && codeOrig == 200 {
				if m := quickNetflixLocRe.FindStringSubmatch(strings.ToLower(origURL)); len(m) == 2 {
					region = strings.ToUpper(m[1])
				}
				return finish("warn", fmt.Sprintf("Originals Only (%s)", nonEmpty(region, "自制剧")), region)
			}
			return finish("no", "No (Not Available)", region)
		}
		if code == 403 {
			return finish("no", "No (IP Banned / 403)", region)
		}
		return finish("warn", fmt.Sprintf("HTTP %d", code), region)
	}

	probeDomain := preferredProbeDomain(service)
	if probeDomain == "" && len(service.Domains) > 0 {
		probeDomain = routingDomain(service.Domains[0])
	}
	if probeDomain == "" {
		return finish("warn", "仅关键词/CIDR规则，无需域名探测", "")
	}
	code, _, body, err := fetchViaProxySNI(ctx, client, "https://"+probeDomain+"/")
	if err != nil {
		return finish("error", "SNI 连接失败: "+compactProbeErr(err), "")
	}
	lower := strings.ToLower(string(body))
	if code == 451 || strings.Contains(lower, "not available in your country") || strings.Contains(lower, "not supported in your country") {
		return finish("no", fmt.Sprintf("No (Region Blocked · HTTP %d)", code), "")
	}
	if code >= 200 && code < 405 {
		return finish("yes", fmt.Sprintf("SNI 连通正常 (HTTP %d)", code), "")
	}
	return finish("warn", fmt.Sprintf("SNI 已响应 (HTTP %d)", code), "")
}

func compactProbeErr(err error) string {
	msg := strings.TrimSpace(err.Error())
	if strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "context deadline exceeded") {
		return "连接超时（请确认解锁机在线且已放行面板 IP 访问 443）"
	}
	if strings.Contains(msg, "connection refused") {
		return "443 端口拒绝连接"
	}
	if len(msg) > 72 {
		return msg[:72] + "..."
	}
	return msg
}

func nonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

type healOrphansRequest struct {
	ReplacementProxyID string `json:"replacement_proxy_id"`
	RemoveOnly         bool   `json:"remove_only"`
}

func (app *App) handleHealOrphans(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	authorization := request.Header.Get("Authorization")
	if !app.authorize(request.Context(), authorization) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	var payload healOrphansRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var nodes []map[string]any
	if err := app.upstreamJSON(request.Context(), authorization, http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	validProxies := make(map[string]bool)
	for _, node := range nodes {
		if valueString(node["role"]) == "proxy" {
			validProxies[valueString(node["id"])] = true
		}
	}
	replacement := strings.TrimSpace(payload.ReplacementProxyID)
	if !payload.RemoveOnly && replacement != "" && !validProxies[replacement] {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "指定的替代解锁机不存在"})
		return
	}

	ipConfigWriteMu.Lock()
	defer ipConfigWriteMu.Unlock()

	records := app.ipStore.Records()
	baseURL := publicBaseURL(request)
	healedConfigs := 0
	healedRoutes := 0

	for _, record := range records {
		nextRoutes := make(map[string]string, len(record.Routes))
		changed := false
		for serviceID, proxyID := range record.Routes {
			if validProxies[proxyID] {
				nextRoutes[serviceID] = proxyID
				continue
			}
			changed = true
			healedRoutes++
			if !payload.RemoveOnly && replacement != "" {
				nextRoutes[serviceID] = replacement
			}
		}
		if !changed {
			continue
		}
		reconciled, err := app.reconcileIPConfigNode(request.Context(), authorization, record)
		if err != nil {
			writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		application, err := app.applyIPRoutes(request.Context(), authorization, reconciled.DNSNodeID, reconciled.Routes, nextRoutes, baseURL)
		if err != nil {
			writeIPRouteError(writer, err)
			return
		}
		reconciled.Routes = application.Routes
		reconciled.TrafficPeers = application.TrafficPeers
		if _, err := app.ipStore.Save(reconciled.IPConfig, reconciled.NodeSecret, application.ProxyPeers); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		healedConfigs++
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"healed_configs": healedConfigs,
		"healed_routes":  healedRoutes,
	})
}
