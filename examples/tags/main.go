// Example: Tags and Tag Members
//
// This example demonstrates how to work with tags and tag assignments
// in VergeOS. Tags allow you to categorize and organize resources like
// VMs, networks, and other objects.
//
// Note: Tags require VergeOS v26 or later.
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
		vergeos.WithInsecureTLS(true), // For self-signed certificates
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// List all tags
	fmt.Println("=== All Tags ===")
	tags, err := client.Tags.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list tags: %v", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found. Tags require VergeOS v26+.")
		fmt.Println("Create tags in the VergeOS UI under System > Tags.")
		return
	}

	// Group tags by category for display
	categoryTags := make(map[int][]vergeos.Tag)
	for _, tag := range tags {
		catID := tag.Category.Int()
		categoryTags[catID] = append(categoryTags[catID], tag)
	}

	for catID, catTags := range categoryTags {
		fmt.Printf("\nCategory %d:\n", catID)
		for _, tag := range catTags {
			fmt.Printf("  - %s (ID: %d) - %s\n", tag.Name, tag.Key.Int(), tag.Description)
		}
	}

	// List all tag assignments
	fmt.Println("\n=== Tag Assignments ===")
	members, err := client.TagMembers.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list tag members: %v", err)
	}

	if len(members) == 0 {
		fmt.Println("No tag assignments found.")
	} else {
		for _, m := range members {
			// Find the tag name
			tagName := fmt.Sprintf("Tag %d", m.Tag.Int())
			for _, t := range tags {
				if t.Key.Int() == m.Tag.Int() {
					tagName = t.Name
					break
				}
			}
			fmt.Printf("  - %s -> %s\n", tagName, m.Member)
		}
	}

	// Demonstrate tag lookup by name
	fmt.Println("\n=== Tag Lookup Example ===")
	if len(tags) > 0 {
		lookupName := tags[0].Name
		tag, err := client.Tags.GetByName(ctx, lookupName)
		if err != nil {
			log.Printf("Failed to get tag by name: %v", err)
		} else {
			fmt.Printf("Found tag '%s' (ID: %d, Category: %d)\n",
				tag.Name, tag.Key.Int(), tag.Category.Int())
		}
	}

	// Demonstrate filtering tags for a specific resource
	fmt.Println("\n=== Tags for VMs Example ===")
	// Get first VM to check its tags
	vms, err := client.VMs.List(ctx, vergeos.WithLimit(1))
	if err != nil {
		log.Printf("Failed to list VMs: %v", err)
	} else if len(vms) > 0 {
		vmID := vms[0].ID
		vmMember := fmt.Sprintf("vms/%d", vmID)

		vmTags, err := client.TagMembers.ListByMember(ctx, vmMember)
		if err != nil {
			log.Printf("Failed to get tags for VM: %v", err)
		} else if len(vmTags) == 0 {
			fmt.Printf("VM '%s' (ID: %d) has no tags assigned.\n", vms[0].Name, vmID)
		} else {
			fmt.Printf("VM '%s' (ID: %d) has %d tag(s):\n", vms[0].Name, vmID, len(vmTags))
			for _, tm := range vmTags {
				tagName := fmt.Sprintf("Tag %d", tm.Tag.Int())
				for _, t := range tags {
					if t.Key.Int() == tm.Tag.Int() {
						tagName = t.Name
						break
					}
				}
				fmt.Printf("  - %s\n", tagName)
			}
		}
	}

	fmt.Println("\nDone!")
}
