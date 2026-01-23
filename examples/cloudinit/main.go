// Example: Cloud-Init File Management
//
// This example demonstrates how to work with cloud-init files in VergeOS.
// Cloud-init files are used to configure VMs during first boot, including
// setting up users, SSH keys, packages, and custom scripts.
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

	// List existing cloud-init files
	fmt.Println("=== Existing Cloud-Init Files ===")
	files, err := client.CloudInitFiles.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list cloud-init files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No cloud-init files found.")
	} else {
		for _, f := range files {
			fmt.Printf("- %s (ID: %d, Size: %d bytes)\n", f.Name, f.ID, f.FileSize)
		}
	}

	// Create a new cloud-init file
	fmt.Println("\n=== Creating Cloud-Init File ===")

	// Example cloud-init user-data content
	cloudConfig := `#cloud-config
# SDK Example Cloud-Init Configuration

# Set hostname
hostname: sdk-example-host

# Create a user
users:
  - name: deploy
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... example@key

# Install packages
packages:
  - curl
  - vim
  - htop

# Run commands on first boot
runcmd:
  - echo "Cloud-init completed at $(date)" >> /var/log/cloud-init-complete.log

# Final message
final_message: "System ready after $UPTIME seconds"
`

	cloudInitFile, err := client.CloudInitFiles.Create(ctx, &vergeos.CloudInitFileCreateRequest{
		Name:     "sdk-example-cloudinit",
		Contents: cloudConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create cloud-init file: %v", err)
	}
	fmt.Printf("Created cloud-init file: %s (ID: %d)\n", cloudInitFile.Name, cloudInitFile.ID)

	// Get the cloud-init file details
	fmt.Println("\n=== Cloud-Init File Details ===")
	cloudInitFile, err = client.CloudInitFiles.Get(ctx, cloudInitFile.ID.Int())
	if err != nil {
		log.Fatalf("Failed to get cloud-init file: %v", err)
	}
	fmt.Printf("Name: %s\n", cloudInitFile.Name)
	fmt.Printf("Size: %d bytes\n", cloudInitFile.FileSize)
	fmt.Printf("Contents Preview:\n")
	// Show first few lines of contents
	lines := 0
	for i, c := range cloudInitFile.Contents {
		fmt.Print(string(c))
		if c == '\n' {
			lines++
			if lines >= 5 {
				fmt.Printf("  ... (%d more bytes)\n", len(cloudInitFile.Contents)-i-1)
				break
			}
		}
	}

	// Update the cloud-init file
	fmt.Println("\n=== Updating Cloud-Init File ===")
	newName := "sdk-example-cloudinit-updated"
	cloudInitFile, err = client.CloudInitFiles.Update(ctx, cloudInitFile.ID.Int(), &vergeos.CloudInitFileUpdateRequest{
		Name: &newName,
	})
	if err != nil {
		log.Fatalf("Failed to update cloud-init file: %v", err)
	}
	fmt.Printf("Updated name to: %s\n", cloudInitFile.Name)

	// Cleanup: Delete the cloud-init file
	fmt.Println("\n=== Cleanup ===")
	if err := client.CloudInitFiles.Delete(ctx, cloudInitFile.ID.Int()); err != nil {
		log.Fatalf("Failed to delete cloud-init file: %v", err)
	}
	fmt.Println("Cloud-init file deleted successfully")

	// Usage tip
	fmt.Println("\n=== Usage Tip ===")
	fmt.Println("To use cloud-init with a VM, set the CloudInit field when creating:")
	fmt.Println("  client.VMs.Create(ctx, &vergeos.VMCreateRequest{")
	fmt.Println("      Name:      \"my-vm\",")
	fmt.Println("      CloudInit: cloudInitID,")
	fmt.Println("      // ... other fields")
	fmt.Println("  })")

	fmt.Println("\nDone!")
}
