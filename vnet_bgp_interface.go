package vergeos

import "context"

const vnetBGPInterfaceEndpoint = "/vnet_bgp_interfaces"

// VNetBGPInterfaceService handles BGP routing interfaces.
type VNetBGPInterfaceService struct {
	client *Client
}

// List returns BGP interfaces, with optional filtering and pagination.
// Rows are ordered by name unless WithSort sets another order.
func (s *VNetBGPInterfaceService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPInterface, error) {
	return listRows[VNetBGPInterface](ctx, s.client, vnetBGPInterfaceEndpoint, vnetBGPInterfaceListFields, vnetBGPInterfaceSort, opts)
}

// ListByBGP returns BGP interfaces for one vnet_bgp record.
func (s *VNetBGPInterfaceService) ListByBGP(ctx context.Context, bgpID int, opts ...ListOption) ([]VNetBGPInterface, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp", bgpID, opts)...)
}

// ListByNetwork returns BGP interfaces for a network.
// A network with no vnet_bgp row returns an empty list.
func (s *VNetBGPInterfaceService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetBGPInterface, error) {
	bgpID, ok, err := s.client.bgpKeyForNetwork(ctx, networkID)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return []VNetBGPInterface{}, nil
	}
	return s.ListByBGP(ctx, bgpID, opts...)
}

// Get returns one BGP interface by key.
func (s *VNetBGPInterfaceService) Get(ctx context.Context, id int) (*VNetBGPInterface, error) {
	return getRow[VNetBGPInterface](ctx, s.client, vnetBGPInterfaceEndpoint, id, vnetBGPInterfaceGetFields, "VNetBGPInterface")
}

// GetByName returns a BGP interface by name on one vnet_bgp record.
func (s *VNetBGPInterfaceService) GetByName(ctx context.Context, bgpID int, name string) (*VNetBGPInterface, error) {
	if err := requireText("name", name); err != nil {
		return nil, err
	}
	rows, err := s.ListByBGP(ctx, bgpID, WithFilter("name eq '"+escapeFilterValue(name)+"'"))
	if err != nil {
		return nil, err
	}
	matched := make([]VNetBGPInterface, 0, len(rows))
	for _, row := range rows {
		if row.Name == name {
			matched = append(matched, row)
		}
	}
	row, err := oneRoutingRow("VNetBGPInterface", name, matched, func(row VNetBGPInterface) any { return row.Key })
	if err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(row.Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VNetBGPInterface", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a BGP interface.
//
// The server creates an associated network and NIC for the interface.
// The returned RoutingRestartStatus reports whether the routing network still
// needs a restart. Pass WithRestartNetwork to restart it in this call. If the
// interface is created but the restart follow-up fails, the interface and
// status are still returned with the error.
func (s *VNetBGPInterfaceService) Create(ctx context.Context, req *VNetBGPInterfaceCreateRequest, opts ...RoutingRestartOption) (*VNetBGPInterface, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp", req.BGP); err != nil {
		return nil, nil, err
	}
	if err := requireText("name", req.Name); err != nil {
		return nil, nil, err
	}
	if err := requireText("ipaddress", req.IPAddress); err != nil {
		return nil, nil, err
	}
	if err := requireText("network", req.Network); err != nil {
		return nil, nil, err
	}
	if err := requireID("interface_vnet", req.InterfaceVNet); err != nil {
		return nil, nil, err
	}
	if err := validateBGPLayer2Type(req.Layer2Type); err != nil {
		return nil, nil, err
	}
	if err := validateBGPInterfaceMTU(req.MTU); err != nil {
		return nil, nil, err
	}
	if req.Layer2ID != nil && *req.Layer2ID < 0 {
		return nil, nil, &ValidationError{Field: "layer2_id", Message: "layer2_id must be zero or greater"}
	}

	id, err := postKey(ctx, s.client, vnetBGPInterfaceEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP interface", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP interface", "created", opts)
}

// Update updates a BGP interface.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// interface is updated but the restart follow-up fails, the interface and
// status are still returned with the error.
func (s *VNetBGPInterfaceService) Update(ctx context.Context, id int, req *VNetBGPInterfaceUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPInterface, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("name", req.Name); err != nil {
		return nil, nil, err
	}
	if err := validateOptionalText("ipaddress", req.IPAddress); err != nil {
		return nil, nil, err
	}
	if err := validateOptionalText("network", req.Network); err != nil {
		return nil, nil, err
	}
	if req.InterfaceVNet != nil {
		if err := requireID("interface_vnet", *req.InterfaceVNet); err != nil {
			return nil, nil, err
		}
	}
	if err := validateBGPLayer2Type(req.Layer2Type); err != nil {
		return nil, nil, err
	}
	if err := validateBGPInterfaceMTU(req.MTU); err != nil {
		return nil, nil, err
	}
	if req.Layer2ID != nil && *req.Layer2ID < 0 {
		return nil, nil, &ValidationError{Field: "layer2_id", Message: "layer2_id must be zero or greater"}
	}
	if err := putRow(ctx, s.client, vnetBGPInterfaceEndpoint, id, req, "VNetBGPInterface"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP interface", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP interface", "updated", opts)
}

// Delete removes a BGP interface and, on the server, its network and NIC.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPInterfaceService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPInterfaceEndpoint, id, "VNetBGPInterface"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP interface", opts)
}

const vnetBGPInterfaceCommandEndpoint = "/vnet_bgp_interface_commands"

// VNetBGPInterfaceCommandService handles BGP interface commands.
type VNetBGPInterfaceCommandService struct {
	client *Client
}

// List returns BGP interface commands, with optional filtering and pagination.
// Rows are ordered by orderid unless WithSort sets another order.
func (s *VNetBGPInterfaceCommandService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPInterfaceCommand, error) {
	return listRows[VNetBGPInterfaceCommand](ctx, s.client, vnetBGPInterfaceCommandEndpoint, vnetBGPInterfaceCommandListFields, vnetBGPCommandSort, opts)
}

// ListByInterface returns commands for one BGP interface.
func (s *VNetBGPInterfaceCommandService) ListByInterface(ctx context.Context, interfaceID int, opts ...ListOption) ([]VNetBGPInterfaceCommand, error) {
	if err := requireID("bgp_interface", interfaceID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp_interface", interfaceID, opts)...)
}

// Get returns one BGP interface command by key.
func (s *VNetBGPInterfaceCommandService) Get(ctx context.Context, id int) (*VNetBGPInterfaceCommand, error) {
	return getRow[VNetBGPInterfaceCommand](ctx, s.client, vnetBGPInterfaceCommandEndpoint, id, vnetBGPInterfaceCommandGetFields, "VNetBGPInterfaceCommand")
}

// Create creates a BGP interface command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is created but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPInterfaceCommandService) Create(ctx context.Context, req *VNetBGPInterfaceCommandCreateRequest, opts ...RoutingRestartOption) (*VNetBGPInterfaceCommand, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp_interface", req.BGPInterface); err != nil {
		return nil, nil, err
	}
	if err := requireText("command", req.Command); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPInterfaceCommandEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGPInterface(ctx, int(row.BGPInterface))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP interface command", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP interface command", "created", opts)
}

// Update updates a BGP interface command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is updated but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPInterfaceCommandService) Update(ctx context.Context, id int, req *VNetBGPInterfaceCommandUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPInterfaceCommand, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("command", req.Command); err != nil {
		return nil, nil, err
	}
	if err := putRow(ctx, s.client, vnetBGPInterfaceCommandEndpoint, id, req, "VNetBGPInterfaceCommand"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGPInterface(ctx, int(row.BGPInterface))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP interface command", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP interface command", "updated", opts)
}

// Delete removes a BGP interface command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPInterfaceCommandService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGPInterface(ctx, int(row.BGPInterface))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPInterfaceCommandEndpoint, id, "VNetBGPInterfaceCommand"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP interface command", opts)
}
