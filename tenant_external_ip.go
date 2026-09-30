package vergeos

import (
	"context"
	"fmt"
	"net"
	"net/url"
)

// TenantExternalIPService assigns virtual IPs from a parent network to a tenant.
//
// Rows live in vnet_addresses. This service always sets the tenant as owner
// ("tenants/{id}") and the address type to virtual, which is the "give tenant
// X this IP" operation. Generic address management stays on VNetAddresses.
type TenantExternalIPService struct {
	client *Client
}

// List returns tenant external IPs (virtual addresses owned by a tenant),
// with optional filtering and pagination.
func (s *TenantExternalIPService) List(ctx context.Context, opts ...ListOption) ([]TenantExternalIP, error) {
	opts = append([]ListOption{WithFilter("type eq 'virtual' and owner bw 'tenants/'")}, opts...)
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = tenantExternalIPListFields
	}

	params := options.toQueryParams()

	var addresses []TenantExternalIP
	if err := s.client.get(ctx, "/vnet_addresses", params, &addresses); err != nil {
		return nil, err
	}

	return addresses, nil
}

// ListByTenant returns external IPs owned by a tenant.
func (s *TenantExternalIPService) ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantExternalIP, error) {
	if tenantID <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("owner eq '%s'", escapeFilterValue(tenantOwnerRef(tenantID)))))
	return s.List(ctx, opts...)
}

// Get returns a tenant external IP by ID.
// An address that is not a virtual IP owned by a tenant is reported as not found.
func (s *TenantExternalIPService) Get(ctx context.Context, id int) (*TenantExternalIP, error) {
	params := url.Values{}
	params.Set("fields", tenantExternalIPGetFields)

	var address TenantExternalIP
	endpoint := fmt.Sprintf("/vnet_addresses/%d", id)
	if err := s.client.get(ctx, endpoint, params, &address); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantExternalIP", ID: id}
		}
		return nil, err
	}
	if address.Type != AddressTypeVirtual || address.TenantKey() == 0 {
		return nil, &NotFoundError{Resource: "TenantExternalIP", ID: id}
	}

	return &address, nil
}

// GetByTenantAndIP returns the external IP a tenant owns for an address.
func (s *TenantExternalIPService) GetByTenantAndIP(ctx context.Context, tenantID int, ip string) (*TenantExternalIP, error) {
	if net.ParseIP(ip) == nil {
		return nil, &ValidationError{Field: "ip", Message: "ip must be a valid IP address"}
	}
	addresses, err := s.ListByTenant(ctx, tenantID, WithFilter(fmt.Sprintf("ip eq '%s'", escapeFilterValue(ip))))
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, &NotFoundError{Resource: "TenantExternalIP", ID: fmt.Sprintf("tenant:%d ip:%s", tenantID, ip)}
	}
	if len(addresses) > 1 {
		keys := make([]any, len(addresses))
		for i, address := range addresses {
			keys[i] = address.Key
		}
		return nil, &AmbiguousNameError{Resource: "TenantExternalIP", Name: ip, Keys: keys}
	}
	return s.Get(ctx, int(addresses[0].Key))
}

// Create gives a tenant a virtual IP from a parent network.
//
// The parent network is left with need_fw_apply until its rules are applied.
// The returned ParentFirewallStatus reports that flag. Pass
// WithApplyParentFirewall to apply the parent network's rules in this call.
//
// If the address is created but the firewall follow-up fails, the address and
// status are still returned with the error.
func (s *TenantExternalIPService) Create(ctx context.Context, req *TenantExternalIPCreateRequest, opts ...ParentFirewallOption) (*TenantExternalIP, *ParentFirewallStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if req.Tenant <= 0 {
		return nil, nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	if req.VNet <= 0 {
		return nil, nil, &ValidationError{Field: "vnet", Message: "vnet is required"}
	}
	if net.ParseIP(req.IP) == nil {
		return nil, nil, &ValidationError{Field: "ip", Message: "ip must be a valid IP address"}
	}

	body := struct {
		VNet        int    `json:"vnet"`
		IP          string `json:"ip"`
		Type        string `json:"type"`
		Owner       string `json:"owner"`
		Hostname    string `json:"hostname,omitempty"`
		Description string `json:"description,omitempty"`
	}{
		VNet:        req.VNet,
		IP:          req.IP,
		Type:        AddressTypeVirtual,
		Owner:       tenantOwnerRef(req.Tenant),
		Hostname:    req.Hostname,
		Description: req.Description,
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnet_addresses", body, &resp); err != nil {
		return nil, nil, err
	}

	id, err := getKey(resp)
	if err != nil {
		return nil, nil, err
	}

	address, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	status, fwErr := parentFirewallFollowUp(ctx, s.client, int(address.VNet), opts)
	if fwErr != nil {
		return address, status, fmt.Errorf("vergeos: external IP %d created but %w", id, fwErr)
	}
	return address, status, nil
}

// Delete removes a tenant external IP.
//
// The returned ParentFirewallStatus reports whether the parent network still
// needs its rules applied. Pass WithApplyParentFirewall to apply them in this
// call. If the address is deleted but the firewall follow-up fails, the status
// is still returned with the error.
func (s *TenantExternalIPService) Delete(ctx context.Context, id int, opts ...ParentFirewallOption) (*ParentFirewallStatus, error) {
	address, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/vnet_addresses/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantExternalIP", ID: id}
		}
		return nil, err
	}

	status, fwErr := parentFirewallFollowUp(ctx, s.client, int(address.VNet), opts)
	if fwErr != nil {
		return status, fmt.Errorf("vergeos: external IP %d deleted but %w", id, fwErr)
	}
	return status, nil
}
