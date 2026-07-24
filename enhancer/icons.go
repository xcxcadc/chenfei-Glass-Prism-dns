package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const maxServiceIconBytes = 256 * 1024

type serviceIconAsset struct {
	contentType string
	body        []byte
}

type serviceIconCache struct {
	mu     sync.RWMutex
	assets map[string]serviceIconAsset
}

func newServiceIconCache() *serviceIconCache {
	return &serviceIconCache{assets: make(map[string]serviceIconAsset)}
}

func (cache *serviceIconCache) get(id string) (serviceIconAsset, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	asset, ok := cache.assets[id]
	return asset, ok
}

func (cache *serviceIconCache) put(id string, asset serviceIconAsset) {
	cache.mu.Lock()
	cache.assets[id] = asset
	cache.mu.Unlock()
}

func (app *App) handleServiceIcon(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	id := strings.TrimPrefix(request.URL.Path, "/enhancer/icons/")
	id = strings.TrimSuffix(id, ".png")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(writer, request)
		return
	}
	service, ok := app.catalog.Service(request.Context(), id)
	if !ok {
		http.NotFound(writer, request)
		return
	}
	asset, ok := app.icons.get(id)
	if !ok {
		asset = app.fetchServiceIcon(request.Context(), service)
		app.icons.put(id, asset)
	}
	writer.Header().Set("Content-Type", asset.contentType)
	writer.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(asset.body)
}

func (app *App) fetchServiceIcon(ctx context.Context, service Service) serviceIconAsset {
	domain := preferredProbeDomain(service)
	if domain != "" {
		iconURL, _ := url.Parse("https://www.google.com/s2/favicons")
		query := iconURL.Query()
		query.Set("domain_url", "https://"+domain)
		query.Set("sz", "64")
		iconURL.RawQuery = query.Encode()
		requestContext, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		request, err := http.NewRequestWithContext(requestContext, http.MethodGet, iconURL.String(), nil)
		if err == nil {
			request.Header.Set("User-Agent", "Prism-DNS-Enhancer/1.3")
			if response, requestErr := app.client.Do(request); requestErr == nil {
				defer response.Body.Close()
				contentType := response.Header.Get("Content-Type")
				if response.StatusCode >= 200 && response.StatusCode < 300 && strings.HasPrefix(contentType, "image/") {
					if body, readErr := io.ReadAll(io.LimitReader(response.Body, maxServiceIconBytes+1)); readErr == nil && len(body) > 0 && len(body) <= maxServiceIconBytes {
						return serviceIconAsset{contentType: contentType, body: body}
					}
				}
			}
		}
	}
	return serviceIconAsset{contentType: "image/svg+xml; charset=utf-8", body: fallbackServiceIcon(service)}
}

func fallbackServiceIcon(service Service) []byte {
	name := strings.TrimSpace(service.Name)
	initial, _ := utf8.DecodeRuneInString(name)
	if initial == utf8.RuneError || initial == 0 {
		initial = 'P'
	}
	sum := sha256.Sum256([]byte(service.ID))
	hue := int(sum[0]) * 360 / 255
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" rx="14" fill="hsl(%d 62%% 42%%)"/><text x="32" y="42" text-anchor="middle" font-family="Arial,sans-serif" font-size="32" font-weight="700" fill="white">%s</text></svg>`,
		hue,
		html.EscapeString(string(initial)),
	)
	return []byte(svg)
}
