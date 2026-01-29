// Example: Basic goVergeOS Usage
//
// This example demonstrates how to create a VergeOS client and perform
// basic operations like listing VMs and getting system information.
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
		vergeos.WithInsecureTLS(true), // For self-signed certificates
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get system information
	fmt.Println("=== System Information ===")
	info, err := client.System.GetInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to get system info: %v", err)
	}
	fmt.Printf("API: %s\n", info.Name)
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("Build Hash: %s\n", info.Hash)

	// Get cloud name
	cloudName, err := client.Settings.GetCloudName(ctx)
	if err != nil {
		log.Fatalf("Failed to get cloud name: %v", err)
	}
	fmt.Printf("Cloud Name: %s\n", cloudName)

	// List clusters
	fmt.Println("\n=== Clusters ===")
	clusters, err := client.Clusters.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list clusters: %v", err)
	}
	for _, cluster := range clusters {
		fmt.Printf("- %s (Key: %v, Enabled: %v)\n", cluster.Name, cluster.Key, cluster.Enabled)
	}

	// List physical nodes
	fmt.Println("\n=== Physical Nodes ===")
	nodes, err := client.Nodes.ListPhysical(ctx)
	if err != nil {
		log.Fatalf("Failed to list nodes: %v", err)
	}
	for _, node := range nodes {
		fmt.Printf("- %s (ID: %d, Cluster: %d, RAM: %dMB, Cores: %d)\n", node.Name, node.ID, node.Cluster, node.RAM, node.Cores)
	}

	// List VMs
	fmt.Println("\n=== Virtual Machines ===")
	vms, err := client.VMs.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list VMs: %v", err)
	}
	for _, vm := range vms {
		status := "stopped"
		if vm.PowerState {
			status = "running"
		}
		fmt.Printf("- %s (ID: %d, CPU: %d, RAM: %dMB, Status: %s)\n",
			vm.Name, vm.ID, vm.CPUCores, vm.RAM, status)
	}

	// List networks
	fmt.Println("\n=== Networks ===")
	networks, err := client.Networks.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list networks: %v", err)
	}
	for _, net := range networks {
		fmt.Printf("- %s (ID: %d, Network: %s, DHCP: %v)\n",
			net.Name, net.ID, net.Network, net.DHCPEnabled)
	}

	fmt.Println("\nDone!")
}
