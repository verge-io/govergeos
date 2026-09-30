package vergeos

import "context"

const vnetBGPIPCommandEndpoint = "/vnet_bgp_ip"

// VNetBGPIPCommandService handles BGP IP commands such as prefix-lists and
// AS-path lists (vnet_bgp_ip).
type VNetBGPIPCommandService struct {
	client *Client
}

// List returns BGP IP commands, with optional filtering and pagination.
// Rows are ordered by orderid unless WithSort sets another order.
func (s *VNetBGPIPCommandService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPIPCommand, error) {
	return listRows[VNetBGPIPCommand](ctx, s.client, vnetBGPIPCommandEndpoint, vnetBGPIPCommandListFields, vnetBGPCommandSort, opts)
}

// ListByBGP returns IP commands for one vnet_bgp record.
func (s *VNetBGPIPCommandService) ListByBGP(ctx context.Context, bgpID int, opts ...ListOption) ([]VNetBGPIPCommand, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp", bgpID, opts)...)
}

// ListByNetwork returns IP commands for a network.
// A network with no vnet_bgp row returns an empty list.
func (s *VNetBGPIPCommandService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetBGPIPCommand, error) {
	bgpID, ok, err := s.client.bgpKeyForNetwork(ctx, networkID)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return []VNetBGPIPCommand{}, nil
	}
	return s.ListByBGP(ctx, bgpID, opts...)
}

// Get returns one BGP IP command by key.
func (s *VNetBGPIPCommandService) Get(ctx context.Context, id int) (*VNetBGPIPCommand, error) {
	return getRow[VNetBGPIPCommand](ctx, s.client, vnetBGPIPCommandEndpoint, id, vnetBGPIPCommandGetFields, "VNetBGPIPCommand")
}

// Create creates a BGP IP command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is created but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPIPCommandService) Create(ctx context.Context, req *VNetBGPIPCommandCreateRequest, opts ...RoutingRestartOption) (*VNetBGPIPCommand, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp", req.BGP); err != nil {
		return nil, nil, err
	}
	if err := requireText("command", req.Command); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPIPCommandEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP IP command", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP IP command", "created", opts)
}

// Update updates a BGP IP command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is updated but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPIPCommandService) Update(ctx context.Context, id int, req *VNetBGPIPCommandUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPIPCommand, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("command", req.Command); err != nil {
		return nil, nil, err
	}
	if err := putRow(ctx, s.client, vnetBGPIPCommandEndpoint, id, req, "VNetBGPIPCommand"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP IP command", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP IP command", "updated", opts)
}

// Delete removes a BGP IP command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPIPCommandService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPIPCommandEndpoint, id, "VNetBGPIPCommand"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP IP command", opts)
}
