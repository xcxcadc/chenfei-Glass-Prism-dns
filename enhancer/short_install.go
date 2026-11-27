package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const shortInstallRepo = "xcxcadc/chenfei-Glass-Prism-dns"

var shortInstallTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{4,128}$`)

func requestPublicOrigin(request *http.Request) string {
	scheme := "https"
	if proto := strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-Proto"), ",")[0]); proto == "http" || proto == "https" {
		scheme = proto
	} else if request.TLS == nil && (strings.HasPrefix(request.Host, "127.0.0.1") || strings.HasPrefix(request.Host, "localhost")) {
		scheme = "http"
	}
	host := strings.TrimSpace(request.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(request.Host)
	}
	if host == "" {
		return "https://localhost"
	}
	return scheme + "://" + host
}

func (app *App) handleShortInstall(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		methodNotAllowed(writer, http.MethodGet, http.MethodHead)
		return
	}
	cred := strings.TrimPrefix(request.URL.Path, "/enhancer/i/")
	if cred == request.URL.Path {
		cred = strings.TrimPrefix(request.URL.Path, "/i/")
	}
	cred = strings.Trim(strings.TrimSpace(cred), "/")
	if !shortInstallTokenPattern.MatchString(cred) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("echo '[ERROR] Invalid Prism install token' >&2; exit 1\n"))
		return
	}

	origin := requestPublicOrigin(request)
	var script string

	// 1. Check managed IP config enrollment token (client / 被解锁机)
	if app.ipStore != nil {
		for _, cfg := range app.ipStore.Records() {
			if cfg.EnrollmentToken != "" && cfg.EnrollmentToken == cred {
				script = fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then exec sudo -E bash -c 'curl -fsSL %q | bash'; fi
  echo "[ERROR] Please run as root (or with sudo)." >&2; exit 1
fi
TMP=$(mktemp /tmp/prism-client.XXXXXX.sh)
trap 'rm -f "$TMP"' EXIT
curl -fsSL --connect-timeout 8 --max-time 60 %q -o "$TMP" 2>/dev/null || \
curl -fsSL --connect-timeout 10 --max-time 60 %q -o "$TMP"
bash "$TMP" --master %q --token %q --one-click --non-interactive
`, origin+"/i/"+cred, origin+"/enhancer/dist/prismdns.sh", "https://raw.githubusercontent.com/"+shortInstallRepo+"/refs/tags/"+prismSourceRef+"/prismdns.sh", origin, cred)
				break
			}
		}
	}

	// 2. Check controller node secret (proxy / manual DNS node)
	if script == "" && app.controllerDB != "" {
		if node, err := controllerNodeBySecret(app.controllerDB, cred); err == nil && node.ID != "" {
			extraArgs := ""
			if node.Role == "proxy" {
				if ip := firstNodeIPv4(node); ip != "" {
					extraArgs = fmt.Sprintf(" --ip %q", ip)
				}
			}
			script = fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then exec sudo -E bash -c 'curl -fsSL %q | bash'; fi
  echo "[ERROR] Please run as root (or with sudo)." >&2; exit 1
fi
TMP=$(mktemp /tmp/prism-agent.XXXXXX.sh)
trap 'rm -f "$TMP"' EXIT
curl -fsSL --connect-timeout 8 --max-time 60 %q -o "$TMP" 2>/dev/null || \
curl -fsSL --connect-timeout 10 --max-time 60 %q -o "$TMP"
bash "$TMP" --master %q --secret %q%s
`, origin+"/i/"+cred, origin+"/enhancer/dist/agent_install.sh", "https://raw.githubusercontent.com/"+shortInstallRepo+"/refs/tags/"+prismSourceRef+"/agent_install.sh", origin, cred, extraArgs)
		}
	}

	if script == "" {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("echo '[ERROR] Prism install token not found or expired' >&2; exit 1\n"))
		return
	}

	setNoCacheHeaders(writer)
	writer.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if request.Method == http.MethodGet {
		_, _ = writer.Write([]byte(script))
	}
}
