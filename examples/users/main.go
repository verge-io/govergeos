// Example: User Management
//
// This example demonstrates how to manage users and groups in VergeOS,
// including listing users, viewing group memberships, and working with
// user properties like SSH keys.
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

	// List all users
	fmt.Println("=== Users ===")
	users, err := client.Users.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}

	for _, user := range users {
		status := "enabled"
		if !user.Enabled {
			status = "disabled"
		}
		userType := user.Type
		if userType == "" {
			userType = "normal"
		}
		fmt.Printf("- %s (Key: %d, Type: %s, Status: %s)\n",
			user.Name, user.Key.Int(), userType, status)

		// Show SSH keys if configured (using helper method)
		sshKeys := user.GetSSHKeys()
		if len(sshKeys) > 0 {
			fmt.Printf("    SSH Keys: %d configured\n", len(sshKeys))
		}
	}

	// List all groups
	fmt.Println("\n=== Groups ===")
	groups, err := client.Groups.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list groups: %v", err)
	}

	for _, group := range groups {
		fmt.Printf("- %s (ID: %d, Type: %s)\n",
			group.Name, group.ID, group.Type)
	}

	// Show group memberships
	fmt.Println("\n=== Group Memberships ===")
	for _, group := range groups {
		members, err := client.Members.ListByGroup(ctx, group.ID.Int())
		if err != nil {
			log.Printf("Failed to list members for group %s: %v", group.Name, err)
			continue
		}

		if len(members) > 0 {
			fmt.Printf("%s:\n", group.Name)
			for _, member := range members {
				fmt.Printf("  - %s\n", member.Member)
			}
		}
	}

	// Demonstrate user lookup by name
	fmt.Println("\n=== User Lookup ===")
	if len(users) > 0 {
		lookupName := users[0].Name
		user, err := client.Users.GetByName(ctx, lookupName)
		if err != nil {
			log.Printf("Failed to get user by name: %v", err)
		} else {
			fmt.Printf("Found user '%s':\n", user.Name)
			fmt.Printf("  Key: %d\n", user.Key.Int())
			fmt.Printf("  Email: %s\n", user.Email)
			fmt.Printf("  Enabled: %v\n", user.Enabled)
			fmt.Printf("  Type: %s\n", user.Type)
			fmt.Printf("  Two-Factor Auth: %v\n", user.TwoFactorAuthentication)
			if user.LastLogin > 0 {
				fmt.Printf("  Last Login: %d\n", user.LastLogin)
			}
		}
	}

	// Note: User Enable/Disable actions are available
	// client.Users.Enable(ctx, userID)
	// client.Users.Disable(ctx, userID)
	fmt.Println("\n=== Available User Actions ===")
	fmt.Println("  - client.Users.Enable(ctx, userID)  - Enable a disabled user")
	fmt.Println("  - client.Users.Disable(ctx, userID) - Disable a user account")

	fmt.Println("\nDone!")
}
