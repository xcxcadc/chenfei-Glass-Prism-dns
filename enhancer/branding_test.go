package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrandingSettingsPersistAndRequireAuthentication(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/nodes" && request.Header.Get("Authorization") == "Bearer valid" {
			writeJSON(writer, http.StatusOK, []any{})
			return
		}
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}))
	defer upstream.Close()

	dataDir := t.TempDir()
	customStore, _ := NewCustomServiceStore(filepath.Join(dataDir, "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(dataDir, "ip-configs.json"))
	app, err := NewApp(upstream.URL, NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore), customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	brandingPath := filepath.Join(dataDir, "branding.json")
	app.branding, _ = NewBrandingStore(brandingPath)

	publicRequest := httptest.NewRequest(http.MethodGet, "/enhancer/api/branding", nil)
	publicResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(publicResponse, publicRequest)
	if publicResponse.Code != http.StatusOK {
		t.Fatalf("public branding read returned %d", publicResponse.Code)
	}

	body := `{"site_name":"辰飞 DNS","browser_title":"辰飞 DNS 管理中心","site_tagline":"全球流媒体解锁"}`
	unauthorizedRequest := httptest.NewRequest(http.MethodPut, "/enhancer/api/branding", strings.NewReader(body))
	unauthorizedResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized update returned %d", unauthorizedResponse.Code)
	}

	updateRequest := httptest.NewRequest(http.MethodPut, "/enhancer/api/branding", strings.NewReader(body))
	updateRequest.Header.Set("Authorization", "Bearer valid")
	updateResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("branding update returned %d: %s", updateResponse.Code, updateResponse.Body.String())
	}

	reloaded, err := NewBrandingStore(brandingPath)
	if err != nil {
		t.Fatal(err)
	}
	settings := reloaded.Get()
	if settings.SiteName != "辰飞 DNS" || settings.BrowserTitle != "辰飞 DNS 管理中心" || settings.SiteTagline != "全球流媒体解锁" {
		t.Fatalf("branding settings were not restored: %+v", settings)
	}

	var responseSettings BrandingSettings
	if err := json.Unmarshal(updateResponse.Body.Bytes(), &responseSettings); err != nil {
		t.Fatal(err)
	}
	if responseSettings != settings {
		t.Fatalf("unexpected update response: %+v", responseSettings)
	}
}

func TestBrandingValidation(t *testing.T) {
	store, err := NewBrandingStore(filepath.Join(t.TempDir(), "branding.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(BrandingSettings{SiteName: "", BrowserTitle: "Title"}); err == nil {
		t.Fatal("expected empty site name to fail")
	}
	if _, err := store.Save(BrandingSettings{SiteName: "Prism", BrowserTitle: strings.Repeat("页", 97)}); err == nil {
		t.Fatal("expected oversized browser title to fail")
	}
}
