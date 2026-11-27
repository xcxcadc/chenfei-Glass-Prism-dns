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
)

type WarpPolicyStore struct {
	mu       sync.RWMutex
	path     string
	policies map[string]map[string]string // proxyNodeID -> serviceID -> mode ("warp"|"warp6"|"warp4"|"auto")
}

func NewWarpPolicyStore(path string) (*WarpPolicyStore, error) {
	store := &WarpPolicyStore{
		path:     path,
		policies: make(map[string]map[string]string),
	}
	if path != "" {
		if err := store.load(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func normalizeWarpMode(mode string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "direct", "off", "none":
		return "", true
	case "warp", "warp6", "warp4", "auto":
		return strings.ToLower(strings.TrimSpace(mode)), true
	default:
		return "", false
	}
}

func (store *WarpPolicyStore) load() error {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read warp policies: %w", err)
	}
	var raw map[string]map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode warp policies: %w", err)
	}
	for proxyID, services := range raw {
		proxyID = strings.TrimSpace(proxyID)
		if proxyID == "" {
			continue
		}
		clean := make(map[string]string)
		for serviceID, mode := range services {
			serviceID = strings.TrimSpace(serviceID)
			norm, ok := normalizeWarpMode(mode)
			if serviceID != "" && ok && norm != "" {
				clean[serviceID] = norm
			}
		}
		if len(clean) > 0 {
			store.policies[proxyID] = clean
		}
	}
	return nil
}

func (store *WarpPolicyStore) saveLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create warp policy directory: %w", err)
	}
	data, err := json.MarshalIndent(store.policies, "", "  ")
	if err != nil {
		return fmt.Errorf("encode warp policies: %w", err)
	}
	tmp := store.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write warp policies: %w", err)
	}
	if err := os.Rename(tmp, store.path); err != nil {
		return fmt.Errorf("replace warp policies: %w", err)
	}
	return nil
}

func (store *WarpPolicyStore) Mode(proxyID, serviceID string) string {
	if store == nil {
		return ""
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	if services, ok := store.policies[proxyID]; ok {
		return services[serviceID]
	}
	return ""
}

func (store *WarpPolicyStore) ForProxy(proxyID string) map[string]string {
	result := make(map[string]string)
	if store == nil {
		return result
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	for k, v := range store.policies[proxyID] {
		result[k] = v
	}
	return result
}

func (store *WarpPolicyStore) Snapshot() map[string]map[string]string {
	out := make(map[string]map[string]string)
	if store == nil {
		return out
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	for proxyID, services := range store.policies {
		copied := make(map[string]string, len(services))
		for k, v := range services {
			copied[k] = v
		}
		out[proxyID] = copied
	}
	return out
}

func (store *WarpPolicyStore) Set(proxyID, serviceID, mode string) (map[string]string, error) {
	proxyID = strings.TrimSpace(proxyID)
	serviceID = strings.TrimSpace(serviceID)
	if proxyID == "" || serviceID == "" {
		return nil, errors.New("proxy_id and service_id are required")
	}
	norm, ok := normalizeWarpMode(mode)
	if !ok {
		return nil, fmt.Errorf("invalid warp mode %q", mode)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if norm == "" {
		if services, exists := store.policies[proxyID]; exists {
			delete(services, serviceID)
			if len(services) == 0 {
				delete(store.policies, proxyID)
			}
		}
	} else {
		if store.policies[proxyID] == nil {
			store.policies[proxyID] = make(map[string]string)
		}
		store.policies[proxyID][serviceID] = norm
	}
	if err := store.saveLocked(); err != nil {
		return nil, err
	}
	result := make(map[string]string, len(store.policies[proxyID]))
	for k, v := range store.policies[proxyID] {
		result[k] = v
	}
	return result, nil
}

func (store *WarpPolicyStore) SetBatch(proxyID string, updates map[string]string, replace bool) (map[string]string, error) {
	proxyID = strings.TrimSpace(proxyID)
	if proxyID == "" {
		return nil, errors.New("proxy_id is required")
	}
	clean := make(map[string]string)
	for serviceID, mode := range updates {
		serviceID = strings.TrimSpace(serviceID)
		if serviceID == "" {
			continue
		}
		norm, ok := normalizeWarpMode(mode)
		if !ok {
			return nil, fmt.Errorf("invalid warp mode %q for %s", mode, serviceID)
		}
		clean[serviceID] = norm
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if replace {
		next := make(map[string]string)
		for serviceID, mode := range clean {
			if mode != "" {
				next[serviceID] = mode
			}
		}
		if len(next) == 0 {
			delete(store.policies, proxyID)
		} else {
			store.policies[proxyID] = next
		}
	} else {
		for serviceID, mode := range clean {
			if mode == "" {
				if store.policies[proxyID] != nil {
					delete(store.policies[proxyID], serviceID)
				}
			} else {
				if store.policies[proxyID] == nil {
					store.policies[proxyID] = make(map[string]string)
				}
				store.policies[proxyID][serviceID] = mode
			}
		}
		if len(store.policies[proxyID]) == 0 {
			delete(store.policies, proxyID)
		}
	}
	if err := store.saveLocked(); err != nil {
		return nil, err
	}
	result := make(map[string]string, len(store.policies[proxyID]))
	for k, v := range store.policies[proxyID] {
		result[k] = v
	}
	return result, nil
}

func (store *WarpPolicyStore) DeleteProxy(proxyID string) error {
	if store == nil || strings.TrimSpace(proxyID) == "" {
		return nil
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.policies[proxyID]; !exists {
		return nil
	}
	delete(store.policies, proxyID)
	return store.saveLocked()
}

type warpPolicyUpdateRequest struct {
	ProxyID   string            `json:"proxy_id"`
	ServiceID string            `json:"service_id,omitempty"`
	Mode      string            `json:"mode,omitempty"`
	Services  map[string]string `json:"services,omitempty"`
	Replace   bool              `json:"replace,omitempty"`
}

func (app *App) handleWarpPolicies(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	switch request.Method {
	case http.MethodGet:
		writeJSON(writer, http.StatusOK, map[string]any{
			"policies":   app.warpPolicies.Snapshot(),
			"warp_ready": app.transport.WarpReadyMap(),
		})
	case http.MethodPut, http.MethodPost:
		var payload warpPolicyUpdateRequest
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		var (
			updated map[string]string
			err     error
		)
		if payload.Services != nil || payload.Replace {
			updated, err = app.warpPolicies.SetBatch(payload.ProxyID, payload.Services, payload.Replace)
		} else {
			updated, err = app.warpPolicies.Set(payload.ProxyID, payload.ServiceID, payload.Mode)
		}
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{
			"proxy_id":   payload.ProxyID,
			"services":   updated,
			"warp_ready": app.transport.ProxyWarpReady(payload.ProxyID),
		})
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPut, http.MethodPost)
	}
}
