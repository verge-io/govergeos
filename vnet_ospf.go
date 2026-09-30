package vergeos

import "context"

const vnetOSPFCommandEndpoint = "/vnet_ospf_commands"

// VNetOSPFCommandService handles OSPF commands.
// OSPF rows use the network's vnet_bgp record as their parent.
type VNetOSPFCommandService struct {
	client *Client
}

// List returns OSPF commands, with optional filtering and pagination.
// Rows are ordered by orderid unless WithSort sets another order.
func (s *VNetOSPFCommandService) List(ctx context.Context, opts ...ListOption) ([]VNetOSPFCommand, error) {
	return listRows[VNetOSPFCommand](ctx, s.client, vnetOSPFCommandEndpoint, vnetOSPFCommandListFields, vnetBGPCommandSort, opts)
}

// ListByBGP returns OSPF commands for one vnet_bgp record.
func (s *VNetOSPFCommandService) ListByBGP(ctx context.Context, bgpID int, opts ...ListOption) ([]VNetOSPFCommand, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp", bgpID, opts)...)
}

// ListByNetwork returns OSPF commands for a network.
// A network with no vnet_bgp row returns an empty list.
func (s *VNetOSPFCommandService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetOSPFCommand, error) {
	bgpID, ok, err := s.client.bgpKeyForNetwork(ctx, networkID)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return []VNetOSPFCommand{}, nil
	}
	return s.ListByBGP(ctx, bgpID, opts...)
}

// Get returns one OSPF command by key.
func (s *VNetOSPFCommandService) Get(ctx context.Context, id int) (*VNetOSPFCommand, error) {
	return getRow[VNetOSPFCommand](ctx, s.client, vnetOSPFCommandEndpoint, id, vnetOSPFCommandGetFields, "VNetOSPFCommand")
}

// Create creates an OSPF command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is created but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetOSPFCommandService) Create(ctx context.Context, req *VNetOSPFCommandCreateRequest, opts ...RoutingRestartOption) (*VNetOSPFCommand, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp", req.BGP); err != nil {
		return nil, nil, err
	}
	if err := requireText("command", req.Command); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetOSPFCommandEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("OSPF command", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "OSPF command", "created", opts)
}

// Update updates an OSPF command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is updated but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetOSPFCommandService) Update(ctx context.Context, id int, req *VNetOSPFCommandUpdateRequest, opts ...RoutingRestartOption) (*VNetOSPFCommand, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("command", req.Command); err != nil {
		return nil, nil, err
	}
	if err := putRow(ctx, s.client, vnetOSPFCommandEndpoint, id, req, "VNetOSPFCommand"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("OSPF command", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "OSPF command", "updated", opts)
}

// Delete removes an OSPF command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetOSPFCommandService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetOSPFCommandEndpoint, id, "VNetOSPFCommand"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "OSPF command", opts)
}
