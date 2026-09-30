//go:build integration

package integration

import (
	"context"
	"testing"
)

// TestVNetProxyList reads proxy configuration and tenant FQDN mappings.
func TestVNetProxyList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	proxies, err := client.VNetProxies.List(ctx)
	if err != nil {
		t.Fatalf("VNetProxies.List failed: %v", err)
	}
	t.Logf("Found %d network proxies", len(proxies))
	if len(proxies) == 0 {
		t.Log("No proxies configured")
		return
	}

	first := proxies[0]
	fetched, err := client.VNetProxies.Get(ctx, int(first.Key))
	if err != nil {
		t.Fatalf("VNetProxies.Get(%d) failed: %v", int(first.Key), err)
	}
	t.Logf("Proxy %d on network %d listens on %s", int(fetched.Key), int(fetched.VNet), fetched.ListenAddress)

	mappings, err := client.VNetProxyTenants.ListByProxy(ctx, int(fetched.Key))
	if err != nil {
		t.Fatalf("VNetProxyTenants.ListByProxy(%d) failed: %v", int(fetched.Key), err)
	}
	t.Logf("Proxy %d has %d tenant mappings", int(fetched.Key), len(mappings))
	if len(mappings) == 0 {
		return
	}
	prettyPrint(t, "Sample proxy tenant", mappings[0])
}
