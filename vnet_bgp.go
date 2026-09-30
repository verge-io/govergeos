package vergeos

import (
	"context"
	"fmt"
)

const vnetBGPEndpoint = "/vnet_bgp"

// VNetBGPService handles the per-network dynamic routing record.
type VNetBGPService struct {
	client *Client
}

// List returns vnet_bgp rows, with optional filtering and pagination.
func (s *VNetBGPService) List(ctx context.Context, opts ...ListOption) ([]VNetBGP, error) {
	return listRows[VNetBGP](ctx, s.client, vnetBGPEndpoint, vnetBGPListFields, "", opts)
}

// ListByNetwork returns the vnet_bgp rows for a network.
func (s *VNetBGPService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetBGP, error) {
	if err := requireID("vnet", networkID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("vnet", networkID, opts)...)
}

// Get returns one vnet_bgp row by key.
func (s *VNetBGPService) Get(ctx context.Context, id int) (*VNetBGP, error) {
	return getRow[VNetBGP](ctx, s.client, vnetBGPEndpoint, id, vnetBGPGetFields, "VNetBGP")
}

// GetByNetwork returns the vnet_bgp row for a network.
func (s *VNetBGPService) GetByNetwork(ctx context.Context, networkID int) (*VNetBGP, error) {
	rows, err := s.ListByNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("vnet:%d", networkID)
	row, err := oneRoutingRow("VNetBGP", name, rows, func(row VNetBGP) any { return row.Key })
	if err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetBGP", ID: networkID}
		}
		return nil, err
	}
	return s.Get(ctx, int(row.Key))
}

// Create creates the vnet_bgp row for a network.
//
// The returned RoutingRestartStatus reports whether that network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the row
// is created but the restart follow-up fails, the row and status are still
// returned with the error.
func (s *VNetBGPService) Create(ctx context.Context, req *VNetBGPCreateRequest, opts ...RoutingRestartOption) (*VNetBGP, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("vnet", req.VNet); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID := int(row.VNet)
	if networkID <= 0 {
		networkID = req.VNet
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP config", "created", opts)
}

// GetOrCreate returns the network's vnet_bgp row, creating one when none exists.
//
// An existing row is returned with a nil status. A created row returns
// RoutingRestartStatus, and WithRestartNetwork restarts the network only in
// that case.
func (s *VNetBGPService) GetOrCreate(ctx context.Context, networkID int, opts ...RoutingRestartOption) (*VNetBGP, *RoutingRestartStatus, error) {
	if err := requireID("vnet", networkID); err != nil {
		return nil, nil, err
	}
	existing, err := s.GetByNetwork(ctx, networkID)
	if err == nil {
		return existing, nil, nil
	}
	if !IsNotFoundError(err) {
		return nil, nil, err
	}
	return s.Create(ctx, &VNetBGPCreateRequest{VNet: networkID}, opts...)
}

// Delete removes the vnet_bgp row and, on the server, the routing rows that
// belong to it.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the row
// is deleted but the restart follow-up fails, the status is still returned
// with the error.
func (s *VNetBGPService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID := int(row.VNet)
	if err := deleteRow(ctx, s.client, vnetBGPEndpoint, id, "VNetBGP"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP config", opts)
}

func oneRoutingRow[T any](resource, name string, rows []T, key func(T) any) (T, error) {
	var zero T
	if len(rows) == 0 {
		return zero, &NotFoundError{Resource: resource, ID: name}
	}
	if err := requireUniqueName(resource, name, rows, key); err != nil {
		return zero, err
	}
	return rows[0], nil
}
