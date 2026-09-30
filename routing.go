package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// RoutingRestartStatus reports the network restart state after a dynamic
// routing change.
//
// VergeOS leaves need_restart set on the network and does not restart it.
// Pending is that flag. Pass WithRestartNetwork to restart the network in the
// same call. Restarted is then true when that restart succeeded. Pending still
// reflects the network after a successful restart, so a flag that remains set
// stays visible.
//
// The restart is NetworkService.Reset. Routing configuration takes effect on
// that restart.
type RoutingRestartStatus struct {
	// NetworkID is the network (vnet) the change belongs to.
	NetworkID int
	// Restarted is true when this call restarted the network.
	Restarted bool
	// Pending is true when the network still has need_restart set.
	Pending bool
}

// RoutingRestartOption configures the restart follow-up for a routing change.
type RoutingRestartOption func(*routingRestartOptions)

type routingRestartOptions struct {
	restart bool
}

// WithRestartNetwork restarts the network after the routing change so the
// change takes effect. Without this option the call still returns
// RoutingRestartStatus, and Pending reports need_restart.
func WithRestartNetwork() RoutingRestartOption {
	return func(opts *routingRestartOptions) {
		opts.restart = true
	}
}

func applyRoutingRestartOptions(opts []RoutingRestartOption) routingRestartOptions {
	var cfg routingRestartOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// routingRestartFollowUp reads need_restart on the network and, when
// requested, restarts it first.
//
// A nil error means the status was read. Pending still reflects the network
// after a successful restart.
func routingRestartFollowUp(ctx context.Context, client *Client, networkID int, opts []RoutingRestartOption) (*RoutingRestartStatus, error) {
	status := &RoutingRestartStatus{NetworkID: networkID}
	if networkID <= 0 {
		return status, nil
	}

	cfg := applyRoutingRestartOptions(opts)
	var restartErr error
	if cfg.restart {
		restartErr = client.Networks.Reset(ctx, networkID, false)
		if restartErr == nil {
			status.Restarted = true
		}
	}

	network, err := client.Networks.Get(ctx, networkID)
	if err != nil {
		if restartErr != nil {
			return status, fmt.Errorf("failed to restart network %d: %w", networkID, restartErr)
		}
		return status, fmt.Errorf("failed to read restart status for network %d: %w", networkID, err)
	}
	status.Pending = network.NeedRestart
	if restartErr != nil {
		return status, fmt.Errorf("failed to restart network %d: %w", networkID, restartErr)
	}
	return status, nil
}

func routingNetworkReadError(noun string, id int, verb string, err error) error {
	return fmt.Errorf("vergeos: %s %d %s but failed to read its network: %w", noun, id, verb, err)
}

func afterRoutingWrite[T any](ctx context.Context, client *Client, row *T, key int, networkID int, noun, verb string, opts []RoutingRestartOption) (*T, *RoutingRestartStatus, error) {
	if networkID <= 0 {
		return row, nil, fmt.Errorf("vergeos: %s %d %s but its network is unknown", noun, key, verb)
	}
	status, err := routingRestartFollowUp(ctx, client, networkID, opts)
	if err != nil {
		return row, status, fmt.Errorf("vergeos: %s %d %s but %w", noun, key, verb, err)
	}
	return row, status, nil
}

func deletedRouting(ctx context.Context, client *Client, key int, networkID int, noun string, opts []RoutingRestartOption) (*RoutingRestartStatus, error) {
	if networkID <= 0 {
		return nil, fmt.Errorf("vergeos: %s %d deleted but its network is unknown", noun, key)
	}
	status, err := routingRestartFollowUp(ctx, client, networkID, opts)
	if err != nil {
		return status, fmt.Errorf("vergeos: %s %d deleted but %w", noun, key, err)
	}
	return status, nil
}

func listRows[T any](ctx context.Context, client *Client, endpoint, defaultFields, defaultSort string, opts []ListOption) ([]T, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = defaultFields
	}
	if options.Sort == "" && defaultSort != "" {
		options.Sort = defaultSort
	}

	var rows []T
	if err := client.get(ctx, endpoint, options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func getRow[T any](ctx context.Context, client *Client, endpoint string, id int, fields, resource string) (*T, error) {
	params := url.Values{}
	params.Set("fields", fields)

	var row T
	path := fmt.Sprintf("%s/%d", endpoint, id)
	if err := client.get(ctx, path, params, &row); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: resource, ID: id}
		}
		return nil, err
	}
	return &row, nil
}

func postKey(ctx context.Context, client *Client, endpoint string, body any) (int, error) {
	var resp apiResponse
	if err := client.post(ctx, endpoint, body, &resp); err != nil {
		return 0, err
	}
	return getKey(resp)
}

func putRow(ctx context.Context, client *Client, endpoint string, id int, body any, resource string) error {
	path := fmt.Sprintf("%s/%d", endpoint, id)
	if err := client.put(ctx, path, body, nil); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: resource, ID: id}
		}
		return err
	}
	return nil
}

func deleteRow(ctx context.Context, client *Client, endpoint string, id int, resource string) error {
	path := fmt.Sprintf("%s/%d", endpoint, id)
	if err := client.delete(ctx, path); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: resource, ID: id}
		}
		return err
	}
	return nil
}

func parentFilter(field string, id int, opts []ListOption) []ListOption {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("%s eq %d", field, id))}
	return append(filterOpts, opts...)
}

func requireID(field string, id int) error {
	if id <= 0 {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	return nil
}

func requireText(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	return nil
}

func validateBGPASN(asn int) error {
	if asn < 1 || int64(asn) > maxBGPASN {
		return &ValidationError{Field: "asn", Message: "asn must be between 1 and 4294967295"}
	}
	return nil
}

func validateEIGRPASN(asn int) error {
	if asn < 1 || asn > maxEIGRPASN {
		return &ValidationError{Field: "asn", Message: "asn must be between 1 and 65535"}
	}
	return nil
}

func validateRouteMapSequence(sequence int) error {
	if sequence < 1 || sequence > maxRouteMapSequence {
		return &ValidationError{Field: "sequence", Message: "sequence must be between 1 and 65535"}
	}
	return nil
}

func validateBGPInterfaceMTU(mtu *int) error {
	if mtu == nil {
		return nil
	}
	if *mtu < minBGPInterfaceMTU || *mtu > maxBGPInterfaceMTU {
		return &ValidationError{Field: "mtu", Message: "mtu must be between 1000 and 65536"}
	}
	return nil
}

func validateBGPLayer2Type(layer2Type *string) error {
	if layer2Type == nil {
		return nil
	}
	if *layer2Type != BGPInterfaceLayer2VLAN && *layer2Type != BGPInterfaceLayer2None {
		return &ValidationError{Field: "layer2_type", Message: "layer2_type must be vlan or none"}
	}
	return nil
}

func validateOptionalText(field string, value *string) error {
	if value != nil && *value == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	return nil
}

func (c *Client) networkIDForBGP(ctx context.Context, bgpID int) (int, error) {
	if err := requireID("bgp", bgpID); err != nil {
		return 0, err
	}
	cfg, err := c.VNetBGP.Get(ctx, bgpID)
	if err != nil {
		return 0, err
	}
	if int(cfg.VNet) <= 0 {
		return 0, fmt.Errorf("vergeos: bgp config %d has no network", bgpID)
	}
	return int(cfg.VNet), nil
}

func (c *Client) networkIDForBGPRouter(ctx context.Context, routerID int) (int, error) {
	if err := requireID("bgp_router", routerID); err != nil {
		return 0, err
	}
	router, err := c.VNetBGPRouters.Get(ctx, routerID)
	if err != nil {
		return 0, err
	}
	return c.networkIDForBGP(ctx, int(router.BGP))
}

func (c *Client) networkIDForBGPInterface(ctx context.Context, interfaceID int) (int, error) {
	if err := requireID("bgp_interface", interfaceID); err != nil {
		return 0, err
	}
	iface, err := c.VNetBGPInterfaces.Get(ctx, interfaceID)
	if err != nil {
		return 0, err
	}
	return c.networkIDForBGP(ctx, int(iface.BGP))
}

func (c *Client) networkIDForBGPRouteMap(ctx context.Context, routeMapID int) (int, error) {
	if err := requireID("bgp_routemap", routeMapID); err != nil {
		return 0, err
	}
	routeMap, err := c.VNetBGPRouteMaps.Get(ctx, routeMapID)
	if err != nil {
		return 0, err
	}
	return c.networkIDForBGP(ctx, int(routeMap.BGP))
}

func (c *Client) networkIDForEIGRPRouter(ctx context.Context, routerID int) (int, error) {
	if err := requireID("eigrp_router", routerID); err != nil {
		return 0, err
	}
	router, err := c.VNetEIGRPRouters.Get(ctx, routerID)
	if err != nil {
		return 0, err
	}
	return c.networkIDForBGP(ctx, int(router.BGP))
}

// bgpKeyForNetwork returns the vnet_bgp key for a network.
// The boolean is false when the network has no routing config.
func (c *Client) bgpKeyForNetwork(ctx context.Context, networkID int) (int, bool, error) {
	if err := requireID("vnet", networkID); err != nil {
		return 0, false, err
	}
	cfg, err := c.VNetBGP.GetByNetwork(ctx, networkID)
	if IsNotFoundError(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(cfg.Key), true, nil
}
