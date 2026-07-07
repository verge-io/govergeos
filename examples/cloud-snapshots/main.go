// Example: Cloud Snapshots - System-Wide Backup Management
//
// This example demonstrates how to:
// - List and manage cloud (system) snapshots
// - View VMs and tenants within snapshots
// - Query snapshot expiration and immutability status
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/cloud-snapshots/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/macstadium/govergeos"
)

func main() {
	// Create client from environment variables
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(os.Getenv("VERGEOS_HOST")),
		vergeos.WithCredentials(os.Getenv("VERGEOS_USERNAME"), os.Getenv("VERGEOS_PASSWORD")),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Run examples
	fmt.Println("=== Cloud Snapshots ===")
	showCloudSnapshots(ctx, client)

	fmt.Println("\n=== Snapshot Contents ===")
	showSnapshotContents(ctx, client)
}

func showCloudSnapshots(ctx context.Context, client *vergeos.Client) {
	// List all cloud snapshots
	snapshots, err := client.CloudSnapshots.List(ctx)
	if err != nil {
		log.Printf("Failed to list cloud snapshots: %v", err)
		return
	}

	fmt.Printf("Found %d cloud snapshots\n", len(snapshots))

	if len(snapshots) == 0 {
		fmt.Println("No cloud snapshots - consider creating one for backup")
		return
	}

	// Categorize snapshots
	localCount := 0
	remoteCount := 0
	immutableCount := 0
	expiringCount := 0
	now := time.Now().Unix()

	for _, s := range snapshots {
		if s.RemoteSync {
			remoteCount++
		} else {
			localCount++
		}
		if s.Immutable {
			immutableCount++
		}
		if s.Expires > 0 && s.Expires < now+86400*7 { // Expiring within 7 days
			expiringCount++
		}
	}

	fmt.Printf("\nSnapshot summary:\n")
	fmt.Printf("  Local: %d\n", localCount)
	fmt.Printf("  Remote (synced): %d\n", remoteCount)
	fmt.Printf("  Immutable (locked): %d\n", immutableCount)
	fmt.Printf("  Expiring within 7 days: %d\n", expiringCount)

	// List local snapshots only
	localSnapshots, err := client.CloudSnapshots.ListLocal(ctx)
	if err != nil {
		log.Printf("Failed to list local snapshots: %v", err)
	} else {
		fmt.Printf("\nLocal snapshots: %d\n", len(localSnapshots))
	}

	// List expiring snapshots
	expiringSnapshots, err := client.CloudSnapshots.ListExpiring(ctx)
	if err != nil {
		log.Printf("Failed to list expiring snapshots: %v", err)
	} else {
		fmt.Printf("Expiring snapshots: %d\n", len(expiringSnapshots))
	}

	// Show snapshot details (limit to first 10)
	fmt.Println("\nRecent snapshots:")
	limit := 10
	if len(snapshots) < limit {
		limit = len(snapshots)
	}
	for i := 0; i < limit; i++ {
		snap := snapshots[i]
		fmt.Printf("\n  Snapshot: %s (ID: %d)\n", snap.Name, int(snap.Key))
		if snap.Description != "" {
			fmt.Printf("    Description: %s\n", snap.Description)
		}
		fmt.Printf("    Status: %s\n", snap.Status)
		if snap.StatusInfo != "" {
			fmt.Printf("    Status Info: %s\n", snap.StatusInfo)
		}

		// Source info
		if snap.RemoteSync {
			fmt.Printf("    Source: Remote (synced from another site)\n")
		} else if snap.Provider {
			fmt.Printf("    Source: Provider tenant\n")
		} else {
			fmt.Printf("    Source: Local\n")
		}

		// Expiration
		if snap.Expires == 0 {
			fmt.Printf("    Expires: Never\n")
		} else {
			expiresTime := time.Unix(snap.Expires, 0)
			remaining := time.Until(expiresTime)
			if remaining > 0 {
				fmt.Printf("    Expires: %s (in %s)\n",
					expiresTime.Format(time.RFC3339),
					formatDuration(remaining))
			} else {
				fmt.Printf("    Expires: %s (EXPIRED)\n", expiresTime.Format(time.RFC3339))
			}
		}

		// Immutability
		if snap.Immutable {
			fmt.Printf("    Immutable: Yes (Status: %s)\n", snap.ImmutableStatus)
			if snap.ImmutableLockExpires > 0 {
				fmt.Printf("    Lock Expires: %s\n",
					time.Unix(snap.ImmutableLockExpires, 0).Format(time.RFC3339))
			}
		}

		// Profile info
		if snap.SnapshotProfile > 0 {
			fmt.Printf("    Profile ID: %d\n", int(snap.SnapshotProfile))
		}

		fmt.Printf("    Created: %s\n", time.Unix(snap.Created, 0).Format(time.RFC3339))
	}

	// Get a specific snapshot by name (if we have any)
	if len(snapshots) > 0 {
		snap, err := client.CloudSnapshots.GetByName(ctx, snapshots[0].Name)
		if err != nil {
			log.Printf("Failed to get snapshot by name: %v", err)
		} else {
			fmt.Printf("\nRetrieved snapshot by name: %s (Key: %d)\n", snap.Name, int(snap.Key))
		}
	}

	// Example: Create a snapshot (commented out to avoid side effects)
	// snapshot, err := client.CloudSnapshots.Create(ctx, &vergeos.CloudSnapshotCreateRequest{
	// 	Name:        "Manual-Backup-" + time.Now().Format("20060102-150405"),
	// 	Description: "Manual backup created via goVergeOS",
	// 	Retention:   ptr(7 * 24 * 60 * 60), // 7 days
	// })
	// if err != nil {
	// 	log.Printf("Failed to create snapshot: %v", err)
	// } else {
	// 	fmt.Printf("Created snapshot: %s (ID: %d)\n", snapshot.Name, int(snapshot.Key))
	// }
}

func showSnapshotContents(ctx context.Context, client *vergeos.Client) {
	// Get all cloud snapshots to find one to inspect
	snapshots, err := client.CloudSnapshots.List(ctx)
	if err != nil {
		log.Printf("Failed to list cloud snapshots: %v", err)
		return
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots to inspect")
		return
	}

	// Use the first snapshot
	snapshotID := int(snapshots[0].Key)
	fmt.Printf("Inspecting snapshot: %s (ID: %d)\n", snapshots[0].Name, snapshotID)

	// List VMs in the snapshot
	vms, err := client.CloudSnapshotVMs.ListBySnapshot(ctx, snapshotID)
	if err != nil {
		log.Printf("Failed to list VMs in snapshot: %v", err)
	} else {
		fmt.Printf("\nVMs in snapshot: %d\n", len(vms))
		if len(vms) > 0 {
			limit := 10
			if len(vms) < limit {
				limit = len(vms)
			}
			for i := 0; i < limit; i++ {
				vm := vms[i]
				fmt.Printf("  - %s (VM ID: %d, CPU: %d, RAM: %d MB)\n",
					vm.Name, int(vm.VM), vm.CPUCores, vm.RAM)
			}
			if len(vms) > limit {
				fmt.Printf("  ... and %d more VMs\n", len(vms)-limit)
			}
		}
	}

	// List tenants in the snapshot
	tenants, err := client.CloudSnapshotTenants.ListBySnapshot(ctx, snapshotID)
	if err != nil {
		log.Printf("Failed to list tenants in snapshot: %v", err)
	} else {
		fmt.Printf("\nTenants in snapshot: %d\n", len(tenants))
		if len(tenants) > 0 {
			for _, tenant := range tenants {
				fmt.Printf("  - %s (Tenant ID: %d)\n", tenant.Name, int(tenant.Tenant))
			}
		}
	}

	// Example: Trigger VM discovery for a snapshot (commented out)
	// if err := client.CloudSnapshots.FindVMs(ctx, snapshotID); err != nil {
	// 	log.Printf("Failed to trigger VM discovery: %v", err)
	// } else {
	// 	fmt.Printf("Triggered VM discovery for snapshot %d\n", snapshotID)
	// }
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
