//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave10VMSnapshots tests the VM Snapshot service against a live VergeOS API.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave10VMSnapshots
func TestWave10VMSnapshots(t *testing.T) {
	client := setupTestClientWave10(t)
	ctx := context.Background()

	t.Run("VMSnapshots", func(t *testing.T) {
		testVMSnapshots(t, ctx, client)
	})
}

func testVMSnapshots(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VMSnapshots service...")

	// List all VM snapshots
	snapshots, err := client.VMSnapshots.List(ctx)
	if err != nil {
		t.Fatalf("VMSnapshots.List failed: %v", err)
	}

	t.Logf("Found %d VM snapshots", len(snapshots))

	if len(snapshots) == 0 {
		t.Log("No VM snapshots found - skipping detailed tests")
		return
	}

	// Log first snapshot to verify field mapping
	first := snapshots[0]
	t.Logf("First snapshot: Key=%d, Name=%q, Machine=%d, MachineDisplay=%q",
		int(first.Key), first.Name, int(first.Machine), first.MachineDisplay)

	// Test Get
	if first.Key > 0 {
		fetched, err := client.VMSnapshots.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VMSnapshots.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VMSnapshots.Get succeeded: Name=%q, Created=%d, Expires=%d",
				fetched.Name, fetched.Created, fetched.Expires)
		}
	}

	// Test ListByVM
	if first.Machine > 0 {
		vmSnapshots, err := client.VMSnapshots.ListByVM(ctx, int(first.Machine))
		if err != nil {
			t.Errorf("VMSnapshots.ListByVM failed: %v", err)
		} else {
			t.Logf("Found %d snapshots for VM %d", len(vmSnapshots), int(first.Machine))
		}
	}

	// Test ListExpiring (snapshots expiring within 30 days)
	expiring, err := client.VMSnapshots.ListExpiring(ctx, 30)
	if err != nil {
		t.Errorf("VMSnapshots.ListExpiring failed: %v", err)
	} else {
		t.Logf("Found %d snapshots expiring within 30 days", len(expiring))
	}

	// Pretty print first snapshot for field verification
	prettyPrintWave10(t, "Sample VMSnapshot", first)
}

// TestWave10VMSnapshotsCRUD tests Create/Update/Delete operations for VM Snapshots.
// Note: Requires an existing VM to test with.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	VERGEOS_TEST_VM_ID=123 \
//	  go test -tags=integration -v ./test/integration/ -run TestWave10VMSnapshotsCRUD
func TestWave10VMSnapshotsCRUD(t *testing.T) {
	client := setupTestClientWave10(t)
	ctx := context.Background()

	// Check for test VM ID
	vmIDStr := os.Getenv("VERGEOS_TEST_VM_ID")
	if vmIDStr == "" {
		t.Skip("Skipping CRUD test: VERGEOS_TEST_VM_ID not set")
	}

	var vmID int
	if err := json.Unmarshal([]byte(vmIDStr), &vmID); err != nil {
		t.Fatalf("Invalid VERGEOS_TEST_VM_ID: %v", err)
	}

	t.Logf("Testing VM Snapshot CRUD with VM ID: %d", vmID)

	// Create a test snapshot
	snapshotName := "sdk-test-snapshot-" + time.Now().Format("20060102-150405")
	snapshot, err := client.VMSnapshots.Create(ctx, &vergeos.VMSnapshotCreateRequest{
		Machine:     vmID,
		Name:        snapshotName,
		Description: "goVergeOS integration test snapshot - safe to delete",
		ExpiresType: "date",
	})
	if err != nil {
		t.Fatalf("VMSnapshots.Create failed: %v", err)
	}
	snapshotID := int(snapshot.Key)
	t.Logf("Created snapshot: [%d] %s", snapshotID, snapshot.Name)

	// Cleanup: delete test snapshot when done
	defer func() {
		t.Log("Cleaning up: deleting test snapshot...")
		if err := client.VMSnapshots.Delete(ctx, snapshotID); err != nil {
			t.Logf("Warning: failed to delete test snapshot: %v", err)
		} else {
			t.Log("Test snapshot deleted successfully")
		}
	}()

	// Read
	snapshot, err = client.VMSnapshots.Get(ctx, snapshotID)
	if err != nil {
		t.Fatalf("VMSnapshots.Get failed: %v", err)
	}
	t.Logf("Read snapshot: [%d] %s (Description: %q)", snapshotID, snapshot.Name, snapshot.Description)

	// Update description
	newDesc := "Updated goVergeOS test snapshot description"
	snapshot, err = client.VMSnapshots.Update(ctx, snapshotID, &vergeos.VMSnapshotUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("VMSnapshots.Update failed: %v", err)
	}
	if snapshot.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, snapshot.Description)
	} else {
		t.Logf("Updated snapshot description: %q", snapshot.Description)
	}

	// Test SetNeverExpires
	snapshot, err = client.VMSnapshots.SetNeverExpires(ctx, snapshotID)
	if err != nil {
		t.Errorf("VMSnapshots.SetNeverExpires failed: %v", err)
	} else {
		t.Logf("Set snapshot to never expire: ExpiresType=%s, Expires=%d", snapshot.ExpiresType, snapshot.Expires)
	}

	// Test SetExpires (7 days from now)
	expires := time.Now().Add(7 * 24 * time.Hour).Unix()
	snapshot, err = client.VMSnapshots.SetExpires(ctx, snapshotID, expires)
	if err != nil {
		t.Errorf("VMSnapshots.SetExpires failed: %v", err)
	} else {
		t.Logf("Set snapshot expires: ExpiresType=%s, Expires=%d", snapshot.ExpiresType, snapshot.Expires)
	}

	// Test GetByName
	byName, err := client.VMSnapshots.GetByName(ctx, vmID, snapshotName)
	if err != nil {
		t.Errorf("VMSnapshots.GetByName failed: %v", err)
	} else {
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	}

	t.Log("VM Snapshot CRUD test completed")
}

// TestWave10TagCategories tests the TagCategory service against a live VergeOS API.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave10TagCategories
func TestWave10TagCategories(t *testing.T) {
	client := setupTestClientWave10(t)
	ctx := context.Background()

	t.Run("TagCategories", func(t *testing.T) {
		testTagCategories(t, ctx, client)
	})
}

func testTagCategories(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing TagCategories service...")

	// List all tag categories
	categories, err := client.TagCategories.List(ctx)
	if err != nil {
		t.Fatalf("TagCategories.List failed: %v", err)
	}

	t.Logf("Found %d tag categories", len(categories))

	if len(categories) == 0 {
		t.Log("No tag categories found")
		return
	}

	// Log first category to verify field mapping
	first := categories[0]
	t.Logf("First category: Key=%d, Name=%q, SingleTagSelection=%v, TaggableVMs=%v",
		int(first.Key), first.Name, first.SingleTagSelection, first.TaggableVMs)

	// Test Get
	if first.Key > 0 {
		fetched, err := client.TagCategories.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("TagCategories.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("TagCategories.Get succeeded: Name=%q", fetched.Name)
		}
	}

	// Test GetByName
	if first.Name != "" {
		byName, err := client.TagCategories.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("TagCategories.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first category for field verification
	prettyPrintWave10(t, "Sample TagCategory", first)
}

// TestWave10TagCategoriesCRUD tests Create/Update/Delete operations for TagCategories and Tags.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave10TagCategoriesCRUD
func TestWave10TagCategoriesCRUD(t *testing.T) {
	client := setupTestClientWave10(t)
	ctx := context.Background()

	// Create a test tag category
	t.Log("Creating test tag category...")
	categoryName := "sdk-test-category-" + time.Now().Format("20060102-150405")
	trueVal := true
	category, err := client.TagCategories.Create(ctx, &vergeos.TagCategoryCreateRequest{
		Name:        categoryName,
		Description: "goVergeOS integration test category - safe to delete",
		TaggableVMs: &trueVal,
	})
	if err != nil {
		t.Fatalf("TagCategories.Create failed: %v", err)
	}
	categoryID := int(category.Key)
	t.Logf("Created category: [%d] %s", categoryID, category.Name)

	// Cleanup: delete test category when done (this also deletes tags)
	defer func() {
		t.Log("Cleaning up: deleting test tag category...")
		if err := client.TagCategories.Delete(ctx, categoryID); err != nil {
			t.Logf("Warning: failed to delete test category: %v", err)
		} else {
			t.Log("Test category deleted successfully")
		}
	}()

	// Update category
	newDesc := "Updated goVergeOS test category description"
	category, err = client.TagCategories.Update(ctx, categoryID, &vergeos.TagCategoryUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("TagCategories.Update failed: %v", err)
	}
	if category.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, category.Description)
	} else {
		t.Logf("Updated category description: %q", category.Description)
	}

	// Test Tags CRUD within this category
	t.Run("TagsCRUD", func(t *testing.T) {
		testTagsCRUD(t, ctx, client, categoryID)
	})
}

func testTagsCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, categoryID int) {
	t.Log("Testing Tags CRUD...")

	// Create a tag
	tagName := "sdk-test-tag-" + time.Now().Format("150405")
	tag, err := client.Tags.Create(ctx, &vergeos.TagCreateRequest{
		Category:    categoryID,
		Name:        tagName,
		Description: "goVergeOS integration test tag - safe to delete",
	})
	if err != nil {
		t.Fatalf("Tags.Create failed: %v", err)
	}
	tagID := int(tag.Key)
	t.Logf("Created tag: [%d] %s (CategoryDisplay: %q)", tagID, tag.Name, tag.CategoryDisplay)

	// Read
	tag, err = client.Tags.Get(ctx, tagID)
	if err != nil {
		t.Fatalf("Tags.Get failed: %v", err)
	}
	t.Logf("Read tag: [%d] %s (Description: %q)", tagID, tag.Name, tag.Description)

	// Update
	newDesc := "Updated goVergeOS test tag description"
	tag, err = client.Tags.Update(ctx, tagID, &vergeos.TagUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Tags.Update failed: %v", err)
	}
	if tag.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, tag.Description)
	} else {
		t.Logf("Updated tag description: %q", tag.Description)
	}

	// Test GetByName
	byName, err := client.Tags.GetByName(ctx, tagName)
	if err != nil {
		t.Errorf("Tags.GetByName failed: %v", err)
	} else {
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	}

	// Test ListByCategory
	tags, err := client.Tags.ListByCategory(ctx, categoryID)
	if err != nil {
		t.Errorf("Tags.ListByCategory failed: %v", err)
	} else {
		t.Logf("Found %d tags in category %d", len(tags), categoryID)
	}

	// Delete
	err = client.Tags.Delete(ctx, tagID)
	if err != nil {
		t.Fatalf("Tags.Delete failed: %v", err)
	}
	t.Log("Tag deleted successfully")

	// Verify deletion
	_, err = client.Tags.Get(ctx, tagID)
	if err == nil {
		t.Error("Expected error after deletion, but got none")
	} else if vergeos.IsNotFoundError(err) {
		t.Log("Verified: tag correctly deleted (NotFoundError)")
	} else {
		t.Logf("Got error after deletion: %v", err)
	}
}

// TestWave10VMMigrate tests the VM migration functionality.
// Note: Requires an existing VM and multiple nodes to test with.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	VERGEOS_TEST_VM_ID=123 VERGEOS_TEST_TARGET_NODE=2 \
//	  go test -tags=integration -v ./test/integration/ -run TestWave10VMMigrate
func TestWave10VMMigrate(t *testing.T) {
	client := setupTestClientWave10(t)
	ctx := context.Background()

	vmIDStr := os.Getenv("VERGEOS_TEST_VM_ID")
	targetNodeStr := os.Getenv("VERGEOS_TEST_TARGET_NODE")
	if vmIDStr == "" || targetNodeStr == "" {
		t.Skip("Skipping migration test: VERGEOS_TEST_VM_ID and VERGEOS_TEST_TARGET_NODE must be set")
	}

	var vmID, targetNode int
	if err := json.Unmarshal([]byte(vmIDStr), &vmID); err != nil {
		t.Fatalf("Invalid VERGEOS_TEST_VM_ID: %v", err)
	}
	if err := json.Unmarshal([]byte(targetNodeStr), &targetNode); err != nil {
		t.Fatalf("Invalid VERGEOS_TEST_TARGET_NODE: %v", err)
	}

	t.Logf("Testing VM migration: VM %d to node %d", vmID, targetNode)

	// Get current VM state
	vm, err := client.VMs.Get(ctx, vmID)
	if err != nil {
		t.Fatalf("VMs.Get failed: %v", err)
	}
	t.Logf("VM state: Name=%q, PowerState=%v", vm.Name, vm.PowerState)

	// Attempt migration
	liveVal := true
	err = client.VMs.Migrate(ctx, vmID, &vergeos.VMMigrateOptions{
		TargetNode: targetNode,
		Live:       &liveVal,
	})
	if err != nil {
		t.Logf("VMs.Migrate failed (may be expected if VM is off or target is same): %v", err)
	} else {
		t.Log("VM migration initiated successfully")
	}
}

// Helper functions

func setupTestClientWave10(t *testing.T) *vergeos.Client {
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		t.Skip("Skipping integration test: VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD must be set")
	}

	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true),
		vergeos.WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

func prettyPrintWave10(t *testing.T, label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}
