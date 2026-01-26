// Example: API Key Authentication
//
// This example demonstrates using API key authentication instead of username/password.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_API_KEY=your-api-key go run ./examples/apikey-auth/
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
	apiKey := os.Getenv("VERGEOS_API_KEY")

	if host == "" || apiKey == "" {
		log.Fatal("VERGEOS_HOST and VERGEOS_API_KEY environment variables are required")
	}

	// Create client using API key authentication
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithAPIKey(apiKey),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test the connection by listing VMs
	fmt.Println("=== Testing API Key Authentication ===")
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("API Key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])

	vms, err := client.VMs.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list VMs: %v", err)
	}

	fmt.Printf("\nSuccess! Found %d VMs\n", len(vms))
	for i, vm := range vms {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(vms)-5)
			break
		}
		fmt.Printf("  - %s (ID: %d, Enabled: %v)\n", vm.Name, int(vm.ID), vm.Enabled)
	}

	// Also test system info
	fmt.Println("\n=== System Info ===")
	info, err := client.System.GetInfo(ctx)
	if err != nil {
		log.Printf("Failed to get system info: %v", err)
	} else {
		fmt.Printf("System: %s\n", info.Name)
		fmt.Printf("Version: %s\n", info.Version)
	}
}
