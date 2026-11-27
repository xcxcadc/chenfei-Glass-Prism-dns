package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultCatalogURL = "https://raw.githubusercontent.com/1-stream/1stream-public-utils/main/stream.smartdns.list"

func detectLocalPublicIPv4() string {
	if envIP := strings.TrimSpace(os.Getenv("PRISM_PANEL_IPV4")); envIP != "" {
		if parsed := net.ParseIP(envIP); parsed != nil && parsed.To4() != nil {
			return parsed.To4().String()
		}
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP == nil {
			continue
		}
		v4 := ipNet.IP.To4()
		if v4 == nil || !v4.IsGlobalUnicast() || v4.IsPrivate() || v4.IsLoopback() || v4.IsLinkLocalUnicast() {
			continue
		}
		return v4.String()
	}
	return ""
}

func main() {
	listen := flag.String("listen", "0.0.0.0:8080", "enhancer listen address")
	upstream := flag.String("upstream", "http://127.0.0.1:18080", "Prism Controller URL")
	dataDir := flag.String("data-dir", "/var/lib/prism-enhancer", "persistent data directory")
	distDir := flag.String("dist-dir", "/opt/prism/dist", "self-hosted node distribution directory")
	controllerDB := flag.String("controller-db", "/opt/prism/data.db", "Prism Controller SQLite database")
	catalogURL := flag.String("catalog-url", defaultCatalogURL, "SmartDNS catalog URL")
	flag.Parse()

	client := &http.Client{Timeout: 25 * time.Second}
	store, err := NewCustomServiceStore(filepath.Join(*dataDir, "custom-services.json"))
	if err != nil {
		log.Fatalf("initialize custom service store: %v", err)
	}
	ipStore, err := NewIPConfigStore(filepath.Join(*dataDir, "ip-configs.json"))
	if err != nil {
		log.Fatalf("initialize IP config store: %v", err)
	}
	transport, err := NewTransportStore(filepath.Join(*dataDir, "transports.json"))
	if err != nil {
		log.Fatalf("initialize transport registry: %v", err)
	}
	nodeLabels, err := NewNodeLabelStore(filepath.Join(*dataDir, "node-labels.json"))
	if err != nil {
		log.Fatalf("initialize node label store: %v", err)
	}
	branding, err := NewBrandingStore(filepath.Join(*dataDir, "branding.json"))
	if err != nil {
		log.Fatalf("initialize branding store: %v", err)
	}
	catalogPreferences, err := NewCatalogPreferenceStore(filepath.Join(*dataDir, "catalog-preferences.json"))
	if err != nil {
		log.Fatalf("initialize catalog preference store: %v", err)
	}
	catalog := NewCatalogManager(*catalogURL, client, store, catalogPreferences)
	app, err := NewApp(*upstream, catalog, store, ipStore, client, *controllerDB)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}
	app.nodeLabels = nodeLabels
	app.branding = branding
	app.transport = transport
	app.distDir = *distDir
	app.panelIPv4 = detectLocalPublicIPv4()
	app.nodeUpdates, err = newNodeUpdateStore(filepath.Join(*dataDir, "node-updates.json"))
	if err != nil {
		log.Fatalf("initialize node updates: %v", err)
	}
	app.mediaChecks, err = NewMediaCheckStore(filepath.Join(*dataDir, "media-checks.json"))
	if err != nil {
		log.Fatalf("initialize media checks: %v", err)
	}
	app.warpPolicies, err = NewWarpPolicyStore(filepath.Join(*dataDir, "warp-policies.json"))
	if err != nil {
		log.Fatalf("initialize warp policies: %v", err)
	}
	app.icons = newServiceIconCache(filepath.Join(*dataDir, "icon-cache"))
	catalogSnapshot := catalog.Snapshot(context.Background(), false)
	go func(services []Service) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		app.prewarmServiceIcons(ctx, services)
		log.Printf("service icon cache prewarm finished")
	}(catalogSnapshot.Services)

	server := &http.Server{
		Addr:              *listen,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	log.Printf("Prism 中文增强层监听 %s，Controller=%s", *listen, *upstream)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
