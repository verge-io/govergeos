package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// VNetProxyTenantService handles tenant FQDN mappings (vnet_proxy_tenants).
//
// A mapping publishes one tenant through a parent network's proxy. Listings
// and lookups that take a proxy ID stay on that proxy.
type VNetProxyTenantService struct {
	client *Client
}

// List returns proxy tenant mappings, with optional filtering and pagination.
func (s *VNetProxyTenantService) List(ctx context.Context, opts ...ListOption) ([]VNetProxyTenant, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vnetProxyTenantListFields
	}

	var mappings []VNetProxyTenant
	if err := s.client.get(ctx, "/vnet_proxy_tenants", options.toQueryParams(), &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

// ListByProxy returns the tenant mappings for one proxy configuration.
func (s *VNetProxyTenantService) ListByProxy(ctx context.Context, proxyID int, opts ...ListOption) ([]VNetProxyTenant, error) {
	if proxyID <= 0 {
		return nil, &ValidationError{Field: "proxy", Message: "proxy is required"}
	}
	opts = append([]ListOption{WithFilter(fmt.Sprintf("proxy eq %d", proxyID))}, opts...)
	return s.List(ctx, opts...)
}

// Get returns a tenant mapping by ID.
func (s *VNetProxyTenantService) Get(ctx context.Context, id int) (*VNetProxyTenant, error) {
	params := url.Values{}
	params.Set("fields", vnetProxyTenantGetFields)

	var mapping VNetProxyTenant
	endpoint := fmt.Sprintf("/vnet_proxy_tenants/%d", id)
	if err := s.client.get(ctx, endpoint, params, &mapping); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "VNetProxyTenant", ID: id}
		}
		return nil, err
	}
	return &mapping, nil
}

// GetByFQDN returns the mapping for an FQDN on a proxy.
func (s *VNetProxyTenantService) GetByFQDN(ctx context.Context, proxyID int, fqdn string) (*VNetProxyTenant, error) {
	fqdn = strings.TrimSpace(fqdn)
	if fqdn == "" {
		return nil, &ValidationError{Field: "fqdn", Message: "fqdn is required"}
	}
	return s.one(ctx, proxyID, fmt.Sprintf("fqdn eq '%s'", escapeFilterValue(fqdn)), fqdn)
}

// GetByTenant returns the mapping for a tenant on a proxy.
// Two FQDNs for the same tenant on that proxy are AmbiguousNameError.
func (s *VNetProxyTenantService) GetByTenant(ctx context.Context, proxyID, tenantID int) (*VNetProxyTenant, error) {
	if tenantID <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	return s.one(ctx, proxyID, fmt.Sprintf("tenant eq %d", tenantID), fmt.Sprintf("tenant %d", tenantID))
}

func (s *VNetProxyTenantService) one(ctx context.Context, proxyID int, filter, name string) (*VNetProxyTenant, error) {
	rows, err := s.ListByProxy(ctx, proxyID, WithFilter(filter))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "VNetProxyTenant", ID: name}
	}
	if err := requireUniqueName("VNetProxyTenant", name, rows, func(row VNetProxyTenant) any {
		return row.Key
	}); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(rows[0].Key))
	if err != nil {
		return nil, err
	}
	if int(got.Proxy) != proxyID {
		return nil, &NotFoundError{Resource: "VNetProxyTenant", ID: name}
	}
	return got, nil
}

// Create publishes a tenant at an FQDN and returns the mapping.
func (s *VNetProxyTenantService) Create(ctx context.Context, req *VNetProxyTenantCreateRequest) (*VNetProxyTenant, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Proxy <= 0 {
		return nil, &ValidationError{Field: "proxy", Message: "proxy is required"}
	}
	if req.Tenant <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	fqdn := strings.TrimSpace(req.FQDN)
	if fqdn == "" {
		return nil, &ValidationError{Field: "fqdn", Message: "fqdn is required"}
	}

	body := struct {
		Proxy  int    `json:"proxy"`
		Tenant int    `json:"tenant"`
		FQDN   string `json:"fqdn"`
	}{
		Proxy:  req.Proxy,
		Tenant: req.Tenant,
		FQDN:   fqdn,
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnet_proxy_tenants", body, &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update updates a tenant mapping and returns the updated row.
func (s *VNetProxyTenantService) Update(ctx context.Context, id int, req *VNetProxyTenantUpdateRequest) (*VNetProxyTenant, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if req.FQDN != nil {
		fqdn := strings.TrimSpace(*req.FQDN)
		if fqdn == "" {
			return nil, &ValidationError{Field: "fqdn", Message: "fqdn is required"}
		}
		req.FQDN = &fqdn
	}
	if req.Tenant != nil && *req.Tenant <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}

	endpoint := fmt.Sprintf("/vnet_proxy_tenants/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete removes a tenant FQDN mapping.
func (s *VNetProxyTenantService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vnet_proxy_tenants/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "VNetProxyTenant", ID: id}
		}
		return err
	}
	return nil
}
