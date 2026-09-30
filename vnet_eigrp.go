package vergeos

import (
	"context"
	"strconv"
)

const vnetEIGRPRouterEndpoint = "/vnet_eigrp_routers"

// VNetEIGRPRouterService handles EIGRP routers.
type VNetEIGRPRouterService struct {
	client *Client
}

// List returns EIGRP routers, with optional filtering and pagination.
// Rows are ordered by ASN unless WithSort sets another order.
func (s *VNetEIGRPRouterService) List(ctx context.Context, opts ...ListOption) ([]VNetEIGRPRouter, error) {
	return listRows[VNetEIGRPRouter](ctx, s.client, vnetEIGRPRouterEndpoint, vnetEIGRPRouterListFields, vnetEIGRPRouterSort, opts)
}

// ListByBGP returns EIGRP routers for one vnet_bgp record.
func (s *VNetEIGRPRouterService) ListByBGP(ctx context.Context, bgpID int, opts ...ListOption) ([]VNetEIGRPRouter, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp", bgpID, opts)...)
}

// ListByNetwork returns EIGRP routers for a network.
// A network with no vnet_bgp row returns an empty list.
func (s *VNetEIGRPRouterService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetEIGRPRouter, error) {
	bgpID, ok, err := s.client.bgpKeyForNetwork(ctx, networkID)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return []VNetEIGRPRouter{}, nil
	}
	return s.ListByBGP(ctx, bgpID, opts...)
}

// Get returns one EIGRP router by key.
func (s *VNetEIGRPRouterService) Get(ctx context.Context, id int) (*VNetEIGRPRouter, error) {
	return getRow[VNetEIGRPRouter](ctx, s.client, vnetEIGRPRouterEndpoint, id, vnetEIGRPRouterGetFields, "VNetEIGRPRouter")
}

// GetByASN returns the EIGRP router with an ASN on one vnet_bgp record.
func (s *VNetEIGRPRouterService) GetByASN(ctx context.Context, bgpID int, asn int) (*VNetEIGRPRouter, error) {
	if err := validateEIGRPASN(asn); err != nil {
		return nil, err
	}
	rows, err := s.ListByBGP(ctx, bgpID, WithFilter("asn eq "+strconv.Itoa(asn)))
	if err != nil {
		return nil, err
	}
	matched := make([]VNetEIGRPRouter, 0, len(rows))
	for _, row := range rows {
		if row.ASN == asn {
			matched = append(matched, row)
		}
	}
	row, err := oneRoutingRow("VNetEIGRPRouter", strconv.Itoa(asn), matched, func(row VNetEIGRPRouter) any { return row.Key })
	if err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetEIGRPRouter", ID: asn}
		}
		return nil, err
	}
	got, err := s.Get(ctx, int(row.Key))
	if err != nil {
		return nil, err
	}
	if got.ASN != asn {
		return nil, &NotFoundError{Resource: "VNetEIGRPRouter", ID: asn}
	}
	return got, nil
}

// Create creates an EIGRP router under a vnet_bgp record.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// router is created but the restart follow-up fails, the router and status
// are still returned with the error.
func (s *VNetEIGRPRouterService) Create(ctx context.Context, req *VNetEIGRPRouterCreateRequest, opts ...RoutingRestartOption) (*VNetEIGRPRouter, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp", req.BGP); err != nil {
		return nil, nil, err
	}
	if err := validateEIGRPASN(req.ASN); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetEIGRPRouterEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("EIGRP router", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "EIGRP router", "created", opts)
}

// Update updates an EIGRP router.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// router is updated but the restart follow-up fails, the router and status
// are still returned with the error.
func (s *VNetEIGRPRouterService) Update(ctx context.Context, id int, req *VNetEIGRPRouterUpdateRequest, opts ...RoutingRestartOption) (*VNetEIGRPRouter, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if req.ASN != nil {
		if err := validateEIGRPASN(*req.ASN); err != nil {
			return nil, nil, err
		}
	}
	if err := putRow(ctx, s.client, vnetEIGRPRouterEndpoint, id, req, "VNetEIGRPRouter"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("EIGRP router", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "EIGRP router", "updated", opts)
}

// Delete removes an EIGRP router and, on the server, its commands.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetEIGRPRouterService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetEIGRPRouterEndpoint, id, "VNetEIGRPRouter"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "EIGRP router", opts)
}

const vnetEIGRPRouterCommandEndpoint = "/vnet_eigrp_router_commands"

// VNetEIGRPRouterCommandService handles EIGRP router commands.
type VNetEIGRPRouterCommandService struct {
	client *Client
}

// List returns EIGRP router commands, with optional filtering and pagination.
// Rows are ordered by orderid unless WithSort sets another order.
func (s *VNetEIGRPRouterCommandService) List(ctx context.Context, opts ...ListOption) ([]VNetEIGRPRouterCommand, error) {
	return listRows[VNetEIGRPRouterCommand](ctx, s.client, vnetEIGRPRouterCommandEndpoint, vnetEIGRPRouterCommandListFields, vnetBGPCommandSort, opts)
}

// ListByRouter returns commands for one EIGRP router.
func (s *VNetEIGRPRouterCommandService) ListByRouter(ctx context.Context, routerID int, opts ...ListOption) ([]VNetEIGRPRouterCommand, error) {
	if err := requireID("eigrp_router", routerID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("eigrp_router", routerID, opts)...)
}

// Get returns one EIGRP router command by key.
func (s *VNetEIGRPRouterCommandService) Get(ctx context.Context, id int) (*VNetEIGRPRouterCommand, error) {
	return getRow[VNetEIGRPRouterCommand](ctx, s.client, vnetEIGRPRouterCommandEndpoint, id, vnetEIGRPRouterCommandGetFields, "VNetEIGRPRouterCommand")
}

// Create creates an EIGRP router command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is created but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetEIGRPRouterCommandService) Create(ctx context.Context, req *VNetEIGRPRouterCommandCreateRequest, opts ...RoutingRestartOption) (*VNetEIGRPRouterCommand, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("eigrp_router", req.EIGRPRouter); err != nil {
		return nil, nil, err
	}
	if err := requireText("command", req.Command); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetEIGRPRouterCommandEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForEIGRPRouter(ctx, int(row.EIGRPRouter))
	if err != nil {
		return row, nil, routingNetworkReadError("EIGRP router command", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "EIGRP router command", "created", opts)
}

// Update updates an EIGRP router command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is updated but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetEIGRPRouterCommandService) Update(ctx context.Context, id int, req *VNetEIGRPRouterCommandUpdateRequest, opts ...RoutingRestartOption) (*VNetEIGRPRouterCommand, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("command", req.Command); err != nil {
		return nil, nil, err
	}
	if err := putRow(ctx, s.client, vnetEIGRPRouterCommandEndpoint, id, req, "VNetEIGRPRouterCommand"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForEIGRPRouter(ctx, int(row.EIGRPRouter))
	if err != nil {
		return row, nil, routingNetworkReadError("EIGRP router command", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "EIGRP router command", "updated", opts)
}

// Delete removes an EIGRP router command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetEIGRPRouterCommandService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForEIGRPRouter(ctx, int(row.EIGRPRouter))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetEIGRPRouterCommandEndpoint, id, "VNetEIGRPRouterCommand"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "EIGRP router command", opts)
}
