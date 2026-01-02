package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// MediaSourceService handles media source read operations.
type MediaSourceService struct {
	client *Client
}

// List returns all media sources, with optional filtering and pagination.
func (s *MediaSourceService) List(ctx context.Context, opts ...ListOption) ([]MediaSource, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = mediaSourceListFields
	}

	params := options.toQueryParams()

	var sources []MediaSource
	if err := s.client.get(ctx, "/files", params, &sources); err != nil {
		return nil, err
	}

	return sources, nil
}

// Get returns a single media source by ID.
func (s *MediaSourceService) Get(ctx context.Context, id int) (*MediaSource, error) {
	params := url.Values{}
	params.Set("fields", mediaSourceListFields)

	var source MediaSource
	endpoint := fmt.Sprintf("/files/%d", id)
	if err := s.client.get(ctx, endpoint, params, &source); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "MediaSource", ID: id}
		}
		return nil, err
	}

	return &source, nil
}

// GetByName returns a media source by name.
func (s *MediaSourceService) GetByName(ctx context.Context, name string) (*MediaSource, error) {
	sources, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", name)))
	if err != nil {
		return nil, err
	}

	if len(sources) == 0 {
		return nil, &NotFoundError{Resource: "MediaSource", ID: name}
	}

	return &sources[0], nil
}

// ListISOs returns all ISO media sources.
func (s *MediaSourceService) ListISOs(ctx context.Context, opts ...ListOption) ([]MediaSource, error) {
	opts = append([]ListOption{WithFilter("type eq 'iso'")}, opts...)
	return s.List(ctx, opts...)
}
