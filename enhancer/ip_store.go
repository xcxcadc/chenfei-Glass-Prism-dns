package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type IPConfig struct {
	ID                      string            `json:"id"`
	IP                      string            `json:"ip"`
	Note                    string            `json:"note,omitempty"`
	DNSNodeID               string            `json:"dns_node_id"`
	NodeName                string            `json:"node_name"`
	ExternalDNSNode         bool              `json:"external_dns_node,omitempty"`
	NodeSecret              string            `json:"-"`
	EnrollmentToken         string            `json:"enrollment_token"`
	Smart                   bool              `json:"smart"`
	Routes                  map[string]string `json:"routes"`
	TrafficPeers            []string          `json:"traffic_peers,omitempty"`
	TrafficRXBytes          uint64            `json:"traffic_rx_bytes"`
	TrafficTXBytes          uint64            `json:"traffic_tx_bytes"`
	TrafficUpdatedAt        *time.Time        `json:"traffic_updated_at,omitempty"`
	DNSReady                bool              `json:"dns_ready"`
	SystemDNSReady          bool              `json:"system_dns_ready"`
	RoutesReady             bool              `json:"routes_ready"`
	HealthyRoutes           int               `json:"healthy_routes"`
	ExpectedRoutes          int               `json:"expected_routes"`
	HealthMessage           string            `json:"health_message,omitempty"`
	HealthUpdatedAt         *time.Time        `json:"health_updated_at,omitempty"`
	ServiceResults          map[string]string `json:"service_results,omitempty"`
	ServiceAuditedAt        *time.Time        `json:"service_audited_at,omitempty"`
	ServiceAuditRequestedAt *time.Time        `json:"service_audit_requested_at,omitempty"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

type ClientHealth struct {
	DNSReady       bool
	SystemDNSReady bool
	RoutesReady    bool
	HealthyRoutes  int
	ExpectedRoutes int
	Message        string
}

type ipConfigRecord struct {
	IPConfig
	NodeSecret  string              `json:"node_secret"`
	ProxyPeers  map[string][]string `json:"proxy_peers,omitempty"`
	LastRXBytes uint64              `json:"last_rx_bytes"`
	LastTXBytes uint64              `json:"last_tx_bytes"`
}

type IPConfigStore struct {
	mu      sync.RWMutex
	path    string
	configs map[string]ipConfigRecord
}

func NewIPConfigStore(path string) (*IPConfigStore, error) {
	store := &IPConfigStore{path: path, configs: make(map[string]ipConfigRecord)}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (store *IPConfigStore) List() []IPConfig {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]IPConfig, 0, len(store.configs))
	for _, record := range store.configs {
		result = append(result, record.public())
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func (store *IPConfigStore) Get(id string) (IPConfig, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, ok := store.configs[id]
	return record.public(), ok
}

func (store *IPConfigStore) Record(id string) (ipConfigRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, ok := store.configs[id]
	record = cloneIPConfigRecord(record)
	return record, ok
}

func (store *IPConfigStore) Records() []ipConfigRecord {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]ipConfigRecord, 0, len(store.configs))
	for _, record := range store.configs {
		result = append(result, cloneIPConfigRecord(record))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (store *IPConfigStore) GetByToken(token string) (ipConfigRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, record := range store.configs {
		if record.EnrollmentToken == token {
			return cloneIPConfigRecord(record), true
		}
	}
	return ipConfigRecord{}, false
}

func (store *IPConfigStore) GetByNodeSecret(secret string) (ipConfigRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, record := range store.configs {
		if secret != "" && record.NodeSecret == secret {
			return cloneIPConfigRecord(record), true
		}
	}
	return ipConfigRecord{}, false
}

func (store *IPConfigStore) Save(config IPConfig, secret string, proxyPeers ...map[string][]string) (IPConfig, error) {
	config.IP = strings.TrimSpace(config.IP)
	config.Note = strings.TrimSpace(config.Note)
	if net.ParseIP(config.IP) == nil {
		return IPConfig{}, errors.New("invalid IP address")
	}
	if len(config.Note) > 80 {
		return IPConfig{}, errors.New("note is too long")
	}
	if config.DNSNodeID == "" || secret == "" {
		return IPConfig{}, errors.New("DNS node credentials are required")
	}
	if config.Routes == nil {
		config.Routes = map[string]string{}
	}
	config.Routes = cloneStringMap(config.Routes)
	config.TrafficPeers = ipv4Peers(config.TrafficPeers)
	for serviceID, proxyID := range config.Routes {
		if strings.TrimSpace(serviceID) == "" || strings.TrimSpace(proxyID) == "" {
			return IPConfig{}, errors.New("service routes must include a proxy node")
		}
	}
	now := time.Now().UTC()
	if config.ID == "" {
		hash := sha256.Sum256([]byte(config.IP))
		config.ID = "ip-" + hex.EncodeToString(hash[:6])
	}
	if config.EnrollmentToken == "" {
		token, err := randomToken()
		if err != nil {
			return IPConfig{}, err
		}
		config.EnrollmentToken = token
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record := ipConfigRecord{IPConfig: config, NodeSecret: secret}
	if existing, ok := store.configs[config.ID]; ok {
		config.CreatedAt = existing.CreatedAt
		record.LastRXBytes = existing.LastRXBytes
		record.LastTXBytes = existing.LastTXBytes
		if stringMapEqual(existing.Routes, config.Routes) {
			config.ServiceResults = existing.ServiceResults
			config.ServiceAuditedAt = existing.ServiceAuditedAt
			config.ServiceAuditRequestedAt = existing.ServiceAuditRequestedAt
		} else {
			config.ServiceResults = nil
			config.ServiceAuditedAt = nil
			auditRequestedAt := monotonicTimestamp(now, existing.ServiceAuditRequestedAt)
			config.ServiceAuditRequestedAt = &auditRequestedAt
		}
		record.ProxyPeers = cloneProxyPeers(existing.ProxyPeers)
	} else {
		config.CreatedAt = now
		config.ServiceAuditRequestedAt = &now
	}
	if len(proxyPeers) > 0 {
		record.ProxyPeers = ipv4ProxyPeers(proxyPeers[0])
	}
	config.UpdatedAt = now
	record.IPConfig = config
	store.configs[config.ID] = record
	if err := store.saveLocked(); err != nil {
		return IPConfig{}, err
	}
	return config, nil
}

func cloneProxyPeers(source map[string][]string) map[string][]string {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string][]string, len(source))
	for proxyID, peers := range source {
		result[proxyID] = append([]string(nil), peers...)
	}
	return result
}

func ipv4ProxyPeers(source map[string][]string) map[string][]string {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string][]string, len(source))
	for proxyID, peers := range source {
		filtered := ipv4Peers(peers)
		if len(filtered) > 0 {
			result[proxyID] = filtered
		}
	}
	return result
}

func (store *IPConfigStore) UpdateClientReport(token string, rxBytes, txBytes uint64, health ClientHealth) (IPConfig, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	for id, record := range store.configs {
		if record.EnrollmentToken != token {
			continue
		}
		hasBaseline := record.TrafficUpdatedAt != nil
		if hasBaseline {
			if rxBytes >= record.LastRXBytes {
				record.TrafficRXBytes += rxBytes - record.LastRXBytes
			} else {
				record.TrafficRXBytes += rxBytes
			}
		}
		if hasBaseline {
			if txBytes >= record.LastTXBytes {
				record.TrafficTXBytes += txBytes - record.LastTXBytes
			} else {
				record.TrafficTXBytes += txBytes
			}
		}
		record.LastRXBytes = rxBytes
		record.LastTXBytes = txBytes
		now := time.Now().UTC()
		record.TrafficUpdatedAt = &now
		record.DNSReady = health.DNSReady
		record.SystemDNSReady = health.SystemDNSReady
		record.RoutesReady = health.RoutesReady
		record.HealthyRoutes = health.HealthyRoutes
		record.ExpectedRoutes = health.ExpectedRoutes
		record.HealthMessage = strings.TrimSpace(health.Message)
		record.HealthUpdatedAt = &now
		store.configs[id] = record
		if err := store.saveLocked(); err != nil {
			return IPConfig{}, err
		}
		return record.public(), nil
	}
	return IPConfig{}, os.ErrNotExist
}

func (store *IPConfigStore) UpdateTraffic(token string, rxBytes, txBytes uint64) (IPConfig, error) {
	return store.UpdateClientReport(token, rxBytes, txBytes, ClientHealth{})
}

func (store *IPConfigStore) UpdateServiceAudit(token string, results map[string]string) (IPConfig, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	for id, record := range store.configs {
		if record.EnrollmentToken != token {
			continue
		}
		filtered := make(map[string]string)
		for serviceID, result := range results {
			if _, selected := record.Routes[serviceID]; !selected {
				continue
			}
			result = strings.TrimSpace(result)
			resultRunes := []rune(result)
			if len(resultRunes) > 160 {
				result = string(resultRunes[:160])
			}
			if result != "" {
				filtered[serviceID] = result
			}
		}
		record.ServiceResults = filtered
		now := time.Now().UTC()
		record.ServiceAuditedAt = &now
		store.configs[id] = record
		if err := store.saveLocked(); err != nil {
			return IPConfig{}, err
		}
		return record.public(), nil
	}
	return IPConfig{}, os.ErrNotExist
}

func (store *IPConfigStore) RequestServiceAudit(id string) (IPConfig, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	record, ok := store.configs[id]
	if !ok {
		return IPConfig{}, os.ErrNotExist
	}
	now := monotonicTimestamp(time.Now().UTC(), record.ServiceAuditRequestedAt)
	record.ServiceAuditRequestedAt = &now
	store.configs[id] = record
	if err := store.saveLocked(); err != nil {
		return IPConfig{}, err
	}
	return record.public(), nil
}

func (store *IPConfigStore) ReplaceRoutes(id string, routes map[string]string) (IPConfig, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	record, ok := store.configs[id]
	if !ok {
		return IPConfig{}, os.ErrNotExist
	}
	if stringMapEqual(record.Routes, routes) {
		return record.public(), nil
	}
	store.applyNormalizedRoutes(&record, routes, time.Now().UTC())
	store.configs[id] = record
	if err := store.saveLocked(); err != nil {
		return IPConfig{}, err
	}
	return record.public(), nil
}

func (store *IPConfigStore) NormalizeRouteConflicts(services []Service) (int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	normalizedCount := 0
	now := time.Now().UTC()
	for id, record := range store.configs {
		routes, changed := normalizeConflictingRoutes(nil, record.Routes, services)
		if !changed {
			continue
		}
		store.applyNormalizedRoutes(&record, routes, now)
		store.configs[id] = record
		normalizedCount++
	}
	if normalizedCount == 0 {
		return 0, nil
	}
	if err := store.saveLocked(); err != nil {
		return 0, err
	}
	return normalizedCount, nil
}

func (store *IPConfigStore) applyNormalizedRoutes(record *ipConfigRecord, routes map[string]string, now time.Time) {
	updated := record.IPConfig
	updated.Routes = cloneStringMap(routes)
	updated.ServiceResults = nil
	updated.ServiceAuditedAt = nil
	auditRequestedAt := monotonicTimestamp(now, record.ServiceAuditRequestedAt)
	updated.ServiceAuditRequestedAt = &auditRequestedAt
	updated.UpdatedAt = now
	record.IPConfig = updated
	pruneUnusedProxyPeers(record)
	record.TrafficPeers = flattenProxyPeers(record.ProxyPeers)
}

func (store *IPConfigStore) ClearTraffic(id string) (IPConfig, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	record, ok := store.configs[id]
	if !ok {
		return IPConfig{}, os.ErrNotExist
	}
	record.TrafficRXBytes = 0
	record.TrafficTXBytes = 0
	record.LastRXBytes = 0
	record.LastTXBytes = 0
	record.TrafficUpdatedAt = nil
	store.configs[id] = record
	if err := store.saveLocked(); err != nil {
		return IPConfig{}, err
	}
	return record.public(), nil
}

func (store *IPConfigStore) Delete(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, ok := store.configs[id]; !ok {
		return os.ErrNotExist
	}
	delete(store.configs, id)
	return store.saveLocked()
}

func (store *IPConfigStore) load() error {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read IP configs: %w", err)
	}
	var records []ipConfigRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode IP configs: %w", err)
	}
	var rawRecords []map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawRecords); err != nil {
		return fmt.Errorf("inspect IP configs: %w", err)
	}
	normalized := false
	for _, rawRecord := range rawRecords {
		if _, exists := rawRecord["failovers"]; exists {
			normalized = true
			break
		}
	}
	for _, record := range records {
		originalTrafficPeers := append([]string(nil), record.TrafficPeers...)
		originalProxyPeers := cloneProxyPeers(record.ProxyPeers)
		record.ProxyPeers = ipv4ProxyPeers(record.ProxyPeers)
		pruneUnusedProxyPeers(&record)
		if len(record.ProxyPeers) > 0 {
			record.TrafficPeers = flattenProxyPeers(record.ProxyPeers)
		} else {
			record.TrafficPeers = ipv4Peers(record.TrafficPeers)
		}
		if !stringSlicesEqual(originalTrafficPeers, record.TrafficPeers) ||
			!proxyPeersEqual(originalProxyPeers, record.ProxyPeers) {
			normalized = true
		}
		store.configs[record.ID] = record
	}
	if normalized {
		return store.saveLocked()
	}
	return nil
}

func (store *IPConfigStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create IP config directory: %w", err)
	}
	records := make([]ipConfigRecord, 0, len(store.configs))
	for _, record := range store.configs {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("encode IP configs: %w", err)
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o600); err != nil {
		return fmt.Errorf("write IP configs: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace IP configs: %w", err)
	}
	return nil
}

func (record ipConfigRecord) public() IPConfig {
	config := record.IPConfig
	config.NodeSecret = ""
	config.Routes = cloneStringMap(config.Routes)
	config.TrafficPeers = append([]string(nil), config.TrafficPeers...)
	return config
}

func cloneIPConfigRecord(record ipConfigRecord) ipConfigRecord {
	record.Routes = cloneStringMap(record.Routes)
	record.ProxyPeers = cloneProxyPeers(record.ProxyPeers)
	record.TrafficPeers = append([]string(nil), record.TrafficPeers...)
	record.ServiceResults = cloneStringMap(record.ServiceResults)
	return record
}

func pruneUnusedProxyPeers(record *ipConfigRecord) {
	used := make(map[string]struct{})
	for _, proxyID := range record.Routes {
		used[proxyID] = struct{}{}
	}
	for proxyID := range record.ProxyPeers {
		if _, exists := used[proxyID]; !exists {
			delete(record.ProxyPeers, proxyID)
		}
	}
}

func flattenProxyPeers(proxyPeers map[string][]string) []string {
	seen := make(map[string]struct{})
	for _, peers := range proxyPeers {
		for _, peer := range ipv4Peers(peers) {
			seen[peer] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for peer := range seen {
		result = append(result, peer)
	}
	sort.Strings(result)
	return result
}

func monotonicTimestamp(current time.Time, previous *time.Time) time.Time {
	if previous != nil && !current.After(*previous) {
		return previous.Add(time.Nanosecond)
	}
	return current
}

func randomToken() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate enrollment token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func stringMapEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func proxyPeersEqual(left, right map[string][]string) bool {
	if len(left) != len(right) {
		return false
	}
	for proxyID, peers := range left {
		if !stringSlicesEqual(peers, right[proxyID]) {
			return false
		}
	}
	return true
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
