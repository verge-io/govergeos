package vergeos

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// TimeSettingsService reads and updates the time settings singleton.
// The table exists on VergeOS 26.2 and later.
type TimeSettingsService struct {
	client *Client
}

// Get returns the time settings row. The key is "self".
func (s *TimeSettingsService) Get(ctx context.Context) (*TimeSettings, error) {
	if err := s.client.require262("time settings"); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("fields", timeSettingsFields)

	var settings TimeSettings
	if err := s.client.get(ctx, timeSettingsPath, params, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// Update writes the fields that are set and returns the stored row.
// A request with no fields returns the current row and does not PUT.
func (s *TimeSettingsService) Update(ctx context.Context, req *TimeSettingsUpdateRequest) (*TimeSettings, error) {
	if err := s.client.require262("time settings"); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if !req.empty() {
		if err := s.client.put(ctx, timeSettingsPath, req, nil); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx)
}

// TimeDashboardService reads the time dashboard view.
type TimeDashboardService struct {
	client *Client
}

// Get returns the time dashboard.
// An array response is treated as the per-node sync list.
func (s *TimeDashboardService) Get(ctx context.Context) (*TimeDashboard, error) {
	if err := s.client.require262("time dashboard"); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("fields", timeDashboardFields)

	var raw json.RawMessage
	if err := s.client.get(ctx, "/time_dashboard", params, &raw); err != nil {
		return nil, err
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return &TimeDashboard{}, nil
	}
	if raw[0] == '[' {
		var nodes []TimeNodeSync
		if err := json.Unmarshal(raw, &nodes); err != nil {
			return nil, err
		}
		return &TimeDashboard{Nodes: nodes}, nil
	}
	var dash TimeDashboard
	if err := json.Unmarshal(raw, &dash); err != nil {
		return nil, err
	}
	return &dash, nil
}
