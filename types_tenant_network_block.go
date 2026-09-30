package vergeos

// TenantNetworkBlock is a CIDR block on a parent network (vnet_cidrs) assigned
// to a tenant. Owner is "tenants/{id}" when the block belongs to a tenant.
type TenantNetworkBlock struct {
	// Key is the unique identifier for the network block.
	Key FlexInt `json:"$key,omitempty"`
	// VNet is the parent network the block is routed from.
	VNet FlexInt `json:"vnet,omitempty"`
	// NetworkName is the parent network name (vnet#name).
	NetworkName string `json:"network_name,omitempty"`
	// CIDR is the block in CIDR notation (for example "192.168.100.0/24").
	CIDR string `json:"cidr,omitempty"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Owner is the owner reference path (for example "tenants/12").
	Owner string `json:"owner,omitempty"`
}

// TenantKey returns the tenant ID parsed from Owner ("tenants/{id}").
// It returns 0 when the block is not owned by a tenant.
func (b TenantNetworkBlock) TenantKey() int {
	return tenantKeyFromOwner(b.Owner)
}

// TenantNetworkBlockCreateRequest assigns a CIDR block to a tenant.
// Tenant is stored as the owner path "tenants/{id}" and is not sent as its own field.
type TenantNetworkBlockCreateRequest struct {
	// Tenant is the tenant to assign the block to (required).
	Tenant int `json:"-"`
	// VNet is the parent network the block is taken from (required).
	VNet int `json:"vnet"`
	// CIDR is the block in CIDR notation (required).
	CIDR string `json:"cidr"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
}

// tenantNetworkBlockListFields are the fields requested when listing network blocks.
const tenantNetworkBlockListFields = "$key,vnet,vnet#name as network_name,cidr,description,owner"

// tenantNetworkBlockGetFields are the fields requested when getting one network block.
const tenantNetworkBlockGetFields = tenantNetworkBlockListFields
