package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type NodeLabel struct {
	Name  string `json:"name,omitempty"`
	Group string `json:"group,omitempty"`
}

type NodeLabelStore struct {
	mu     sync.RWMutex
	path   string
	labels map[string]NodeLabel
}

func NewNodeLabelStore(path string) (*NodeLabelStore, error) {
	store := &NodeLabelStore{path: path, labels: make(map[string]NodeLabel)}
	if path != "" {
		if err := store.load(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (store *NodeLabelStore) Overlay(node map[string]any) {
	if store == nil {
		return
	}
	store.mu.RLock()
	label, ok := store.labels[valueString(node["id"])]
	store.mu.RUnlock()
	if !ok {
		return
	}
	if label.Name != "" {
		node["name"] = label.Name
	}
	if label.Group != "" {
		node["group"] = label.Group
	}
}

func (store *NodeLabelStore) Set(id string, display, controller NodeLabel) error {
	if store == nil || id == "" {
		return nil
	}
	label := NodeLabel{}
	if display.Name != controller.Name {
		label.Name = display.Name
	}
	if display.Group != controller.Group {
		label.Group = display.Group
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if label.Name == "" && label.Group == "" {
		delete(store.labels, id)
	} else {
		store.labels[id] = label
	}
	return store.saveLocked()
}

func (store *NodeLabelStore) Delete(id string) error {
	if store == nil || id == "" {
		return nil
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.labels, id)
	return store.saveLocked()
}

func (store *NodeLabelStore) load() error {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read node labels: %w", err)
	}
	if err := json.Unmarshal(data, &store.labels); err != nil {
		return fmt.Errorf("decode node labels: %w", err)
	}
	return nil
}

func (store *NodeLabelStore) saveLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	data, err := json.MarshalIndent(store.labels, "", "  ")
	if err != nil {
		return fmt.Errorf("encode node labels: %w", err)
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o640); err != nil {
		return fmt.Errorf("write node labels: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace node labels: %w", err)
	}
	return nil
}

func validateNodeLabel(label NodeLabel) error {
	label.Name = strings.TrimSpace(label.Name)
	label.Group = strings.TrimSpace(label.Group)
	if label.Name == "" {
		return errors.New("node name is required")
	}
	if utf8.RuneCountInString(label.Name) > 64 {
		return errors.New("node name is too long")
	}
	if utf8.RuneCountInString(label.Group) > 64 {
		return errors.New("node group is too long")
	}
	for _, value := range []string{label.Name, label.Group} {
		for _, character := range value {
			if unicode.IsLetter(character) || unicode.IsNumber(character) || strings.ContainsRune(" .,_-", character) {
				continue
			}
			return errors.New("node name/group contains unsupported characters")
		}
	}
	return nil
}

func controllerNodeLabel(value string, keepComma bool) string {
	var normalized strings.Builder
	hasIdentifier := false
	for _, character := range strings.TrimSpace(value) {
		identifier := character <= unicode.MaxASCII && (unicode.IsLetter(character) || unicode.IsNumber(character))
		valid := identifier || character == ' ' || (keepComma && character == ',')
		if valid {
			normalized.WriteRune(character)
			hasIdentifier = hasIdentifier || identifier
		} else {
			normalized.WriteByte(' ')
		}
	}
	if !hasIdentifier {
		return ""
	}
	return strings.Join(strings.Fields(normalized.String()), " ")
}

func controllerNodeName(payload map[string]any, displayName string) string {
	if name := controllerNodeLabel(displayName, false); name != "" {
		return name
	}
	if address := strings.TrimSpace(valueString(payload["public_ip"])); address != "" {
		return safeNodeName(address)
	}
	return "Node"
}

func (app *App) handleEnhancedNodes(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
		return
	}
	switch request.Method {
	case http.MethodGet:
		var nodes []map[string]any
		if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodGet, "/api/nodes", nil, &nodes); err != nil {
			writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		for _, node := range nodes {
			app.nodeLabels.Overlay(node)
		}
		writeJSON(writer, http.StatusOK, nodes)
	case http.MethodPost:
		app.createEnhancedNode(writer, request)
	default:
		methodNotAllowed(writer, http.MethodGet, http.MethodPost)
	}
}

func (app *App) handleEnhancedNode(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "login expired"})
		return
	}
	id := strings.TrimPrefix(request.URL.Path, "/enhancer/api/nodes/")
	if id == "" || strings.Contains(id, "/") {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid node ID"})
		return
	}
	switch request.Method {
	case http.MethodPut:
		app.updateEnhancedNode(writer, request, id)
	case http.MethodDelete:
		if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+url.PathEscape(id), nil, nil); err != nil {
			writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := app.nodeLabels.Delete(id); err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(writer, http.MethodPut, http.MethodDelete)
	}
}

func (app *App) createEnhancedNode(writer http.ResponseWriter, request *http.Request) {
	var payload map[string]any
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	display := NodeLabel{Name: strings.TrimSpace(valueString(payload["name"])), Group: strings.TrimSpace(valueString(payload["group"]))}
	if err := validateNodeLabel(display); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	controller := NodeLabel{Name: controllerNodeName(payload, display.Name), Group: controllerNodeLabel(display.Group, true)}
	payload["name"], payload["group"] = controller.Name, controller.Group
	var created map[string]any
	if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodPost, "/api/nodes", payload, &created); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	id := valueString(created["id"])
	if err := app.nodeLabels.Set(id, display, controller); err != nil {
		_ = app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodDelete, "/api/nodes/"+url.PathEscape(id), nil, nil)
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	app.nodeLabels.Overlay(created)
	writeJSON(writer, http.StatusCreated, created)
}

func (app *App) updateEnhancedNode(writer http.ResponseWriter, request *http.Request, id string) {
	var payload map[string]any
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	display := NodeLabel{Name: strings.TrimSpace(valueString(payload["name"])), Group: strings.TrimSpace(valueString(payload["group"]))}
	if err := validateNodeLabel(display); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	controller := NodeLabel{Name: controllerNodeName(payload, display.Name), Group: controllerNodeLabel(display.Group, true)}
	payload["name"], payload["group"] = controller.Name, controller.Group
	if err := app.upstreamJSON(request.Context(), request.Header.Get("Authorization"), http.MethodPut, "/api/nodes/"+url.PathEscape(id), payload, nil); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	if err := app.nodeLabels.Set(id, display, controller); err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
}
