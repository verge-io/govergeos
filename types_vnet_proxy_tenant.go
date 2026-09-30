package vergeos

// VNetProxyTenant maps a tenant to an FQDN on a network proxy (vnet_proxy_tenants).
//
// Several tenants can share one proxy address. The proxy chooses the tenant
// from the name in the request.
type VNetProxyTenant struct {
	// Key is the unique identifier for the mapping.
	Key FlexInt `json:"$key,omitempty"`
	// Proxy is the parent proxy configuration ID.
	Proxy FlexInt `json:"proxy,omitempty"`
	// Tenant is the tenant ID published at FQDN.
	Tenant FlexInt `json:"tenant,omitempty"`
	// TenantName is the tenant name (tenant#name).
	TenantName string `json:"tenant_name,omitempty"`
	// FQDN is the fully qualified domain name for this tenant.
	FQDN string `json:"fqdn,omitempty"`
	// Modified is the last modification timestamp (Unix epoch).
	Modified int64 `json:"modified,omitempty"`
}

// VNetProxyTenantCreateRequest publishes a tenant at an FQDN.
type VNetProxyTenantCreateRequest struct {
	// Proxy is the parent proxy configuration ID (required).
	Proxy int `json:"proxy"`
	// Tenant is the tenant to publish (required).
	Tenant int `json:"tenant"`
	// FQDN is the fully qualified domain name (required).
	FQDN string `json:"fqdn"`
}

// VNetProxyTenantUpdateRequest is the request body for updating a mapping.
// Nil fields are left unchanged.
type VNetProxyTenantUpdateRequest struct {
	// Tenant is the tenant to publish.
	Tenant *int `json:"tenant,omitempty"`
	// FQDN is the fully qualified domain name.
	FQDN *string `json:"fqdn,omitempty"`
}

// vnetProxyTenantListFields are the fields to request when listing mappings.
const vnetProxyTenantListFields = "$key,proxy,tenant,tenant#name as tenant_name,fqdn,modified"

// vnetProxyTenantGetFields are the fields to request when getting one mapping.
const vnetProxyTenantGetFields = vnetProxyTenantListFields
