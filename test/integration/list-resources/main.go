// Integration test: List all goVergeOS resources
//
// This tests goVergeOS List methods against a live VergeOS instance.
//
// Usage:
//
//	export VERGEOS_HOST=https://your-vergeos-host
//	export VERGEOS_USERNAME=your-username
//	export VERGEOS_PASSWORD=your-password
//	go run ./test/integration/list-resources/
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	vergeos "github.com/verge-io/goVergeOS"
)

func main() {
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		log.Fatal("Please set VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD environment variables")
	}

	// Create a new client
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Test Nodes
	fmt.Println("=== Nodes ===")
	nodes, err := client.Nodes.List(ctx)
	if err != nil {
		log.Printf("Nodes.List failed: %v", err)
	} else {
		fmt.Printf("Found %d nodes\n", len(nodes))
		for _, node := range nodes {
			fmt.Printf("  - %s (Cluster: %d)\n", node.Name, node.Cluster)
		}
	}

	// Test Clusters
	fmt.Println("\n=== Clusters ===")
	clusters, err := client.Clusters.List(ctx)
	if err != nil {
		log.Printf("Clusters.List failed: %v", err)
	} else {
		fmt.Printf("Found %d clusters\n", len(clusters))
		for _, cluster := range clusters {
			fmt.Printf("  - %s (ID: %d)\n", cluster.Name, cluster.Key)
		}
	}

	// Test VMs
	fmt.Println("\n=== VMs ===")
	vms, err := client.VMs.List(ctx)
	if err != nil {
		log.Printf("VMs.List failed: %v", err)
	} else {
		fmt.Printf("Found %d VMs\n", len(vms))
		for _, vm := range vms {
			fmt.Printf("  - %s (RAM: %dMB, Running: %t)\n", vm.Name, vm.RAM, vm.PowerState)
		}
	}

	// Test Networks
	fmt.Println("\n=== Networks ===")
	networks, err := client.Networks.List(ctx)
	if err != nil {
		log.Printf("Networks.List failed: %v", err)
	} else {
		fmt.Printf("Found %d networks\n", len(networks))
		for _, net := range networks {
			fmt.Printf("  - %s (Type: %s)\n", net.Name, net.Type)
		}
	}

	// Test Users
	fmt.Println("\n=== Users ===")
	users, err := client.Users.List(ctx)
	if err != nil {
		log.Printf("Users.List failed: %v", err)
	} else {
		fmt.Printf("Found %d users\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (Enabled: %t)\n", user.Name, user.Enabled)
		}
	}

	// Test Groups
	fmt.Println("\n=== Groups ===")
	groups, err := client.Groups.List(ctx)
	if err != nil {
		log.Printf("Groups.List failed: %v", err)
	} else {
		fmt.Printf("Found %d groups\n", len(groups))
		for _, group := range groups {
			fmt.Printf("  - %s\n", group.Name)
		}
	}

	// Test MediaSources
	fmt.Println("\n=== Media Sources ===")
	mediaSources, err := client.MediaSources.List(ctx)
	if err != nil {
		log.Printf("MediaSources.List failed: %v", err)
	} else {
		fmt.Printf("Found %d media sources\n", len(mediaSources))
		for _, ms := range mediaSources {
			fmt.Printf("  - %s (Type: %s)\n", ms.Name, ms.Type)
		}
	}

	// Test Settings
	fmt.Println("\n=== Settings ===")
	settings, err := client.Settings.List(ctx)
	if err != nil {
		log.Printf("Settings.List failed: %v", err)
	} else {
		fmt.Printf("Found %d settings\n", len(settings))
	}

	fmt.Println("\n=== All tests complete ===")
}
