//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/macstadium/govergeos"
)

// TestWireGuardList tests the VNetWireGuards service.
func TestWireGuardList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VNetWireGuards.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetWireGuards.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetWireGuards.Get succeeded: Name=%q, MTU=%d, EndpointIP=%q",
				fetched.Name, fetched.MTU, fetched.EndpointIP)
		}
	})

	// Test ListByNetwork
	t.Run("ListByNetwork", func(t *testing.T) {
		if first.VNet == 0 {
			t.Skip("No VNet ID available")
		}
		netWGs, err := client.VNetWireGuards.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetWireGuards.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d WireGuard interfaces in network %d", len(netWGs), int(first.VNet))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.VNet == 0 {
			t.Skip("No name or VNet ID available")
		}
		byName, err := client.VNetWireGuards.GetByName(ctx, int(first.VNet), first.Name)
		if err != nil {
			t.Errorf("VNetWireGuards.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first WireGuard for field verification
	prettyPrint(t, "Sample VNetWireGuard", first)
}

// TestWireGuardPeersList tests the VNetWireGuardPeers service.
func TestWireGuardPeersList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VNetWireGuardPeers.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetWireGuardPeers.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetWireGuardPeers.Get succeeded: Name=%q, AllowedIPs=%q, ConfigureFirewall=%q",
				fetched.Name, fetched.AllowedIPs, fetched.ConfigureFirewall)
		}
	})

	// Test ListByWireGuard
	t.Run("ListByWireGuard", func(t *testing.T) {
		if first.WireGuard == 0 {
			t.Skip("No WireGuard ID available")
		}
		wgPeers, err := client.VNetWireGuardPeers.ListByWireGuard(ctx, int(first.WireGuard))
		if err != nil {
			t.Errorf("VNetWireGuardPeers.ListByWireGuard failed: %v", err)
		} else {
			t.Logf("Found %d peers in WireGuard %d", len(wgPeers), int(first.WireGuard))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.WireGuard == 0 {
			t.Skip("No name or WireGuard ID available")
		}
		byName, err := client.VNetWireGuardPeers.GetByName(ctx, int(first.WireGuard), first.Name)
		if err != nil {
			t.Errorf("VNetWireGuardPeers.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Test peer status
	t.Run("PeerStatus", func(t *testing.T) {
		status, err := client.VNetWireGuardPeerStatus.GetByPeer(ctx, int(first.Key))
		if err != nil {
			t.Logf("VNetWireGuardPeerStatus.GetByPeer: %v (may not exist if peer never connected)", err)
		} else {
			t.Logf("Peer status: LastHandshake=%d, TXBytes=%d, RXBytes=%d",
				status.LastHandshake, status.TXBytes, status.RXBytes)
		}
	})

	// Pretty print first peer for field verification
	prettyPrint(t, "Sample VNetWireGuardPeer", first)
}

// TestIPSecList tests the VNetIPSecs service.
func TestIPSecList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VNetIPSecs.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetIPSecs.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetIPSecs.Get succeeded: Mode=%q, Compress=%v, ExcludeNetwork=%v",
				fetched.Mode, fetched.Compress, fetched.ExcludeNetwork)
		}
	})

	// Test GetByNetwork
	t.Run("GetByNetwork", func(t *testing.T) {
		if first.VNet == 0 {
			t.Skip("No VNet ID available")
		}
		byNet, err := client.VNetIPSecs.GetByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetIPSecs.GetByNetwork failed: %v", err)
		} else {
			t.Logf("GetByNetwork succeeded: Key=%d", int(byNet.Key))
		}
	})

	// Pretty print first IPSec for field verification
	prettyPrint(t, "Sample VNetIPSec", first)
}

// TestIPSecPhase1List tests the VNetIPSecPhase1s service.
func TestIPSecPhase1List(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VNetIPSecPhase1s.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetIPSecPhase1s.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetIPSecPhase1s.Get succeeded: Name=%q, Auth=%q, Auto=%q, DPDAction=%q",
				fetched.Name, fetched.Auth, fetched.Auto, fetched.DPDAction)
		}
	})

	// Test ListByIPSec
	t.Run("ListByIPSec", func(t *testing.T) {
		if first.IPSec == 0 {
			t.Skip("No IPSec ID available")
		}
		ipsecPhase1s, err := client.VNetIPSecPhase1s.ListByIPSec(ctx, int(first.IPSec))
		if err != nil {
			t.Errorf("VNetIPSecPhase1s.ListByIPSec failed: %v", err)
		} else {
			t.Logf("Found %d Phase 1 configs in IPSec %d", len(ipsecPhase1s), int(first.IPSec))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.IPSec == 0 {
			t.Skip("No name or IPSec ID available")
		}
		byName, err := client.VNetIPSecPhase1s.GetByName(ctx, int(first.IPSec), first.Name)
		if err != nil {
			t.Errorf("VNetIPSecPhase1s.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first Phase 1 for field verification
	prettyPrint(t, "Sample VNetIPSecPhase1", first)
}

// TestIPSecPhase2List tests the VNetIPSecPhase2s service.
func TestIPSecPhase2List(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VNetIPSecPhase2s.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetIPSecPhase2s.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetIPSecPhase2s.Get succeeded: Name=%q, Protocol=%q, Lifetime=%d, Ciphers=%q",
				fetched.Name, fetched.Protocol, fetched.Lifetime, fetched.Ciphers)
		}
	})

	// Test ListByPhase1
	t.Run("ListByPhase1", func(t *testing.T) {
		if first.Phase1 == 0 {
			t.Skip("No Phase1 ID available")
		}
		phase1Phase2s, err := client.VNetIPSecPhase2s.ListByPhase1(ctx, int(first.Phase1))
		if err != nil {
			t.Errorf("VNetIPSecPhase2s.ListByPhase1 failed: %v", err)
		} else {
			t.Logf("Found %d Phase 2 configs in Phase 1 %d", len(phase1Phase2s), int(first.Phase1))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.Phase1 == 0 {
			t.Skip("No name or Phase1 ID available")
		}
		byName, err := client.VNetIPSecPhase2s.GetByName(ctx, int(first.Phase1), first.Name)
		if err != nil {
			t.Errorf("VNetIPSecPhase2s.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first Phase 2 for field verification
	prettyPrint(t, "Sample VNetIPSecPhase2", first)
}

// TestIPSecConnectionsList tests the VNetIPSecConnections service.
func TestIPSecConnectionsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VNetIPSecConnections.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetIPSecConnections.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetIPSecConnections.Get succeeded: LocalNetwork=%q, RemoteNetwork=%q, Interface=%q",
				fetched.LocalNetwork, fetched.RemoteNetwork, fetched.Interface)
		}
	})

	// Test ListByNetwork
	t.Run("ListByNetwork", func(t *testing.T) {
		if first.VNet == 0 {
			t.Skip("No VNet ID available")
		}
		netConns, err := client.VNetIPSecConnections.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetIPSecConnections.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d connections in network %d", len(netConns), int(first.VNet))
		}
	})

	// Test ListByPhase1
	t.Run("ListByPhase1", func(t *testing.T) {
		if first.Phase1 == 0 {
			t.Skip("No Phase1 ID available")
		}
		phase1Conns, err := client.VNetIPSecConnections.ListByPhase1(ctx, int(first.Phase1))
		if err != nil {
			t.Errorf("VNetIPSecConnections.ListByPhase1 failed: %v", err)
		} else {
			t.Logf("Found %d connections for Phase 1 %d", len(phase1Conns), int(first.Phase1))
		}
	})

	// Pretty print first connection for field verification
	prettyPrint(t, "Sample VNetIPSecConnection", first)
}

// TestWireGuardCRUD tests Create/Update/Delete operations for WireGuard.
func TestWireGuardCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test network for CRUD operations
	t.Log("Creating test network for WireGuard CRUD tests...")
	testNetwork, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
		Name:        "sdk-vpn-test-network",
		Description: "Temporary network for goVergeOS VPN integration testing - safe to delete",
		Network:     "10.254.0.0/24",
		IPAddress:   "10.254.0.1",
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

	// Create WireGuard interface
	wg, err := client.VNetWireGuards.Create(ctx, &vergeos.VNetWireGuardCreateRequest{
		VNet:       networkID,
		Name:       "test-wg0",
		IP:         "10.254.255.1/24",
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

	// Create a peer (using dummy public key for testing)
	testPeerPublicKey := "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=" // dummy base64
	peer, err := client.VNetWireGuardPeers.Create(ctx, &vergeos.VNetWireGuardPeerCreateRequest{
		WireGuard:         wgID,
		Name:              "test-peer",
		PeerIP:            "10.254.255.2",
		PublicKey:         testPeerPublicKey,
		AllowedIPs:        "10.254.255.2/32",
		ConfigureFirewall: ptr(vergeos.WireGuardPeerFirewallRemoteUser),
	})
	if err != nil {
		t.Logf("VNetWireGuardPeers.Create failed (may require valid public key): %v", err)
	} else {
		peerID := int(peer.Key)
		t.Logf("Created peer: [%d] %s -> %s", peerID, peer.Name, peer.PeerIP)

		// Update peer
		newAllowed := "10.254.255.2/32,192.168.0.0/24"
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
