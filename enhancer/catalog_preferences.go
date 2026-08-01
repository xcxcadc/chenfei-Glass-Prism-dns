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
	"unicode"
	"unicode/utf8"
)

type CatalogPreferences struct {
	Categories        []string            `json:"categories"`
	ServiceCategories map[string]string   `json:"service_categories"`
	ServiceDomains    map[string][]string `json:"service_domains,omitempty"`
	DeletedServices   []string            `json:"deleted_services,omitempty"`
}

type CatalogPreferenceStore struct {
	mu          sync.RWMutex
	path        string
	categories  map[string]struct{}
	assignments map[string]string
	domains     map[string][]string
	deleted     map[string]struct{}
}

func NewCatalogPreferenceStore(path string) (*CatalogPreferenceStore, error) {
	store := &CatalogPreferenceStore{
		path:        path,
		categories:  make(map[string]struct{}),
		assignments: make(map[string]string),
		domains:     make(map[string][]string),
		deleted:     make(map[string]struct{}),
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (store *CatalogPreferenceStore) Categories() []string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return sortedCategoryKeys(store.categories)
}

func (store *CatalogPreferenceStore) IsCustomCategory(category string) bool {
	store.mu.RLock()
	defer store.mu.RUnlock()
	_, ok := store.categories[category]
	return ok
}

func (store *CatalogPreferenceStore) AddCategory(category string) (string, error) {
	normalized, err := normalizeCategory(category)
	if err != nil {
		return "", err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	_, existed := store.categories[normalized]
	store.categories[normalized] = struct{}{}
	if err := store.saveLocked(); err != nil {
		if !existed {
			delete(store.categories, normalized)
		}
		return "", err
	}
	return normalized, nil
}

func (store *CatalogPreferenceStore) DeleteCategory(category string) error {
	normalized, err := normalizeCategory(category)
	if err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, ok := store.categories[normalized]; !ok {
		return os.ErrNotExist
	}
	delete(store.categories, normalized)
	if err := store.saveLocked(); err != nil {
		store.categories[normalized] = struct{}{}
		return err
	}
	return nil
}

func (store *CatalogPreferenceStore) SetServiceCategory(serviceID, category string) error {
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" {
		return errors.New("service id is required")
	}
	normalized, err := normalizeCategory(category)
	if err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	previous, existed := store.assignments[serviceID]
	store.assignments[serviceID] = normalized
	if err := store.saveLocked(); err != nil {
		if existed {
			store.assignments[serviceID] = previous
		} else {
			delete(store.assignments, serviceID)
		}
		return err
	}
	return nil
}

func (store *CatalogPreferenceStore) ClearServiceCategory(serviceID string) error {
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" {
		return errors.New("service id is required")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	previous, ok := store.assignments[serviceID]
	if !ok {
		return nil
	}
	delete(store.assignments, serviceID)
	if err := store.saveLocked(); err != nil {
		store.assignments[serviceID] = previous
		return err
	}
	return nil
}

func (store *CatalogPreferenceStore) SetServiceDomains(serviceID string, domains []string) error {
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" {
		return errors.New("service id is required")
	}
	normalized, err := normalizeCustomDomains(domains)
	if err != nil {
		return err
	}
	if len(normalized) == 0 {
		return errors.New("at least one valid domain is required")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	previous, existed := store.domains[serviceID]
	store.domains[serviceID] = append([]string(nil), normalized...)
	if err := store.saveLocked(); err != nil {
		if existed {
			store.domains[serviceID] = previous
		} else {
			delete(store.domains, serviceID)
		}
		return err
	}
	return nil
}

func (store *CatalogPreferenceStore) ClearServiceDomains(serviceID string) error {
	serviceID = strings.TrimSpace(serviceID)
	if serviceID == "" {
		return errors.New("service id is required")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	previous, existed := store.domains[serviceID]
	if !existed {
		return nil
	}
	delete(store.domains, serviceID)
	if err := store.saveLocked(); err != nil {
		store.domains[serviceID] = previous
		return err
	}
	return nil
}

func (store *CatalogPreferenceStore) DeleteService(serviceID string, aliases []string) error {
	ids := uniqueServiceIDs(serviceID, aliases)
	if len(ids) == 0 {
		return errors.New("service id is required")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	previousDeleted := make(map[string]bool, len(ids))
	previousAssignments := make(map[string]string, len(ids))
	previousDomains := make(map[string][]string, len(ids))
	for _, id := range ids {
		_, previousDeleted[id] = store.deleted[id]
		if category, ok := store.assignments[id]; ok {
			previousAssignments[id] = category
		}
		if domains, ok := store.domains[id]; ok {
			previousDomains[id] = append([]string(nil), domains...)
		}
		store.deleted[id] = struct{}{}
		delete(store.assignments, id)
		delete(store.domains, id)
	}
	if err := store.saveLocked(); err != nil {
		for _, id := range ids {
			if previousDeleted[id] {
				store.deleted[id] = struct{}{}
			} else {
				delete(store.deleted, id)
			}
			if category, ok := previousAssignments[id]; ok {
				store.assignments[id] = category
			} else {
				delete(store.assignments, id)
			}
			if domains, ok := previousDomains[id]; ok {
				store.domains[id] = append([]string(nil), domains...)
			} else {
				delete(store.domains, id)
			}
		}
		return err
	}
	return nil
}

func (store *CatalogPreferenceStore) Apply(services []Service) []Service {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]Service, 0, len(services))
	for _, service := range services {
		if serviceIDsDeleted(service, store.deleted) {
			continue
		}
		result = append(result, service)
		index := len(result) - 1
		result[index].OriginalCategory = ""
		result[index].DomainOverride = false
		preferenceIDs := append([]string{result[index].ID}, result[index].Aliases...)
		category := ""
		categoryFound := false
		for _, preferenceID := range preferenceIDs {
			if value, ok := store.assignments[preferenceID]; ok {
				category = value
				categoryFound = true
				break
			}
		}
		if categoryFound && category != result[index].Category {
			result[index].OriginalCategory = result[index].Category
			result[index].Category = category
		}
		for _, preferenceID := range preferenceIDs {
			domains, exists := store.domains[preferenceID]
			if !exists {
				continue
			}
			result[index].Domains = append([]string(nil), domains...)
			result[index].DomainOverride = true
			break
		}
	}
	return result
}

func (store *CatalogPreferenceStore) load() error {
	if store.path == "" {
		return nil
	}
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read catalog preferences: %w", err)
	}
	var preferences CatalogPreferences
	if err := json.Unmarshal(data, &preferences); err != nil {
		return fmt.Errorf("decode catalog preferences: %w", err)
	}
	for _, category := range preferences.Categories {
		normalized, normalizeErr := normalizeCategory(category)
		if normalizeErr != nil {
			return fmt.Errorf("decode catalog preferences: %w", normalizeErr)
		}
		store.categories[normalized] = struct{}{}
	}
	for serviceID, category := range preferences.ServiceCategories {
		normalized, normalizeErr := normalizeCategory(category)
		if strings.TrimSpace(serviceID) == "" || normalizeErr != nil {
			return fmt.Errorf("decode catalog preferences: invalid service category")
		}
		store.assignments[serviceID] = normalized
	}
	for serviceID, domains := range preferences.ServiceDomains {
		if strings.TrimSpace(serviceID) == "" {
			return fmt.Errorf("decode catalog preferences: invalid service domains")
		}
		normalized, normalizeErr := normalizeCustomDomains(domains)
		if normalizeErr != nil {
			return fmt.Errorf("decode catalog preferences: invalid service domains")
		}
		store.domains[serviceID] = normalized
	}
	for _, serviceID := range preferences.DeletedServices {
		serviceID = strings.TrimSpace(serviceID)
		if serviceID == "" {
			return fmt.Errorf("decode catalog preferences: invalid deleted service")
		}
		store.deleted[serviceID] = struct{}{}
	}
	return nil
}

func (store *CatalogPreferenceStore) saveLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	preferences := CatalogPreferences{
		Categories:        sortedCategoryKeys(store.categories),
		ServiceCategories: make(map[string]string, len(store.assignments)),
		ServiceDomains:    make(map[string][]string, len(store.domains)),
		DeletedServices:   sortedServiceIDs(store.deleted),
	}
	for serviceID, category := range store.assignments {
		preferences.ServiceCategories[serviceID] = category
	}
	for serviceID, domains := range store.domains {
		preferences.ServiceDomains[serviceID] = append([]string(nil), domains...)
	}
	data, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return fmt.Errorf("encode catalog preferences: %w", err)
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o640); err != nil {
		return fmt.Errorf("write catalog preferences: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace catalog preferences: %w", err)
	}
	return nil
}

func uniqueServiceIDs(serviceID string, aliases []string) []string {
	seen := make(map[string]struct{}, len(aliases)+1)
	result := make([]string, 0, len(aliases)+1)
	for _, value := range append([]string{serviceID}, aliases...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func serviceIDsDeleted(service Service, deleted map[string]struct{}) bool {
	for _, serviceID := range uniqueServiceIDs(service.ID, service.Aliases) {
		if _, ok := deleted[serviceID]; ok {
			return true
		}
	}
	return false
}

func sortedServiceIDs(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for serviceID := range values {
		result = append(result, serviceID)
	}
	sort.Strings(result)
	return result
}

func normalizeCategory(category string) (string, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return "", errors.New("category is required")
	}
	if utf8.RuneCountInString(category) > 64 {
		return "", errors.New("category must be 64 characters or fewer")
	}
	for _, character := range category {
		if unicode.IsControl(character) {
			return "", errors.New("category contains unsupported characters")
		}
	}
	return category, nil
}

func sortedCategoryKeys(categories map[string]struct{}) []string {
	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	sort.Strings(result)
	return result
}
