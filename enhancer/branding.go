package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type BrandingSettings struct {
	SiteName     string `json:"site_name"`
	BrowserTitle string `json:"browser_title"`
	SiteTagline  string `json:"site_tagline"`
}

type BrandingStore struct {
	mu       sync.RWMutex
	path     string
	settings BrandingSettings
}

func NewBrandingStore(path string) (*BrandingStore, error) {
	store := &BrandingStore{path: path}
	if path != "" {
		if err := store.load(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (store *BrandingStore) Get() BrandingSettings {
	if store == nil {
		return BrandingSettings{}
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.settings
}

func (store *BrandingStore) Save(settings BrandingSettings) (BrandingSettings, error) {
	if store == nil {
		return BrandingSettings{}, errors.New("branding store is unavailable")
	}
	settings.SiteName = strings.TrimSpace(settings.SiteName)
	settings.BrowserTitle = strings.TrimSpace(settings.BrowserTitle)
	settings.SiteTagline = strings.TrimSpace(settings.SiteTagline)
	if err := validateBranding(settings); err != nil {
		return BrandingSettings{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.settings = settings
	if err := store.saveLocked(); err != nil {
		return BrandingSettings{}, err
	}
	return store.settings, nil
}

func validateBranding(settings BrandingSettings) error {
	if settings.SiteName == "" {
		return errors.New("site name is required")
	}
	if settings.BrowserTitle == "" {
		return errors.New("browser title is required")
	}
	for _, field := range []struct {
		name  string
		value string
		limit int
	}{
		{name: "site name", value: settings.SiteName, limit: 48},
		{name: "browser title", value: settings.BrowserTitle, limit: 96},
		{name: "site tagline", value: settings.SiteTagline, limit: 120},
	} {
		if utf8.RuneCountInString(field.value) > field.limit {
			return fmt.Errorf("%s is too long", field.name)
		}
		for _, character := range field.value {
			if unicode.IsControl(character) {
				return fmt.Errorf("%s contains unsupported characters", field.name)
			}
		}
	}
	return nil
}

func (store *BrandingStore) load() error {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read branding settings: %w", err)
	}
	if err := json.Unmarshal(data, &store.settings); err != nil {
		return fmt.Errorf("decode branding settings: %w", err)
	}
	return nil
}

func (store *BrandingStore) saveLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	data, err := json.MarshalIndent(store.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode branding settings: %w", err)
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o640); err != nil {
		return fmt.Errorf("write branding settings: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace branding settings: %w", err)
	}
	return nil
}

func (app *App) handleBranding(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		writeJSON(writer, http.StatusOK, app.branding.Get())
	case http.MethodPut:
		if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
			return
		}
		var settings BrandingSettings
		if err := decodeJSON(request, &settings); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		saved, err := app.branding.Save(settings)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusOK, saved)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPut)
	}
}
