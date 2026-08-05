package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestRuleSetHandler(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/nodes" && request.Header.Get("Authorization") == "Bearer test" {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[]`))
			return
		}
		http.NotFound(writer, request)
	}))
	defer upstream.Close()

	store, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	service, err := store.Upsert(Service{Name: "Custom", Domains: []string{"example.com", "*.example.com", "*.cdn.example.com"}, DomainKeywords: []string{"Twitter"}, CIDRs: []string{"192.133.76.0/22"}})
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, err := NewApp(upstream.URL, catalog, store, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/enhancer/rules/"+service.ID+".list", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), "DOMAIN-SUFFIX,example.com") {
		t.Fatalf("unexpected ruleset: %s", body)
	}
	if !strings.Contains(string(body), "DOMAIN-SUFFIX,cdn.example.com") || strings.Contains(string(body), "DOMAIN-SUFFIX,*.") {
		t.Fatalf("wildcard ruleset was not compiled safely: %s", body)
	}
	if strings.Count(string(body), "DOMAIN-SUFFIX,example.com") != 1 {
		t.Fatalf("duplicate compiled suffix rule: %s", body)
	}
	if !strings.Contains(string(body), "DOMAIN-KEYWORD,twitter") || !strings.Contains(string(body), "IP-CIDR,192.133.76.0/22") {
		t.Fatalf("keyword or CIDR rules were not compiled: %s", body)
	}
}

func TestCustomServiceWriteRequiresAuthentication(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") == "Bearer valid" {
			_, _ = writer.Write([]byte(`[]`))
			return
		}
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	store, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp(upstream.URL, catalog, store, ipStore, upstream.Client())
	body := `{"name":"Private","domains":["example.com"]}`

	unauthorized := httptest.NewRequest(http.MethodPost, "/enhancer/api/custom-services", strings.NewReader(body))
	unauthorizedResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", unauthorizedResponse.Code)
	}

	authorized := httptest.NewRequest(http.MethodPost, "/enhancer/api/custom-services", strings.NewReader(body))
	authorized.Header.Set("Authorization", "Bearer valid")
	authorizedResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(authorizedResponse, authorized)
	if authorizedResponse.Code != http.StatusCreated {
		t.Fatalf("expected created, got %d: %s", authorizedResponse.Code, authorizedResponse.Body.String())
	}
}

func TestCustomServiceDeleteRequiresAuthenticationAndPersists(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") == "Bearer valid" {
			_, _ = writer.Write([]byte(`[]`))
			return
		}
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	path := filepath.Join(t.TempDir(), "services.json")
	store, err := NewCustomServiceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	service, err := store.Upsert(Service{Name: "Delete Test", Domains: []string{"delete.example"}})
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, err := NewApp(upstream.URL, catalog, store, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	requestPath := "/enhancer/api/custom-services/" + service.ID

	request := httptest.NewRequest(http.MethodDelete, requestPath, nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated custom service delete should fail, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, requestPath, nil)
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("authenticated custom service delete failed: %d %s", response.Code, response.Body.String())
	}
	if _, ok := store.Get(service.ID); ok {
		t.Fatal("deleted custom service remained in memory")
	}

	reloaded, err := NewCustomServiceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reloaded.Get(service.ID); ok {
		t.Fatal("deleted custom service remained on disk")
	}
}

func TestServiceDomainWriteRequiresAuthenticationAndCanRestore(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/nodes" && request.Header.Get("Authorization") == "Bearer valid" {
			_, _ = writer.Write([]byte(`[]`))
			return
		}
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	store, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	service, err := store.Upsert(Service{Name: "Domain Test", Domains: []string{"original.example"}})
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, err := NewApp(upstream.URL, catalog, store, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	path := "/enhancer/api/service-domains/" + service.ID

	request := httptest.NewRequest(http.MethodPut, path, strings.NewReader(`{"domains":["*.custom.example","custom.example"]}`))
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated domain write should fail, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPut, path, strings.NewReader(`{"domains":["*.custom.example","custom.example"]}`))
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"domain_override":true`) {
		t.Fatalf("domain override write failed: %d %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPut, path, strings.NewReader(`{"domains":["https://invalid.example"]}`))
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid domain should fail, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, path, nil)
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `"domain_override":true`) || !strings.Contains(response.Body.String(), `"original.example"`) {
		t.Fatalf("domain override restore failed: %d %s", response.Code, response.Body.String())
	}
}

func TestBuiltInServiceDeleteRequiresAuthenticationAndCleansRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/nodes" && request.Header.Get("Authorization") == "Bearer valid" {
			_, _ = writer.Write([]byte(`[]`))
			return
		}
		if request.URL.Path == "/catalog" {
			_, _ = writer.Write([]byte("# ---------- > Global\n# > YouTube\nnameserver /youtube.com/group\n"))
			return
		}
		http.NotFound(writer, request)
	}))
	defer upstream.Close()

	dataDir := t.TempDir()
	customStore, err := NewCustomServiceStore(filepath.Join(dataDir, "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	preferences, err := NewCatalogPreferenceStore(filepath.Join(dataDir, "catalog-preferences.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore, preferences)
	ipStore, err := NewIPConfigStore(filepath.Join(dataDir, "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	services := catalog.Snapshot(context.Background(), true).Services
	if len(services) != 1 {
		t.Fatalf("unexpected catalog: %+v", services)
	}
	service := services[0]
	config, err := ipStore.Save(IPConfig{
		IP:        "203.0.113.41",
		DNSNodeID: "dns-a",
		NodeName:  "Target",
		Routes:    map[string]string{service.ID: "proxy-a"},
	}, "secret", map[string][]string{"proxy-a": {"198.51.100.12"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ipStore.UpdateServiceAudit(config.EnrollmentToken, map[string]string{service.ID: "YES"}); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodDelete, "/enhancer/api/services/"+service.ID, nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated service delete should fail, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/enhancer/api/services/"+service.ID, nil)
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("built-in service delete failed: %d %s", response.Code, response.Body.String())
	}
	if services := catalog.Snapshot(context.Background(), true).Services; len(services) != 0 {
		t.Fatalf("deleted service returned after catalog refresh: %+v", services)
	}
	updated, ok := ipStore.Get(config.ID)
	record, recordOK := ipStore.Record(config.ID)
	if !ok || !recordOK || len(updated.Routes) != 0 || len(record.ProxyPeers) != 0 || len(updated.ServiceResults) != 0 {
		t.Fatalf("deleted service route data remained: %+v", updated)
	}
}

func TestServiceIconHandlerCachesFetchedIcon(t *testing.T) {
	store, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	service, _ := store.Upsert(Service{Name: "Example", Domains: []string{"example.com"}})
	requests := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"image/png"}},
			Body:       io.NopCloser(strings.NewReader("png")),
			Request:    request,
		}, nil
	})}
	catalog := NewCatalogManager("https://catalog.invalid/list", client, store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp("http://127.0.0.1:1", catalog, store, ipStore, client)

	for range 2 {
		request := httptest.NewRequest(http.MethodGet, "/enhancer/icons/"+service.ID+".png", nil)
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/png" || response.Body.String() != "png" {
			t.Fatalf("unexpected icon response: %d %q %q", response.Code, response.Header().Get("Content-Type"), response.Body.String())
		}
		if !strings.Contains(response.Header().Get("Cache-Control"), "immutable") {
			t.Fatalf("fetched icon should be immutable: %q", response.Header().Get("Cache-Control"))
		}
		if response.Header().Get("X-Prism-Icon-Source") != "remote" {
			t.Fatalf("fetched icon source is missing: %q", response.Header().Get("X-Prism-Icon-Source"))
		}
	}
	if requests != 1 {
		t.Fatalf("icon should be fetched once, got %d requests", requests)
	}
}

func TestServiceIconHandlerFallsBackToSVG(t *testing.T) {
	store, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	service, _ := store.Upsert(Service{Name: "Fallback", Domains: []string{"fallback.example"}})
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})}
	catalog := NewCatalogManager("https://catalog.invalid/list", client, store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp("http://127.0.0.1:1", catalog, store, ipStore, client)

	request := httptest.NewRequest(http.MethodGet, "/enhancer/icons/"+service.ID+".png", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.HasPrefix(response.Header().Get("Content-Type"), "image/svg+xml") || !strings.Contains(response.Body.String(), "<svg") {
		t.Fatalf("unexpected fallback icon: %d %q %q", response.Code, response.Header().Get("Content-Type"), response.Body.String())
	}
	if !strings.Contains(response.Header().Get("Cache-Control"), "no-store") {
		t.Fatalf("fallback icon must not be cached by browsers: %q", response.Header().Get("Cache-Control"))
	}
	if response.Header().Get("X-Prism-Icon-Source") != "fallback" {
		t.Fatalf("fallback icon source is missing: %q", response.Header().Get("X-Prism-Icon-Source"))
	}
}

func TestServiceIconHandlerCanSkipFallbackPlaceholder(t *testing.T) {
	store, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	service, _ := store.Upsert(Service{Name: "Fallback Skip", Domains: []string{"fallback-skip.example"}})
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})}
	catalog := NewCatalogManager("https://catalog.invalid/list", client, store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp("http://127.0.0.1:1", catalog, store, ipStore, client)

	request := httptest.NewRequest(http.MethodGet, "/enhancer/icons/"+service.ID+".png?fallback=0", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected placeholder-skipping request to return 404, got %d", response.Code)
	}
}

func TestServiceIconHandlerDoesNotFetchCustomDomainsDirectly(t *testing.T) {
	store, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	service, _ := store.Upsert(Service{Name: "Internal", Domains: []string{"internal.example"}})
	var requestedHosts []string
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestedHosts = append(requestedHosts, request.URL.Hostname())
		return nil, errors.New("offline")
	})}
	catalog := NewCatalogManager("https://catalog.invalid/list", client, store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp("http://127.0.0.1:1", catalog, store, ipStore, client)

	request := httptest.NewRequest(http.MethodGet, "/enhancer/icons/"+service.ID+".png", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected icon response: %d", response.Code)
	}
	if len(requestedHosts) == 0 {
		t.Fatal("expected the favicon proxy to be queried")
	}
	for _, host := range requestedHosts {
		if host != "www.google.com" && host != "icons.duckduckgo.com" {
			t.Fatalf("custom service triggered a direct request to %q", host)
		}
	}
}

func TestServiceIconHandlerReusesPersistentCache(t *testing.T) {
	dataDir := t.TempDir()
	store, _ := NewCustomServiceStore(filepath.Join(dataDir, "services.json"))
	service, _ := store.Upsert(Service{Name: "Example", Domains: []string{"example.com"}})
	requests := 0
	fetchingClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"image/png"}},
			Body:       io.NopCloser(strings.NewReader("persisted-png")),
			Request:    request,
		}, nil
	})}
	catalog := NewCatalogManager("https://catalog.invalid/list", fetchingClient, store)
	ipStore, _ := NewIPConfigStore(filepath.Join(dataDir, "ip-configs.json"))
	first, _ := NewApp("http://127.0.0.1:1", catalog, store, ipStore, fetchingClient)
	first.icons = newServiceIconCache(filepath.Join(dataDir, "icon-cache"))

	path := "/enhancer/icons/" + service.ID + ".png"
	response := httptest.NewRecorder()
	first.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != http.StatusOK || response.Body.String() != "persisted-png" {
		t.Fatalf("unexpected first icon response: %d %q", response.Code, response.Body.String())
	}

	offlineClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})}
	restartedCatalog := NewCatalogManager("https://catalog.invalid/list", offlineClient, store)
	restarted, _ := NewApp("http://127.0.0.1:1", restartedCatalog, store, ipStore, offlineClient)
	restarted.icons = newServiceIconCache(filepath.Join(dataDir, "icon-cache"))
	response = httptest.NewRecorder()
	restarted.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != http.StatusOK || response.Body.String() != "persisted-png" {
		t.Fatalf("persistent icon was not reused: %d %q", response.Code, response.Body.String())
	}
	if requests != 1 {
		t.Fatalf("expected one remote icon request across restart, got %d", requests)
	}
}

func TestWebAssetsExposeVersionAndDisableCaching(t *testing.T) {
	store, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", http.DefaultClient, store)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp("http://127.0.0.1:1", catalog, store, ipStore, http.DefaultClient)

	for _, path := range []string{"/", "/assets/app.js?v=" + uiVersion} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, response.Code)
		}
		if cacheControl := response.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "no-store") || !strings.Contains(cacheControl, "no-cache") {
			t.Fatalf("%s has weak cache control: %q", path, cacheControl)
		}
		if response.Header().Get("Pragma") != "no-cache" || response.Header().Get("Expires") != "0" {
			t.Fatalf("%s is missing legacy no-cache headers", path)
		}
		if response.Header().Get("X-Prism-UI-Version") != uiVersion {
			t.Fatalf("%s is missing UI version header", path)
		}
		if path != "/" && (!strings.Contains(response.Body.String(), "service-custom-edit") || !strings.Contains(response.Body.String(), "service-custom-delete")) {
			t.Fatalf("%s is missing direct custom service management controls", path)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), "/assets/app.js?v="+uiVersion) {
		t.Fatalf("index does not reference versioned UI assets: %s", response.Body.String())
	}
	if csp := response.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "img-src 'self' data: https:") {
		t.Fatalf("CSP blocks external service icon fallbacks: %q", csp)
	}

	request = httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("favicon returned %d", response.Code)
	}
	if response.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("favicon has unexpected content type: %q", response.Header().Get("Content-Type"))
	}
	if !bytes.HasPrefix(response.Body.Bytes(), []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("favicon is not a PNG")
	}
}
