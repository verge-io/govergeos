// Example: Files
//
// This example demonstrates how to list, upload, download, and delete files
// (ISOs, images, and other files) in VergeOS. Files are used for VM installation
// and boot media.
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

	vergeos "github.com/verge-io/goVergeOS"
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

	// List all files
	fmt.Println("=== All Files ===")
	files, err := client.Files.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No files found.")
		fmt.Println("Upload ISOs via the VergeOS UI under Media Images.")
		return
	}

	// Group by type
	isoFiles := []vergeos.File{}
	otherFiles := []vergeos.File{}

	for _, f := range files {
		if f.Type == "iso" {
			isoFiles = append(isoFiles, f)
		} else {
			otherFiles = append(otherFiles, f)
		}
	}

	// Display ISO files
	if len(isoFiles) > 0 {
		fmt.Println("\n=== ISO Images ===")
		for _, iso := range isoFiles {
			sizeMB := iso.Filesize / (1024 * 1024)
			fmt.Printf("- %s\n", iso.Name)
			fmt.Printf("    ID: %d, Size: %d MB\n", iso.ID, sizeMB)
			if iso.Description != "" {
				fmt.Printf("    Description: %s\n", iso.Description)
			}
		}
	}

	// Display other files
	if len(otherFiles) > 0 {
		fmt.Println("\n=== Other Files ===")
		for _, f := range otherFiles {
			sizeMB := f.Filesize / (1024 * 1024)
			fmt.Printf("- %s (Type: %s, Size: %d MB)\n", f.Name, f.Type, sizeMB)
		}
	}

	// Filter ISO files only
	fmt.Println("\n=== Filtering: ISO Files Only ===")
	isos, err := client.Files.List(ctx,
		vergeos.WithFilter("type eq 'iso'"),
	)
	if err != nil {
		log.Fatalf("Failed to filter files: %v", err)
	}
	fmt.Printf("Found %d ISO file(s)\n", len(isos))

	// Get details of first file
	if len(files) > 0 {
		fmt.Println("\n=== File Details ===")
		f, err := client.Files.Get(ctx, files[0].ID.Int())
		if err != nil {
			log.Fatalf("Failed to get file: %v", err)
		}
		fmt.Printf("Name: %s\n", f.Name)
		fmt.Printf("ID: %d\n", f.ID)
		fmt.Printf("Type: %s\n", f.Type)
		fmt.Printf("Filesize: %d bytes (%.2f GB)\n", f.Filesize, float64(f.Filesize)/(1024*1024*1024))
		fmt.Printf("Allocated: %d bytes\n", f.AllocatedBytes)
		fmt.Printf("Used: %d bytes\n", f.UsedBytes)
		if f.PreferredTier != "" {
			fmt.Printf("Preferred Tier: %s\n", f.PreferredTier)
		}
		if f.Creator != "" {
			fmt.Printf("Creator: %s\n", f.Creator)
		}
	}

	// Usage tip for attaching ISO to VM
	fmt.Println("\n=== Usage Tip ===")
	fmt.Println("To attach an ISO to a VM drive:")
	fmt.Println("  drive, err := client.VMDrives.Create(ctx, vmMachineID, &vergeos.VMDriveCreateRequest{")
	fmt.Println("      Name:      \"cdrom\",")
	fmt.Println("      Interface: \"ide\",")
	fmt.Println("      Media:     \"cdrom\",")
	fmt.Println("      File:      fileID,  // ID of the ISO")
	fmt.Println("  })")

	// Upload example (commented out - uncomment to test)
	fmt.Println("\n=== Upload Example (Code) ===")
	fmt.Println("// Upload a local file:")
	fmt.Println("// file, err := client.Files.UploadFromFile(ctx, \"/path/to/local.iso\", &vergeos.FileCreateRequest{")
	fmt.Println("//     Name:        \"my-uploaded.iso\",")
	fmt.Println("//     Description: \"Uploaded via SDK\",")
	fmt.Println("// })")
	fmt.Println("//")
	fmt.Println("// Or create entry first, then upload:")
	fmt.Println("// entry, _ := client.Files.Create(ctx, &vergeos.FileCreateRequest{")
	fmt.Println("//     Name:           \"my-file.iso\",")
	fmt.Println("//     AllocatedBytes: \"1073741824\", // 1GB")
	fmt.Println("// })")
	fmt.Println("// file, _ := client.Files.Upload(ctx, entry.ID.Int(), reader, size)")

	// Download example (commented out - uncomment to test)
	fmt.Println("\n=== Download Example (Code) ===")
	fmt.Println("// Download to file:")
	fmt.Println("// path, err := client.Files.DownloadToFile(ctx, fileID, \"/local/path/\")")
	fmt.Println("//")
	fmt.Println("// Or get reader for streaming:")
	fmt.Println("// reader, file, err := client.Files.Download(ctx, fileID)")
	fmt.Println("// defer reader.Close()")
	fmt.Println("// io.Copy(destination, reader)")

	fmt.Println("\nDone!")
}
