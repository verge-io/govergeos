package vergeos

// VMNIC represents a network interface attached to a VM.
type VMNIC struct {
	// ID is the unique identifier for the NIC.
	ID FlexInt `json:"$key,omitempty"`
	// Machine is the machine reference ID.
	Machine int `json:"machine,omitempty"`
	// Name is the NIC name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Interface is the interface type.
	Interface string `json:"interface,omitempty"`
	// Driver is the network driver.
	Driver string `json:"driver,omitempty"`
	// Model is the NIC model.
	Model string `json:"model,omitempty"`
	// Vendor is the NIC vendor.
	Vendor string `json:"vendor,omitempty"`
	// Port is the port number.
	Port int `json:"port,omitempty"`
	// Enabled indicates whether the NIC is enabled.
	Enabled bool `json:"enabled"`
	// VNET is the virtual network ID.
	VNET int `json:"vnet,omitempty"`
	// MAC is the MAC address.
	MAC string `json:"macaddress,omitempty"`
	// IPAddress is the assigned IP address.
	IPAddress string `json:"ipaddress,omitempty"`
	// Asset is the asset tag.
	Asset string `json:"asset,omitempty"`
	// PowerState is the NIC power state ("up" or "down").
	PowerState string `json:"powerState,omitempty"`
}

// VMNICCreateRequest is the request body for creating a NIC.
type VMNICCreateRequest struct {
	// Machine is the VM's machine ID.
	Machine int `json:"machine"`
	// Name is the NIC name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Interface is the interface type.
	Interface string `json:"interface,omitempty"`
	// Driver is the network driver.
	Driver string `json:"driver,omitempty"`
	// Model is the NIC model.
	Model string `json:"model,omitempty"`
	// Vendor is the NIC vendor.
	Vendor string `json:"vendor,omitempty"`
	// Port is the port number.
	Port *int `json:"port,omitempty"`
	// Enabled indicates whether the NIC is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// VNET is the virtual network ID.
	VNET int `json:"vnet,omitempty"`
	// MAC is the MAC address (optional, auto-generated if not specified).
	MAC string `json:"macaddress,omitempty"`
	// AssignIPAddress indicates whether to assign an IP address.
	AssignIPAddress bool `json:"-"` // Handled separately via vnet_addresses
	// Asset is the asset tag.
	Asset string `json:"asset,omitempty"`
}

// VMNICUpdateRequest is the request body for updating a NIC.
type VMNICUpdateRequest struct {
	// Name is the NIC name.
	Name *string `json:"name,omitempty"`
	// Description is the description.
	Description *string `json:"description,omitempty"`
	// Interface is the interface type.
	Interface *string `json:"interface,omitempty"`
	// Driver is the network driver.
	Driver *string `json:"driver,omitempty"`
	// Model is the NIC model.
	Model *string `json:"model,omitempty"`
	// Vendor is the NIC vendor.
	Vendor *string `json:"vendor,omitempty"`
	// Port is the port number.
	Port *int `json:"port,omitempty"`
	// Enabled indicates whether the NIC is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// VNET is the virtual network ID.
	VNET *int `json:"vnet,omitempty"`
	// Asset is the asset tag.
	Asset *string `json:"asset,omitempty"`
}

// vnetAddressRequest is the request body for assigning an IP address to a NIC.
type vnetAddressRequest struct {
	VNET int    `json:"vnet"`
	MAC  string `json:"macaddress"`
	Type string `json:"type"`
}

// nicListFields are the fields to request when listing NICs.
const nicListFields = "$key,machine,name,description,interface,driver,model,vendor,port,enabled,vnet,macaddress,asset"

// nicGetFields are the fields to request when getting a single NIC (includes power state).
const nicGetFields = nicListFields + ",status#status as powerState"
