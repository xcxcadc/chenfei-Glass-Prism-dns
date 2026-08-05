package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"mime/multipart"
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

func TestBrandingIconValidationAndPersistence(t *testing.T) {
	data, err := fs.ReadFile(embeddedWeb, "web/favicon.png")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBrandingIcon(data); err != nil {
		t.Fatalf("default favicon should be a transparent PNG: %v", err)
	}

	store, err := NewBrandingStore(filepath.Join(t.TempDir(), "branding.json"))
	if err != nil {
		t.Fatal(err)
	}
	custom := testBrandingIcon(t)
	saved, err := store.SaveIcon(custom)
	if err != nil {
		t.Fatal(err)
	}
	if saved.IconVersion == 0 {
		t.Fatal("expected icon version after upload")
	}
	stored, err := store.Icon()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, custom) {
		t.Fatal("stored icon differs from uploaded icon")
	}
	if _, err := store.SaveIcon(testOpaqueBrandingIcon(t)); err == nil {
		t.Fatal("expected opaque icon to be rejected")
	}

	reloaded, err := NewBrandingStore(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Get().IconVersion != saved.IconVersion {
		t.Fatalf("icon version was not persisted: %+v", reloaded.Get())
	}
}

func TestBrandingIconEndpointRequiresAuthentication(t *testing.T) {
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
	app.branding, _ = NewBrandingStore(filepath.Join(dataDir, "branding.json"))
	custom := testBrandingIcon(t)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("icon", "prism.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(custom); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/enhancer/api/branding/icon", &body)
	request.Header.Set("Content-Type", form.FormDataContentType())
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("icon upload returned %d: %s", response.Code, response.Body.String())
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/enhancer/api/branding/icon", nil)
	getResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK || !bytes.Equal(getResponse.Body.Bytes(), custom) {
		t.Fatalf("uploaded icon was not served: status=%d", getResponse.Code)
	}
	rootIconRequest := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rootIconResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(rootIconResponse, rootIconRequest)
	if rootIconResponse.Code != http.StatusOK || !bytes.Equal(rootIconResponse.Body.Bytes(), custom) {
		t.Fatalf("uploaded favicon was not served: status=%d", rootIconResponse.Code)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/enhancer/api/branding/icon", nil)
	deleteRequest.Header.Set("Authorization", "Bearer valid")
	deleteResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("icon reset returned %d: %s", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func testBrandingIcon(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	canvas.SetNRGBA(16, 16, color.NRGBA{R: 30, G: 100, B: 240, A: 255})
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func testOpaqueBrandingIcon(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			canvas.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
