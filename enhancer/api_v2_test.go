package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHandleV2Dashboard(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/nodes" && r.Header.Get("Authorization") == "Bearer test-tok" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"p1","name":"SG","role":"proxy","public_ip":"203.0.113.1"},{"id":"d1","name":"Client","role":"dns","public_ip":"203.0.113.2"}]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()

	dir := t.TempDir()
	customStore, err := NewCustomServiceStore(filepath.Join(dir, "custom.json"))
	if err != nil {
		t.Fatal(err)
	}
	ipStore, _ := NewIPConfigStore(filepath.Join(dir, "ips.json"))
	app, err := NewApp(upstream.URL, NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), customStore), customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/enhancer/api/v2/dashboard", nil)
	req.Header.Set("Authorization", "Bearer test-tok")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["proxies_total"].(float64) != 1 || payload["clients_total"].(float64) != 1 {
		t.Fatalf("unexpected dashboard counts: %#v", payload)
	}
}
