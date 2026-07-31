package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestConnectivityRejectsOversizedDomainSetWithoutTruncating(t *testing.T) {
	domains := make([]string, maxConnectivityDomains+1)
	for index := range domains {
		domains[index] = fmt.Sprintf("dependency-%d.example.com", index)
	}
	_, err := TestConnectivity(context.Background(), ConnectivityRequest{
		ProxyServer: "127.0.0.1:443",
		Domains:     domains,
	})
	if err == nil || !strings.Contains(err.Error(), "maximum is 100") {
		t.Fatalf("expected explicit domain limit error, got %v", err)
	}
}
