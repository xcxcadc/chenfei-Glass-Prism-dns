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
	}, "node-secret")
	if err != nil {
		t.Fatal(err)
	}
	if config.EnrollmentToken == "" {
		t.Fatal("expected enrollment token")
	}

	if _, err := store.UpdateTraffic(config.EnrollmentToken, 0, 0); err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateTraffic(config.EnrollmentToken, 50, 60)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrafficRXBytes != 50 || updated.TrafficTXBytes != 60 {
		t.Fatalf("unexpected traffic after second report: %+v", updated)
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
		t.Fatalf("traffic after clear used the wrong baseline: %+v", updated)
	}

	reloaded, err := NewIPConfigStore(path)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := reloaded.Record(config.ID)
	if !ok || record.NodeSecret != "node-secret" || record.TrafficRXBytes != 10 {
		t.Fatalf("persisted record mismatch: %+v", record)
	}
}
