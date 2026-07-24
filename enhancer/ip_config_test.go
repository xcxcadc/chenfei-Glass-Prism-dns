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

func TestIPConfigCreateBootstrapAndTraffic(t *testing.T) {
	var override map[string]string
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/catalog" {
			_, _ = writer.Write([]byte("# ---------- > Global Platform\n# > Netflix\nnameserver /netflix.com/group\n"))
			return
		}
		if request.Header.Get("Authorization") != "Bearer valid" {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		switch {
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{map[string]any{"id": 2, "role": "proxy", "public_ip": "198.51.100.20"}})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodPost:
			writeJSON(writer, http.StatusCreated, map[string]any{"id": 7, "secret": "controller-secret"})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodPost:
			var rule map[string]any
			_ = json.NewDecoder(request.Body).Decode(&rule)
			rule["id"] = 11
			writeJSON(writer, http.StatusCreated, rule)
		case request.URL.Path == "/api/rules/11/override" && request.Method == http.MethodPost:
			_ = json.NewDecoder(request.Body).Decode(&override)
			writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	catalog := NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore)
	snapshot := catalog.Snapshot(context.Background(), true)
	if len(snapshot.Services) != 1 {
		t.Fatalf("unexpected catalog: %+v", snapshot)
	}
	serviceID := snapshot.Services[0].ID
	app, err := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}

	body := `{"ip":"203.0.113.10","note":"edge","smart":true,"routes":{"` + serviceID + `":"2"}}`
	request := httptest.NewRequest(http.MethodPost, "/enhancer/api/ip-configs", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	var config IPConfig
	if err := json.Unmarshal(response.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	if config.DNSNodeID != "7" || config.EnrollmentToken == "" || override["proxy_node_id"] != "2" || len(config.TrafficPeers) != 1 || config.TrafficPeers[0] != "198.51.100.20" {
		t.Fatalf("configuration was not orchestrated: config=%+v override=%+v", config, override)
	}

	bootstrap := httptest.NewRequest(http.MethodGet, "/enhancer/api/bootstrap/"+config.EnrollmentToken, nil)
	bootstrapResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(bootstrapResponse, bootstrap)
	if bootstrapResponse.Code != http.StatusOK || !strings.Contains(bootstrapResponse.Body.String(), "controller-secret") {
		t.Fatalf("bootstrap failed: %d %s", bootstrapResponse.Code, bootstrapResponse.Body.String())
	}
	if !strings.Contains(bootstrapResponse.Body.String(), `"health_probes"`) || !strings.Contains(bootstrapResponse.Body.String(), `"netflix.com"`) {
		t.Fatalf("bootstrap did not include route probes: %s", bootstrapResponse.Body.String())
	}
	if !strings.Contains(bootstrapResponse.Body.String(), `"unlock_test":"Netflix"`) {
		t.Fatalf("bootstrap did not include UnlockTests provider: %s", bootstrapResponse.Body.String())
	}

	for _, report := range []string{
		`{"token":"` + config.EnrollmentToken + `","scope":"unlock_peers","interface":"nftables","rx_bytes":100,"tx_bytes":200,"dns_ready":true,"system_dns_ready":true,"routes_ready":true,"healthy_routes":1,"expected_routes":1}`,
		`{"token":"` + config.EnrollmentToken + `","scope":"unlock_peers","interface":"nftables","rx_bytes":160,"tx_bytes":280,"dns_ready":true,"system_dns_ready":true,"routes_ready":true,"healthy_routes":1,"expected_routes":1}`,
	} {
		trafficRequest := httptest.NewRequest(http.MethodPost, "/enhancer/api/traffic/report", strings.NewReader(report))
		trafficResponse := httptest.NewRecorder()
		app.Handler().ServeHTTP(trafficResponse, trafficRequest)
		if trafficResponse.Code != http.StatusOK {
			t.Fatalf("traffic report failed: %d %s", trafficResponse.Code, trafficResponse.Body.String())
		}
	}
	stored, _ := ipStore.Get(config.ID)
	if stored.TrafficRXBytes != 60 || stored.TrafficTXBytes != 80 {
		t.Fatalf("traffic was not accumulated: %+v", stored)
	}
	if !stored.DNSReady || !stored.SystemDNSReady || !stored.RoutesReady || stored.HealthyRoutes != 1 {
		t.Fatalf("client health was not stored: %+v", stored)
	}

	auditRequest := httptest.NewRequest(http.MethodPost, "/enhancer/api/audit/report", strings.NewReader(
		`{"token":"`+config.EnrollmentToken+`","scope":"unlock_services","results":{"`+serviceID+`":"YES (Region: SG) [Via DNS]"}}`,
	))
	auditResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(auditResponse, auditRequest)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("service audit report failed: %d %s", auditResponse.Code, auditResponse.Body.String())
	}
	stored, _ = ipStore.Get(config.ID)
	if stored.ServiceResults[serviceID] == "" || stored.ServiceAuditedAt == nil {
		t.Fatalf("service audit was not stored: %+v", stored)
	}
}
