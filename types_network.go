package vergeos

// Network represents a VergeOS virtual network (vnet).
type Network struct {
	// ID is the unique identifier for the network.
	ID FlexInt `json:"$key,omitempty"`
	// Name is the network name.
	Name string `json:"name"`
	// Enabled indicates whether the network is enabled.
	Enabled bool `json:"enabled"`
	// DefaultGateway is the default gateway.
	DefaultGateway int `json:"vnet_default_gateway,omitempty"`
	// IPAddress is the network's IP address.
	IPAddress string `json:"ipaddress,omitempty"`
	// Network is the CIDR notation (e.g., "10.0.0.0/24").
	Network string `json:"network,omitempty"`
	// Gateway is the gateway IP address for the network.
	Gateway string `json:"gateway,omitempty"`
	// DNS is the DNS service mode (disabled, simple, bind, network).
	DNS string `json:"dns,omitempty"`
	// Domain is the domain name for the network.
	Domain string `json:"domain,omitempty"`

	// DHCP configuration
	// DHCPEnabled indicates whether DHCP is enabled.
	DHCPEnabled bool `json:"dhcp_enabled,omitempty"`
	// DHCPDynamic indicates whether dynamic DHCP is enabled.
	DHCPDynamic bool `json:"dhcp_dynamic,omitempty"`
	// DHCPSequential indicates whether DHCP assigns IPs sequentially.
	DHCPSequential bool `json:"dhcp_sequential,omitempty"`
	// DHCPStart is the start of the DHCP range.
	DHCPStart string `json:"dhcp_start,omitempty"`
	// DHCPStop is the end of the DHCP range.
	DHCPStop string `json:"dhcp_stop,omitempty"`

	// Network configuration
	// OnPowerLoss is the behavior on power loss.
	OnPowerLoss string `json:"on_power_loss,omitempty"`
	// PowerState is the network power state.
	PowerState string `json:"powerstate,omitempty"`
	// Type is the network type (internal, external, bgp, dmz, core, physical, port_mirror, vpn).
	Type string `json:"type,omitempty"`
	// VLAN is the VLAN tag (layer2_id).
	VLAN int `json:"layer2_id,omitempty"`
	// MTU is the maximum transmission unit.
	MTU int `json:"mtu,omitempty"`
	// InterfaceVnet is the interface vnet ID.
	InterfaceVnet int `json:"interface_vnet,omitempty"`
	// IPAddressType is the IP address type.
	IPAddressType string `json:"ipaddress_type,omitempty"`
	// Layer2Type is the layer 2 type.
	Layer2Type string `json:"layer2_type,omitempty"`

	// Bonding configuration
	// EnableBonding indicates whether bonding is enabled.
	EnableBonding bool `json:"enable_bonding,omitempty"`
	// BondInterfacesArgs are the bonding interface arguments.
	BondInterfacesArgs []int `json:"bond_interfaces_args,omitempty"`
}

// NetworkCreateRequest is the request body for creating a network.
type NetworkCreateRequest struct {
	// Name is the network name (required).
	Name string `json:"name"`
	// Enabled indicates whether the network is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// IPAddress is the network's IP address.
	IPAddress string `json:"ipaddress,omitempty"`
	// Network is the CIDR notation (e.g., "10.0.0.0/24").
	Network string `json:"network,omitempty"`
	// Gateway is the gateway IP address for the network.
	Gateway string `json:"gateway,omitempty"`
	// DNS is the DNS service mode (disabled, simple, bind, network).
	DNS string `json:"dns,omitempty"`
	// Domain is the domain name for the network.
	Domain string `json:"domain,omitempty"`

	// DHCP configuration
	// DHCPEnabled indicates whether DHCP is enabled.
	DHCPEnabled *bool `json:"dhcp_enabled,omitempty"`
	// DHCPDynamic indicates whether dynamic DHCP is enabled.
	DHCPDynamic *bool `json:"dhcp_dynamic,omitempty"`
	// DHCPSequential indicates whether DHCP assigns IPs sequentially.
	DHCPSequential *bool `json:"dhcp_sequential,omitempty"`
	// DHCPStart is the start of the DHCP range.
	DHCPStart string `json:"dhcp_start,omitempty"`
	// DHCPStop is the end of the DHCP range.
	DHCPStop string `json:"dhcp_stop,omitempty"`

	// Network configuration
	// OnPowerLoss is the behavior on power loss.
	OnPowerLoss string `json:"on_power_loss,omitempty"`
	// Type is the network type (internal, external, bgp, dmz, core, physical, port_mirror, vpn).
	Type string `json:"type,omitempty"`
	// VLAN is the VLAN tag (layer2_id).
	VLAN *int `json:"layer2_id,omitempty"`
	// MTU is the maximum transmission unit.
	MTU *int `json:"mtu,omitempty"`
	// InterfaceVnet is the interface vnet ID.
	InterfaceVnet *int `json:"interface_vnet,omitempty"`
	// IPAddressType is the IP address type.
	IPAddressType string `json:"ipaddress_type,omitempty"`
	// Layer2Type is the layer 2 type.
	Layer2Type string `json:"layer2_type,omitempty"`

	// Bonding configuration
	// EnableBonding indicates whether bonding is enabled.
	EnableBonding *bool `json:"enable_bonding,omitempty"`
	// BondInterfacesArgs are the bonding interface arguments.
	BondInterfacesArgs []int `json:"bond_interfaces_args,omitempty"`
}

// NetworkUpdateRequest is the request body for updating a network.
type NetworkUpdateRequest struct {
	// Name is the network name.
	Name *string `json:"name,omitempty"`
	// Enabled indicates whether the network is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// IPAddress is the network's IP address.
	IPAddress *string `json:"ipaddress,omitempty"`
	// Network is the CIDR notation.
	Network *string `json:"network,omitempty"`
	// Gateway is the gateway IP address for the network.
	Gateway *string `json:"gateway,omitempty"`
	// DNS is the DNS service mode (disabled, simple, bind, network).
	DNS *string `json:"dns,omitempty"`
	// Domain is the domain name for the network.
	Domain *string `json:"domain,omitempty"`

	// DHCP configuration
	// DHCPEnabled indicates whether DHCP is enabled.
	DHCPEnabled *bool `json:"dhcp_enabled,omitempty"`
	// DHCPDynamic indicates whether dynamic DHCP is enabled.
	DHCPDynamic *bool `json:"dhcp_dynamic,omitempty"`
	// DHCPSequential indicates whether DHCP assigns IPs sequentially.
	DHCPSequential *bool `json:"dhcp_sequential,omitempty"`
	// DHCPStart is the start of the DHCP range.
	DHCPStart *string `json:"dhcp_start,omitempty"`
	// DHCPStop is the end of the DHCP range.
	DHCPStop *string `json:"dhcp_stop,omitempty"`

	// Network configuration
	// OnPowerLoss is the behavior on power loss.
	OnPowerLoss *string `json:"on_power_loss,omitempty"`
	// Type is the network type (internal, external, bgp, dmz, core, physical, port_mirror, vpn).
	Type *string `json:"type,omitempty"`
	// VLAN is the VLAN tag (layer2_id).
	VLAN *int `json:"layer2_id,omitempty"`
	// MTU is the maximum transmission unit.
	MTU *int `json:"mtu,omitempty"`
	// InterfaceVnet is the interface vnet ID.
	InterfaceVnet *int `json:"interface_vnet,omitempty"`

	// Bonding configuration
	// EnableBonding indicates whether bonding is enabled.
	EnableBonding *bool `json:"enable_bonding,omitempty"`
	// BondInterfacesArgs are the bonding interface arguments.
	BondInterfacesArgs []int `json:"bond_interfaces_args,omitempty"`
}

// vnetAction represents a network action request.
type vnetAction struct {
	VNet   int         `json:"vnet"`
	Action string      `json:"action"`
	Params interface{} `json:"params"`
}

// networkListFields are the fields to request when listing networks.
const networkListFields = "$key,name,enabled,ipaddress,network,gateway,dns,domain,dhcp_enabled,dhcp_dynamic,dhcp_sequential,dhcp_start,dhcp_stop,on_power_loss,type,layer2_id,mtu,interface_vnet,ipaddress_type,layer2_type,enable_bonding,bond_interfaces_args"

// networkGetFields are the fields to request when getting a single network.
const networkGetFields = networkListFields + ",vnet_default_gateway"
