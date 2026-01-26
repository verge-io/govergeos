// Example: NAS Volume Management
//
// This example demonstrates how to manage NAS volumes in VergeOS.
// Volumes provide file-level storage for CIFS/NFS shares, VM disks, and general storage.
//
// Note: This example requires an existing NAS service. Volumes use SHA1 hash strings
// as their IDs, unlike other resources that use integers.
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

	vergeos "github.com/verge-io/vergeos-go-sdk"
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
	// List existing volumes
	// =========================================================================
	fmt.Println("=== Listing NAS Volumes ===")
	volumes, err := client.Volumes.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list volumes: %v", err)
	}

	fmt.Printf("Found %d volume(s)\n", len(volumes))
	for _, v := range volumes {
		sizeGB := v.MaxSize / (1024 * 1024 * 1024)
		fmt.Printf("  - %s (ID: %s...)\n", v.Name, truncateID(v.Key))
		fmt.Printf("    Type: %s, Size: %d GB, Enabled: %v\n", v.FSType, sizeGB, v.Enabled)
		if v.SnapshotProfile > 0 {
			fmt.Printf("    Snapshot Profile: %d\n", v.SnapshotProfile)
		}
	}

	if len(volumes) == 0 {
		fmt.Println("\nNo volumes found. This example requires an existing NAS service.")
		fmt.Println("Skipping volume operations demo.")
		showVolumeSharesExample(ctx, client)
		return
	}

	// =========================================================================
	// Get a specific volume by ID
	// =========================================================================
	fmt.Println("\n=== Getting Volume Details ===")
	vol := volumes[0]
	volume, err := client.Volumes.Get(ctx, vol.Key)
	if err != nil {
		log.Fatalf("Failed to get volume: %v", err)
	}
	fmt.Printf("Volume: %s\n", volume.Name)
	fmt.Printf("  ID: %s\n", volume.ID)
	fmt.Printf("  Service: %d\n", volume.Service)
	fmt.Printf("  Filesystem: %s\n", volume.FSType)
	fmt.Printf("  Max Size: %d bytes\n", volume.MaxSize)
	fmt.Printf("  Preferred Tier: %s\n", volume.PreferredTier)
	fmt.Printf("  Owner: %s:%s\n", volume.OwnerUser, volume.OwnerGroup)
	fmt.Printf("  Encrypted: %v\n", volume.Encrypt)
	fmt.Printf("  TRIM/Discard: %v\n", volume.Discard)
	fmt.Printf("  Read-Only: %v\n", volume.ReadOnly)

	// =========================================================================
	// Get volume by name (within a service)
	// =========================================================================
	if volume.Service > 0 {
		fmt.Println("\n=== Lookup Volume by Name ===")
		foundVol, err := client.Volumes.GetByName(ctx, int(volume.Service), volume.Name)
		if err != nil {
			fmt.Printf("Warning: Failed to get volume by name: %v\n", err)
		} else {
			fmt.Printf("Found volume '%s' by name (ID: %s...)\n", foundVol.Name, truncateID(foundVol.Key))
		}
	}

	// =========================================================================
	// List volumes by service
	// =========================================================================
	if volume.Service > 0 {
		fmt.Println("\n=== Volumes in Same Service ===")
		serviceVols, err := client.Volumes.ListByService(ctx, int(volume.Service))
		if err != nil {
			fmt.Printf("Warning: Failed to list service volumes: %v\n", err)
		} else {
			fmt.Printf("Service %d has %d volume(s):\n", volume.Service, len(serviceVols))
			for _, sv := range serviceVols {
				fmt.Printf("  - %s (%s)\n", sv.Name, sv.FSType)
			}
		}
	}

	// =========================================================================
	// Show CIFS/NFS shares for volumes
	// =========================================================================
	showVolumeSharesExample(ctx, client)

	fmt.Println("\n=== Volume Operations Reference ===")
	fmt.Println("The SDK supports the following volume operations:")
	fmt.Println("  - client.Volumes.List(ctx)                    - List all volumes")
	fmt.Println("  - client.Volumes.ListByService(ctx, svcID)    - List volumes by NAS service")
	fmt.Println("  - client.Volumes.Get(ctx, id)                 - Get volume by SHA1 ID")
	fmt.Println("  - client.Volumes.GetByName(ctx, svcID, name)  - Get volume by name")
	fmt.Println("  - client.Volumes.Create(ctx, req)             - Create a new volume")
	fmt.Println("  - client.Volumes.Update(ctx, id, req)         - Update volume settings")
	fmt.Println("  - client.Volumes.Delete(ctx, id)              - Delete a volume")
	fmt.Println("  - client.Volumes.Enable(ctx, id)              - Enable a volume")
	fmt.Println("  - client.Volumes.Disable(ctx, id)             - Disable a volume")
	fmt.Println("  - client.Volumes.Reset(ctx, id)               - Reset a volume")

	fmt.Println("\n=== Done ===")
}

// showVolumeSharesExample demonstrates CIFS and NFS share listing
func showVolumeSharesExample(ctx context.Context, client *vergeos.Client) {
	fmt.Println("\n=== CIFS Shares ===")
	cifsShares, err := client.VolumeCIFSShares.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list CIFS shares: %v\n", err)
	} else {
		fmt.Printf("Found %d CIFS share(s)\n", len(cifsShares))
		for _, share := range cifsShares {
			fmt.Printf("  - %s (Path: %s, Read-Only: %v)\n", share.Name, share.SharePath, share.ReadOnly)
		}
	}

	fmt.Println("\n=== NFS Shares ===")
	nfsShares, err := client.VolumeNFSShares.List(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list NFS shares: %v\n", err)
	} else {
		fmt.Printf("Found %d NFS share(s)\n", len(nfsShares))
		for _, share := range nfsShares {
			fmt.Printf("  - %s (Path: %s, Squash: %s)\n", share.Name, share.SharePath, share.Squash)
		}
	}
}

// truncateID returns the first 12 characters of a SHA1 ID for display
func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
