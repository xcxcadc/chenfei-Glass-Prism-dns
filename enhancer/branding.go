package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxBrandingIconBytes     = 2 << 20
	maxBrandingIconDimension = 2048
)

type BrandingSettings struct {
	SiteName     string `json:"site_name"`
	BrowserTitle string `json:"browser_title"`
	SiteTagline  string `json:"site_tagline"`
	IconVersion  int64  `json:"icon_version,omitempty"`
}

type BrandingStore struct {
	mu       sync.RWMutex
	path     string
	iconPath string
	settings BrandingSettings
}

func NewBrandingStore(path string) (*BrandingStore, error) {
	store := &BrandingStore{path: path}
	if path != "" {
		store.iconPath = filepath.Join(filepath.Dir(path), "branding-icon.png")
	}
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
	settings.IconVersion = store.settings.IconVersion
	store.settings = settings
	if err := store.saveLocked(); err != nil {
		return BrandingSettings{}, err
	}
	return store.settings, nil
}

func (store *BrandingStore) Icon() ([]byte, error) {
	if store == nil {
		return nil, os.ErrNotExist
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.iconPath == "" {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(store.iconPath)
}

func (store *BrandingStore) SaveIcon(data []byte) (BrandingSettings, error) {
	if store == nil || store.iconPath == "" {
		return BrandingSettings{}, errors.New("branding icon storage is unavailable")
	}
	if err := validateBrandingIcon(data); err != nil {
		return BrandingSettings{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(store.iconPath), 0o750); err != nil {
		return BrandingSettings{}, fmt.Errorf("create icon directory: %w", err)
	}
	temporaryPath := store.iconPath + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o640); err != nil {
		return BrandingSettings{}, fmt.Errorf("write branding icon: %w", err)
	}
	if err := os.Rename(temporaryPath, store.iconPath); err != nil {
		_ = os.Remove(temporaryPath)
		return BrandingSettings{}, fmt.Errorf("replace branding icon: %w", err)
	}
	store.settings.IconVersion = time.Now().UnixNano()
	if err := store.saveLocked(); err != nil {
		return BrandingSettings{}, err
	}
	return store.settings, nil
}

func (store *BrandingStore) ResetIcon() (BrandingSettings, error) {
	if store == nil || store.iconPath == "" {
		return BrandingSettings{}, errors.New("branding icon storage is unavailable")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := os.Remove(store.iconPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return BrandingSettings{}, fmt.Errorf("remove branding icon: %w", err)
	}
	store.settings.IconVersion = time.Now().UnixNano()
	if err := store.saveLocked(); err != nil {
		return BrandingSettings{}, err
	}
	return store.settings, nil
}

func validateBrandingIcon(data []byte) error {
	if len(data) == 0 || len(data) > maxBrandingIconBytes {
		return fmt.Errorf("branding icon must be between 1 byte and %d MB", maxBrandingIconBytes/(1<<20))
	}
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return errors.New("branding icon must be a valid PNG")
	}
	if config.Width < 16 || config.Height < 16 || config.Width > maxBrandingIconDimension || config.Height > maxBrandingIconDimension {
		return fmt.Errorf("branding icon must be between 16 and %d pixels", maxBrandingIconDimension)
	}
	image, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return errors.New("branding icon PNG could not be decoded")
	}
	transparent := false
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y && !transparent; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			if color.RGBAModel.Convert(image.At(x, y)).(color.RGBA).A < 255 {
				transparent = true
				break
			}
		}
	}
	if !transparent {
		return errors.New("branding icon must have a transparent background")
	}
	return nil
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

func (app *App) handleBrandingIcon(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		icon, err := app.brandingIcon()
		if err != nil {
			http.Error(writer, "branding icon unavailable", http.StatusNotFound)
			return
		}
		writer.Header().Set("Content-Type", "image/png")
		writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		_, _ = writer.Write(icon)
	case http.MethodPut:
		if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
			return
		}
		request.Body = http.MaxBytesReader(writer, request.Body, maxBrandingIconBytes+64<<10)
		file, _, err := request.FormFile("icon")
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "请选择 PNG 图标文件"})
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, maxBrandingIconBytes+1))
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "读取图标文件失败"})
			return
		}
		saved, err := app.branding.SaveIcon(data)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusOK, saved)
	case http.MethodDelete:
		if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
			writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
			return
		}
		settings, err := app.branding.ResetIcon()
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusOK, settings)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (app *App) brandingIcon() ([]byte, error) {
	if app.branding != nil {
		icon, err := app.branding.Icon()
		if err == nil {
			return icon, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return fs.ReadFile(app.web, "favicon.png")
}
