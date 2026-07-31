package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed web/*
var embeddedWeb embed.FS

const uiVersion = "1.5.4"

type App struct {
	catalog      *CatalogManager
	store        *CustomServiceStore
	ipStore      *IPConfigStore
	nodeLabels   *NodeLabelStore
	branding     *BrandingStore
	preferences  *CatalogPreferenceStore
	transport    *TransportStore
	upstream     *url.URL
	proxy        *httputil.ReverseProxy
	client       *http.Client
	icons        *serviceIconCache
	web          fs.FS
	indexHTML    []byte
	controllerDB string
}

func NewApp(upstreamURL string, catalog *CatalogManager, store *CustomServiceStore, ipStore *IPConfigStore, client *http.Client, controllerDB ...string) (*App, error) {
	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("parse upstream URL: %w", err)
	}
	web, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		return nil, fmt.Errorf("open embedded web assets: %w", err)
	}
	indexHTML, err := fs.ReadFile(web, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded index: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.FlushInterval = -1
	originalDirector := proxy.Director
	proxy.Director = func(request *http.Request) {
		originalDirector(request)
		if request.URL.Path == "/api/sync" {
			request.Header.Del("Accept-Encoding")
		}
	}
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, proxyErr error) {
		log.Printf("upstream request failed: %s %s: %v", request.Method, request.URL.Path, proxyErr)
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": "Controller 暂时不可用"})
	}
	databasePath := "/opt/prism/data.db"
	if len(controllerDB) > 0 && strings.TrimSpace(controllerDB[0]) != "" {
		databasePath = controllerDB[0]
	}
	app := &App{
		catalog:      catalog,
		store:        store,
		ipStore:      ipStore,
		nodeLabels:   &NodeLabelStore{labels: make(map[string]NodeLabel)},
		branding:     &BrandingStore{},
		preferences:  catalog.preferences,
		upstream:     upstream,
		proxy:        proxy,
		client:       client,
		icons:        newServiceIconCache(),
		web:          web,
		indexHTML:    indexHTML,
		controllerDB: databasePath,
	}
	app.transport, _ = NewTransportStore("")
	proxy.ModifyResponse = app.modifyUpstreamResponse
	return app, nil
}

func (app *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/enhancer/api/health", app.handleHealth)
	mux.HandleFunc("/enhancer/api/catalog", app.handleCatalog)
	mux.HandleFunc("/enhancer/api/custom-services", app.handleCustomServices)
	mux.HandleFunc("/enhancer/api/custom-services/", app.handleCustomService)
	mux.HandleFunc("/enhancer/api/service-domains/", app.handleServiceDomains)
	mux.HandleFunc("/enhancer/api/categories", app.handleCategories)
	mux.HandleFunc("/enhancer/api/service-categories/", app.handleServiceCategory)
	mux.HandleFunc("/enhancer/api/connectivity", app.handleConnectivity)
	mux.HandleFunc("/enhancer/api/account", app.handleAccountUpdate)
	mux.HandleFunc("/enhancer/api/branding", app.handleBranding)
	mux.HandleFunc("/enhancer/api/nodes", app.handleEnhancedNodes)
	mux.HandleFunc("/enhancer/api/nodes/", app.handleEnhancedNode)
	mux.HandleFunc("/enhancer/api/ip-configs", app.handleIPConfigs)
	mux.HandleFunc("/enhancer/api/ip-configs/", app.handleIPConfig)
	mux.HandleFunc("/enhancer/api/bootstrap/", app.handleBootstrap)
	mux.HandleFunc("/enhancer/api/transport/proxy", app.handleProxyTransport)
	mux.HandleFunc("/enhancer/api/transport/client", app.handleClientTransport)
	mux.HandleFunc("/enhancer/api/traffic/report", app.handleTrafficReport)
	mux.HandleFunc("/enhancer/api/audit/report", app.handleServiceAuditReport)
	mux.HandleFunc("/enhancer/api/traffic/", app.handleTrafficClear)
	mux.HandleFunc("/enhancer/rules/", app.handleRuleSet)
	mux.HandleFunc("/enhancer/icons/", app.handleServiceIcon)
	mux.Handle("/assets/", noStore(http.StripPrefix("/assets/", http.FileServer(http.FS(app.web)))))
	mux.HandleFunc("/", app.handleRoot)
	return securityHeaders(requestLogger(mux))
}

func (app *App) handleRoot(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/favicon.ico" && request.Method == http.MethodGet {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	if request.URL.Path == "/" && request.Method == http.MethodGet {
		setNoCacheHeaders(writer)
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(app.indexHTML)
		return
	}
	app.proxy.ServeHTTP(writer, request)
}

func (app *App) handleHealth(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	snapshot := app.catalog.Snapshot(request.Context(), false)
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":                     "ok",
		"catalog_updated":            snapshot.UpdatedAt,
		"catalog_error":              snapshot.Error,
		"service_count":              len(snapshot.Services),
		"custom_count":               len(app.store.List()),
		"controller_target":          app.upstream.String(),
		"ui_version":                 uiVersion,
		"authorization_mode":         "panel_allowlist",
		"authorization_sync_seconds": 5,
	})
}

func (app *App) handleCatalog(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	force := request.URL.Query().Get("refresh") == "1"
	snapshot := app.catalog.Snapshot(request.Context(), force)
	status := http.StatusOK
	if len(snapshot.Services) == 0 && snapshot.Error != "" {
		status = http.StatusBadGateway
	}
	writeJSON(writer, status, snapshot)
}

func (app *App) handleCustomServices(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		writeJSON(writer, http.StatusOK, app.store.List())
	case http.MethodPost:
		if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
			return
		}
		var service Service
		if err := decodeJSON(request, &service); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		requestedCategory := strings.TrimSpace(service.Category)
		if requestedCategory == "" {
			requestedCategory = "自定义服务"
		}
		categoryWasKnown := app.catalogHasCategory(request.Context(), requestedCategory)
		saved, err := app.store.Upsert(service)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if !categoryWasKnown {
			if _, err := app.preferences.AddCategory(saved.Category); err != nil {
				writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		if err := app.preferences.ClearServiceCategory(saved.ID); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if err := app.preferences.ClearServiceDomains(saved.ID); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusCreated, saved)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPost)
	}
}

func (app *App) handleCustomService(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	id := strings.TrimPrefix(request.URL.Path, "/enhancer/api/custom-services/")
	if id == "" || !strings.HasPrefix(id, "custom-") {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "无效的服务 ID"})
		return
	}
	switch request.Method {
	case http.MethodPut:
		var service Service
		if err := decodeJSON(request, &service); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		requestedCategory := strings.TrimSpace(service.Category)
		if requestedCategory == "" {
			requestedCategory = "自定义服务"
		}
		categoryWasKnown := app.catalogHasCategory(request.Context(), requestedCategory)
		service.ID = id
		saved, err := app.store.Upsert(service)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if !categoryWasKnown {
			if _, err := app.preferences.AddCategory(saved.Category); err != nil {
				writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		if err := app.preferences.ClearServiceCategory(saved.ID); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusOK, saved)
	case http.MethodDelete:
		if err := app.store.Delete(id); errors.Is(err, os.ErrNotExist) {
			writeJSON(writer, http.StatusNotFound, map[string]string{"error": "服务不存在"})
			return
		} else if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		_ = app.preferences.ClearServiceCategory(id)
		_ = app.preferences.ClearServiceDomains(id)
		writer.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(writer, http.MethodPut, http.MethodDelete)
	}
}

func (app *App) handleServiceDomains(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	serviceID := strings.TrimPrefix(request.URL.Path, "/enhancer/api/service-domains/")
	if serviceID == "" {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "无效的服务 ID"})
		return
	}
	service, ok := app.catalog.Service(request.Context(), serviceID)
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "服务不存在"})
		return
	}
	switch request.Method {
	case http.MethodPut:
		var payload struct {
			Domains []string `json:"domains"`
		}
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := app.preferences.SetServiceDomains(service.ID, payload.Domains); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	case http.MethodDelete:
		if err := app.preferences.ClearServiceDomains(service.ID); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	default:
		methodNotAllowed(writer, http.MethodPut, http.MethodDelete)
		return
	}
	updated, _ := app.catalog.Service(request.Context(), service.ID)
	writeJSON(writer, http.StatusOK, updated)
}

func (app *App) handleCategories(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		snapshot := app.catalog.Snapshot(request.Context(), false)
		writeJSON(writer, http.StatusOK, map[string]any{
			"categories":        snapshot.Categories,
			"custom_categories": app.preferences.Categories(),
		})
	case http.MethodPost:
		if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
			return
		}
		var payload struct {
			Name string `json:"name"`
		}
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		category, err := normalizeCategory(payload.Name)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if app.catalogHasCategory(request.Context(), category) {
			writeJSON(writer, http.StatusConflict, map[string]string{"error": "分类已存在"})
			return
		}
		category, err = app.preferences.AddCategory(category)
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusCreated, map[string]string{"name": category})
	case http.MethodDelete:
		if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
			return
		}
		var payload struct {
			Name string `json:"name"`
		}
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		category, err := normalizeCategory(payload.Name)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if !app.preferences.IsCustomCategory(category) {
			writeJSON(writer, http.StatusNotFound, map[string]string{"error": "自定义分类不存在"})
			return
		}
		for _, service := range app.catalog.Snapshot(request.Context(), false).Services {
			if service.Category == category {
				writeJSON(writer, http.StatusConflict, map[string]string{"error": "该分类仍有服务，请先移动这些服务"})
				return
			}
		}
		if err := app.preferences.DeleteCategory(category); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPost, http.MethodDelete)
	}
}

func (app *App) handleServiceCategory(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	serviceID := strings.TrimPrefix(request.URL.Path, "/enhancer/api/service-categories/")
	if serviceID == "" {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "无效的服务 ID"})
		return
	}
	service, ok := app.catalog.Service(request.Context(), serviceID)
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "服务不存在"})
		return
	}
	switch request.Method {
	case http.MethodPut:
		var payload struct {
			Category string `json:"category"`
		}
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		category, err := normalizeCategory(payload.Category)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		known := false
		for _, available := range app.catalog.Snapshot(request.Context(), false).Categories {
			if available == category {
				known = true
				break
			}
		}
		if !known {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "分类不存在，请先新建分类"})
			return
		}
		if err := app.preferences.SetServiceCategory(service.ID, category); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	case http.MethodDelete:
		if err := app.preferences.ClearServiceCategory(service.ID); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	default:
		methodNotAllowed(writer, http.MethodPut, http.MethodDelete)
		return
	}
	updated, _ := app.catalog.Service(request.Context(), service.ID)
	writeJSON(writer, http.StatusOK, updated)
}

func (app *App) catalogHasCategory(ctx context.Context, category string) bool {
	for _, available := range app.catalog.Snapshot(ctx, false).Categories {
		if available == category {
			return true
		}
	}
	return false
}

func (app *App) handleRuleSet(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	id := strings.TrimPrefix(request.URL.Path, "/enhancer/rules/")
	id = strings.TrimSuffix(id, ".list")
	service, ok := app.catalog.Service(request.Context(), id)
	if !ok {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("Cache-Control", "public, max-age=1800")
	_, _ = fmt.Fprintf(writer, "# %s | %s\n", service.Category, service.Name)
	for _, domain := range routingDomains(service.Domains) {
		_, _ = fmt.Fprintf(writer, "DOMAIN-SUFFIX,%s\n", domain)
	}
}

func (app *App) handleConnectivity(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	var payload ConnectivityRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), 90*time.Second)
	defer cancel()
	results, err := TestConnectivity(ctx, payload)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"results": results})
}

func (app *App) authorize(ctx context.Context, authorization string) bool {
	if authorization == "" {
		return false
	}
	checkURL := *app.upstream
	checkURL.Path = "/api/nodes"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL.String(), nil)
	if err != nil {
		return false
	}
	request.Header.Set("Authorization", authorization)
	response, err := app.client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode >= 200 && response.StatusCode < 300
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("JSON 格式错误: %w", err)
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func methodNotAllowed(writer http.ResponseWriter, methods ...string) {
	writer.Header().Set("Allow", strings.Join(methods, ", "))
	writeJSON(writer, http.StatusMethodNotAllowed, map[string]string{"error": "请求方法不受支持"})
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		setNoCacheHeaders(writer)
		next.ServeHTTP(writer, request)
	})
}

func setNoCacheHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	writer.Header().Set("Pragma", "no-cache")
	writer.Header().Set("Expires", "0")
	writer.Header().Set("X-Prism-UI-Version", uiVersion)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set("Referrer-Policy", "same-origin")
		writer.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; font-src 'self' data: https://cdn.jsdelivr.net; script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		next.ServeHTTP(writer, request)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(writer, request)
		if !strings.HasPrefix(request.URL.Path, "/api/sse") {
			log.Printf("%s %s %s", request.Method, request.URL.Path, time.Since(started).Round(time.Millisecond))
		}
	})
}
