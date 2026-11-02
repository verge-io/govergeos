//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// passthroughDeviceTypes are machine_devices.type values that draw from a
// resource group. Order prefers pools that are cheaper to borrow for a test
// than a physical GPU.
var passthroughDeviceTypes = []string{
	vergeos.DeviceTypeUSB,
	vergeos.DeviceTypeSRIOVNIC,
	vergeos.DeviceTypePCI,
	vergeos.DeviceTypeVGPU,
	vergeos.DeviceTypeVGPULegacy,
	vergeos.DeviceTypeHostGPU,
	vergeos.DeviceTypeGPULegacy,
}

func TestTenantNodeDevicesRead(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	node, ok := firstTenantNode(t, client, ctx)
	if !ok {
		t.Skip("No tenant node with a machine key")
	}
	t.Logf("tenant node %q key=%d machine=%d", node.Name, node.Key.Int(), node.Machine.Int())

	devices, err := client.TenantNodeDevices.List(ctx, node.Key.Int())
	if err != nil {
		t.Fatalf("TenantNodeDevices.List(%d) failed: %v", node.Key.Int(), err)
	}
	t.Logf("Found %d devices on tenant node %d", len(devices), node.Key.Int())
	if len(devices) == 0 {
		return
	}

	first := devices[0]
	got, err := client.TenantNodeDevices.Get(ctx, first.Key.Int())
	if err != nil {
		t.Fatalf("TenantNodeDevices.Get(%d) failed: %v", first.Key.Int(), err)
	}
	if got.Machine != node.Machine.Int() {
		t.Errorf("device machine %d, tenant node machine %d (node key %d)", got.Machine, node.Machine.Int(), node.Key.Int())
	}
	if got.MachineType != vergeos.DeviceMachineTypeTenant {
		t.Errorf("machine_type = %q, want %q", got.MachineType, vergeos.DeviceMachineTypeTenant)
	}
	t.Logf("device %d type=%s resource_group=%s machine_type=%s", got.Key.Int(), got.Type, got.ResourceGroup, got.MachineType)
}

func TestTenantNodeDevicesRejectVMDevice(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	vms, err := client.VMs.List(ctx, vergeos.WithFilter("is_snapshot eq false"), vergeos.WithLimit(20))
	if err != nil {
		t.Fatalf("VMs.List failed: %v", err)
	}
	for _, vm := range vms {
		devices, err := client.VMDevices.List(ctx, vm.Key.Int())
		if err != nil {
			t.Fatalf("VMDevices.List(%d) failed: %v", vm.Key.Int(), err)
		}
		if len(devices) == 0 {
			continue
		}
		device := devices[0]
		_, err = client.TenantNodeDevices.Get(ctx, device.Key.Int())
		if !vergeos.IsNotFoundError(err) {
			t.Fatalf("TenantNodeDevices.Get(VM device %d, machine_type %q) = %v, want not found", device.Key.Int(), device.MachineType, err)
		}
		t.Logf("VM device %d machine_type=%q is not returned as a tenant node device", device.Key.Int(), device.MachineType)
		return
	}
	t.Skip("No VM device to compare")
}

// TestTenantNodeDevicesAttach borrows one device from a passthrough resource
// group, checks it landed on the tenant node's machine, and deletes it.
// It skips when the lab has no tenant node or no passthrough resource group.
func TestTenantNodeDevicesAttach(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	node, ok := firstTenantNode(t, client, ctx)
	if !ok {
		t.Skip("No tenant node with a machine key")
	}
	group, ok := passthroughResourceGroup(t, client, ctx)
	if !ok {
		t.Skip("No passthrough resource group to attach")
	}

	name := fmt.Sprintf("govergeos-tnd-%d", time.Now().Unix())
	count := 1
	optional := true
	var createdID int
	t.Logf("attaching %s from resource group %s (%s) to tenant node %d (machine %d)", group.Type, group.Key, group.Name, node.Key.Int(), node.Machine.Int())

	t.Cleanup(func() {
		if createdID != 0 {
			if err := client.TenantNodeDevices.Delete(ctx, createdID); err != nil && !vergeos.IsNotFoundError(err) {
				t.Logf("cleanup delete %d: %v", createdID, err)
			}
		}
		devices, err := client.TenantNodeDevices.List(ctx, node.Key.Int())
		if err != nil {
			t.Logf("cleanup list: %v", err)
			return
		}
		for _, device := range devices {
			if device.Name != name && !strings.HasPrefix(device.Name, name) {
				continue
			}
			if err := client.TenantNodeDevices.Delete(ctx, device.Key.Int()); err != nil && !vergeos.IsNotFoundError(err) {
				t.Logf("cleanup delete %d: %v", device.Key.Int(), err)
			}
		}
	})

	device, err := client.TenantNodeDevices.Create(ctx, node.Key.Int(), &vergeos.VMDeviceCreateRequest{
		Name:          name,
		Type:          group.Type,
		ResourceGroup: vergeos.DeviceResourceGroup(group.Key),
		Count:         &count,
		Optional:      &optional,
		Description:   "govergeos tenant node device test",
	})
	if err != nil {
		if deviceAttachUnavailable(err) {
			t.Skipf("No spare device in resource group %s: %v", group.Name, err)
		}
		t.Fatalf("TenantNodeDevices.Create failed: %v", err)
	}
	createdID = device.Key.Int()
	if device.Machine != node.Machine.Int() {
		t.Errorf("created device machine %d, tenant node machine %d (node key %d)", device.Machine, node.Machine.Int(), node.Key.Int())
	}
	if device.MachineType != vergeos.DeviceMachineTypeTenant {
		t.Errorf("created machine_type = %q, want %q", device.MachineType, vergeos.DeviceMachineTypeTenant)
	}
	if device.ResourceGroup.String() != "" && device.ResourceGroup.String() != group.Key {
		t.Errorf("resource group %q, posted %q", device.ResourceGroup, group.Key)
	}

	got, err := client.TenantNodeDevices.Get(ctx, device.Key.Int())
	if err != nil {
		t.Fatalf("TenantNodeDevices.Get(%d) failed: %v", device.Key.Int(), err)
	}
	if got.Name != name {
		t.Errorf("Get name %q, want %q", got.Name, name)
	}

	renamed := name + "-renamed"
	updated, err := client.TenantNodeDevices.Update(ctx, device.Key.Int(), &vergeos.VMDeviceUpdateRequest{
		Name: &renamed,
	})
	if err != nil {
		t.Fatalf("TenantNodeDevices.Update failed: %v", err)
	}
	if updated.Name != renamed {
		t.Errorf("updated name %q, want %q", updated.Name, renamed)
	}

	if err := client.TenantNodeDevices.Delete(ctx, device.Key.Int()); err != nil {
		t.Fatalf("TenantNodeDevices.Delete failed: %v", err)
	}
	_, err = client.TenantNodeDevices.Get(ctx, device.Key.Int())
	if !vergeos.IsNotFoundError(err) {
		t.Fatalf("device %d still present after delete: %v", device.Key.Int(), err)
	}
}

func firstTenantNode(t *testing.T, client *vergeos.Client, ctx context.Context) (*vergeos.TenantNode, bool) {
	t.Helper()
	nodes, err := client.TenantNodes.List(ctx)
	if err != nil {
		t.Fatalf("TenantNodes.List failed: %v", err)
	}
	for i := range nodes {
		node := &nodes[i]
		if node.IsSnapshot || node.Key.Int() <= 0 || node.Machine.Int() <= 0 {
			continue
		}
		return node, true
	}
	return nil, false
}

func passthroughResourceGroup(t *testing.T, client *vergeos.Client, ctx context.Context) (*vergeos.ResourceGroup, bool) {
	t.Helper()
	groups, err := client.ResourceGroups.List(ctx)
	if err != nil {
		t.Fatalf("ResourceGroups.List failed: %v", err)
	}
	byType := map[string][]vergeos.ResourceGroup{}
	for _, group := range groups {
		if !group.Enabled || group.Key == "" || group.Type == "" {
			continue
		}
		byType[group.Type] = append(byType[group.Type], group)
	}
	for _, deviceType := range passthroughDeviceTypes {
		if len(byType[deviceType]) > 0 {
			group := byType[deviceType][0]
			return &group, true
		}
	}
	t.Logf("resource groups: %d, none of the passthrough types were enabled", len(groups))
	return nil, false
}

func deviceAttachUnavailable(err error) bool {
	msg := strings.ToLower(err.Error())
	for _, phrase := range []string{
		"no device",
		"no available",
		"none available",
		"not available",
		"not enough",
		"insufficient",
		"unavailable",
	} {
		if strings.Contains(msg, phrase) {
			return true
		}
	}
	return false
}
