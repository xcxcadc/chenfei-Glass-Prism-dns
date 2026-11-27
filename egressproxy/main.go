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
	geminiRegionPattern   = regexp.MustCompile(`,2,1,200,"([A-Z]{2,3})"`)
	geminiTruePattern     = regexp.MustCompile(`(?:"45631641"|45631641\\"):true|45631641,null,true`)
	geminiFalsePattern    = regexp.MustCompile(`(?:"45631641"|45631641\\"):false|45631641,null,false`)
	playCountryPattern    = regexp.MustCompile(`<div class="yVZQTb">([^<(]+)`)
	youtubeCNPattern      = regexp.MustCompile(`"countryCode"\s*:\s*"CN"|"INNERTUBE_CONTEXT_GL"\s*:\s*"CN"`)
	blockedPagePattern    = regexp.MustCompile(`(?i)(not supported in your (country|region)|not available in your (country|region)|doesn.?t currently support your (country|region)|unsupported country|region blocked|app-unavailable-in-region)`)
	geminiBlockedRegions  = map[string]struct{}{
		"CHN": {}, "CN": {}, "HKG": {}, "HK": {}, "MAC": {}, "MO": {},
		"RUS": {}, "RU": {}, "BLR": {}, "BY": {}, "IRN": {}, "IR": {},
		"PRK": {}, "KP": {}, "SYR": {}, "SY": {}, "CUB": {}, "CU": {},
	}
	fallbackDNSServers = []string{"1.1.1.1:53", "8.8.8.8:53"}
)

type servicePolicy struct {
	ID             string   `json:"service_id"`
	Name           string   `json:"name"`
	Domains        []string `json:"domains"`
	DomainKeywords []string `json:"domain_keywords,omitempty"`
	CIDRs          []string `json:"cidrs,omitempty"`
	ProxyIP        string   `json:"proxy_ip,omitempty"`
	ProbeDomains   []string `json:"probe_domains"`
	IPv6Candidate  bool     `json:"ipv6_candidate,omitempty"`
	EgressMode     string   `json:"egress_mode,omitempty"`
}

const (
	egressNative4 = ""
	egressNative6 = "native6"
	egressWarp6   = "warp6"
	egressWarp4   = "warp4"
)

type policyConfig struct {
	Services []servicePolicy `json:"services"`
}

type probeResult struct {
	Reachable bool
	Code      int
	Blocked   bool
	Region    string
}

type cidrRoute struct {
	net     *net.IPNet
	proxyIP string
}

type keywordRoute struct {
	keyword string
	proxyIP string
}

type familyPolicy struct {
	mu           sync.RWMutex
	path         string
	warpSocks    string
	hash         string
	services     []servicePolicy
	prefer6      map[string]bool
	allow6       map[string]bool
	egressModes  map[string]string
	routeIPs     map[string]string
	keywords     []keywordRoute
	cidrs        []cidrRoute
	knownProxies    map[string]struct{}
	refresh         chan struct{}
	warpFailCount   int
	warpBypassUntil time.Time
}

func newFamilyPolicy(path string) *familyPolicy {
	return &familyPolicy{
		path:         path,
		prefer6:      make(map[string]bool),
		allow6:       make(map[string]bool),
		egressModes:  make(map[string]string),
		routeIPs:     make(map[string]string),
		knownProxies: make(map[string]struct{}),
		refresh:      make(chan struct{}, 1),
	}
}

func (policy *familyPolicy) shouldBypassWARP() bool {
	policy.mu.RLock()
	until := policy.warpBypassUntil
	policy.mu.RUnlock()
	return !until.IsZero() && time.Now().Before(until)
}

func (policy *familyPolicy) recordWARPResult(ok bool) {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	if ok {
		policy.warpFailCount = 0
		policy.warpBypassUntil = time.Time{}
		return
	}
	policy.warpFailCount++
	if policy.warpFailCount >= 2 {
		policy.warpBypassUntil = time.Now().Add(25 * time.Second)
		policy.warpFailCount = 0
		log.Printf("warp circuit breaker tripped: bypassing unhealthy WARP to native for 25s")
	}
}

func (policy *familyPolicy) egressModeFor(host string) string {
	host = normalizeHost(host)
	policy.mu.RLock()
	defer policy.mu.RUnlock()
	for candidate := host; candidate != ""; {
		if mode, exists := policy.egressModes[candidate]; exists && mode != "" {
			return mode
		}
		index := strings.IndexByte(candidate, '.')
		if index < 0 {
			break
		}
		candidate = candidate[index+1:]
	}
	return egressNative4
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

func (policy *familyPolicy) allowIPv6Fallback(host string) bool {
	host = normalizeHost(host)
	policy.mu.RLock()
	defer policy.mu.RUnlock()
	for candidate := host; candidate != ""; {
		if allow, exists := policy.allow6[candidate]; exists {
			return allow
		}
		index := strings.IndexByte(candidate, '.')
		if index < 0 {
			break
		}
		candidate = candidate[index+1:]
	}
	return false
}

func (policy *familyPolicy) routeIP(host string) string {
	return policy.routeDestination(host, nil)
}

func (policy *familyPolicy) routeDestination(host string, dstIP net.IP) string {
	host = normalizeHost(host)
	policy.mu.RLock()
	defer policy.mu.RUnlock()
	for candidate := host; candidate != ""; {
		if routeIP, exists := policy.routeIPs[candidate]; exists {
			return routeIP
		}
		index := strings.IndexByte(candidate, '.')
		if index < 0 {
			break
		}
		candidate = candidate[index+1:]
	}
	if host != "" {
		for _, item := range policy.keywords {
			if strings.Contains(host, item.keyword) {
				return item.proxyIP
			}
		}
	}
	if dstIP != nil {
		for _, item := range policy.cidrs {
			if item.net.Contains(dstIP) {
				return item.proxyIP
			}
		}
	}
	return ""
}

func (policy *familyPolicy) isKnownProxyIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	policy.mu.RLock()
	_, ok := policy.knownProxies[ip.String()]
	policy.mu.RUnlock()
	return ok
}

func (policy *familyPolicy) requestRefresh() {
	select {
	case policy.refresh <- struct{}{}:
	default:
	}
}

func (policy *familyPolicy) run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
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
		if !errors.Is(os.ErrNotExist, err) && !errors.Is(err, os.ErrNotExist) {
			log.Printf("read egress policy: %v", err)
		}
		return
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	var config policyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("decode egress policy: %v", err)
		return
	}
	services := normalizeServices(config.Services)

	policy.mu.Lock()
	unchanged := hash == policy.hash
	previousPrefer6 := policy.prefer6
	previousModes := policy.egressModes
	initial := make(map[string]bool)
	allow6 := make(map[string]bool)
	initialModes := make(map[string]string)
	routeIPs := make(map[string]string)
	knownProxies := make(map[string]struct{})
	var keywords []keywordRoute
	var cidrs []cidrRoute
	hasClientRoutes := false

	for _, service := range services {
		proxyIP := ""
		if parsed := net.ParseIP(strings.TrimSpace(service.ProxyIP)); parsed != nil && parsed.To4() != nil {
			proxyIP = parsed.To4().String()
			knownProxies[proxyIP] = struct{}{}
			hasClientRoutes = true
		}
		explicitMode := strings.ToLower(strings.TrimSpace(service.EgressMode))
		for _, domain := range service.Domains {
			if service.IPv6Candidate && previousPrefer6[domain] {
				initial[domain] = true
			} else {
				initial[domain] = false
			}
			if service.IPv6Candidate {
				allow6[domain] = true
			} else if _, exists := allow6[domain]; !exists {
				allow6[domain] = false
			}
			switch explicitMode {
			case egressWarp6:
				initialModes[domain] = egressWarp6
			case egressWarp4:
				initialModes[domain] = egressWarp4
			case "warp":
				if prev := previousModes[domain]; prev == egressWarp6 || prev == egressWarp4 {
					initialModes[domain] = prev
				} else {
					initialModes[domain] = egressWarp6
				}
			case "auto":
				if prev := previousModes[domain]; prev != "" {
					initialModes[domain] = prev
				}
			default:
				if service.IPv6Candidate && previousPrefer6[domain] {
					initialModes[domain] = egressNative6
				}
			}
			if proxyIP != "" {
				routeIPs[domain] = proxyIP
			}
		}
		if proxyIP != "" {
			for _, kw := range service.DomainKeywords {
				keywords = append(keywords, keywordRoute{keyword: kw, proxyIP: proxyIP})
			}
			for _, rawCIDR := range service.CIDRs {
				if _, ipNet, parseErr := net.ParseCIDR(rawCIDR); parseErr == nil && ipNet != nil {
					cidrs = append(cidrs, cidrRoute{net: ipNet, proxyIP: proxyIP})
				}
			}
		}
	}

	policy.hash = hash
	policy.services = services
	policy.prefer6 = initial
	policy.allow6 = allow6
	policy.egressModes = initialModes
	policy.routeIPs = routeIPs
	policy.keywords = keywords
	policy.cidrs = cidrs
	policy.knownProxies = knownProxies
	policy.mu.Unlock()

	if !unchanged {
		log.Printf("loaded %d egress service policies (client_routes=%t, warp_socks=%q)", len(services), hasClientRoutes, policy.warpSocks)
	}

	// Client-side policies forward to ProxyIPv4 and do not need local IPv4/IPv6/WARP probing.
	if hasClientRoutes {
		return
	}
	go policy.evaluate(context.WithoutCancel(ctx), hash, services)
}

func (policy *familyPolicy) evaluate(ctx context.Context, hash string, services []servicePolicy) {
	type result struct {
		service servicePolicy
		mode    string
		prefer6 bool
	}
	warpUp := warpSocksReachable(ctx, policy.warpSocks)
	results := make(chan result, len(services))
	workers := make(chan struct{}, 4)
	var waitGroup sync.WaitGroup
	for _, service := range services {
		service := service
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			explicit := strings.ToLower(strings.TrimSpace(service.EgressMode))
			switch explicit {
			case egressWarp6:
				results <- result{service: service, mode: egressWarp6}
				return
			case egressWarp4:
				results <- result{service: service, mode: egressWarp4}
				return
			case "warp":
				if !warpUp {
					results <- result{service: service, mode: egressWarp6}
					return
				}
				workers <- struct{}{}
				defer func() { <-workers }()
				chosen := evaluateWarpChoice(ctx, service, policy.warpSocks)
				results <- result{service: service, mode: chosen}
				return
			case "auto":
				workers <- struct{}{}
				defer func() { <-workers }()
				chosen := evaluateAutoService(ctx, service, policy.warpSocks, warpUp)
				results <- result{service: service, mode: chosen, prefer6: chosen == egressNative6}
				return
			}
			if !service.IPv6Candidate {
				results <- result{service: service, mode: egressNative4, prefer6: false}
				return
			}
			workers <- struct{}{}
			defer func() { <-workers }()
			p6 := evaluateService(ctx, service)
			m := egressNative4
			if p6 {
				m = egressNative6
			}
			results <- result{service: service, mode: m, prefer6: p6}
		}()
	}
	waitGroup.Wait()
	close(results)
	nextPrefer6 := make(map[string]bool)
	nextModes := make(map[string]string)
	for item := range results {
		for _, domain := range item.service.Domains {
			if item.prefer6 {
				nextPrefer6[domain] = true
			} else if _, exists := nextPrefer6[domain]; !exists {
				nextPrefer6[domain] = false
			}
			if item.mode == egressWarp6 || item.mode == egressWarp4 {
				nextModes[domain] = item.mode
			} else if item.mode != "" {
				if cur := nextModes[domain]; cur != egressWarp6 && cur != egressWarp4 {
					nextModes[domain] = item.mode
				}
			}
		}
		if item.service.IPv6Candidate || item.service.EgressMode != "" {
			modeLabel := item.mode
			if modeLabel == "" {
				modeLabel = "native4"
			}
			log.Printf("egress route service=%q policy=%q selected=%s", item.service.Name, item.service.EgressMode, modeLabel)
		}
	}
	policy.mu.Lock()
	defer policy.mu.Unlock()
	if policy.hash == hash {
		policy.prefer6 = nextPrefer6
		policy.egressModes = nextModes
	}
}

func normalizeServices(services []servicePolicy) []servicePolicy {
	result := make([]servicePolicy, 0, len(services))
	for _, service := range services {
		service.ID = strings.TrimSpace(service.ID)
		service.Name = strings.TrimSpace(service.Name)
		service.EgressMode = strings.ToLower(strings.TrimSpace(service.EgressMode))
		service.Domains = normalizeDomains(service.Domains)
		service.ProbeDomains = normalizeDomains(service.ProbeDomains)
		service.DomainKeywords = normalizeKeywords(service.DomainKeywords)
		service.CIDRs = normalizeCIDRList(service.CIDRs)
		if service.ID == "" || (len(service.Domains) == 0 && len(service.DomainKeywords) == 0 && len(service.CIDRs) == 0) {
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

func normalizeKeywords(keywords []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, kw := range keywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw == "" {
			continue
		}
		if _, ok := seen[kw]; ok {
			continue
		}
		seen[kw] = struct{}{}
		result = append(result, kw)
	}
	sort.Strings(result)
	return result
}

func normalizeCIDRList(cidrs []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, raw := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err != nil || network == nil {
			continue
		}
		val := network.String()
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		result = append(result, val)
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

func representativeProbeTargets(service servicePolicy) []string {
	nameLower := strings.ToLower(service.Name)
	switch {
	case strings.Contains(nameLower, "gemini") || strings.Contains(nameLower, "ai studio"):
		return []string{"gemini.google.com"}
	case strings.Contains(nameLower, "youtube"):
		return []string{"www.youtube.com"}
	case strings.Contains(nameLower, "play store"):
		return []string{"play.google.com"}
	case strings.Contains(nameLower, "claude"):
		return []string{"claude.ai"}
	case strings.Contains(nameLower, "chatgpt") || strings.Contains(nameLower, "openai"):
		return []string{"chatgpt.com"}
	}
	if len(service.ProbeDomains) > 2 {
		return service.ProbeDomains[:2]
	}
	return service.ProbeDomains
}

func evaluateService(ctx context.Context, service servicePolicy) bool {
	if !service.IPv6Candidate {
		return false
	}
	for _, domain := range representativeProbeTargets(service) {
		probeCtx, cancel := context.WithTimeout(ctx, 14*time.Second)
		var ipv4, ipv6 probeResult
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		go func() { defer waitGroup.Done(); ipv4 = probeDomain(probeCtx, domain, false) }()
		go func() { defer waitGroup.Done(); ipv6 = probeDomain(probeCtx, domain, true) }()
		waitGroup.Wait()
		cancel()
		log.Printf("probe service=%q domain=%s v4={ok:%t code:%d blocked:%t region:%s} v6={ok:%t code:%d blocked:%t region:%s}",
			service.Name, domain,
			ipv4.Reachable, ipv4.Code, ipv4.Blocked, ipv4.Region,
			ipv6.Reachable, ipv6.Code, ipv6.Blocked, ipv6.Region)
		if preferIPv6Result(service, ipv4, ipv6) {
			return true
		}
		if ipv4.Reachable && !ipv4.Blocked && ipv4.Code >= 200 && ipv4.Code < 400 && ipv4.Region != "CHN" && ipv4.Region != "CN" {
			return false
		}
	}
	return false
}

func preferIPv6Result(service servicePolicy, ipv4, ipv6 probeResult) bool {
	if !service.IPv6Candidate {
		return false
	}
	if !ipv6.Reachable || ipv6.Blocked {
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
	if (ipv4.Region == "CHN" || ipv4.Region == "CN") && ipv6.Region != "CHN" && ipv6.Region != "CN" && ipv6.Code >= 200 && ipv6.Code < 400 {
		return true
	}
	return false
}

func probeURLForDomain(domain string) (string, string) {
	switch domain {
	case "gemini.google.com":
		return "gemini.google.com", "https://gemini.google.com/app"
	case "youtube.com", "www.youtube.com":
		return "www.youtube.com", "https://www.youtube.com/premium"
	default:
		return domain, "https://" + domain + "/"
	}
}

func probeDomainViaWARP(ctx context.Context, domain, warpSocks, family string) probeResult {
	dialDomain, targetURL := probeURLForDomain(domain)
	transport := &http.Transport{
		Proxy:               nil,
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 6 * time.Second,
		DialContext: func(dialContext context.Context, _, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			if host == "" {
				host = dialDomain
			}
			return dialViaWARP(dialContext, warpSocks, host, port, family)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 12 * time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return probeResult{}
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	response, err := client.Do(request)
	if err != nil {
		return probeResult{}
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, probeBodyLimit))
	finalURL := ""
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}
	return inspectProbeResponse(dialDomain, response.StatusCode, finalURL, body)
}

func probeResultHealthy(r probeResult) bool {
	if !r.Reachable || r.Blocked {
		return false
	}
	if r.Region == "CHN" || r.Region == "CN" {
		return false
	}
	return r.Code >= 200 && r.Code < 400
}

// evaluateWarpChoice selects between WARP IPv6 (preferred when clean) and WARP IPv4.
func evaluateWarpChoice(ctx context.Context, service servicePolicy, warpSocks string) string {
	targets := representativeProbeTargets(service)
	if len(targets) == 0 {
		return egressWarp6
	}
	domain := targets[0]
	probeCtx, cancel := context.WithTimeout(ctx, 14*time.Second)
	defer cancel()
	var w6, w4 probeResult
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); w6 = probeDomainViaWARP(probeCtx, domain, warpSocks, "ipv6") }()
	go func() { defer wg.Done(); w4 = probeDomainViaWARP(probeCtx, domain, warpSocks, "ipv4") }()
	wg.Wait()
	log.Printf("probe warp service=%q domain=%s w6={ok:%t code:%d blocked:%t region:%s} w4={ok:%t code:%d blocked:%t region:%s}",
		service.Name, domain,
		w6.Reachable, w6.Code, w6.Blocked, w6.Region,
		w4.Reachable, w4.Code, w4.Blocked, w4.Region)
	if probeResultHealthy(w6) {
		return egressWarp6
	}
	if probeResultHealthy(w4) {
		return egressWarp4
	}
	if w6.Reachable && !w6.Blocked {
		return egressWarp6
	}
	if w4.Reachable && !w4.Blocked {
		return egressWarp4
	}
	return egressWarp6
}

// evaluateAutoService runs 4-tier adaptive selection: native IPv4 -> native IPv6 -> WARP IPv6 -> WARP IPv4.
func evaluateAutoService(ctx context.Context, service servicePolicy, warpSocks string, warpUp bool) string {
	targets := representativeProbeTargets(service)
	if len(targets) == 0 {
		return egressNative4
	}
	domain := targets[0]
	probeCtx, cancel := context.WithTimeout(ctx, 14*time.Second)
	var v4, v6 probeResult
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); v4 = probeDomain(probeCtx, domain, false) }()
	go func() { defer wg.Done(); v6 = probeDomain(probeCtx, domain, true) }()
	wg.Wait()
	cancel()
	log.Printf("probe auto service=%q domain=%s v4={ok:%t code:%d blocked:%t region:%s} v6={ok:%t code:%d blocked:%t region:%s}",
		service.Name, domain,
		v4.Reachable, v4.Code, v4.Blocked, v4.Region,
		v6.Reachable, v6.Code, v6.Blocked, v6.Region)
	if probeResultHealthy(v4) {
		return egressNative4
	}
	if preferIPv6Result(servicePolicy{IPv6Candidate: true}, v4, v6) {
		return egressNative6
	}
	if warpUp {
		choice := evaluateWarpChoice(ctx, service, warpSocks)
		return choice
	}
	return egressNative4
}

func probeDomain(ctx context.Context, domain string, ipv6 bool) probeResult {
	dialDomain, targetURL := probeURLForDomain(domain)
	transport := &http.Transport{
		Proxy:               nil,
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 6 * time.Second,
		DialContext: func(dialContext context.Context, _, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			if host == "" {
				host = dialDomain
			}
			return dialHost(dialContext, host, port, ipv6, true, false, "")
		},
	}
	client := &http.Client{Transport: transport, Timeout: 12 * time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return probeResult{}
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	response, err := client.Do(request)
	if err != nil {
		return probeResult{}
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, probeBodyLimit))
	finalURL := ""
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}
	return inspectProbeResponse(dialDomain, response.StatusCode, finalURL, body)
}

func inspectProbeResponse(domain string, statusCode int, finalURL string, body []byte) probeResult {
	result := probeResult{
		Reachable: true,
		Code:      statusCode,
		Blocked:   blockedPagePattern.Match(body) || blockedPagePattern.MatchString(finalURL),
	}
	if match := geminiRegionPattern.FindSubmatch(body); len(match) == 2 {
		result.Region = string(match[1])
	}
	if strings.Contains(domain, "gemini.google.com") {
		if geminiFalsePattern.Match(body) || strings.Contains(finalURL, "/faq") {
			result.Blocked = true
		} else if !geminiTruePattern.Match(body) {
			if _, regionBlocked := geminiBlockedRegions[result.Region]; regionBlocked {
				result.Blocked = true
			}
		}
	}
	if strings.Contains(domain, "youtube.com") && youtubeCNPattern.Match(body) {
		result.Region = "CHN"
		result.Blocked = true
	}
	if strings.Contains(domain, "play.google.com") {
		if match := playCountryPattern.FindSubmatch(body); len(match) == 2 {
			country := strings.TrimSpace(string(match[1]))
			if strings.EqualFold(country, "China") {
				result.Region = "CHN"
				result.Blocked = true
			} else if country != "" && result.Region == "" {
				result.Region = country
			}
		}
	}
	return result
}

type proxyServer struct {
	policy      *familyPolicy
	dialTimeout time.Duration
	mode        string // "proxy" or "client"
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
	log.Printf("listening on %s (mode=%s)", address, server.mode)
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

func isUsableDirectIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast()
}

func (server *proxyServer) handle(client net.Conn, port string, hostname func([]byte) (string, error)) {
	defer client.Close()

	var origAddr string
	var origIP net.IP
	if server.mode == "client" {
		origAddr, origIP, _ = originalDestination(client)
	}

	_ = client.SetReadDeadline(time.Now().Add(8 * time.Second))
	initial, readErr := readInitial(client, port)
	_ = client.SetReadDeadline(time.Time{})

	var host string
	var parseErr error
	if readErr == nil {
		host, parseErr = hostname(initial)
	}

	// If client mode cannot extract SNI/Host (e.g. IP-direct TLS, ECH, MTProto, custom protocol),
	// check CIDR routing first, then fall back cleanly to SO_ORIGINAL_DST without dropping traffic!
	if readErr != nil || parseErr != nil {
		if server.mode == "client" {
			if routeIP := server.policy.routeDestination("", origIP); routeIP != "" {
				server.forwardToAddress(client, net.JoinHostPort(routeIP, port), initial)
				return
			}
			if origAddr != "" && isUsableDirectIP(origIP) && !server.policy.isKnownProxyIP(origIP) {
				server.forwardToAddress(client, origAddr, initial)
				return
			}
		}
		return
	}

	if server.mode == "client" {
		routeIP := server.policy.routeDestination(host, origIP)
		if routeIP != "" {
			server.forwardToAddress(client, net.JoinHostPort(routeIP, port), initial)
			return
		}
		// Unmanaged domain on client: pass straight through to the kernel's original destination IP
		// (unless that IP was a stale Prism proxy IP, in which case resolve fresh via clean DNS).
		if origAddr != "" && isUsableDirectIP(origIP) && !server.policy.isKnownProxyIP(origIP) {
			server.forwardToAddress(client, origAddr, initial)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), server.dialTimeout)
		upstream, err := dialHost(ctx, host, port, false, false, true, "")
		cancel()
		if err != nil {
			log.Printf("client direct dial %s:%s: %v", host, port, err)
			return
		}
		defer upstream.Close()
		if len(initial) > 0 {
			if _, err := upstream.Write(initial); err != nil {
				return
			}
		}
		relay(client, upstream)
		return
	}

	// Proxy mode: per-service WARP or smart native IPv4/IPv6 egress selection
	egressMode := server.policy.egressModeFor(host)
	var upstream net.Conn
	var err error
	if (egressMode == egressWarp6 || egressMode == egressWarp4) && server.policy.warpSocks != "" && !server.policy.shouldBypassWARP() {
		family := "ipv6"
		if egressMode == egressWarp4 {
			family = "ipv4"
		}
		warpTimeout := 4 * time.Second
		if server.dialTimeout < warpTimeout {
			warpTimeout = server.dialTimeout
		}
		warpCtx, warpCancel := context.WithTimeout(context.Background(), warpTimeout)
		upstream, err = dialViaWARP(warpCtx, server.policy.warpSocks, host, port, family)
		warpCancel()
		if err != nil {
			server.policy.recordWARPResult(false)
			log.Printf("warp dial %s:%s (%s) failed, falling back to native with fresh context: %v", host, port, egressMode, err)
		} else {
			server.policy.recordWARPResult(true)
		}
	}
	if upstream == nil {
		nativeCtx, nativeCancel := context.WithTimeout(context.Background(), server.dialTimeout)
		prefer6 := egressMode == egressNative6 || server.policy.preferIPv6(host)
		allow6 := prefer6 || server.policy.allowIPv6Fallback(host)
		upstream, err = dialHost(nativeCtx, host, port, prefer6, !allow6, allow6, "")
		nativeCancel()
	}
	if err != nil {
		log.Printf("dial %s:%s (mode=%s): %v", host, port, egressMode, err)
		return
	}
	defer upstream.Close()
	if _, err := upstream.Write(initial); err != nil {
		return
	}
	relay(client, upstream)
}

func (server *proxyServer) forwardToAddress(client net.Conn, targetAddr string, initial []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), server.dialTimeout)
	dialer := &net.Dialer{Timeout: 6 * time.Second, KeepAlive: 30 * time.Second}
	upstream, err := dialer.DialContext(ctx, "tcp", targetAddr)
	cancel()
	if err != nil {
		log.Printf("forward %s: %v", targetAddr, err)
		return
	}
	defer upstream.Close()
	if len(initial) > 0 {
		if _, err := upstream.Write(initial); err != nil {
			return
		}
	}
	relay(client, upstream)
}

// readInitial reads the TLS ClientHello or HTTP request header while preserving
// every byte already read from connection so non-HTTP/non-TLS streams can still be relayed.
func readInitial(connection net.Conn, port string) ([]byte, error) {
	if port == "443" {
		header := make([]byte, 5)
		n, err := io.ReadFull(connection, header)
		if err != nil {
			return header[:n], err
		}
		length := int(binary.BigEndian.Uint16(header[3:5]))
		if header[0] != 22 || length <= 0 || length > maxInitialBytes-5 {
			return header, errors.New("invalid TLS client hello")
		}
		payload := make([]byte, length)
		pn, err := io.ReadFull(connection, payload)
		if err != nil {
			return append(header, payload[:pn]...), err
		}
		return append(header, payload...), nil
	}
	buffer := make([]byte, 0, 4096)
	chunk := make([]byte, 1)
	for len(buffer) < maxInitialBytes {
		if _, err := io.ReadFull(connection, chunk); err != nil {
			return buffer, err
		}
		buffer = append(buffer, chunk[0])
		if len(buffer) >= 4 && bytes.Equal(buffer[len(buffer)-4:], []byte("\r\n\r\n")) {
			return buffer, nil
		}
	}
	return buffer, errors.New("HTTP headers exceed limit")
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

func lookupPublicIPs(ctx context.Context, host string) ([]net.IPAddr, error) {
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err == nil && len(addresses) > 0 {
		for _, a := range addresses {
			if isUsableDirectIP(a.IP) {
				return addresses, nil
			}
		}
	}
	for _, dnsServer := range fallbackDNSServers {
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(dialCtx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 3 * time.Second}).DialContext(dialCtx, "udp", dnsServer)
			},
		}
		if fallbackAddrs, fallbackErr := resolver.LookupIPAddr(ctx, host); fallbackErr == nil && len(fallbackAddrs) > 0 {
			return fallbackAddrs, nil
		}
	}
	return addresses, err
}

func dialHost(ctx context.Context, host, port string, prefer6, singleFamily, allowOtherFamilyFallback bool, forcedIP string) (net.Conn, error) {
	if parsed := net.ParseIP(strings.TrimSpace(forcedIP)); parsed != nil && parsed.To4() != nil {
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		return dialer.DialContext(ctx, "tcp", net.JoinHostPort(parsed.To4().String(), port))
	}
	addresses, err := lookupPublicIPs(ctx, host)
	if err != nil && len(addresses) == 0 {
		return nil, err
	}
	filtered := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		ip := address.IP
		if !isUsableDirectIP(ip) {
			continue
		}
		isIPv6 := ip.To4() == nil
		if singleFamily && isIPv6 != prefer6 {
			continue
		}
		if !allowOtherFamilyFallback && !prefer6 && isIPv6 {
			continue
		}
		filtered = append(filtered, ip)
	}
	// If host is IPv6-only (no A record at all) and we aren't running a strict single-family probe, allow IPv6.
	if len(filtered) == 0 && !singleFamily {
		for _, address := range addresses {
			if isUsableDirectIP(address.IP) {
				filtered = append(filtered, address.IP)
			}
		}
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

func validateListenAddress(address string, allowWildcard bool) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid listen address %s: %w", address, err)
	}
	if !allowWildcard && host != "127.0.0.1" && host != "::1" {
		return fmt.Errorf("listen address must use loopback: %s", address)
	}
	if allowWildcard && host != "" && host != "0.0.0.0" && host != "::" && host != "127.0.0.1" && host != "::1" {
		if net.ParseIP(host) == nil {
			return fmt.Errorf("invalid listen host: %s", address)
		}
	}
	if parsedPort, parseErr := strconv.Atoi(port); parseErr != nil || parsedPort < 1 || parsedPort > 65535 {
		return fmt.Errorf("invalid listen port: %s", address)
	}
	return nil
}

func main() {
	var tlsListen, httpListen, policyFile, mode, warpSocks string
	var dialTimeout time.Duration
	flag.StringVar(&tlsListen, "tls-listen", "127.0.0.1:19443", "TLS SNI listen address")
	flag.StringVar(&httpListen, "http-listen", "127.0.0.1:19080", "HTTP Host listen address")
	flag.StringVar(&policyFile, "policy", "/etc/prismdns/egress-policy.json", "service egress policy file")
	flag.StringVar(&mode, "mode", "auto", "operating mode: auto, proxy, or client")
	flag.StringVar(&warpSocks, "warp-socks", "127.0.0.1:40000", "local WARP/SOCKS5 proxy address for per-service WARP routing")
	flag.DurationVar(&dialTimeout, "dial-timeout", 20*time.Second, "upstream dial timeout")
	flag.Parse()

	if dialTimeout < time.Second || dialTimeout > time.Minute {
		log.Fatal("dial-timeout must be between 1s and 1m")
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "auto" {
		if strings.Contains(strings.ToLower(policyFile), "client") || strings.HasSuffix(tlsListen, ":19444") {
			mode = "client"
		} else {
			mode = "proxy"
		}
	}
	if mode != "proxy" && mode != "client" {
		log.Fatalf("invalid mode %q: must be auto, proxy, or client", mode)
	}
	allowWildcard := mode == "proxy"
	for _, address := range []string{tlsListen, httpListen} {
		if err := validateListenAddress(address, allowWildcard); err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	policy := newFamilyPolicy(policyFile)
	policy.warpSocks = strings.TrimSpace(warpSocks)
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

	server := &proxyServer{policy: policy, dialTimeout: dialTimeout, mode: mode}
	errorsChannel := make(chan error, 2)
	go func() { errorsChannel <- server.serve(ctx, tlsListen, "443", tlsHostname) }()
	go func() { errorsChannel <- server.serve(ctx, httpListen, "80", httpHostname) }()
	if err := <-errorsChannel; err != nil {
		log.Fatal(err)
	}
}
