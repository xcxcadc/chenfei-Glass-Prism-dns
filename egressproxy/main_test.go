package main

import (
	"crypto/tls"
	"io"
	"net"
	"testing"
	"time"
)

func TestHTTPHostname(t *testing.T) {
	host, err := httpHostname([]byte("GET / HTTP/1.1\r\nHost: Gemini.Google.com:80\r\n\r\n"))
	if err != nil || host != "gemini.google.com" {
		t.Fatalf("unexpected HTTP host %q: %v", host, err)
	}
}

func TestTLSHostname(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	deadline := time.Now().Add(2 * time.Second)
	_ = client.SetDeadline(deadline)
	_ = server.SetDeadline(deadline)
	go func() {
		connection := tls.Client(client, &tls.Config{ServerName: "gemini.google.com", MinVersion: tls.VersionTLS12})
		_ = connection.Handshake()
	}()
	header := make([]byte, 5)
	if _, err := io.ReadFull(server, header); err != nil {
		t.Fatal(err)
	}
	length := int(header[3])<<8 | int(header[4])
	payload := make([]byte, length)
	if _, err := io.ReadFull(server, payload); err != nil {
		t.Fatal(err)
	}
	host, err := tlsHostname(append(header, payload...))
	if err != nil || host != "gemini.google.com" {
		t.Fatalf("unexpected TLS host %q: %v", host, err)
	}
}

func TestFamilyPolicyMatchesDomainSuffix(t *testing.T) {
	policy := newFamilyPolicy("")
	policy.prefer6 = map[string]bool{"googleapis.com": true, "example.com": false}
	if !policy.preferIPv6("generativelanguage.googleapis.com") {
		t.Fatal("expected suffix policy to prefer IPv6")
	}
	if policy.preferIPv6("cdn.example.com") || policy.preferIPv6("unmanaged.test") {
		t.Fatal("IPv4 must remain the default")
	}
}

func TestPreferIPv6OnlyWhenItImprovesThePath(t *testing.T) {
	service := servicePolicy{IPv6Candidate: true}
	if !preferIPv6Result(service,
		probeResult{Reachable: true, Code: 200, Region: "CHN"},
		probeResult{Reachable: true, Code: 200, Region: "SGP"}) {
		t.Fatal("expected Gemini region improvement to select IPv6")
	}
	if preferIPv6Result(service,
		probeResult{Reachable: true, Code: 200, Region: "SGP"},
		probeResult{Reachable: true, Code: 200, Region: "SGP"}) {
		t.Fatal("IPv4 must remain preferred when both families are equivalent")
	}
	if !preferIPv6Result(service,
		probeResult{Reachable: true, Code: 403},
		probeResult{Reachable: true, Code: 200}) {
		t.Fatal("expected a blocked IPv4 path to select IPv6")
	}
}

func TestNormalizeServices(t *testing.T) {
	services := normalizeServices([]servicePolicy{{
		ID: " gemini ", Name: " Gemini ", Domains: []string{"*.Google.com", "google.com", ""},
		ProbeDomains: []string{"Gemini.Google.com"},
	}})
	if len(services) != 1 || services[0].Domains[0] != "google.com" {
		t.Fatalf("unexpected normalized services: %#v", services)
	}
}
