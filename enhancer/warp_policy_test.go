package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestWarpPolicyStorePersistenceAndIsolation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "warp-policies.json")

	store, err := NewWarpPolicyStore(path)
	if err != nil {
		t.Fatalf("NewWarpPolicyStore failed: %v", err)
	}
	if _, err := store.Set("proxy-sg", "gemini", "warp"); err != nil {
		t.Fatalf("Set gemini failed: %v", err)
	}
	if _, err := store.Set("proxy-sg", "netflix", "warp6"); err != nil {
		t.Fatalf("Set netflix failed: %v", err)
	}
	if _, err := store.Set("proxy-sg", "claude", "direct"); err != nil {
		t.Fatalf("Set direct should delete entry: %v", err)
	}
	if got := store.Mode("proxy-sg", "gemini"); got != "warp" {
		t.Fatalf("expected gemini=warp, got %q", got)
	}
	if got := store.Mode("proxy-sg", "netflix"); got != "warp6" {
		t.Fatalf("expected netflix=warp6, got %q", got)
	}
	if got := store.Mode("proxy-sg", "claude"); got != "" {
		t.Fatalf("expected claude empty (direct), got %q", got)
	}
	if got := store.Mode("proxy-jp", "gemini"); got != "" {
		t.Fatalf("other proxy must remain isolated, got %q", got)
	}

	reloaded, err := NewWarpPolicyStore(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.Mode("proxy-sg", "gemini") != "warp" || reloaded.Mode("proxy-sg", "netflix") != "warp6" {
		t.Fatalf("reloaded store mismatch: %#v", reloaded.Snapshot())
	}

	if err := reloaded.DeleteProxy("proxy-sg"); err != nil {
		t.Fatalf("DeleteProxy failed: %v", err)
	}
	if len(reloaded.ForProxy("proxy-sg")) != 0 {
		t.Fatalf("expected proxy-sg cleared after DeleteProxy")
	}
}

func TestProxyEgressServicesIncludesArbitraryWarpService(t *testing.T) {
	dir := t.TempDir()
	warpStore, err := NewWarpPolicyStore(filepath.Join(dir, "warp.json"))
	if err != nil {
		t.Fatal(err)
	}
	ipStore, err := NewIPConfigStore(filepath.Join(dir, "ips.json"))
	if err != nil {
		t.Fatal(err)
	}
	customStore, err := NewCustomServiceStore(filepath.Join(dir, "custom.json"))
	if err != nil {
		t.Fatal(err)
	}
	geminiSvc, err := customStore.Upsert(Service{Name: "Gemini", Category: "AI", Domains: []string{"gemini.google.com"}})
	if err != nil {
		t.Fatal(err)
	}
	netflixSvc, err := customStore.Upsert(Service{Name: "Netflix", Category: "Streaming", Domains: []string{"netflix.com"}})
	if err != nil {
		t.Fatal(err)
	}
	disneySvc, err := customStore.Upsert(Service{Name: "Disney+", Category: "Streaming", Domains: []string{"disneyplus.com"}})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := warpStore.Set("p1", netflixSvc.ID, "warp6"); err != nil {
		t.Fatal(err)
	}
	if _, err := warpStore.Set("p1", geminiSvc.ID, "warp"); err != nil {
		t.Fatal(err)
	}

	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", &http.Client{}, customStore)
	app := &App{
		catalog:      catalog,
		ipStore:      ipStore,
		warpPolicies: warpStore,
	}

	services := app.proxyEgressServices(context.Background(), "p1")
	byID := map[string]transportEgressService{}
	for _, svc := range services {
		byID[svc.ServiceID] = svc
	}
	if byID[geminiSvc.ID].EgressMode != "warp" {
		t.Fatalf("expected gemini egress_mode=warp, got %#v", byID[geminiSvc.ID])
	}
	if byID[netflixSvc.ID].EgressMode != "warp6" {
		t.Fatalf("expected non-IPv6-candidate netflix to be exported with egress_mode=warp6, got %#v", byID[netflixSvc.ID])
	}
	if _, exists := byID[disneySvc.ID]; exists {
		t.Fatalf("unselected non-IPv6-candidate service must stay native direct and not be exported")
	}
}

func TestHandleWarpPoliciesAPI(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/nodes" && r.Header.Get("Authorization") == "Bearer test-token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()

	dir := t.TempDir()
	warpStore, err := NewWarpPolicyStore(filepath.Join(dir, "warp.json"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(upstream.URL, NewCatalogManager("", nil, nil), nil, nil, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	app.warpPolicies = warpStore

	body, _ := json.Marshal(map[string]any{
		"proxy_id":   "proxy-1",
		"service_id": "youtube",
		"mode":       "warp",
	})
	req := httptest.NewRequest(http.MethodPut, "/enhancer/api/warp-policies", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	app.handleWarpPolicies(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT failed: %d %s", rec.Code, rec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/enhancer/api/warp-policies?proxy_id=proxy-1", nil)
	getReq.Header.Set("Authorization", "Bearer test-token")
	getRec := httptest.NewRecorder()
	app.handleWarpPolicies(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET failed: %d %s", getRec.Code, getRec.Body.String())
	}
	var resp struct {
		Policies map[string]map[string]string `json:"policies"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Policies["proxy-1"]["youtube"] != "warp" {
		t.Fatalf("expected youtube=warp, got %#v", resp.Policies)
	}
}
