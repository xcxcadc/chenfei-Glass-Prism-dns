package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotateEnrollmentTokenRevokesPreviousToken(t *testing.T) {
	store, err := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	config, err := store.Save(IPConfig{IP: "203.0.113.40", DNSNodeID: "dns-1", Routes: map[string]string{"service": "proxy"}}, "secret")
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := store.RotateEnrollmentToken(config.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.EnrollmentToken == config.EnrollmentToken {
		t.Fatal("token rotation returned the old token")
	}
	if _, ok := store.GetByToken(config.EnrollmentToken); ok {
		t.Fatal("old enrollment token remained valid")
	}
	if _, ok := store.GetByToken(rotated.EnrollmentToken); !ok {
		t.Fatal("new enrollment token was not persisted")
	}
	if rotated.DNSReady || rotated.RoutesReady || rotated.ServiceAuditedAt != nil {
		t.Fatalf("rotation must invalidate old health and audit state: %+v", rotated)
	}
}

func TestIPConfigExportIsAuthenticatedAndRedacted(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/nodes" && request.Header.Get("Authorization") == "Bearer valid" {
			writeJSON(writer, http.StatusOK, []any{map[string]any{"id": "proxy-1", "role": "proxy", "name": "Japan", "public_ip": "198.51.100.40"}})
			return
		}
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	services, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	config, err := store.Save(IPConfig{IP: "203.0.113.41", DNSNodeID: "dns-1", Routes: map[string]string{"service": "proxy-1"}}, "node-secret")
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("http://127.0.0.1:1/catalog", upstream.Client(), services)
	app, err := NewApp(upstream.URL, catalog, services, store, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}

	unauthorized := httptest.NewRecorder()
	app.Handler().ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/enhancer/api/ip-configs/export", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized export returned %d", unauthorized.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/enhancer/api/ip-configs/export", nil)
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("export returned %d: %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), config.EnrollmentToken) || strings.Contains(response.Body.String(), "node-secret") || strings.Contains(response.Body.String(), "traffic_rx_bytes") {
		t.Fatalf("export leaked runtime credentials or counters: %s", response.Body.String())
	}
	var exported ipConfigExport
	if err := json.Unmarshal(response.Body.Bytes(), &exported); err != nil {
		t.Fatal(err)
	}
	if exported.Version != ipConfigExportVersion || len(exported.Configs) != 1 || exported.Configs[0].Routes["service"].PublicIP != "198.51.100.40" {
		t.Fatalf("unexpected export payload: %+v", exported)
	}
}

func TestResolveImportedRoutesMatchesChangedNodeIDByPublicIP(t *testing.T) {
	routes, err := resolveImportedRoutes(map[string]ipConfigExportRoute{
		"service": {NodeID: "old-id", PublicIP: "198.51.100.50"},
	}, []map[string]any{
		{"id": "new-id", "role": "proxy", "name": "Japan", "public_ip": "198.51.100.50"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if routes["service"] != "new-id" {
		t.Fatalf("route was not remapped by public IP: %+v", routes)
	}
}

func TestIPConfigImportCreatesFreshDNSNodeWithoutImportingSecrets(t *testing.T) {
	var ruleID = 0
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer valid" && request.URL.Path != "/catalog" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case request.URL.Path == "/catalog":
			_, _ = writer.Write([]byte("# ---------- > Test\n# > Test service\nnameserver /example.com/group\n"))
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{map[string]any{"id": "proxy-new", "role": "proxy", "name": "Japan", "public_ip": "198.51.100.60"}})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodPost:
			writeJSON(writer, http.StatusCreated, map[string]any{"id": "dns-new", "secret": "generated-node-secret"})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodPost:
			ruleID++
			var rule map[string]any
			_ = json.NewDecoder(request.Body).Decode(&rule)
			rule["id"] = ruleID
			writeJSON(writer, http.StatusCreated, rule)
		case strings.HasPrefix(request.URL.Path, "/api/rules/") && strings.HasSuffix(request.URL.Path, "/override"):
			writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	services, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), services)
	service := catalog.Snapshot(context.Background(), true).Services[0]
	app, err := NewApp(upstream.URL, catalog, services, store, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	payload := ipConfigExport{Version: ipConfigExportVersion, Configs: []ipConfigExportEntry{{IP: "203.0.113.42", Note: "imported", Routes: map[string]ipConfigExportRoute{service.ID: {NodeID: "proxy-new", PublicIP: "198.51.100.60"}}}}}
	body, _ := json.Marshal(payload)
	request := httptest.NewRequest(http.MethodPost, "/enhancer/api/ip-configs/import", strings.NewReader(string(body)))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("import returned %d: %s", response.Code, response.Body.String())
	}
	var result ipConfigImportResult
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 || len(result.Configs) != 1 || result.Configs[0].DNSNodeID != "dns-new" || result.Configs[0].EnrollmentToken == "" {
		t.Fatalf("unexpected import result: %+v", result)
	}
	if strings.Contains(response.Body.String(), "generated-node-secret") {
		t.Fatalf("import response leaked the DNS node secret: %s", response.Body.String())
	}
}
