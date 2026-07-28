package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type CatalogSnapshot struct {
	Source     string    `json:"source"`
	UpdatedAt  time.Time `json:"updated_at"`
	Services   []Service `json:"services"`
	Categories []string  `json:"categories"`
	Error      string    `json:"error,omitempty"`
}

type CatalogManager struct {
	mu          sync.RWMutex
	sourceURL   string
	client      *http.Client
	store       *CustomServiceStore
	preferences *CatalogPreferenceStore
	snapshot    CatalogSnapshot
	cacheTTL    time.Duration
}

func NewCatalogManager(sourceURL string, client *http.Client, store *CustomServiceStore, preferences ...*CatalogPreferenceStore) *CatalogManager {
	preferenceStore, _ := NewCatalogPreferenceStore("")
	if len(preferences) > 0 && preferences[0] != nil {
		preferenceStore = preferences[0]
	}
	return &CatalogManager{
		sourceURL:   sourceURL,
		client:      client,
		store:       store,
		preferences: preferenceStore,
		cacheTTL:    6 * time.Hour,
		snapshot:    CatalogSnapshot{Source: sourceURL},
	}
}

func (manager *CatalogManager) Snapshot(ctx context.Context, force bool) CatalogSnapshot {
	manager.mu.RLock()
	snapshot := manager.snapshot
	manager.mu.RUnlock()
	if force || snapshot.UpdatedAt.IsZero() || time.Since(snapshot.UpdatedAt) > manager.cacheTTL {
		if err := manager.refresh(ctx); err != nil {
			manager.mu.Lock()
			manager.snapshot.Error = err.Error()
			snapshot = manager.snapshot
			manager.mu.Unlock()
		} else {
			manager.mu.RLock()
			snapshot = manager.snapshot
			manager.mu.RUnlock()
		}
	}
	snapshot.Services = manager.preferences.Apply(mergeServices(snapshot.Services, manager.store.List()))
	sortServices(snapshot.Services)
	snapshot.Categories = manager.categories(snapshot.Services)
	return snapshot
}

func (manager *CatalogManager) Service(ctx context.Context, id string) (Service, bool) {
	if service, ok := manager.store.Get(id); ok {
		return manager.preferences.Apply([]Service{service})[0], true
	}
	snapshot := manager.Snapshot(ctx, false)
	for _, service := range snapshot.Services {
		if service.ID == id {
			return service, true
		}
	}
	return Service{}, false
}

func (manager *CatalogManager) refresh(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, manager.sourceURL, nil)
	if err != nil {
		return fmt.Errorf("create catalog request: %w", err)
	}
	response, err := manager.client.Do(request)
	if err != nil {
		return fmt.Errorf("download catalog: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("download catalog: unexpected status %s", response.Status)
	}
	services, err := ParseSmartDNS(response.Body)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return fmt.Errorf("download catalog: no services found")
	}
	manager.mu.Lock()
	manager.snapshot = CatalogSnapshot{
		Source:    manager.sourceURL,
		UpdatedAt: time.Now().UTC(),
		Services:  services,
	}
	manager.mu.Unlock()
	return nil
}

func mergeServices(base, custom []Service) []Service {
	result := make([]Service, 0, len(base)+len(custom))
	result = append(result, custom...)
	result = append(result, base...)
	sortServices(result)
	return result
}

func sortServices(services []Service) {
	sort.SliceStable(services, func(i, j int) bool {
		if services[i].Custom != services[j].Custom {
			return services[i].Custom
		}
		if services[i].Category == services[j].Category {
			return services[i].Name < services[j].Name
		}
		return services[i].Category < services[j].Category
	})
}

func (manager *CatalogManager) categories(services []Service) []string {
	categories := make(map[string]struct{})
	for _, service := range services {
		if service.Category != "" {
			categories[service.Category] = struct{}{}
		}
	}
	for _, category := range manager.preferences.Categories() {
		categories[category] = struct{}{}
	}
	return sortedCategoryKeys(categories)
}
