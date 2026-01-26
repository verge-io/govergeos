// Example: Snapshot Profile Lifecycle Management
//
// This example demonstrates how to create, configure, and manage snapshot profiles
// including adding scheduling periods with different frequencies and retention policies.
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

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

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

	// List existing snapshot profiles
	fmt.Println("=== Existing Snapshot Profiles ===")
	profiles, err := client.SnapshotProfiles.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list snapshot profiles: %v", err)
	}
	for _, p := range profiles {
		fmt.Printf("  [%d] %s: %s\n", p.Key, p.Name, p.Description)
	}

	// Create a new snapshot profile
	fmt.Println("\n=== Creating Snapshot Profile ===")
	profile, err := client.SnapshotProfiles.Create(ctx, &vergeos.SnapshotProfileCreateRequest{
		Name:        "sdk-example-profile",
		Description: "Example profile created by goVergeOS",
	})
	if err != nil {
		log.Fatalf("Failed to create snapshot profile: %v", err)
	}
	fmt.Printf("Created profile: [%d] %s\n", profile.Key, profile.Name)
	profileID := int(profile.Key)

	// Add a daily snapshot period (3:00 AM, retain for 7 days)
	fmt.Println("\n=== Adding Daily Snapshot Period ===")
	dailyPeriod, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
		Profile:   profileID,
		Name:      "Daily at 3AM",
		Frequency: vergeos.FrequencyDaily,
		Hour:      ptr(3),
		Minute:    ptr(0),
		Retention: 7 * 24 * 60 * 60, // 7 days in seconds
	})
	if err != nil {
		log.Fatalf("Failed to create daily period: %v", err)
	}
	fmt.Printf("Created period: [%d] %s (retention: %d days)\n",
		dailyPeriod.Key, dailyPeriod.Name, dailyPeriod.Retention/(24*60*60))

	// Add a weekly snapshot period (Sundays at midnight, retain for 4 weeks)
	fmt.Println("\n=== Adding Weekly Snapshot Period ===")
	weeklyPeriod, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
		Profile:   profileID,
		Name:      "Weekly on Sunday",
		Frequency: vergeos.FrequencyWeekly,
		DayOfWeek: ptr(vergeos.DayOfWeekSunday),
		Hour:      ptr(0),
		Minute:    ptr(0),
		Retention: 28 * 24 * 60 * 60, // 4 weeks in seconds
	})
	if err != nil {
		log.Fatalf("Failed to create weekly period: %v", err)
	}
	fmt.Printf("Created period: [%d] %s (retention: %d days)\n",
		weeklyPeriod.Key, weeklyPeriod.Name, weeklyPeriod.Retention/(24*60*60))

	// Add a monthly snapshot period (1st of month at 1:00 AM, retain for 1 year)
	fmt.Println("\n=== Adding Monthly Snapshot Period ===")
	monthlyPeriod, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
		Profile:      profileID,
		Name:         "Monthly on 1st",
		Frequency:    vergeos.FrequencyMonthly,
		DayOfMonth:   ptr(1),
		Hour:         ptr(1),
		Minute:       ptr(0),
		Retention:    365 * 24 * 60 * 60, // 1 year in seconds
		MinSnapshots: ptr(3),             // Keep at least 3 even if older than retention
	})
	if err != nil {
		log.Fatalf("Failed to create monthly period: %v", err)
	}
	fmt.Printf("Created period: [%d] %s (retention: %d days, min_snapshots: %d)\n",
		monthlyPeriod.Key, monthlyPeriod.Name, monthlyPeriod.Retention/(24*60*60), monthlyPeriod.MinSnapshots)

	// List all periods for the profile
	fmt.Println("\n=== All Periods for Profile ===")
	periods, err := client.SnapshotProfilePeriods.ListByProfile(ctx, profileID)
	if err != nil {
		log.Fatalf("Failed to list periods: %v", err)
	}
	for _, p := range periods {
		fmt.Printf("  [%d] %s: freq=%s, hour=%d:%02d, retention=%d days\n",
			p.Key, p.Name, p.Frequency, p.Hour, p.Minute, p.Retention/(24*60*60))
	}

	// Update the daily period to enable quiescing
	fmt.Println("\n=== Updating Daily Period (enable quiesce) ===")
	dailyPeriod, err = client.SnapshotProfilePeriods.Update(ctx, int(dailyPeriod.Key), &vergeos.SnapshotProfilePeriodUpdateRequest{
		Quiesce: ptr(true),
	})
	if err != nil {
		log.Fatalf("Failed to update daily period: %v", err)
	}
	fmt.Printf("Updated period: [%d] %s (quiesce: %v)\n",
		dailyPeriod.Key, dailyPeriod.Name, dailyPeriod.Quiesce)

	// Update the profile description
	fmt.Println("\n=== Updating Profile Description ===")
	profile, err = client.SnapshotProfiles.Update(ctx, profileID, &vergeos.SnapshotProfileUpdateRequest{
		Description: ptr("Production backup schedule - Daily/Weekly/Monthly"),
	})
	if err != nil {
		log.Fatalf("Failed to update profile: %v", err)
	}
	fmt.Printf("Updated profile: [%d] %s - %s\n", profile.Key, profile.Name, profile.Description)

	// Get profile by name
	fmt.Println("\n=== Get Profile by Name ===")
	profile, err = client.SnapshotProfiles.GetByName(ctx, "sdk-example-profile")
	if err != nil {
		log.Fatalf("Failed to get profile by name: %v", err)
	}
	fmt.Printf("Found profile: [%d] %s\n", profile.Key, profile.Name)

	// Cleanup: Delete all periods first, then the profile
	fmt.Println("\n=== Cleanup: Deleting Periods ===")
	for _, p := range periods {
		if err := client.SnapshotProfilePeriods.Delete(ctx, int(p.Key)); err != nil {
			log.Printf("Warning: Failed to delete period %d: %v", p.Key, err)
		} else {
			fmt.Printf("Deleted period: [%d] %s\n", p.Key, p.Name)
		}
	}

	fmt.Println("\n=== Cleanup: Deleting Profile ===")
	if err := client.SnapshotProfiles.Delete(ctx, profileID); err != nil {
		log.Fatalf("Failed to delete profile: %v", err)
	}
	fmt.Printf("Deleted profile: [%d] %s\n", profile.Key, profile.Name)

	fmt.Println("\n=== Done! ===")
}
