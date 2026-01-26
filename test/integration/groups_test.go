//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestGroupsList tests listing groups.
func TestGroupsList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	groups, err := client.Groups.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	t.Logf("Found %d group(s)", len(groups))
	for i, g := range groups {
		if i >= 5 {
			t.Logf("  ... and %d more", len(groups)-5)
			break
		}
		t.Logf("  - ID=%d Name=%q SystemGroup=%v Enabled=%v",
			g.ID, g.Name, g.SystemGroup, g.Enabled)
	}
}

// TestGroupCRUD tests full CRUD operations on groups.
func TestGroupCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	testName := "sdk-test-group-" + time.Now().Format("20060102150405")

	// Create
	t.Logf("Creating group: %s", testName)
	enabled := true
	created, err := client.Groups.Create(ctx, &vergeos.GroupCreateRequest{
		Name:        testName,
		Description: "goVergeOS integration test group - safe to delete",
		Enabled:     &enabled,
	})
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	groupID := int(created.ID)
	t.Logf("Created group with ID: %d", groupID)

	// Cleanup
	defer func() {
		t.Logf("Cleaning up group: %d", groupID)
		if err := client.Groups.Delete(ctx, groupID); err != nil {
			t.Errorf("Failed to delete group: %v", err)
		}
	}()

	// Verify creation
	if created.Name != testName {
		t.Errorf("Expected name %q, got %q", testName, created.Name)
	}
	if !created.Enabled {
		t.Error("Expected group to be enabled")
	}

	// Get by ID
	t.Logf("Getting group by ID: %d", groupID)
	got, err := client.Groups.Get(ctx, groupID)
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}
	if got.Name != testName {
		t.Errorf("Expected name %q, got %q", testName, got.Name)
	}

	// Get by name
	t.Logf("Getting group by name: %s", testName)
	byName, err := client.Groups.GetByName(ctx, testName)
	if err != nil {
		t.Fatalf("Failed to get group by name: %v", err)
	}
	if int(byName.ID) != groupID {
		t.Errorf("Expected ID %d, got %d", groupID, byName.ID)
	}

	// Update
	newDesc := "Updated description - " + time.Now().Format(time.RFC3339)
	t.Logf("Updating group description")
	updated, err := client.Groups.Update(ctx, groupID, &vergeos.GroupUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}
	if updated.Description != newDesc {
		t.Errorf("Expected description %q, got %q", newDesc, updated.Description)
	}

	t.Log("Group CRUD test completed successfully")
}
