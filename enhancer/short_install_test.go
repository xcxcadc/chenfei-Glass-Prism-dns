package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleShortInstallClientToken(t *testing.T) {
	dir := t.TempDir()
	ipStore, err := NewIPConfigStore(filepath.Join(dir, "ips.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ipStore.Save(IPConfig{
		IP:         "203.0.113.10",
		Note:       "Demo",
		DNSNodeID:  "dns-demo-1",
		NodeSecret: "demosecret123",
		Routes:     map[string]string{"gemini": "proxy-1"},
	}, "demosecret123")
	if err != nil {
		t.Fatal(err)
	}

	app, err := NewApp("http://127.0.0.1:18080", NewCatalogManager("http://127.0.0.1:1/unavailable", &http.Client{}, nil), nil, ipStore, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://dns.example.com/i/"+cfg.EnrollmentToken, nil)
	req.Host = "dns.example.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "prismdns.sh") || !strings.Contains(body, cfg.EnrollmentToken) || !strings.Contains(body, "https://dns.example.com") {
		t.Fatalf("unexpected short install script:\n%s", body)
	}

	badReq := httptest.NewRequest(http.MethodGet, "https://dns.example.com/i/nonexistent-token-999", nil)
	badRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown token, got %d", badRec.Code)
	}
}
