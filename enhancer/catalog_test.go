package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSmartDNS(t *testing.T) {
	input := `# ---------- > Global Platform
# > Disney+
nameserver /disneyplus.com/group
nameserver /*.bamgrid.com/group
nameserver /disneyplus.com/group
# > Netflix
nameserver /netflix.com/group
invalid line
`
	services, err := ParseSmartDNS(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseSmartDNS returned error: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
	if services[0].Name != "Disney+" || len(services[0].Domains) != 2 {
		t.Fatalf("unexpected first service: %#v", services[0])
	}
	if services[0].Domains[0] != "bamgrid.com" || services[0].Domains[1] != "disneyplus.com" {
		t.Fatalf("unexpected normalized domains: %#v", services[0].Domains)
	}
}

func TestCustomServiceStorePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "services.json")
	store, err := NewCustomServiceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := store.Upsert(Service{Name: "My Service", Category: "Custom", Domains: []string{"Example.com", "*.cdn.example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if !saved.Custom || len(saved.Domains) != 2 {
		t.Fatalf("unexpected saved service: %#v", saved)
	}
	reloaded, err := NewCustomServiceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	service, ok := reloaded.Get(saved.ID)
	if !ok || service.Name != "My Service" {
		t.Fatalf("service was not persisted: %#v", service)
	}
	if err := reloaded.Delete(saved.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("store file missing after delete: %v", err)
	}
}

func TestNormalizeDomainsRejectsInvalidValues(t *testing.T) {
	domains := normalizeDomains([]string{"ok.example", "bad value", "https://example.com", "ok.example"})
	if len(domains) != 1 || domains[0] != "ok.example" {
		t.Fatalf("unexpected domains: %#v", domains)
	}
}

func TestParseSmartDNSCanonicalizesLabelsWithoutChangingIDs(t *testing.T) {
	input := `# ---------- > Global Plaform
# > Claude 2
nameserver /claude.ai/group
# > Openai
nameserver /openai.com/group
`
	services, err := ParseSmartDNS(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if services[0].Name != "Claude" || services[0].Category != "Global Platform" {
		t.Fatalf("Claude label was not normalized: %+v", services[0])
	}
	if services[0].ID != stableServiceID("Global Plaform", "Claude 2") {
		t.Fatalf("existing service ID changed: %s", services[0].ID)
	}
	if services[1].Name != "ChatGPT / OpenAI" {
		t.Fatalf("OpenAI label was not normalized: %+v", services[1])
	}
}
