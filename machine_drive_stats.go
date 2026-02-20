package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// MachineDriveStatsService handles machine drive I/O statistics read operations.
// This service provides per-drive reads, writes, throughput, and utilization metrics.
type MachineDriveStatsService struct {
	client *Client
}

// List returns all machine drive stats, with optional filtering.
func (s *MachineDriveStatsService) List(ctx context.Context, opts ...ListOption) ([]MachineDriveStats, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = machineDriveStatsListFields
	}

	params := options.toQueryParams()

	var stats []MachineDriveStats
	if err := s.client.get(ctx, "/machine_drive_stats", params, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// ListPhysical returns stats for physical drives only.
func (s *MachineDriveStatsService) ListPhysical(ctx context.Context, opts ...ListOption) ([]MachineDriveStats, error) {
	opts = append(opts, WithFilter("physical eq true"))
	return s.List(ctx, opts...)
}

// GetByDrive retrieves stats for a specific drive by parent drive ID.
func (s *MachineDriveStatsService) GetByDrive(ctx context.Context, driveID int) (*MachineDriveStats, error) {
	params := url.Values{}
	params.Set("fields", machineDriveStatsListFields)
	params.Set("filter", fmt.Sprintf("parent_drive eq %d", driveID))

	var stats []MachineDriveStats
	if err := s.client.get(ctx, "/machine_drive_stats", params, &stats); err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, &NotFoundError{Resource: "MachineDriveStats", ID: fmt.Sprintf("parent_drive=%d", driveID)}
	}

	return &stats[0], nil
}

// Get returns a single machine drive stats record by ID.
func (s *MachineDriveStatsService) Get(ctx context.Context, id int) (*MachineDriveStats, error) {
	params := url.Values{}
	params.Set("fields", machineDriveStatsListFields)

	var stats MachineDriveStats
	endpoint := fmt.Sprintf("/machine_drive_stats/%d", id)
	if err := s.client.get(ctx, endpoint, params, &stats); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "MachineDriveStats", ID: id}
		}
		return nil, err
	}

	return &stats, nil
}
