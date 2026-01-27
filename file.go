package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// FileService handles file read operations.
type FileService struct {
	client *Client
}

// List returns all files, with optional filtering and pagination.
func (s *FileService) List(ctx context.Context, opts ...ListOption) ([]File, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = fileListFields
	}

	params := options.toQueryParams()

	var files []File
	if err := s.client.get(ctx, "/files", params, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// Get returns a single file by ID.
func (s *FileService) Get(ctx context.Context, id int) (*File, error) {
	params := url.Values{}
	params.Set("fields", fileListFields)

	var file File
	endpoint := fmt.Sprintf("/files/%d", id)
	if err := s.client.get(ctx, endpoint, params, &file); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "File", ID: id}
		}
		return nil, err
	}

	return &file, nil
}

// GetByName returns a file by name.
func (s *FileService) GetByName(ctx context.Context, name string) (*File, error) {
	files, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", name)))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, &NotFoundError{Resource: "File", ID: name}
	}

	return &files[0], nil
}

// ListISOs returns all ISO files.
func (s *FileService) ListISOs(ctx context.Context, opts ...ListOption) ([]File, error) {
	opts = append([]ListOption{WithFilter("type eq 'iso'")}, opts...)
	return s.List(ctx, opts...)
}
