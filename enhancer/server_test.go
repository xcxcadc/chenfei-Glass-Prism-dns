package main

import (
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
	service, err := store.Upsert(Service{Name: "Custom", Domains: []string{"example.com", "*.example.com", "*.cdn.example.com"}})
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
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), "/assets/app.js?v="+uiVersion) {
		t.Fatalf("index does not reference versioned UI assets: %s", response.Body.String())
	}
}
