package main

import "testing"

func TestNormalizeConflictingRoutesUsesChangedServiceProxy(t *testing.T) {
	services := []Service{
		{ID: "max", Domains: []string{"max.com", "hbomax.com"}},
		{ID: "hbo", Domains: []string{"*.max.com"}},
		{ID: "other", Domains: []string{"example.com"}},
	}
	previous := map[string]string{"max": "proxy-a", "hbo": "proxy-a", "other": "proxy-c"}
	current := map[string]string{"max": "proxy-a", "hbo": "proxy-b", "other": "proxy-c"}

	normalized, changed := normalizeConflictingRoutes(previous, current, services)
	if !changed {
		t.Fatal("expected overlapping routes to be normalized")
	}
	if normalized["max"] != "proxy-b" || normalized["hbo"] != "proxy-b" {
		t.Fatalf("changed service proxy did not win: %+v", normalized)
	}
	if normalized["other"] != "proxy-c" {
		t.Fatalf("unrelated service was changed: %+v", normalized)
	}
}

func TestConflictingServiceIDsIncludesTransitiveDomainOverlap(t *testing.T) {
	services := []Service{
		{ID: "a", Domains: []string{"example.com"}},
		{ID: "b", Domains: []string{"api.example.com", "video.example.net"}},
		{ID: "c", Domains: []string{"*.video.example.net"}},
		{ID: "d", Domains: []string{"independent.test"}},
	}
	routes := map[string]string{"a": "p1", "b": "p1", "c": "p1", "d": "p2"}

	linked := conflictingServiceIDs("a", routes, services)
	if len(linked) != 3 || linked[0] != "a" || linked[1] != "b" || linked[2] != "c" {
		t.Fatalf("unexpected linked services: %+v", linked)
	}
}

func TestNormalizeConflictingRoutesIncludesParentAndChildDomains(t *testing.T) {
	services := []Service{
		{ID: "parent", Domains: []string{"example.com"}},
		{ID: "child", Domains: []string{"api.example.com"}},
	}
	normalized, changed := normalizeConflictingRoutes(
		map[string]string{"parent": "proxy-a", "child": "proxy-a"},
		map[string]string{"parent": "proxy-a", "child": "proxy-b"},
		services,
	)
	if !changed || normalized["parent"] != "proxy-b" || normalized["child"] != "proxy-b" {
		t.Fatalf("parent and child DNS routes were not linked: %+v", normalized)
	}
}

func TestNormalizeConflictingRoutesCanonicalizesLegacyServiceIDs(t *testing.T) {
	services := []Service{
		{ID: "google-gemini-hk-only-726ddd4d", Name: "Google Gemini /HK Only/", Domains: []string{"gemini.google.com"}},
		{ID: "openai-47b4550d", Name: "ChatGPT / OpenAI", Domains: []string{"openai.com"}},
	}
	current := map[string]string{
		"google-gemini-892afd44":   "proxy-sg",
		"openai-47b4550d":          "proxy-my",
		"deleted-service-a1b2c3d4": "proxy-old",
	}

	normalized, changed := normalizeConflictingRoutes(nil, current, services)
	if !changed {
		t.Fatal("expected legacy service ids to be normalized")
	}
	if normalized["google-gemini-hk-only-726ddd4d"] != "proxy-sg" {
		t.Fatalf("legacy Gemini route was not migrated: %+v", normalized)
	}
	if normalized["openai-47b4550d"] != "proxy-my" {
		t.Fatalf("canonical route was changed: %+v", normalized)
	}
	if _, exists := normalized["google-gemini-892afd44"]; exists {
		t.Fatalf("legacy route id leaked into normalized routes: %+v", normalized)
	}
	if _, exists := normalized["deleted-service-a1b2c3d4"]; exists {
		t.Fatalf("unknown stale route id should be dropped: %+v", normalized)
	}
}

func TestStalePreviousRouteIDsIncludesLegacyAndUnknownIDs(t *testing.T) {
	services := []Service{
		{ID: "claude-2-hk-only-22ccd555", Name: "Claude 2 /HK Only/", Domains: []string{"claude.ai"}},
	}
	previous := map[string]string{
		"claude-2-67be3388":         "proxy-vn",
		"unknown-service-11111111":  "proxy-old",
		"claude-2-hk-only-22ccd555": "proxy-vn",
	}
	stale := stalePreviousRouteIDs(previous, services)
	if len(stale) != 2 || stale[0] != "claude-2-67be3388" || stale[1] != "unknown-service-11111111" {
		t.Fatalf("unexpected stale route ids: %+v", stale)
	}
}
