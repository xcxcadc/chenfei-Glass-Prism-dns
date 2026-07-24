package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

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
	service, err := store.Upsert(Service{Name: "Custom", Domains: []string{"example.com", "cdn.example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), store)
	app, err := NewApp(upstream.URL, catalog, store, upstream.Client())
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
	app, _ := NewApp(upstream.URL, catalog, store, upstream.Client())
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
