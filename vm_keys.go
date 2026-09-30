package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// machineKeyForVM returns the machine key (vms.machine) for a VM $key.
//
// Drives, NICs, devices, and snapshots are stored against the machine key.
// vm_actions and VM.Get use the VM $key. On a system that has been in use
// the two numbers differ, and each is often some other object's key.
func (c *Client) machineKeyForVM(ctx context.Context, vmID int) (int, error) {
	if vmID <= 0 {
		return 0, &ValidationError{Field: "vm", Message: "VM ID is required"}
	}

	params := url.Values{}
	params.Set("fields", "$key,machine")

	var vm VM
	endpoint := fmt.Sprintf("/vms/%d", vmID)
	if err := c.get(ctx, endpoint, params, &vm); err != nil {
		if IsNotFoundError(err) {
			return 0, &NotFoundError{Resource: "VM", ID: vmID}
		}
		return 0, err
	}
	if vm.Machine <= 0 {
		return 0, &ValidationError{Field: "machine", Message: fmt.Sprintf("VM %d has no machine key", vmID)}
	}
	return vm.Machine, nil
}

// vmKeyForMachine returns the vms.$key of the VM whose machine field equals
// machineKey. snapshot selects the snapshot's own VM row (is_snapshot true),
// which is what vm_actions restore must target. The live VM is the row with
// is_snapshot false, which is what hot-unplug must target.
//
// A row whose machine field is present and does not match machineKey is
// ignored, so a key collision cannot select another VM's row.
func (c *Client) vmKeyForMachine(ctx context.Context, machineKey int, snapshot bool) (int, error) {
	if machineKey <= 0 {
		return 0, &ValidationError{Field: "machine", Message: "machine key is required"}
	}

	params := url.Values{}
	params.Set("fields", "$key,machine,is_snapshot")
	params.Set("filter", fmt.Sprintf("machine eq %d", machineKey))

	var vms []VM
	if err := c.get(ctx, "/vms", params, &vms); err != nil {
		return 0, err
	}

	for _, vm := range vms {
		if vm.IsSnapshot != snapshot {
			continue
		}
		if vm.Machine != 0 && vm.Machine != machineKey {
			continue
		}
		id := vm.Key.Int()
		if id <= 0 {
			continue
		}
		return id, nil
	}

	resource := "VM"
	if snapshot {
		resource = "snapshot VM"
	}
	return 0, &NotFoundError{Resource: resource, ID: machineKey}
}
