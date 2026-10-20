package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestMediaCheckHTTPAuthenticationAndGeneration(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer admin" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeJSON(writer, http.StatusOK, []any{})
	}))
	defer upstream.Close()
	directory := t.TempDir()
	databasePath := filepath.Join(directory, "controller.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TABLE nodes (id TEXT PRIMARY KEY, role TEXT, secret TEXT, public_ip TEXT, address TEXT);
		INSERT INTO nodes VALUES ('proxy-a','proxy','proxy-secret','198.51.100.1',''),('dns-a','client','client-secret','203.0.113.1','')`); err != nil {
		t.Fatal(err)
	}
	database.Close()
	custom, _ := NewCustomServiceStore(filepath.Join(directory, "services.json"))
	service, err := custom.Upsert(Service{Name: "OpenAI", Domains: []string{"chatgpt.com"}})
	if err != nil {
		t.Fatal(err)
	}
	ipStore, _ := NewIPConfigStore(filepath.Join(directory, "ips.json"))
	config, err := ipStore.Save(IPConfig{IP: "203.0.113.1", DNSNodeID: "dns-a", Routes: map[string]string{service.ID: "proxy-a"}}, "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager(upstream.URL, upstream.Client(), custom)
	app, err := NewApp(upstream.URL, catalog, custom, ipStore, upstream.Client(), databasePath)
	if err != nil {
		t.Fatal(err)
	}
	invoke := func(method, path string, body any, authorization string) *httptest.ResponseRecorder {
		encoded, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		request := httptest.NewRequest(method, path, bytes.NewReader(encoded))
		request.Header.Set("Authorization", authorization)
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, request)
		return response
	}
	for _, path := range []string{"/enhancer/api/media-checks?config_id=" + config.ID, "/enhancer/api/media-checks/unknown"} {
		if response := invoke(http.MethodGet, path, nil, ""); response.Code != 401 {
			t.Fatal("unauthorized read accepted", response.Code)
		}
	}
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks", map[string]string{"config_id": config.ID}, ""); response.Code != 401 {
		t.Fatal("unauthorized create accepted")
	}
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks", map[string]string{"node_id": "dns-a"}, "Bearer admin"); response.Code != 404 {
		t.Fatal("client node accepted as proxy")
	}
	response := invoke(http.MethodPost, "/enhancer/api/media-checks", map[string]string{"config_id": config.ID}, "Bearer admin")
	if response.Code != 202 {
		t.Fatal(response.Code, response.Body.String())
	}
	var job MediaCheckJob
	json.Unmarshal(response.Body.Bytes(), &job)
	worker := map[string]any{"role": "client", "credential": "wrong", "action": "claim"}
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks/worker", worker, ""); response.Code != 401 {
		t.Fatal("invalid worker credential accepted")
	}
	worker["credential"] = config.EnrollmentToken
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks/worker", worker, ""); response.Code != 200 {
		t.Fatal("claim failed", response.Body.String())
	}
	worker["action"], worker["job_id"], worker["status"], worker["sequence"] = "report", job.ID, "running", 1
	worker["results"] = map[string]string{"ChatGPT": "Yes"}
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks/worker", worker, ""); response.Code != 200 {
		t.Fatal("report failed", response.Body.String())
	}
	stored, _ := ipStore.Record(config.ID)
	if stored.ServiceResults[service.ID] != "Yes" {
		t.Fatal("raw official result did not reach service view", stored.ServiceResults)
	}
	legacy := map[string]any{"token": config.EnrollmentToken, "scope": "unlock_services", "service_audit_requested_at": mediaGeneration(stored), "results": map[string]string{service.ID: "No"}}
	if response := invoke(http.MethodPost, "/enhancer/api/audit/report", legacy, ""); response.Code != 409 {
		t.Fatal("legacy audit overwrote official worker")
	}
	if _, err := ipStore.RequestServiceAudit(config.ID); err != nil {
		t.Fatal(err)
	}
	worker["sequence"] = 2
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks/worker", worker, ""); response.Code != 409 {
		t.Fatal("old configuration result accepted", response.Code)
	}
	worker["role"], worker["credential"] = "proxy", "proxy-secret"
	if response := invoke(http.MethodPost, "/enhancer/api/media-checks/worker", worker, ""); response.Code != 404 {
		t.Fatal("proxy could overwrite client result")
	}
}

func TestMediaCheckRestartPreservesPartialResults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checks.json")
	store, _ := NewMediaCheckStore(path)
	job, _ := store.Create("proxy", "", "")
	store.Claim("proxy", "")
	store.Report(job.ID, "proxy", "", "running", 1, map[string]string{"Netflix": "Yes (Region: SG)"}, "")
	restored, err := NewMediaCheckStore(path)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := restored.Get(job.ID)
	if got.Status != "failed" || got.Results["Netflix"] != "Yes (Region: SG)" || got.Error == "" {
		t.Fatal("restart lost progress or left task spinning", got)
	}
}

func TestMediaCheckStreamsPreservesRawResults(t *testing.T) {
	store, err := NewMediaCheckStore(filepath.Join(t.TempDir(), "checks.json"))
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.Create("proxy-a", "", "")
	if err != nil {
		t.Fatal(err)
	}
	again, _ := store.Create("proxy-a", "", "")
	if job.ID != again.ID {
		t.Fatal("duplicate active job")
	}
	claimed, err := store.Claim("proxy-a", "")
	if err != nil || claimed == nil {
		t.Fatal("job not claimed", err)
	}
	if duplicate, _ := store.Claim("proxy-a", ""); duplicate != nil {
		t.Fatal("claimed twice")
	}
	partial, err := store.Report(job.ID, "proxy-a", "", "running", 1, map[string]string{"Netflix": "Yes (Region: SG)"}, "")
	if err != nil || partial.Status != "running" || partial.Results["Netflix"] != "Yes (Region: SG)" {
		t.Fatal(partial, err)
	}
	partial.Results["Netflix"] = "wrong"
	if _, err := store.Report(job.ID, "proxy-b", "", "running", 2, nil, ""); err == nil {
		t.Fatal("cross-node report accepted")
	}
	if _, err := store.Report(job.ID, "proxy-a", "", "running", 1, nil, ""); !errors.Is(err, errMediaConflict) {
		t.Fatal("old sequence accepted")
	}
	if _, err := store.Report(job.ID, "proxy-a", "", "running", 2, map[string]string{"IPv6": "Yes"}, ""); err == nil {
		t.Fatal("IPv6 accepted")
	}
	final, err := store.Report(job.ID, "proxy-a", "", "completed", 3, map[string]string{"Claude": "No (Banned by WAF)"}, "")
	if err != nil || final.Results["Netflix"] != "Yes (Region: SG)" || final.Results["Claude"] != "No (Banned by WAF)" || final.FinishedAt == nil {
		t.Fatal(final, err)
	}
	if _, err := store.Report(job.ID, "proxy-a", "", "running", 4, nil, ""); !errors.Is(err, errMediaConflict) {
		t.Fatal("finished job overwritten")
	}
	reloaded, err := NewMediaCheckStore(store.path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reloaded.Get(job.ID)
	if !ok || got.Status != "completed" || got.Results["Claude"] != final.Results["Claude"] {
		t.Fatal(got)
	}
}

func TestMediaCheckDoesNotCutOffAt30Seconds(t *testing.T) {
	store, _ := NewMediaCheckStore("")
	job, _ := store.Create("", "client-a", "generation-a")
	store.Claim("", "client-a")
	stored := store.jobs[job.ID]
	stored.RequestedAt = time.Now().Add(-4 * time.Minute)
	stored.UpdatedAt = time.Now().Add(-25 * time.Second)
	store.jobs[job.ID] = stored
	got, _ := store.Get(job.ID)
	if got.Status != "running" {
		t.Fatal("active worker timed out", got)
	}
	stored.UpdatedAt = time.Now().Add(-100 * time.Second)
	store.jobs[job.ID] = stored
	got, _ = store.Get(job.ID)
	if got.Status != "failed" || got.Error == "" {
		t.Fatal("dead worker not reported", got)
	}
	newJob, _ := store.Create("", "client-a", "generation-b")
	if newJob.ID == job.ID {
		t.Fatal("new generation reused stale job")
	}
}

func TestMediaCheckSnapshotConcurrentReports(t *testing.T) {
	store, _ := NewMediaCheckStore("")
	job, _ := store.Create("proxy", "", "")
	store.Claim("proxy", "")
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		for sequence := 1; sequence < 50; sequence++ {
			store.Report(job.ID, "proxy", "", "running", sequence, map[string]string{"ChatGPT": "Yes"}, "")
		}
	}()
	go func() {
		defer workers.Done()
		for count := 0; count < 50; count++ {
			snapshot, _ := store.Get(job.ID)
			snapshot.Results["ChatGPT"] = "No"
		}
	}()
	workers.Wait()
	got, _ := store.Get(job.ID)
	if got.Results["ChatGPT"] != "Yes" {
		t.Fatal("snapshot mutated store")
	}
}

func TestMediaResultNoProbeOverride(t *testing.T) {
	service := Service{Name: "OpenAI", Domains: []string{"chatgpt.com"}}
	job := MediaCheckJob{Status: "running", Results: map[string]string{"ChatGPT": "Yes"}}
	if got := mediaServiceResult(service, job); got != "Yes" {
		t.Fatal(got)
	}
	job.Results["ChatGPT"] = "No (Unsupported Country)"
	if got := mediaServiceResult(service, job); got != "No (Unsupported Country)" {
		t.Fatal(got)
	}
	if got := mediaServiceResult(Service{Name: "Unknown service"}, job); got != "NOT COVERED (check.unlock.media IPv4 / menu 0)" {
		t.Fatal(got)
	}
	if got := mediaServiceResult(Service{Name: "Claude"}, MediaCheckJob{Status: "running", Results: map[string]string{"Claude": "Yes (Region: SG)"}}); got != "Yes (Region: SG)" {
		t.Fatal(got)
	}
}
