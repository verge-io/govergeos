package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// NodeLLDPNeighborService reads LLDP neighbors (node_lldp_neighbors).
//
// Neighbors are discovered on the node and are read-only. They show which
// switch port each NIC is cabled to.
type NodeLLDPNeighborService struct {
	client *Client
}

// List returns LLDP neighbors, with optional filtering and pagination.
func (s *NodeLLDPNeighborService) List(ctx context.Context, opts ...ListOption) ([]NodeLLDPNeighbor, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeLLDPNeighborListFields
	}

	var neighbors []NodeLLDPNeighbor
	if err := s.client.get(ctx, "/node_lldp_neighbors", options.toQueryParams(), &neighbors); err != nil {
		return nil, err
	}
	return neighbors, nil
}

// ListByNode returns LLDP neighbors discovered on one node.
func (s *NodeLLDPNeighborService) ListByNode(ctx context.Context, nodeID int, opts ...ListOption) ([]NodeLLDPNeighbor, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node eq %d", nodeID)))
	return s.List(ctx, opts...)
}

// ListByNIC returns LLDP neighbors discovered on one NIC.
func (s *NodeLLDPNeighborService) ListByNIC(ctx context.Context, nicID int, opts ...ListOption) ([]NodeLLDPNeighbor, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("nic eq %d", nicID)))
	return s.List(ctx, opts...)
}

// Get returns one LLDP neighbor by key.
func (s *NodeLLDPNeighborService) Get(ctx context.Context, id int) (*NodeLLDPNeighbor, error) {
	params := url.Values{}
	params.Set("fields", nodeLLDPNeighborListFields)

	var neighbor NodeLLDPNeighbor
	endpoint := fmt.Sprintf("/node_lldp_neighbors/%d", id)
	if err := s.client.get(ctx, endpoint, params, &neighbor); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeLLDPNeighbor", ID: id}
		}
		return nil, err
	}
	return &neighbor, nil
}
