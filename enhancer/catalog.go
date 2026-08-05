package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	categoryPattern          = regexp.MustCompile(`^#\s*-+\s*>\s*(.+?)\s*$`)
	servicePattern           = regexp.MustCompile(`^#\s*>\s*(.+?)\s*$`)
	domainPattern            = regexp.MustCompile(`^nameserver\s+/([^/]+)/`)
	geminiApplicationDomains = []string{
		"gemini.google.com",
		"accountcapabilities-pa.googleapis.com",
		"accounts.google.com",
		"aisandbox-pa.googleapis.com",
		"alkaliminer-pa.googleapis.com",
		"android-context-data.googleapis.com",
		"android.googleapis.com",
		"app-measurement.com",
		"clients3.google.com",
		"firebaseinstallations.googleapis.com",
		"firebaseremoteconfig.googleapis.com",
		"gemini.gstatic.com",
		"growth-pa.googleapis.com",
		"lh3.googleusercontent.com",
		"notifications-pa.googleapis.com",
		"ogads-pa.clients6.google.com",
		"ogs.google.com",
		"oauth2.googleapis.com",
		"oauthaccountmanager.googleapis.com",
		"people-pa.googleapis.com",
		"play.google.com",
		"play.googleapis.com",
		"proactivebackend-pa.googleapis.com",
		"robinfrontend-pa.googleapis.com",
		"signaler-pa.clients6.google.com",
		"signaler-pa.googleapis.com",
		"ssl.gstatic.com",
		"subscriptionsfirstparty-pa.googleapis.com",
		"voilatile-pa.googleapis.com",
		"waa-pa.clients6.google.com",
		"www.googleapis.com",
		"www.google.com",
		"www.google.com.hk",
		"www.gstatic.com",
	}
	grokDomains = []string{
		"ads-twitter.com",
		"cms-twdigitalassets.com",
		"periscope.tv",
		"pscp.tv",
		"t.co",
		"tellapart.com",
		"tweetdeck.com",
		"twimg.co",
		"twimg.com",
		"twimg.org",
		"twitpic.com",
		"twitter.biz",
		"twitter.com",
		"twitter.jp",
		"twittercommunity.com",
		"twitterflightschool.com",
		"twitterinc.com",
		"twitteroauth.com",
		"twitterstat.us",
		"twtrdns.net",
		"twttr.com",
		"twttr.net",
		"twvid.com",
		"vine.co",
		"x.com",
		"grok.com",
		"www.grok.com",
		"x.ai",
		"api.x.ai",
		"grok.x.com",
		"accounts.x.com",
	}
	grokDomainKeywords = []string{"twitter"}
	grokCIDRs = []string{
		"192.133.76.0/22",
		"199.59.148.0/22",
		"199.96.56.0/21",
		"202.160.128.0/22",
		"209.237.192.0/19",
		"69.195.160.0/19",
	}
	serviceDomainSupplements = map[string][]string{
		"FR:France.tv": {
			"france.tv",
		},
		"Gemini": geminiApplicationDomains,
		"Grok":   grokDomains,
		"iQIYI": {
			"iq.com",
			"iqiyi.com",
		},
		"HOY TV": {
			"hoy.tv",
		},
		"J:com On Demand": {
			"jcom.co.jp",
			"myjcom.jp",
		},
		"Karaoke@DAM": {
			"clubdam.com",
		},
		"LiTV": {
			"litvfreepc.akamaized.net",
		},
		"NZ:Neon TV": {
			"neontv.co.nz",
		},
		"NZ:ThreeNow": {
			"threenow.co.nz",
		},
		"YouTube": {
			"accounts.youtube.com",
			"ggpht.com",
			"googlevideo.com",
			"gvt1.com",
			"gvt2.com",
			"i.ytimg.com",
			"music.youtube.com",
			"youtu.be",
			"youtube-nocookie.com",
			"ytimg.com",
			"s.youtube.com",
			"youtube.googleapis.com",
		},
		"Videomarket": {
			"videomarket.jp",
		},
		"Viu.TV": {
			"viu.com",
			"viu.tv",
		},
		"VN:Galaxy Play": {
			"galaxyplay.vn",
		},
		"VN:K+": {
			"k-plus.tv",
			"kplus.vn",
		},
		"Wavve": {
			"wavve.com",
		},
		"AU:10 play": {
			"global.ssl.fastly.net",
		},
	}
	serviceDomainKeywordSupplements = map[string][]string{
		"Grok": grokDomainKeywords,
	}
	serviceCIDRSupplements = map[string][]string{
		"Grok": grokCIDRs,
	}
	retiredServices = map[string]struct{}{
		"Crackle":  {},
		"FR:Salto": {},
		"GYAO!":    {},
	}
)

type Service struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	OriginalCategory string   `json:"original_category,omitempty"`
	Domains          []string `json:"domains"`
	DomainKeywords   []string `json:"domain_keywords,omitempty"`
	CIDRs            []string `json:"cidrs,omitempty"`
	Custom           bool     `json:"custom"`
	DomainOverride   bool     `json:"domain_override,omitempty"`
	Aliases          []string `json:"aliases,omitempty"`
}

func ParseSmartDNS(reader io.Reader) ([]Service, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)

	category := "其他服务"
	serviceName := "未分类"
	services := make(map[string]*Service)
	order := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := categoryPattern.FindStringSubmatch(line); len(matches) == 2 {
			category = cleanLabel(matches[1], "其他服务")
			serviceName = "未分类"
			continue
		}
		if matches := servicePattern.FindStringSubmatch(line); len(matches) == 2 {
			serviceName = cleanLabel(matches[1], "未分类")
			continue
		}

		matches := domainPattern.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}
		domain := normalizeDomain(matches[1])
		if domain == "" {
			continue
		}

		key := category + "\x00" + serviceName
		service, ok := services[key]
		if !ok {
			service = &Service{
				ID:       stableServiceID(category, serviceName),
				Name:     canonicalServiceName(serviceName, category),
				Category: canonicalCategory(category, serviceName),
				Domains:  make([]string, 0),
			}
			services[key] = service
			order = append(order, key)
		}
		if !contains(service.Domains, domain) {
			service.Domains = append(service.Domains, domain)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan smartdns list: %w", err)
	}

	result := make([]Service, 0, len(order))
	for _, key := range order {
		service := services[key]
		if _, retired := retiredServices[service.Name]; retired {
			continue
		}
		service.Domains = normalizeDomains(append(service.Domains, serviceDomainSupplements[service.Name]...))
		if service.Name == "Grok" {
			service.Domains = excludeDomain(service.Domains, "twitter")
		}
		service.DomainKeywords = normalizeDomainKeywords(append(service.DomainKeywords, serviceDomainKeywordSupplements[service.Name]...))
		service.CIDRs = normalizeCIDRs(append(service.CIDRs, serviceCIDRSupplements[service.Name]...))
		result = append(result, *service)
	}
	return result, nil
}

func ensureBuiltInServices(services []Service) []Service {
	for _, service := range []Service{
		{
			ID:       stableServiceID("AI Platform", "Grok"),
			Name:     "Grok",
			Category: "AI Platform",
			Domains:  normalizeDomains(grokDomains),
			DomainKeywords: normalizeDomainKeywords(grokDomainKeywords),
			CIDRs: normalizeCIDRs(grokCIDRs),
		},
	} {
		present := false
		for index, existing := range services {
			if serviceIdentityKey(existing.Name) == serviceIdentityKey(service.Name) {
				services[index].Domains = normalizeDomains(append(existing.Domains, service.Domains...))
				if services[index].Name == "Grok" {
					services[index].Domains = excludeDomain(services[index].Domains, "twitter")
				}
				services[index].DomainKeywords = normalizeDomainKeywords(append(existing.DomainKeywords, service.DomainKeywords...))
				services[index].CIDRs = normalizeCIDRs(append(existing.CIDRs, service.CIDRs...))
				present = true
				break
			}
		}
		if !present {
			services = append(services, service)
		}
	}
	return services
}

func canonicalServiceName(name, category string) string {
	aliases := map[string]string{
		"claude 2":                              "Claude",
		"google aistudio":                       "Google AI Studio",
		"google gemini":                         "Gemini",
		"microsoft copilot for image generates": "Microsoft Copilot Image Creator",
		"openai":                                "ChatGPT / OpenAI",
		"bilibili":                              "Bilibili",
		"netease cloud music":                   "NetEase Cloud Music",
		"youtube":                               "YouTube",
		"tiktok":                                "TikTok",
		"iqiyi":                                 "iQIYI",
		"eu:skyshowtime":                        "EU:SkyShowtime",
		"eurosport":                             "Eurosport",
		"viaplay":                               "Viaplay",
		"gb:sky go /<replace with groupname>skygonz/<replace with groupname>": "GB:Sky Go",
		"it rai play":       "IT:RaiPlay",
		"nl:videoland":      "NL:Videoland",
		"tr:digiturkplay":   "TR:Digiturk Play",
		"huluusa":           "Hulu (US)",
		"exhantai/e-hentai": "ExHentai / E-Hentai",
	}
	if canonical, ok := aliases[strings.ToLower(strings.TrimSpace(name))]; ok {
		return canonical
	}
	if name == "未分类" && strings.EqualFold(category, "Indian Media") {
		return "Indian Media Bundle"
	}
	return name
}

func canonicalCategory(category, serviceName string) string {
	if strings.EqualFold(strings.TrimSpace(serviceName), "YouTube") {
		return "Global Platform"
	}
	value := strings.TrimSpace(category)
	switch strings.ToLower(value) {
	case "global plaform":
		return "Global Platform"
	case "southeastasia media":
		return "Southeast Asia Media"
	case "? media":
		if strings.EqualFold(strings.TrimSpace(serviceName), "Setanta Sports") {
			return "Sports Media"
		}
	}
	return value
}

func normalizeDomain(value string) string {
	return routingDomain(normalizeDomainPattern(value))
}

func normalizeDomainPattern(value string) string {
	domain := strings.ToLower(strings.TrimSpace(value))
	wildcard := strings.HasPrefix(domain, "*.")
	if wildcard {
		domain = strings.TrimPrefix(domain, "*.")
	}
	domain = strings.TrimPrefix(domain, ".")
	domain = strings.TrimSuffix(domain, ".")
	if !validDomainName(domain) {
		return ""
	}
	if wildcard {
		return "*." + domain
	}
	return domain
}

func normalizeDomains(domains []string) []string {
	return collectDomains(domains, normalizeDomain)
}

func excludeDomain(domains []string, excluded string) []string {
	excluded = routingDomain(excluded)
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		if routingDomain(domain) != excluded {
			result = append(result, domain)
		}
	}
	return result
}

func normalizeCustomDomains(domains []string) ([]string, error) {
	values := splitDomainValues(domains)
	if len(values) > 1000 {
		return nil, fmt.Errorf("域名数量过多，最多支持 1000 项")
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		domain := normalizeDomainPattern(value)
		if domain == "" {
			return nil, fmt.Errorf("域名或泛域名格式无效 %q；请使用 example.com 或 *.example.com", strings.TrimSpace(value))
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	sort.Strings(result)
	return result, nil
}

func normalizeDomainKeywords(keywords []string) []string {
	return collectRuleValues(keywords, normalizeDomainKeyword)
}

func normalizeCustomDomainKeywords(keywords []string) ([]string, error) {
	values := splitDomainValues(keywords)
	if len(values) > 1000 {
		return nil, fmt.Errorf("domain keyword count exceeds 1000")
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		keyword := normalizeDomainKeyword(value)
		if keyword == "" {
			return nil, fmt.Errorf("invalid domain keyword %q", strings.TrimSpace(value))
		}
		result = append(result, keyword)
	}
	return uniqueSortedStrings(result), nil
}

func normalizeDomainKeyword(value string) string {
	keyword := strings.ToLower(strings.TrimSpace(value))
	if keyword == "" || len(keyword) > 128 {
		return ""
	}
	for index := 0; index < len(keyword); index++ {
		character := keyword[index]
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '-' && character != '_' && character != '.' {
			return ""
		}
	}
	return keyword
}

func normalizeCIDRs(cidrs []string) []string {
	return collectRuleValues(cidrs, normalizeCIDR)
}

func normalizeCustomCIDRs(cidrs []string) ([]string, error) {
	values := splitDomainValues(cidrs)
	if len(values) > 1000 {
		return nil, fmt.Errorf("CIDR count exceeds 1000")
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		cidr := normalizeCIDR(value)
		if cidr == "" {
			return nil, fmt.Errorf("invalid CIDR %q", strings.TrimSpace(value))
		}
		result = append(result, cidr)
	}
	return uniqueSortedStrings(result), nil
}

func normalizeCIDR(value string) string {
	_, network, err := net.ParseCIDR(strings.TrimSpace(value))
	if err != nil || network == nil {
		return ""
	}
	return network.String()
}

func collectRuleValues(values []string, normalize func(string) string) []string {
	result := make([]string, 0, len(values))
	for _, value := range splitDomainValues(values) {
		if normalized := normalize(value); normalized != "" {
			result = append(result, normalized)
		}
	}
	return uniqueSortedStrings(result)
}

func uniqueSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func routingDomains(domains []string) []string {
	return collectDomains(domains, routingDomain)
}

func routingDomain(value string) string {
	domain := normalizeDomainPattern(value)
	return strings.TrimPrefix(domain, "*.")
}

func collectDomains(domains []string, normalize func(string) string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(domains))
	for _, value := range splitDomainValues(domains) {
		domain := normalize(value)
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	sort.Strings(result)
	return result
}

func splitDomainValues(domains []string) []string {
	result := make([]string, 0, len(domains))
	for _, line := range domains {
		for _, value := range strings.FieldsFunc(line, func(r rune) bool {
			return r == '\n' || r == '\r' || r == ',' || r == ';'
		}) {
			if strings.TrimSpace(value) != "" {
				result = append(result, value)
			}
		}
	}
	return result
}

func validDomainName(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}
	for _, label := range strings.Split(domain, ".") {
		if len(label) == 0 || len(label) > 63 || !asciiAlphaNumeric(label[0]) || !asciiAlphaNumeric(label[len(label)-1]) {
			return false
		}
		for index := 1; index < len(label)-1; index++ {
			if !asciiAlphaNumeric(label[index]) && label[index] != '-' {
				return false
			}
		}
	}
	return true
}

func asciiAlphaNumeric(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9'
}

func stableServiceID(category, name string) string {
	base := slug(name)
	if base == "" {
		base = "service"
	}
	sum := sha256.Sum256([]byte(category + "\x00" + name))
	return base + "-" + hex.EncodeToString(sum[:4])
}

func serviceIdentityKey(name string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func customServiceID(name string) string {
	sum := sha256.Sum256([]byte(name))
	base := slug(name)
	if base == "" {
		base = "service"
	}
	return "custom-" + base + "-" + hex.EncodeToString(sum[:4])
}

func slug(value string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) && r <= unicode.MaxASCII, unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case builder.Len() > 0 && !lastDash:
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func cleanLabel(value, fallback string) string {
	value = strings.TrimSpace(strings.TrimSuffix(value, ":"))
	if value == "" {
		return fallback
	}
	return value
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
