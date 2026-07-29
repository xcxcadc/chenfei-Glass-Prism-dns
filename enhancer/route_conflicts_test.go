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
