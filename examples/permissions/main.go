// Example: Permissions Management
//
// This example demonstrates how to manage resource-level permissions in VergeOS.
// Permissions grant identities (users or groups) access to specific resources
// with granular control over list, read, create, modify, and delete operations.
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

	// =========================================================================
	// List All Permissions
	// =========================================================================
	fmt.Println("=== All Permissions ===")
	permissions, err := client.Permissions.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list permissions: %v", err)
	}

	fmt.Printf("Found %d permission(s)\n", len(permissions))
	for _, p := range permissions {
		fmt.Printf("  - ID: %d, Identity: %d, Table: %s, Row: %d\n",
			p.Key, p.Identity, p.Table, p.Row)
		fmt.Printf("    Flags: list=%v, read=%v, create=%v, modify=%v, delete=%v\n",
			p.List, p.Read, p.Create, p.Modify, p.Delete)
		if p.IdentityDisplay != "" {
			fmt.Printf("    Identity: %s, Resource: %s\n", p.IdentityDisplay, p.RowDisplay)
		}
	}

	// =========================================================================
	// List Permissions by Resource Type
	// =========================================================================
	fmt.Println("\n=== Permissions by Resource Type ===")

	// VM permissions
	vmPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableVMs)
	if err != nil {
		fmt.Printf("Warning: Failed to list VM permissions: %v\n", err)
	} else {
		fmt.Printf("VM permissions: %d\n", len(vmPerms))
	}

	// Network permissions
	netPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableNetworks)
	if err != nil {
		fmt.Printf("Warning: Failed to list network permissions: %v\n", err)
	} else {
		fmt.Printf("Network permissions: %d\n", len(netPerms))
	}

	// Volume permissions
	volPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableVolumes)
	if err != nil {
		fmt.Printf("Warning: Failed to list volume permissions: %v\n", err)
	} else {
		fmt.Printf("Volume permissions: %d\n", len(volPerms))
	}

	// Tenant permissions
	tenantPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableTenants)
	if err != nil {
		fmt.Printf("Warning: Failed to list tenant permissions: %v\n", err)
	} else {
		fmt.Printf("Tenant permissions: %d\n", len(tenantPerms))
	}

	// =========================================================================
	// List Permissions by Identity
	// =========================================================================
	if len(permissions) > 0 && permissions[0].Identity > 0 {
		fmt.Println("\n=== Permissions by Identity ===")
		identityID := int(permissions[0].Identity)
		identityPerms, err := client.Permissions.ListByIdentity(ctx, identityID)
		if err != nil {
			fmt.Printf("Warning: Failed to list identity permissions: %v\n", err)
		} else {
			fmt.Printf("Identity %d has %d permission(s)\n", identityID, len(identityPerms))
			for _, p := range identityPerms {
				flags := formatFlags(p)
				fmt.Printf("  - %s/%d: %s\n", p.Table, p.Row, flags)
			}
		}
	}

	// =========================================================================
	// Get Specific Permission
	// =========================================================================
	if len(permissions) > 0 {
		fmt.Println("\n=== Get Permission by ID ===")
		perm, err := client.Permissions.Get(ctx, int(permissions[0].Key))
		if err != nil {
			fmt.Printf("Warning: Failed to get permission: %v\n", err)
		} else {
			fmt.Printf("Permission ID %d:\n", perm.Key)
			fmt.Printf("  Identity: %d (%s)\n", perm.Identity, perm.IdentityDisplay)
			fmt.Printf("  Resource: %s/%d (%s)\n", perm.Table, perm.Row, perm.RowDisplay)
			fmt.Printf("  Flags: %s\n", formatFlags(*perm))
		}
	}

	// =========================================================================
	// Operations Reference
	// =========================================================================
	showOperationsReference()

	fmt.Println("\n=== Done ===")
}

// formatFlags returns a human-readable string of permission flags
func formatFlags(p vergeos.Permission) string {
	flags := ""
	if p.List {
		flags += "L"
	} else {
		flags += "-"
	}
	if p.Read {
		flags += "R"
	} else {
		flags += "-"
	}
	if p.Create {
		flags += "C"
	} else {
		flags += "-"
	}
	if p.Modify {
		flags += "M"
	} else {
		flags += "-"
	}
	if p.Delete {
		flags += "D"
	} else {
		flags += "-"
	}
	return flags
}

func showOperationsReference() {
	fmt.Println("\n=== Operations Reference ===")

	fmt.Println("\nListing Permissions:")
	fmt.Println("  client.Permissions.List(ctx)                           - List all permissions")
	fmt.Println("  client.Permissions.ListByIdentity(ctx, identityID)     - List by user/group")
	fmt.Println("  client.Permissions.ListByTable(ctx, table)             - List by resource type")
	fmt.Println("  client.Permissions.ListByResource(ctx, table, rowID)   - List by specific resource")

	fmt.Println("\nGetting Permissions:")
	fmt.Println("  client.Permissions.Get(ctx, id)                                    - Get by ID")
	fmt.Println("  client.Permissions.GetByIdentityAndResource(ctx, id, table, row)   - Get specific")

	fmt.Println("\nCreating/Updating Permissions:")
	fmt.Println("  client.Permissions.Create(ctx, req)        - Create with full control")
	fmt.Println("  client.Permissions.Update(ctx, id, req)    - Update permission flags")
	fmt.Println("  client.Permissions.Delete(ctx, id)         - Delete a permission")

	fmt.Println("\nConvenience Methods:")
	fmt.Println("  client.Permissions.Grant(ctx, identity, table, row, read, modify, delete)")
	fmt.Println("  client.Permissions.GrantReadOnly(ctx, identity, table, row)")
	fmt.Println("  client.Permissions.GrantFullAccess(ctx, identity, table, row)")
	fmt.Println("  client.Permissions.Revoke(ctx, identity, table, row)")

	fmt.Println("\nCommon Table Constants:")
	fmt.Println("  vergeos.PermissionTableVMs        = \"vms\"")
	fmt.Println("  vergeos.PermissionTableNetworks   = \"vnets\"")
	fmt.Println("  vergeos.PermissionTableVolumes    = \"volumes\"")
	fmt.Println("  vergeos.PermissionTableTenants    = \"tenants\"")
	fmt.Println("  vergeos.PermissionTableUsers      = \"users\"")
	fmt.Println("  vergeos.PermissionTableGroups     = \"groups\"")

	fmt.Println("\nPermission Flags:")
	fmt.Println("  L = List   - See resource in listings")
	fmt.Println("  R = Read   - View resource details (automatically enables List)")
	fmt.Println("  C = Create - Create child/related resources")
	fmt.Println("  M = Modify - Update the resource")
	fmt.Println("  D = Delete - Delete the resource")
}
