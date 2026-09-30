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

// TestVMKeyAndMachineKeyDiffer exercises snapshot, drive, and NIC calls on a
// VM whose $key and machine key are different numbers. That is the normal
// case on a system that has been in use, and it is where posting the machine
// key to vm_actions hits another VM.
//
// Requires VERGEOS_TEST_VM_ID. The test powers the VM, creates and deletes a
// drive and a NIC, takes a snapshot, and restores that snapshot. The VM's
// description and power state are put back when the test finishes.
func TestVMKeyAndMachineKeyDiffer(t *testing.T) {
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
	if vm.Key.Int() == vm.Machine {
		t.Skipf("VM %d key equals machine key %d; this test needs them to differ", vm.Key.Int(), vm.Machine)
	}
	t.Logf("VM %q key=%d machine=%d running=%v", vm.Name, vm.Key.Int(), vm.Machine, vm.PowerState)

	wasRunning := vm.PowerState
	originalDesc := vm.Description
	var driveID, nicID, snapshotID int

	defer func() {
		if driveID != 0 {
			if err := client.VMDrives.Delete(ctx, driveID); err != nil && !vergeos.IsNotFoundError(err) {
				t.Logf("cleanup drive %d: %v", driveID, err)
			}
		}
		if nicID != 0 {
			if err := client.VMNICs.Delete(ctx, nicID); err != nil && !vergeos.IsNotFoundError(err) {
				t.Logf("cleanup NIC %d: %v", nicID, err)
			}
		}
		if snapshotID != 0 {
			if err := client.VMSnapshots.Delete(ctx, snapshotID); err != nil && !vergeos.IsNotFoundError(err) {
				t.Logf("cleanup snapshot %d: %v", snapshotID, err)
			}
		}
		if _, err := client.VMs.Update(ctx, vmID, &vergeos.VMUpdateRequest{Description: &originalDesc}); err != nil {
			t.Logf("restore description: %v", err)
		}
		if wasRunning {
			if err := client.VMs.PowerOn(ctx, vmID); err != nil {
				t.Logf("restore power on: %v", err)
			}
		} else if err := client.VMs.Kill(ctx, vmID); err != nil {
			t.Logf("restore power off: %v", err)
		}
	}()

	if !wasRunning {
		if err := client.VMs.PowerOn(ctx, vmID); err != nil {
			t.Fatalf("PowerOn(%d) failed: %v", vmID, err)
		}
	}

	t.Run("DeleteDriveFromRunningVM", func(t *testing.T) {
		stamp := time.Now().Format("20060102-150405")
		drive, err := client.VMDrives.Create(ctx, vmID, &vergeos.VMDriveCreateRequest{
			Name:      "sdk-key-disk-" + stamp,
			Interface: "virtio-scsi",
			Media:     "disk",
			SizeGB:    1,
		})
		if err != nil {
			t.Fatalf("VMDrives.Create(%d) failed: %v", vmID, err)
		}
		driveID = drive.Key.Int()
		if drive.Machine != vm.Machine {
			t.Fatalf("drive machine = %d, want VM machine %d", drive.Machine, vm.Machine)
		}
		if drive.Machine == vm.Key.Int() {
			t.Fatalf("drive was attached using the VM $key %d", vm.Key.Int())
		}

		drives, err := client.VMDrives.List(ctx, vmID)
		if err != nil {
			t.Fatalf("VMDrives.List(%d) failed: %v", vmID, err)
		}
		found := false
		for _, d := range drives {
			if d.Key.Int() == driveID {
				found = true
				if d.Machine != vm.Machine {
					t.Errorf("listed drive machine = %d, want %d", d.Machine, vm.Machine)
				}
			}
		}
		if !found {
			t.Fatalf("List(%d) did not return drive %d", vmID, driveID)
		}

		if err := client.VMDrives.HotplugDrive(ctx, vmID, driveID); err != nil {
			t.Fatalf("HotplugDrive(%d, %d) failed: %v", vmID, driveID, err)
		}
		plugged, err := client.VMDrives.Get(ctx, driveID)
		if err != nil {
			t.Fatalf("VMDrives.Get(%d) failed: %v", driveID, err)
		}
		if plugged.PowerState != "online" {
			t.Fatalf("drive power state = %q, want online before delete", plugged.PowerState)
		}

		if err := client.VMDrives.Delete(ctx, driveID); err != nil {
			t.Fatalf("VMDrives.Delete(%d) on running VM failed: %v", driveID, err)
		}
		driveID = 0
		if _, err := client.VMDrives.Get(ctx, plugged.Key.Int()); !vergeos.IsNotFoundError(err) {
			t.Fatalf("drive still present after delete: %v", err)
		}
	})

	t.Run("DeleteNICFromRunningVM", func(t *testing.T) {
		networks, err := client.Networks.List(ctx, vergeos.WithLimit(1))
		if err != nil {
			t.Fatalf("Networks.List failed: %v", err)
		}
		if len(networks) == 0 {
			t.Skip("No networks available to attach a NIC")
		}

		stamp := time.Now().Format("20060102-150405")
		nic, err := client.VMNICs.Create(ctx, vmID, &vergeos.VMNICCreateRequest{
			Name:      "sdk-key-nic-" + stamp,
			Interface: "virtio",
			VNET:      networks[0].Key.Int(),
		})
		if err != nil {
			t.Fatalf("VMNICs.Create(%d) failed: %v", vmID, err)
		}
		nicID = nic.Key.Int()
		if nic.Machine != vm.Machine {
			t.Fatalf("NIC machine = %d, want VM machine %d", nic.Machine, vm.Machine)
		}
		if nic.Machine == vm.Key.Int() {
			t.Fatalf("NIC was attached using the VM $key %d", vm.Key.Int())
		}

		nics, err := client.VMNICs.List(ctx, vmID)
		if err != nil {
			t.Fatalf("VMNICs.List(%d) failed: %v", vmID, err)
		}
		found := false
		for _, n := range nics {
			if n.Key.Int() == nicID {
				found = true
			}
		}
		if !found {
			t.Fatalf("List(%d) did not return NIC %d", vmID, nicID)
		}

		current, err := client.VMNICs.Get(ctx, nicID)
		if err != nil {
			t.Fatalf("VMNICs.Get(%d) failed: %v", nicID, err)
		}
		t.Logf("NIC %d power state %q before delete", nicID, current.PowerState)
		if current.PowerState == "" || current.PowerState == "down" {
			t.Logf("NIC is not up; delete will not hot-unplug. Unit tests cover that path.")
		}

		if err := client.VMNICs.Delete(ctx, nicID); err != nil {
			t.Fatalf("VMNICs.Delete(%d) on running VM failed: %v", nicID, err)
		}
		nicID = 0
		if _, err := client.VMNICs.Get(ctx, current.Key.Int()); !vergeos.IsNotFoundError(err) {
			t.Fatalf("NIC still present after delete: %v", err)
		}
	})

	t.Run("Restore", func(t *testing.T) {
		stamp := time.Now().Format("20060102-150405")
		before := "govergeos-key-before-" + stamp
		after := "govergeos-key-after-" + stamp
		if _, err := client.VMs.Update(ctx, vmID, &vergeos.VMUpdateRequest{Description: &before}); err != nil {
			t.Fatalf("set description: %v", err)
		}

		expires := time.Now().Add(time.Hour).Unix()
		snapshotName := "sdk-key-snapshot-" + stamp
		snapshot, err := client.VMSnapshots.Create(ctx, &vergeos.VMSnapshotCreateRequest{
			VM:          vmID,
			Name:        snapshotName,
			Description: "goVergeOS key resolution test - safe to delete",
			ExpiresType: "date",
			Expires:     &expires,
		})
		if err != nil {
			t.Fatalf("VMSnapshots.Create VM %d failed: %v", vmID, err)
		}
		snapshotID = int(snapshot.Key)
		if int(snapshot.Machine) != vm.Machine {
			t.Fatalf("snapshot machine = %d, want %d", int(snapshot.Machine), vm.Machine)
		}

		listed, err := client.VMSnapshots.ListByVM(ctx, vmID)
		if err != nil {
			t.Fatalf("ListByVM(%d) failed: %v", vmID, err)
		}
		seen := false
		for _, snap := range listed {
			if int(snap.Key) == snapshotID {
				seen = true
				if int(snap.Machine) != vm.Machine {
					t.Errorf("listed snapshot machine = %d, want %d", int(snap.Machine), vm.Machine)
				}
			}
		}
		if !seen {
			t.Fatalf("ListByVM(%d) did not return snapshot %d", vmID, snapshotID)
		}

		byName, err := client.VMSnapshots.GetByName(ctx, vmID, snapshotName)
		if err != nil {
			t.Fatalf("GetByName(%d, %q) failed: %v", vmID, snapshotName, err)
		}
		if int(byName.Key) != snapshotID {
			t.Fatalf("GetByName key = %d, want %d", int(byName.Key), snapshotID)
		}

		// Listing by the machine key must not be treated as this VM's $key.
		// When that number is another VM, this snapshot must not come back.
		other, err := client.VMs.Get(ctx, vm.Machine)
		if vergeos.IsNotFoundError(err) {
			if _, listErr := client.VMSnapshots.ListByVM(ctx, vm.Machine); listErr == nil {
				t.Fatalf("ListByVM(%d) succeeded for a VM $key that does not exist", vm.Machine)
			}
		} else if err != nil {
			t.Fatalf("VMs.Get(%d) failed: %v", vm.Machine, err)
		} else if other.Machine != vm.Machine {
			others, err := client.VMSnapshots.ListByVM(ctx, vm.Machine)
			if err != nil {
				t.Fatalf("ListByVM(%d) failed: %v", vm.Machine, err)
			}
			for _, snap := range others {
				if int(snap.Key) == snapshotID {
					t.Fatalf("ListByVM(machine key %d) returned snapshot %d, which belongs to VM %d", vm.Machine, snapshotID, vmID)
				}
			}
		}

		if _, err := client.VMs.Update(ctx, vmID, &vergeos.VMUpdateRequest{Description: &after}); err != nil {
			t.Fatalf("change description: %v", err)
		}
		// Kill stops the VM when the guest ignores a shutdown request.
		// Restore runs against a powered-off VM.
		if err := client.VMs.Kill(ctx, vmID); err != nil {
			t.Fatalf("Kill(%d) before restore failed: %v", vmID, err)
		}
		if err := client.VMSnapshots.Restore(ctx, snapshotID, nil); err != nil {
			t.Fatalf("Restore(%d) failed: %v", snapshotID, err)
		}

		restored, err := client.VMs.Get(ctx, vmID)
		if err != nil {
			t.Fatalf("VMs.Get(%d) after restore failed: %v", vmID, err)
		}
		if restored.Description != before {
			t.Fatalf("description after restore = %q, want %q", restored.Description, before)
		}
		if restored.Key.Int() != vmID || restored.Machine != vm.Machine {
			t.Fatalf("restore changed VM identity: key=%d machine=%d", restored.Key.Int(), restored.Machine)
		}

		if err := client.VMSnapshots.Delete(ctx, snapshotID); err != nil {
			t.Fatalf("delete snapshot %d: %v", snapshotID, err)
		}
		snapshotID = 0
	})
}
