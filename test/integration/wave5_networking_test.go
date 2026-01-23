//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/vergeos-go-sdk"
)

// TestWave5Networking tests the Wave 5 networking services (VNetAddresses, VNetDNSViews,
// VNetDNSZones, VNetDNSRecords, VNetHosts) against a live VergeOS API.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave5
func TestWave5Networking(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("VNetAddresses", func(t *testing.T) {
		testVNetAddresses(t, ctx, client)
	})

	t.Run("VNetDNSViews", func(t *testing.T) {
		testVNetDNSViews(t, ctx, client)
	})

	t.Run("VNetDNSZones", func(t *testing.T) {
		testVNetDNSZones(t, ctx, client)
	})

	t.Run("VNetDNSRecords", func(t *testing.T) {
		testVNetDNSRecords(t, ctx, client)
	})

	t.Run("VNetHosts", func(t *testing.T) {
		testVNetHosts(t, ctx, client)
	})
}

func testVNetAddresses(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetAddresses service...")

	// List all addresses
	addresses, err := client.VNetAddresses.List(ctx)
	if err != nil {
		t.Fatalf("VNetAddresses.List failed: %v", err)
	}

	t.Logf("Found %d addresses", len(addresses))

	if len(addresses) == 0 {
		t.Log("No addresses found - networks may not have any IP addresses configured")
		return
	}

	// Log first address to verify field mapping
	first := addresses[0]
	t.Logf("First address: Key=%d, IP=%q, Type=%q, MAC=%q, Hostname=%q, VNet=%d",
		int(first.Key), first.IP, first.Type, first.MAC, first.Hostname, int(first.VNet))

	// Test Get by ID
	fetched, err := client.VNetAddresses.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetAddresses.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetAddresses.Get succeeded: IP=%q, Owner=%q", fetched.IP, fetched.Owner)
	}

	// Test ListByNetwork
	if first.VNet > 0 {
		netAddrs, err := client.VNetAddresses.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetAddresses.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d addresses in network %d", len(netAddrs), int(first.VNet))
		}
	}

	// Test ListByType
	staticAddrs, err := client.VNetAddresses.ListByType(ctx, int(first.VNet), vergeos.AddressTypeStatic)
	if err != nil {
		t.Errorf("VNetAddresses.ListByType failed: %v", err)
	} else {
		t.Logf("Found %d static addresses in network %d", len(staticAddrs), int(first.VNet))
	}

	// Test GetByIP if we have an IP
	if first.IP != "" && first.VNet > 0 {
		byIP, err := client.VNetAddresses.GetByIP(ctx, int(first.VNet), first.IP)
		if err != nil {
			t.Errorf("VNetAddresses.GetByIP failed: %v", err)
		} else {
			t.Logf("GetByIP succeeded: Key=%d", int(byIP.Key))
		}
	}

	// Pretty print first address for field verification
	prettyPrint(t, "Sample VNetAddress", first)
}

func testVNetDNSViews(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetDNSViews service...")

	// List all DNS views
	views, err := client.VNetDNSViews.List(ctx)
	if err != nil {
		t.Fatalf("VNetDNSViews.List failed: %v", err)
	}

	t.Logf("Found %d DNS views", len(views))

	if len(views) == 0 {
		t.Log("No DNS views found - this is normal if DNS is not configured")
		return
	}

	// Log first view to verify field mapping
	first := views[0]
	t.Logf("First view: Key=%d, Name=%q, VNet=%d, Recursion=%v, MatchClients=%q",
		int(first.Key), first.Name, int(first.VNet), first.Recursion, first.MatchClients)

	// Test Get by ID
	fetched, err := client.VNetDNSViews.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetDNSViews.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetDNSViews.Get succeeded: Name=%q, MaxCacheSize=%d", fetched.Name, fetched.MaxCacheSize)
	}

	// Test ListByNetwork
	if first.VNet > 0 {
		netViews, err := client.VNetDNSViews.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetDNSViews.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d DNS views in network %d", len(netViews), int(first.VNet))
		}
	}

	// Test GetByName
	if first.Name != "" && first.VNet > 0 {
		byName, err := client.VNetDNSViews.GetByName(ctx, int(first.VNet), first.Name)
		if err != nil {
			t.Errorf("VNetDNSViews.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first view for field verification
	prettyPrint(t, "Sample VNetDNSView", first)
}

func testVNetDNSZones(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetDNSZones service...")

	// List all DNS zones
	zones, err := client.VNetDNSZones.List(ctx)
	if err != nil {
		t.Fatalf("VNetDNSZones.List failed: %v", err)
	}

	t.Logf("Found %d DNS zones", len(zones))

	if len(zones) == 0 {
		t.Log("No DNS zones found - this is normal if DNS is not configured")
		return
	}

	// Log first zone to verify field mapping
	first := zones[0]
	t.Logf("First zone: Key=%d, Domain=%q, Type=%q, View=%d, DefaultTTL=%q, SerialNumber=%d",
		int(first.Key), first.Domain, first.Type, int(first.View), first.DefaultTTL, first.SerialNumber)

	// Test Get by ID
	fetched, err := client.VNetDNSZones.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetDNSZones.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetDNSZones.Get succeeded: Domain=%q, Nameserver=%q", fetched.Domain, fetched.Nameserver)
	}

	// Test ListByView
	if first.View > 0 {
		viewZones, err := client.VNetDNSZones.ListByView(ctx, int(first.View))
		if err != nil {
			t.Errorf("VNetDNSZones.ListByView failed: %v", err)
		} else {
			t.Logf("Found %d DNS zones in view %d", len(viewZones), int(first.View))
		}
	}

	// Test GetByDomain
	if first.Domain != "" && first.View > 0 {
		byDomain, err := client.VNetDNSZones.GetByDomain(ctx, int(first.View), first.Domain)
		if err != nil {
			t.Errorf("VNetDNSZones.GetByDomain failed: %v", err)
		} else {
			t.Logf("GetByDomain succeeded: Key=%d", int(byDomain.Key))
		}
	}

	// Pretty print first zone for field verification
	prettyPrint(t, "Sample VNetDNSZone", first)
}

func testVNetDNSRecords(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetDNSRecords service...")

	// List all DNS records
	records, err := client.VNetDNSRecords.List(ctx)
	if err != nil {
		t.Fatalf("VNetDNSRecords.List failed: %v", err)
	}

	t.Logf("Found %d DNS records", len(records))

	if len(records) == 0 {
		t.Log("No DNS records found - this is normal if DNS is not configured")
		return
	}

	// Log first record to verify field mapping
	first := records[0]
	t.Logf("First record: Key=%d, Host=%q, Type=%q, Value=%q, Zone=%d, TTL=%q",
		int(first.Key), first.Host, first.Type, first.Value, int(first.Zone), first.TTL)

	// Test Get by ID
	fetched, err := client.VNetDNSRecords.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetDNSRecords.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetDNSRecords.Get succeeded: Host=%q, MXPreference=%d", fetched.Host, fetched.MXPreference)
	}

	// Test ListByZone
	if first.Zone > 0 {
		zoneRecords, err := client.VNetDNSRecords.ListByZone(ctx, int(first.Zone))
		if err != nil {
			t.Errorf("VNetDNSRecords.ListByZone failed: %v", err)
		} else {
			t.Logf("Found %d DNS records in zone %d", len(zoneRecords), int(first.Zone))
		}
	}

	// Test ListByType
	if first.Zone > 0 && first.Type != "" {
		typeRecords, err := client.VNetDNSRecords.ListByType(ctx, int(first.Zone), first.Type)
		if err != nil {
			t.Errorf("VNetDNSRecords.ListByType failed: %v", err)
		} else {
			t.Logf("Found %d %s records in zone %d", len(typeRecords), first.Type, int(first.Zone))
		}
	}

	// Test GetByHostAndType
	if first.Zone > 0 && first.Type != "" {
		byHostType, err := client.VNetDNSRecords.GetByHostAndType(ctx, int(first.Zone), first.Host, first.Type)
		if err != nil {
			t.Errorf("VNetDNSRecords.GetByHostAndType failed: %v", err)
		} else {
			t.Logf("GetByHostAndType succeeded: Key=%d", int(byHostType.Key))
		}
	}

	// Pretty print first record for field verification
	prettyPrint(t, "Sample VNetDNSRecord", first)
}

func testVNetHosts(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetHosts service...")

	// List all host overrides
	hosts, err := client.VNetHosts.List(ctx)
	if err != nil {
		t.Fatalf("VNetHosts.List failed: %v", err)
	}

	t.Logf("Found %d host overrides", len(hosts))

	if len(hosts) == 0 {
		t.Log("No host overrides found - this is normal if none are configured")
		return
	}

	// Log first host to verify field mapping
	first := hosts[0]
	t.Logf("First host: Key=%d, Host=%q, IP=%q, Type=%q, VNet=%d",
		int(first.Key), first.Host, first.IP, first.Type, int(first.VNet))

	// Test Get by ID
	fetched, err := client.VNetHosts.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetHosts.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetHosts.Get succeeded: Host=%q, IP=%q", fetched.Host, fetched.IP)
	}

	// Test ListByNetwork
	if first.VNet > 0 {
		netHosts, err := client.VNetHosts.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetHosts.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d host overrides in network %d", len(netHosts), int(first.VNet))
		}
	}

	// Test GetByHost
	if first.Host != "" && first.VNet > 0 {
		byHost, err := client.VNetHosts.GetByHost(ctx, int(first.VNet), first.Host)
		if err != nil {
			t.Errorf("VNetHosts.GetByHost failed: %v", err)
		} else {
			t.Logf("GetByHost succeeded: Key=%d", int(byHost.Key))
		}
	}

	// Test GetByIP
	if first.IP != "" && first.VNet > 0 {
		byIP, err := client.VNetHosts.GetByIP(ctx, int(first.VNet), first.IP)
		if err != nil {
			t.Errorf("VNetHosts.GetByIP failed: %v", err)
		} else {
			t.Logf("GetByIP succeeded: Key=%d", int(byIP.Key))
		}
	}

	// Pretty print first host for field verification
	prettyPrint(t, "Sample VNetHost", first)
}

// TestWave5NetworkingCRUD tests Create/Update/Delete operations for Wave 5 services.
// This test creates a dedicated test network to avoid modifying production data.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave5NetworkingCRUD
func TestWave5NetworkingCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test network for CRUD operations
	t.Log("Creating test network for CRUD tests...")
	testNetwork, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
		Name:        "sdk-wave5-test-network",
		Description: "Temporary network for Wave 5 SDK integration testing - safe to delete",
		Network:     "10.252.0.0/24",
		IPAddress:   "10.252.0.1",
		DHCPEnabled: ptr(false),
	})
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}
	networkID := int(testNetwork.ID)
	t.Logf("Created test network: %s (ID: %d)", testNetwork.Name, networkID)

	// Cleanup: delete test network when done
	defer func() {
		t.Log("Cleaning up: deleting test network...")
		if err := client.Networks.Delete(ctx, networkID); err != nil {
			t.Logf("Warning: failed to delete test network: %v", err)
		} else {
			t.Log("Test network deleted successfully")
		}
	}()

	// Run CRUD subtests
	t.Run("VNetHostsCRUD", func(t *testing.T) {
		testVNetHostsCRUD(t, ctx, client, networkID)
	})

	t.Run("VNetAddressesCRUD", func(t *testing.T) {
		testVNetAddressesCRUD(t, ctx, client, networkID)
	})

	t.Run("VNetDNSCRUD", func(t *testing.T) {
		testVNetDNSCRUD(t, ctx, client, networkID)
	})
}

func testVNetHostsCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, networkID int) {
	t.Log("Testing VNetHosts CRUD...")

	// Create
	host, err := client.VNetHosts.Create(ctx, &vergeos.VNetHostCreateRequest{
		VNet: networkID,
		Host: "test-server.local",
		IP:   "10.252.0.100",
	})
	if err != nil {
		t.Fatalf("VNetHosts.Create failed: %v", err)
	}
	hostID := int(host.Key)
	t.Logf("Created host: [%d] %s -> %s", hostID, host.Host, host.IP)

	// Read
	host, err = client.VNetHosts.Get(ctx, hostID)
	if err != nil {
		t.Fatalf("VNetHosts.Get failed: %v", err)
	}
	t.Logf("Read host: [%d] %s -> %s (type: %s)", hostID, host.Host, host.IP, host.Type)

	// Update
	host, err = client.VNetHosts.Update(ctx, hostID, &vergeos.VNetHostUpdateRequest{
		IP: ptr("10.252.0.101"),
	})
	if err != nil {
		t.Fatalf("VNetHosts.Update failed: %v", err)
	}
	if host.IP != "10.252.0.101" {
		t.Errorf("Update verification failed: expected IP 10.252.0.101, got %s", host.IP)
	} else {
		t.Logf("Updated host IP to: %s", host.IP)
	}

	// Delete
	err = client.VNetHosts.Delete(ctx, hostID)
	if err != nil {
		t.Fatalf("VNetHosts.Delete failed: %v", err)
	}
	t.Log("Deleted host successfully")

	// Verify deletion
	_, err = client.VNetHosts.Get(ctx, hostID)
	if err == nil {
		t.Error("Expected error after deletion, but got none")
	} else if !vergeos.IsNotFoundError(err) {
		t.Logf("Got expected error after deletion: %v", err)
	} else {
		t.Log("Verified: host correctly deleted (NotFoundError)")
	}
}

func testVNetAddressesCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, networkID int) {
	t.Log("Testing VNetAddresses CRUD...")

	// Create
	addr, err := client.VNetAddresses.Create(ctx, &vergeos.VNetAddressCreateRequest{
		VNet:        networkID,
		Type:        vergeos.AddressTypeStatic,
		IP:          "10.252.0.200",
		Hostname:    "test-static",
		Description: "SDK integration test address",
	})
	if err != nil {
		t.Fatalf("VNetAddresses.Create failed: %v", err)
	}
	addrID := int(addr.Key)
	t.Logf("Created address: [%d] %s (type: %s, hostname: %s)", addrID, addr.IP, addr.Type, addr.Hostname)

	// Read
	addr, err = client.VNetAddresses.Get(ctx, addrID)
	if err != nil {
		t.Fatalf("VNetAddresses.Get failed: %v", err)
	}
	t.Logf("Read address: [%d] %s (hostname: %s)", addrID, addr.IP, addr.Hostname)

	// Update
	addr, err = client.VNetAddresses.Update(ctx, addrID, &vergeos.VNetAddressUpdateRequest{
		Hostname: ptr("test-static-updated"),
	})
	if err != nil {
		t.Fatalf("VNetAddresses.Update failed: %v", err)
	}
	if addr.Hostname != "test-static-updated" {
		t.Errorf("Update verification failed: expected hostname test-static-updated, got %s", addr.Hostname)
	} else {
		t.Logf("Updated address hostname to: %s", addr.Hostname)
	}

	// Delete
	err = client.VNetAddresses.Delete(ctx, addrID)
	if err != nil {
		t.Fatalf("VNetAddresses.Delete failed: %v", err)
	}
	t.Log("Deleted address successfully")
}

func testVNetDNSCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, networkID int) {
	t.Log("Testing VNetDNS CRUD (Views, Zones, Records)...")

	// Create DNS View
	view, err := client.VNetDNSViews.Create(ctx, &vergeos.VNetDNSViewCreateRequest{
		VNet:         networkID,
		Name:         "test-view",
		Recursion:    ptr(true),
		MatchClients: ptr("10.252.0.0/24;"),
	})
	if err != nil {
		t.Fatalf("VNetDNSViews.Create failed: %v", err)
	}
	viewID := int(view.Key)
	t.Logf("Created DNS view: [%d] %s (recursion: %v)", viewID, view.Name, view.Recursion)

	// Update DNS View
	view, err = client.VNetDNSViews.Update(ctx, viewID, &vergeos.VNetDNSViewUpdateRequest{
		MatchClients: ptr("10.252.0.0/24;192.168.0.0/16;"),
	})
	if err != nil {
		t.Errorf("VNetDNSViews.Update failed: %v", err)
	} else {
		t.Logf("Updated DNS view match_clients to: %s", view.MatchClients)
	}

	// Create DNS Zone
	zone, err := client.VNetDNSZones.Create(ctx, &vergeos.VNetDNSZoneCreateRequest{
		View:       viewID,
		Domain:     "test.local",
		Type:       ptr(vergeos.DNSZoneTypeMaster),
		Nameserver: ptr("ns1.test.local"),
		Email:      ptr("admin@test.local"),
		DefaultTTL: ptr("1h"),
	})
	if err != nil {
		t.Fatalf("VNetDNSZones.Create failed: %v", err)
	}
	zoneID := int(zone.Key)
	t.Logf("Created DNS zone: [%d] %s (type: %s)", zoneID, zone.Domain, zone.Type)

	// Update DNS Zone
	zone, err = client.VNetDNSZones.Update(ctx, zoneID, &vergeos.VNetDNSZoneUpdateRequest{
		DefaultTTL: ptr("2h"),
	})
	if err != nil {
		t.Errorf("VNetDNSZones.Update failed: %v", err)
	} else {
		t.Logf("Updated DNS zone TTL to: %s", zone.DefaultTTL)
	}

	// Create DNS Record (A)
	aRecord, err := client.VNetDNSRecords.Create(ctx, &vergeos.VNetDNSRecordCreateRequest{
		Zone:  zoneID,
		Host:  "www",
		Type:  vergeos.DNSRecordTypeA,
		Value: "10.252.0.100",
		TTL:   ptr("30m"),
	})
	if err != nil {
		t.Fatalf("VNetDNSRecords.Create (A) failed: %v", err)
	}
	aRecordID := int(aRecord.Key)
	t.Logf("Created A record: [%d] %s -> %s", aRecordID, aRecord.Host, aRecord.Value)

	// Create DNS Record (MX)
	mxRecord, err := client.VNetDNSRecords.Create(ctx, &vergeos.VNetDNSRecordCreateRequest{
		Zone:         zoneID,
		Host:         "@",
		Type:         vergeos.DNSRecordTypeMX,
		Value:        "mail.test.local",
		MXPreference: ptr(10),
	})
	if err != nil {
		t.Fatalf("VNetDNSRecords.Create (MX) failed: %v", err)
	}
	mxRecordID := int(mxRecord.Key)
	t.Logf("Created MX record: [%d] %s -> %s (pref: %d)", mxRecordID, mxRecord.Host, mxRecord.Value, mxRecord.MXPreference)

	// List records by zone
	records, err := client.VNetDNSRecords.ListByZone(ctx, zoneID)
	if err != nil {
		t.Errorf("VNetDNSRecords.ListByZone failed: %v", err)
	} else {
		t.Logf("Found %d records in zone", len(records))
	}

	// Update A Record
	aRecord, err = client.VNetDNSRecords.Update(ctx, aRecordID, &vergeos.VNetDNSRecordUpdateRequest{
		Value: ptr("10.252.0.101"),
	})
	if err != nil {
		t.Errorf("VNetDNSRecords.Update failed: %v", err)
	} else {
		t.Logf("Updated A record value to: %s", aRecord.Value)
	}

	// Cleanup: Delete records, zone, view (in reverse order)
	t.Log("Cleaning up DNS resources...")

	if err := client.VNetDNSRecords.Delete(ctx, mxRecordID); err != nil {
		t.Errorf("Failed to delete MX record: %v", err)
	}
	if err := client.VNetDNSRecords.Delete(ctx, aRecordID); err != nil {
		t.Errorf("Failed to delete A record: %v", err)
	}
	t.Log("Deleted DNS records")

	if err := client.VNetDNSZones.Delete(ctx, zoneID); err != nil {
		t.Errorf("Failed to delete zone: %v", err)
	}
	t.Log("Deleted DNS zone")

	if err := client.VNetDNSViews.Delete(ctx, viewID); err != nil {
		t.Errorf("Failed to delete view: %v", err)
	}
	t.Log("Deleted DNS view")

	t.Log("VNetDNS CRUD test complete")
}

// ptr is a helper to create pointers to values
func ptr[T any](v T) *T {
	return &v
}
