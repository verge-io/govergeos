package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// NodeService handles node read operations.
type NodeService struct {
	client *Client
}

// List returns all nodes, with optional filtering and pagination.
func (s *NodeService) List(ctx context.Context, opts ...ListOption) ([]Node, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = nodeListFields
	}

	params := options.toQueryParams()

	var nodes []Node
	if err := s.client.get(ctx, "/nodes", params, &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

// ListPhysical returns all physical nodes.
func (s *NodeService) ListPhysical(ctx context.Context, opts ...ListOption) ([]Node, error) {
	// Prepend physical filter
	opts = append([]ListOption{WithFilter("physical eq true")}, opts...)
	return s.List(ctx, opts...)
}

// Get returns a single node by ID.
func (s *NodeService) Get(ctx context.Context, id int) (*Node, error) {
	params := url.Values{}
	params.Set("fields", nodeListFields)

	var node Node
	endpoint := fmt.Sprintf("/nodes/%d", id)
	if err := s.client.get(ctx, endpoint, params, &node); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Node", ID: id}
		}
		return nil, err
	}

	return &node, nil
}

// GetByName returns a node by name.
func (s *NodeService) GetByName(ctx context.Context, name string) (*Node, error) {
	nodes, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", name)))
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, &NotFoundError{Resource: "Node", ID: name}
	}

	return &nodes[0], nil
}

// GetDashboard returns detailed dashboard information for a node.
func (s *NodeService) GetDashboard(ctx context.Context, id int) (*Node, error) {
	params := url.Values{}
	params.Set("fields", "dashboard")

	var node Node
	endpoint := fmt.Sprintf("/nodes/%d", id)
	if err := s.client.get(ctx, endpoint, params, &node); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Node", ID: id}
		}
		return nil, err
	}

	return &node, nil
}
