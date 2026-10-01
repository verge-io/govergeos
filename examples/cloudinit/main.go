// Example: Cloud-Init File Management
//
// This example demonstrates how to work with cloud-init files in VergeOS.
// Cloud-init files are used to configure VMs during first boot, including
// setting up users, SSH keys, packages, and custom scripts.
//
// Every file belongs to a VM. The owner reference uses the VM $key
// (vms/<VM.Key>), not the machine key. CreateForVM fills that in.
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

	vergeos "github.com/verge-io/govergeos"
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
	if err := run(ctx, client); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nDone!")
}

func run(ctx context.Context, client *vergeos.Client) error {
	// List existing cloud-init files
	fmt.Println("=== Existing Cloud-Init Files ===")
	files, err := client.CloudInitFiles.List(ctx)
	if err != nil {
		return fmt.Errorf("list cloud-init files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No cloud-init files found.")
	} else {
		for _, f := range files {
			fmt.Printf("- %s (ID: %d, Owner: %s, Size: %d bytes)\n", f.Name, f.Key, f.Owner, f.FileSize)
		}
	}

	// Cloud-init files require an owner, so create a throwaway VM first.
	fmt.Println("\n=== Creating VM ===")
	vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
		Name:        "sdk-example-cloudinit-" + time.Now().Format("20060102-150405"),
		Description: "Created by goVergeOS cloud-init example",
		CPUCores:    1,
		RAM:         512,
		OSFamily:    "linux",
	})
	if err != nil {
		return fmt.Errorf("create VM: %w", err)
	}
	fmt.Printf("Created VM: %s (ID: %d, machine: %d)\n", vm.Name, vm.Key, vm.Machine)
	defer func() {
		fmt.Println("\n=== Cleanup VM ===")
		if err := client.VMs.Delete(ctx, vm.Key.Int()); err != nil {
			log.Printf("Failed to delete VM: %v", err)
			return
		}
		fmt.Println("VM deleted successfully")
	}()

	// Create a new cloud-init file
	fmt.Println("\n=== Creating Cloud-Init File ===")

	// Example cloud-init user-data content
	cloudConfig := `#cloud-config
# goVergeOS Example Cloud-Init Configuration

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

	// CreateForVM sets owner to vms/<VM.Key>. Pass vm.Key, not vm.Machine.
	cloudInitFile, err := client.CloudInitFiles.CreateForVM(ctx, vm.Key.Int(), &vergeos.CloudInitFileCreateRequest{
		Name:     "/user-data",
		Contents: cloudConfig,
	})
	if err != nil {
		return fmt.Errorf("create cloud-init file: %w", err)
	}
	fmt.Printf("Created cloud-init file: %s (ID: %d, Owner: %s)\n", cloudInitFile.Name, cloudInitFile.Key, cloudInitFile.Owner)

	// List files for this VM
	fmt.Println("\n=== Cloud-Init Files For VM ===")
	vmFiles, err := client.CloudInitFiles.ListByVM(ctx, vm.Key.Int())
	if err != nil {
		return fmt.Errorf("list cloud-init files for VM: %w", err)
	}
	for _, f := range vmFiles {
		fmt.Printf("- %s (ID: %d, Owner: %s)\n", f.Name, f.Key, f.Owner)
	}

	// Get the cloud-init file details
	fmt.Println("\n=== Cloud-Init File Details ===")
	cloudInitFile, err = client.CloudInitFiles.Get(ctx, cloudInitFile.Key.Int())
	if err != nil {
		return fmt.Errorf("get cloud-init file: %w", err)
	}
	fmt.Printf("Name: %s\n", cloudInitFile.Name)
	fmt.Printf("Owner: %s\n", cloudInitFile.Owner)
	fmt.Printf("Size: %d bytes\n", cloudInitFile.FileSize)
	// Get leaves Contents empty. The body is GET ?download=1.
	contents, err := client.CloudInitFiles.GetContents(ctx, cloudInitFile.Key.Int())
	if err != nil {
		return fmt.Errorf("read cloud-init file: %w", err)
	}
	fmt.Printf("Contents Preview:\n")
	// Show first few lines of contents
	lines := 0
	for i, c := range contents {
		fmt.Print(string(c))
		if c == '\n' {
			lines++
			if lines >= 5 {
				fmt.Printf("  ... (%d more bytes)\n", len(contents)-i-1)
				break
			}
		}
	}

	// Update the cloud-init file. Owner cannot be changed.
	fmt.Println("\n=== Updating Cloud-Init File ===")
	newName := "/user-data-updated"
	cloudInitFile, err = client.CloudInitFiles.Update(ctx, cloudInitFile.Key.Int(), &vergeos.CloudInitFileUpdateRequest{
		Name: &newName,
	})
	if err != nil {
		return fmt.Errorf("update cloud-init file: %w", err)
	}
	fmt.Printf("Updated name to: %s\n", cloudInitFile.Name)

	// Cleanup: Delete the cloud-init file, then the VM (deferred above).
	fmt.Println("\n=== Cleanup ===")
	if err := client.CloudInitFiles.Delete(ctx, cloudInitFile.Key.Int()); err != nil {
		return fmt.Errorf("delete cloud-init file: %w", err)
	}
	fmt.Println("Cloud-init file deleted successfully")
	return nil
}
