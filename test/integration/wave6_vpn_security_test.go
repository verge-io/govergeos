//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave6VPNSecurity tests the Wave 6 VPN and Security services (WireGuard, IPSec, Certificates)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave6
func TestWave6VPNSecurity(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("Certificates", func(t *testing.T) {
		testCertificates(t, ctx, client)
	})

	t.Run("VNetWireGuards", func(t *testing.T) {
		testVNetWireGuards(t, ctx, client)
	})

	t.Run("VNetWireGuardPeers", func(t *testing.T) {
		testVNetWireGuardPeers(t, ctx, client)
	})

	t.Run("VNetIPSecs", func(t *testing.T) {
		testVNetIPSecs(t, ctx, client)
	})

	t.Run("VNetIPSecPhase1s", func(t *testing.T) {
		testVNetIPSecPhase1s(t, ctx, client)
	})

	t.Run("VNetIPSecPhase2s", func(t *testing.T) {
		testVNetIPSecPhase2s(t, ctx, client)
	})

	t.Run("VNetIPSecConnections", func(t *testing.T) {
		testVNetIPSecConnections(t, ctx, client)
	})
}

func testCertificates(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Certificates service...")

	// List all certificates
	certs, err := client.Certificates.List(ctx)
	if err != nil {
		t.Fatalf("Certificates.List failed: %v", err)
	}

	t.Logf("Found %d certificates", len(certs))

	if len(certs) == 0 {
		t.Log("No certificates found - this is normal for fresh installations")
		return
	}

	// Log first certificate to verify field mapping
	first := certs[0]
	t.Logf("First certificate: Key=%d, Domain=%q, Type=%q, Valid=%v, Expires=%d",
		int(first.Key), first.Domain, first.Type, first.Valid, first.Expires)

	// Test Get by ID
	fetched, err := client.Certificates.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("Certificates.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("Certificates.Get succeeded: Domain=%q, KeyType=%q, AutoCreated=%v",
			fetched.Domain, fetched.KeyType, fetched.AutoCreated)
	}

	// Test ListValid
	validCerts, err := client.Certificates.ListValid(ctx)
	if err != nil {
		t.Errorf("Certificates.ListValid failed: %v", err)
	} else {
		t.Logf("Found %d valid certificates", len(validCerts))
	}

	// Test ListByType
	if first.Type != "" {
		typeCerts, err := client.Certificates.ListByType(ctx, first.Type)
		if err != nil {
			t.Errorf("Certificates.ListByType failed: %v", err)
		} else {
			t.Logf("Found %d certificates of type %q", len(typeCerts), first.Type)
		}
	}

	// Test GetByDomain
	if first.Domain != "" {
		byDomain, err := client.Certificates.GetByDomain(ctx, first.Domain)
		if err != nil {
			t.Errorf("Certificates.GetByDomain failed: %v", err)
		} else {
			t.Logf("GetByDomain succeeded: Key=%d", int(byDomain.Key))
		}
	}

	// Pretty print first certificate for field verification
	prettyPrint(t, "Sample Certificate", first)
}

func testVNetWireGuards(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetWireGuards service...")

	// List all WireGuard interfaces
	wgs, err := client.VNetWireGuards.List(ctx)
	if err != nil {
		t.Fatalf("VNetWireGuards.List failed: %v", err)
	}

	t.Logf("Found %d WireGuard interfaces", len(wgs))

	if len(wgs) == 0 {
		t.Log("No WireGuard interfaces found - this is normal if WireGuard VPNs are not configured")
		return
	}

	// Log first WireGuard to verify field mapping
	first := wgs[0]
	t.Logf("First WireGuard: Key=%d, Name=%q, VNet=%d, IP=%q, ListenPort=%d, Enabled=%v",
		int(first.Key), first.Name, int(first.VNet), first.IP, first.ListenPort, first.Enabled)

	// Verify PublicKey is populated (should be auto-generated)
	if first.PublicKey == "" {
		t.Log("Warning: PublicKey is empty - may not be returned by default")
	} else {
		t.Logf("PublicKey length: %d chars", len(first.PublicKey))
	}

	// Test Get by ID
	fetched, err := client.VNetWireGuards.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetWireGuards.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetWireGuards.Get succeeded: Name=%q, MTU=%d, EndpointIP=%q",
			fetched.Name, fetched.MTU, fetched.EndpointIP)
	}

	// Test ListByNetwork
	if first.VNet > 0 {
		netWGs, err := client.VNetWireGuards.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetWireGuards.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d WireGuard interfaces in network %d", len(netWGs), int(first.VNet))
		}
	}

	// Test GetByName
	if first.Name != "" && first.VNet > 0 {
		byName, err := client.VNetWireGuards.GetByName(ctx, int(first.VNet), first.Name)
		if err != nil {
			t.Errorf("VNetWireGuards.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first WireGuard for field verification
	prettyPrint(t, "Sample VNetWireGuard", first)
}

func testVNetWireGuardPeers(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetWireGuardPeers service...")

	// List all WireGuard peers
	peers, err := client.VNetWireGuardPeers.List(ctx)
	if err != nil {
		t.Fatalf("VNetWireGuardPeers.List failed: %v", err)
	}

	t.Logf("Found %d WireGuard peers", len(peers))

	if len(peers) == 0 {
		t.Log("No WireGuard peers found - this is normal if no peers are configured")
		return
	}

	// Log first peer to verify field mapping
	first := peers[0]
	t.Logf("First peer: Key=%d, Name=%q, WireGuard=%d, PeerIP=%q, Endpoint=%q, Enabled=%v",
		int(first.Key), first.Name, int(first.WireGuard), first.PeerIP, first.Endpoint, first.Enabled)

	// Test Get by ID
	fetched, err := client.VNetWireGuardPeers.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetWireGuardPeers.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetWireGuardPeers.Get succeeded: Name=%q, AllowedIPs=%q, ConfigureFirewall=%q",
			fetched.Name, fetched.AllowedIPs, fetched.ConfigureFirewall)
	}

	// Test ListByWireGuard
	if first.WireGuard > 0 {
		wgPeers, err := client.VNetWireGuardPeers.ListByWireGuard(ctx, int(first.WireGuard))
		if err != nil {
			t.Errorf("VNetWireGuardPeers.ListByWireGuard failed: %v", err)
		} else {
			t.Logf("Found %d peers in WireGuard %d", len(wgPeers), int(first.WireGuard))
		}
	}

	// Test GetByName
	if first.Name != "" && first.WireGuard > 0 {
		byName, err := client.VNetWireGuardPeers.GetByName(ctx, int(first.WireGuard), first.Name)
		if err != nil {
			t.Errorf("VNetWireGuardPeers.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Test GetConfig (only works if autogenerate_peer is enabled)
	if first.AutogeneratePeer {
		config, err := client.VNetWireGuardPeers.GetConfig(ctx, int(first.Key))
		if err != nil {
			t.Logf("VNetWireGuardPeers.GetConfig failed (expected if wg_config field not available): %v", err)
		} else if config != "" {
			t.Logf("GetConfig succeeded: config length=%d bytes", len(config))
		}
	}

	// Test peer status
	status, err := client.VNetWireGuardPeerStatus.GetByPeer(ctx, int(first.Key))
	if err != nil {
		t.Logf("VNetWireGuardPeerStatus.GetByPeer: %v (may not exist if peer never connected)", err)
	} else {
		t.Logf("Peer status: LastHandshake=%d, TXBytes=%d, RXBytes=%d",
			status.LastHandshake, status.TXBytes, status.RXBytes)
	}

	// Pretty print first peer for field verification
	prettyPrint(t, "Sample VNetWireGuardPeer", first)
}

func testVNetIPSecs(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetIPSecs service...")

	// List all IPSec configurations
	ipsecs, err := client.VNetIPSecs.List(ctx)
	if err != nil {
		t.Fatalf("VNetIPSecs.List failed: %v", err)
	}

	t.Logf("Found %d IPSec configurations", len(ipsecs))

	if len(ipsecs) == 0 {
		t.Log("No IPSec configurations found - this is normal if IPSec VPNs are not configured")
		return
	}

	// Log first IPSec to verify field mapping
	first := ipsecs[0]
	t.Logf("First IPSec: Key=%d, VNet=%d, Enabled=%v, Mode=%q, UniqueIDs=%q",
		int(first.Key), int(first.VNet), first.Enabled, first.Mode, first.UniqueIDs)

	// Test Get by ID
	fetched, err := client.VNetIPSecs.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetIPSecs.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetIPSecs.Get succeeded: Mode=%q, Compress=%v, ExcludeNetwork=%v",
			fetched.Mode, fetched.Compress, fetched.ExcludeNetwork)
	}

	// Test GetByNetwork
	if first.VNet > 0 {
		byNet, err := client.VNetIPSecs.GetByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetIPSecs.GetByNetwork failed: %v", err)
		} else {
			t.Logf("GetByNetwork succeeded: Key=%d", int(byNet.Key))
		}
	}

	// Pretty print first IPSec for field verification
	prettyPrint(t, "Sample VNetIPSec", first)
}

func testVNetIPSecPhase1s(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetIPSecPhase1s service...")

	// List all Phase 1 configurations
	phase1s, err := client.VNetIPSecPhase1s.List(ctx)
	if err != nil {
		t.Fatalf("VNetIPSecPhase1s.List failed: %v", err)
	}

	t.Logf("Found %d IPSec Phase 1 configurations", len(phase1s))

	if len(phase1s) == 0 {
		t.Log("No IPSec Phase 1 configurations found - this is normal if IPSec VPNs are not configured")
		return
	}

	// Log first Phase 1 to verify field mapping
	first := phase1s[0]
	t.Logf("First Phase1: Key=%d, Name=%q, IPSec=%d, RemoteGateway=%q, KeyExchange=%q, Enabled=%v",
		int(first.Key), first.Name, int(first.IPSec), first.RemoteGateway, first.KeyExchange, first.Enabled)

	// Test Get by ID
	fetched, err := client.VNetIPSecPhase1s.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetIPSecPhase1s.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetIPSecPhase1s.Get succeeded: Name=%q, Auth=%q, Auto=%q, DPDAction=%q",
			fetched.Name, fetched.Auth, fetched.Auto, fetched.DPDAction)
	}

	// Test ListByIPSec
	if first.IPSec > 0 {
		ipsecPhase1s, err := client.VNetIPSecPhase1s.ListByIPSec(ctx, int(first.IPSec))
		if err != nil {
			t.Errorf("VNetIPSecPhase1s.ListByIPSec failed: %v", err)
		} else {
			t.Logf("Found %d Phase 1 configs in IPSec %d", len(ipsecPhase1s), int(first.IPSec))
		}
	}

	// Test GetByName
	if first.Name != "" && first.IPSec > 0 {
		byName, err := client.VNetIPSecPhase1s.GetByName(ctx, int(first.IPSec), first.Name)
		if err != nil {
			t.Errorf("VNetIPSecPhase1s.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first Phase 1 for field verification
	prettyPrint(t, "Sample VNetIPSecPhase1", first)
}

func testVNetIPSecPhase2s(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetIPSecPhase2s service...")

	// List all Phase 2 configurations
	phase2s, err := client.VNetIPSecPhase2s.List(ctx)
	if err != nil {
		t.Fatalf("VNetIPSecPhase2s.List failed: %v", err)
	}

	t.Logf("Found %d IPSec Phase 2 configurations", len(phase2s))

	if len(phase2s) == 0 {
		t.Log("No IPSec Phase 2 configurations found - this is normal if IPSec VPNs are not configured")
		return
	}

	// Log first Phase 2 to verify field mapping
	first := phase2s[0]
	t.Logf("First Phase2: Key=%d, Name=%q, Phase1=%d, Local=%q, Remote=%q, Mode=%q, Enabled=%v",
		int(first.Key), first.Name, int(first.Phase1), first.Local, first.Remote, first.Mode, first.Enabled)

	// Test Get by ID
	fetched, err := client.VNetIPSecPhase2s.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetIPSecPhase2s.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetIPSecPhase2s.Get succeeded: Name=%q, Protocol=%q, Lifetime=%d, Ciphers=%q",
			fetched.Name, fetched.Protocol, fetched.Lifetime, fetched.Ciphers)
	}

	// Test ListByPhase1
	if first.Phase1 > 0 {
		phase1Phase2s, err := client.VNetIPSecPhase2s.ListByPhase1(ctx, int(first.Phase1))
		if err != nil {
			t.Errorf("VNetIPSecPhase2s.ListByPhase1 failed: %v", err)
		} else {
			t.Logf("Found %d Phase 2 configs in Phase 1 %d", len(phase1Phase2s), int(first.Phase1))
		}
	}

	// Test GetByName
	if first.Name != "" && first.Phase1 > 0 {
		byName, err := client.VNetIPSecPhase2s.GetByName(ctx, int(first.Phase1), first.Name)
		if err != nil {
			t.Errorf("VNetIPSecPhase2s.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first Phase 2 for field verification
	prettyPrint(t, "Sample VNetIPSecPhase2", first)
}

func testVNetIPSecConnections(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetIPSecConnections service...")

	// List all active IPSec connections
	conns, err := client.VNetIPSecConnections.List(ctx)
	if err != nil {
		t.Fatalf("VNetIPSecConnections.List failed: %v", err)
	}

	t.Logf("Found %d active IPSec connections", len(conns))

	if len(conns) == 0 {
		t.Log("No active IPSec connections found - this is normal if no tunnels are established")
		return
	}

	// Log first connection to verify field mapping
	first := conns[0]
	t.Logf("First connection: Key=%d, VNet=%d, Connection=%q, Local=%q, Remote=%q, Protocol=%q",
		int(first.Key), int(first.VNet), first.Connection, first.Local, first.Remote, first.Protocol)

	// Test Get by ID
	fetched, err := client.VNetIPSecConnections.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("VNetIPSecConnections.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("VNetIPSecConnections.Get succeeded: LocalNetwork=%q, RemoteNetwork=%q, Interface=%q",
			fetched.LocalNetwork, fetched.RemoteNetwork, fetched.Interface)
	}

	// Test ListByNetwork
	if first.VNet > 0 {
		netConns, err := client.VNetIPSecConnections.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetIPSecConnections.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d connections in network %d", len(netConns), int(first.VNet))
		}
	}

	// Test ListByPhase1
	if first.Phase1 > 0 {
		phase1Conns, err := client.VNetIPSecConnections.ListByPhase1(ctx, int(first.Phase1))
		if err != nil {
			t.Errorf("VNetIPSecConnections.ListByPhase1 failed: %v", err)
		} else {
			t.Logf("Found %d connections for Phase 1 %d", len(phase1Conns), int(first.Phase1))
		}
	}

	// Pretty print first connection for field verification
	prettyPrint(t, "Sample VNetIPSecConnection", first)
}

// TestWave6VPNCRUD tests Create/Update/Delete operations for Wave 6 VPN services.
// This test creates a dedicated test network to avoid modifying production data.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave6VPNCRUD
func TestWave6VPNCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test network for CRUD operations
	t.Log("Creating test network for Wave 6 CRUD tests...")
	testNetwork, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
		Name:        "sdk-wave6-test-network",
		Description: "Temporary network for Wave 6 goVergeOS integration testing - safe to delete",
		Network:     "10.253.0.0/24",
		IPAddress:   "10.253.0.1",
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
	t.Run("WireGuardCRUD", func(t *testing.T) {
		testWireGuardCRUD(t, ctx, client, networkID)
	})

	// Note: IPSec CRUD requires external gateway, skip in automated tests
	// t.Run("IPSecCRUD", func(t *testing.T) {
	// 	testIPSecCRUD(t, ctx, client, networkID)
	// })
}

func testWireGuardCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, networkID int) {
	t.Log("Testing WireGuard CRUD...")

	// Create WireGuard interface
	wg, err := client.VNetWireGuards.Create(ctx, &vergeos.VNetWireGuardCreateRequest{
		VNet:       networkID,
		Name:       "test-wg0",
		IP:         "10.253.255.1/24",
		ListenPort: ptr(51820),
	})
	if err != nil {
		t.Fatalf("VNetWireGuards.Create failed: %v", err)
	}
	wgID := int(wg.Key)
	t.Logf("Created WireGuard: [%d] %s IP=%s Port=%d", wgID, wg.Name, wg.IP, wg.ListenPort)
	t.Logf("Generated PublicKey: %s", wg.PublicKey)

	// Read
	wg, err = client.VNetWireGuards.Get(ctx, wgID)
	if err != nil {
		t.Fatalf("VNetWireGuards.Get failed: %v", err)
	}
	t.Logf("Read WireGuard: [%d] %s Enabled=%v", wgID, wg.Name, wg.Enabled)

	// Update
	newDesc := "Updated test WireGuard"
	wg, err = client.VNetWireGuards.Update(ctx, wgID, &vergeos.VNetWireGuardUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("VNetWireGuards.Update failed: %v", err)
	}
	if wg.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, wg.Description)
	} else {
		t.Logf("Updated WireGuard description to: %q", wg.Description)
	}

	// Create a peer
	// Note: We need a valid base64 public key for the peer
	// Using a dummy key for testing (would need real key for actual connection)
	testPeerPublicKey := "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=" // dummy base64
	peer, err := client.VNetWireGuardPeers.Create(ctx, &vergeos.VNetWireGuardPeerCreateRequest{
		WireGuard:         wgID,
		Name:              "test-peer",
		PeerIP:            "10.253.255.2",
		PublicKey:         testPeerPublicKey,
		AllowedIPs:        "10.253.255.2/32",
		ConfigureFirewall: ptr(vergeos.WireGuardPeerFirewallRemoteUser),
	})
	if err != nil {
		t.Logf("VNetWireGuardPeers.Create failed (may require valid public key): %v", err)
	} else {
		peerID := int(peer.Key)
		t.Logf("Created peer: [%d] %s -> %s", peerID, peer.Name, peer.PeerIP)

		// Update peer
		newAllowed := "10.253.255.2/32,192.168.0.0/24"
		peer, err = client.VNetWireGuardPeers.Update(ctx, peerID, &vergeos.VNetWireGuardPeerUpdateRequest{
			AllowedIPs: &newAllowed,
		})
		if err != nil {
			t.Errorf("VNetWireGuardPeers.Update failed: %v", err)
		} else {
			t.Logf("Updated peer AllowedIPs to: %s", peer.AllowedIPs)
		}

		// Delete peer
		if err := client.VNetWireGuardPeers.Delete(ctx, peerID); err != nil {
			t.Errorf("VNetWireGuardPeers.Delete failed: %v", err)
		} else {
			t.Log("Deleted peer successfully")
		}
	}

	// Delete WireGuard
	err = client.VNetWireGuards.Delete(ctx, wgID)
	if err != nil {
		t.Fatalf("VNetWireGuards.Delete failed: %v", err)
	}
	t.Log("Deleted WireGuard successfully")

	// Verify deletion
	_, err = client.VNetWireGuards.Get(ctx, wgID)
	if err == nil {
		t.Error("Expected error after deletion, but got none")
	} else if !vergeos.IsNotFoundError(err) {
		t.Logf("Got expected error after deletion: %v", err)
	} else {
		t.Log("Verified: WireGuard correctly deleted (NotFoundError)")
	}
}

// TestCertificateCRUD tests certificate create/update/delete operations.
// This creates a self-signed certificate to avoid needing external domains.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestCertificateCRUD
func TestCertificateCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Certificate CRUD with self-signed certificate...")

	// Create a self-signed certificate
	cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
		DomainName:  "sdk-test.local",
		Type:        vergeos.CertificateTypeSelfSigned,
		Description: "goVergeOS integration test certificate - safe to delete",
		KeyType:     vergeos.CertificateKeyTypeECDSA,
	})
	if err != nil {
		t.Fatalf("Certificates.Create (self-signed) failed: %v", err)
	}
	certID := int(cert.Key)
	t.Logf("Created certificate: [%d] Domain=%q Type=%q Valid=%v", certID, cert.Domain, cert.Type, cert.Valid)

	// Read
	cert, err = client.Certificates.Get(ctx, certID)
	if err != nil {
		t.Fatalf("Certificates.Get failed: %v", err)
	}
	t.Logf("Read certificate: [%d] Domain=%q Expires=%d", certID, cert.Domain, cert.Expires)

	// Get with keys
	certWithKeys, err := client.Certificates.GetWithKeys(ctx, certID)
	if err != nil {
		t.Errorf("Certificates.GetWithKeys failed: %v", err)
	} else {
		t.Logf("GetWithKeys: Public key length=%d, Private present=%v",
			len(certWithKeys.Public), certWithKeys.Private != "")
	}

	// Update
	newDesc := "Updated goVergeOS test certificate"
	cert, err = client.Certificates.Update(ctx, certID, &vergeos.CertificateUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Errorf("Certificates.Update failed: %v", err)
	} else {
		t.Logf("Updated certificate description to: %q", cert.Description)
	}

	// Delete
	err = client.Certificates.Delete(ctx, certID)
	if err != nil {
		t.Fatalf("Certificates.Delete failed: %v", err)
	}
	t.Log("Deleted certificate successfully")

	// Verify deletion
	_, err = client.Certificates.Get(ctx, certID)
	if err == nil {
		t.Error("Expected error after deletion, but got none")
	} else if !vergeos.IsNotFoundError(err) {
		t.Logf("Got expected error after deletion: %v", err)
	} else {
		t.Log("Verified: Certificate correctly deleted (NotFoundError)")
	}
}
