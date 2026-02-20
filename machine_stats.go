package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// MachineStatsService handles machine statistics read operations.
// This service provides per-machine CPU, RAM, and temperature metrics.
type MachineStatsService struct {
	client *Client
}

// List returns all machine stats, with optional filtering.
func (s *MachineStatsService) List(ctx context.Context, opts ...ListOption) ([]MachineStats, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = machineStatsListFields
	}

	params := options.toQueryParams()

	var stats []MachineStats
	if err := s.client.get(ctx, "/machine_stats", params, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// GetByMachine retrieves stats for a specific machine by machine ID.
func (s *MachineStatsService) GetByMachine(ctx context.Context, machineID int) (*MachineStats, error) {
	params := url.Values{}
	params.Set("fields", machineStatsListFields)
	params.Set("filter", fmt.Sprintf("machine eq %d", machineID))

	var stats []MachineStats
	if err := s.client.get(ctx, "/machine_stats", params, &stats); err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, &NotFoundError{Resource: "MachineStats", ID: fmt.Sprintf("machine=%d", machineID)}
	}

	return &stats[0], nil
}

// Get returns a single machine stats record by ID.
func (s *MachineStatsService) Get(ctx context.Context, id int) (*MachineStats, error) {
	params := url.Values{}
	params.Set("fields", machineStatsListFields)

	var stats MachineStats
	endpoint := fmt.Sprintf("/machine_stats/%d", id)
	if err := s.client.get(ctx, endpoint, params, &stats); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "MachineStats", ID: id}
		}
		return nil, err
	}

	return &stats, nil
}
