package vergeos

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// VNetProxyService handles network proxy configuration (vnet_proxy).
//
// The proxy publishes tenant UIs through a parent network. One network has
// one proxy. Tenant names on that proxy are managed by VNetProxyTenants.
type VNetProxyService struct {
	client *Client
}

// List returns proxy configurations, with optional filtering and pagination.
func (s *VNetProxyService) List(ctx context.Context, opts ...ListOption) ([]VNetProxy, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vnetProxyListFields
	}

	var proxies []VNetProxy
	if err := s.client.get(ctx, "/vnet_proxy", options.toQueryParams(), &proxies); err != nil {
		return nil, err
	}
	return proxies, nil
}

// ListByNetwork returns the proxy configurations for a network.
// A network has at most one. More than one is left for the caller to see;
// GetByNetwork reports that as AmbiguousNameError.
func (s *VNetProxyService) ListByNetwork(ctx context.Context, vnetID int, opts ...ListOption) ([]VNetProxy, error) {
	if vnetID <= 0 {
		return nil, &ValidationError{Field: "vnet", Message: "vnet is required"}
	}
	opts = append([]ListOption{WithFilter(fmt.Sprintf("vnet eq %d", vnetID))}, opts...)
	return s.List(ctx, opts...)
}

// Get returns a proxy configuration by ID.
func (s *VNetProxyService) Get(ctx context.Context, id int) (*VNetProxy, error) {
	params := url.Values{}
	params.Set("fields", vnetProxyGetFields)

	var proxy VNetProxy
	endpoint := fmt.Sprintf("/vnet_proxy/%d", id)
	if err := s.client.get(ctx, endpoint, params, &proxy); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetProxy", ID: id}
		}
		return nil, err
	}
	return &proxy, nil
}

// GetByNetwork returns the proxy configuration for a network.
func (s *VNetProxyService) GetByNetwork(ctx context.Context, vnetID int) (*VNetProxy, error) {
	proxies, err := s.ListByNetwork(ctx, vnetID)
	if err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, &NotFoundError{Resource: "VNetProxy", ID: vnetID}
	}
	if err := requireUniqueName("VNetProxy", fmt.Sprintf("network %d", vnetID), proxies, func(p VNetProxy) any {
		return p.Key
	}); err != nil {
		return nil, err
	}
	return s.Get(ctx, int(proxies[0].Key))
}

// Exists reports whether a network already has a proxy configuration.
func (s *VNetProxyService) Exists(ctx context.Context, vnetID int) (bool, error) {
	proxies, err := s.ListByNetwork(ctx, vnetID)
	if err != nil {
		return false, err
	}
	if len(proxies) > 1 {
		return false, requireUniqueName("VNetProxy", fmt.Sprintf("network %d", vnetID), proxies, func(p VNetProxy) any {
			return p.Key
		})
	}
	return len(proxies) == 1, nil
}

// Create enables the proxy on a network and returns the created configuration.
//
// A second proxy on the same network is refused before POST. ListenAddress
// defaults to 0.0.0.0. DefaultSelf defaults to true.
func (s *VNetProxyService) Create(ctx context.Context, req *VNetProxyCreateRequest) (*VNetProxy, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.VNet <= 0 {
		return nil, &ValidationError{Field: "vnet", Message: "vnet is required"}
	}
	listen := strings.TrimSpace(req.ListenAddress)
	if listen == "" {
		listen = defaultProxyListenAddress
	}
	if net.ParseIP(listen) == nil {
		return nil, &ValidationError{Field: "listen_address", Message: "listen_address must be a valid IP address"}
	}
	defaultSelf := true
	if req.DefaultSelf != nil {
		defaultSelf = *req.DefaultSelf
	}

	exists, err := s.Exists(ctx, req.VNet)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &ValidationError{
			Field:   "vnet",
			Message: fmt.Sprintf("proxy already configured for network %d", req.VNet),
		}
	}

	body := struct {
		VNet          int    `json:"vnet"`
		ListenAddress string `json:"listen_address"`
		DefaultSelf   bool   `json:"default_self"`
		Name          string `json:"name,omitempty"`
	}{
		VNet:          req.VNet,
		ListenAddress: listen,
		DefaultSelf:   defaultSelf,
		Name:          req.Name,
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnet_proxy", body, &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// GetOrCreate returns the network's proxy, creating one when it is missing.
func (s *VNetProxyService) GetOrCreate(ctx context.Context, req *VNetProxyCreateRequest) (*VNetProxy, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	proxy, err := s.GetByNetwork(ctx, req.VNet)
	if err == nil {
		return proxy, nil
	}
	if !IsNotFoundError(err) {
		return nil, err
	}
	return s.Create(ctx, req)
}

// Update updates a proxy configuration and returns the updated row.
func (s *VNetProxyService) Update(ctx context.Context, id int, req *VNetProxyUpdateRequest) (*VNetProxy, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if req.ListenAddress != nil {
		listen := strings.TrimSpace(*req.ListenAddress)
		if net.ParseIP(listen) == nil {
			return nil, &ValidationError{Field: "listen_address", Message: "listen_address must be a valid IP address"}
		}
		req.ListenAddress = &listen
	}

	endpoint := fmt.Sprintf("/vnet_proxy/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetProxy", ID: id}
		}
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete removes a proxy configuration.
// Tenant FQDN mappings for the proxy are removed with it.
func (s *VNetProxyService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vnet_proxy/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VNetProxy", ID: id}
		}
		return err
	}
	return nil
}

// DeleteByNetwork removes the proxy configuration for a network.
func (s *VNetProxyService) DeleteByNetwork(ctx context.Context, vnetID int) error {
	proxy, err := s.GetByNetwork(ctx, vnetID)
	if err != nil {
		return err
	}
	return s.Delete(ctx, int(proxy.Key))
}
