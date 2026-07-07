//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/macstadium/govergeos"
)

// TestUpdateSettings tests the UpdateSettings service against a live VergeOS API.
func TestUpdateSettings(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("Get", func(t *testing.T) {
		settings, err := client.UpdateSettings.Get(ctx)
		if err != nil {
			t.Fatalf("UpdateSettings.Get failed: %v", err)
		}

		t.Logf("Update settings: Branch=%d, BranchName=%q, Source=%d",
			settings.Branch, settings.BranchName, settings.Source)
		t.Logf("Flags: AutoRefresh=%v, AutoUpdate=%v, RebootRequired=%v, Installed=%v",
			settings.AutoRefresh, settings.AutoUpdate, settings.RebootRequired, settings.Installed)
		t.Logf("Snapshot: SnapshotCloudOnUpdate=%v, ExpireSeconds=%d",
			settings.SnapshotCloudOnUpdate, settings.SnapshotCloudExpireSeconds)

		prettyPrint(t, "UpdateSettings", settings)
	})
}

// TestUpdateBranches tests the UpdateBranches service against a live VergeOS API.
func TestUpdateBranches(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		branches, err := client.UpdateBranches.List(ctx)
		if err != nil {
			t.Fatalf("UpdateBranches.List failed: %v", err)
		}
		t.Logf("Found %d update branches", len(branches))

		if len(branches) == 0 {
			t.Log("No update branches found")
			return
		}

		for _, b := range branches {
			t.Logf("  Branch: Key=%d, Name=%q, Description=%q",
				int(b.Key), b.Name, b.Description)
		}
		prettyPrint(t, "Sample UpdateBranch", branches[0])
	})

	t.Run("Get", func(t *testing.T) {
		branches, err := client.UpdateBranches.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(branches) == 0 {
			t.Skip("No update branches available")
		}

		first := branches[0]
		fetched, err := client.UpdateBranches.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("UpdateBranches.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("UpdateBranches.Get succeeded: Name=%q", fetched.Name)
	})
}

// TestUpdateSourcePackages tests the UpdateSourcePackages service against a live VergeOS API.
func TestUpdateSourcePackages(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		packages, err := client.UpdateSourcePackages.List(ctx)
		if err != nil {
			t.Fatalf("UpdateSourcePackages.List failed: %v", err)
		}
		t.Logf("Found %d update source packages", len(packages))

		if len(packages) == 0 {
			t.Log("No update source packages found")
			return
		}

		for _, p := range packages {
			t.Logf("  Package: Key=%d, Name=%q, Version=%q, Branch=%d, Downloaded=%v",
				int(p.Key), p.Name, p.Version, p.Branch, p.Downloaded)
		}
		prettyPrint(t, "Sample UpdateSourcePackage", packages[0])
	})

	t.Run("ListByBranchAndSource", func(t *testing.T) {
		// Get settings to find current branch and source
		settings, err := client.UpdateSettings.Get(ctx)
		if err != nil {
			t.Skip("Could not get update settings")
		}

		if settings.Branch == 0 || settings.Source == 0 {
			t.Skip("Update settings have no branch/source configured")
		}

		packages, err := client.UpdateSourcePackages.ListByBranchAndSource(ctx, settings.Branch, settings.Source)
		if err != nil {
			t.Fatalf("UpdateSourcePackages.ListByBranchAndSource(%d, %d) failed: %v",
				settings.Branch, settings.Source, err)
		}
		t.Logf("Found %d packages for branch=%d, source=%d", len(packages), settings.Branch, settings.Source)

		for _, p := range packages {
			t.Logf("  Package: Name=%q, Version=%q, Downloaded=%v", p.Name, p.Version, p.Downloaded)
		}
	})
}
