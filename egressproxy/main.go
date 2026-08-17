package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	maxInitialBytes = 64 * 1024
	probeBodyLimit  = 2 * 1024 * 1024
)

var (
	geminiRegionPattern = regexp.MustCompile(`,2,1,200,"([A-Z]{3})"`)
	blockedPagePattern  = regexp.MustCompile(`(?i)(not supported in your (country|region)|not available in your (country|region)|doesn.?t currently support your (country|region)|unsupported country|region blocked)`)
)

type servicePolicy struct {
	ID            string   `json:"service_id"`
	Name          string   `json:"name"`
	Domains       []string `json:"domains"`
	ProbeDomains  []string `json:"probe_domains"`
	IPv6Candidate bool     `json:"ipv6_candidate,omitempty"`
}

type policyConfig struct {
	Services []servicePolicy `json:"services"`
}

type probeResult struct {
	Reachable bool
	Code      int
	Blocked   bool
	Region    string
}

type familyPolicy struct {
	mu       sync.RWMutex
	path     string
	hash     string
	services []servicePolicy
	prefer6  map[string]bool
	refresh  chan struct{}
}

func newFamilyPolicy(path string) *familyPolicy {
	return &familyPolicy{path: path, prefer6: make(map[string]bool), refresh: make(chan struct{}, 1)}
}

func (policy *familyPolicy) preferIPv6(host string) bool {
	host = normalizeHost(host)
	policy.mu.RLock()
	defer policy.mu.RUnlock()
	for candidate := host; candidate != ""; {
		if prefer, exists := policy.prefer6[candidate]; exists {
			return prefer
		}
		index := strings.IndexByte(candidate, '.')
		if index < 0 {
			break
		}
		candidate = candidate[index+1:]
	}
	return false
}

func (policy *familyPolicy) requestRefresh() {
	select {
	case policy.refresh <- struct{}{}:
	default:
	}
}

func (policy *familyPolicy) run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	policy.refreshNow(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			policy.refreshNow(ctx)
		case <-policy.refresh:
			policy.refreshNow(ctx)
		}
	}
}

func (policy *familyPolicy) refreshNow(ctx context.Context) {
	data, err := os.ReadFile(policy.path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("read egress policy: %v", err)
		}
		return
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	policy.mu.RLock()
	unchanged := hash == policy.hash
	policy.mu.RUnlock()
	if unchanged {
		return
	}
	var config policyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("decode egress policy: %v", err)
		return
	}
	services := normalizeServices(config.Services)
	initial := make(map[string]bool)
	for _, service := range services {
		for _, domain := range service.Domains {
			initial[domain] = service.IPv6Candidate
		}
	}
	policy.mu.Lock()
	policy.hash = hash
	policy.services = services
	policy.prefer6 = initial
	policy.mu.Unlock()
	log.Printf("loaded %d egress service policies", len(services))

	go policy.evaluate(context.WithoutCancel(ctx), hash, services)
}

func (policy *familyPolicy) evaluate(ctx context.Context, hash string, services []servicePolicy) {
	type result struct {
		service servicePolicy
		prefer6 bool
	}
	results := make(chan result, len(services))
	workers := make(chan struct{}, 4)
	var waitGroup sync.WaitGroup
	for _, service := range services {
		service := service
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			workers <- struct{}{}
			defer func() { <-workers }()
			results <- result{service: service, prefer6: evaluateService(ctx, service)}
		}()
	}
	waitGroup.Wait()
	close(results)
	next := make(map[string]bool)
	for item := range results {
		for _, domain := range item.service.Domains {
			if item.prefer6 {
				next[domain] = true
			} else if _, exists := next[domain]; !exists {
				next[domain] = false
			}
		}
		log.Printf("egress family service=%q ipv6=%t", item.service.Name, item.prefer6)
	}
	policy.mu.Lock()
	defer policy.mu.Unlock()
	if policy.hash == hash {
		policy.prefer6 = next
	}
}

func normalizeServices(services []servicePolicy) []servicePolicy {
	result := make([]servicePolicy, 0, len(services))
	for _, service := range services {
		service.ID = strings.TrimSpace(service.ID)
		service.Name = strings.TrimSpace(service.Name)
		service.Domains = normalizeDomains(service.Domains)
		service.ProbeDomains = normalizeDomains(service.ProbeDomains)
		if service.ID == "" || len(service.Domains) == 0 {
			continue
		}
		result = append(result, service)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].ID < result[right].ID })
	return result
}

func normalizeDomains(domains []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = normalizeHost(domain)
		domain = strings.TrimPrefix(domain, "*.")
		if domain == "" || net.ParseIP(domain) != nil {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	sort.Strings(result)
	return result
}

func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(host, ".")))
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	return strings.Trim(host, "[]")
}

func evaluateService(ctx context.Context, service servicePolicy) bool {
	for _, domain := range service.ProbeDomains {
		probeCtx, cancel := context.WithTimeout(ctx, 14*time.Second)
		var ipv4, ipv6 probeResult
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		go func() { defer waitGroup.Done(); ipv4 = probeDomain(probeCtx, domain, false) }()
		go func() { defer waitGroup.Done(); ipv6 = probeDomain(probeCtx, domain, true) }()
		waitGroup.Wait()
		cancel()
		if preferIPv6Result(service, ipv4, ipv6) {
			return true
		}
		if ipv4.Reachable && !service.IPv6Candidate {
			return false
		}
	}
	return false
}

func preferIPv6Result(service servicePolicy, ipv4, ipv6 probeResult) bool {
	if !service.IPv6Candidate {
		return false
	}
	if !ipv6.Reachable {
		return false
	}
	if !ipv4.Reachable {
		return true
	}
	if ipv4.Blocked && !ipv6.Blocked {
		return true
	}
	if (ipv4.Code == http.StatusForbidden || ipv4.Code == http.StatusUnavailableForLegalReasons || ipv4.Code >= 500) &&
		ipv6.Code >= 200 && ipv6.Code < 500 && ipv6.Code != http.StatusForbidden && ipv6.Code != http.StatusUnavailableForLegalReasons {
		return true
	}
	if service.IPv6Candidate && ipv4.Region == "CHN" && ipv6.Region != "" && ipv6.Region != "CHN" {
		return true
	}
	return false
}

func probeDomain(ctx context.Context, domain string, ipv6 bool) probeResult {
	transport := &http.Transport{
		Proxy:               nil,
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 6 * time.Second,
		DialContext: func(dialContext context.Context, _, address string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			return dialHost(dialContext, domain, port, ipv6, true)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 12 * time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+domain+"/", nil)
	if err != nil {
		return probeResult{}
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/138 Safari/537.36")
	response, err := client.Do(request)
	if err != nil {
		return probeResult{}
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, probeBodyLimit))
	result := probeResult{Reachable: true, Code: response.StatusCode, Blocked: blockedPagePattern.Match(body)}
	if match := geminiRegionPattern.FindSubmatch(body); len(match) == 2 {
		result.Region = string(match[1])
	}
	return result
}

type proxyServer struct {
	policy      *familyPolicy
	dialTimeout time.Duration
}

func (server *proxyServer) serve(ctx context.Context, address string, port string, hostname func([]byte) (string, error)) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	log.Printf("listening on %s", address)
	for {
		connection, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("accept %s: %v", address, err)
			continue
		}
		go server.handle(connection, port, hostname)
	}
}

func (server *proxyServer) handle(client net.Conn, port string, hostname func([]byte) (string, error)) {
	defer client.Close()
	_ = client.SetReadDeadline(time.Now().Add(10 * time.Second))
	initial, err := readInitial(client, port)
	if err != nil {
		log.Printf("read client preface: %v", err)
		return
	}
	host, err := hostname(initial)
	if err != nil {
		log.Printf("parse destination: %v", err)
		return
	}
	_ = client.SetReadDeadline(time.Time{})
	ctx, cancel := context.WithTimeout(context.Background(), server.dialTimeout)
	upstream, err := dialHost(ctx, host, port, server.policy.preferIPv6(host), false)
	cancel()
	if err != nil {
		log.Printf("dial %s:%s: %v", host, port, err)
		return
	}
	defer upstream.Close()
	if _, err := upstream.Write(initial); err != nil {
		return
	}
	relay(client, upstream)
}

func readInitial(connection net.Conn, port string) ([]byte, error) {
	if port == "443" {
		header := make([]byte, 5)
		if _, err := io.ReadFull(connection, header); err != nil {
			return nil, err
		}
		length := int(binary.BigEndian.Uint16(header[3:5]))
		if header[0] != 22 || length <= 0 || length > maxInitialBytes-5 {
			return nil, errors.New("invalid TLS client hello")
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(connection, payload); err != nil {
			return nil, err
		}
		return append(header, payload...), nil
	}
	buffer := make([]byte, 0, 4096)
	chunk := make([]byte, 1)
	for len(buffer) < maxInitialBytes {
		if _, err := io.ReadFull(connection, chunk); err != nil {
			return nil, err
		}
		buffer = append(buffer, chunk[0])
		if len(buffer) >= 4 && bytes.Equal(buffer[len(buffer)-4:], []byte("\r\n\r\n")) {
			return buffer, nil
		}
	}
	return nil, errors.New("HTTP headers exceed limit")
}

func httpHostname(data []byte) (string, error) {
	lines := strings.Split(string(data), "\r\n")
	for _, line := range lines[1:] {
		name, value, found := strings.Cut(line, ":")
		if found && strings.EqualFold(strings.TrimSpace(name), "host") {
			host := normalizeHost(value)
			if host != "" {
				return host, nil
			}
		}
	}
	return "", errors.New("HTTP Host header is missing")
}

func tlsHostname(data []byte) (string, error) {
	if len(data) < 9 || data[0] != 22 || data[5] != 1 {
		return "", errors.New("TLS ClientHello is missing")
	}
	payload := data[9:]
	if len(payload) < 34 {
		return "", errors.New("TLS ClientHello is truncated")
	}
	offset := 34
	if offset >= len(payload) {
		return "", errors.New("TLS session ID is missing")
	}
	offset += 1 + int(payload[offset])
	if offset+2 > len(payload) {
		return "", errors.New("TLS cipher suites are missing")
	}
	cipherLength := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2 + cipherLength
	if offset >= len(payload) {
		return "", errors.New("TLS compression methods are missing")
	}
	offset += 1 + int(payload[offset])
	if offset+2 > len(payload) {
		return "", errors.New("TLS extensions are missing")
	}
	extensionLength := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2
	end := offset + extensionLength
	if end > len(payload) {
		return "", errors.New("TLS extensions are truncated")
	}
	for offset+4 <= end {
		extensionType := binary.BigEndian.Uint16(payload[offset : offset+2])
		length := int(binary.BigEndian.Uint16(payload[offset+2 : offset+4]))
		offset += 4
		if offset+length > end {
			return "", errors.New("TLS extension is truncated")
		}
		if extensionType == 0 {
			serverNames := payload[offset : offset+length]
			if len(serverNames) < 5 {
				return "", errors.New("TLS SNI is truncated")
			}
			nameLength := int(binary.BigEndian.Uint16(serverNames[3:5]))
			if serverNames[2] != 0 || 5+nameLength > len(serverNames) {
				return "", errors.New("TLS SNI hostname is invalid")
			}
			host := normalizeHost(string(serverNames[5 : 5+nameLength]))
			if host != "" {
				return host, nil
			}
		}
		offset += length
	}
	return "", errors.New("TLS SNI hostname is missing")
}

func dialHost(ctx context.Context, host, port string, prefer6, singleFamily bool) (net.Conn, error) {
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	filtered := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		ip := address.IP
		if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		isIPv6 := ip.To4() == nil
		if singleFamily && isIPv6 != prefer6 {
			continue
		}
		filtered = append(filtered, ip)
	}
	sort.SliceStable(filtered, func(left, right int) bool {
		leftIPv6 := filtered[left].To4() == nil
		rightIPv6 := filtered[right].To4() == nil
		if leftIPv6 == rightIPv6 {
			return filtered[left].String() < filtered[right].String()
		}
		return leftIPv6 == prefer6
	})
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no usable address for %s", host)
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	var failures []string
	for _, ip := range filtered {
		connection, dialErr := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return connection, nil
		}
		failures = append(failures, ip.String()+": "+dialErr.Error())
	}
	return nil, errors.New(strings.Join(failures, "; "))
}

func relay(left, right net.Conn) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	copyConnection := func(destination, source net.Conn) {
		defer waitGroup.Done()
		_, _ = io.Copy(destination, source)
		if tcp, ok := destination.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
	}
	go copyConnection(left, right)
	go copyConnection(right, left)
	waitGroup.Wait()
}

func main() {
	var tlsListen, httpListen, policyFile string
	var dialTimeout time.Duration
	flag.StringVar(&tlsListen, "tls-listen", "127.0.0.1:19443", "TLS SNI listen address")
	flag.StringVar(&httpListen, "http-listen", "127.0.0.1:19080", "HTTP Host listen address")
	flag.StringVar(&policyFile, "policy", "/etc/prismdns/egress-policy.json", "service egress policy file")
	flag.DurationVar(&dialTimeout, "dial-timeout", 20*time.Second, "upstream dial timeout")
	flag.Parse()
	if dialTimeout < time.Second || dialTimeout > time.Minute {
		log.Fatal("dial-timeout must be between 1s and 1m")
	}
	for _, address := range []string{tlsListen, httpListen} {
		host, port, err := net.SplitHostPort(address)
		if err != nil || (host != "127.0.0.1" && host != "::1") {
			log.Fatalf("listen address must use loopback: %s", address)
		}
		if parsedPort, parseErr := strconv.Atoi(port); parseErr != nil || parsedPort < 1 || parsedPort > 65535 {
			log.Fatalf("invalid listen port: %s", address)
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	policy := newFamilyPolicy(policyFile)
	go policy.run(ctx)
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-hup:
				policy.mu.Lock()
				policy.hash = ""
				policy.mu.Unlock()
				policy.requestRefresh()
			}
		}
	}()

	server := &proxyServer{policy: policy, dialTimeout: dialTimeout}
	errorsChannel := make(chan error, 2)
	go func() { errorsChannel <- server.serve(ctx, tlsListen, "443", tlsHostname) }()
	go func() { errorsChannel <- server.serve(ctx, httpListen, "80", httpHostname) }()
	if err := <-errorsChannel; err != nil {
		log.Fatal(err)
	}
}
