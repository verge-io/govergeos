// Example: VM Lifecycle Management
//
// This example demonstrates how to create, configure, and manage VMs
// including adding drives and NICs, power operations, and cleanup.
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

	// Create a new VM
	fmt.Println("Creating VM...")
	vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
		Name:        "sdk-example-vm",
		Description: "Created by VergeOS Go SDK example",
		CPUCores:    2,
		RAM:         2048, // 2GB
		OSFamily:    "linux",
	})
	if err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}
	fmt.Printf("Created VM: %s (ID: %d)\n", vm.Name, vm.ID)

	// Add a virtual disk
	fmt.Println("Adding virtual disk...")
	drive, err := client.VMDrives.Create(ctx, vm.Machine, &vergeos.VMDriveCreateRequest{
		Name:      "disk0",
		Interface: "virtio",
		Media:     "disk",
		SizeGB:    20, // 20GB disk
	})
	if err != nil {
		log.Fatalf("Failed to create drive: %v", err)
	}
	fmt.Printf("Created drive: %s (ID: %d, Size: %dGB)\n", drive.Name, drive.ID, drive.SizeGB)

	// List available networks to attach
	networks, err := client.Networks.List(ctx, vergeos.WithLimit(1))
	if err != nil {
		log.Fatalf("Failed to list networks: %v", err)
	}

	if len(networks) > 0 {
		// Add a network interface
		fmt.Println("Adding network interface...")
		nic, err := client.VMNICs.Create(ctx, vm.Machine, &vergeos.VMNICCreateRequest{
			Name: "nic0",
			VNET: networks[0].ID.Int(),
		})
		if err != nil {
			log.Fatalf("Failed to create NIC: %v", err)
		}
		fmt.Printf("Created NIC: %s (ID: %d, MAC: %s, Network: %s)\n",
			nic.Name, nic.ID, nic.MAC, networks[0].Name)
	}

	// Power on the VM
	fmt.Println("Powering on VM...")
	if err := client.VMs.PowerOn(ctx, vm.ID.Int()); err != nil {
		log.Fatalf("Failed to power on VM: %v", err)
	}
	fmt.Println("VM is now running")

	// Get updated VM status
	vm, err = client.VMs.Get(ctx, vm.ID.Int())
	if err != nil {
		log.Fatalf("Failed to get VM: %v", err)
	}
	fmt.Printf("VM Status: Running=%v\n", vm.PowerState)

	// List all drives for the VM
	fmt.Println("\nVM Drives:")
	drives, err := client.VMDrives.List(ctx, vm.Machine)
	if err != nil {
		log.Fatalf("Failed to list drives: %v", err)
	}
	for _, d := range drives {
		fmt.Printf("  - %s: %dGB (%s)\n", d.Name, d.SizeGB, d.Interface)
	}

	// List all NICs for the VM
	fmt.Println("\nVM NICs:")
	nics, err := client.VMNICs.List(ctx, vm.Machine)
	if err != nil {
		log.Fatalf("Failed to list NICs: %v", err)
	}
	for _, n := range nics {
		fmt.Printf("  - %s: %s (VNET: %d)\n", n.Name, n.MAC, n.VNET)
	}

	// Power off the VM
	fmt.Println("\nPowering off VM...")
	if err := client.VMs.PowerOff(ctx, vm.ID.Int()); err != nil {
		log.Fatalf("Failed to power off VM: %v", err)
	}
	fmt.Println("VM is now stopped")

	// Cleanup: Delete the VM
	fmt.Println("\nCleaning up - deleting VM...")
	if err := client.VMs.Delete(ctx, vm.ID.Int()); err != nil {
		log.Fatalf("Failed to delete VM: %v", err)
	}
	fmt.Println("VM deleted successfully")

	fmt.Println("\nDone!")
}
