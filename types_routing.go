package vergeos

// Dynamic routing types cover the VergeOS BGP, OSPF, and EIGRP tables.
//
// Every protocol row hangs off one vnet_bgp record for the network.
// VNetBGP.GetOrCreate returns that record. Child services take its key in BGP.
//
// A create, update, or delete leaves the network needing a restart before the
// change takes effect. Those methods return RoutingRestartStatus. Pending is
// the network's need_restart flag. WithRestartNetwork restarts the network in
// the same call.

// BGP interface Layer 2 types.
const (
	// BGPInterfaceLayer2VLAN tags the interface with a VLAN.
	BGPInterfaceLayer2VLAN = "vlan"
	// BGPInterfaceLayer2None leaves the interface untagged.
	BGPInterfaceLayer2None = "none"
)

// BGP router command values. The command column is a string; these are the
// values the routing API documents. Other command strings are still accepted.
const (
	BGPRouterCommandAggregateAddress = "aggregate-address"
	BGPRouterCommandBGP              = "bgp"
	BGPRouterCommandBMP              = "bmp"
	BGPRouterCommandCoalesceTime     = "coalesce-time"
	BGPRouterCommandDistance         = "distance"
	BGPRouterCommandMaximumPaths     = "maximum-paths"
	BGPRouterCommandNeighbor         = "neighbor"
	BGPRouterCommandNetwork          = "network"
	BGPRouterCommandReadQuanta       = "read-quanta"
	BGPRouterCommandRedistribute     = "redistribute"
	BGPRouterCommandSegmentRouting   = "segment-routing"
	BGPRouterCommandTableMap         = "table-map"
	BGPRouterCommandTimers           = "timers"
	BGPRouterCommandUpdateDelay      = "update-delay"
	BGPRouterCommandVNC              = "vnc"
	BGPRouterCommandVRFPolicy        = "vrf-policy"
	BGPRouterCommandWriteQuanta      = "write-quanta"
)

// BGP interface command values.
const (
	BGPInterfaceCommandBandwidth   = "bandwidth"
	BGPInterfaceCommandDescription = "description"
	BGPInterfaceCommandIP          = "ip"
	BGPInterfaceCommandLinkDetect  = "link-detect"
	BGPInterfaceCommandMPLSTE      = "mpls-te"
	BGPInterfaceCommandMulticast   = "multicast"
	BGPInterfaceCommandNo          = "no"
	BGPInterfaceCommandOSPF        = "ospf"
)

// BGP route map command values.
const (
	BGPRouteMapCommandCall        = "call"
	BGPRouteMapCommandContinue    = "continue"
	BGPRouteMapCommandDescription = "description"
	BGPRouteMapCommandMatch       = "match"
	BGPRouteMapCommandOn          = "on"
	BGPRouteMapCommandSet         = "set"
)

// BGP IP command values (prefix-lists, AS-path lists, community lists, and
// other IP-level commands on vnet_bgp_ip).
const (
	BGPIPCommandASPath           = "as-path"
	BGPIPCommandCommunityList    = "community-list"
	BGPIPCommandExtCommunityList = "extcommunity-list"
	BGPIPCommandForwarding       = "forwarding"
	BGPIPCommandMRoute           = "mroute"
	BGPIPCommandMulticast        = "multicast"
	BGPIPCommandMulticastRouting = "multicast-routing"
	BGPIPCommandOSPF             = "ospf"
	BGPIPCommandPrefixList       = "prefix-list"
	BGPIPCommandProtocol         = "protocol"
	BGPIPCommandRoute            = "route"
	BGPIPCommandSSMPingD         = "ssmpingd"
)

// OSPF command values.
const (
	OSPFCommandAdvanced            = "advanced"
	OSPFCommandArea                = "area"
	OSPFCommandAutoCost            = "auto-cost"
	OSPFCommandCapability          = "capability"
	OSPFCommandDefaultInformation  = "default-information"
	OSPFCommandDefaultMetric       = "default-metric"
	OSPFCommandDistance            = "distance"
	OSPFCommandDistributeList      = "distribute-list"
	OSPFCommandLogAdjacencyChanges = "log-adjacency-changes"
	OSPFCommandMaxMetric           = "max-metric"
	OSPFCommandMPLSTE              = "mpls-te"
	OSPFCommandNeighbor            = "neighbor"
	OSPFCommandNetwork             = "network"
	OSPFCommandNo                  = "no"
	OSPFCommandOSPF                = "ospf"
	OSPFCommandPassiveInterface    = "passive-interface"
	OSPFCommandRedistribute        = "redistribute"
	OSPFCommandRouterID            = "router-id"
	OSPFCommandTimers              = "timers"
)

// EIGRP router command values.
const (
	EIGRPRouterCommandEIGRP            = "eigrp"
	EIGRPRouterCommandCoalesceTime     = "coalesce-time"
	EIGRPRouterCommandMaximumPaths     = "maximum-paths"
	EIGRPRouterCommandMetric           = "metric"
	EIGRPRouterCommandNeighbor         = "neighbor"
	EIGRPRouterCommandNetwork          = "network"
	EIGRPRouterCommandPassiveInterface = "passive-interface"
	EIGRPRouterCommandRedistribute     = "redistribute"
	EIGRPRouterCommandTimers           = "timers"
	EIGRPRouterCommandVariance         = "variance"
)

const (
	vnetBGPListFields                 = "$key,vnet"
	vnetBGPRouterListFields           = "$key,bgp,asn"
	vnetBGPRouterCommandListFields    = "$key,bgp_router,orderid,enabled,no,command,params"
	vnetBGPInterfaceListFields        = "$key,bgp,name,description,ipaddress,network,mtu,layer2_type,layer2_id,interface_vnet,bgp_vnet,nic"
	vnetBGPInterfaceCommandListFields = "$key,bgp_interface,orderid,command,params"
	vnetBGPRouteMapListFields         = "$key,bgp,tag,permit,sequence"
	vnetBGPRouteMapCommandListFields  = "$key,bgp_routemap,orderid,command,params"
	vnetBGPIPCommandListFields        = "$key,bgp,orderid,command,params"
	vnetOSPFCommandListFields         = "$key,bgp,orderid,command,params"
	vnetEIGRPRouterListFields         = "$key,bgp,asn"
	vnetEIGRPRouterCommandListFields  = "$key,eigrp_router,orderid,enabled,no,command,params"

	vnetBGPGetFields                 = vnetBGPListFields
	vnetBGPRouterGetFields           = vnetBGPRouterListFields
	vnetBGPRouterCommandGetFields    = vnetBGPRouterCommandListFields
	vnetBGPInterfaceGetFields        = vnetBGPInterfaceListFields
	vnetBGPInterfaceCommandGetFields = vnetBGPInterfaceCommandListFields
	vnetBGPRouteMapGetFields         = vnetBGPRouteMapListFields
	vnetBGPRouteMapCommandGetFields  = vnetBGPRouteMapCommandListFields
	vnetBGPIPCommandGetFields        = vnetBGPIPCommandListFields
	vnetOSPFCommandGetFields         = vnetOSPFCommandListFields
	vnetEIGRPRouterGetFields         = vnetEIGRPRouterListFields
	vnetEIGRPRouterCommandGetFields  = vnetEIGRPRouterCommandListFields

	vnetBGPRouterSort          = "asn"
	vnetBGPInterfaceSort       = "name"
	vnetBGPRouteMapSort        = "tag,sequence"
	vnetBGPCommandSort         = "orderid"
	vnetEIGRPRouterSort        = "asn"
	maxBGPASN            int64 = 4294967295
	maxEIGRPASN                = 65535
	maxRouteMapSequence        = 65535
	minBGPInterfaceMTU         = 1000
	maxBGPInterfaceMTU         = 65536
)

// VNetBGP is the per-network dynamic routing record (vnet_bgp).
// BGP, OSPF, and EIGRP rows all reference this key.
type VNetBGP struct {
	// Key is the vnet_bgp row key.
	Key FlexInt `json:"$key,omitempty"`
	// VNet is the network this routing configuration belongs to.
	VNet FlexInt `json:"vnet,omitempty"`
}

// VNetBGPCreateRequest creates a vnet_bgp row for a network.
type VNetBGPCreateRequest struct {
	// VNet is the network key (required).
	VNet int `json:"vnet"`
}

// VNetBGPRouter is a BGP router (vnet_bgp_routers).
type VNetBGPRouter struct {
	// Key is the router row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGP is the parent vnet_bgp key.
	BGP FlexInt `json:"bgp,omitempty"`
	// ASN is the local autonomous system number.
	ASN int `json:"asn,omitempty"`
}

// VNetBGPRouterCreateRequest creates a BGP router.
type VNetBGPRouterCreateRequest struct {
	// BGP is the parent vnet_bgp key (required).
	BGP int `json:"bgp"`
	// ASN is the autonomous system number, from 1 to 4294967295 (required).
	ASN int `json:"asn"`
}

// VNetBGPRouterUpdateRequest updates a BGP router.
// A nil field is omitted.
type VNetBGPRouterUpdateRequest struct {
	// ASN is the autonomous system number, from 1 to 4294967295.
	ASN *int `json:"asn,omitempty"`
}

func (r *VNetBGPRouterUpdateRequest) empty() bool {
	return r == nil || r.ASN == nil
}

// VNetBGPRouterCommand is one BGP router command (vnet_bgp_router_commands).
type VNetBGPRouterCommand struct {
	// Key is the command row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGPRouter is the parent BGP router key.
	BGPRouter FlexInt `json:"bgp_router,omitempty"`
	// OrderID is the command order.
	OrderID int `json:"orderid,omitempty"`
	// Enabled reports whether the command is enabled.
	Enabled bool `json:"enabled,omitempty"`
	// No reports whether the command is negated.
	No bool `json:"no,omitempty"`
	// Command is the command type (neighbor, network, redistribute, ...).
	Command string `json:"command,omitempty"`
	// Params are the command parameters.
	Params string `json:"params,omitempty"`
}

// VNetBGPRouterCommandCreateRequest creates a BGP router command.
// Nil bool pointers are omitted.
type VNetBGPRouterCommandCreateRequest struct {
	// BGPRouter is the parent BGP router key (required).
	BGPRouter int `json:"bgp_router"`
	// Command is the command type (required).
	Command string `json:"command"`
	// Params are the command parameters.
	Params string `json:"params"`
	// Enabled sets whether the command is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// No negates the command when true (API column "no").
	No *bool `json:"no,omitempty"`
}

// VNetBGPRouterCommandUpdateRequest updates a BGP router command.
// A nil field is omitted.
type VNetBGPRouterCommandUpdateRequest struct {
	// Command is the command type.
	Command *string `json:"command,omitempty"`
	// Params are the command parameters.
	Params *string `json:"params,omitempty"`
	// Enabled sets whether the command is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// No negates the command when true (API column "no").
	No *bool `json:"no,omitempty"`
}

func (r *VNetBGPRouterCommandUpdateRequest) empty() bool {
	return r == nil || (r.Command == nil && r.Params == nil && r.Enabled == nil && r.No == nil)
}

// VNetBGPInterface is a routing interface (vnet_bgp_interfaces).
// Creating one also creates an associated network and NIC on the server.
type VNetBGPInterface struct {
	// Key is the interface row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGP is the parent vnet_bgp key.
	BGP FlexInt `json:"bgp,omitempty"`
	// Name is the interface name.
	Name string `json:"name,omitempty"`
	// Description is the interface description.
	Description string `json:"description,omitempty"`
	// IPAddress is the interface address.
	IPAddress string `json:"ipaddress,omitempty"`
	// Network is the interface network in CIDR form.
	Network string `json:"network,omitempty"`
	// MTU is the interface MTU.
	MTU int `json:"mtu,omitempty"`
	// Layer2Type is "vlan" or "none".
	Layer2Type string `json:"layer2_type,omitempty"`
	// Layer2ID is the VLAN id when Layer2Type is vlan. Nil when unset.
	Layer2ID *FlexInt `json:"layer2_id,omitempty"`
	// InterfaceVNet is the uplink network key. Nil when unset.
	InterfaceVNet *FlexInt `json:"interface_vnet,omitempty"`
	// BGPVNet is the network the server created for this interface. Nil when unset.
	BGPVNet *FlexInt `json:"bgp_vnet,omitempty"`
	// NIC is the NIC the server created for this interface. Nil when unset.
	NIC *FlexInt `json:"nic,omitempty"`
}

// VNetBGPInterfaceCreateRequest creates a BGP interface.
// Nil pointers are omitted.
type VNetBGPInterfaceCreateRequest struct {
	// BGP is the parent vnet_bgp key (required).
	BGP int `json:"bgp"`
	// Name is the interface name (required).
	Name string `json:"name"`
	// IPAddress is the interface address (required).
	IPAddress string `json:"ipaddress"`
	// Network is the interface network in CIDR form (required).
	Network string `json:"network"`
	// InterfaceVNet is the uplink network key (required).
	InterfaceVNet int `json:"interface_vnet"`
	// Layer2Type is "vlan" or "none".
	Layer2Type *string `json:"layer2_type,omitempty"`
	// Layer2ID is the VLAN id.
	Layer2ID *int `json:"layer2_id,omitempty"`
	// MTU is the interface MTU, from 1000 to 65536.
	MTU *int `json:"mtu,omitempty"`
	// Description is the interface description.
	Description string `json:"description,omitempty"`
}

// VNetBGPInterfaceUpdateRequest updates a BGP interface.
// A nil field is omitted. Description can be set to an empty string.
type VNetBGPInterfaceUpdateRequest struct {
	// Name is the interface name.
	Name *string `json:"name,omitempty"`
	// IPAddress is the interface address.
	IPAddress *string `json:"ipaddress,omitempty"`
	// Network is the interface network in CIDR form.
	Network *string `json:"network,omitempty"`
	// InterfaceVNet is the uplink network key.
	InterfaceVNet *int `json:"interface_vnet,omitempty"`
	// Layer2Type is "vlan" or "none".
	Layer2Type *string `json:"layer2_type,omitempty"`
	// Layer2ID is the VLAN id.
	Layer2ID *int `json:"layer2_id,omitempty"`
	// MTU is the interface MTU, from 1000 to 65536.
	MTU *int `json:"mtu,omitempty"`
	// Description is the interface description.
	Description *string `json:"description,omitempty"`
}

func (r *VNetBGPInterfaceUpdateRequest) empty() bool {
	return r == nil || (r.Name == nil && r.IPAddress == nil && r.Network == nil &&
		r.InterfaceVNet == nil && r.Layer2Type == nil && r.Layer2ID == nil &&
		r.MTU == nil && r.Description == nil)
}

// VNetBGPInterfaceCommand is one interface command (vnet_bgp_interface_commands).
type VNetBGPInterfaceCommand struct {
	// Key is the command row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGPInterface is the parent interface key.
	BGPInterface FlexInt `json:"bgp_interface,omitempty"`
	// OrderID is the command order.
	OrderID int `json:"orderid,omitempty"`
	// Command is the command type (ip, ospf, bandwidth, ...).
	Command string `json:"command,omitempty"`
	// Params are the command parameters.
	Params string `json:"params,omitempty"`
}

// VNetBGPInterfaceCommandCreateRequest creates an interface command.
type VNetBGPInterfaceCommandCreateRequest struct {
	// BGPInterface is the parent interface key (required).
	BGPInterface int `json:"bgp_interface"`
	// Command is the command type (required).
	Command string `json:"command"`
	// Params are the command parameters.
	Params string `json:"params"`
}

// VNetBGPInterfaceCommandUpdateRequest updates an interface command.
// A nil field is omitted.
type VNetBGPInterfaceCommandUpdateRequest struct {
	// Command is the command type.
	Command *string `json:"command,omitempty"`
	// Params are the command parameters.
	Params *string `json:"params,omitempty"`
}

func (r *VNetBGPInterfaceCommandUpdateRequest) empty() bool {
	return r == nil || (r.Command == nil && r.Params == nil)
}

// VNetBGPRouteMap is one route map entry (vnet_bgp_routemaps).
// Entries that share a tag form one route map and are ordered by sequence.
type VNetBGPRouteMap struct {
	// Key is the route map row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGP is the parent vnet_bgp key.
	BGP FlexInt `json:"bgp,omitempty"`
	// Tag is the route map name.
	Tag string `json:"tag,omitempty"`
	// Permit is true for a permit entry and false for a deny entry.
	Permit bool `json:"permit,omitempty"`
	// Sequence is the entry sequence number.
	Sequence int `json:"sequence,omitempty"`
}

// VNetBGPRouteMapCreateRequest creates a route map entry.
// A nil Permit is omitted.
type VNetBGPRouteMapCreateRequest struct {
	// BGP is the parent vnet_bgp key (required).
	BGP int `json:"bgp"`
	// Tag is the route map name (required).
	Tag string `json:"tag"`
	// Sequence is the entry sequence, from 1 to 65535 (required).
	Sequence int `json:"sequence"`
	// Permit is true for permit and false for deny.
	Permit *bool `json:"permit,omitempty"`
}

// VNetBGPRouteMapUpdateRequest updates a route map entry.
// A nil field is omitted.
type VNetBGPRouteMapUpdateRequest struct {
	// Tag is the route map name.
	Tag *string `json:"tag,omitempty"`
	// Sequence is the entry sequence, from 1 to 65535.
	Sequence *int `json:"sequence,omitempty"`
	// Permit is true for permit and false for deny.
	Permit *bool `json:"permit,omitempty"`
}

func (r *VNetBGPRouteMapUpdateRequest) empty() bool {
	return r == nil || (r.Tag == nil && r.Sequence == nil && r.Permit == nil)
}

// VNetBGPRouteMapCommand is one route map command (vnet_bgp_routemap_commands).
type VNetBGPRouteMapCommand struct {
	// Key is the command row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGPRouteMap is the parent route map key.
	BGPRouteMap FlexInt `json:"bgp_routemap,omitempty"`
	// OrderID is the command order.
	OrderID int `json:"orderid,omitempty"`
	// Command is the command type (match, set, call, ...).
	Command string `json:"command,omitempty"`
	// Params are the command parameters.
	Params string `json:"params,omitempty"`
}

// VNetBGPRouteMapCommandCreateRequest creates a route map command.
type VNetBGPRouteMapCommandCreateRequest struct {
	// BGPRouteMap is the parent route map key (required).
	BGPRouteMap int `json:"bgp_routemap"`
	// Command is the command type (required).
	Command string `json:"command"`
	// Params are the command parameters.
	Params string `json:"params"`
}

// VNetBGPRouteMapCommandUpdateRequest updates a route map command.
// A nil field is omitted.
type VNetBGPRouteMapCommandUpdateRequest struct {
	// Command is the command type.
	Command *string `json:"command,omitempty"`
	// Params are the command parameters.
	Params *string `json:"params,omitempty"`
}

func (r *VNetBGPRouteMapCommandUpdateRequest) empty() bool {
	return r == nil || (r.Command == nil && r.Params == nil)
}

// VNetBGPIPCommand is an IP-level BGP command (vnet_bgp_ip), such as a
// prefix-list or an AS-path access list.
type VNetBGPIPCommand struct {
	// Key is the command row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGP is the parent vnet_bgp key.
	BGP FlexInt `json:"bgp,omitempty"`
	// OrderID is the command order.
	OrderID int `json:"orderid,omitempty"`
	// Command is the command type (prefix-list, as-path, ...).
	Command string `json:"command,omitempty"`
	// Params are the command parameters.
	Params string `json:"params,omitempty"`
}

// VNetBGPIPCommandCreateRequest creates a BGP IP command.
type VNetBGPIPCommandCreateRequest struct {
	// BGP is the parent vnet_bgp key (required).
	BGP int `json:"bgp"`
	// Command is the command type (required).
	Command string `json:"command"`
	// Params are the command parameters.
	Params string `json:"params"`
}

// VNetBGPIPCommandUpdateRequest updates a BGP IP command.
// A nil field is omitted.
type VNetBGPIPCommandUpdateRequest struct {
	// Command is the command type.
	Command *string `json:"command,omitempty"`
	// Params are the command parameters.
	Params *string `json:"params,omitempty"`
}

func (r *VNetBGPIPCommandUpdateRequest) empty() bool {
	return r == nil || (r.Command == nil && r.Params == nil)
}

// VNetOSPFCommand is an OSPF command (vnet_ospf_commands).
// OSPF commands use the same vnet_bgp parent as BGP.
type VNetOSPFCommand struct {
	// Key is the command row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGP is the parent vnet_bgp key.
	BGP FlexInt `json:"bgp,omitempty"`
	// OrderID is the command order.
	OrderID int `json:"orderid,omitempty"`
	// Command is the command type (network, area, router-id, ...).
	Command string `json:"command,omitempty"`
	// Params are the command parameters.
	Params string `json:"params,omitempty"`
}

// VNetOSPFCommandCreateRequest creates an OSPF command.
type VNetOSPFCommandCreateRequest struct {
	// BGP is the parent vnet_bgp key (required).
	BGP int `json:"bgp"`
	// Command is the command type (required).
	Command string `json:"command"`
	// Params are the command parameters.
	Params string `json:"params"`
}

// VNetOSPFCommandUpdateRequest updates an OSPF command.
// A nil field is omitted.
type VNetOSPFCommandUpdateRequest struct {
	// Command is the command type.
	Command *string `json:"command,omitempty"`
	// Params are the command parameters.
	Params *string `json:"params,omitempty"`
}

func (r *VNetOSPFCommandUpdateRequest) empty() bool {
	return r == nil || (r.Command == nil && r.Params == nil)
}

// VNetEIGRPRouter is an EIGRP router (vnet_eigrp_routers).
type VNetEIGRPRouter struct {
	// Key is the router row key.
	Key FlexInt `json:"$key,omitempty"`
	// BGP is the parent vnet_bgp key.
	BGP FlexInt `json:"bgp,omitempty"`
	// ASN is the EIGRP autonomous system number.
	ASN int `json:"asn,omitempty"`
}

// VNetEIGRPRouterCreateRequest creates an EIGRP router.
type VNetEIGRPRouterCreateRequest struct {
	// BGP is the parent vnet_bgp key (required).
	BGP int `json:"bgp"`
	// ASN is the autonomous system number, from 1 to 65535 (required).
	ASN int `json:"asn"`
}

// VNetEIGRPRouterUpdateRequest updates an EIGRP router.
// A nil field is omitted.
type VNetEIGRPRouterUpdateRequest struct {
	// ASN is the autonomous system number, from 1 to 65535.
	ASN *int `json:"asn,omitempty"`
}

func (r *VNetEIGRPRouterUpdateRequest) empty() bool {
	return r == nil || r.ASN == nil
}

// VNetEIGRPRouterCommand is one EIGRP router command (vnet_eigrp_router_commands).
type VNetEIGRPRouterCommand struct {
	// Key is the command row key.
	Key FlexInt `json:"$key,omitempty"`
	// EIGRPRouter is the parent EIGRP router key.
	EIGRPRouter FlexInt `json:"eigrp_router,omitempty"`
	// OrderID is the command order.
	OrderID int `json:"orderid,omitempty"`
	// Enabled reports whether the command is enabled.
	Enabled bool `json:"enabled,omitempty"`
	// No reports whether the command is negated.
	No bool `json:"no,omitempty"`
	// Command is the command type (network, redistribute, variance, ...).
	Command string `json:"command,omitempty"`
	// Params are the command parameters.
	Params string `json:"params,omitempty"`
}

// VNetEIGRPRouterCommandCreateRequest creates an EIGRP router command.
// Nil bool pointers are omitted.
type VNetEIGRPRouterCommandCreateRequest struct {
	// EIGRPRouter is the parent EIGRP router key (required).
	EIGRPRouter int `json:"eigrp_router"`
	// Command is the command type (required).
	Command string `json:"command"`
	// Params are the command parameters.
	Params string `json:"params"`
	// Enabled sets whether the command is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// No negates the command when true (API column "no").
	No *bool `json:"no,omitempty"`
}

// VNetEIGRPRouterCommandUpdateRequest updates an EIGRP router command.
// A nil field is omitted.
type VNetEIGRPRouterCommandUpdateRequest struct {
	// Command is the command type.
	Command *string `json:"command,omitempty"`
	// Params are the command parameters.
	Params *string `json:"params,omitempty"`
	// Enabled sets whether the command is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// No negates the command when true (API column "no").
	No *bool `json:"no,omitempty"`
}

func (r *VNetEIGRPRouterCommandUpdateRequest) empty() bool {
	return r == nil || (r.Command == nil && r.Params == nil && r.Enabled == nil && r.No == nil)
}
