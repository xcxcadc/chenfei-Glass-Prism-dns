package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	maxConnectivityDomains = 100
	connectivityWorkers    = 8
)

type ConnectivityRequest struct {
	DNSServer   string   `json:"dns_server"`
	ProxyServer string   `json:"proxy_server"`
	Domains     []string `json:"domains"`
	TimeoutMS   int      `json:"timeout_ms"`
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
	domains := routingDomains(request.Domains)
	if len(domains) == 0 {
		return nil, fmt.Errorf("at least one valid domain is required")
	}
	if len(domains) > maxConnectivityDomains {
		return nil, fmt.Errorf("too many domains: maximum is %d", maxConnectivityDomains)
	}
	timeout := time.Duration(request.TimeoutMS) * time.Millisecond
	if timeout < time.Second || timeout > 20*time.Second {
		timeout = 6 * time.Second
	}
	dnsServer := strings.TrimSpace(request.DNSServer)
	proxyServer := strings.TrimSpace(request.ProxyServer)
	if dnsServer == "" && proxyServer == "" {
		return nil, fmt.Errorf("dns server or proxy server is required")
	}
	if dnsServer != "" {
		if _, _, err := net.SplitHostPort(dnsServer); err != nil {
			dnsServer = net.JoinHostPort(strings.Trim(dnsServer, "[]"), "53")
		}
	}
	if proxyServer != "" {
		if _, _, err := net.SplitHostPort(proxyServer); err != nil {
			proxyServer = net.JoinHostPort(strings.Trim(proxyServer, "[]"), "443")
		}
	}
	if dnsServer == "" {
		dnsServer = "127.0.0.1:53"
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
	results := make([]DomainTestResult, len(domains))
	workers := make(chan struct{}, connectivityWorkers)
	var waitGroup sync.WaitGroup
	for index, domain := range domains {
		waitGroup.Add(1)
		go func(index int, domain string) {
			defer waitGroup.Done()
			workers <- struct{}{}
			defer func() { <-workers }()

			result := DomainTestResult{Domain: domain}
			target := proxyServer
			if target != "" {
				host, _, _ := net.SplitHostPort(target)
				result.Addresses = []string{strings.Trim(host, "[]")}
			} else {
				started := time.Now()
				addresses, err := resolver.LookupIPAddr(ctx, domain)
				result.DNSMS = time.Since(started).Milliseconds()
				if err != nil {
					result.Error = "DNS: " + err.Error()
					results[index] = result
					return
				}
				for _, address := range addresses {
					result.Addresses = append(result.Addresses, address.IP.String())
				}
				if len(addresses) == 0 {
					result.Error = "DNS: no address returned"
					results[index] = result
					return
				}
				target = net.JoinHostPort(addresses[0].IP.String(), "443")
			}

			tlsStarted := time.Now()
			connection, err := tls.DialWithDialer(
				&net.Dialer{Timeout: timeout},
				"tcp",
				target,
				&tls.Config{MinVersion: tls.VersionTLS12, ServerName: domain},
			)
			result.TLSMS = time.Since(tlsStarted).Milliseconds()
			if err != nil {
				result.Error = "TLS: " + err.Error()
				results[index] = result
				return
			}
			_ = connection.Close()
			result.Success = true
			results[index] = result
		}(index, domain)
	}
	waitGroup.Wait()
	return results, nil
}
