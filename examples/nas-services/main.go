// Example: NAS Services Management
//
// This example demonstrates how to manage NAS services, users, volume syncs,
// and volume snapshots in VergeOS. NAS services are specialized VMs that provide
// file storage functionality including CIFS/NFS shares and volume replication.
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
	// NAS Services
	// =========================================================================
	fmt.Println("=== NAS Services ===")
	services, err := client.NASServices.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list NAS services: %v", err)
	}

	fmt.Printf("Found %d NAS service(s)\n", len(services))
	for _, svc := range services {
		fmt.Printf("  - %s (ID: %d, VM: %d)\n", svc.Name, svc.Key, svc.VM)
		fmt.Printf("    Enabled: %v, Max Imports: %d, Max Syncs: %d\n",
			svc.Enabled, svc.MaxImports, svc.MaxSyncs)
	}

	if len(services) == 0 {
		fmt.Println("\nNo NAS services found. Create a NAS VM first.")
		showOperationsReference()
		return
	}

	// Get first service for demos
	service := services[0]

	// =========================================================================
	// Get NAS Service by different methods
	// =========================================================================
	fmt.Println("\n=== Get NAS Service ===")

	// By ID
	svc, err := client.NASServices.Get(ctx, int(service.Key))
	if err != nil {
		fmt.Printf("Warning: Failed to get by ID: %v\n", err)
	} else {
		fmt.Printf("Got by ID: %s\n", svc.Name)
	}

	// By VM
	if service.VM > 0 {
		svc, err := client.NASServices.GetByVM(ctx, int(service.VM))
		if err != nil {
			fmt.Printf("Warning: Failed to get by VM: %v\n", err)
		} else {
			fmt.Printf("Got by VM %d: %s\n", service.VM, svc.Name)
		}
	}

	// By Name
	svc, err = client.NASServices.GetByName(ctx, service.Name)
	if err != nil {
		fmt.Printf("Warning: Failed to get by name: %v\n", err)
	} else {
		fmt.Printf("Got by name '%s': ID %d\n", service.Name, svc.Key)
	}

	// =========================================================================
	// NAS Service Users
	// =========================================================================
	fmt.Println("\n=== NAS Service Users ===")
	users, err := client.NASServiceUsers.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list users: %v\n", err)
	} else {
		fmt.Printf("Found %d NAS service user(s)\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (ID: %s..., Service: %d)\n",
				user.Name, truncateID(user.ID), user.Service)
			fmt.Printf("    Display: %s, Enabled: %v\n",
				user.DisplayName, user.Enabled)
		}
	}

	// List users for specific service
	fmt.Printf("\n--- Users for service %d ---\n", service.Key)
	serviceUsers, err := client.NASServiceUsers.ListByService(ctx, int(service.Key))
	if err != nil {
		fmt.Printf("Warning: Failed to list service users: %v\n", err)
	} else {
		fmt.Printf("Found %d user(s) for this service\n", len(serviceUsers))
	}

	// =========================================================================
	// Volume Syncs
	// =========================================================================
	fmt.Println("\n=== Volume Syncs ===")
	syncs, err := client.VolumeSyncs.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list volume syncs: %v\n", err)
	} else {
		fmt.Printf("Found %d volume sync(s)\n", len(syncs))
		for _, sync := range syncs {
			fmt.Printf("  - %s (ID: %s...)\n", sync.Name, truncateID(sync.ID))
			fmt.Printf("    Enabled: %v, Method: %s, Workers: %d\n",
				sync.Enabled, sync.SyncMethod, sync.Workers)
		}
	}

	// List enabled syncs
	enabledSyncs, err := client.VolumeSyncs.ListEnabled(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list enabled syncs: %v\n", err)
	} else {
		fmt.Printf("Enabled syncs: %d\n", len(enabledSyncs))
	}

	// List syncs for specific service
	serviceSyncs, err := client.VolumeSyncs.ListByService(ctx, int(service.Key))
	if err != nil {
		fmt.Printf("Warning: Failed to list service syncs: %v\n", err)
	} else {
		fmt.Printf("Syncs for service %d: %d\n", service.Key, len(serviceSyncs))
	}

	// =========================================================================
	// Volume Snapshots
	// =========================================================================
	fmt.Println("\n=== Volume Snapshots ===")
	snapshots, err := client.VolumeSnapshots.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list volume snapshots: %v\n", err)
	} else {
		fmt.Printf("Found %d volume snapshot(s)\n", len(snapshots))
		for _, snap := range snapshots {
			fmt.Printf("  - %s (ID: %d, Volume: %d)\n", snap.Name, snap.Key, snap.Volume)
			fmt.Printf("    Enabled: %v, Expires: %s\n", snap.Enabled, snap.ExpiresType)
		}
	}

	// List expiring snapshots (within 30 days)
	expiring, err := client.VolumeSnapshots.ListExpiring(ctx, 30)
	if err != nil {
		fmt.Printf("Warning: Failed to list expiring snapshots: %v\n", err)
	} else {
		fmt.Printf("Snapshots expiring within 30 days: %d\n", len(expiring))
	}

	// List manually created snapshots
	manual, err := client.VolumeSnapshots.ListManual(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list manual snapshots: %v\n", err)
	} else {
		fmt.Printf("Manually created snapshots: %d\n", len(manual))
	}

	// =========================================================================
	// Operations Reference
	// =========================================================================
	showOperationsReference()

	fmt.Println("\n=== Done ===")
}

func showOperationsReference() {
	fmt.Println("\n=== Operations Reference ===")

	fmt.Println("\nNAS Services:")
	fmt.Println("  client.NASServices.List(ctx)              - List all NAS services")
	fmt.Println("  client.NASServices.Get(ctx, id)           - Get by ID")
	fmt.Println("  client.NASServices.GetByVM(ctx, vmID)     - Get by VM ID")
	fmt.Println("  client.NASServices.GetByName(ctx, name)   - Get by name")
	fmt.Println("  client.NASServices.Create(ctx, req)       - Create service")
	fmt.Println("  client.NASServices.Update(ctx, id, req)   - Update settings")
	fmt.Println("  client.NASServices.Delete(ctx, id)        - Delete service")

	fmt.Println("\nNAS Service Users (SHA1 string IDs):")
	fmt.Println("  client.NASServiceUsers.List(ctx)                      - List all users")
	fmt.Println("  client.NASServiceUsers.ListByService(ctx, svcID)      - List by service")
	fmt.Println("  client.NASServiceUsers.Get(ctx, id)                   - Get by SHA1 ID")
	fmt.Println("  client.NASServiceUsers.GetByName(ctx, svcID, name)    - Get by name")
	fmt.Println("  client.NASServiceUsers.Create(ctx, req)               - Create user")
	fmt.Println("  client.NASServiceUsers.Update(ctx, id, req)           - Update user")
	fmt.Println("  client.NASServiceUsers.Delete(ctx, id)                - Delete user")
	fmt.Println("  client.NASServiceUsers.Enable(ctx, id)                - Enable user")
	fmt.Println("  client.NASServiceUsers.Disable(ctx, id)               - Disable user")

	fmt.Println("\nVolume Syncs (SHA1 string IDs):")
	fmt.Println("  client.VolumeSyncs.List(ctx)                       - List all syncs")
	fmt.Println("  client.VolumeSyncs.ListByService(ctx, svcID)       - List by service")
	fmt.Println("  client.VolumeSyncs.ListEnabled(ctx)                - List enabled syncs")
	fmt.Println("  client.VolumeSyncs.Get(ctx, id)                    - Get by SHA1 ID")
	fmt.Println("  client.VolumeSyncs.GetByName(ctx, svcID, name)     - Get by name")
	fmt.Println("  client.VolumeSyncs.Create(ctx, req)                - Create sync job")
	fmt.Println("  client.VolumeSyncs.Update(ctx, id, req)            - Update sync")
	fmt.Println("  client.VolumeSyncs.Delete(ctx, id)                 - Delete sync")
	fmt.Println("  client.VolumeSyncs.Start(ctx, id)                  - Start sync now")
	fmt.Println("  client.VolumeSyncs.Stop(ctx, id)                   - Stop running sync")

	fmt.Println("\nVolume Snapshots:")
	fmt.Println("  client.VolumeSnapshots.List(ctx)                      - List all snapshots")
	fmt.Println("  client.VolumeSnapshots.ListByVolume(ctx, volID)       - List by volume")
	fmt.Println("  client.VolumeSnapshots.ListExpiring(ctx, days)        - List expiring soon")
	fmt.Println("  client.VolumeSnapshots.ListManual(ctx)                - List manual snapshots")
	fmt.Println("  client.VolumeSnapshots.Get(ctx, id)                   - Get by ID")
	fmt.Println("  client.VolumeSnapshots.Create(ctx, req)               - Create snapshot")
	fmt.Println("  client.VolumeSnapshots.Update(ctx, id, req)           - Update snapshot")
	fmt.Println("  client.VolumeSnapshots.Delete(ctx, id)                - Delete snapshot")
	fmt.Println("  client.VolumeSnapshots.SetNeverExpires(ctx, id)       - Set to never expire")
	fmt.Println("  client.VolumeSnapshots.SetExpires(ctx, id, timestamp) - Set expiration")
}

// truncateID returns the first 12 characters of a SHA1 ID for display
func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
