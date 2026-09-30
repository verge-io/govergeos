//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestCloudInitFileCreateForVM creates a cloud-init file on a throwaway VM.
// Owner must be the VM $key reference (vms/<VM.ID>), not the machine key.
// The test runs when the lab client env vars are set (see setupTestClient).
func TestCloudInitFileCreateForVM(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	stamp := time.Now().Format("20060102-150405")
	vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
		Name:        "sdk-cloudinit-" + stamp,
		Description: "goVergeOS cloud-init integration test",
		CPUCores:    1,
		RAM:         512,
		OSFamily:    "linux",
	})
	if err != nil {
		t.Fatalf("VMs.Create failed: %v", err)
	}
	vmID := vm.ID.Int()
	t.Logf("Created VM %q key=%d machine=%d", vm.Name, vmID, vm.Machine)

	var fileID int
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if fileID != 0 {
			if err := client.CloudInitFiles.Delete(cleanupCtx, fileID); err != nil && !vergeos.IsNotFoundError(err) {
				t.Logf("cleanup cloud-init file %d: %v", fileID, err)
			}
		}
		if err := client.VMs.Delete(cleanupCtx, vmID); err != nil && !vergeos.IsNotFoundError(err) {
			t.Logf("cleanup VM %d: %v", vmID, err)
		}
	}()

	contents := "#cloud-config\nhostname: sdk-cloudinit\n"
	file, err := client.CloudInitFiles.CreateForVM(ctx, vmID, &vergeos.CloudInitFileCreateRequest{
		Name:     "/user-data",
		Contents: contents,
	})
	if err != nil {
		t.Fatalf("CloudInitFiles.CreateForVM(%d) failed: %v", vmID, err)
	}
	fileID = file.ID.Int()

	wantOwner := fmt.Sprintf("vms/%d", vmID)
	if file.Owner != wantOwner {
		t.Fatalf("owner = %q, want VM $key reference %q (machine key is %d)", file.Owner, wantOwner, vm.Machine)
	}

	got, err := client.CloudInitFiles.Get(ctx, fileID)
	if err != nil {
		t.Fatalf("CloudInitFiles.Get(%d) failed: %v", fileID, err)
	}
	if got.Owner != wantOwner {
		t.Fatalf("Get owner = %q, want %q", got.Owner, wantOwner)
	}
	if got.Name != "/user-data" {
		t.Fatalf("Get name = %q, want /user-data", got.Name)
	}

	listed, err := client.CloudInitFiles.ListByVM(ctx, vmID)
	if err != nil {
		t.Fatalf("CloudInitFiles.ListByVM(%d) failed: %v", vmID, err)
	}
	found := false
	for _, f := range listed {
		if f.ID.Int() == fileID {
			found = true
			if f.Owner != wantOwner {
				t.Errorf("listed owner = %q, want %q", f.Owner, wantOwner)
			}
		}
	}
	if !found {
		t.Fatalf("ListByVM(%d) did not return file %d", vmID, fileID)
	}

	// Listing by the machine key asks for a different VM's files when the
	// two keys differ. This file must not show up there.
	if vm.Machine > 0 && vm.Machine != vmID {
		byMachine, err := client.CloudInitFiles.ListByVM(ctx, vm.Machine)
		if err != nil {
			t.Fatalf("CloudInitFiles.ListByVM(machine %d) failed: %v", vm.Machine, err)
		}
		for _, f := range byMachine {
			if f.ID.Int() == fileID {
				t.Fatalf("file %d was listed for machine key %d", fileID, vm.Machine)
			}
		}
	}
}
