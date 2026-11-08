package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalServiceNameStripsRegionQualifier(t *testing.T) {
	cases := map[string]string{
		"Claude 2 /HK Only/":        "Claude",
		"Google Gemini /HK Only/":   "Gemini",
		"Google AiStudio /HK Only/": "Google AI Studio",
		"Netflix":                   "Netflix",
		"Some Service /JP Only/":    "Some Service",
	}
	for input, want := range cases {
		if got := canonicalServiceName(input, "AI Platform"); got != want {
			t.Errorf("canonicalServiceName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseSmartDNSAppliesSupplementsToQualifiedNames(t *testing.T) {
	list := "# ---> AI Platform\n" +
		"# > Google Gemini /HK Only/\n" +
		"nameserver /gemini.google.com/<replace with groupname>\n" +
		"# > Claude 2 /HK Only/\n" +
		"nameserver /claude.ai/<replace with groupname>\n"
	services, err := ParseSmartDNS(strings.NewReader(list))
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Service{}
	for _, service := range services {
		byName[service.Name] = service
	}
	gemini, ok := byName["Gemini"]
	if !ok {
		t.Fatalf("Gemini not found, got %v", services)
	}
	if !contains(gemini.Domains, "accounts.google.com") || !contains(gemini.Domains, "www.gstatic.com") {
		t.Fatalf("Gemini supplements were not applied: %v", gemini.Domains)
	}
	// 路由以服务 ID 保存，改名不能改变 ID，否则已有路由会失效。
	if gemini.ID != stableServiceID("AI Platform", "Google Gemini /HK Only/") {
		t.Fatalf("service ID must stay derived from the upstream name, got %s", gemini.ID)
	}
	if _, ok := byName["Claude"]; !ok {
		t.Fatalf("Claude not found: %v", services)
	}
}

func TestMediaTestSpecForQualifiedServices(t *testing.T) {
	cases := map[string]string{
		"Claude 2 /HK Only/":      "Claude",
		"Google Gemini /HK Only/": "Google Gemini",
	}
	for upstream, label := range cases {
		service := Service{Name: canonicalServiceName(upstream, "AI Platform")}
		spec := mediaTestSpecForService(service)
		if len(spec.Required) != 1 || spec.Required[0] != label {
			t.Errorf("%s: spec = %+v, want label %q", upstream, spec, label)
		}
	}
}

func TestSingleLabelDomainsAreRejected(t *testing.T) {
	for _, value := range []string{"bbc", "twitter", "colab", "localhost"} {
		if got := normalizeDomain(value); got != "" {
			t.Errorf("normalizeDomain(%q) = %q, want empty", value, got)
		}
	}
	if _, err := normalizeCustomDomains([]string{"example.com", "colab"}); err == nil {
		t.Fatal("expected single-label custom domain to be rejected")
	}
	if got := normalizeDomain("*.Example.com."); got != "example.com" {
		t.Fatalf("wildcard normalization changed: %q", got)
	}
}

func TestCatalogPreferencesLoadDropsLegacySingleLabelDomains(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog-preferences.json")
	body := `{"service_domains":{"svc":["gemini.google.com","colab","twitter","deepmind.com"]}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := NewCatalogPreferenceStore(path)
	if err != nil {
		t.Fatalf("legacy preferences must still load: %v", err)
	}
	applied := store.Apply([]Service{{ID: "svc", Name: "Svc", Domains: []string{"x.example.com"}}})
	if len(applied) != 1 {
		t.Fatalf("unexpected result %v", applied)
	}
	want := []string{"deepmind.com", "gemini.google.com"}
	if strings.Join(applied[0].Domains, ",") != strings.Join(want, ",") {
		t.Fatalf("domains = %v, want %v", applied[0].Domains, want)
	}
}
