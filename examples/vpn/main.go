// Example: WireGuard VPN Management
//
// This example demonstrates how to manage WireGuard VPN tunnels in VergeOS.
// WireGuard provides fast, modern VPN connectivity for site-to-site and
// remote user access scenarios.
//
// Usage:
//
//	export VERGEOS_HOST=https://your-vergeos-host
//	export VERGEOS_USERNAME=admin
//	export VERGEOS_PASSWORD=yourpassword
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

func main() {
	// Get configuration from environment
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		log.Fatal("Please set VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD environment variables")
	}

	// Create client
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// =========================================================================
	// List existing WireGuard interfaces
	// =========================================================================
	fmt.Println("=== Listing WireGuard Interfaces ===")
	wgInterfaces, err := client.VNetWireGuards.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list WireGuard interfaces: %v", err)
	}

	fmt.Printf("Found %d WireGuard interface(s)\n", len(wgInterfaces))
	for _, wg := range wgInterfaces {
		fmt.Printf("  - %s (ID: %d, Network: %d)\n", wg.Name, wg.Key, wg.VNet)
		fmt.Printf("    IP: %s, Port: %d, Enabled: %v\n", wg.IP, wg.ListenPort, wg.Enabled)
		if wg.PublicKey != "" {
			fmt.Printf("    Public Key: %s...\n", truncateKey(wg.PublicKey))
		}
	}

	// =========================================================================
	// List WireGuard peers
	// =========================================================================
	fmt.Println("\n=== Listing WireGuard Peers ===")
	peers, err := client.VNetWireGuardPeers.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list WireGuard peers: %v", err)
	}

	fmt.Printf("Found %d WireGuard peer(s)\n", len(peers))
	for _, peer := range peers {
		fmt.Printf("  - %s (ID: %d)\n", peer.Name, peer.Key)
		fmt.Printf("    Peer IP: %s, Allowed IPs: %s\n", peer.PeerIP, peer.AllowedIPs)
		fmt.Printf("    Enabled: %v, Firewall: %s\n", peer.Enabled, peer.ConfigureFirewall)
		if peer.Endpoint != "" {
			fmt.Printf("    Endpoint: %s:%d\n", peer.Endpoint, peer.Port)
		} else {
			fmt.Println("    Endpoint: (roaming client)")
		}
	}

	// =========================================================================
	// Get peer connection status
	// =========================================================================
	if len(peers) > 0 {
		fmt.Println("\n=== Peer Connection Status ===")
		for _, peer := range peers {
			status, err := client.VNetWireGuardPeerStatus.GetByPeer(ctx, int(peer.Key))
			if err != nil {
				fmt.Printf("  %s: Status unavailable\n", peer.Name)
				continue
			}

			lastHandshake := "Never"
			if status.LastHandshake > 0 {
				t := time.Unix(status.LastHandshake, 0)
				lastHandshake = t.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("  %s:\n", peer.Name)
			fmt.Printf("    Last Handshake: %s\n", lastHandshake)
			fmt.Printf("    TX: %s, RX: %s\n", formatBytes(status.TXBytes), formatBytes(status.RXBytes))
		}
	}

	// =========================================================================
	// If we have a WireGuard interface, show more details
	// =========================================================================
	if len(wgInterfaces) > 0 {
		wg := wgInterfaces[0]

		fmt.Println("\n=== WireGuard Interface Details ===")
		wgDetail, err := client.VNetWireGuards.Get(ctx, int(wg.Key))
		if err != nil {
			fmt.Printf("Warning: Failed to get WireGuard details: %v\n", err)
		} else {
			fmt.Printf("Name: %s\n", wgDetail.Name)
			fmt.Printf("Description: %s\n", wgDetail.Description)
			fmt.Printf("Network ID: %d\n", wgDetail.VNet)
			fmt.Printf("IP Address: %s\n", wgDetail.IP)
			fmt.Printf("Listen Port: %d\n", wgDetail.ListenPort)
			fmt.Printf("MTU: %d (0 = auto)\n", wgDetail.MTU)
			fmt.Printf("Public Key: %s\n", wgDetail.PublicKey)
			if wgDetail.EndpointIP != "" {
				fmt.Printf("Endpoint IP: %s\n", wgDetail.EndpointIP)
			}
		}

		// List peers for this specific interface
		fmt.Printf("\n=== Peers for %s ===\n", wg.Name)
		wgPeers, err := client.VNetWireGuardPeers.ListByWireGuard(ctx, int(wg.Key))
		if err != nil {
			fmt.Printf("Warning: Failed to list peers: %v\n", err)
		} else {
			fmt.Printf("Found %d peer(s) on this interface\n", len(wgPeers))
			for _, p := range wgPeers {
				fmt.Printf("  - %s: %s -> %s\n", p.Name, p.PeerIP, p.AllowedIPs)
			}
		}
	}

	// =========================================================================
	// IPSec VPN overview
	// =========================================================================
	fmt.Println("\n=== IPSec VPN Overview ===")
	ipsecs, err := client.VNetIPSecs.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list IPSec configurations: %v\n", err)
	} else {
		fmt.Printf("Found %d IPSec configuration(s)\n", len(ipsecs))
		for _, ipsec := range ipsecs {
			fmt.Printf("  - Network %d (ID: %d)\n", ipsec.VNet, ipsec.Key)

			// List Phase 1 connections
			phase1s, _ := client.VNetIPSecPhase1s.ListByIPSec(ctx, int(ipsec.Key))
			for _, p1 := range phase1s {
				fmt.Printf("    Phase 1: %s -> %s (IKE: %s)\n",
					p1.Name, p1.RemoteGateway, p1.KeyExchange)

				// List Phase 2 connections
				phase2s, _ := client.VNetIPSecPhase2s.ListByPhase1(ctx, int(p1.Key))
				for _, p2 := range phase2s {
					fmt.Printf("      Phase 2: %s (%s <-> %s)\n",
						p2.Name, p2.Local, p2.Remote)
				}
			}
		}
	}

	// =========================================================================
	// Active IPSec connections
	// =========================================================================
	fmt.Println("\n=== Active IPSec Connections ===")
	conns, err := client.VNetIPSecConnections.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list IPSec connections: %v\n", err)
	} else {
		if len(conns) == 0 {
			fmt.Println("No active IPSec connections")
		}
		for _, conn := range conns {
			fmt.Printf("  - %s: %s <-> %s\n",
				conn.Connection, conn.LocalNetwork, conn.RemoteNetwork)
		}
	}

	fmt.Println("\n=== VPN Operations Reference ===")
	fmt.Println("WireGuard:")
	fmt.Println("  - client.VNetWireGuards.List/Get/Create/Update/Delete")
	fmt.Println("  - client.VNetWireGuardPeers.List/Get/Create/Update/Delete")
	fmt.Println("  - client.VNetWireGuardPeers.GetConfig(ctx, peerID) - Get peer config file")
	fmt.Println("  - client.VNetWireGuardPeerStatus.GetByPeer(ctx, peerID) - Connection stats")
	fmt.Println("\nIPSec:")
	fmt.Println("  - client.VNetIPSecs.List/Get/Create/Update/Delete")
	fmt.Println("  - client.VNetIPSecPhase1s.List/Get/Create/Update/Delete")
	fmt.Println("  - client.VNetIPSecPhase2s.List/Get/Create/Update/Delete")
	fmt.Println("  - client.VNetIPSecConnections.List - Active tunnel status")

	fmt.Println("\n=== Done ===")
}

// truncateKey returns the first 20 characters of a base64 key for display
func truncateKey(key string) string {
	if len(key) > 20 {
		return key[:20]
	}
	return key
}

// formatBytes formats bytes as a human-readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
