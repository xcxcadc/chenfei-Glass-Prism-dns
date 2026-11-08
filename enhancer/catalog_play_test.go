package main

import (
	"strings"
	"testing"
)

func TestGooglePlayStoreBuiltInService(t *testing.T) {
	services := ensureBuiltInServices(nil)
	var play, youtube *Service
	list := "# ---> Global Platform\n# > YouTube\nnameserver /youtube.com/<replace with groupname>\n"
	parsed, err := ParseSmartDNS(strings.NewReader(list))
	if err != nil {
		t.Fatal(err)
	}
	for index := range services {
		if services[index].Name == "Google Play Store" {
			play = &services[index]
		}
	}
	for index := range parsed {
		if parsed[index].Name == "YouTube" {
			youtube = &parsed[index]
		}
	}
	if play == nil || youtube == nil {
		t.Fatalf("missing services: play=%v youtube=%v", play, youtube)
	}
	for _, want := range []string{"play-fe.googleapis.com", "android.clients.google.com", "play.google.com"} {
		if !contains(play.Domains, want) {
			t.Errorf("Play Store missing %s: %v", want, play.Domains)
		}
	}
	if domainsOverlap(routingDomains(play.Domains), routingDomains(youtube.Domains)) {
		t.Errorf("Play Store must not overlap YouTube: %v vs %v", play.Domains, youtube.Domains)
	}
	spec := mediaTestSpecForService(*play)
	if len(spec.Required) != 1 || !globalMediaLabel(spec.Required[0]) {
		t.Errorf("Play Store audit spec not covered: %+v", spec)
	}
}
