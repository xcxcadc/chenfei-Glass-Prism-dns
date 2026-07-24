package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultCatalogURL = "https://raw.githubusercontent.com/1-stream/1stream-public-utils/main/stream.smartdns.list"

func main() {
	listen := flag.String("listen", "0.0.0.0:8080", "enhancer listen address")
	upstream := flag.String("upstream", "http://127.0.0.1:18080", "Prism Controller URL")
	dataDir := flag.String("data-dir", "/var/lib/prism-enhancer", "persistent data directory")
	catalogURL := flag.String("catalog-url", defaultCatalogURL, "SmartDNS catalog URL")
	flag.Parse()

	client := &http.Client{Timeout: 25 * time.Second}
	store, err := NewCustomServiceStore(filepath.Join(*dataDir, "custom-services.json"))
	if err != nil {
		log.Fatalf("initialize custom service store: %v", err)
	}
	catalog := NewCatalogManager(*catalogURL, client, store)
	app, err := NewApp(*upstream, catalog, store, client)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}

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
