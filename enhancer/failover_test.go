package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestAutomaticFailoverUsesHealthyIPv4ProxyAndRequestsReaudit(t *testing.T) {
	catalogServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("# ---------- > AI Platform\n# > Claude\nnameserver /claude.ai/group\n"))
	}))
	defer catalogServer.Close()

	customStore, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager(catalogServer.URL, catalogServer.Client(), customStore)
	snapshot := catalog.Snapshot(context.Background(), true)
	if len(snapshot.Services) != 1 {
		t.Fatalf("unexpected catalog: %+v", snapshot)
	}
	serviceID := snapshot.Services[0].ID

	databasePath := filepath.Join(t.TempDir(), "data.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TABLE nodes (
		id TEXT PRIMARY KEY,
		name TEXT,
		role TEXT,
		public_ip TEXT,
		address TEXT,
		priority INTEGER,
		latency INTEGER,
		last_heartbeat TEXT,
		unlock_json TEXT
	)`); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, values := range [][]any{
		{"sg", "Singapore", "proxy", "198.51.100.20, 2001:db8::20", "", 80, 30, now, `{"Claude":"Yes"}`},
		{"vn", "Vietnam", "proxy", "198.51.100.30, 2001:db8::30", "", 70, 20, now, `{"Claude":"YES (Region: VN)"}`},
		{"v6", "IPv6 only", "proxy", "2001:db8::40", "", 100, 1, now, `{"Claude":"Yes"}`},
	} {
		if _, err := database.Exec(`INSERT INTO nodes
			(id, name, role, public_ip, address, priority, latency, last_heartbeat, unlock_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, values...); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	ipStore, err := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	config, err := ipStore.Save(IPConfig{
		IP:        "203.0.113.10",
		DNSNodeID: "dns-1",
		NodeName:  "Target",
		Routes:    map[string]string{serviceID: "sg"},
	}, "node-secret", map[string][]string{"sg": {"198.51.100.20", "2001:db8::20"}})
	if err != nil {
		t.Fatal(err)
	}
	initialRequest := *config.ServiceAuditRequestedAt

	app, err := NewApp(catalogServer.URL, catalog, customStore, ipStore, catalogServer.Client(), databasePath)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := app.reconcileAutomaticFailover(config, map[string]string{
		serviceID: "UNSTABLE (2/3 YES; Banned (WAF))",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Routes[serviceID] != "vn" {
		t.Fatalf("expected the IPv4-capable VN proxy, got %+v", updated.Routes)
	}
	if len(updated.TrafficPeers) != 1 || updated.TrafficPeers[0] != "198.51.100.30" {
		t.Fatalf("automatic failover must publish only proxy IPv4 peers: %+v", updated.TrafficPeers)
	}
	event := updated.Failovers[serviceID]
	if event.Status != "switched" || event.FromProxyID != "sg" || event.ToProxyID != "vn" {
		t.Fatalf("failover event was not persisted: %+v", event)
	}
	if updated.ServiceAuditRequestedAt == nil || !updated.ServiceAuditRequestedAt.After(initialRequest) {
		t.Fatalf("failover did not request a fresh target audit: %+v", updated.ServiceAuditRequestedAt)
	}

	recovered, err := app.reconcileAutomaticFailover(updated, map[string]string{
		serviceID: "YES (Region: VN) [Via DNS]",
	})
	if err != nil {
		t.Fatal(err)
	}
	event = recovered.Failovers[serviceID]
	if event.Status != "recovered" || event.RecoveredResult == "" {
		t.Fatalf("successful re-audit did not mark failover recovered: %+v", event)
	}
}

func TestAuditResultPassedRejectsUnstableResults(t *testing.T) {
	if auditResultPassed("UNSTABLE (2/3 YES; Banned (WAF))") {
		t.Fatal("unstable audit results must trigger failover")
	}
	for _, result := range []string{"YES (Region: SG)", "PASS (HTTPS reachable)"} {
		if !auditResultPassed(result) {
			t.Fatalf("expected passing audit result: %s", result)
		}
	}
}

func TestParseControllerTimeAcceptsLongSQLiteFraction(t *testing.T) {
	parsed := parseControllerTime("2026-07-29 03:32:14.6489228052+00:00")
	if parsed.IsZero() || parsed.Nanosecond() != 648922805 {
		t.Fatalf("controller heartbeat was not normalized: %v", parsed)
	}
}

func TestAutomaticFailoverKeepsOverlappingServicesOnCompatibleProxy(t *testing.T) {
	catalogServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("# ---------- > AI Platform\n# > Claude\nnameserver /shared.example/group\n# > Gemini\nnameserver /shared.example/group\n"))
	}))
	defer catalogServer.Close()

	customStore, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager(catalogServer.URL, catalogServer.Client(), customStore)
	snapshot := catalog.Snapshot(context.Background(), true)
	if len(snapshot.Services) != 2 {
		t.Fatalf("unexpected catalog: %+v", snapshot.Services)
	}
	serviceIDs := make(map[string]string)
	for _, service := range snapshot.Services {
		serviceIDs[service.Name] = service.ID
	}

	databasePath := filepath.Join(t.TempDir(), "data.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TABLE nodes (
		id TEXT PRIMARY KEY,
		name TEXT,
		role TEXT,
		public_ip TEXT,
		address TEXT,
		priority INTEGER,
		latency INTEGER,
		last_heartbeat TEXT,
		unlock_json TEXT
	)`); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, values := range [][]any{
		{"current", "Current", "proxy", "198.51.100.10", "", 80, 30, now, `{"Claude":"YES","Gemini":"YES"}`},
		{"claude-only", "Claude only", "proxy", "198.51.100.20", "", 80, 10, now, `{"Claude":"YES","Gemini":"NO"}`},
		{"compatible", "Compatible", "proxy", "198.51.100.30", "", 70, 20, now, `{"Claude":"YES","Gemini":"YES"}`},
	} {
		if _, err := database.Exec(`INSERT INTO nodes
			(id, name, role, public_ip, address, priority, latency, last_heartbeat, unlock_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, values...); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	ipStore, err := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	routes := map[string]string{
		serviceIDs["Claude"]: "current",
		serviceIDs["Gemini"]: "current",
	}
	config, err := ipStore.Save(IPConfig{
		IP:             "203.0.113.20",
		DNSNodeID:      "dns-2",
		NodeName:       "Target",
		Routes:         routes,
		ServiceResults: map[string]string{serviceIDs["Claude"]: "UNSTABLE", serviceIDs["Gemini"]: "YES"},
	}, "node-secret", map[string][]string{"current": {"198.51.100.10"}})
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(catalogServer.URL, catalog, customStore, ipStore, catalogServer.Client(), databasePath)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := app.reconcileAutomaticFailover(config, map[string]string{
		serviceIDs["Claude"]: "UNSTABLE (1/3 YES; TIMEOUT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, serviceID := range serviceIDs {
		if updated.Routes[serviceID] != "compatible" {
			t.Fatalf("%s route did not follow the compatible shared proxy: %+v", name, updated.Routes)
		}
		if event := updated.Failovers[serviceID]; event.Status != "switched" || event.ToProxyID != "compatible" {
			t.Fatalf("%s failover event was not linked: %+v", name, event)
		}
		if _, exists := updated.ServiceResults[serviceID]; exists {
			t.Fatalf("%s stale audit result survived failover: %+v", name, updated.ServiceResults)
		}
	}
	if len(updated.TrafficPeers) != 1 || updated.TrafficPeers[0] != "198.51.100.30" {
		t.Fatalf("linked failover did not publish the compatible IPv4 peer: %+v", updated.TrafficPeers)
	}
}
