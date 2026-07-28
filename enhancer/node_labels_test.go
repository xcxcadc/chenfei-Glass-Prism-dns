package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnhancedNodeUnicodeLabels(t *testing.T) {
	controllerNode := map[string]any{
		"id": "7", "name": "EU", "group": "", "role": "dns", "public_ip": "203.0.113.10",
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer valid" {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		switch {
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{controllerNode})
		case request.URL.Path == "/api/nodes/7" && request.Method == http.MethodPut:
			var payload map[string]any
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			controllerNode["name"] = payload["name"]
			controllerNode["group"] = payload["group"]
			writer.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	dataDir := t.TempDir()
	customStore, _ := NewCustomServiceStore(filepath.Join(dataDir, "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(dataDir, "ip-configs.json"))
	labelPath := filepath.Join(dataDir, "node-labels.json")
	labelStore, _ := NewNodeLabelStore(labelPath)
	app, err := NewApp(upstream.URL, NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore), customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	app.nodeLabels = labelStore

	body := `{"name":"EU俄罗斯","group":"欧洲,俄罗斯","public_ip":"203.0.113.10"}`
	request := httptest.NewRequest(http.MethodPut, "/enhancer/api/nodes/7", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("update returned %d: %s", response.Code, response.Body.String())
	}
	if valueString(controllerNode["name"]) != "EU" || valueString(controllerNode["group"]) != "" {
		t.Fatalf("controller received incompatible labels: %+v", controllerNode)
	}

	reloaded, err := NewNodeLabelStore(labelPath)
	if err != nil {
		t.Fatal(err)
	}
	app.nodeLabels = reloaded
	request = httptest.NewRequest(http.MethodGet, "/enhancer/api/nodes", nil)
	request.Header.Set("Authorization", "Bearer valid")
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("list returned %d: %s", response.Code, response.Body.String())
	}
	var nodes []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &nodes); err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0]["name"] != "EU俄罗斯" || nodes[0]["group"] != "欧洲,俄罗斯" {
		t.Fatalf("display labels were not restored: %+v", nodes)
	}
}

func TestEnhancedNodeCreateWithUnicodeName(t *testing.T) {
	var controllerPayload map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer valid" {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		switch {
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, []any{})
		case request.URL.Path == "/api/nodes" && request.Method == http.MethodPost:
			if err := json.NewDecoder(request.Body).Decode(&controllerPayload); err != nil {
				t.Fatal(err)
			}
			created := make(map[string]any, len(controllerPayload)+1)
			for key, value := range controllerPayload {
				created[key] = value
			}
			created["id"] = "8"
			writeJSON(writer, http.StatusCreated, created)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer upstream.Close()

	dataDir := t.TempDir()
	customStore, _ := NewCustomServiceStore(filepath.Join(dataDir, "services.json"))
	ipStore, _ := NewIPConfigStore(filepath.Join(dataDir, "ip-configs.json"))
	app, err := NewApp(upstream.URL, NewCatalogManager(upstream.URL+"/catalog", upstream.Client(), customStore), customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	app.nodeLabels, _ = NewNodeLabelStore(filepath.Join(dataDir, "node-labels.json"))

	body := `{"name":"俄罗斯节点","group":"欧洲","role":"dns","public_ip":"192.0.2.8","priority":1,"secret":"secret"}`
	request := httptest.NewRequest(http.MethodPost, "/enhancer/api/nodes", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	if valueString(controllerPayload["name"]) != "IP 192 0 2 8" || valueString(controllerPayload["group"]) != "" {
		t.Fatalf("controller received incompatible labels: %+v", controllerPayload)
	}
	var created map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created["name"] != "俄罗斯节点" || created["group"] != "欧洲" {
		t.Fatalf("display labels were not returned: %+v", created)
	}
}
