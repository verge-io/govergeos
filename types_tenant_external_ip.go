package vergeos

// TenantExternalIP is a virtual IP on a parent network (vnet_addresses) assigned
// to a tenant. Owner is "tenants/{id}". These are the addresses a parent hands
// a tenant; VNetAddresses remains the generic address table.
type TenantExternalIP struct {
	// Key is the unique identifier for the address.
	Key FlexInt `json:"$key,omitempty"`
	// VNet is the parent network the address is taken from.
	VNet FlexInt `json:"vnet,omitempty"`
	// NetworkName is the parent network name (vnet#name).
	NetworkName string `json:"network_name,omitempty"`
	// IP is the assigned address.
	IP string `json:"ip,omitempty"`
	// Type is the address type. Tenant external IPs are "virtual".
	Type string `json:"type,omitempty"`
	// Hostname is an optional hostname.
	Hostname string `json:"hostname,omitempty"`
	// MAC is the MAC address, when one is set.
	MAC string `json:"mac,omitempty"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Owner is the owner reference path (for example "tenants/12").
	Owner string `json:"owner,omitempty"`
}

// TenantKey returns the tenant ID parsed from Owner ("tenants/{id}").
// It returns 0 when the address is not owned by a tenant.
func (a TenantExternalIP) TenantKey() int {
	return tenantKeyFromOwner(a.Owner)
}

// TenantExternalIPCreateRequest gives a tenant a virtual IP from a parent network.
// The service sends type "virtual" and owner "tenants/{id}".
type TenantExternalIPCreateRequest struct {
	// Tenant is the tenant to assign the address to (required).
	Tenant int `json:"-"`
	// VNet is the parent network the address is taken from (required).
	VNet int `json:"vnet"`
	// IP is the address to assign (required).
	IP string `json:"ip"`
	// Hostname is an optional hostname.
	Hostname string `json:"hostname,omitempty"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
}

// tenantExternalIPListFields are the fields requested when listing tenant external IPs.
const tenantExternalIPListFields = "$key,vnet,vnet#name as network_name,ip,type,hostname,mac,description,owner"

// tenantExternalIPGetFields are the fields requested when getting one tenant external IP.
const tenantExternalIPGetFields = tenantExternalIPListFields
