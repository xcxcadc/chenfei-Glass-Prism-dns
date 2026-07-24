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
	ID               string            `json:"id"`
	IP               string            `json:"ip"`
	Note             string            `json:"note,omitempty"`
	DNSNodeID        string            `json:"dns_node_id"`
	NodeName         string            `json:"node_name"`
	NodeSecret       string            `json:"-"`
	EnrollmentToken  string            `json:"enrollment_token"`
	Smart            bool              `json:"smart"`
	Routes           map[string]string `json:"routes"`
	TrafficPeers     []string          `json:"traffic_peers,omitempty"`
	TrafficRXBytes   uint64            `json:"traffic_rx_bytes"`
	TrafficTXBytes   uint64            `json:"traffic_tx_bytes"`
	TrafficUpdatedAt *time.Time        `json:"traffic_updated_at,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type ipConfigRecord struct {
	IPConfig
	NodeSecret  string `json:"node_secret"`
	LastRXBytes uint64 `json:"last_rx_bytes"`
	LastTXBytes uint64 `json:"last_tx_bytes"`
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
	return record, ok
}

func (store *IPConfigStore) GetByToken(token string) (ipConfigRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, record := range store.configs {
		if record.EnrollmentToken == token {
			return record, true
		}
	}
	return ipConfigRecord{}, false
}

func (store *IPConfigStore) Save(config IPConfig, secret string) (IPConfig, error) {
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
	} else {
		config.CreatedAt = now
	}
	config.UpdatedAt = now
	record.IPConfig = config
	store.configs[config.ID] = record
	if err := store.saveLocked(); err != nil {
		return IPConfig{}, err
	}
	return config, nil
}

func (store *IPConfigStore) UpdateTraffic(token string, rxBytes, txBytes uint64) (IPConfig, error) {
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
		store.configs[id] = record
		if err := store.saveLocked(); err != nil {
			return IPConfig{}, err
		}
		return record.public(), nil
	}
	return IPConfig{}, os.ErrNotExist
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
	for _, record := range records {
		store.configs[record.ID] = record
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
	return config
}

func randomToken() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate enrollment token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}
