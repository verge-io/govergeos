package vergeos

import (
	"context"
	"fmt"
	"net"
	"net/url"
)

// TenantNetworkBlockService handles tenant network block operations on vnet_cidrs.
type TenantNetworkBlockService struct {
	client *Client
}

// List returns network blocks, with optional filtering and pagination.
func (s *TenantNetworkBlockService) List(ctx context.Context, opts ...ListOption) ([]TenantNetworkBlock, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = tenantNetworkBlockListFields
	}

	params := options.toQueryParams()

	var blocks []TenantNetworkBlock
	if err := s.client.get(ctx, "/vnet_cidrs", params, &blocks); err != nil {
		return nil, err
	}

	return blocks, nil
}

// ListByTenant returns network blocks owned by a tenant.
func (s *TenantNetworkBlockService) ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantNetworkBlock, error) {
	if tenantID <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("owner eq '%s'", escapeFilterValue(tenantOwnerRef(tenantID)))))
	return s.List(ctx, opts...)
}

// Get returns a single network block by ID.
func (s *TenantNetworkBlockService) Get(ctx context.Context, id int) (*TenantNetworkBlock, error) {
	params := url.Values{}
	params.Set("fields", tenantNetworkBlockGetFields)

	var block TenantNetworkBlock
	endpoint := fmt.Sprintf("/vnet_cidrs/%d", id)
	if err := s.client.get(ctx, endpoint, params, &block); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TenantNetworkBlock", ID: id}
		}
		return nil, err
	}

	return &block, nil
}

// GetByTenantAndCIDR returns the network block a tenant owns for a CIDR.
func (s *TenantNetworkBlockService) GetByTenantAndCIDR(ctx context.Context, tenantID int, cidr string) (*TenantNetworkBlock, error) {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return nil, &ValidationError{Field: "cidr", Message: "cidr must be a valid CIDR block"}
	}
	blocks, err := s.ListByTenant(ctx, tenantID, WithFilter(fmt.Sprintf("cidr eq '%s'", escapeFilterValue(cidr))))
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, &NotFoundError{Resource: "TenantNetworkBlock", ID: fmt.Sprintf("tenant:%d cidr:%s", tenantID, cidr)}
	}
	if len(blocks) > 1 {
		keys := make([]any, len(blocks))
		for i, block := range blocks {
			keys[i] = block.Key
		}
		return nil, &AmbiguousNameError{Resource: "TenantNetworkBlock", Name: cidr, Keys: keys}
	}
	return s.Get(ctx, int(blocks[0].Key))
}

// Create assigns a CIDR block to a tenant and returns the created block.
//
// The parent network is left with need_fw_apply until its rules are applied.
// The returned ParentFirewallStatus reports that flag. Pass
// WithApplyParentFirewall to apply the parent network's rules in this call.
//
// If the block is created but the firewall follow-up fails, the block and
// status are still returned with the error.
func (s *TenantNetworkBlockService) Create(ctx context.Context, req *TenantNetworkBlockCreateRequest, opts ...ParentFirewallOption) (*TenantNetworkBlock, *ParentFirewallStatus, error) {
	if req == nil {
		return nil, nil, &ValidationError{Message: "create request is required"}
	}
	if req.Tenant <= 0 {
		return nil, nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	if req.VNet <= 0 {
		return nil, nil, &ValidationError{Field: "vnet", Message: "vnet is required"}
	}
	if _, _, err := net.ParseCIDR(req.CIDR); err != nil {
		return nil, nil, &ValidationError{Field: "cidr", Message: "cidr must be a valid CIDR block"}
	}

	body := struct {
		VNet        int    `json:"vnet"`
		CIDR        string `json:"cidr"`
		Description string `json:"description,omitempty"`
		Owner       string `json:"owner"`
	}{
		VNet:        req.VNet,
		CIDR:        req.CIDR,
		Description: req.Description,
		Owner:       tenantOwnerRef(req.Tenant),
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnet_cidrs", body, &resp); err != nil {
		return nil, nil, err
	}

	id, err := getKey(resp)
	if err != nil {
		return nil, nil, err
	}

	block, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	status, fwErr := parentFirewallFollowUp(ctx, s.client, int(block.VNet), opts)
	if fwErr != nil {
		return block, status, fmt.Errorf("vergeos: network block %d created but %w", id, fwErr)
	}
	return block, status, nil
}

// Delete removes a network block.
//
// The returned ParentFirewallStatus reports whether the parent network still
// needs its rules applied. Pass WithApplyParentFirewall to apply them in this
// call. If the block is deleted but the firewall follow-up fails, the status
// is still returned with the error.
func (s *TenantNetworkBlockService) Delete(ctx context.Context, id int, opts ...ParentFirewallOption) (*ParentFirewallStatus, error) {
	block, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/vnet_cidrs/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TenantNetworkBlock", ID: id}
		}
		return nil, err
	}

	status, fwErr := parentFirewallFollowUp(ctx, s.client, int(block.VNet), opts)
	if fwErr != nil {
		return status, fmt.Errorf("vergeos: network block %d deleted but %w", id, fwErr)
	}
	return status, nil
}
