// Example: Network Management
//
// This example demonstrates how to create and manage virtual networks
// including DHCP configuration and network lifecycle.
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

	vergeos "github.com/macstadium/govergeos"
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

	// List existing networks
	fmt.Println("=== Existing Networks ===")
	networks, err := client.Networks.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list networks: %v", err)
	}
	for _, net := range networks {
		fmt.Printf("- %s (ID: %d, Network: %s, DHCP: %v)\n",
			net.Name, net.ID, net.Network, net.DHCPEnabled)
	}

	// Create a new network with DHCP
	fmt.Println("\n=== Creating Network ===")
	dhcpEnabled := true
	network, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
		Name:        "sdk-example-network",
		Network:     "192.168.100.0/24",
		IPAddress:   "192.168.100.1",
		DHCPEnabled: &dhcpEnabled,
		DHCPStart:   "192.168.100.100",
		DHCPStop:    "192.168.100.200",
	})
	if err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}
	fmt.Printf("Created network: %s (ID: %d)\n", network.Name, network.ID)
	fmt.Printf("  Network: %s\n", network.Network)
	fmt.Printf("  IP Address: %s\n", network.IPAddress)
	fmt.Printf("  DHCP Enabled: %v\n", network.DHCPEnabled)
	fmt.Printf("  DHCP Range: %s - %s\n", network.DHCPStart, network.DHCPStop)

	// Update the network
	fmt.Println("\n=== Updating Network ===")
	newDHCPStop := "192.168.100.250"
	network, err = client.Networks.Update(ctx, network.ID.Int(), &vergeos.NetworkUpdateRequest{
		DHCPStop: &newDHCPStop,
	})
	if err != nil {
		log.Fatalf("Failed to update network: %v", err)
	}
	fmt.Printf("Updated DHCP range to: %s - %s\n", network.DHCPStart, network.DHCPStop)

	// Get network by ID
	fmt.Println("\n=== Get Network Details ===")
	network, err = client.Networks.Get(ctx, network.ID.Int())
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}
	fmt.Printf("Network: %s\n", network.Name)
	fmt.Printf("  ID: %d\n", network.ID)
	fmt.Printf("  Enabled: %v\n", network.Enabled)
	fmt.Printf("  Network: %s\n", network.Network)
	fmt.Printf("  Type: %s\n", network.Type)
	fmt.Printf("  MTU: %d\n", network.MTU)

	// Filter networks by name
	fmt.Println("\n=== Filter Networks ===")
	filtered, err := client.Networks.List(ctx,
		vergeos.WithFilter("name eq 'sdk-example-network'"),
	)
	if err != nil {
		log.Fatalf("Failed to filter networks: %v", err)
	}
	fmt.Printf("Found %d network(s) matching filter\n", len(filtered))

	// Cleanup: Delete the network
	fmt.Println("\n=== Cleanup ===")
	if err := client.Networks.Delete(ctx, network.ID.Int()); err != nil {
		log.Fatalf("Failed to delete network: %v", err)
	}
	fmt.Println("Network deleted successfully")

	fmt.Println("\nDone!")
}
