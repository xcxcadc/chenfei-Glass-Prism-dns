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
	var createdRule map[string]any
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
			writeJSON(writer, http.StatusOK, []any{map[string]any{"id": 2, "role": "proxy", "public_ip": "198.51.100.20, 2001:db8::20"}})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodPost:
			writeJSON(writer, http.StatusCreated, map[string]any{"id": 7, "secret": "controller-secret"})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodPost:
			var rule map[string]any
			_ = json.NewDecoder(request.Body).Decode(&rule)
			createdRule = rule
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
	if config.DNSNodeID != "7" || config.EnrollmentToken == "" || config.ServiceAuditRequestedAt == nil || override["proxy_node_id"] != "2" || len(config.TrafficPeers) != 2 || config.TrafficPeers[0] != "198.51.100.20" || config.TrafficPeers[1] != "2001:db8::20" {
		t.Fatalf("configuration was not orchestrated: config=%+v override=%+v", config, override)
	}
	if createdRule["target_type"] != "group" || createdRule["target_val"] != "__prism_enhancer_direct__" {
		t.Fatalf("IP rule must use the direct baseline group: %+v", createdRule)
	}

	updateBody := `{"ip":"203.0.113.10","note":"updated","smart":true,"routes":{"` + serviceID + `":"2"}}`
	updateRequest := httptest.NewRequest(http.MethodPut, "/enhancer/api/ip-configs/"+config.ID, strings.NewReader(updateBody))
	updateRequest.Header.Set("Authorization", "Bearer valid")
	updateResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("existing IP update returned %d: %s", updateResponse.Code, updateResponse.Body.String())
	}
	var updated IPConfig
	if err := json.Unmarshal(updateResponse.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.ID != config.ID || updated.IP != config.IP || updated.Note != "updated" {
		t.Fatalf("existing IP update did not preserve identity: before=%+v after=%+v", config, updated)
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
	if !strings.Contains(bootstrapResponse.Body.String(), `"route_domains":["netflix.com"]`) {
		t.Fatalf("bootstrap did not include every routed domain: %s", bootstrapResponse.Body.String())
	}
	if !strings.Contains(bootstrapResponse.Body.String(), `"media_source":"https://media.ispvps.com"`) || !strings.Contains(bootstrapResponse.Body.String(), `"media_tests":["Netflix"]`) {
		t.Fatalf("bootstrap did not include media.ispvps.com checks: %s", bootstrapResponse.Body.String())
	}
	if !strings.Contains(bootstrapResponse.Body.String(), `"traffic_peers":["198.51.100.20","2001:db8::20"]`) {
		t.Fatalf("bootstrap did not preserve both proxy address families: %s", bootstrapResponse.Body.String())
	}
	if !strings.Contains(bootstrapResponse.Body.String(), `"service_audit_requested_at"`) {
		t.Fatalf("bootstrap did not include the service audit request: %s", bootstrapResponse.Body.String())
	}
	if !strings.Contains(bootstrapResponse.Body.String(), `"smart":true`) {
		t.Fatalf("managed IP bootstrap must preserve Agent smart mode: %s", bootstrapResponse.Body.String())
	}

	triggerRequest := httptest.NewRequest(http.MethodPost, "/enhancer/api/ip-configs/"+config.ID+"/audit", nil)
	triggerRequest.Header.Set("Authorization", "Bearer valid")
	triggerResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(triggerResponse, triggerRequest)
	if triggerResponse.Code != http.StatusAccepted {
		t.Fatalf("service audit trigger failed: %d %s", triggerResponse.Code, triggerResponse.Body.String())
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

func TestHealthProbesPreservePersistedOverlappingSelections(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/catalog" {
			_, _ = writer.Write([]byte("# ---------- > Test\n# > Alpha\nnameserver /shared.example/group\n# > Beta\nnameserver /api.shared.example/group\n"))
			return
		}
		http.NotFound(writer, request)
	}))
	defer upstream.Close()

	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	catalog := NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore)
	services := catalog.Snapshot(context.Background(), true).Services
	if len(services) != 2 {
		t.Fatalf("unexpected catalog: %+v", services)
	}
	routes := map[string]string{services[0].ID: "proxy-a", services[1].ID: "proxy-b"}
	config, err := ipStore.Save(IPConfig{
		IP:        "203.0.113.20",
		DNSNodeID: "dns-1",
		Routes:    routes,
	}, "secret", map[string][]string{
		"proxy-a": {"198.51.100.10"},
		"proxy-b": {"198.51.100.20"},
	})
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	record, ok := ipStore.Record(config.ID)
	if !ok {
		t.Fatal("saved IP configuration was not found")
	}
	probes := app.healthProbes(context.Background(), record)
	if len(probes) != 2 {
		t.Fatalf("expected two health probes, got %+v", probes)
	}
	for _, probe := range probes {
		serviceID := probe["service_id"].(string)
		peers := probe["traffic_peers"].([]string)
		if serviceID == services[0].ID && (len(peers) != 1 || peers[0] != "198.51.100.10") {
			t.Fatalf("alpha route was silently changed: %+v", probe)
		}
		if serviceID == services[1].ID && (len(peers) != 1 || peers[0] != "198.51.100.20") {
			t.Fatalf("beta route was silently changed: %+v", probe)
		}
	}
}

func TestApplyIPRoutesLinksOverlappingDomainsBeforeControllerDelivery(t *testing.T) {
	nextRuleID := 10
	overrides := make(map[string]string)
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/catalog":
			_, _ = writer.Write([]byte("# ---------- > AI Platform\n# > Claude\nnameserver /shared.example/group\n# > Gemini\nnameserver /api.shared.example/group\n"))
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{
				map[string]any{"id": "proxy-a", "role": "proxy", "public_ip": "198.51.100.10"},
				map[string]any{"id": "proxy-b", "role": "proxy", "public_ip": "198.51.100.20, 2001:db8::20"},
			})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodPost:
			nextRuleID++
			var rule map[string]any
			if err := json.NewDecoder(request.Body).Decode(&rule); err != nil {
				t.Fatal(err)
			}
			rule["id"] = nextRuleID
			writeJSON(writer, http.StatusCreated, rule)
		case strings.HasPrefix(request.URL.Path, "/api/rules/") && strings.HasSuffix(request.URL.Path, "/override"):
			var payload map[string]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			overrides[request.URL.Path] = payload["proxy_node_id"]
			writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	catalog := NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore)
	services := catalog.Snapshot(context.Background(), true).Services
	serviceIDs := make(map[string]string)
	for _, service := range services {
		serviceIDs[service.Name] = service.ID
	}
	app, err := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	previous := map[string]string{
		serviceIDs["Claude"]: "proxy-a",
		serviceIDs["Gemini"]: "proxy-a",
	}
	current := map[string]string{
		serviceIDs["Claude"]: "proxy-a",
		serviceIDs["Gemini"]: "proxy-b",
	}
	application, err := app.applyIPRoutes(context.Background(), "Bearer valid", "dns-1", previous, current, "https://panel.example")
	if err != nil {
		t.Fatal(err)
	}
	for serviceName, serviceID := range serviceIDs {
		if application.Routes[serviceID] != "proxy-b" {
			t.Fatalf("%s did not follow the latest overlapping-domain selection: %+v", serviceName, application.Routes)
		}
	}
	if len(overrides) != 2 {
		t.Fatalf("expected two linked controller overrides, got %+v", overrides)
	}
	for path, proxyID := range overrides {
		if proxyID != "proxy-b" {
			t.Fatalf("controller override %s used conflicting proxy %s", path, proxyID)
		}
	}
	if len(application.TrafficPeers) != 2 || application.TrafficPeers[0] != "198.51.100.20" || application.TrafficPeers[1] != "2001:db8::20" {
		t.Fatalf("linked route did not preserve both proxy address families: %+v", application.TrafficPeers)
	}
}

func TestIPConfigAdoptsExistingDNSNode(t *testing.T) {
	createdNode := false
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
			writeJSON(writer, http.StatusOK, []any{
				map[string]any{"id": "2", "role": "proxy", "public_ip": "198.51.100.20"},
				map[string]any{"id": "7", "name": "EU", "role": "dns", "public_ip": "203.0.113.10", "secret": "existing-secret"},
			})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodPost:
			createdNode = true
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "unexpected node creation"})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodPost:
			writeJSON(writer, http.StatusCreated, map[string]any{"id": "11"})
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
	serviceID := catalog.Snapshot(context.Background(), true).Services[0].ID
	app, err := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}

	body := `{"ip":"203.0.113.10","note":"EU","smart":true,"existing_dns_node_id":"7","routes":{"` + serviceID + `":"2"}}`
	request := httptest.NewRequest(http.MethodPost, "/enhancer/api/ip-configs", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("adopt returned %d: %s", response.Code, response.Body.String())
	}
	var config IPConfig
	if err := json.Unmarshal(response.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	if createdNode || config.DNSNodeID != "7" || !config.ExternalDNSNode || override["dns_node_id"] != "7" || override["proxy_node_id"] != "2" {
		t.Fatalf("existing DNS node was not adopted correctly: config=%+v override=%+v created=%v", config, override, createdNode)
	}
	record, ok := ipStore.Record(config.ID)
	if !ok || record.NodeSecret != "existing-secret" {
		t.Fatalf("existing node secret was not retained: %+v", record)
	}
}

func TestIPConfigUpdateRecreatesMissingDNSNode(t *testing.T) {
	createdNode := false
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
			writeJSON(writer, http.StatusOK, []any{map[string]any{"id": "proxy-1", "role": "proxy", "public_ip": "198.51.100.20"}})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodPost:
			createdNode = true
			writeJSON(writer, http.StatusCreated, map[string]any{"id": "dns-recreated", "name": "RU", "secret": "new-secret"})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/rules" && request.Method == http.MethodPost:
			writeJSON(writer, http.StatusCreated, map[string]any{"id": "rule-1"})
		case request.URL.Path == "/api/rules/rule-1/override" && request.Method == http.MethodPost:
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
	serviceID := catalog.Snapshot(context.Background(), true).Services[0].ID
	config, err := ipStore.Save(IPConfig{IP: "203.0.113.40", Note: "RU", DNSNodeID: "missing-dns", NodeName: "RU", ExternalDNSNode: true, Routes: map[string]string{serviceID: "proxy-1"}}, "old-secret")
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/enhancer/api/ip-configs/"+config.ID, strings.NewReader(`{"ip":"203.0.113.40","note":"RU","smart":true,"routes":{"`+serviceID+`":"proxy-1"}}`))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("missing DNS node update returned %d: %s", response.Code, response.Body.String())
	}
	var updated IPConfig
	if err := json.Unmarshal(response.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if !createdNode || updated.DNSNodeID != "dns-recreated" || updated.ExternalDNSNode || override["dns_node_id"] != "dns-recreated" || override["proxy_node_id"] != "proxy-1" {
		t.Fatalf("missing DNS node was not recreated and rebound: config=%+v override=%+v created=%v", updated, override, createdNode)
	}
	record, ok := ipStore.Record(config.ID)
	if !ok || record.NodeSecret != "new-secret" {
		t.Fatalf("recreated DNS credentials were not persisted: %+v", record)
	}
}

func TestDeleteAdoptedIPConfigKeepsExistingDNSNode(t *testing.T) {
	deletedNode := false
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/nodes/7" && request.Method == http.MethodDelete:
			deletedNode = true
			writer.WriteHeader(http.StatusNoContent)
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	customStore, _ := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	stored, err := ipStore.Save(IPConfig{IP: "203.0.113.10", DNSNodeID: "7", NodeName: "EU", ExternalDNSNode: true, Routes: map[string]string{}}, "existing-secret")
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(upstream.URL, NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore), customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/enhancer/api/ip-configs/"+stored.ID, nil)
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete returned %d: %s", response.Code, response.Body.String())
	}
	if deletedNode {
		t.Fatal("deleting an adopted IP configuration must not delete the existing DNS node")
	}
	if _, ok := ipStore.Record(stored.ID); ok {
		t.Fatal("IP configuration was not deleted")
	}
}

func TestApplyIPRoutesMigratesLegacyRuleBeforeClearingOverride(t *testing.T) {
	var migrated map[string]any
	var override map[string]string
	var serviceID string
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/catalog":
			_, _ = writer.Write([]byte("# ---------- > Global Platform\n# > Netflix\nnameserver /netflix.com/group\n"))
		case request.URL.Path == "/api/rules" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{map[string]any{
				"id": "11", "name": "Stream · Netflix", "type": "RULE-SET",
				"source_type": "all", "source_val": "", "target_type": "node", "target_val": "2",
				"value": "https://panel.example/enhancer/rules/" + serviceID + ".list", "enabled": true,
			}})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{map[string]any{"id": "2", "role": "proxy", "public_ip": "198.51.100.20"}})
		case request.URL.Path == "/api/rules/11" && request.Method == http.MethodPut:
			_ = json.NewDecoder(request.Body).Decode(&migrated)
			writer.WriteHeader(http.StatusNoContent)
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
	serviceID = catalog.Snapshot(context.Background(), true).Services[0].ID
	app, _ := NewApp(upstream.URL, catalog, customStore, ipStore, upstream.Client())

	if _, err := app.applyIPRoutes(context.Background(), "Bearer valid", "7", map[string]string{serviceID: "2"}, map[string]string{}, "https://panel.example"); err != nil {
		t.Fatal(err)
	}
	if migrated["target_type"] != "group" || migrated["target_val"] != "__prism_enhancer_direct__" {
		t.Fatalf("legacy rule was not migrated to the direct baseline: %+v", migrated)
	}
	if override["dns_node_id"] != "7" || override["proxy_node_id"] != "" {
		t.Fatalf("override was not cleared: %+v", override)
	}
}

func TestPreferredProbeDomainsUsesSeveralStableGenericCandidates(t *testing.T) {
	service := Service{
		Name: "Future Service",
		Domains: []string{
			"legacy-api.long-example.invalid",
			"stream.example.com",
			"example.com",
			"cdn.example.net",
			"api.example.org",
		},
	}

	domains := preferredProbeDomains(service)
	if len(domains) != 4 {
		t.Fatalf("expected four generic probe domains, got %#v", domains)
	}
	if domains[0] != "example.com" {
		t.Fatalf("expected the shortest root domain first, got %#v", domains)
	}
	if contains(domains, "legacy-api.long-example.invalid") {
		t.Fatalf("deep legacy hostname should not displace stable root candidates: %#v", domains)
	}
}

func TestPreferredProbeDomainsCompilesWildcardPatterns(t *testing.T) {
	service := Service{Name: "Custom", Domains: []string{"*.cdn.example.com", "*.example.com"}}
	domains := preferredProbeDomains(service)
	if len(domains) != 2 || domains[0] != "example.com" || domains[1] != "cdn.example.com" {
		t.Fatalf("wildcard patterns reached probes: %#v", domains)
	}
}

func TestPreferredProbeDomainsIncludesGeminiApplicationDependencies(t *testing.T) {
	service := Service{Name: "Gemini", Domains: append([]string(nil), geminiApplicationDomains...)}
	domains := preferredProbeDomains(service)
	if len(domains) != len(service.Domains) {
		t.Fatalf("expected every Gemini application dependency, got %#v", domains)
	}
	for _, domain := range service.Domains {
		if !contains(domains, domain) {
			t.Fatalf("missing Gemini application dependency %q from %#v", domain, domains)
		}
	}
}

func TestUnlockTestProvidersUseMediaISPVPSLabels(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{name: "Gemini", want: []string{"Google Gemini"}},
		{name: "Disney+", want: []string{"Disney+"}},
		{name: "YouTube", want: []string{"YouTube Premium"}},
		{name: "Netflix", want: []string{"Netflix"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := unlockTestProviders(Service{Name: test.name})
			if strings.Join(got, "\x00") != strings.Join(test.want, "\x00") {
				t.Fatalf("media labels for %q = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

func TestMediaTestSpecSeparatesRequiredAndAlternativeLabels(t *testing.T) {
	youtube := mediaTestSpecForService(Service{Name: "YouTube"})
	if strings.Join(youtube.Required, "\x00") != "YouTube Premium" || len(youtube.Any) != 0 {
		t.Fatalf("YouTube media spec = %+v, want only YouTube Premium required", youtube)
	}
	bilibili := mediaTestSpecForService(Service{Name: "Bilibili"})
	if len(bilibili.Required) != 0 || strings.Join(bilibili.Any, "\x00") != "BiliBili China Mainland Only\x00BiliBili Hongkong/Macau/Taiwan\x00Bilibili Taiwan Only" {
		t.Fatalf("Bilibili media spec = %+v, want alternative regional labels", bilibili)
	}
	claude := mediaTestSpecForService(Service{Name: "Claude"})
	if strings.Join(claude.Required, "\x00") != "Claude" || len(claude.Any) != 0 {
		t.Fatalf("Claude should be tested as an exact media.ispvps.com label: %+v", claude)
	}
}

func TestMediaTestSpecAcceptsCatalogAliases(t *testing.T) {
	for _, test := range []struct {
		name  string
		label string
	}{
		{name: "Hulu", label: "Hulu"},
		{name: "HBO Max", label: "HBO Max"},
		{name: "Peacock TV", label: "Peacock TV"},
		{name: "Amazon Prime Video", label: "Amazon Prime Video"},
		{name: "OpenAI", label: "ChatGPT"},
	} {
		spec := mediaTestSpecForService(Service{Name: test.name})
		if len(spec.Required) != 1 || spec.Required[0] != test.label {
			t.Fatalf("media alias %q = %+v, want %q", test.name, spec, test.label)
		}
	}
}

func TestMediaTestSpecCoversRegionalCatalogLabels(t *testing.T) {
	for _, test := range []struct {
		name  string
		label string
	}{
		{name: "Abema", label: "Abema.TV"},
		{name: "Bilibili", label: "BiliBili Hongkong/Macau/Taiwan"},
		{name: "Hulu Japan", label: "Hulu Japan"},
		{name: "Viu.TV", label: "Viu.TV"},
		{name: "YouTube", label: "YouTube Premium"},
	} {
		spec := mediaTestSpecForService(Service{Name: test.name})
		labels := append(append([]string(nil), spec.Required...), spec.Any...)
		if !contains(labels, test.label) {
			t.Fatalf("regional media label %q = %+v, want %q", test.name, spec, test.label)
		}
	}
}

func TestMediaTestSpecUsesDomainsForCustomNames(t *testing.T) {
	for _, test := range []struct {
		name   string
		domain string
		label  string
	}{
		{name: "我的 AI 服务", domain: "gemini.google.com", label: "Google Gemini"},
		{name: "视频服务", domain: "*.youtube.com", label: "YouTube Premium"},
		{name: "自定义流媒体", domain: "disneyplus.com", label: "Disney+"},
	} {
		spec := mediaTestSpecForService(Service{Name: test.name, Domains: []string{test.domain}})
		if len(spec.Required) != 1 || spec.Required[0] != test.label {
			t.Fatalf("domain mapping %q/%q = %+v, want %q", test.name, test.domain, spec, test.label)
		}
	}
}

func TestPreferredProbeDomainsUsesReachableOpenAIDependencyHosts(t *testing.T) {
	service := Service{Name: "ChatGPT / OpenAI", Domains: []string{
		"chatgpt.com",
		"openai.com",
		"oaistatic.com",
		"oaiusercontent.com",
		"sora.com",
	}}
	domains := preferredProbeDomains(service)
	for _, domain := range []string{"chatgpt.com", "openai.com", "cdn.oaistatic.com", "files.oaiusercontent.com", "sora.com"} {
		if !contains(domains, domain) {
			t.Fatalf("missing reachable OpenAI dependency %q from %#v", domain, domains)
		}
	}
	for _, rootOnly := range []string{"oaistatic.com", "oaiusercontent.com"} {
		if contains(domains, rootOnly) {
			t.Fatalf("non-serving root %q should not be used as an HTTPS dependency: %#v", rootOnly, domains)
		}
	}
}

func TestCopilotImageCreatorUsesItsOwnPathChecks(t *testing.T) {
	service := Service{Name: "Microsoft Copilot Image Creator", Domains: []string{"copilot.microsoft.com"}}
	if providers := unlockTestProviders(service); len(providers) != 1 || providers[0] != service.Name {
		t.Fatalf("custom media label should be tested explicitly: %#v", providers)
	}
	if domains := preferredProbeDomains(service); len(domains) != 1 || domains[0] != "copilot.microsoft.com" {
		t.Fatalf("expected the Copilot page probe, got %#v", domains)
	}
}
