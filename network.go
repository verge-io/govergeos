package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

const (
	// Network action constants
	networkActionPowerOn  = "poweron"
	networkActionPowerOff = "poweroff"
	networkActionKill     = "killpower"
	networkActionReset    = "reset"
	networkActionApply    = "apply"
	networkActionApplyDNS = "applydns"

	// Network power state polling
	networkPowerStateMaxRetries   = 30
	networkPowerStatePollInterval = 5 * time.Second
)

// NetworkService handles network operations.
type NetworkService struct {
	client *Client
}

// List returns all networks, with optional filtering and pagination.
func (s *NetworkService) List(ctx context.Context, opts ...ListOption) ([]Network, error) {
	options := applyListOptions(opts)

	// Use network-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = networkListFields
	}

	params := options.toQueryParams()

	var networks []Network
	if err := s.client.get(ctx, "/vnets", params, &networks); err != nil {
		return nil, err
	}

	return networks, nil
}

// Get returns a single network by ID.
func (s *NetworkService) Get(ctx context.Context, id int) (*Network, error) {
	params := url.Values{}
	params.Set("fields", networkGetFields)

	var network Network
	endpoint := fmt.Sprintf("/vnets/%d", id)
	if err := s.client.get(ctx, endpoint, params, &network); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Network", ID: id}
		}
		return nil, err
	}

	return &network, nil
}

// Create creates a new network and returns the created network.
func (s *NetworkService) Create(ctx context.Context, req *NetworkCreateRequest) (*Network, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	// Set defaults
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnets", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created network's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created network
	return s.Get(ctx, id)
}

// Update updates a network and returns the updated network.
func (s *NetworkService) Update(ctx context.Context, id int, req *NetworkUpdateRequest) (*Network, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/vnets/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Network", ID: id}
		}
		return nil, err
	}

	// Read back the updated network
	return s.Get(ctx, id)
}

// Delete deletes a network.
func (s *NetworkService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vnets/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return nil // Already deleted
		}
		return err
	}
	return nil
}

// PowerOn powers on a network and waits for it to start.
func (s *NetworkService) PowerOn(ctx context.Context, id int) error {
	// Get current state
	network, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Already running
	if network.PowerState {
		return nil
	}

	// Power on is typically done by enabling the network
	enabled := true
	updateReq := &NetworkUpdateRequest{
		Enabled: &enabled,
	}

	_, err = s.Update(ctx, id, updateReq)
	if err != nil {
		return fmt.Errorf("vergeos: failed to power on network %d: %w", id, err)
	}

	// Wait for network to start
	return s.waitForPowerState(ctx, id, true)
}

// PowerOff powers off a network and waits for it to stop.
func (s *NetworkService) PowerOff(ctx context.Context, id int) error {
	// Get current state
	network, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Already stopped
	if !network.PowerState {
		return nil
	}

	// Send kill action
	action := vnetAction{
		VNet:   id,
		Action: networkActionKill,
		Params: struct{}{},
	}

	if err := s.client.post(ctx, "/vnet_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power off network %d: %w", id, err)
	}

	// Wait for network to stop
	return s.waitForPowerState(ctx, id, false)
}

// waitForPowerState waits for a network to reach the desired power state.
func (s *NetworkService) waitForPowerState(ctx context.Context, id int, desiredState bool) error {
	stateDesc := "stopped"
	if desiredState {
		stateDesc = "running"
	}

	for i := 0; i < networkPowerStateMaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(networkPowerStatePollInterval):
		}

		network, err := s.Get(ctx, id)
		if err != nil {
			return err
		}

		if network.PowerState == desiredState {
			return nil
		}
	}

	return fmt.Errorf("vergeos: timeout waiting for network %d to become %s", id, stateDesc)
}

// Kill forcefully powers off a network (hard power off).
func (s *NetworkService) Kill(ctx context.Context, id int) error {
	action := vnetAction{
		VNet:   id,
		Action: networkActionKill,
		Params: struct{}{},
	}

	if err := s.client.post(ctx, "/vnet_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to kill network %d: %w", id, err)
	}

	return nil
}

// Reset restarts a network.
func (s *NetworkService) Reset(ctx context.Context, id int, applyFirewall bool) error {
	params := struct {
		Apply bool `json:"apply"`
	}{
		Apply: applyFirewall,
	}

	action := vnetAction{
		VNet:   id,
		Action: networkActionReset,
		Params: params,
	}

	if err := s.client.post(ctx, "/vnet_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to reset network %d: %w", id, err)
	}

	return nil
}

// ApplyRules applies firewall rules to a running network.
func (s *NetworkService) ApplyRules(ctx context.Context, id int) error {
	action := vnetAction{
		VNet:   id,
		Action: networkActionApply,
		Params: struct{}{},
	}

	if err := s.client.post(ctx, "/vnet_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to apply rules to network %d: %w", id, err)
	}

	return nil
}

// ApplyDNS applies DNS configuration to a running network.
func (s *NetworkService) ApplyDNS(ctx context.Context, id int) error {
	action := vnetAction{
		VNet:   id,
		Action: networkActionApplyDNS,
		Params: struct{}{},
	}

	if err := s.client.post(ctx, "/vnet_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to apply DNS to network %d: %w", id, err)
	}

	return nil
}
