package main

import "testing"

func TestEnsureBuiltInSocialServices(t *testing.T) {
	services := ensureBuiltInServices(nil)
	byName := map[string]Service{}
	for _, svc := range services {
		byName[svc.Name] = svc
	}
	expected := map[string]string{
		"X (Twitter)":     "x.com",
		"Instagram (INS)": "instagram.com",
		"Facebook":        "facebook.com",
		"Threads":         "threads.net",
		"WhatsApp":        "whatsapp.com",
		"Telegram":        "telegram.org",
		"Discord":         "discord.com",
		"Reddit":          "reddit.com",
		"Pinterest":       "pinterest.com",
		"Line":            "line.me",
	}
	for name, requiredDomain := range expected {
		svc, ok := byName[name]
		if !ok {
			t.Fatalf("missing built-in social service %q", name)
		}
		if svc.Category != "Social Platform" {
			t.Fatalf("expected %q category Social Platform, got %q", name, svc.Category)
		}
		if !contains(svc.Domains, requiredDomain) {
			t.Fatalf("service %q missing domain %q: %v", name, requiredDomain, svc.Domains)
		}
		domain, probeURL := nativeProbeTarget(svc)
		if domain == "" || probeURL == "" {
			t.Fatalf("service %q missing native probe target", name)
		}
	}
}
