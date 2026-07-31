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
	if saved.Domains[0] != "*.cdn.example.com" || saved.Domains[1] != "example.com" {
		t.Fatalf("wildcard domain was not preserved: %#v", saved.Domains)
	}
	reloaded, err := NewCustomServiceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	service, ok := reloaded.Get(saved.ID)
	if !ok || service.Name != "My Service" {
		t.Fatalf("service was not persisted: %#v", service)
	}
	if service.Domains[0] != "*.cdn.example.com" || service.Domains[1] != "example.com" {
		t.Fatalf("reloaded wildcard domain changed: %#v", service.Domains)
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

func TestNormalizeCustomDomainsPreservesWildcards(t *testing.T) {
	domains, err := normalizeCustomDomains([]string{"Example.COM.", "*.API.Example.COM.", "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"*.api.example.com", "example.com"}
	if len(domains) != len(expected) {
		t.Fatalf("unexpected domains: %#v", domains)
	}
	for index := range expected {
		if domains[index] != expected[index] {
			t.Fatalf("unexpected domains: %#v", domains)
		}
	}
	routes := routingDomains(domains)
	if len(routes) != 2 || routes[0] != "api.example.com" || routes[1] != "example.com" {
		t.Fatalf("unexpected routing domains: %#v", routes)
	}
}

func TestNormalizeCustomDomainsRejectsMalformedWildcards(t *testing.T) {
	for _, value := range []string{"*", "*example.com", "foo.*.example.com", "**.example.com", "https://example.com", "-bad.example"} {
		if _, err := normalizeCustomDomains([]string{value}); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestCustomServiceStoreRollsBackFailedWrites(t *testing.T) {
	root := t.TempDir()
	store, err := NewCustomServiceStore(filepath.Join(root, "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	saved, err := store.Upsert(Service{Name: "Original", Domains: []string{"example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	store.path = filepath.Join(blocked, "services.json")
	if _, err := store.Upsert(Service{ID: saved.ID, Name: "Changed", Domains: []string{"changed.example"}}); err == nil {
		t.Fatal("expected update persistence failure")
	}
	current, ok := store.Get(saved.ID)
	if !ok || current.Name != "Original" || current.Domains[0] != "example.com" {
		t.Fatalf("failed update was not rolled back: %#v", current)
	}
	if err := store.Delete(saved.ID); err == nil {
		t.Fatal("expected delete persistence failure")
	}
	if _, ok := store.Get(saved.ID); !ok {
		t.Fatal("failed delete was not rolled back")
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

func TestParseSmartDNSSupplementsYouTubeTrafficDomains(t *testing.T) {
	input := `# ---------- > China Media
# > YouTube
nameserver /youtube.com/group
nameserver /youtubei.googleapis.com/group
`
	services, err := ParseSmartDNS(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 {
		t.Fatalf("expected one service, got %d", len(services))
	}
	if services[0].Category != "Global Platform" {
		t.Fatalf("YouTube should be categorized as a global service: %+v", services[0])
	}
	for _, domain := range []string{"googlevideo.com", "ggpht.com", "ytimg.com", "youtube.com", "youtubei.googleapis.com"} {
		if !contains(services[0].Domains, domain) {
			t.Fatalf("YouTube traffic domain %q missing from %#v", domain, services[0].Domains)
		}
	}
}

func TestParseSmartDNSSupplementsGeminiApplicationDependencies(t *testing.T) {
	input := `# ---------- > AI Platform
# > Google Gemini
nameserver /gemini.google.com/group
nameserver /proactivebackend-pa.googleapis.com/group
`
	services, err := ParseSmartDNS(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 {
		t.Fatalf("expected one service, got %d", len(services))
	}
	for _, domain := range []string{
		"accountcapabilities-pa.googleapis.com",
		"accounts.google.com",
		"firebaseinstallations.googleapis.com",
		"gemini.gstatic.com",
		"lh3.googleusercontent.com",
		"ogads-pa.clients6.google.com",
		"oauthaccountmanager.googleapis.com",
		"people-pa.googleapis.com",
		"play.googleapis.com",
		"signaler-pa.googleapis.com",
		"subscriptionsfirstparty-pa.googleapis.com",
		"waa-pa.clients6.google.com",
		"www.google.com",
		"www.gstatic.com",
	} {
		if !contains(services[0].Domains, domain) {
			t.Fatalf("Gemini application dependency %q missing from %#v", domain, services[0].Domains)
		}
	}
}

func TestParseSmartDNSSupplementsCurrentViuDomains(t *testing.T) {
	input := `# ---------- > Global Platform
# > Viu.TV
nameserver /viu.now.com/group
`
	services, err := ParseSmartDNS(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 {
		t.Fatalf("expected one service, got %d", len(services))
	}
	for _, domain := range []string{"viu.com", "viu.tv"} {
		if !contains(services[0].Domains, domain) {
			t.Fatalf("Viu traffic domain %q missing from %#v", domain, services[0].Domains)
		}
	}
}

func TestParseSmartDNSRefreshesKnownStaleDomains(t *testing.T) {
	input := strings.NewReader(`# ---------- > Europe Media
# > FR:France.tv
nameserver /ftven.fr/group
# ---------- > Hong Kong Media
# > HOY TV
nameserver /hoy.tv/group
# ---------- > Japan Media
# > J:com On Demand
nameserver /id.zaq.ne.jp/group
# ---------- > Southeast Asia Media
# > VN:Galaxy Play
nameserver /glxplay.io/group
# > VN:K+
nameserver /solocoo.tv/group
`)
	services, err := ParseSmartDNS(input)
	if err != nil {
		t.Fatalf("ParseSmartDNS returned error: %v", err)
	}
	expected := map[string][]string{
		"FR:France.tv":    {"france.tv"},
		"HOY TV":          {"hoy.tv"},
		"J:com On Demand": {"jcom.co.jp", "myjcom.jp"},
		"VN:Galaxy Play":  {"galaxyplay.vn"},
		"VN:K+":           {"k-plus.tv", "kplus.vn"},
	}
	for _, service := range services {
		domains, ok := expected[service.Name]
		if !ok {
			continue
		}
		for _, domain := range domains {
			if !contains(service.Domains, domain) {
				t.Fatalf("%s missing refreshed domain %q: %#v", service.Name, domain, service.Domains)
			}
		}
		delete(expected, service.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing refreshed services: %#v", expected)
	}
}

func TestParseSmartDNSOmitsRetiredServices(t *testing.T) {
	input := strings.NewReader(`# ---------- > North America Media
# > Crackle
nameserver /crackle.com/group
# ---------- > Japan Media
# > GYAO!
nameserver /gyao.yahoo.co.jp/group
# ---------- > Europe Media
# > FR:Salto
nameserver /salto.fr/group
# > FR:France.tv
nameserver /ftven.fr/group
`)
	services, err := ParseSmartDNS(input)
	if err != nil {
		t.Fatalf("ParseSmartDNS returned error: %v", err)
	}
	if len(services) != 1 || services[0].Name != "FR:France.tv" {
		t.Fatalf("retired services were not removed: %#v", services)
	}
}

func TestMergeServicesDeduplicatesNamesAndKeepsLegacyIDs(t *testing.T) {
	merged := mergeServices(
		[]Service{{ID: "youtube-old", Name: "YouTube", Category: "Global Platform", Domains: []string{"youtube.com"}}},
		[]Service{{ID: "youtube-new", Name: "youtube", Category: "Custom", Custom: true, Domains: []string{"*.googlevideo.com", "youtube.com"}}},
	)
	if len(merged) != 1 {
		t.Fatalf("expected one service after merge, got %d: %#v", len(merged), merged)
	}
	service := merged[0]
	if service.ID != "youtube-old" || !contains(service.Aliases, "youtube-new") {
		t.Fatalf("canonical and legacy IDs were not preserved: %#v", service)
	}
	if len(service.Domains) != 2 || !contains(service.Domains, "youtube.com") || !contains(service.Domains, "googlevideo.com") {
		t.Fatalf("merged domains were not unique and normalized: %#v", service.Domains)
	}
}
