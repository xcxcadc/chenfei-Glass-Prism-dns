package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestAgentSyncRestoresPersistentIPRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"role":"dns",
			"rules":{"https://panel/enhancer/rules/stale.list:stale.example":{"name":"https://panel/enhancer/rules/stale.list","pattern":"stale.example","ips":["192.0.2.1"]},"public:example.net":{"name":"public","pattern":"example.net","ips":["192.0.2.2"]}},
			"rule_overrides":{"stale.example":"old-proxy","example.net":"public-proxy"}
		}`))
	}))
	defer upstream.Close()

	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	service, err := customStore.Upsert(Service{Name: "Persistent", Domains: []string{"example.com", "cdn.example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	_, err = ipStore.Save(IPConfig{
		IP: "203.0.113.10", DNSNodeID: "dns-1", NodeName: "Target", Smart: true,
		Routes: map[string]string{service.ID: "proxy-1"}, TrafficPeers: []string{"198.51.100.20"},
	}, "node-secret", map[string][]string{"proxy-1": {"198.51.100.20"}})
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), customStore)
	app, _ := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())

	request := httptest.NewRequest(http.MethodGet, "/api/sync", nil)
	request.Header.Set("Authorization", "Bearer node-secret")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	var payload struct {
		Rules         map[string]map[string]any `json:"rules"`
		RuleOverrides map[string]string         `json:"rule_overrides"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.RuleOverrides["example.com"] != "proxy-1" || payload.RuleOverrides["cdn.example.com"] != "proxy-1" {
		t.Fatalf("persistent overrides were not restored: %+v", payload.RuleOverrides)
	}
	if payload.RuleOverrides["stale.example"] != "old-proxy" || payload.RuleOverrides["example.net"] != "public-proxy" {
		t.Fatalf("unmanaged overrides changed: %+v", payload.RuleOverrides)
	}
	rule := payload.Rules["enhancer:"+service.ID+":example.com"]
	if rule == nil || rule["pattern"] != "example.com" {
		t.Fatalf("persistent rule missing: %+v", payload.Rules)
	}
	if _, exists := payload.Rules["https://panel/enhancer/rules/stale.list:stale.example"]; exists {
		t.Fatal("stale enhancer rule was not removed")
	}
	if _, exists := payload.Rules["public:example.net"]; !exists {
		t.Fatal("unmanaged rule was removed")
	}
}

func TestAgentSyncIgnoresUnknownNodeSecret(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"role":"dns","rules":{},"rule_overrides":{}}`))
	}))
	defer upstream.Close()
	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	catalog := NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), customStore)
	app, _ := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	request := httptest.NewRequest(http.MethodGet, "/api/sync", nil)
	request.Header.Set("Authorization", "Bearer unknown")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != `{"role":"dns","rules":{},"rule_overrides":{}}` {
		t.Fatalf("unknown agent response changed: %d %s", response.Code, response.Body.String())
	}
}

func TestEffectiveProxyPeersMigratesSingleProxyConfig(t *testing.T) {
	record := ipConfigRecord{IPConfig: IPConfig{
		Routes:       map[string]string{"service-a": "proxy-1", "service-b": "proxy-1"},
		TrafficPeers: []string{"198.51.100.20", "2001:db8::20"},
	}}
	peers := effectiveProxyPeers(record)
	if len(peers) != 1 || len(peers["proxy-1"]) != 2 {
		t.Fatalf("legacy single-proxy config was not migrated: %+v", peers)
	}
	record.Routes["service-b"] = "proxy-2"
	if peers := effectiveProxyPeers(record); peers != nil {
		t.Fatalf("ambiguous legacy multi-proxy config should not be guessed: %+v", peers)
	}
}
