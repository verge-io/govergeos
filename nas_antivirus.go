package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// NASServiceAntivirusService reads and updates antivirus settings on NAS
// services (vm_service_antivirus).
type NASServiceAntivirusService struct {
	client *Client
}

// List returns NAS service antivirus rows, with optional filtering and pagination.
func (s *NASServiceAntivirusService) List(ctx context.Context, opts ...ListOption) ([]NASServiceAntivirus, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nasServiceAntivirusListFields
	}

	var rows []NASServiceAntivirus
	if err := s.client.get(ctx, "/vm_service_antivirus", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByService returns antivirus settings for one NAS service.
func (s *NASServiceAntivirusService) ListByService(ctx context.Context, serviceID int, opts ...ListOption) ([]NASServiceAntivirus, error) {
	if serviceID <= 0 {
		return nil, &ValidationError{Field: "service", Message: "service is required"}
	}
	opts = append([]ListOption{WithFilter(fmt.Sprintf("service eq %d", serviceID))}, opts...)
	return s.List(ctx, opts...)
}

// Get returns one antivirus settings row by key.
func (s *NASServiceAntivirusService) Get(ctx context.Context, id int) (*NASServiceAntivirus, error) {
	params := url.Values{}
	params.Set("fields", nasServiceAntivirusListFields)

	var row NASServiceAntivirus
	endpoint := fmt.Sprintf("/vm_service_antivirus/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "NASServiceAntivirus", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

// GetByService returns the antivirus settings row for one NAS service.
//
// The table keeps one row per service.
func (s *NASServiceAntivirusService) GetByService(ctx context.Context, serviceID int) (*NASServiceAntivirus, error) {
	rows, err := s.ListByService(ctx, serviceID, WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "NASServiceAntivirus", ID: serviceID}
	}
	return &rows[0], nil
}

// Update changes antivirus settings and returns the stored row.
//
// A request with no fields reads the row back and does not PUT.
func (s *NASServiceAntivirusService) Update(ctx context.Context, id int, req *NASServiceAntivirusUpdateRequest) (*NASServiceAntivirus, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if id <= 0 {
		return nil, &ValidationError{Field: "id", Message: "antivirus settings key is required"}
	}
	if req.MaxRecursion != nil && (*req.MaxRecursion < 0 || *req.MaxRecursion > 100) {
		return nil, &ValidationError{Field: "max_recursion", Message: "max_recursion must be between 0 and 100"}
	}
	if nasAntivirusUpdateEmpty(req) {
		return s.Get(ctx, id)
	}

	endpoint := fmt.Sprintf("/vm_service_antivirus/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func nasAntivirusUpdateEmpty(req *NASServiceAntivirusUpdateRequest) bool {
	return req.Enabled == nil &&
		req.MaxRecursion == nil &&
		req.DatabasePrivateMirror == nil &&
		req.DatabaseLocation == nil &&
		req.DatabaseUpdatesEnabled == nil
}
