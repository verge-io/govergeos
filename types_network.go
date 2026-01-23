package vergeos

import "strings"

// Network represents a VergeOS virtual network (vnet).
type Network struct {
	// ID is the unique identifier for the network.
	ID FlexInt `json:"$key,omitempty"`
	// Name is the network name.
	Name string `json:"name"`
	// Description is an optional description of the network.
	Description string `json:"description,omitempty"`
	// Enabled indicates whether the network is enabled.
	Enabled bool `json:"enabled"`
	// Creator is the username that created this network.
	Creator string `json:"creator,omitempty"`

	// IP configuration
	// IPAddress is the network's router IP address.
	IPAddress string `json:"ipaddress,omitempty"`
	// DMZIPAddress is the DMZ IP address for this router.
	DMZIPAddress string `json:"dmz_ipaddress,omitempty"`
	// MACAddress is the router's MAC address (readonly).
	MACAddress string `json:"macaddress,omitempty"`
	// Network is the CIDR notation (e.g., "10.0.0.0/24").
	Network string `json:"network,omitempty"`
	// Gateway is the gateway IP address sent to DHCP clients.
	Gateway string `json:"gateway,omitempty"`
	// DefaultGateway is the default gateway network ID (0 = none).
	DefaultGateway FlexInt `json:"vnet_default_gateway,omitempty"`
	// IPAddressType is the IP address type (static, dynamic, bgp, none).
	IPAddressType string `json:"ipaddress_type,omitempty"`
	// Hostname is the router hostname.
	Hostname string `json:"hostname,omitempty"`

	// DNS configuration
	// DNS is the DNS service mode (disabled, simple, bind, network).
	DNS string `json:"dns,omitempty"`
	// Domain is the domain name for the network.
	Domain string `json:"domain,omitempty"`
	// DNSList is the list of DNS servers (newline, space, or comma separated).
	DNSList string `json:"dnslist,omitempty"`
	// OverrideDHCPDNS indicates whether to ignore DNS servers from DHCP.
	OverrideDHCPDNS bool `json:"override_dhcp_dns,omitempty"`
	// NetworkDNS is the DNS network ID for forwarding (0 = none).
	NetworkDNS FlexInt `json:"network_dns,omitempty"`

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
	// Type is the network type (internal, external, bgp, dmz, core, physical, port_mirror, vpn).
	Type string `json:"type,omitempty"`
	// Layer2Type is the layer 2 type (vlan, vxlan, none, bond, bond_slave).
	Layer2Type string `json:"layer2_type,omitempty"`
	// VLAN is the VLAN/VXLAN tag (layer2_id).
	VLAN int `json:"layer2_id,omitempty"`
	// VXLANMulticast is the multicast address for VXLAN.
	VXLANMulticast string `json:"vxlan_multicast,omitempty"`
	// MTU is the maximum transmission unit.
	MTU int `json:"mtu,omitempty"`
	// InterfaceVnet is the interface vnet ID (0 = none).
	InterfaceVnet FlexInt `json:"interface_vnet,omitempty"`
	// PhysicalBridged indicates whether this is a bridged physical network.
	PhysicalBridged bool `json:"physical_bridged,omitempty"`

	// Power and scheduling
	// OnPowerLoss is the behavior on power loss (power_on, last_state, leave_off).
	OnPowerLoss string `json:"on_power_loss,omitempty"`
	// PowerState indicates whether the network is running.
	PowerState bool `json:"powerstate,omitempty"`
	// Cluster is the cluster ID for this network.
	Cluster FlexInt `json:"cluster,omitempty"`
	// ClusterFailover is the failover cluster ID (0 = none).
	ClusterFailover FlexInt `json:"cluster_failover,omitempty"`
	// PreferredNode is the preferred node ID (0 = none).
	PreferredNode FlexInt `json:"preferred_node,omitempty"`
	// HAGroup is the HA group for anti-affinity.
	HAGroup string `json:"ha_group,omitempty"`

	// Bonding configuration
	// EnableBonding indicates whether bonding is enabled.
	EnableBonding bool `json:"enable_bonding,omitempty"`
	// BondInterfacesArgs are the bonding interface arguments.
	BondInterfacesArgs []int `json:"bond_interfaces_args,omitempty"`

	// Port mirroring
	// PortMirroring is the port mirroring mode (off, east_west, north_south).
	PortMirroring string `json:"port_mirroring,omitempty"`
	// PortMirroringVnet is the network ID for mirrored traffic (0 = none).
	PortMirroringVnet FlexInt `json:"port_mirroring_vnet,omitempty"`

	// Firewall and statistics
	// Statistics indicates whether to track statistics for all rules.
	Statistics bool `json:"statistics,omitempty"`
	// DMZStatistics indicates whether to track DMZ statistics.
	DMZStatistics bool `json:"dmz_statistics,omitempty"`
	// Trace indicates whether to trace/debug firewall rules.
	Trace bool `json:"trace,omitempty"`
	// MirrorLogs indicates whether to mirror syslog to UI.
	MirrorLogs bool `json:"mirror_logs,omitempty"`
	// NeedRestart indicates whether the network needs a restart.
	NeedRestart bool `json:"need_restart,omitempty"`
	// NeedFWApply indicates whether firewall rules need to be applied.
	NeedFWApply bool `json:"need_fw_apply,omitempty"`
	// NeedDNSApply indicates whether DNS configuration needs to be applied.
	NeedDNSApply bool `json:"need_dns_apply,omitempty"`
	// NeedProxyApply indicates whether proxy configuration needs to be applied.
	NeedProxyApply bool `json:"need_proxy_apply,omitempty"`

	// Rate limiting
	// RateLimit is the rate limit value.
	RateLimit int64 `json:"rate_limit,omitempty"`
	// RateLimitType is the rate limit unit type.
	RateLimitType string `json:"rate_limit_type,omitempty"`
	// RateLimitBurst is the rate limit burst value.
	RateLimitBurst int64 `json:"rate_limit_burst,omitempty"`

	// BGP configuration
	// BGPASN is the BGP autonomous system number.
	BGPASN int `json:"bgp_asn,omitempty"`

	// IPsec
	// IPsecEnabled indicates whether IPsec is enabled.
	IPsecEnabled bool `json:"ipsec_enabled,omitempty"`

	// Proxy
	// ProxyEnabled indicates whether the proxy is enabled.
	ProxyEnabled bool `json:"proxy_enabled,omitempty"`
	// ProxyListenAddress is the proxy listen address.
	ProxyListenAddress string `json:"proxy_listen_address,omitempty"`

	// PXE boot
	// PXE is the PXE boot mode (none, ybos, custom).
	PXE string `json:"pxe,omitempty"`
	// TFTPServer is the TFTP server IP for custom PXE.
	TFTPServer string `json:"tftp_server,omitempty"`

	// Gateway monitoring
	// MonitorGateway indicates whether to monitor the gateway.
	MonitorGateway bool `json:"monitor_gateway,omitempty"`
	// MonitorIP is the IP to monitor (blank for default route).
	MonitorIP string `json:"monitor_ip,omitempty"`
	// MonitorIntervalMS is the monitoring interval in milliseconds.
	MonitorIntervalMS int `json:"monitor_interval_ms,omitempty"`

	// Notes
	// Note is a free-form note about the network.
	Note string `json:"note,omitempty"`
}

// NetworkCreateRequest is the request body for creating a network.
type NetworkCreateRequest struct {
	// Name is the network name (required).
	Name string `json:"name"`
	// Description is an optional description of the network.
	Description string `json:"description,omitempty"`
	// Enabled indicates whether the network is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// IP configuration
	// IPAddress is the network's IP address.
	IPAddress string `json:"ipaddress,omitempty"`
	// DMZIPAddress is the DMZ IP address.
	DMZIPAddress string `json:"dmz_ipaddress,omitempty"`
	// Network is the CIDR notation (e.g., "10.0.0.0/24").
	Network string `json:"network,omitempty"`
	// Gateway is the gateway IP address for the network.
	Gateway string `json:"gateway,omitempty"`
	// DefaultGateway is the default gateway network ID.
	DefaultGateway *int `json:"vnet_default_gateway,omitempty"`
	// IPAddressType is the IP address type.
	IPAddressType string `json:"ipaddress_type,omitempty"`
	// Hostname is the router hostname.
	Hostname string `json:"hostname,omitempty"`

	// DNS configuration
	// DNS is the DNS service mode (disabled, simple, bind, network).
	DNS string `json:"dns,omitempty"`
	// Domain is the domain name for the network.
	Domain string `json:"domain,omitempty"`
	// DNSList is the list of DNS servers.
	DNSList string `json:"dnslist,omitempty"`
	// OverrideDHCPDNS indicates whether to ignore DNS servers from DHCP.
	OverrideDHCPDNS *bool `json:"override_dhcp_dns,omitempty"`
	// NetworkDNS is the DNS network ID for forwarding.
	NetworkDNS *int `json:"network_dns,omitempty"`

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
	// Type is the network type (internal, external, bgp, dmz, core, physical, port_mirror, vpn).
	Type string `json:"type,omitempty"`
	// Layer2Type is the layer 2 type.
	Layer2Type string `json:"layer2_type,omitempty"`
	// VLAN is the VLAN tag (layer2_id).
	VLAN *int `json:"layer2_id,omitempty"`
	// VXLANMulticast is the multicast address for VXLAN.
	VXLANMulticast string `json:"vxlan_multicast,omitempty"`
	// MTU is the maximum transmission unit.
	MTU *int `json:"mtu,omitempty"`
	// InterfaceVnet is the interface vnet ID.
	InterfaceVnet *int `json:"interface_vnet,omitempty"`
	// PhysicalBridged indicates whether this is a bridged physical network.
	PhysicalBridged *bool `json:"physical_bridged,omitempty"`

	// Power and scheduling
	// OnPowerLoss is the behavior on power loss.
	OnPowerLoss string `json:"on_power_loss,omitempty"`
	// Cluster is the cluster ID for this network.
	Cluster *int `json:"cluster,omitempty"`
	// ClusterFailover is the failover cluster ID.
	ClusterFailover *int `json:"cluster_failover,omitempty"`
	// PreferredNode is the preferred node ID.
	PreferredNode *int `json:"preferred_node,omitempty"`
	// HAGroup is the HA group for anti-affinity.
	HAGroup string `json:"ha_group,omitempty"`

	// Bonding configuration
	// EnableBonding indicates whether bonding is enabled.
	EnableBonding *bool `json:"enable_bonding,omitempty"`
	// BondInterfacesArgs are the bonding interface arguments.
	BondInterfacesArgs []int `json:"bond_interfaces_args,omitempty"`

	// Port mirroring
	// PortMirroring is the port mirroring mode.
	PortMirroring string `json:"port_mirroring,omitempty"`
	// PortMirroringVnet is the network ID for mirrored traffic.
	PortMirroringVnet *int `json:"port_mirroring_vnet,omitempty"`

	// Firewall and statistics
	// Statistics indicates whether to track statistics for all rules.
	Statistics *bool `json:"statistics,omitempty"`
	// DMZStatistics indicates whether to track DMZ statistics.
	DMZStatistics *bool `json:"dmz_statistics,omitempty"`
	// Trace indicates whether to trace/debug firewall rules.
	Trace *bool `json:"trace,omitempty"`
	// MirrorLogs indicates whether to mirror syslog to UI.
	MirrorLogs *bool `json:"mirror_logs,omitempty"`

	// Rate limiting
	// RateLimit is the rate limit value.
	RateLimit *int64 `json:"rate_limit,omitempty"`
	// RateLimitType is the rate limit unit type.
	RateLimitType string `json:"rate_limit_type,omitempty"`
	// RateLimitBurst is the rate limit burst value.
	RateLimitBurst *int64 `json:"rate_limit_burst,omitempty"`

	// BGP configuration
	// BGPASN is the BGP autonomous system number.
	BGPASN *int `json:"bgp_asn,omitempty"`

	// Proxy
	// ProxyEnabled indicates whether the proxy is enabled.
	ProxyEnabled *bool `json:"proxy_enabled,omitempty"`
	// ProxyListenAddress is the proxy listen address.
	ProxyListenAddress string `json:"proxy_listen_address,omitempty"`

	// PXE boot
	// PXE is the PXE boot mode (none, ybos, custom).
	PXE string `json:"pxe,omitempty"`
	// TFTPServer is the TFTP server IP for custom PXE.
	TFTPServer string `json:"tftp_server,omitempty"`

	// Gateway monitoring
	// MonitorGateway indicates whether to monitor the gateway.
	MonitorGateway *bool `json:"monitor_gateway,omitempty"`
	// MonitorIP is the IP to monitor.
	MonitorIP string `json:"monitor_ip,omitempty"`
	// MonitorIntervalMS is the monitoring interval in milliseconds.
	MonitorIntervalMS *int `json:"monitor_interval_ms,omitempty"`

	// Notes
	// Note is a free-form note about the network.
	Note string `json:"note,omitempty"`
}

// NetworkUpdateRequest is the request body for updating a network.
type NetworkUpdateRequest struct {
	// Name is the network name.
	Name *string `json:"name,omitempty"`
	// Description is the network description.
	Description *string `json:"description,omitempty"`
	// Enabled indicates whether the network is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// IP configuration
	// IPAddress is the network's IP address.
	IPAddress *string `json:"ipaddress,omitempty"`
	// DMZIPAddress is the DMZ IP address.
	DMZIPAddress *string `json:"dmz_ipaddress,omitempty"`
	// Network is the CIDR notation.
	Network *string `json:"network,omitempty"`
	// Gateway is the gateway IP address for the network.
	Gateway *string `json:"gateway,omitempty"`
	// DefaultGateway is the default gateway network ID.
	DefaultGateway *int `json:"vnet_default_gateway,omitempty"`
	// Hostname is the router hostname.
	Hostname *string `json:"hostname,omitempty"`

	// DNS configuration
	// DNS is the DNS service mode (disabled, simple, bind, network).
	DNS *string `json:"dns,omitempty"`
	// Domain is the domain name for the network.
	Domain *string `json:"domain,omitempty"`
	// DNSList is the list of DNS servers.
	DNSList *string `json:"dnslist,omitempty"`
	// OverrideDHCPDNS indicates whether to ignore DNS servers from DHCP.
	OverrideDHCPDNS *bool `json:"override_dhcp_dns,omitempty"`
	// NetworkDNS is the DNS network ID for forwarding.
	NetworkDNS *int `json:"network_dns,omitempty"`

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
	// Layer2Type is the layer 2 type.
	Layer2Type *string `json:"layer2_type,omitempty"`
	// VLAN is the VLAN tag (layer2_id).
	VLAN *int `json:"layer2_id,omitempty"`
	// VXLANMulticast is the multicast address for VXLAN.
	VXLANMulticast *string `json:"vxlan_multicast,omitempty"`
	// MTU is the maximum transmission unit.
	MTU *int `json:"mtu,omitempty"`
	// InterfaceVnet is the interface vnet ID.
	InterfaceVnet *int `json:"interface_vnet,omitempty"`
	// PhysicalBridged indicates whether this is a bridged physical network.
	PhysicalBridged *bool `json:"physical_bridged,omitempty"`

	// Power and scheduling
	// OnPowerLoss is the behavior on power loss.
	OnPowerLoss *string `json:"on_power_loss,omitempty"`
	// Cluster is the cluster ID for this network.
	Cluster *int `json:"cluster,omitempty"`
	// ClusterFailover is the failover cluster ID.
	ClusterFailover *int `json:"cluster_failover,omitempty"`
	// PreferredNode is the preferred node ID.
	PreferredNode *int `json:"preferred_node,omitempty"`
	// HAGroup is the HA group for anti-affinity.
	HAGroup *string `json:"ha_group,omitempty"`

	// Bonding configuration
	// EnableBonding indicates whether bonding is enabled.
	EnableBonding *bool `json:"enable_bonding,omitempty"`
	// BondInterfacesArgs are the bonding interface arguments.
	BondInterfacesArgs []int `json:"bond_interfaces_args,omitempty"`

	// Port mirroring
	// PortMirroring is the port mirroring mode.
	PortMirroring *string `json:"port_mirroring,omitempty"`
	// PortMirroringVnet is the network ID for mirrored traffic.
	PortMirroringVnet *int `json:"port_mirroring_vnet,omitempty"`

	// Firewall and statistics
	// Statistics indicates whether to track statistics for all rules.
	Statistics *bool `json:"statistics,omitempty"`
	// DMZStatistics indicates whether to track DMZ statistics.
	DMZStatistics *bool `json:"dmz_statistics,omitempty"`
	// Trace indicates whether to trace/debug firewall rules.
	Trace *bool `json:"trace,omitempty"`
	// MirrorLogs indicates whether to mirror syslog to UI.
	MirrorLogs *bool `json:"mirror_logs,omitempty"`

	// Rate limiting
	// RateLimit is the rate limit value.
	RateLimit *int64 `json:"rate_limit,omitempty"`
	// RateLimitType is the rate limit unit type.
	RateLimitType *string `json:"rate_limit_type,omitempty"`
	// RateLimitBurst is the rate limit burst value.
	RateLimitBurst *int64 `json:"rate_limit_burst,omitempty"`

	// Proxy
	// ProxyEnabled indicates whether the proxy is enabled.
	ProxyEnabled *bool `json:"proxy_enabled,omitempty"`
	// ProxyListenAddress is the proxy listen address.
	ProxyListenAddress *string `json:"proxy_listen_address,omitempty"`

	// PXE boot
	// PXE is the PXE boot mode (none, ybos, custom).
	PXE *string `json:"pxe,omitempty"`
	// TFTPServer is the TFTP server IP for custom PXE.
	TFTPServer *string `json:"tftp_server,omitempty"`

	// Gateway monitoring
	// MonitorGateway indicates whether to monitor the gateway.
	MonitorGateway *bool `json:"monitor_gateway,omitempty"`
	// MonitorIP is the IP to monitor.
	MonitorIP *string `json:"monitor_ip,omitempty"`
	// MonitorIntervalMS is the monitoring interval in milliseconds.
	MonitorIntervalMS *int `json:"monitor_interval_ms,omitempty"`

	// Notes
	// Note is a free-form note about the network.
	Note *string `json:"note,omitempty"`
}

// vnetAction represents a network action request.
type vnetAction struct {
	VNet   int         `json:"vnet"`
	Action string      `json:"action"`
	Params interface{} `json:"params"`
}

// networkListFields are the fields to request when listing networks.
const networkListFields = "$key,name,description,enabled,creator,ipaddress,dmz_ipaddress,macaddress,network,gateway,ipaddress_type,hostname,dns,domain,dnslist,override_dhcp_dns,network_dns,dhcp_enabled,dhcp_dynamic,dhcp_sequential,dhcp_start,dhcp_stop,type,layer2_type,layer2_id,vxlan_multicast,mtu,interface_vnet,physical_bridged,on_power_loss,powerstate,cluster,cluster_failover,preferred_node,ha_group,enable_bonding,port_mirroring,port_mirroring_vnet,statistics,dmz_statistics,trace,mirror_logs,need_restart,need_fw_apply,need_dns_apply,need_proxy_apply,rate_limit,rate_limit_type,rate_limit_burst,bgp_asn,ipsec_enabled,proxy_enabled,proxy_listen_address,pxe,tftp_server,monitor_gateway,monitor_ip,monitor_interval_ms,note"

// networkGetFields are the fields to request when getting a single network.
const networkGetFields = networkListFields + ",vnet_default_gateway,bond_interfaces_args"

// GetDNSServers returns the DNS list as a slice of server addresses.
// The DNSList field can be newline, space, or comma separated.
func (n *Network) GetDNSServers() []string {
	if n.DNSList == "" {
		return nil
	}
	// Split on common delimiters (newline, comma, space)
	var servers []string
	for _, line := range strings.Split(n.DNSList, "\n") {
		line = strings.TrimSpace(line)
		// Also split on comma and space within each line
		for _, part := range strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ' '
		}) {
			part = strings.TrimSpace(part)
			if part != "" {
				servers = append(servers, part)
			}
		}
	}
	return servers
}
