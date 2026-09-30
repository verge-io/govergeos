package vergeos

import (
	"context"
	"strconv"
)

const vnetBGPRouteMapEndpoint = "/vnet_bgp_routemaps"

// VNetBGPRouteMapService handles BGP route map entries.
type VNetBGPRouteMapService struct {
	client *Client
}

// List returns BGP route maps, with optional filtering and pagination.
// Rows are ordered by tag, then sequence, unless WithSort sets another order.
func (s *VNetBGPRouteMapService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPRouteMap, error) {
	return listRows[VNetBGPRouteMap](ctx, s.client, vnetBGPRouteMapEndpoint, vnetBGPRouteMapListFields, vnetBGPRouteMapSort, opts)
}

// ListByBGP returns route maps for one vnet_bgp record.
func (s *VNetBGPRouteMapService) ListByBGP(ctx context.Context, bgpID int, opts ...ListOption) ([]VNetBGPRouteMap, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp", bgpID, opts)...)
}

// ListByNetwork returns route maps for a network.
// A network with no vnet_bgp row returns an empty list.
func (s *VNetBGPRouteMapService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetBGPRouteMap, error) {
	bgpID, ok, err := s.client.bgpKeyForNetwork(ctx, networkID)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return []VNetBGPRouteMap{}, nil
	}
	return s.ListByBGP(ctx, bgpID, opts...)
}

// Get returns one route map by key.
func (s *VNetBGPRouteMapService) Get(ctx context.Context, id int) (*VNetBGPRouteMap, error) {
	return getRow[VNetBGPRouteMap](ctx, s.client, vnetBGPRouteMapEndpoint, id, vnetBGPRouteMapGetFields, "VNetBGPRouteMap")
}

// GetByTagAndSequence returns the route map entry with a tag and sequence on
// one vnet_bgp record.
func (s *VNetBGPRouteMapService) GetByTagAndSequence(ctx context.Context, bgpID int, tag string, sequence int) (*VNetBGPRouteMap, error) {
	if err := requireText("tag", tag); err != nil {
		return nil, err
	}
	if err := validateRouteMapSequence(sequence); err != nil {
		return nil, err
	}
	filter := "tag eq '" + escapeFilterValue(tag) + "' and sequence eq " + strconv.Itoa(sequence)
	rows, err := s.ListByBGP(ctx, bgpID, WithFilter(filter))
	if err != nil {
		return nil, err
	}
	matched := make([]VNetBGPRouteMap, 0, len(rows))
	for _, row := range rows {
		if row.Tag == tag && row.Sequence == sequence {
			matched = append(matched, row)
		}
	}
	name := tag + "/" + strconv.Itoa(sequence)
	row, err := oneRoutingRow("VNetBGPRouteMap", name, matched, func(row VNetBGPRouteMap) any { return row.Key })
	if err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(row.Key))
	if err != nil {
		return nil, err
	}
	if got.Tag != tag || got.Sequence != sequence {
		return nil, &NotFoundError{Resource: "VNetBGPRouteMap", ID: name}
	}
	return got, nil
}

// Create creates a route map entry.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// entry is created but the restart follow-up fails, the entry and status are
// still returned with the error.
func (s *VNetBGPRouteMapService) Create(ctx context.Context, req *VNetBGPRouteMapCreateRequest, opts ...RoutingRestartOption) (*VNetBGPRouteMap, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp", req.BGP); err != nil {
		return nil, nil, err
	}
	if err := requireText("tag", req.Tag); err != nil {
		return nil, nil, err
	}
	if err := validateRouteMapSequence(req.Sequence); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPRouteMapEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP route map", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP route map", "created", opts)
}

// Update updates a route map entry.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// entry is updated but the restart follow-up fails, the entry and status are
// still returned with the error.
func (s *VNetBGPRouteMapService) Update(ctx context.Context, id int, req *VNetBGPRouteMapUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPRouteMap, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("tag", req.Tag); err != nil {
		return nil, nil, err
	}
	if req.Sequence != nil {
		if err := validateRouteMapSequence(*req.Sequence); err != nil {
			return nil, nil, err
		}
	}
	if err := putRow(ctx, s.client, vnetBGPRouteMapEndpoint, id, req, "VNetBGPRouteMap"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP route map", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP route map", "updated", opts)
}

// Delete removes a route map entry and, on the server, its commands.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPRouteMapService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGP(ctx, int(row.BGP))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPRouteMapEndpoint, id, "VNetBGPRouteMap"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP route map", opts)
}

const vnetBGPRouteMapCommandEndpoint = "/vnet_bgp_routemap_commands"

// VNetBGPRouteMapCommandService handles BGP route map commands.
type VNetBGPRouteMapCommandService struct {
	client *Client
}

// List returns route map commands, with optional filtering and pagination.
// Rows are ordered by orderid unless WithSort sets another order.
func (s *VNetBGPRouteMapCommandService) List(ctx context.Context, opts ...ListOption) ([]VNetBGPRouteMapCommand, error) {
	return listRows[VNetBGPRouteMapCommand](ctx, s.client, vnetBGPRouteMapCommandEndpoint, vnetBGPRouteMapCommandListFields, vnetBGPCommandSort, opts)
}

// ListByRouteMap returns commands for one route map entry.
func (s *VNetBGPRouteMapCommandService) ListByRouteMap(ctx context.Context, routeMapID int, opts ...ListOption) ([]VNetBGPRouteMapCommand, error) {
	if err := requireID("bgp_routemap", routeMapID); err != nil {
		return nil, err
	}
	return s.List(ctx, parentFilter("bgp_routemap", routeMapID, opts)...)
}

// Get returns one route map command by key.
func (s *VNetBGPRouteMapCommandService) Get(ctx context.Context, id int) (*VNetBGPRouteMapCommand, error) {
	return getRow[VNetBGPRouteMapCommand](ctx, s.client, vnetBGPRouteMapCommandEndpoint, id, vnetBGPRouteMapCommandGetFields, "VNetBGPRouteMapCommand")
}

// Create creates a route map command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is created but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPRouteMapCommandService) Create(ctx context.Context, req *VNetBGPRouteMapCommandCreateRequest, opts ...RoutingRestartOption) (*VNetBGPRouteMapCommand, *RoutingRestartStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if err := requireID("bgp_routemap", req.BGPRouteMap); err != nil {
		return nil, nil, err
	}
	if err := requireText("command", req.Command); err != nil {
		return nil, nil, err
	}

	id, err := postKey(ctx, s.client, vnetBGPRouteMapCommandEndpoint, req)
	if err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGPRouteMap(ctx, int(row.BGPRouteMap))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP route map command", id, "created", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP route map command", "created", opts)
}

// Update updates a route map command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call. If the
// command is updated but the restart follow-up fails, the command and status
// are still returned with the error.
func (s *VNetBGPRouteMapCommandService) Update(ctx context.Context, id int, req *VNetBGPRouteMapCommandUpdateRequest, opts ...RoutingRestartOption) (*VNetBGPRouteMapCommand, *RoutingRestartStatus, error) {
	if req.empty() {
		return nil, nil, &ValidationError{Message: "no update parameters provided"}
	}
	if err := validateOptionalText("command", req.Command); err != nil {
		return nil, nil, err
	}
	if err := putRow(ctx, s.client, vnetBGPRouteMapCommandEndpoint, id, req, "VNetBGPRouteMapCommand"); err != nil {
		return nil, nil, err
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	networkID, err := s.client.networkIDForBGPRouteMap(ctx, int(row.BGPRouteMap))
	if err != nil {
		return row, nil, routingNetworkReadError("BGP route map command", id, "updated", err)
	}
	return afterRoutingWrite(ctx, s.client, row, id, networkID, "BGP route map command", "updated", opts)
}

// Delete removes a route map command.
//
// The returned RoutingRestartStatus reports whether the network still needs
// a restart. Pass WithRestartNetwork to restart it in this call.
func (s *VNetBGPRouteMapCommandService) Delete(ctx context.Context, id int, opts ...RoutingRestartOption) (*RoutingRestartStatus, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	networkID, err := s.client.networkIDForBGPRouteMap(ctx, int(row.BGPRouteMap))
	if err != nil {
		return nil, err
	}
	if err := deleteRow(ctx, s.client, vnetBGPRouteMapCommandEndpoint, id, "VNetBGPRouteMapCommand"); err != nil {
		return nil, err
	}
	return deletedRouting(ctx, s.client, id, networkID, "BGP route map command", opts)
}
