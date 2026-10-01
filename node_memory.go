package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// NodeMemoryService reads DIMM inventory (node_memory).
//
// Rows are discovered by the platform. IsHealthy on a row is true when
// status is online.
type NodeMemoryService struct {
	client *Client
}

// List returns DIMM rows, with optional filtering and pagination.
func (s *NodeMemoryService) List(ctx context.Context, opts ...ListOption) ([]NodeMemory, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeMemoryListFields
	}

	var dimms []NodeMemory
	if err := s.client.get(ctx, "/node_memory", options.toQueryParams(), &dimms); err != nil {
		return nil, err
	}
	return dimms, nil
}

// ListByNode returns DIMM rows for one node.
func (s *NodeMemoryService) ListByNode(ctx context.Context, nodeID int, opts ...ListOption) ([]NodeMemory, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node eq %d", nodeID)))
	return s.List(ctx, opts...)
}

// ListUnhealthy returns DIMM rows whose status is not online.
func (s *NodeMemoryService) ListUnhealthy(ctx context.Context, opts ...ListOption) ([]NodeMemory, error) {
	opts = append(opts, WithFilter("status ne 'online'"))
	return s.List(ctx, opts...)
}

// Get returns one DIMM row by key.
func (s *NodeMemoryService) Get(ctx context.Context, id int) (*NodeMemory, error) {
	params := url.Values{}
	params.Set("fields", nodeMemoryListFields)

	var dimm NodeMemory
	endpoint := fmt.Sprintf("/node_memory/%d", id)
	if err := s.client.get(ctx, endpoint, params, &dimm); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "NodeMemory", ID: id}
		}
		return nil, err
	}
	return &dimm, nil
}
