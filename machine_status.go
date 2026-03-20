package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// MachineStatusService handles machine status operations.
type MachineStatusService struct {
	client *Client
}

// List returns all machine statuses.
func (s *MachineStatusService) List(ctx context.Context, opts ...ListOption) ([]MachineStatus, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = machineStatusListFields
	}
	params := options.toQueryParams()

	var statuses []MachineStatus
	if err := s.client.get(ctx, "/machine_status", params, &statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}

// Get returns the status for a specific machine by its machine key.
func (s *MachineStatusService) Get(ctx context.Context, machineKey int) (*MachineStatus, error) {
	// machine_status is keyed by machine, so we filter by machine eq <key>.
	params := url.Values{}
	params.Set("fields", machineStatusGetFields)
	params.Set("filter", fmt.Sprintf("machine eq %d", machineKey))

	var statuses []MachineStatus
	if err := s.client.get(ctx, "/machine_status", params, &statuses); err != nil {
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, &NotFoundError{Resource: "MachineStatus", ID: machineKey}
	}

	return &statuses[0], nil
}

// GetByKey returns the status by its direct row key.
func (s *MachineStatusService) GetByKey(ctx context.Context, key int) (*MachineStatus, error) {
	params := url.Values{}
	params.Set("fields", machineStatusGetFields)

	var status MachineStatus
	endpoint := fmt.Sprintf("/machine_status/%d", key)
	if err := s.client.get(ctx, endpoint, params, &status); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "MachineStatus", ID: key}
		}
		return nil, err
	}

	return &status, nil
}
