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
	Categories        []string          `json:"categories"`
	ServiceCategories map[string]string `json:"service_categories"`
}

type CatalogPreferenceStore struct {
	mu          sync.RWMutex
	path        string
	categories  map[string]struct{}
	assignments map[string]string
}

func NewCatalogPreferenceStore(path string) (*CatalogPreferenceStore, error) {
	store := &CatalogPreferenceStore{
		path:        path,
		categories:  make(map[string]struct{}),
		assignments: make(map[string]string),
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

func (store *CatalogPreferenceStore) Apply(services []Service) []Service {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]Service, len(services))
	copy(result, services)
	for index := range result {
		result[index].OriginalCategory = ""
		category, ok := store.assignments[result[index].ID]
		if !ok {
			continue
		}
		if category != result[index].Category {
			result[index].OriginalCategory = result[index].Category
			result[index].Category = category
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
	}
	for serviceID, category := range store.assignments {
		preferences.ServiceCategories[serviceID] = category
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
