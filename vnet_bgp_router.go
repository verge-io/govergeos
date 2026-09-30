package vergeos

import (
	"context"
	"strconv"
)

const vnetBGPRouterEndpoint = "/vnet_bgp_routers"

// VNetBGPRouterService handles BGP routers.
type VNetBGPRouterService struct {
	client *Client
}

// List returns BGP routers, with optional filtering and pagination.
// Rows are ordered by ASN unless WithSort sets another order.
func (s *VNetBGPRouterService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPRouter, error) {
	return listRows[VNetBGPRouter](ctx, s.client, vnetBGPRouterEndpoint, vnetBGPRouterListFields, vnetBGPRouterSort, opts)
}

// ListByBGP returns BGP routers for one vnet_bgp record.
func (s *VNetBGPRouterService) ListByBGP(ctx context.Context, bgpID int, opts ...ListOption) ([]VNetBGPRouter, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp", bgpID, opts)...)
}

// ListByNetwork returns BGP routers for a network.
// A network with no vnet_bgp row returns an empty list.
func (s *VNetBGPRouterService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetBGPRouter, error) {
	bgpID, ok, err := s.client.bgpKeyForNetwork(ctx, networkID)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return []VNetBGPRouter{}, nil
	}
	return s.ListByBGP(ctx, bgpID, opts...)
}

// Get returns one BGP router by key.
func (s *VNetBGPRouterService) Get(ctx context.Context, id int) (*VNetBGPRouter, error) {
	return getRow[VNetBGPRouter](ctx, s.client, vnetBGPRouterEndpoint, id, vnetBGPRouterGetFields, "VNetBGPRouter")
}

// GetByASN returns the BGP router with an ASN on one vnet_bgp record.
func (s *VNetBGPRouterService) GetByASN(ctx context.Context, bgpID int, asn int) (*VNetBGPRouter, error) {
	if err := validateBGPASN(asn); err != nil {
		return nil, err
	}
	rows, err := s.ListByBGP(ctx, bgpID, WithFilter("asn eq "+strconv.Itoa(asn)))
	if err != nil {
		return nil, err
	}
	matched := make([]VNetBGPRouter, 0, len(rows))
	for _, row := range rows {
		if row.ASN == asn {
			matched = append(matched, row)
		}
	}
	name := strconv.Itoa(asn)
	row, err := oneRoutingRow("VNetBGPRouter", name, matched, func(row VNetBGPRouter) any { return row.Key })
	if err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetBGPRouter", ID: asn}
		}
		return nil, err
	}
	got, err := s.Get(ctx, int(row.Key))
	if err != nil {
		return nil, err
	}
	if got.ASN != asn {
		return nil, &NotFoundError{Resource: "VNetBGPRouter", ID: asn}
	}
	return got, nil
}

// Create creates a BGP router under a vnet_bgp record.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// router is created but the restart follow-up fails, the router and status
// are still returned with the error.
func (s *VNetBGPRouterService) Create(ctx context.Context, req *VNetBGPRouterCreateRequest, opts ...RoutingRestartOption) (*VNetBGPRouter, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp", req.BGP); err != nil {
		return nil, nil, err
	}
	if err := validateBGPASN(req.ASN); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPRouterEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP router", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP router", "created", opts)
}

// Update updates a BGP router.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// router is updated but the restart follow-up fails, the router and status
// are still returned with the error.
func (s *VNetBGPRouterService) Update(ctx context.Context, id int, req *VNetBGPRouterUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPRouter, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if req.ASN != nil {
		if err := validateBGPASN(*req.ASN); err != nil {
			return nil, nil, err
		}
	}
	if err := putRow(ctx, s.client, vnetBGPRouterEndpoint, id, req, "VNetBGPRouter"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP router", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP router", "updated", opts)
}

// Delete removes a BGP router and, on the server, its commands.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPRouterService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPRouterEndpoint, id, "VNetBGPRouter"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP router", opts)
}

const vnetBGPRouterCommandEndpoint = "/vnet_bgp_router_commands"

// VNetBGPRouterCommandService handles BGP router commands.
type VNetBGPRouterCommandService struct {
	client *Client
}

// List returns BGP router commands, with optional filtering and pagination.
// Rows are ordered by orderid unless WithSort sets another order.
func (s *VNetBGPRouterCommandService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPRouterCommand, error) {
	return listRows[VNetBGPRouterCommand](ctx, s.client, vnetBGPRouterCommandEndpoint, vnetBGPRouterCommandListFields, vnetBGPCommandSort, opts)
}

// ListByRouter returns commands for one BGP router.
func (s *VNetBGPRouterCommandService) ListByRouter(ctx context.Context, routerID int, opts ...ListOption) ([]VNetBGPRouterCommand, error) {
	if err := requireID("bgp_router", routerID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp_router", routerID, opts)...)
}

// Get returns one BGP router command by key.
func (s *VNetBGPRouterCommandService) Get(ctx context.Context, id int) (*VNetBGPRouterCommand, error) {
	return getRow[VNetBGPRouterCommand](ctx, s.client, vnetBGPRouterCommandEndpoint, id, vnetBGPRouterCommandGetFields, "VNetBGPRouterCommand")
}

// Create creates a BGP router command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is created but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPRouterCommandService) Create(ctx context.Context, req *VNetBGPRouterCommandCreateRequest, opts ...RoutingRestartOption) (*VNetBGPRouterCommand, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp_router", req.BGPRouter); err != nil {
		return nil, nil, err
	}
	if err := requireText("command", req.Command); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPRouterCommandEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGPRouter(ctx, int(row.BGPRouter))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP router command", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP router command", "created", opts)
}

// Update updates a BGP router command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is updated but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPRouterCommandService) Update(ctx context.Context, id int, req *VNetBGPRouterCommandUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPRouterCommand, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("command", req.Command); err != nil {
		return nil, nil, err
	}
	if err := putRow(ctx, s.client, vnetBGPRouterCommandEndpoint, id, req, "VNetBGPRouterCommand"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGPRouter(ctx, int(row.BGPRouter))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP router command", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP router command", "updated", opts)
}

// Delete removes a BGP router command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPRouterCommandService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGPRouter(ctx, int(row.BGPRouter))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPRouterCommandEndpoint, id, "VNetBGPRouterCommand"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP router command", opts)
}
