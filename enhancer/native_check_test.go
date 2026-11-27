package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestEvaluateNativeProbeResponse(t *testing.T) {
	gemini := Service{ID: "gemini", Name: "Gemini", Domains: []string{"gemini.google.com"}}
	st, reg, _ := evaluateNativeProbeResponse(gemini, 200, "https://gemini.google.com/", `... ,2,1,200,"NLD" ... "45631641\":false ...`, "")
	if st != "no" || reg != "NLD" {
		t.Fatalf("expected Gemini NLD false -> no/NLD, got %s/%s", st, reg)
	}
	st, reg, _ = evaluateNativeProbeResponse(gemini, 200, "https://gemini.google.com/", `... ,2,1,200,"SGP" ... "45631641\":true ...`, "")
	if st != "yes" || reg != "SGP" {
		t.Fatalf("expected Gemini SGP true -> yes/SGP, got %s/%s", st, reg)
	}

	claude := Service{ID: "claude", Name: "Claude", Domains: []string{"claude.ai"}}
	st, reg, _ = evaluateNativeProbeResponse(claude, 200, "https://claude.ai/cdn-cgi/trace", "ip=78.142.230.156\nloc=NL\n", "")
	if st != "yes" || reg != "NL" {
		t.Fatalf("expected Claude NL -> yes/NL, got %s/%s", st, reg)
	}
}

func TestClientNativeCheckLifecycle(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/nodes" && r.Header.Get("Authorization") == "Bearer admin-tok" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()

	dir := t.TempDir()
	customStore, _ := NewCustomServiceStore(filepath.Join(dir, "custom.json"))
	svc, _ := customStore.Upsert(Service{Name: "Claude", Category: "AI", Domains: []string{"claude.ai"}})
	ipStore, _ := NewIPConfigStore(filepath.Join(dir, "ips.json"))
	cfg, err := ipStore.Save(IPConfig{IP: "78.142.230.156", DNSNodeID: "dns-1", NodeSecret: "sec-1"}, "sec-1")
	if err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(upstream.URL, NewCatalogManager("http://127.0.0.1:1/unavailable", upstream.Client(), customStore), customStore, ipStore, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}

	// 1. Request native check
	reqBody, _ := json.Marshal(map[string]any{"config_id": cfg.ID, "service_id": svc.ID})
	req := httptest.NewRequest(http.MethodPost, "/enhancer/api/client-native-check", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer admin-tok")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"pending"`)) {
		t.Fatalf("expected pending response, got %d %s", rec.Code, rec.Body.String())
	}

	// 2. Pending task visible to client
	tasks := app.nativeChecks.PendingForClient(cfg.ID)
	if len(tasks) != 1 || tasks[0].Domain != "claude.ai" {
		t.Fatalf("unexpected pending tasks: %#v", tasks)
	}

	// 3. Client reports native check result
	repBody, _ := json.Marshal(map[string]any{
		"token":        cfg.EnrollmentToken,
		"task_id":      tasks[0].TaskID,
		"service_id":   svc.ID,
		"http_code":    200,
		"final_url":    "https://claude.ai/cdn-cgi/trace",
		"body_snippet": "ip=78.142.230.156\nloc=NL\n",
		"latency_ms":   120,
	})
	repReq := httptest.NewRequest(http.MethodPost, "/enhancer/api/client-native-check/report", bytes.NewReader(repBody))
	repRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(repRec, repReq)
	if repRec.Code != http.StatusOK {
		t.Fatalf("report failed: %d %s", repRec.Code, repRec.Body.String())
	}

	// 4. Subsequent GET returns ready + yes
	getReq := httptest.NewRequest(http.MethodGet, "/enhancer/api/client-native-check?config_id="+cfg.ID+"&service_id="+svc.ID, nil)
	getReq.Header.Set("Authorization", "Bearer admin-tok")
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK || !bytes.Contains(getRec.Body.Bytes(), []byte(`"status":"yes"`)) {
		t.Fatalf("expected ready yes result, got %d %s", getRec.Code, getRec.Body.String())
	}
}
