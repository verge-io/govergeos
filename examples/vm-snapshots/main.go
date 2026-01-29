// Package main demonstrates VM snapshot and tag operations using goVergeOS.
//
// This example shows how to:
// - List and manage VM snapshots
// - Create and restore snapshots
// - Manage tag categories and tags
// - Migrate VMs between nodes
//
// Run with:
//
//	VERGEOS_HOST=https://your-host \
//	VERGEOS_USERNAME=user \
//	VERGEOS_PASSWORD=pass \
//	go run ./examples/vm-snapshots/
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

	// Demonstrate VM snapshot operations
	fmt.Println("=== VM Snapshot Operations ===")
	demonstrateSnapshots(ctx, client)

	// Demonstrate tag category and tag operations
	fmt.Println("\n=== Tag Category Operations ===")
	demonstrateTags(ctx, client)
}

func demonstrateSnapshots(ctx context.Context, client *vergeos.Client) {
	// List all VM snapshots
	snapshots, err := client.VMSnapshots.List(ctx)
	if err != nil {
		log.Printf("Failed to list snapshots: %v", err)
		return
	}
	fmt.Printf("Found %d VM snapshots\n", len(snapshots))

	// List snapshots expiring within 7 days
	expiring, err := client.VMSnapshots.ListExpiring(ctx, 7)
	if err != nil {
		log.Printf("Failed to list expiring snapshots: %v", err)
	} else {
		fmt.Printf("Found %d snapshots expiring within 7 days\n", len(expiring))
	}

	// If we have snapshots, show details of the first one
	if len(snapshots) > 0 {
		first := snapshots[0]
		fmt.Printf("\nFirst snapshot:\n")
		fmt.Printf("  ID: %d\n", int(first.Key))
		fmt.Printf("  Name: %s\n", first.Name)
		fmt.Printf("  VM: %s (ID: %d)\n", first.MachineDisplay, int(first.Machine))
		fmt.Printf("  Created: %s\n", time.Unix(first.Created, 0).Format(time.RFC3339))
		if first.Expires > 0 {
			fmt.Printf("  Expires: %s\n", time.Unix(first.Expires, 0).Format(time.RFC3339))
		} else {
			fmt.Printf("  Expires: Never\n")
		}

		// List all snapshots for this VM
		vmSnapshots, err := client.VMSnapshots.ListByVM(ctx, int(first.Machine))
		if err != nil {
			log.Printf("Failed to list VM snapshots: %v", err)
		} else {
			fmt.Printf("\nSnapshots for VM %d: %d total\n", int(first.Machine), len(vmSnapshots))
		}
	}

	// Example: Creating a snapshot (commented out to avoid side effects)
	/*
		vmID := 123 // Replace with actual VM ID
		snapshot, err := client.VMSnapshots.Create(ctx, &vergeos.VMSnapshotCreateRequest{
			Machine:     vmID,
			Name:        "my-snapshot-" + time.Now().Format("20060102-150405"),
			Description: "Example snapshot",
			ExpiresType: "date",
		})
		if err != nil {
			log.Printf("Failed to create snapshot: %v", err)
		} else {
			fmt.Printf("Created snapshot: %s (ID: %d)\n", snapshot.Name, int(snapshot.Key))

			// Set to never expire
			snapshot, _ = client.VMSnapshots.SetNeverExpires(ctx, int(snapshot.Key))
			fmt.Println("Set snapshot to never expire")

			// Restore snapshot (warning: this reverts the VM!)
			// err = client.VMSnapshots.Restore(ctx, int(snapshot.Key), &vergeos.VMSnapshotRestoreOptions{
			//     PowerOn: true,
			// })
		}
	*/
}

func demonstrateTags(ctx context.Context, client *vergeos.Client) {
	// List all tag categories
	categories, err := client.TagCategories.List(ctx)
	if err != nil {
		log.Printf("Failed to list tag categories: %v", err)
		return
	}
	fmt.Printf("Found %d tag categories\n", len(categories))

	// Show details of each category
	for _, cat := range categories {
		fmt.Printf("\nCategory: %s (ID: %d)\n", cat.Name, int(cat.Key))
		fmt.Printf("  Description: %s\n", cat.Description)
		fmt.Printf("  Single Selection: %v\n", cat.SingleTagSelection)

		// Show which resource types can be tagged
		var taggable []string
		if cat.TaggableVMs {
			taggable = append(taggable, "VMs")
		}
		if cat.TaggableVNets {
			taggable = append(taggable, "Networks")
		}
		if cat.TaggableVolumes {
			taggable = append(taggable, "Volumes")
		}
		if cat.TaggableTenants {
			taggable = append(taggable, "Tenants")
		}
		if cat.TaggableNodes {
			taggable = append(taggable, "Nodes")
		}
		if cat.TaggableClusters {
			taggable = append(taggable, "Clusters")
		}
		fmt.Printf("  Taggable: %v\n", taggable)

		// List tags in this category
		tags, err := client.Tags.ListByCategory(ctx, int(cat.Key))
		if err != nil {
			log.Printf("Failed to list tags: %v", err)
			continue
		}
		fmt.Printf("  Tags (%d):\n", len(tags))
		for _, tag := range tags {
			fmt.Printf("    - %s (ID: %d)\n", tag.Name, int(tag.Key))
		}
	}

	// List all tags
	tags, err := client.Tags.List(ctx)
	if err != nil {
		log.Printf("Failed to list tags: %v", err)
	} else {
		fmt.Printf("\nTotal tags across all categories: %d\n", len(tags))
	}

	// Example: Creating a tag category and tag (commented out to avoid side effects)
	/*
		trueVal := true
		category, err := client.TagCategories.Create(ctx, &vergeos.TagCategoryCreateRequest{
			Name:        "Environment",
			Description: "Environment classification",
			TaggableVMs: &trueVal,
			TaggableVNets: &trueVal,
		})
		if err != nil {
			log.Printf("Failed to create category: %v", err)
			return
		}
		fmt.Printf("Created category: %s (ID: %d)\n", category.Name, int(category.Key))

		// Create tags in the category
		for _, tagName := range []string{"Production", "Development", "Staging"} {
			tag, err := client.Tags.Create(ctx, &vergeos.TagCreateRequest{
				Category:    int(category.Key),
				Name:        tagName,
				Description: tagName + " environment",
			})
			if err != nil {
				log.Printf("Failed to create tag %s: %v", tagName, err)
				continue
			}
			fmt.Printf("Created tag: %s (ID: %d)\n", tag.Name, int(tag.Key))
		}

		// Assign tag to a VM
		vmID := 123 // Replace with actual VM ID
		tagID := int(tag.Key)
		member, err := client.TagMembers.Assign(ctx, tagID, fmt.Sprintf("vms/%d", vmID))
		if err != nil {
			log.Printf("Failed to assign tag: %v", err)
		} else {
			fmt.Printf("Assigned tag to VM: member ID %d\n", int(member.Key))
		}
	*/
}
