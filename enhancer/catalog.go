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
		"gemini.app.google",
		"ai.google",
		"labs.google",
		"deepmind.google",
		"notebooklm.google",
		"notebooklm.google.com",
		"jules.google",
		"jules.google.com",
		"idx.google.com",
		"accountcapabilities-pa.googleapis.com",
		"accounts.google.com",
		"ai.googleusercontent.com",
		"aisandbox-pa.googleapis.com",
		"alkaliminer-pa.googleapis.com",
		"antigravity.google.com",
		"android-context-data.googleapis.com",
		"android.googleapis.com",
		"app-measurement.com",
		"aratea-pa.googleapis.com",
		"autopush-cloudcode-pa.googleapis.com",
		"bard.google.com",
		"clients1.google.com",
		"clients2.google.com",
		"clients3.google.com",
		"clients4.google.com",
		"cloudaicompanion.googleapis.com",
		"cloudcode-pa.googleapis.com",
		"content.googleapis.com",
		"content-push.googleapis.com",
		"daily-cloudcode-pa.googleapis.com",
		"firebaseinstallations.googleapis.com",
		"firebaseremoteconfig.googleapis.com",
		"footprints-pa.googleapis.com",
		"geller-pa.googleapis.com",
		"gemini.gstatic.com",
		"geminiweb-pa.clients.google.com",
		"geminiweb-pa.clients6.google.com",
		"growth-pa.googleapis.com",
		"lh3.googleusercontent.com",
		"notifications-pa.googleapis.com",
		"ogads-pa.clients6.google.com",
		"ogs.google.com",
		"oauth2.googleapis.com",
		"oauthaccountmanager.googleapis.com",
		"optimizationguide-pa.googleapis.com",
		"people-pa.googleapis.com",
		"play.google.com",
		"play.googleapis.com",
		"proactivebackend-pa.googleapis.com",
		"push.clients6.google.com",
		"robinfrontend-pa.googleapis.com",
		"scone-pa.clients6.google.com",
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
	grokCIDRs          = []string{
		"192.133.76.0/22",
		"199.59.148.0/22",
		"199.96.56.0/21",
		"202.160.128.0/22",
		"209.237.192.0/19",
		"69.195.160.0/19",
	}
	serviceDomainSupplements = map[string][]string{
		"ChatGPT / OpenAI": {
			"chatgpt.com",
			"openai.com",
			"ab.chatgpt.com",
			"api.openai.com",
			"auth.openai.com",
			"auth0.openai.com",
			"browser-intake-datadoghq.com",
			"cdn.oaistatic.com",
			"oaistatic.com",
			"chat.openai.com",
			"challenges.cloudflare.com",
			"events.statsig.com",
			"featuregates.org",
			"files.oaiusercontent.com",
			"oaiusercontent.com",
			"id.openai.com",
			"platform.openai.com",
			"prodregistryv2.org",
			"static.cloudflareinsights.com",
			"statsigapi.net",
			"sora.com",
			"ws.chatgpt.com",
		},
		"FR:France.tv": {
			"france.tv",
		},
		"Gemini": geminiApplicationDomains,
		"Google AI Studio": {
			"aistudio.google.com",
			"alkalicore-pa.clients6.google.com",
			"alkalimakersuite-pa.clients6.google.com",
			"generativelanguage.googleapis.com",
			"makersuite.google.com",
			"waa-pa.clients6.google.com",
			"webchannel-alkalimakersuite-pa.clients6.google.com",
		},
		"Grok": grokDomains,
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
		// YouTube：页面/接口域名用于地区判定；视频流与图片 CDN（googlevideo.com、
		// ggpht.com、ytimg.com、gvt1.com、gvt2.com）也必须经解锁机，否则在对 SNI 做
		// 过滤的机房网络里（已实测）视频直连会被丢弃，表现为“页面能开、
		// 视频一直转圈”。此类网络必须使用加密传输模式（prism_transport.sh --mode
		// encrypted），明文直连模式下这些 SNI 同样会被丢弃。
		"YouTube": {
			"accounts.youtube.com",
			"ggpht.com",
			"googlevideo.com",
			"gvt1.com",
			"gvt2.com",
			"i.ytimg.com",
			"jnn-pa.googleapis.com",
			"music.youtube.com",
			"s.youtube.com",
			"youtu.be",
			"youtube-nocookie.com",
			"youtube.com",
			"youtube.googleapis.com",
			"youtubeembeddedplayer.googleapis.com",
			"youtubeeducation.com",
			"youtubekids.com",
			"yt.be",
			"ytimg.com",
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
	// Google Play 商店域名。刻意不包含 play.googleapis.com / android.googleapis.com：
	// 它们已属于 Gemini，重叠会让两个服务被强制绑定到同一节点；
	// 也不包含 dl.google.com、play-lh、play-games（下载与图片 CDN，应直连）。
	// play.google.com 在 SNI 过滤网络中需要加密传输模式才能经解锁机转发。
	googlePlayStoreDomains = []string{
		"android.clients.google.com",
		"play-fe.googleapis.com",
		"play.google.com",
		"playatoms-pa.googleapis.com",
	}
	xTwitterDomains = []string{
		"x.com",
		"twitter.com",
		"twimg.com",
		"twimg.co",
		"twimg.org",
		"t.co",
		"twttr.com",
		"twttr.net",
		"twtrdns.net",
		"twvid.com",
		"tweetdeck.com",
		"twitter.jp",
		"twitter.biz",
		"twitteroauth.com",
		"twitterstat.us",
		"twitterinc.com",
		"twittercommunity.com",
		"ads-twitter.com",
		"cms-twdigitalassets.com",
		"pscp.tv",
		"periscope.tv",
		"vine.co",
		"twitpic.com",
		"tellapart.com",
	}
	instagramDomains = []string{
		"instagram.com",
		"cdninstagram.com",
		"ig.me",
		"instagr.am",
		"igcdn.com",
		"igsonar.com",
		"i.instagram.com",
		"graph.instagram.com",
		"help.instagram.com",
		"scontent.cdninstagram.com",
		"instagramstatic-a.akamaihd.net",
	}
	facebookDomains = []string{
		"facebook.com",
		"fb.com",
		"fb.me",
		"fb.gg",
		"fb.watch",
		"fbcdn.net",
		"fbsbx.com",
		"facebook.net",
		"messenger.com",
		"m.me",
		"accountkit.com",
		"atdmt.com",
		"fbinfra.net",
		"fbpigeon.com",
		"fbthirdpartypixel.com",
		"meta.com",
		"workplace.com",
		"oculus.com",
		"oculuscdn.com",
	}
	threadsDomains = []string{
		"threads.net",
		"threads.com",
	}
	whatsappDomains = []string{
		"whatsapp.com",
		"whatsapp.net",
		"wa.me",
		"whatsapp-plus.info",
		"whatsapp-plus.me",
		"whatsapp-plus.net",
		"whatsappbrand.com",
	}
	telegramDomains = []string{
		"telegram.org",
		"t.me",
		"telegram.me",
		"telegram.dog",
		"telegram.space",
		"telesco.pe",
		"telegra.ph",
		"tdesktop.com",
		"tg.dev",
		"tx.me",
		"fragment.com",
		"contest.com",
		"graph.org",
		"legit.tg",
		"nicegram.app",
	}
	telegramCIDRs = []string{
		"91.108.4.0/22",
		"91.108.8.0/22",
		"91.108.12.0/22",
		"91.108.16.0/22",
		"91.108.20.0/22",
		"91.108.56.0/22",
		"149.154.160.0/20",
		"185.76.151.0/24",
	}
	discordDomains = []string{
		"discord.com",
		"discord.gg",
		"discord.media",
		"discord.co",
		"discord.dev",
		"discord.new",
		"discord.gift",
		"discord.gifts",
		"discord.design",
		"discord.status",
		"discordapp.com",
		"discordapp.net",
		"discordcdn.com",
		"discordstatus.com",
		"dis.gd",
		"watchanimeattheoffice.com",
	}
	redditDomains = []string{
		"reddit.com",
		"redd.it",
		"redditmedia.com",
		"redditstatic.com",
		"redditinc.com",
		"redditstatus.com",
		"redditmail.com",
		"redditblog.com",
		"redditspace.com",
	}
	pinterestDomains = []string{
		"pinterest.com",
		"pinimg.com",
		"pin.it",
		"pinterest.co.uk",
		"pinterest.jp",
		"pinterest.de",
		"pinterest.fr",
		"pinterest.ca",
	}
	lineDomains = []string{
		"line.me",
		"line-apps.com",
		"line-cdn.net",
		"line-scdn.net",
		"linecorp.com",
		"lin.ee",
		"nhncorp.jp",
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
			ID:       stableServiceID("Global Platform", "Google Play Store"),
			Name:     "Google Play Store",
			Category: "Global Platform",
			Domains:  normalizeDomains(googlePlayStoreDomains),
		},
		{
			ID:             stableServiceID("AI Platform", "Grok"),
			Name:           "Grok",
			Category:       "AI Platform",
			Domains:        normalizeDomains(grokDomains),
			DomainKeywords: normalizeDomainKeywords(grokDomainKeywords),
			CIDRs:          normalizeCIDRs(grokCIDRs),
		},
		{
			ID:             stableServiceID("Social Platform", "X (Twitter)"),
			Name:           "X (Twitter)",
			Category:       "Social Platform",
			Domains:        normalizeDomains(xTwitterDomains),
			DomainKeywords: normalizeDomainKeywords(grokDomainKeywords),
			CIDRs:          normalizeCIDRs(grokCIDRs),
			Aliases:        []string{"X", "Twitter", "推特"},
		},
		{
			ID:       stableServiceID("Social Platform", "Instagram (INS)"),
			Name:     "Instagram (INS)",
			Category: "Social Platform",
			Domains:  normalizeDomains(instagramDomains),
			Aliases:  []string{"INS", "IG", "Instagram"},
		},
		{
			ID:       stableServiceID("Social Platform", "Facebook"),
			Name:     "Facebook",
			Category: "Social Platform",
			Domains:  normalizeDomains(facebookDomains),
			Aliases:  []string{"FB", "Facebook", "脸书", "Messenger", "Meta"},
		},
		{
			ID:       stableServiceID("Social Platform", "Threads"),
			Name:     "Threads",
			Category: "Social Platform",
			Domains:  normalizeDomains(threadsDomains),
			Aliases:  []string{"Threads", "串"},
		},
		{
			ID:       stableServiceID("Social Platform", "WhatsApp"),
			Name:     "WhatsApp",
			Category: "Social Platform",
			Domains:  normalizeDomains(whatsappDomains),
			Aliases:  []string{"WhatsApp", "WA"},
		},
		{
			ID:       stableServiceID("Social Platform", "Telegram"),
			Name:     "Telegram",
			Category: "Social Platform",
			Domains:  normalizeDomains(telegramDomains),
			CIDRs:    normalizeCIDRs(telegramCIDRs),
			Aliases:  []string{"Telegram", "TG", "电报"},
		},
		{
			ID:       stableServiceID("Social Platform", "Discord"),
			Name:     "Discord",
			Category: "Social Platform",
			Domains:  normalizeDomains(discordDomains),
			Aliases:  []string{"Discord", "DC"},
		},
		{
			ID:       stableServiceID("Social Platform", "Reddit"),
			Name:     "Reddit",
			Category: "Social Platform",
			Domains:  normalizeDomains(redditDomains),
			Aliases:  []string{"Reddit"},
		},
		{
			ID:       stableServiceID("Social Platform", "Pinterest"),
			Name:     "Pinterest",
			Category: "Social Platform",
			Domains:  normalizeDomains(pinterestDomains),
			Aliases:  []string{"Pinterest"},
		},
		{
			ID:       stableServiceID("Social Platform", "Line"),
			Name:     "Line",
			Category: "Social Platform",
			Domains:  normalizeDomains(lineDomains),
			Aliases:  []string{"Line", "LINE"},
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

// regionQualifierPattern 匹配上游名称末尾的区域限定标记，例如 "/HK Only/"。
// 该标记只说明规则面向哪类客户端，并不属于服务本身的名称；
// 若不剥离，Claude / Gemini / Google AI Studio 会与别名表、补充域名表、
// 探测域名表和审计标签全部匹配失败。
var regionQualifierPattern = regexp.MustCompile(`(?i)\s*/\s*[a-z]{2,3}\s+only\s*/\s*$`)

func stripRegionQualifier(name string) string {
	return strings.TrimSpace(regionQualifierPattern.ReplaceAllString(name, ""))
}

func canonicalServiceName(name, category string) string {
	name = stripRegionQualifier(name)
	aliases := map[string]string{
		"claude 2":                              "Claude",
		"google aistudio":                       "Google AI Studio",
		"google ai studio":                      "Google AI Studio",
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

// sanitizeStoredDomains 宽松地规范化已存储的域名：保留合法项，丢弃无效项。
func sanitizeStoredDomains(domains []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(domains))
	for _, value := range splitDomainValues(domains) {
		domain := normalizeDomainPattern(value)
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
	labels := strings.Split(domain, ".")
	// 单标签（如 "bbc"、"colab"、"twitter"）在 dnsmasq 中只会匹配顶级域，
	// 不会命中任何真实站点，只会污染规则与冲突检测，因此一律拒绝。
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
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
