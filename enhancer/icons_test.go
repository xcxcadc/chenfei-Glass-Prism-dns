package main

import "testing"

func TestPreferredIconDomainUsesServiceHomepage(t *testing.T) {
	tests := []struct {
		name   string
		domain string
	}{
		{name: "Gemini", domain: "gemini.google.com"},
		{name: "Google AI Studio", domain: "aistudio.google.com"},
		{name: "ChatGPT / OpenAI", domain: "chatgpt.com"},
		{name: "Claude", domain: "claude.ai"},
		{name: "Grok", domain: "grok.com"},
		{name: "Disney+", domain: "disneyplus.com"},
		{name: "HBO / Max", domain: "max.com"},
		{name: "Microsoft Copilot Image Creator", domain: "copilot.microsoft.com"},
		{name: "X", domain: "x.com"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := Service{Name: test.name, Domains: []string{"dependency.invalid"}}
			if got := preferredIconDomain(service); got != test.domain {
				t.Fatalf("preferredIconDomain(%q) = %q, want %q", test.name, got, test.domain)
			}
		})
	}
}

func TestPreferredIconURLsUseOfficialAssets(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "CBC Gem", url: "https://gem.cbc.ca/favicon.ico"},
		{name: "Grok", url: "https://grok.com/favicon.ico"},
		{name: "X", url: "https://x.com/favicon.ico"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urls := preferredIconURLs(Service{Name: test.name})
			if len(urls) == 0 || urls[0] != test.url {
				t.Fatalf("preferredIconURLs(%q) = %v, want first URL %q", test.name, urls, test.url)
			}
		})
	}
}

func TestIsAllowedIconRedirectOnlyAllowsGoogleStaticAssets(t *testing.T) {
	tests := []struct {
		source string
		target string
		want   bool
	}{
		{source: "www.google.com", target: "t2.gstatic.com", want: true},
		{source: "www.google.com", target: "gstatic.com", want: true},
		{source: "www.google.com", target: "evil.gstatic.com.example", want: false},
		{source: "example.com", target: "cdn.example.com", want: false},
	}
	for _, test := range tests {
		if got := isAllowedIconRedirect(test.source, test.target); got != test.want {
			t.Fatalf("isAllowedIconRedirect(%q, %q) = %v, want %v", test.source, test.target, got, test.want)
		}
	}
}
