package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type CustomServiceStore struct {
	mu       sync.RWMutex
	path     string
	services map[string]Service
}

func NewCustomServiceStore(path string) (*CustomServiceStore, error) {
	store := &CustomServiceStore{path: path, services: make(map[string]Service)}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (store *CustomServiceStore) List() []Service {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]Service, 0, len(store.services))
	for _, service := range store.services {
		result = append(result, service)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Category == result[j].Category {
			return result[i].Name < result[j].Name
		}
		return result[i].Category < result[j].Category
	})
	return result
}

func (store *CustomServiceStore) Get(id string) (Service, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	service, ok := store.services[id]
	return service, ok
}

func (store *CustomServiceStore) Upsert(service Service) (Service, error) {
	service.Name = strings.TrimSpace(service.Name)
	service.Category = strings.TrimSpace(service.Category)
	service.Domains = normalizeDomains(service.Domains)
	service.Custom = true
	if service.Name == "" {
		return Service{}, errors.New("service name is required")
	}
	if service.Category == "" {
		service.Category = "自定义服务"
	}
	if len(service.Domains) == 0 {
		return Service{}, errors.New("at least one valid domain is required")
	}
	if service.ID == "" {
		service.ID = customServiceID(service.Name)
	}
	if !strings.HasPrefix(service.ID, "custom-") {
		return Service{}, errors.New("invalid custom service id")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.services[service.ID] = service
	if err := store.saveLocked(); err != nil {
		return Service{}, err
	}
	return service, nil
}

func (store *CustomServiceStore) Delete(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, ok := store.services[id]; !ok {
		return os.ErrNotExist
	}
	delete(store.services, id)
	return store.saveLocked()
}

func (store *CustomServiceStore) load() error {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read custom services: %w", err)
	}
	var services []Service
	if err := json.Unmarshal(data, &services); err != nil {
		return fmt.Errorf("decode custom services: %w", err)
	}
	for _, service := range services {
		service.Custom = true
		store.services[service.ID] = service
	}
	return nil
}

func (store *CustomServiceStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	services := make([]Service, 0, len(store.services))
	for _, service := range store.services {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].ID < services[j].ID })
	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return fmt.Errorf("encode custom services: %w", err)
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o640); err != nil {
		return fmt.Errorf("write custom services: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace custom services: %w", err)
	}
	return nil
}
