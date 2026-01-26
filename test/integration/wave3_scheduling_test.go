//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave3Scheduling tests the Wave 3 scheduling services (SnapshotProfiles, SnapshotProfilePeriods)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave3
func TestWave3Scheduling(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("SnapshotProfiles", func(t *testing.T) {
		testSnapshotProfiles(t, ctx, client)
	})

	t.Run("SnapshotProfilePeriods", func(t *testing.T) {
		testSnapshotProfilePeriods(t, ctx, client)
	})
}

func testSnapshotProfiles(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing SnapshotProfiles service...")

	// List all snapshot profiles
	profiles, err := client.SnapshotProfiles.List(ctx)
	if err != nil {
		t.Fatalf("SnapshotProfiles.List failed: %v", err)
	}

	t.Logf("Found %d snapshot profiles", len(profiles))

	if len(profiles) == 0 {
		t.Log("No snapshot profiles found")
		return
	}

	// Log first profile to verify field mapping
	first := profiles[0]
	t.Logf("First profile: Key=%d, Name=%q, Description=%q, IgnoreWarnings=%v",
		int(first.Key), first.Name, first.Description, first.IgnoreWarnings)

	// Test Get
	if first.Key > 0 {
		fetched, err := client.SnapshotProfiles.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("SnapshotProfiles.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("SnapshotProfiles.Get succeeded: Name=%q", fetched.Name)
		}
	}

	// Test GetByName
	if first.Name != "" {
		byName, err := client.SnapshotProfiles.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("SnapshotProfiles.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first profile for field verification
	prettyPrint(t, "Sample SnapshotProfile", first)
}

func testSnapshotProfilePeriods(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing SnapshotProfilePeriods service...")

	// List all snapshot profile periods
	periods, err := client.SnapshotProfilePeriods.List(ctx)
	if err != nil {
		t.Fatalf("SnapshotProfilePeriods.List failed: %v", err)
	}

	t.Logf("Found %d snapshot profile periods", len(periods))

	if len(periods) == 0 {
		t.Log("No snapshot profile periods found")
		return
	}

	// Log first period to verify field mapping
	first := periods[0]
	t.Logf("First period: Key=%d, Name=%q, Profile=%d, Frequency=%q, Retention=%d sec",
		int(first.Key), first.Name, int(first.Profile), first.Frequency, first.Retention)

	// Test Get
	if first.Key > 0 {
		fetched, err := client.SnapshotProfilePeriods.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("SnapshotProfilePeriods.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("SnapshotProfilePeriods.Get succeeded: Name=%q, Hour=%d, Minute=%d, DayOfWeek=%q",
				fetched.Name, fetched.Hour, fetched.Minute, fetched.DayOfWeek)
		}
	}

	// Test ListByProfile
	if first.Profile > 0 {
		profilePeriods, err := client.SnapshotProfilePeriods.ListByProfile(ctx, int(first.Profile))
		if err != nil {
			t.Errorf("SnapshotProfilePeriods.ListByProfile failed: %v", err)
		} else {
			t.Logf("Found %d periods in profile %d", len(profilePeriods), int(first.Profile))
		}

		// Test GetByName
		if first.Name != "" {
			byName, err := client.SnapshotProfilePeriods.GetByName(ctx, int(first.Profile), first.Name)
			if err != nil {
				t.Errorf("SnapshotProfilePeriods.GetByName failed: %v", err)
			} else {
				t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
			}
		}
	}

	// Pretty print first period for field verification
	prettyPrint(t, "Sample SnapshotProfilePeriod", first)
}

// TestWave3SchedulingCRUD tests Create/Update/Delete operations for Wave 3 scheduling services.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave3SchedulingCRUD
func TestWave3SchedulingCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test snapshot profile
	t.Log("Creating test snapshot profile...")
	profile, err := client.SnapshotProfiles.Create(ctx, &vergeos.SnapshotProfileCreateRequest{
		Name:        "sdk-test-profile",
		Description: "goVergeOS integration test profile - safe to delete",
	})
	if err != nil {
		t.Fatalf("SnapshotProfiles.Create failed: %v", err)
	}
	profileID := int(profile.Key)
	t.Logf("Created profile: [%d] %s", profileID, profile.Name)

	// Cleanup: delete test profile when done (this also deletes periods)
	defer func() {
		t.Log("Cleaning up: deleting test snapshot profile...")
		if err := client.SnapshotProfiles.Delete(ctx, profileID); err != nil {
			t.Logf("Warning: failed to delete test profile: %v", err)
		} else {
			t.Log("Test profile deleted successfully")
		}
	}()

	// Run CRUD subtests
	t.Run("SnapshotProfilesCRUD", func(t *testing.T) {
		testSnapshotProfilesCRUD(t, ctx, client, profileID, profile)
	})

	t.Run("SnapshotProfilePeriodsCRUD", func(t *testing.T) {
		testSnapshotProfilePeriodsCRUD(t, ctx, client, profileID)
	})
}

func testSnapshotProfilesCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, profileID int, profile *vergeos.SnapshotProfile) {
	t.Log("Testing SnapshotProfiles CRUD (using existing test profile)...")

	// Read
	profile, err := client.SnapshotProfiles.Get(ctx, profileID)
	if err != nil {
		t.Fatalf("SnapshotProfiles.Get failed: %v", err)
	}
	t.Logf("Read profile: [%d] %s (Description: %q)", profileID, profile.Name, profile.Description)

	// Update
	newDesc := "Updated goVergeOS test profile description"
	profile, err = client.SnapshotProfiles.Update(ctx, profileID, &vergeos.SnapshotProfileUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("SnapshotProfiles.Update failed: %v", err)
	}
	if profile.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, profile.Description)
	} else {
		t.Logf("Updated profile description: %q", profile.Description)
	}

	// Test GetByName
	byName, err := client.SnapshotProfiles.GetByName(ctx, "sdk-test-profile")
	if err != nil {
		t.Errorf("SnapshotProfiles.GetByName failed: %v", err)
	} else {
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	}

	// Note: Delete is tested in the parent function's defer
	t.Log("SnapshotProfiles CRUD test completed (delete will happen in cleanup)")
}

func testSnapshotProfilePeriodsCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, profileID int) {
	t.Log("Testing SnapshotProfilePeriods CRUD...")

	// Create a daily period
	period, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
		Profile:   profileID,
		Name:      "sdk-daily-test",
		Frequency: "daily",
		Hour:      ptrInt(2), // 2 AM
		Minute:    ptrInt(0),
		Retention: 7 * 24 * 60 * 60, // 7 days in seconds
	})
	if err != nil {
		t.Fatalf("SnapshotProfilePeriods.Create failed: %v", err)
	}
	periodID := int(period.Key)
	t.Logf("Created period: [%d] %s (Frequency: %s, Hour: %d, Retention: %d sec)",
		periodID, period.Name, period.Frequency, period.Hour, period.Retention)

	// Read
	period, err = client.SnapshotProfilePeriods.Get(ctx, periodID)
	if err != nil {
		t.Fatalf("SnapshotProfilePeriods.Get failed: %v", err)
	}
	t.Logf("Read period: [%d] %s (Estimated snapshots: %d)", periodID, period.Name, period.EstimatedSnapshotCount)

	// Update
	newRetention := 14 * 24 * 60 * 60 // 14 days
	period, err = client.SnapshotProfilePeriods.Update(ctx, periodID, &vergeos.SnapshotProfilePeriodUpdateRequest{
		Retention: &newRetention,
	})
	if err != nil {
		t.Fatalf("SnapshotProfilePeriods.Update failed: %v", err)
	}
	if period.Retention != newRetention {
		t.Errorf("Update verification failed: expected retention %d, got %d", newRetention, period.Retention)
	} else {
		t.Logf("Updated period retention to: %d seconds", period.Retention)
	}

	// Test GetByName
	byName, err := client.SnapshotProfilePeriods.GetByName(ctx, profileID, "sdk-daily-test")
	if err != nil {
		t.Errorf("SnapshotProfilePeriods.GetByName failed: %v", err)
	} else {
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	}

	// Test ListByProfile
	periods, err := client.SnapshotProfilePeriods.ListByProfile(ctx, profileID)
	if err != nil {
		t.Errorf("SnapshotProfilePeriods.ListByProfile failed: %v", err)
	} else {
		t.Logf("Found %d periods in profile %d", len(periods), profileID)
	}

	// Create a second period (weekly) to test multiple periods
	weeklyPeriod, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
		Profile:   profileID,
		Name:      "sdk-weekly-test",
		Frequency: "weekly",
		DayOfWeek: ptrString("sun"),
		Hour:      ptrInt(0), // Midnight
		Minute:    ptrInt(0),
		Retention: 28 * 24 * 60 * 60, // 4 weeks in seconds
	})
	if err != nil {
		t.Errorf("Failed to create weekly period: %v", err)
	} else {
		t.Logf("Created weekly period: [%d] %s (DayOfWeek: %s)", int(weeklyPeriod.Key), weeklyPeriod.Name, weeklyPeriod.DayOfWeek)

		// Delete weekly period
		err = client.SnapshotProfilePeriods.Delete(ctx, int(weeklyPeriod.Key))
		if err != nil {
			t.Errorf("Failed to delete weekly period: %v", err)
		} else {
			t.Log("Deleted weekly period successfully")
		}
	}

	// Delete daily period
	err = client.SnapshotProfilePeriods.Delete(ctx, periodID)
	if err != nil {
		t.Fatalf("SnapshotProfilePeriods.Delete failed: %v", err)
	}
	t.Log("Deleted daily period successfully")

	// Verify deletion
	_, err = client.SnapshotProfilePeriods.Get(ctx, periodID)
	if err == nil {
		t.Error("Expected error after deletion, but got none")
	} else if !vergeos.IsNotFoundError(err) {
		t.Logf("Got expected error after deletion: %v", err)
	} else {
		t.Log("Verified: period correctly deleted (NotFoundError)")
	}
}

// Helper functions for pointer values
func ptrInt(v int) *int {
	return &v
}

func ptrString(v string) *string {
	return &v
}
