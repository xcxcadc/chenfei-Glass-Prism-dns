package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	maxServiceIconBytes    = 256 * 1024
	fallbackIconCacheTTL   = 30 * time.Second
	serviceIconWorkerCount = 8
)

type serviceIconAsset struct {
	contentType string
	body        []byte
	fallback    bool
	expiresAt   time.Time
}

type storedServiceIconAsset struct {
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

type serviceIconCache struct {
	mu        sync.RWMutex
	assets    map[string]serviceIconAsset
	directory string
	locks     sync.Map
}

func newServiceIconCache(directories ...string) *serviceIconCache {
	cache := &serviceIconCache{assets: make(map[string]serviceIconAsset)}
	if len(directories) > 0 {
		cache.directory = strings.TrimSpace(directories[0])
		if cache.directory != "" {
			if err := os.MkdirAll(cache.directory, 0o755); err != nil {
				log.Printf("initialize service icon cache: %v", err)
				cache.directory = ""
			}
		}
	}
	return cache
}

func (cache *serviceIconCache) get(key string) (serviceIconAsset, bool) {
	cache.mu.RLock()
	asset, ok := cache.assets[key]
	cache.mu.RUnlock()
	if ok {
		if !asset.fallback || asset.expiresAt.IsZero() || time.Now().Before(asset.expiresAt) {
			return asset, true
		}
		cache.mu.Lock()
		if current, exists := cache.assets[key]; exists && current.fallback && current.expiresAt.Equal(asset.expiresAt) {
			delete(cache.assets, key)
		}
		cache.mu.Unlock()
	}
	if cache.directory == "" {
		return serviceIconAsset{}, false
	}
	body, err := os.ReadFile(cache.path(key))
	if err != nil {
		return serviceIconAsset{}, false
	}
	var stored storedServiceIconAsset
	if json.Unmarshal(body, &stored) != nil ||
		!strings.HasPrefix(stored.ContentType, "image/") ||
		len(stored.Body) == 0 ||
		len(stored.Body) > maxServiceIconBytes {
		return serviceIconAsset{}, false
	}
	asset = serviceIconAsset{contentType: stored.ContentType, body: stored.Body}
	cache.mu.Lock()
	cache.assets[key] = asset
	cache.mu.Unlock()
	return asset, true
}

func (cache *serviceIconCache) put(key string, asset serviceIconAsset) {
	cache.mu.Lock()
	cache.assets[key] = asset
	cache.mu.Unlock()
	if asset.fallback || cache.directory == "" {
		return
	}
	stored, err := json.Marshal(storedServiceIconAsset{ContentType: asset.contentType, Body: asset.body})
	if err != nil {
		return
	}
	path := cache.path(key)
	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, stored, 0o644); err != nil {
		log.Printf("persist service icon %s: %v", key, err)
		return
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		if !os.IsExist(err) {
			log.Printf("activate service icon %s: %v", key, err)
		}
	}
}

func (cache *serviceIconCache) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(cache.directory, hex.EncodeToString(sum[:])+".json")
}

func (cache *serviceIconCache) lock(key string) func() {
	value, _ := cache.locks.LoadOrStore(key, &sync.Mutex{})
	mutex := value.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
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
	asset := app.serviceIcon(request.Context(), service, false)
	if asset.fallback {
		writer.Header().Set("Cache-Control", "no-store, max-age=0")
		writer.Header().Set("X-Prism-Icon-Source", "fallback")
	} else {
		writer.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		writer.Header().Set("X-Prism-Icon-Source", "remote")
	}
	writer.Header().Set("Content-Type", asset.contentType)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(asset.body)
}

func (app *App) serviceIcon(ctx context.Context, service Service, requireAccurate bool) serviceIconAsset {
	key := serviceIconCacheKey(service)
	if asset, ok := app.icons.get(key); ok && (!requireAccurate || !asset.fallback) {
		return asset
	}
	unlock := app.icons.lock(key)
	defer unlock()
	if asset, ok := app.icons.get(key); ok && (!requireAccurate || !asset.fallback) {
		return asset
	}
	asset := app.fetchServiceIcon(ctx, service)
	app.icons.put(key, asset)
	return asset
}

func serviceIconCacheKey(service Service) string {
	return service.ID + "\x00" + preferredIconDomain(service)
}

func (app *App) prewarmServiceIcons(ctx context.Context, services []Service) {
	if len(services) == 0 {
		return
	}
	prioritized := make([]Service, 0, len(services))
	byID := make(map[string]Service, len(services))
	for _, service := range services {
		byID[service.ID] = service
	}
	seen := make(map[string]struct{}, len(services))
	for _, config := range app.ipStore.List() {
		for serviceID := range config.Routes {
			service, ok := byID[serviceID]
			if !ok {
				continue
			}
			if _, exists := seen[serviceID]; exists {
				continue
			}
			seen[serviceID] = struct{}{}
			prioritized = append(prioritized, service)
		}
	}
	for _, service := range services {
		if _, exists := seen[service.ID]; exists {
			continue
		}
		prioritized = append(prioritized, service)
	}

	for attempt := 0; attempt < 3; attempt++ {
		jobs := make(chan Service)
		var workers sync.WaitGroup
		for range serviceIconWorkerCount {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for service := range jobs {
					if ctx.Err() != nil {
						return
					}
					app.serviceIcon(ctx, service, true)
				}
			}()
		}
		for _, service := range prioritized {
			select {
			case <-ctx.Done():
				close(jobs)
				workers.Wait()
				return
			case jobs <- service:
			}
		}
		close(jobs)
		workers.Wait()
		if attempt < 2 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(attempt+1) * time.Second):
			}
		}
	}
}

func (app *App) fetchServiceIcon(ctx context.Context, service Service) serviceIconAsset {
	domains := []string{preferredIconDomain(service), preferredProbeDomain(service)}
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}
		iconURL, _ := url.Parse("https://www.google.com/s2/favicons")
		query := iconURL.Query()
		query.Set("domain_url", "https://"+domain)
		query.Set("sz", "64")
		iconURL.RawQuery = query.Encode()
		candidateURLs := append([]string(nil), preferredIconURLs(service)...)
		candidateURLs = append(candidateURLs, iconURL.String())
		for _, candidateURL := range candidateURLs {
			if asset, ok := app.fetchIconURL(ctx, candidateURL); ok {
				return asset
			}
		}
	}
	return serviceIconAsset{
		contentType: "image/svg+xml; charset=utf-8",
		body:        fallbackServiceIcon(service),
		fallback:    true,
		expiresAt:   time.Now().Add(fallbackIconCacheTTL),
	}
}

func (app *App) fetchIconURL(ctx context.Context, candidateURL string) (serviceIconAsset, bool) {
	requestContext, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, candidateURL, nil)
	if err != nil {
		return serviceIconAsset{}, false
	}
	request.Header.Set("User-Agent", "Prism-DNS-Enhancer/1.4.5")
	allowedHost := request.URL.Hostname()
	client := *app.client
	client.CheckRedirect = func(next *http.Request, previous []*http.Request) error {
		if len(previous) >= 3 ||
			next.URL.Scheme != "https" ||
			!strings.EqualFold(next.URL.Hostname(), allowedHost) {
			return http.ErrUseLastResponse
		}
		return nil
	}
	response, err := client.Do(request)
	if err != nil {
		return serviceIconAsset{}, false
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return serviceIconAsset{}, false
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxServiceIconBytes+1))
	if err != nil || len(body) == 0 || len(body) > maxServiceIconBytes {
		return serviceIconAsset{}, false
	}
	contentType := response.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		contentType = http.DetectContentType(body)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return serviceIconAsset{}, false
	}
	return serviceIconAsset{contentType: contentType, body: body}, true
}

func preferredIconDomain(service Service) string {
	preferred := map[string]string{
		"AU:9 now":                        "9now.com.au",
		"AU:ABC iView":                    "abc.net.au",
		"AU:Optus":                        "optus.com.au",
		"AU:SBS on Demand":                "sbs.com.au",
		"Amazon Prime Video":              "primevideo.com",
		"AnimeFesta":                      "animefesta.iowl.jp",
		"Apple TV+":                       "tv.apple.com",
		"Bilibili":                        "bilibili.com",
		"CBC Gem":                         "gem.cbc.ca",
		"ChatGPT / OpenAI":                "chatgpt.com",
		"Claude":                          "claude.ai",
		"Crave TV":                        "crave.ca",
		"DAZN":                            "dazn.com",
		"Directv Stream":                  "directv.com",
		"DirecTV":                         "directv.com",
		"Discovery+":                      "discoveryplus.com",
		"Disney+":                         "disneyplus.com",
		"EU:SkyShowtime":                  "skyshowtime.com",
		"Fuji TV":                         "www.fujitv.co.jp",
		"GB:BBC":                          "bbc.co.uk",
		"GB:Discovery+":                   "discoveryplus.com",
		"Games":                           "umamusume.jp",
		"Gemini":                          "gemini.google.com",
		"Google AI Studio":                "aistudio.google.com",
		"Grok":                            "grok.com",
		"Hami Video":                      "hamivideo.hinet.net",
		"HBO / Max":                       "max.com",
		"ID:Vidio":                        "vidio.com",
		"IN:Jio Cinema":                   "jiohotstar.com",
		"IT:RaiPlay":                      "raiplay.it",
		"KKTV":                            "kktv.me",
		"Microsoft Copilot Image Creator": "copilot.microsoft.com",
		"Music.jp":                        "music-book.jp",
		"NZ:SkyGO NZ":                     "sky.co.nz",
		"NetEase Cloud Music":             "music.163.com",
		"Netflix":                         "netflix.com",
		"NicoNico":                        "nicovideo.jp",
		"Reads":                           "magazine.rakuten.co.jp",
		"Setanta Sports":                  "setantasports.com",
		"Spotify":                         "spotify.com",
		"Suno":                            "suno.com",
		"TH:AIS Play":                     "ais.th",
		"TikTok":                          "tiktok.com",
		"U-NEXT":                          "video.unext.jp",
		"UA:MEGOGO":                       "megogo.net",
		"VN:K+":                           "www.kplus.vn",
		"Viaplay":                         "viaplay.com",
		"Viu":                             "viu.com",
		"Wavve":                           "wavve.com",
		"X":                               "x.com",
		"YouTube":                         "youtube.com",
		"Youku":                           "youku.com",
	}
	if domain := preferred[service.Name]; domain != "" {
		return domain
	}
	return preferredProbeDomain(service)
}

func preferredIconURLs(service Service) []string {
	preferred := map[string][]string{
		"CBC Gem": {"https://gem.cbc.ca/favicon.ico"},
		"Grok":    {"https://x.ai/favicon.ico"},
		"VN:K+":   {"https://www.kplus.vn/logo-kplus.svg"},
		"X":       {"https://x.com/favicon.ico"},
	}
	return append([]string(nil), preferred[service.Name]...)
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
