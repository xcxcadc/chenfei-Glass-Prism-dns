package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

type ConnectivityRequest struct {
	DNSServer string   `json:"dns_server"`
	Domains   []string `json:"domains"`
	TimeoutMS int      `json:"timeout_ms"`
}

type DomainTestResult struct {
	Domain    string   `json:"domain"`
	Addresses []string `json:"addresses,omitempty"`
	DNSMS     int64    `json:"dns_ms"`
	TLSMS     int64    `json:"tls_ms,omitempty"`
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
}

func TestConnectivity(ctx context.Context, request ConnectivityRequest) ([]DomainTestResult, error) {
	domains := normalizeDomains(request.Domains)
	if len(domains) == 0 {
		return nil, fmt.Errorf("at least one valid domain is required")
	}
	if len(domains) > 10 {
		domains = domains[:10]
	}
	timeout := time.Duration(request.TimeoutMS) * time.Millisecond
	if timeout < time.Second || timeout > 20*time.Second {
		timeout = 6 * time.Second
	}
	dnsServer := strings.TrimSpace(request.DNSServer)
	if dnsServer == "" {
		return nil, fmt.Errorf("dns server is required")
	}
	if _, _, err := net.SplitHostPort(dnsServer); err != nil {
		dnsServer = net.JoinHostPort(strings.Trim(dnsServer, "[]"), "53")
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(dialContext context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: timeout}).DialContext(dialContext, "udp", dnsServer)
		},
	}
	results := make([]DomainTestResult, 0, len(domains))
	for _, domain := range domains {
		result := DomainTestResult{Domain: domain}
		started := time.Now()
		addresses, err := resolver.LookupIPAddr(ctx, domain)
		result.DNSMS = time.Since(started).Milliseconds()
		if err != nil {
			result.Error = "DNS: " + err.Error()
			results = append(results, result)
			continue
		}
		for _, address := range addresses {
			result.Addresses = append(result.Addresses, address.IP.String())
		}
		if len(addresses) == 0 {
			result.Error = "DNS: no address returned"
			results = append(results, result)
			continue
		}

		tlsStarted := time.Now()
		connection, err := tls.DialWithDialer(
			&net.Dialer{Timeout: timeout},
			"tcp",
			net.JoinHostPort(addresses[0].IP.String(), "443"),
			&tls.Config{MinVersion: tls.VersionTLS12, ServerName: domain},
		)
		result.TLSMS = time.Since(tlsStarted).Milliseconds()
		if err != nil {
			result.Error = "TLS: " + err.Error()
			results = append(results, result)
			continue
		}
		_ = connection.Close()
		result.Success = true
		results = append(results, result)
	}
	return results, nil
}
