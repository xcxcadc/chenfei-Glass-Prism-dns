package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func newReleaseTestHandler(t *testing.T) http.Handler {
	t.Helper()
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		t.Errorf("unexpected upstream request: %s", request.URL)
		return nil, fmt.Errorf("network disabled in release asset tests")
	})}
	store, err := NewCustomServiceStore(filepath.Join(t.TempDir(), "services.json"))
	if err != nil {
		t.Fatal(err)
	}
	ipStore, err := NewIPConfigStore(filepath.Join(t.TempDir(), "ip-configs.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalogManager("https://catalog.invalid/list", client, store)
	app, err := NewApp("https://controller.invalid", catalog, store, ipStore, client)
	if err != nil {
		t.Fatal(err)
	}
	return app.Handler()
}

func TestReleaseAssetsUsePublishedTag(t *testing.T) {
	if uiVersion != "1.5.30" || prismSourceRef != "v"+uiVersion {
		t.Fatalf("release version mismatch: UI %q, source %q", uiVersion, prismSourceRef)
	}
	handler := newReleaseTestHandler(t)
	wantCSP := "default-src 'self'; connect-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; font-src 'self' data: https://cdn.jsdelivr.net; script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"
	for _, path := range []string{"/", "/assets/app.js", "/assets/app.js?v=" + uiVersion, "/assets/app.js?v=1.5.28", "/assets/app.js?v=1.5.29"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("unexpected status: %d", response.Code)
			}
			if response.Header().Get("X-Prism-UI-Version") != uiVersion {
				t.Fatalf("inconsistent UI version header: %q", response.Header().Get("X-Prism-UI-Version"))
			}
			if cache := response.Header().Get("Cache-Control"); !strings.Contains(cache, "no-store") || !strings.Contains(cache, "no-cache") {
				t.Fatalf("weak cache control: %q", cache)
			}
			if response.Header().Get("Content-Security-Policy") != wantCSP {
				t.Fatalf("CSP changed: %q", response.Header().Get("Content-Security-Policy"))
			}
			body := response.Body.String()
			for _, stale := range []string{"__PRISM_", "v1.5.28", "v1.5.29", "?v=1.5.28", "?v=1.5.29"} {
				if strings.Contains(body, stale) {
					t.Errorf("response contains stale ref or placeholder %q", stale)
				}
			}
			if path == "/" {
				for _, asset := range []string{"/assets/app.js?v=" + uiVersion, "/assets/styles.css?v=" + uiVersion} {
					if !strings.Contains(body, asset) {
						t.Errorf("HTML is missing versioned asset %q", asset)
					}
				}
				return
			}
			if response.Header().Get("Content-Type") != "text/javascript; charset=utf-8" {
				t.Fatalf("unexpected JavaScript content type: %q", response.Header().Get("Content-Type"))
			}
			for _, fragment := range []string{
				`const prismSourceRef = "` + prismSourceRef + `";`,
				"https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/${prismSourceRef}/${file}",
				`curl -fsSL ${prismRawURL("agent_install.sh")} | bash -s -- --master ${location.origin}`,
			} {
				if !strings.Contains(body, fragment) {
					t.Errorf("served install command source is missing %q", fragment)
				}
			}
		})
	}
	if got, want := prismRawURL("agent_install.sh"), "https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/"+prismSourceRef+"/agent_install.sh"; got != want {
		t.Fatalf("backend install URL = %q, want %q", got, want)
	}
}

func TestReleaseAppScriptHandlerMethods(t *testing.T) {
	handler := newReleaseTestHandler(t)
	getResponse := httptest.NewRecorder()
	handler.ServeHTTP(getResponse, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	headResponse := httptest.NewRecorder()
	handler.ServeHTTP(headResponse, httptest.NewRequest(http.MethodHead, "/assets/app.js", nil))
	if headResponse.Code != http.StatusOK || headResponse.Body.Len() != 0 {
		t.Fatalf("unexpected HEAD response: status %d, body size %d", headResponse.Code, headResponse.Body.Len())
	}
	if got, want := headResponse.Header().Get("Content-Length"), fmt.Sprint(getResponse.Body.Len()); got != want {
		t.Fatalf("HEAD content length = %q, want %q", got, want)
	}
	postResponse := httptest.NewRecorder()
	handler.ServeHTTP(postResponse, httptest.NewRequest(http.MethodPost, "/assets/app.js", nil))
	if postResponse.Code != http.StatusMethodNotAllowed || postResponse.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("unexpected POST response: status %d, Allow %q", postResponse.Code, postResponse.Header().Get("Allow"))
	}
	for _, path := range []string{"/assets/styles.css", "/assets/app.js/extra"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		wantStatus := http.StatusOK
		if strings.HasSuffix(path, "/extra") {
			wantStatus = http.StatusNotFound
		}
		if response.Code != wantStatus {
			t.Errorf("%s status = %d, want %d", path, response.Code, wantStatus)
		}
	}
}

func TestReleaseMediaUIKeepsFailureDetailsAndRawResults(t *testing.T) {
	response := httptest.NewRecorder()
	newReleaseTestHandler(t).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	for _, fragment := range []string{
		`mediaNotStarted: "未启动"`,
		`mediaExecutionFailed: "执行失败"`,
		`mediaUnverified: "服务可用性未确认"`,
		"检测任务失败不代表所有服务不可用",
		`"Worker has not responded for 90 seconds; check its connection and timer":"媒体检测 Worker 超过 90 秒未响应`,
		"原始错误：${raw}",
		"element.title = nodeMediaDetail(node);",
		"title=\"${escapeHTML(status.tooltip || status.detail)}\"",
		`result.textContent = String(raw ?? "")`,
		`card.querySelector(".media-new").addEventListener("click", () => syncMediaCheck(record, "create"))`,
		`card.querySelector(".media-reconnect").addEventListener("click", () => syncMediaCheck(record))`,
	} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Errorf("served media UI source is missing %q", fragment)
		}
	}
}
