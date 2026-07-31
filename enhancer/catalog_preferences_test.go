package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCatalogPreferenceStorePersistsUnicodeCategories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog-preferences.json")
	store, err := NewCatalogPreferenceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	category, err := store.AddCategory("海外视频 / 自定义")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetServiceCategory("youtube-test", category); err != nil {
		t.Fatal(err)
	}
	service := store.Apply([]Service{{ID: "youtube-test", Name: "YouTube", Category: "China Media"}})[0]
	if service.Category != category || service.OriginalCategory != "China Media" {
		t.Fatalf("unexpected applied category: %+v", service)
	}

	reloaded, err := NewCatalogPreferenceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	service = reloaded.Apply([]Service{{ID: "youtube-test", Name: "YouTube", Category: "China Media"}})[0]
	if service.Category != category || len(reloaded.Categories()) != 1 {
		t.Fatalf("preferences did not persist: %+v %#v", service, reloaded.Categories())
	}
	if err := reloaded.ClearServiceCategory("youtube-test"); err != nil {
		t.Fatal(err)
	}
	service = reloaded.Apply([]Service{{ID: "youtube-test", Name: "YouTube", Category: "China Media"}})[0]
	if service.Category != "China Media" || service.OriginalCategory != "" {
		t.Fatalf("category override was not cleared: %+v", service)
	}
}

func TestCatalogPreferenceStorePersistsDomainOverridesAndLegacyAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog-preferences.json")
	store, err := NewCatalogPreferenceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetServiceDomains("legacy-youtube", []string{"Example.COM", "*.cdn.example.com"}); err != nil {
		t.Fatal(err)
	}
	service := store.Apply([]Service{{ID: "canonical-youtube", Aliases: []string{"legacy-youtube"}, Name: "YouTube", Domains: []string{"youtube.com"}}})[0]
	if !service.DomainOverride || len(service.Domains) != 2 || service.Domains[0] != "*.cdn.example.com" {
		t.Fatalf("alias domain override was not applied: %#v", service)
	}

	reloaded, err := NewCatalogPreferenceStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := reloaded.ClearServiceDomains("legacy-youtube"); err != nil {
		t.Fatal(err)
	}
	service = reloaded.Apply([]Service{{ID: "canonical-youtube", Aliases: []string{"legacy-youtube"}, Name: "YouTube", Domains: []string{"youtube.com"}}})[0]
	if service.DomainOverride || len(service.Domains) != 1 || service.Domains[0] != "youtube.com" {
		t.Fatalf("clearing domain override did not restore catalog values: %#v", service)
	}
}

func TestCatalogSnapshotIncludesEmptyCustomCategories(t *testing.T) {
	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	preferences, _ := NewCatalogPreferenceStore(filepath.Join(t.TempDir(), "catalog-preferences.json"))
	_, _ = preferences.AddCategory("空分类")
	_ = preferences.SetServiceCategory("youtube-test", "全球视频")
	manager := NewCatalogManager("http://127.0.0.1:1/unavailable", http.DefaultClient, customStore, preferences)
	manager.snapshot = CatalogSnapshot{
		Source:    "test",
		UpdatedAt: time.Now(),
		Services:  []Service{{ID: "youtube-test", Name: "YouTube", Category: "Global Platform"}},
	}

	snapshot := manager.Snapshot(context.Background(), false)
	if len(snapshot.Services) != 1 || snapshot.Services[0].Category != "全球视频" {
		t.Fatalf("service category override missing: %+v", snapshot.Services)
	}
	for _, expected := range []string{"全球视频", "空分类"} {
		found := false
		for _, category := range snapshot.Categories {
			found = found || category == expected
		}
		if !found {
			t.Fatalf("category %q missing from %#v", expected, snapshot.Categories)
		}
	}
}

func TestCategoryAPIManagesAssignmentsWithoutChangingServiceID(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/nodes":
			if request.Header.Get("Authorization") != "Bearer valid" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = writer.Write([]byte(`[]`))
		case "/catalog":
			_, _ = writer.Write([]byte("# ---------- > China Media\n# > YouTube\nnameserver /youtube.com/group\n"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	preferences, _ := NewCatalogPreferenceStore(filepath.Join(t.TempDir(), "catalog-preferences.json"))
	manager := NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore, preferences)
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	app, _ := NewApp(upstream.URL, manager, customStore, ipStore, upstream.Client())
	handler := app.Handler()
	service := manager.Snapshot(context.Background(), true).Services[0]
	if service.Category != "Global Platform" {
		t.Fatalf("YouTube should default to Global Platform: %+v", service)
	}

	request := httptest.NewRequest(http.MethodPost, "/enhancer/api/categories", strings.NewReader(`{"name":"海外视频"}`))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create category failed: %d %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/enhancer/api/categories", strings.NewReader(`{"name":"海外视频"}`))
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("duplicate category should be rejected: %d %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPut, "/enhancer/api/service-categories/"+service.ID, strings.NewReader(`{"category":"海外视频"}`))
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"original_category":"Global Platform"`) {
		t.Fatalf("assign category failed: %d %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/enhancer/api/categories", strings.NewReader(`{"name":"海外视频"}`))
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("used category should not be deleted: %d %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/enhancer/api/service-categories/"+service.ID, nil)
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"category":"Global Platform"`) {
		t.Fatalf("restore category failed: %d %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/enhancer/api/categories", strings.NewReader(`{"name":"海外视频"}`))
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete category failed: %d %s", response.Code, response.Body.String())
	}
}
