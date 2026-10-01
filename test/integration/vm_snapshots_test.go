//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestVMSnapshots tests the VM Snapshot service against a live VergeOS API.
func TestVMSnapshots(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		snapshots, err := client.VMSnapshots.List(ctx)
		if err != nil {
			t.Fatalf("VMSnapshots.List failed: %v", err)
		}
		t.Logf("Found %d VM snapshots", len(snapshots))

		if len(snapshots) == 0 {
			t.Log("No VM snapshots found - skipping detailed tests")
			return
		}

		first := snapshots[0]
		t.Logf("First snapshot: Key=%d, Name=%q, Machine=%d, MachineDisplay=%q",
			int(first.Key), first.Name, int(first.Machine), first.MachineDisplay)
		prettyPrint(t, "Sample VMSnapshot", first)
	})

	t.Run("Get", func(t *testing.T) {
		snapshots, err := client.VMSnapshots.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(snapshots) == 0 {
			t.Skip("No VM snapshots available")
		}

		first := snapshots[0]
		fetched, err := client.VMSnapshots.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("VMSnapshots.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("VMSnapshots.Get succeeded: Name=%q, Created=%d, Expires=%d",
			fetched.Name, fetched.Created, fetched.Expires)
	})

	t.Run("ListByVM", func(t *testing.T) {
		vms, err := client.VMs.List(ctx, vergeos.WithLimit(50))
		if err != nil {
			t.Fatalf("VMs.List failed: %v", err)
		}

		var vm *vergeos.VM
		for i := range vms {
			if vms[i].IsSnapshot || vms[i].Key.Int() <= 0 || vms[i].Machine <= 0 {
				continue
			}
			if vm == nil || vms[i].Key.Int() != vms[i].Machine {
				chosen := vms[i]
				vm = &chosen
				if vms[i].Key.Int() != vms[i].Machine {
					break
				}
			}
		}
		if vm == nil {
			t.Skip("No VMs available")
		}

		vmSnapshots, err := client.VMSnapshots.ListByVM(ctx, vm.Key.Int())
		if err != nil {
			t.Fatalf("VMSnapshots.ListByVM(%d) failed: %v", vm.Key.Int(), err)
		}
		t.Logf("Found %d snapshots for VM %d (machine %d)", len(vmSnapshots), vm.Key.Int(), vm.Machine)
		for _, snap := range vmSnapshots {
			if int(snap.Machine) != vm.Machine {
				t.Errorf("snapshot %d machine=%d, VM %d machine=%d", int(snap.Key), int(snap.Machine), vm.Key.Int(), vm.Machine)
			}
		}
	})

	t.Run("ListExpiring", func(t *testing.T) {
		expiring, err := client.VMSnapshots.ListExpiring(ctx, 30)
		if err != nil {
			t.Fatalf("VMSnapshots.ListExpiring failed: %v", err)
		}
		t.Logf("Found %d snapshots expiring within 30 days", len(expiring))
	})
}

// TestVMSnapshotsCRUD tests Create/Update/Delete operations for VM Snapshots.
// Requires VERGEOS_TEST_VM_ID environment variable.
func TestVMSnapshotsCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
		VM:          vmID,
		Name:        snapshotName,
		Description: "goVergeOS integration test snapshot - safe to delete",
		ExpiresType: "date",
	})
	if err != nil {
		t.Fatalf("VMSnapshots.Create failed: %v", err)
	}
	snapshotID := int(snapshot.Key)
	t.Logf("Created snapshot: [%d] %s", snapshotID, snapshot.Name)

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
		t.Logf("Set snapshot to never expire: Expires=%d", snapshot.Expires)
	}

	// Test SetExpires (7 days from now)
	expires := time.Now().Add(7 * 24 * time.Hour).Unix()
	snapshot, err = client.VMSnapshots.SetExpires(ctx, snapshotID, expires)
	if err != nil {
		t.Errorf("VMSnapshots.SetExpires failed: %v", err)
	} else {
		t.Logf("Set snapshot expires: Expires=%d", snapshot.Expires)
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

// TestVMsSnapshot creates a snapshot through VMs.Snapshot and checks that a
// machine_snapshots row exists for the VM. Requires VERGEOS_TEST_VM_ID.
// Quiesce is left false so the test does not depend on a guest agent.
func TestVMsSnapshot(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	vmIDStr := os.Getenv("VERGEOS_TEST_VM_ID")
	if vmIDStr == "" {
		t.Skip("Skipping: VERGEOS_TEST_VM_ID not set")
	}

	var vmID int
	if err := json.Unmarshal([]byte(vmIDStr), &vmID); err != nil {
		t.Fatalf("Invalid VERGEOS_TEST_VM_ID: %v", err)
	}

	vm, err := client.VMs.Get(ctx, vmID)
	if err != nil {
		t.Fatalf("VMs.Get(%d) failed: %v", vmID, err)
	}
	if vm.Machine <= 0 {
		t.Fatalf("VM %d has no machine key", vmID)
	}

	snapshotName := "sdk-test-vms-snapshot-" + time.Now().Format("20060102-150405")
	snapshot, err := client.VMs.Snapshot(ctx, vmID, &vergeos.VMSnapshotOptions{
		Name:      snapshotName,
		Retention: 3600,
	})
	if err != nil {
		t.Fatalf("VMs.Snapshot failed: %v", err)
	}
	if snapshot == nil || int(snapshot.Key) == 0 {
		t.Fatal("VMs.Snapshot returned no snapshot")
	}
	snapshotID := int(snapshot.Key)
	t.Logf("VMs.Snapshot created [%d] %s", snapshotID, snapshot.Name)

	defer func() {
		t.Log("Cleaning up: deleting test snapshot...")
		if err := client.VMSnapshots.Delete(ctx, snapshotID); err != nil {
			t.Logf("Warning: failed to delete test snapshot: %v", err)
		}
	}()

	fetched, err := client.VMSnapshots.Get(ctx, snapshotID)
	if err != nil {
		t.Fatalf("VMSnapshots.Get(%d) failed: %v", snapshotID, err)
	}
	if fetched.Name != snapshotName {
		t.Errorf("snapshot name = %q, want %q", fetched.Name, snapshotName)
	}

	byName, err := client.VMSnapshots.GetByName(ctx, vmID, snapshotName)
	if err != nil {
		t.Fatalf("snapshot %q not found on VM %d: %v", snapshotName, vmID, err)
	}
	if int(byName.Key) != snapshotID {
		t.Errorf("GetByName key = %d, want %d", int(byName.Key), snapshotID)
	}
	t.Logf("Found machine_snapshots row: Key=%d Name=%q Machine=%d Expires=%d",
		int(byName.Key), byName.Name, int(byName.Machine), byName.Expires)
}

// TestVMMigrate tests the VM migration functionality.
// Requires VERGEOS_TEST_VM_ID and VERGEOS_TEST_TARGET_NODE environment variables.
func TestVMMigrate(t *testing.T) {
	client := setupTestClient(t)
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

	vm, err := client.VMs.Get(ctx, vmID)
	if err != nil {
		t.Fatalf("VMs.Get failed: %v", err)
	}
	t.Logf("VM state: Name=%q, PowerState=%v", vm.Name, vm.PowerState)

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
