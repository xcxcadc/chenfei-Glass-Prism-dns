package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	categoryPattern = regexp.MustCompile(`^#\s*-+\s*>\s*(.+?)\s*$`)
	servicePattern  = regexp.MustCompile(`^#\s*>\s*(.+?)\s*$`)
	domainPattern   = regexp.MustCompile(`^nameserver\s+/([^/]+)/`)
	domainValid     = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$|^[a-z0-9]$`)
)

type Service struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Domains  []string `json:"domains"`
	Custom   bool     `json:"custom"`
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
				Name:     serviceName,
				Category: category,
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
		sort.Strings(service.Domains)
		result = append(result, *service)
	}
	return result, nil
}

func normalizeDomain(value string) string {
	domain := strings.ToLower(strings.TrimSpace(value))
	domain = strings.TrimPrefix(domain, "*.")
	domain = strings.TrimPrefix(domain, ".")
	domain = strings.TrimSuffix(domain, ".")
	if domain == "" || len(domain) > 253 || strings.Contains(domain, " ") || !domainValid.MatchString(domain) {
		return ""
	}
	return domain
}

func normalizeDomains(domains []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(domains))
	for _, line := range domains {
		for _, value := range strings.FieldsFunc(line, func(r rune) bool {
			return r == '\n' || r == '\r' || r == ',' || r == ';'
		}) {
			domain := normalizeDomain(value)
			if domain == "" {
				continue
			}
			if _, ok := seen[domain]; ok {
				continue
			}
			seen[domain] = struct{}{}
			result = append(result, domain)
		}
	}
	sort.Strings(result)
	return result
}

func stableServiceID(category, name string) string {
	base := slug(name)
	if base == "" {
		base = "service"
	}
	sum := sha256.Sum256([]byte(category + "\x00" + name))
	return base + "-" + hex.EncodeToString(sum[:4])
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
