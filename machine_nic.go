package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// MachineNICService handles machine NIC read operations.
// This service provides per-NIC traffic counters and link status for physical nodes.
type MachineNICService struct {
	client *Client
}

// List returns all machine NICs with stats and status expanded, with optional filtering.
func (s *MachineNICService) List(ctx context.Context, opts ...ListOption) ([]MachineNIC, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = machineNICListFields
	}

	params := options.toQueryParams()

	var nics []MachineNIC
	if err := s.client.get(ctx, "/machine_nics", params, &nics); err != nil {
		return nil, err
	}

	return nics, nil
}

// ListByMachine returns all NICs for a given machine, with stats and status expanded.
func (s *MachineNICService) ListByMachine(ctx context.Context, machineID int, opts ...ListOption) ([]MachineNIC, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("machine eq %d", machineID)))
	return s.List(ctx, opts...)
}

// Get returns a single machine NIC by ID.
func (s *MachineNICService) Get(ctx context.Context, id int) (*MachineNIC, error) {
	params := url.Values{}
	params.Set("fields", machineNICListFields)

	var nic MachineNIC
	endpoint := fmt.Sprintf("/machine_nics/%d", id)
	if err := s.client.get(ctx, endpoint, params, &nic); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "MachineNIC", ID: id}
		}
		return nil, err
	}

	return &nic, nil
}
