package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const transportFreshness = 90 * time.Second

type proxyTransportRecord struct {
	NodeID     string    `json:"node_id"`
	Endpoint   string    `json:"endpoint"`
	SSHHostKey string    `json:"ssh_host_key"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type clientTransportRecord struct {
	ConfigID     string               `json:"config_id"`
	SSHPublicKey string               `json:"ssh_public_key"`
	ReadyProxies map[string]time.Time `json:"ready_proxies,omitempty"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type transportState struct {
	Proxies map[string]proxyTransportRecord  `json:"proxies"`
	Clients map[string]clientTransportRecord `json:"clients"`
}

type TransportStore struct {
	mu      sync.RWMutex
	path    string
	proxies map[string]proxyTransportRecord
	clients map[string]clientTransportRecord
}

type proxyTransportRequest struct {
	Secret     string `json:"secret"`
	SSHHostKey string `json:"ssh_host_key"`
	PublicKey  string `json:"public_key,omitempty"`
	ListenPort int    `json:"listen_port,omitempty"`
}

type clientTransportRequest struct {
	Token        string    `json:"token"`
	SSHPublicKey string    `json:"ssh_public_key"`
	ReadyProxies *[]string `json:"ready_proxies,omitempty"`
	PublicKey    string    `json:"public_key,omitempty"`
}

type transportPeer struct {
	ProxyID       string `json:"proxy_id,omitempty"`
	ConfigID      string `json:"config_id,omitempty"`
	SSHPublicKey  string `json:"ssh_public_key,omitempty"`
	SSHHostKey    string `json:"ssh_host_key,omitempty"`
	SSHHost       string `json:"ssh_host,omitempty"`
	SSHPort       int    `json:"ssh_port,omitempty"`
	ProxyIP       string `json:"proxy_ip,omitempty"`
	ClientIP      string `json:"client_ip,omitempty"`
	PublicProxyIP string `json:"public_proxy_ip,omitempty"`
}

type transportConfig struct {
	Role      string          `json:"role"`
	Interface string          `json:"interface"`
	Peers     []transportPeer `json:"peers"`
}

type controllerNode struct {
	ID       string
	Role     string
	PublicIP string
	Address  string
}

func NewTransportStore(path string) (*TransportStore, error) {
	store := &TransportStore{
		path:    path,
		proxies: make(map[string]proxyTransportRecord),
		clients: make(map[string]clientTransportRecord),
	}
	if path != "" {
		if err := store.load(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (store *TransportStore) RegisterProxy(nodeID, endpoint, sshHostKey string) error {
	if net.ParseIP(endpoint) == nil {
		return errors.New("invalid proxy endpoint")
	}
	normalizedHostKey, err := normalizeSSHKey(sshHostKey)
	if err != nil {
		return fmt.Errorf("invalid SSH host key: %w", err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.proxies[nodeID] = proxyTransportRecord{
		NodeID:     nodeID,
		Endpoint:   endpoint,
		SSHHostKey: normalizedHostKey,
		UpdatedAt:  time.Now().UTC(),
	}
	return store.saveLocked()
}

func (store *TransportStore) RegisterClient(configID, sshPublicKey string, readyProxies *[]string) error {
	normalizedPublicKey, err := normalizeSSHKey(sshPublicKey)
	if err != nil {
		return fmt.Errorf("invalid SSH client key: %w", err)
	}
	now := time.Now().UTC()
	store.mu.Lock()
	defer store.mu.Unlock()
	record := store.clients[configID]
	if record.SSHPublicKey != "" && record.SSHPublicKey != normalizedPublicKey {
		record.ReadyProxies = nil
	}
	record.ConfigID = configID
	record.SSHPublicKey = normalizedPublicKey
	record.UpdatedAt = now
	if readyProxies != nil {
		record.ReadyProxies = make(map[string]time.Time, len(*readyProxies))
		for _, proxyID := range *readyProxies {
			proxyID = strings.TrimSpace(proxyID)
			if proxyID != "" {
				record.ReadyProxies[proxyID] = now
			}
		}
	}
	store.clients[configID] = record
	return store.saveLocked()
}

func (store *TransportStore) ClientConfig(record ipConfigRecord) transportConfig {
	store.mu.RLock()
	defer store.mu.RUnlock()
	_, exists := store.clients[record.ID]
	config := transportConfig{Role: "client", Interface: "loopback", Peers: []transportPeer{}}
	if !exists {
		return config
	}
	for _, proxyID := range uniqueProxyIDs(record.Routes) {
		proxy, ok := store.proxies[proxyID]
		if !ok || !transportRecordFresh(proxy.UpdatedAt) {
			continue
		}
		proxyIP, clientIP := transportPairIPs(proxyID, record.ID)
		config.Peers = append(config.Peers, transportPeer{
			ProxyID:       proxyID,
			SSHHostKey:    proxy.SSHHostKey,
			SSHHost:       proxy.Endpoint,
			SSHPort:       22,
			ProxyIP:       proxyIP,
			ClientIP:      clientIP,
			PublicProxyIP: proxy.Endpoint,
		})
	}
	sort.Slice(config.Peers, func(i, j int) bool { return config.Peers[i].ProxyID < config.Peers[j].ProxyID })
	return config
}

func (store *TransportStore) ProxyConfig(proxyID string, records []ipConfigRecord) transportConfig {
	store.mu.RLock()
	defer store.mu.RUnlock()
	config := transportConfig{Role: "proxy", Interface: "openssh", Peers: []transportPeer{}}
	if _, exists := store.proxies[proxyID]; !exists {
		return config
	}
	for _, record := range records {
		if !routesUseProxy(record.Routes, proxyID) {
			continue
		}
		client, ok := store.clients[record.ID]
		if !ok || !transportRecordFresh(client.UpdatedAt) {
			continue
		}
		proxyIP, clientIP := transportPairIPs(proxyID, record.ID)
		config.Peers = append(config.Peers, transportPeer{
			ConfigID:     record.ID,
			SSHPublicKey: client.SSHPublicKey,
			ProxyIP:      proxyIP,
			ClientIP:     clientIP,
		})
	}
	sort.Slice(config.Peers, func(i, j int) bool { return config.Peers[i].ConfigID < config.Peers[j].ConfigID })
	return config
}

func (store *TransportStore) EffectiveProxyIP(configID, proxyID string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	client, clientExists := store.clients[configID]
	proxy, proxyExists := store.proxies[proxyID]
	if !clientExists || !proxyExists || !transportRecordFresh(client.UpdatedAt) || !transportRecordFresh(proxy.UpdatedAt) {
		return "", false
	}
	readyAt, ready := client.ReadyProxies[proxyID]
	if !ready || !transportRecordFresh(readyAt) {
		return "", false
	}
	proxyIP, _ := transportPairIPs(proxyID, configID)
	return proxyIP, true
}

func (store *TransportStore) load() error {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read transport registry: %w", err)
	}
	var state transportState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("decode transport registry: %w", err)
	}
	if state.Proxies != nil {
		store.proxies = state.Proxies
	}
	if state.Clients != nil {
		store.clients = state.Clients
	}
	return nil
}

func (store *TransportStore) saveLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return fmt.Errorf("create transport directory: %w", err)
	}
	data, err := json.MarshalIndent(transportState{Proxies: store.proxies, Clients: store.clients}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode transport registry: %w", err)
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, data, 0o600); err != nil {
		return fmt.Errorf("write transport registry: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace transport registry: %w", err)
	}
	return nil
}

func (app *App) handleProxyTransport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload proxyTransportRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	node, err := controllerNodeBySecret(app.controllerDB, payload.Secret)
	if errors.Is(err, os.ErrNotExist) || node.Role != "proxy" {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid proxy node secret"})
		return
	}
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	endpoint := firstNodeIP(node)
	if endpoint == "" {
		endpoint = clientIP(request)
	}
	if err := app.transport.RegisterProxy(node.ID, endpoint, payload.SSHHostKey); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, app.transport.ProxyConfig(node.ID, app.ipStore.Records()))
}

func (app *App) handleClientTransport(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload clientTransportRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	record, ok := app.ipStore.GetByToken(strings.TrimSpace(payload.Token))
	if !ok {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "invalid enrollment token"})
		return
	}
	if err := app.transport.RegisterClient(record.ID, payload.SSHPublicKey, payload.ReadyProxies); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, app.transport.ClientConfig(record))
}

func controllerNodeBySecret(databasePath, secret string) (controllerNode, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return controllerNode{}, os.ErrNotExist
	}
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return controllerNode{}, fmt.Errorf("open controller database: %w", err)
	}
	defer database.Close()
	var node controllerNode
	err = database.QueryRow(`SELECT id, role, COALESCE(public_ip, ''), COALESCE(address, '') FROM nodes WHERE secret = ?`, secret).
		Scan(&node.ID, &node.Role, &node.PublicIP, &node.Address)
	if errors.Is(err, sql.ErrNoRows) {
		return controllerNode{}, os.ErrNotExist
	}
	if err != nil {
		return controllerNode{}, fmt.Errorf("find controller node: %w", err)
	}
	return node, nil
}

func normalizeSSHKey(value string) (string, error) {
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(value)))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(publicKey))), nil
}

func transportPairIPs(proxyID, configID string) (string, string) {
	sum := sha256.Sum256([]byte(proxyID + "\x00" + configID))
	network := (uint16(sum[0])<<8 | uint16(sum[1])) & 0xfffc
	third := byte(network >> 8)
	fourth := byte(network & 0xfc)
	return net.IPv4(10, 250, third, fourth+1).String(), net.IPv4(10, 250, third, fourth+2).String()
}

func transportRecordFresh(updatedAt time.Time) bool {
	return !updatedAt.IsZero() && time.Since(updatedAt) <= transportFreshness
}

func routesUseProxy(routes map[string]string, proxyID string) bool {
	for _, currentProxyID := range routes {
		if currentProxyID == proxyID {
			return true
		}
	}
	return false
}

func uniqueProxyIDs(routes map[string]string) []string {
	seen := make(map[string]struct{})
	for _, proxyID := range routes {
		if proxyID != "" {
			seen[proxyID] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for proxyID := range seen {
		result = append(result, proxyID)
	}
	sort.Strings(result)
	return result
}

func firstNodeIP(node controllerNode) string {
	for _, value := range []string{node.PublicIP, node.Address} {
		for _, candidate := range strings.FieldsFunc(value, func(character rune) bool {
			return character == ',' || character == ';' || character == ' ' || character == '\t'
		}) {
			candidate = strings.Trim(candidate, "[]")
			if host, _, err := net.SplitHostPort(candidate); err == nil {
				candidate = strings.Trim(host, "[]")
			}
			if parsed := net.ParseIP(candidate); parsed != nil {
				return parsed.String()
			}
		}
	}
	return ""
}
