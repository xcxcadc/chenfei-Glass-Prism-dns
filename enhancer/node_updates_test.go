package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNodeUpdateLifecycle(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/nodes" || request.Header.Get("Authorization") != "Bearer admin" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeJSON(writer, http.StatusOK, []map[string]string{{"id": "one", "role": "proxy"}, {"id": "two", "role": "proxy"}, {"id": "three", "role": "dns"}})
	}))
	t.Cleanup(upstream.Close)
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "controller.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(`CREATE TABLE nodes (id TEXT, role TEXT, secret TEXT, public_ip TEXT, address TEXT);
		INSERT INTO nodes VALUES ('one','proxy','proxy-one','',''),('two','proxy','proxy-two','',''),('three','dns','client-secret','','')`)
	_ = database.Close()
	if err != nil {
		t.Fatal(err)
	}
	custom, _ := NewCustomServiceStore("")
	ips, err := NewIPConfigStore(filepath.Join(directory, "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	client, err := ips.Save(IPConfig{IP: "203.0.113.11", DNSNodeID: "three", Routes: map[string]string{"service": "one"}}, "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	immutable := false
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host == "api.github.com" {
			if request.Method != http.MethodGet || request.URL.Path != "/repos/xcxcadc/chenfei-Glass-Prism-dns/releases/tags/"+prismSourceRef || request.Header.Get("Authorization") != "" {
				t.Fatalf("invalid release lookup: %s %s", request.Method, request.URL)
			}
			body, _ := json.Marshal(map[string]any{"tag_name": prismSourceRef, "immutable": immutable, "draft": false, "prerelease": false})
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
		}
		return http.DefaultTransport.RoundTrip(request)
	})}
	app, err := NewApp(upstream.URL, NewCatalogManager(upstream.URL, upstream.Client(), custom), custom, ips, httpClient, databasePath)
	if err != nil {
		t.Fatal(err)
	}
	jobFile := filepath.Join(directory, "node-updates.json")
	app.nodeUpdates, err = newNodeUpdateStore(jobFile)
	if err != nil {
		t.Fatal(err)
	}
	handler := app.Handler()
	call := func(method, path, token, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, response.Code, status, response.Body.String())
		}
		return response
	}
	const workerPath = "/enhancer/api/node-updates/worker"
	const queuePath = "/enhancer/api/node-updates"
	const releasePath = "/enhancer/api/node-release"
	const poll = `{"action":"poll","role":"proxy","credential":"proxy-one","version":"v1.5.36","transport_version":"2.5.2","agent_version":"v1.2.1","ipv4":"203.0.113.1"}`
	clientPoll := `{"action":"poll","role":"client","credential":"` + client.EnrollmentToken + `","version":"v1.5.36","agent_version":"v1.3"}`
	queue := `{"node_id":"one","version":"` + prismSourceRef + `"}`
	call(http.MethodPost, queuePath, "", queue, 401)
	call(http.MethodGet, releasePath, "", "", 401)
	call(http.MethodGet, releasePath, "admin", "", 503)
	call(http.MethodPost, queuePath, "admin", queue, 503)
	immutable = true
	app.nodeUpdates.releaseUntil = time.Time{}
	call(http.MethodGet, releasePath, "admin", "", 200)
	call(http.MethodPost, queuePath, "admin", queue, 409)
	call(http.MethodPost, workerPath, "", strings.Replace(poll, "proxy-one", "invalid", 1), 401)
	call(http.MethodPost, workerPath, "", poll, 200)
	call(http.MethodPost, workerPath, "", clientPoll, 200)
	call(http.MethodPost, queuePath, "admin", strings.Replace(queue, prismSourceRef, "v1.5.35", 1), 400)
	response := call(http.MethodPost, queuePath, "admin", queue, 202)
	var job nodeUpdate
	if err := json.Unmarshal(response.Body.Bytes(), &job); err != nil || job.ID == "" || job.Status != "queued" {
		t.Fatalf("unexpected queued job: %s", response.Body.String())
	}
	other := call(http.MethodPost, workerPath, "", strings.Replace(poll, "proxy-one", "proxy-two", 1), 200)
	if !strings.Contains(other.Body.String(), `"job":null`) {
		t.Fatal("another node must not receive the update")
	}
	app.nodeUpdates, err = newNodeUpdateStore(jobFile)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		response = call(http.MethodPost, workerPath, "", poll, 200)
		if !strings.Contains(response.Body.String(), job.ID) || !strings.Contains(response.Body.String(), `"status":"running"`) {
			t.Fatalf("unexpected job claim: %s", response.Body.String())
		}
	}
	report := `{"action":"report","role":"proxy","credential":"proxy-one","id":"` + job.ID + `","version":"` + prismSourceRef + `","status":"succeeded"}`
	call(http.MethodPost, workerPath, "", strings.Replace(report, "proxy-one", "proxy-two", 1), 409)
	call(http.MethodPost, workerPath, "", report, 200)
	call(http.MethodPost, workerPath, "", report, 200)
	call(http.MethodPost, queuePath, "admin", queue, 409)
	response = call(http.MethodGet, "/enhancer/api/nodes", "admin", "", 200)
	body := response.Body.String()
	for _, expected := range []string{`"version":"` + prismSourceRef + `"`, `"status":"succeeded"`, `"ipv4":"203.0.113.1"`, `"agent_version":"v1.2.1"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("inventory missing %s", expected)
		}
	}
	if strings.Contains(body, "proxy-one") || strings.Contains(body, "binding") {
		t.Fatal("private management data leaked")
	}
	info := app.nodeUpdates.inventory["one"]
	info.LastSeen = time.Now().Add(-2 * nodeHeartbeatTTL)
	app.nodeUpdates.inventory["one"] = info
	response = call(http.MethodGet, "/enhancer/api/nodes", "admin", "", 200)
	if !strings.Contains(response.Body.String(), `"update_supported":false`) {
		t.Fatal("expired worker heartbeat incorrectly shown as available")
	}
	call(http.MethodPost, workerPath, "", clientPoll, 200)
	clientQueue := `{"node_id":"three","version":"` + prismSourceRef + `"}`
	response = call(http.MethodPost, queuePath, "admin", clientQueue, 202)
	var clientJob nodeUpdate
	if err := json.Unmarshal(response.Body.Bytes(), &clientJob); err != nil || clientJob.ID == "" {
		t.Fatalf("unexpected queued client job: %s", response.Body.String())
	}
	response = call(http.MethodPost, workerPath, "", clientPoll, 200)
	if !strings.Contains(response.Body.String(), clientJob.ID) || !strings.Contains(response.Body.String(), `"status":"running"`) {
		t.Fatalf("unexpected client job claim: %s", response.Body.String())
	}
	clientReport := `{"action":"report","role":"client","credential":"` + client.EnrollmentToken + `","id":"` + clientJob.ID + `","version":"` + prismSourceRef + `","status":"failed"}`
	call(http.MethodPost, workerPath, "", clientReport, 200)
	call(http.MethodPost, workerPath, "", clientReport, 200)
	clientInfo := app.nodeUpdates.inventory["three"]
	if clientInfo.Version != "v1.5.36" || clientInfo.AgentVersion != "v1.3" {
		t.Fatalf("failed update changed the reported client version: %#v", clientInfo)
	}
	response = call(http.MethodPost, queuePath, "admin", clientQueue, 202)
	if strings.Contains(response.Body.String(), clientJob.ID) {
		t.Fatal("retry reused an already failed update")
	}
}
