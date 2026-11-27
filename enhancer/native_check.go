package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	nativeCheckCacheTTL   = 5 * time.Minute
	nativeCheckTaskExpiry = 45 * time.Second
)

type NativeProbeTask struct {
	TaskID      string    `json:"task_id"`
	ServiceID   string    `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Domain      string    `json:"domain"`
	ProbeURL    string    `json:"probe_url"`
	CreatedAt   time.Time `json:"-"`
}

type NativeCheckResult struct {
	ConfigID  string    `json:"config_id"`
	ServiceID string    `json:"service_id"`
	Status    string    `json:"status"` // "yes", "no", "warn", "error"
	Summary   string    `json:"summary"`
	Region    string    `json:"region,omitempty"`
	LatencyMS int64     `json:"latency_ms"`
	CheckedAt time.Time `json:"checked_at"`
}

type NativeCheckStore struct {
	mu      sync.Mutex
	results map[string]NativeCheckResult // key: configID + "|" + serviceID
	pending map[string]map[string]NativeProbeTask // configID -> serviceID -> task
}

func NewNativeCheckStore() *NativeCheckStore {
	return &NativeCheckStore{
		results: make(map[string]NativeCheckResult),
		pending: make(map[string]map[string]NativeProbeTask),
	}
}

func nativeCheckKey(configID, serviceID string) string {
	return configID + "|" + serviceID
}

func nativeProbeTarget(service Service) (domain, probeURL string) {
	nameLower := strings.ToLower(strings.TrimSpace(service.Name))
	switch {
	case strings.Contains(nameLower, "gemini") || strings.Contains(nameLower, "ai studio"):
		return "gemini.google.com", "https://gemini.google.com/"
	case strings.Contains(nameLower, "claude"):
		return "claude.ai", "https://claude.ai/cdn-cgi/trace"
	case strings.Contains(nameLower, "chatgpt") || strings.Contains(nameLower, "openai"):
		return "chatgpt.com", "https://chatgpt.com/cdn-cgi/trace"
	case strings.Contains(nameLower, "netflix"):
		return "www.netflix.com", "https://www.netflix.com/title/81280792"
	case strings.Contains(nameLower, "disney"):
		return "www.disneyplus.com", "https://www.disneyplus.com/"
	case strings.Contains(nameLower, "youtube"):
		return "www.youtube.com", "https://www.youtube.com/premium"
	case strings.Contains(nameLower, "play store"):
		return "play.google.com", "https://play.google.com/store/apps"
	case strings.Contains(nameLower, "grok"):
		return "grok.com", "https://grok.com/"
	case strings.Contains(nameLower, "tiktok"):
		return "www.tiktok.com", "https://www.tiktok.com/"
	case strings.Contains(nameLower, "prime video"):
		return "www.primevideo.com", "https://www.primevideo.com/"
	}
	for _, d := range service.Domains {
		clean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(d)), "*.")
		if clean != "" && strings.Contains(clean, ".") {
			return clean, "https://" + clean + "/"
		}
	}
	return "", ""
}

func randomTaskID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return "nc-" + hex.EncodeToString(b[:])
}

func (store *NativeCheckStore) GetCached(configID, serviceID string) (NativeCheckResult, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	res, ok := store.results[nativeCheckKey(configID, serviceID)]
	if !ok || time.Since(res.CheckedAt) > nativeCheckCacheTTL {
		return NativeCheckResult{}, false
	}
	return res, true
}

func (store *NativeCheckStore) IsPending(configID, serviceID string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	tasks := store.pending[configID]
	task, ok := tasks[serviceID]
	if !ok {
		return false
	}
	if time.Since(task.CreatedAt) > nativeCheckTaskExpiry {
		delete(tasks, serviceID)
		return false
	}
	return true
}

func (store *NativeCheckStore) Request(configID string, service Service, force bool) (*NativeCheckResult, bool, string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	key := nativeCheckKey(configID, service.ID)
	if !force {
		if res, ok := store.results[key]; ok && time.Since(res.CheckedAt) <= nativeCheckCacheTTL {
			copied := res
			return &copied, false, ""
		}
	}
	if store.pending[configID] == nil {
		store.pending[configID] = make(map[string]NativeProbeTask)
	}
	if existing, ok := store.pending[configID][service.ID]; ok && time.Since(existing.CreatedAt) <= nativeCheckTaskExpiry {
		return nil, true, existing.TaskID
	}
	domain, probeURL := nativeProbeTarget(service)
	if domain == "" {
		res := NativeCheckResult{
			ConfigID:  configID,
			ServiceID: service.ID,
			Status:    "warn",
			Summary:   "无可用探测域名",
			CheckedAt: time.Now().UTC(),
		}
		store.results[key] = res
		return &res, false, ""
	}
	task := NativeProbeTask{
		TaskID:      randomTaskID(),
		ServiceID:   service.ID,
		ServiceName: service.Name,
		Domain:      domain,
		ProbeURL:    probeURL,
		CreatedAt:   time.Now().UTC(),
	}
	store.pending[configID][service.ID] = task
	return nil, true, task.TaskID
}

func (store *NativeCheckStore) PendingForClient(configID string) []NativeProbeTask {
	if store == nil || configID == "" {
		return nil
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	tasks := store.pending[configID]
	if len(tasks) == 0 {
		return nil
	}
	now := time.Now()
	out := make([]NativeProbeTask, 0, len(tasks))
	for svcID, task := range tasks {
		if now.Sub(task.CreatedAt) > nativeCheckTaskExpiry {
			delete(tasks, svcID)
			continue
		}
		out = append(out, task)
		if len(out) >= 4 {
			break
		}
	}
	return out
}

func evaluateNativeProbeResponse(service Service, code int, finalURL, body string, probeErr string) (status, region, summary string) {
	if strings.TrimSpace(probeErr) != "" || code == 0 {
		msg := strings.TrimSpace(probeErr)
		if msg == "" {
			msg = "原生连接超时/失败"
		}
		return "error", "", msg
	}
	nameLower := strings.ToLower(strings.TrimSpace(service.Name))
	lowerBody := strings.ToLower(body)
	lowerURL := strings.ToLower(finalURL)

	switch {
	case strings.Contains(nameLower, "gemini") || strings.Contains(nameLower, "ai studio"):
		if m := quickGeminiRegionRe.FindStringSubmatch(body); len(m) > 1 {
			region = m[1]
		}
		isTrue := quickGeminiTrueRe.MatchString(body)
		isFalse := quickGeminiFalseRe.MatchString(body)
		if isFalse || region == "CHN" || region == "HKG" || region == "RUS" || region == "BLR" || region == "MAC" {
			if region != "" {
				return "no", region, fmt.Sprintf("原生不可用 (%s)", region)
			}
			return "no", "", "原生不可用"
		}
		if isTrue || (code == 200 && region != "") {
			return "yes", region, fmt.Sprintf("原生可用 (%s)", region)
		}
		if code == 200 {
			return "warn", region, "HTTP 200 (待确认)"
		}
		return "no", region, fmt.Sprintf("HTTP %d", code)

	case strings.Contains(nameLower, "claude") || strings.Contains(nameLower, "chatgpt") || strings.Contains(nameLower, "openai"):
		if m := quickTraceLocRe.FindStringSubmatch(body); len(m) > 1 {
			region = m[1]
		}
		if region == "CN" || region == "HK" || region == "MO" || region == "RU" || region == "BY" || region == "IR" || region == "KP" {
			return "no", region, fmt.Sprintf("原生地区受限 (%s)", region)
		}
		if code >= 200 && code < 400 && region != "" {
			return "yes", region, fmt.Sprintf("原生可用 (%s)", region)
		}
		if code == 403 {
			return "no", region, "原生被拦截 (403)"
		}
		return "warn", region, fmt.Sprintf("HTTP %d", code)

	case strings.Contains(nameLower, "netflix"):
		if m := quickNetflixLocRe.FindStringSubmatch(lowerURL); len(m) > 1 {
			region = strings.ToUpper(m[1])
		}
		if code == 200 && !strings.Contains(lowerBody, "not available") {
			if region != "" {
				return "yes", region, fmt.Sprintf("原生完整解锁 (%s)", region)
			}
			return "yes", "", "原生可用"
		}
		if code == 404 {
			return "no", region, "仅自制剧 (Originals Only)"
		}
		return "no", region, fmt.Sprintf("原生不可用 (HTTP %d)", code)

	case strings.Contains(nameLower, "disney"):
		if strings.Contains(lowerURL, "unavailable") || strings.Contains(lowerBody, "not available in your region") || code == 403 {
			return "no", "", "原生地区不可用"
		}
		if code >= 200 && code < 400 {
			return "yes", "", "原生可用"
		}
		return "no", "", fmt.Sprintf("HTTP %d", code)

	case strings.Contains(nameLower, "youtube"):
		if m := quickYouTubeGLRe.FindStringSubmatch(body); len(m) > 1 {
			if m[1] != "" {
				region = m[1]
			} else {
				region = m[2]
			}
		}
		if strings.Contains(lowerBody, "youtube premium is not available in your country") || region == "CN" || region == "RU" {
			return "no", region, fmt.Sprintf("原生不可用 (%s)", region)
		}
		if code == 200 {
			if region != "" {
				return "yes", region, fmt.Sprintf("原生可用 (%s)", region)
			}
			return "yes", "", "原生可用"
		}
		return "no", region, fmt.Sprintf("HTTP %d", code)

	case strings.Contains(nameLower, "play store"):
		if m := quickPlayCountryRe.FindStringSubmatch(body); len(m) > 1 {
			region = strings.TrimSpace(m[1])
		}
		if strings.HasPrefix(strings.ToLower(region), "china") {
			return "no", region, "原生送中 (China)"
		}
		if code == 200 && region != "" {
			return "yes", region, fmt.Sprintf("原生可用 (%s)", region)
		}
		return "warn", region, fmt.Sprintf("HTTP %d", code)
	}

	if parsedURL, err := url.Parse(finalURL); err == nil && strings.Contains(strings.ToLower(parsedURL.Path), "unavailable") {
		return "no", "", "原生地区不可用"
	}
	if code == 403 || code == 451 {
		return "no", "", fmt.Sprintf("原生受限 (HTTP %d)", code)
	}
	if code >= 200 && code < 400 {
		return "yes", "", fmt.Sprintf("原生可直连 (HTTP %d)", code)
	}
	return "warn", "", fmt.Sprintf("HTTP %d", code)
}

func (store *NativeCheckStore) Complete(configID, serviceID string, service Service, code int, finalURL, body string, latencyMS int64, probeErr string) NativeCheckResult {
	status, region, summary := evaluateNativeProbeResponse(service, code, finalURL, body, probeErr)
	res := NativeCheckResult{
		ConfigID:  configID,
		ServiceID: serviceID,
		Status:    status,
		Region:    region,
		Summary:   summary,
		LatencyMS: latencyMS,
		CheckedAt: time.Now().UTC(),
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.results[nativeCheckKey(configID, serviceID)] = res
	if tasks := store.pending[configID]; tasks != nil {
		delete(tasks, serviceID)
	}
	return res
}

type clientNativeCheckRequest struct {
	ConfigID  string `json:"config_id"`
	ServiceID string `json:"service_id"`
	Force     bool   `json:"force,omitempty"`
}

type clientNativeReportRequest struct {
	Token       string `json:"token"`
	TaskID      string `json:"task_id,omitempty"`
	ServiceID   string `json:"service_id"`
	HTTPCode    int    `json:"http_code"`
	FinalURL    string `json:"final_url,omitempty"`
	BodySnippet string `json:"body_snippet,omitempty"`
	LatencyMS   int64  `json:"latency_ms,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (app *App) handleClientNativeCheck(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	switch request.Method {
	case http.MethodGet:
		configID := strings.TrimSpace(request.URL.Query().Get("config_id"))
		serviceID := strings.TrimSpace(request.URL.Query().Get("service_id"))
		if configID == "" || serviceID == "" {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "config_id 和 service_id 不能为空"})
			return
		}
		if res, ok := app.nativeChecks.GetCached(configID, serviceID); ok {
			writeJSON(writer, http.StatusOK, map[string]any{"status": "ready", "result": res})
			return
		}
		if app.nativeChecks.IsPending(configID, serviceID) {
			writeJSON(writer, http.StatusOK, map[string]any{"status": "pending"})
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"status": "idle"})

	case http.MethodPost:
		var payload clientNativeCheckRequest
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		configID := strings.TrimSpace(payload.ConfigID)
		serviceID := strings.TrimSpace(payload.ServiceID)
		if configID == "" || serviceID == "" {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "config_id 和 service_id 不能为空"})
			return
		}
		service, ok := app.catalog.Service(request.Context(), serviceID)
		if !ok {
			writeJSON(writer, http.StatusNotFound, map[string]string{"error": "服务不存在"})
			return
		}
		res, pending, taskID := app.nativeChecks.Request(configID, service, payload.Force)
		if res != nil {
			writeJSON(writer, http.StatusOK, map[string]any{"status": "ready", "result": res})
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"status": "pending", "pending": pending, "task_id": taskID})

	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPost)
	}
}

func (app *App) handleClientNativeReport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload clientNativeReportRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	token := strings.TrimSpace(payload.Token)
	if app.ipStore == nil || token == "" {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid client token"})
		return
	}
	matched, found := app.ipStore.GetByToken(token)
	if !found {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid client token"})
		return
	}
	service, ok := app.catalog.Service(request.Context(), strings.TrimSpace(payload.ServiceID))
	if !ok {
		service = Service{ID: strings.TrimSpace(payload.ServiceID), Name: strings.TrimSpace(payload.ServiceID)}
	}
	res := app.nativeChecks.Complete(matched.ID, service.ID, service, payload.HTTPCode, payload.FinalURL, payload.BodySnippet, payload.LatencyMS, payload.Error)
	writeJSON(writer, http.StatusOK, map[string]any{"ok": true, "result": res})
}
