package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

const (
	// Power action constants
	vmActionPowerOn = "poweron"
	vmActionKill    = "kill"

	// Polling configuration
	powerStateMaxRetries   = 30
	powerStatePollInterval = 5 * time.Second
)

// VMService handles VM operations.
type VMService struct {
	client *Client
}

// List returns all VMs, with optional filtering and pagination.
func (s *VMService) List(ctx context.Context, opts ...ListOption) ([]VM, error) {
	options := applyListOptions(opts)

	// Use VM-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = vmListFields
	}

	params := options.toQueryParams()

	var vms []VM
	if err := s.client.get(ctx, "/vms", params, &vms); err != nil {
		return nil, err
	}

	return vms, nil
}

// Get returns a single VM by ID.
func (s *VMService) Get(ctx context.Context, id int) (*VM, error) {
	params := url.Values{}
	params.Set("fields", vmGetFields)

	var vm VM
	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.get(ctx, endpoint, params, &vm); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VM", ID: id}
		}
		return nil, err
	}

	return &vm, nil
}

// Create creates a new VM and returns the created VM.
func (s *VMService) Create(ctx context.Context, req *VMCreateRequest) (*VM, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.CPUCores <= 0 {
		return nil, &ValidationError{Field: "cpu_cores", Message: "cpu_cores must be positive"}
	}
	if req.RAM <= 0 {
		return nil, &ValidationError{Field: "ram", Message: "ram must be positive"}
	}

	// Set defaults
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vms", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created VM's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created VM
	return s.Get(ctx, id)
}

// Update updates a VM and returns the updated VM.
func (s *VMService) Update(ctx context.Context, id int, req *VMUpdateRequest) (*VM, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VM", ID: id}
		}
		return nil, err
	}

	// Read back the updated VM
	return s.Get(ctx, id)
}

// Delete deletes a VM.
func (s *VMService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VM", ID: id}
		}
		return err
	}
	return nil
}

// PowerOn powers on a VM and waits for it to start.
func (s *VMService) PowerOn(ctx context.Context, id int) error {
	// Get current state
	vm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Already running
	if vm.PowerState {
		return nil
	}

	// Send power on action
	action := vmAction{
		VM:     id,
		Action: vmActionPowerOn,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power on VM %d: %w", id, err)
	}

	// Wait for VM to start
	return s.waitForPowerState(ctx, id, true)
}

// PowerOff powers off a VM and waits for it to stop.
func (s *VMService) PowerOff(ctx context.Context, id int) error {
	// Get current state
	vm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Already stopped
	if !vm.PowerState {
		return nil
	}

	// Send kill action
	action := vmAction{
		VM:     id,
		Action: vmActionKill,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power off VM %d: %w", id, err)
	}

	// Wait for VM to stop
	return s.waitForPowerState(ctx, id, false)
}

// waitForPowerState waits for a VM to reach the desired power state.
func (s *VMService) waitForPowerState(ctx context.Context, id int, desiredState bool) error {
	stateStr := "stopped"
	if desiredState {
		stateStr = "running"
	}

	for i := 0; i < powerStateMaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(powerStatePollInterval):
		}

		vm, err := s.Get(ctx, id)
		if err != nil {
			return err
		}

		if vm.PowerState == desiredState {
			return nil
		}
	}

	return fmt.Errorf("vergeos: timeout waiting for VM %d to become %s", id, stateStr)
}
