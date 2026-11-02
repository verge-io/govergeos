package vergeos

import "context"

// TenantNodeDeviceService attaches host devices to a tenant node.
//
// Rows live in machine_devices, the same table VMDevices uses. List and
// Create take the tenant node $key and resolve tenant_nodes.machine before
// they touch that table. Passing the node $key to VMDevices looks the key
// up in /vms instead.
//
// A device passed to a tenant is thick provisioned. The tenant owns it
// even when no VM inside the tenant is using it, so it cannot be assigned
// to another VM or tenant until it is removed.
type TenantNodeDeviceService struct {
	client *Client
}

func (s *TenantNodeDeviceService) devices() *VMDeviceService {
	return &VMDeviceService{client: s.client}
}

// List returns devices attached to a tenant node.
// tenantNodeID is the tenant node $key (TenantNode.Key).
func (s *TenantNodeDeviceService) List(ctx context.Context, tenantNodeID int) ([]VMDevice, error) {
	machine, err := s.client.machineKeyForTenantNode(ctx, tenantNodeID)
	if err != nil {
		return nil, err
	}
	return s.devices().listByMachine(ctx, machine)
}

// Get returns a device by ID.
// A device whose machine_type is not tenant is reported as not found.
// VMDevices.Get is the call for a VM device.
func (s *TenantNodeDeviceService) Get(ctx context.Context, deviceID int) (*VMDevice, error) {
	device, err := s.devices().Get(ctx, deviceID)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantNodeDevice", ID: deviceID}
		}
		return nil, err
	}
	if device.MachineType != DeviceMachineTypeTenant {
		return nil, &NotFoundError{Resource: "TenantNodeDevice", ID: deviceID}
	}
	return device, nil
}

// Create attaches a device to a tenant node and returns it.
// tenantNodeID is the tenant node $key. The machine key is resolved from
// tenant_nodes and written on the device. Name and type are required.
//
// The tenant owns the device once it is attached, including while nothing
// inside the tenant is using it.
func (s *TenantNodeDeviceService) Create(ctx context.Context, tenantNodeID int, req *VMDeviceCreateRequest) (*VMDevice, error) {
	if err := validateDeviceCreate(req); err != nil {
		return nil, err
	}
	machine, err := s.client.machineKeyForTenantNode(ctx, tenantNodeID)
	if err != nil {
		return nil, err
	}
	return s.devices().createOnMachine(ctx, machine, req)
}

// Update updates a device and returns it.
// A device whose machine_type is not tenant is left unchanged and reported
// as not found.
func (s *TenantNodeDeviceService) Update(ctx context.Context, deviceID int, req *VMDeviceUpdateRequest) (*VMDevice, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if _, err := s.Get(ctx, deviceID); err != nil {
		return nil, err
	}
	return s.devices().Update(ctx, deviceID, req)
}

// Delete removes a device.
// A device whose machine_type is not tenant is left in place and reported
// as not found.
func (s *TenantNodeDeviceService) Delete(ctx context.Context, deviceID int) error {
	if _, err := s.Get(ctx, deviceID); err != nil {
		return err
	}
	return s.devices().Delete(ctx, deviceID)
}
