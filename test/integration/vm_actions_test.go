//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestVMGuestReboot sends a graceful reset to the VM identified by
// VERGEOS_TEST_VM_ID. The VM must already be running. The call reboots
// that guest, so it is skipped unless the variable is set.
func TestVMGuestReboot(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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
	if !vm.PowerState {
		t.Skip("VM is not running; guest reboot requires a running guest")
	}

	if err := client.VMs.GuestReboot(ctx, vmID); err != nil {
		t.Fatalf("VMs.GuestReboot(%d) failed: %v", vmID, err)
	}
}
