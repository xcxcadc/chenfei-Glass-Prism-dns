package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIPConfigStoreTrafficLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ip-configs.json")
	store, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	config, err := store.Save(IPConfig{
		IP:        "203.0.113.10",
		DNSNodeID: "12",
		NodeName:  "IP 203 0 113 10",
		Smart:     true,
		Routes:    map[string]string{"netflix": "2"},
	}, "node-secret", map[string][]string{"2": {"198.51.100.20"}})
	if err != nil {
		t.Fatal(err)
	}
	if config.EnrollmentToken == "" {
		t.Fatal("expected enrollment token")
	}
	if config.ServiceAuditRequestedAt == nil {
		t.Fatal("new route configuration should request a service audit")
	}
	requested, err := store.RequestServiceAudit(config.ID)
	if err != nil || requested.ServiceAuditRequestedAt == nil {
		t.Fatalf("service audit request was not persisted: config=%+v err=%v", requested, err)
	}

	if _, err := store.UpdateClientReport(config.EnrollmentToken, 0, 0, ClientHealth{
		DNSReady: true, SystemDNSReady: true, RoutesReady: true, HealthyRoutes: 1, ExpectedRoutes: 1,
	}); err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateClientReport(config.EnrollmentToken, 50, 60, ClientHealth{
		DNSReady: true, SystemDNSReady: true, RoutesReady: true, HealthyRoutes: 1, ExpectedRoutes: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrafficRXBytes != 50 || updated.TrafficTXBytes != 60 {
		t.Fatalf("unexpected traffic after second report: %+v", updated)
	}
	if !updated.DNSReady || !updated.SystemDNSReady || !updated.RoutesReady || updated.HealthyRoutes != 1 || updated.HealthUpdatedAt == nil {
		t.Fatalf("client health was not persisted: %+v", updated)
	}
	audited, err := store.UpdateServiceAudit(config.EnrollmentToken, map[string]string{"netflix": "YES (Region: SG) [Via DNS]", "other": "YES"})
	if err != nil {
		t.Fatal(err)
	}
	if audited.ServiceResults["netflix"] == "" || audited.ServiceResults["other"] != "" || audited.ServiceAuditedAt == nil {
		t.Fatalf("service audit was not filtered and persisted: %+v", audited)
	}
	updated, err = store.UpdateTraffic(config.EnrollmentToken, 10, 20)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrafficRXBytes != 60 || updated.TrafficTXBytes != 80 {
		t.Fatalf("counter reset was not accumulated: %+v", updated)
	}
	cleared, err := store.ClearTraffic(config.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cleared.TrafficRXBytes != 0 || cleared.TrafficTXBytes != 0 {
		t.Fatalf("traffic was not cleared: %+v", cleared)
	}
	updated, err = store.UpdateTraffic(config.EnrollmentToken, 20, 30)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrafficRXBytes != 10 || updated.TrafficTXBytes != 10 {
		t.Fatalf("traffic after clear did not continue from the preserved counter baseline: %+v", updated)
	}
	updated, err = store.UpdateTraffic(config.EnrollmentToken, 30, 45)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrafficRXBytes != 20 || updated.TrafficTXBytes != 25 {
		t.Fatalf("traffic after the new baseline was not accumulated: %+v", updated)
	}

	reloaded, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := reloaded.Record(config.ID)
	if !ok || record.NodeSecret != "node-secret" || record.TrafficRXBytes != 20 || record.ProxyPeers["2"][0] != "198.51.100.20" || record.ServiceAuditRequestedAt == nil {
		t.Fatalf("persisted record mismatch: %+v", record)
	}
	bySecret, ok := reloaded.GetByNodeSecret("node-secret")
	if !ok || bySecret.ID != config.ID {
		t.Fatalf("node secret lookup failed: %+v", bySecret)
	}
}

func TestIPConfigStorePreservesDualStackProxyPeers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ip-configs.json")
	data, err := json.Marshal([]ipConfigRecord{{
		IPConfig: IPConfig{
			ID:              "ip-legacy",
			IP:              "203.0.113.30",
			DNSNodeID:       "dns-a",
			NodeName:        "Legacy",
			EnrollmentToken: "token",
			Routes:          map[string]string{"gemini": "proxy-a"},
			TrafficPeers:    []string{"198.51.100.10", "2001:db8::10"},
		},
		NodeSecret: "secret",
		ProxyPeers: map[string][]string{
			"proxy-a": {"2001:db8::10", "198.51.100.10"},
			"unused":  {"198.51.100.20"},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := store.Record("ip-legacy")
	if !ok {
		t.Fatal("legacy record was not loaded")
	}
	if len(record.ProxyPeers) != 1 || len(record.ProxyPeers["proxy-a"]) != 2 || record.ProxyPeers["proxy-a"][0] != "198.51.100.10" || record.ProxyPeers["proxy-a"][1] != "2001:db8::10" {
		t.Fatalf("dual-stack proxy peers were not preserved: %+v", record.ProxyPeers)
	}
	if len(record.TrafficPeers) != 2 || record.TrafficPeers[0] != "198.51.100.10" || record.TrafficPeers[1] != "2001:db8::10" {
		t.Fatalf("dual-stack traffic peers were not preserved: %+v", record.TrafficPeers)
	}

	persistedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var persisted []ipConfigRecord
	if err := json.Unmarshal(persistedData, &persisted); err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 1 || len(persisted[0].TrafficPeers) != 2 || persisted[0].TrafficPeers[1] != "2001:db8::10" {
		t.Fatalf("dual-stack peers were not persisted: %+v", persisted)
	}
}

func TestIPConfigStoreRemovesLegacyAutomaticFailoverState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ip-configs.json")
	legacy := `[{
		"id":"ip-legacy",
		"ip":"203.0.113.31",
		"dns_node_id":"dns-a",
		"node_name":"Legacy",
		"enrollment_token":"token",
		"routes":{"gemini":"proxy-a"},
		"node_secret":"secret",
		"failovers":{"gemini":{"status":"switched","from_proxy_id":"proxy-b","to_proxy_id":"proxy-a"}}
	}]`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewIPConfigStore(path); err != nil {
		t.Fatal(err)
	}
	persisted, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(persisted), `"failovers"`) {
		t.Fatalf("legacy automatic failover state was not removed: %s", persisted)
	}
}

func TestNormalizeRouteConflictsPreservesManagedSmartMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ip-configs.json")
	store, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	config, err := store.Save(IPConfig{
		IP:        "203.0.113.45",
		DNSNodeID: "dns-45",
		NodeName:  "Target",
		Smart:     true,
		Routes:    map[string]string{"service-a": "proxy-a"},
	}, "secret", map[string][]string{"proxy-a": {"198.51.100.45"}})
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := store.NormalizeRouteConflicts([]Service{{ID: "service-a", Domains: []string{"example.com"}}})
	if err != nil {
		t.Fatal(err)
	}
	if normalized != 0 {
		t.Fatalf("normalized count = %d, want 0", normalized)
	}
	updated, ok := store.Get(config.ID)
	if !ok || !updated.Smart {
		t.Fatalf("managed smart mode was not preserved: %+v", updated)
	}
	reloaded, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	persisted, ok := reloaded.Get(config.ID)
	if !ok || !persisted.Smart {
		t.Fatalf("managed smart mode was not persisted: %+v", persisted)
	}
}
