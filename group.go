package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// GroupService handles group read operations.
type GroupService struct {
	client *Client
}

// List returns all groups, with optional filtering and pagination.
func (s *GroupService) List(ctx context.Context, opts ...ListOption) ([]Group, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = groupListFields
	}

	params := options.toQueryParams()

	var groups []Group
	if err := s.client.get(ctx, "/groups", params, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// Get returns a single group by ID.
func (s *GroupService) Get(ctx context.Context, id int) (*Group, error) {
	params := url.Values{}
	params.Set("fields", groupListFields)

	var group Group
	endpoint := fmt.Sprintf("/groups/%d", id)
	if err := s.client.get(ctx, endpoint, params, &group); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Group", ID: id}
		}
		return nil, err
	}

	return &group, nil
}

// GetByName returns a group by name.
func (s *GroupService) GetByName(ctx context.Context, name string) (*Group, error) {
	groups, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", name)))
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, &NotFoundError{Resource: "Group", ID: name}
	}

	return &groups[0], nil
}
