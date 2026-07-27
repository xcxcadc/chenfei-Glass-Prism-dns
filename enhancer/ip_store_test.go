package main

import (
	"path/filepath"
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
	if updated.TrafficRXBytes != 0 || updated.TrafficTXBytes != 0 {
		t.Fatalf("first traffic report after clear should only establish a baseline: %+v", updated)
	}
	updated, err = store.UpdateTraffic(config.EnrollmentToken, 30, 45)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrafficRXBytes != 10 || updated.TrafficTXBytes != 15 {
		t.Fatalf("traffic after the new baseline was not accumulated: %+v", updated)
	}

	reloaded, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := reloaded.Record(config.ID)
	if !ok || record.NodeSecret != "node-secret" || record.TrafficRXBytes != 10 || record.ProxyPeers["2"][0] != "198.51.100.20" {
		t.Fatalf("persisted record mismatch: %+v", record)
	}
	bySecret, ok := reloaded.GetByNodeSecret("node-secret")
	if !ok || bySecret.ID != config.ID {
		t.Fatalf("node secret lookup failed: %+v", bySecret)
	}
}
