package main

import (
	"strings"
	"testing"
)

func TestYouTubeSupplementsAreApplied(t *testing.T) {
	list := "# ---> Global Platform\n# > YouTube\nnameserver /youtube.com/<replace with groupname>\n"
	services, err := ParseSmartDNS(strings.NewReader(list))
	if err != nil {
		t.Fatal(err)
	}
	var youtube *Service
	for index := range services {
		if services[index].Name == "YouTube" {
			youtube = &services[index]
		}
	}
	if youtube == nil {
		t.Fatalf("YouTube not found: %v", services)
	}
	for _, want := range []string{"youtube.com", "googlevideo.com", "ytimg.com", "ggpht.com", "youtubekids.com", "jnn-pa.googleapis.com", "yt.be", "youtube.googleapis.com"} {
		if !contains(youtube.Domains, want) {
			t.Errorf("YouTube is missing %s: %v", want, youtube.Domains)
		}
	}
}
