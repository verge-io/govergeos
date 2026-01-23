// Example: Media Sources
//
// This example demonstrates how to list and work with media sources
// (ISOs, images, and other files) in VergeOS. Media sources are used
// for VM installation and boot media.
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

	// List all media sources
	fmt.Println("=== All Media Sources ===")
	media, err := client.MediaSources.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list media sources: %v", err)
	}

	if len(media) == 0 {
		fmt.Println("No media sources found.")
		fmt.Println("Upload ISOs via the VergeOS UI under Media Images.")
		return
	}

	// Group by type
	isoFiles := []vergeos.MediaSource{}
	otherFiles := []vergeos.MediaSource{}

	for _, m := range media {
		if m.Type == "iso" {
			isoFiles = append(isoFiles, m)
		} else {
			otherFiles = append(otherFiles, m)
		}
	}

	// Display ISO files
	if len(isoFiles) > 0 {
		fmt.Println("\n=== ISO Images ===")
		for _, iso := range isoFiles {
			sizeMB := iso.Size / (1024 * 1024)
			fmt.Printf("- %s\n", iso.Name)
			fmt.Printf("    ID: %d, Size: %d MB\n", iso.ID, sizeMB)
			if iso.Description != "" {
				fmt.Printf("    Description: %s\n", iso.Description)
			}
		}
	}

	// Display other files
	if len(otherFiles) > 0 {
		fmt.Println("\n=== Other Media Files ===")
		for _, f := range otherFiles {
			sizeMB := f.Size / (1024 * 1024)
			fmt.Printf("- %s (Type: %s, Size: %d MB)\n", f.Name, f.Type, sizeMB)
		}
	}

	// Filter ISO files only
	fmt.Println("\n=== Filtering: ISO Files Only ===")
	isos, err := client.MediaSources.List(ctx,
		vergeos.WithFilter("type eq 'iso'"),
	)
	if err != nil {
		log.Fatalf("Failed to filter media sources: %v", err)
	}
	fmt.Printf("Found %d ISO file(s)\n", len(isos))

	// Get details of first media source
	if len(media) > 0 {
		fmt.Println("\n=== Media Source Details ===")
		m, err := client.MediaSources.Get(ctx, media[0].ID.Int())
		if err != nil {
			log.Fatalf("Failed to get media source: %v", err)
		}
		fmt.Printf("Name: %s\n", m.Name)
		fmt.Printf("ID: %d\n", m.ID)
		fmt.Printf("Type: %s\n", m.Type)
		fmt.Printf("Size: %d bytes (%.2f GB)\n", m.Size, float64(m.Size)/(1024*1024*1024))
		if m.Path != "" {
			fmt.Printf("Path: %s\n", m.Path)
		}
	}

	// Usage tip for attaching ISO to VM
	fmt.Println("\n=== Usage Tip ===")
	fmt.Println("To attach an ISO to a VM drive:")
	fmt.Println("  drive, err := client.VMDrives.Create(ctx, vmMachineID, &vergeos.VMDriveCreateRequest{")
	fmt.Println("      Name:      \"cdrom\",")
	fmt.Println("      Interface: \"ide\",")
	fmt.Println("      Media:     \"cdrom\",")
	fmt.Println("      MediaFile:  mediaSourceID,  // ID of the ISO")
	fmt.Println("  })")

	fmt.Println("\nDone!")
}
