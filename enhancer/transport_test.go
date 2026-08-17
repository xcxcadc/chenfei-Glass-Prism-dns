package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestTransportStoreBuildsReadyPair(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transports.json")
	store, err := NewTransportStore(path)
	if err != nil {
		t.Fatal(err)
	}
	proxyKey := testSSHKey(t)
	clientKey := testSSHKey(t)
	if err := store.RegisterProxy("proxy-a", "203.0.113.10", proxyKey, true); err != nil {
		t.Fatal(err)
	}
	if err := store.RegisterClient("ip-a", clientKey, nil); err != nil {
		t.Fatal(err)
	}
	record := ipConfigRecord{IPConfig: IPConfig{ID: "ip-a", IP: "198.51.100.30", Routes: map[string]string{"youtube": "proxy-a"}}}
	clientConfig := store.ClientConfig(record)
	if len(clientConfig.Peers) != 1 || clientConfig.Peers[0].SSHHost != "203.0.113.10" || clientConfig.Peers[0].SSHPort != 22 ||
		clientConfig.Peers[0].RemoteHTTP != 19080 || clientConfig.Peers[0].RemoteHTTPS != 19443 {
		t.Fatalf("unexpected client config: %#v", clientConfig)
	}
	proxyConfig := store.ProxyConfig("proxy-a", []ipConfigRecord{record})
	if len(proxyConfig.Peers) != 1 || !strings.HasPrefix(proxyConfig.Peers[0].SSHPublicKey, "ssh-ed25519 ") {
		t.Fatalf("unexpected proxy config: %#v", proxyConfig)
	}
	if len(proxyConfig.AuthorizedIPs) != 1 || proxyConfig.AuthorizedIPs[0] != "198.51.100.30" {
		t.Fatalf("panel-managed target IP was not authorized: %#v", proxyConfig)
	}
	if _, ready := store.EffectiveProxyIP("ip-a", "proxy-a"); ready {
		t.Fatal("transport must not replace DNS before the client reports readiness")
	}
	readyProxies := []string{"proxy-a"}
	if err := store.RegisterClient("ip-a", clientKey, &readyProxies); err != nil {
		t.Fatal(err)
	}
	transportIP, ready := store.EffectiveProxyIP("ip-a", "proxy-a")
	if !ready || transportIP != clientConfig.Peers[0].ProxyIP {
		t.Fatalf("unexpected ready transport: %q %v", transportIP, ready)
	}
	reloaded, err := NewTransportStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloadedTransportIP, ok := reloaded.EffectiveProxyIP("ip-a", "proxy-a"); !ok || reloadedTransportIP != transportIP {
		t.Fatalf("transport readiness was not persisted: %q %v", reloadedTransportIP, ok)
	}
}

func TestTransportPairIPsAreStableAndScoped(t *testing.T) {
	firstProxy, firstClient := transportPairIPs("proxy-a", "ip-a")
	repeatedProxy, repeatedClient := transportPairIPs("proxy-a", "ip-a")
	otherProxy, otherClient := transportPairIPs("proxy-b", "ip-a")
	if firstProxy != repeatedProxy || firstClient != repeatedClient {
		t.Fatal("transport pair allocation is not stable")
	}
	if firstProxy == firstClient || firstProxy == otherProxy || firstClient == otherClient {
		t.Fatalf("transport pair allocation collided: %s %s %s %s", firstProxy, firstClient, otherProxy, otherClient)
	}
}

func TestControllerNodeBySecret(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "data.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TABLE nodes (
		id TEXT PRIMARY KEY,
		role TEXT,
		secret TEXT,
		public_ip TEXT,
		address TEXT
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO nodes (id, role, secret, public_ip, address) VALUES (?, ?, ?, ?, ?)`,
		"node-a", "proxy", "secret-a", "203.0.113.20", "proxy.example"); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	node, err := controllerNodeBySecret(databasePath, "secret-a")
	if err != nil {
		t.Fatal(err)
	}
	if node.ID != "node-a" || node.Role != "proxy" || firstNodeIPv4(node) != "203.0.113.20" {
		t.Fatalf("unexpected node: %#v", node)
	}
}

func TestFirstNodeIPv4NeverSelectsIPv6(t *testing.T) {
	node := controllerNode{
		PublicIP: "2001:db8::20, 203.0.113.20",
		Address:  "[2001:db8::21]:22 198.51.100.21:22",
	}
	if result := firstNodeIPv4(node); result != "203.0.113.20" {
		t.Fatalf("expected proxy IPv4 preference, got %q", result)
	}
	node.PublicIP = "2001:db8::20"
	node.Address = "[2001:db8::21]:22"
	if result := firstNodeIPv4(node); result != "" {
		t.Fatalf("IPv6-only proxy must not be registered for unlock traffic: %q", result)
	}
}

func testSSHKey(t *testing.T) string {
	t.Helper()
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(ssh.MarshalAuthorizedKey(sshPublicKey))
}
