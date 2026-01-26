// Example: Multi-Tenant Management
//
// This example demonstrates how to manage tenants (virtual data centers) in VergeOS.
// Tenants are isolated environments that can contain their own VMs, networks, and storage.
// This is essential for MSPs and enterprises with multi-tenant requirements.
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

	vergeos "github.com/verge-io/goVergeOS"
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
	// List existing tenants
	// =========================================================================
	fmt.Println("=== Listing Existing Tenants ===")
	tenants, err := client.Tenants.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list tenants: %v", err)
	}
	fmt.Printf("Found %d existing tenant(s)\n", len(tenants))
	for _, t := range tenants {
		fmt.Printf("  - %s (ID: %d, Isolated: %v)\n", t.Name, t.Key, t.Isolate)
	}

	// =========================================================================
	// Create a new tenant
	// =========================================================================
	fmt.Println("\n=== Creating New Tenant ===")
	tenantName := fmt.Sprintf("sdk-example-tenant-%d", time.Now().Unix())
	tenant, err := client.Tenants.Create(ctx, &vergeos.TenantCreateRequest{
		Name:        tenantName,
		Description: "Created by goVergeOS example",
		Password:    "SecureP@ssw0rd!", // Admin password for the tenant
	})
	if err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}
	fmt.Printf("Created tenant: %s (ID: %d)\n", tenant.Name, tenant.Key)
	fmt.Printf("  UUID: %s\n", tenant.UUID)
	fmt.Printf("  Auto-created VNet: %d\n", tenant.VNet)

	// Ensure cleanup on exit
	defer func() {
		fmt.Println("\n=== Cleaning Up ===")
		// Power off tenant first (required before deletion)
		fmt.Println("Powering off tenant...")
		if err := client.Tenants.PowerOff(ctx, int(tenant.Key)); err != nil {
			fmt.Printf("Warning: Failed to power off tenant: %v\n", err)
		}
		// Wait for power off
		time.Sleep(2 * time.Second)

		// Delete the tenant
		fmt.Println("Deleting tenant...")
		if err := client.Tenants.Delete(ctx, int(tenant.Key)); err != nil {
			fmt.Printf("Warning: Failed to delete tenant: %v\n", err)
		} else {
			fmt.Println("Tenant deleted successfully")
		}
	}()

	// =========================================================================
	// Add a virtual node to the tenant
	// =========================================================================
	fmt.Println("\n=== Adding Tenant Node ===")
	node, err := client.TenantNodes.Create(ctx, &vergeos.TenantNodeCreateRequest{
		Tenant:   int(tenant.Key),
		Name:     "node1",
		CPUCores: 2,
		RAM:      4096, // 4GB minimum recommended
	})
	if err != nil {
		log.Fatalf("Failed to create tenant node: %v", err)
	}
	fmt.Printf("Created tenant node: %s (ID: %d)\n", node.Name, node.Key)
	fmt.Printf("  CPU Cores: %d\n", node.CPUCores)
	fmt.Printf("  RAM: %d MB\n", node.RAM)

	// =========================================================================
	// List storage tiers and allocate storage
	// =========================================================================
	fmt.Println("\n=== Allocating Storage ===")

	// First, check existing storage allocations for the tenant
	existingStorage, err := client.TenantStorage.ListByTenant(ctx, int(tenant.Key))
	if err != nil {
		log.Fatalf("Failed to list tenant storage: %v", err)
	}

	if len(existingStorage) > 0 {
		fmt.Printf("Tenant already has %d storage allocation(s):\n", len(existingStorage))
		for _, s := range existingStorage {
			fmt.Printf("  - Tier %d: %d GB provisioned, %d GB used\n",
				s.Tier, s.Provisioned/(1024*1024*1024), s.Used/(1024*1024*1024))
		}
	} else {
		fmt.Println("No storage allocated yet (storage allocation requires tier configuration)")
	}

	// =========================================================================
	// Power on the tenant
	// =========================================================================
	fmt.Println("\n=== Powering On Tenant ===")
	if err := client.Tenants.PowerOn(ctx, int(tenant.Key)); err != nil {
		log.Fatalf("Failed to power on tenant: %v", err)
	}
	fmt.Println("Tenant power on initiated")

	// Wait for tenant to start
	fmt.Println("Waiting for tenant to start...")
	time.Sleep(3 * time.Second)

	// =========================================================================
	// Update tenant settings
	// =========================================================================
	fmt.Println("\n=== Updating Tenant Settings ===")
	newDesc := "Updated by goVergeOS example - multi-tenant demo"
	allowBranding := true
	tenant, err = client.Tenants.Update(ctx, int(tenant.Key), &vergeos.TenantUpdateRequest{
		Description:   &newDesc,
		AllowBranding: &allowBranding,
	})
	if err != nil {
		log.Fatalf("Failed to update tenant: %v", err)
	}
	fmt.Printf("Updated tenant description: %s\n", tenant.Description)
	fmt.Printf("Allow branding: %v\n", tenant.AllowBranding)

	// =========================================================================
	// Network isolation
	// =========================================================================
	fmt.Println("\n=== Network Isolation Demo ===")
	fmt.Printf("Current isolation status: %v\n", tenant.Isolate)

	// Enable network isolation
	fmt.Println("Enabling network isolation...")
	if err := client.Tenants.IsolateOn(ctx, int(tenant.Key)); err != nil {
		fmt.Printf("Warning: Failed to enable isolation: %v\n", err)
	} else {
		// Verify isolation status
		tenant, _ = client.Tenants.Get(ctx, int(tenant.Key))
		fmt.Printf("Isolation enabled: %v\n", tenant.Isolate)
	}

	// Disable network isolation
	fmt.Println("Disabling network isolation...")
	if err := client.Tenants.IsolateOff(ctx, int(tenant.Key)); err != nil {
		fmt.Printf("Warning: Failed to disable isolation: %v\n", err)
	} else {
		tenant, _ = client.Tenants.Get(ctx, int(tenant.Key))
		fmt.Printf("Isolation disabled: %v\n", !tenant.Isolate)
	}

	// =========================================================================
	// List tenant nodes
	// =========================================================================
	fmt.Println("\n=== Tenant Nodes ===")
	nodes, err := client.TenantNodes.ListByTenant(ctx, int(tenant.Key))
	if err != nil {
		log.Fatalf("Failed to list tenant nodes: %v", err)
	}
	fmt.Printf("Tenant has %d node(s):\n", len(nodes))
	for _, n := range nodes {
		fmt.Printf("  - %s: %d cores, %d MB RAM (ID: %d)\n",
			n.Name, n.CPUCores, n.RAM, n.Key)
	}

	// =========================================================================
	// Update tenant node resources
	// =========================================================================
	fmt.Println("\n=== Scaling Tenant Node ===")
	newCores := 4
	newRAM := 8192 // 8GB
	node, err = client.TenantNodes.Update(ctx, int(node.Key), &vergeos.TenantNodeUpdateRequest{
		CPUCores: &newCores,
		RAM:      &newRAM,
	})
	if err != nil {
		log.Fatalf("Failed to update tenant node: %v", err)
	}
	fmt.Printf("Scaled node %s to %d cores, %d MB RAM\n", node.Name, node.CPUCores, node.RAM)

	// =========================================================================
	// Get tenant by name (useful for automation)
	// =========================================================================
	fmt.Println("\n=== Lookup by Name ===")
	foundTenant, err := client.Tenants.GetByName(ctx, tenantName)
	if err != nil {
		log.Fatalf("Failed to get tenant by name: %v", err)
	}
	fmt.Printf("Found tenant by name: %s (ID: %d)\n", foundTenant.Name, foundTenant.Key)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The tenant will now be cleaned up...")
}
