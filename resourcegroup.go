package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// ResourceGroupService handles resource group read operations.
type ResourceGroupService struct {
	client *Client
}

// List returns all resource groups, with optional filtering and pagination.
func (s *ResourceGroupService) List(ctx context.Context, opts ...ListOption) ([]ResourceGroup, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = resourceGroupListFields
	}

	params := options.toQueryParams()

	var groups []ResourceGroup
	if err := s.client.get(ctx, "/resource_groups", params, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// Get returns a single resource group by ID.
func (s *ResourceGroupService) Get(ctx context.Context, id int) (*ResourceGroup, error) {
	params := url.Values{}
	params.Set("fields", resourceGroupListFields)

	var group ResourceGroup
	endpoint := fmt.Sprintf("/resource_groups/%d", id)
	if err := s.client.get(ctx, endpoint, params, &group); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "ResourceGroup", ID: id}
		}
		return nil, err
	}

	return &group, nil
}

// GetByName returns a resource group by name.
func (s *ResourceGroupService) GetByName(ctx context.Context, name string) (*ResourceGroup, error) {
	groups, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, &NotFoundError{Resource: "ResourceGroup", ID: name}
	}
	if err := requireUniqueName("ResourceGroup", name, groups, func(g ResourceGroup) any { return g.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("ResourceGroup", name, groups[0].Name, name); err != nil {
		return nil, err
	}

	return &groups[0], nil
}
