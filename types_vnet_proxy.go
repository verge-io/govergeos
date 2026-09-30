package vergeos

// VNetProxy is the proxy configuration for one network (vnet_proxy).
//
// The proxy publishes tenant UIs through that network. Each tenant is
// reached by an FQDN mapped on VNetProxyTenants. A network has one proxy.
type VNetProxy struct {
	// Key is the unique identifier for the proxy configuration.
	Key FlexInt `json:"$key,omitempty"`
	// VNet is the parent network ID.
	VNet FlexInt `json:"vnet,omitempty"`
	// NetworkName is the parent network name (vnet#name).
	NetworkName string `json:"network_name,omitempty"`
	// Name is the optional proxy name.
	Name string `json:"name,omitempty"`
	// ListenAddress is the address the proxy listens on.
	// "0.0.0.0" listens on every address of the network.
	ListenAddress string `json:"listen_address,omitempty"`
	// DefaultSelf serves this system's UI when no tenant FQDN matches.
	DefaultSelf bool `json:"default_self,omitempty"`
	// Modified is the last modification timestamp (Unix epoch).
	Modified int64 `json:"modified,omitempty"`
}

// VNetProxyCreateRequest enables the proxy on a network.
//
// An empty ListenAddress is stored as "0.0.0.0". A nil DefaultSelf is
// stored as true, which is the platform default: unmatched names show
// this system's UI.
type VNetProxyCreateRequest struct {
	// VNet is the parent network ID (required).
	VNet int `json:"vnet"`
	// ListenAddress is the address to listen on.
	ListenAddress string `json:"listen_address,omitempty"`
	// DefaultSelf serves this system's UI when no tenant FQDN matches.
	DefaultSelf *bool `json:"default_self,omitempty"`
	// Name is an optional proxy name.
	Name string `json:"name,omitempty"`
}

// VNetProxyUpdateRequest is the request body for updating a proxy.
// Nil fields are left unchanged.
type VNetProxyUpdateRequest struct {
	// ListenAddress is the address to listen on.
	ListenAddress *string `json:"listen_address,omitempty"`
	// DefaultSelf serves this system's UI when no tenant FQDN matches.
	DefaultSelf *bool `json:"default_self,omitempty"`
	// Name is the proxy name.
	Name *string `json:"name,omitempty"`
}

// defaultProxyListenAddress is the listen address used when Create omits one.
const defaultProxyListenAddress = "0.0.0.0"

// vnetProxyListFields are the fields to request when listing proxies.
const vnetProxyListFields = "$key,vnet,vnet#name as network_name,name,listen_address,default_self,modified"

// vnetProxyGetFields are the fields to request when getting one proxy.
const vnetProxyGetFields = vnetProxyListFields
