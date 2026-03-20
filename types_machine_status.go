package vergeos

import "encoding/json"

// MachineStatus represents the runtime status of a VergeOS machine (VM).
// This includes power state, node assignment, and guest agent information.
type MachineStatus struct {
	// Key is the status row key.
	Key FlexInt `json:"$key,omitempty"`
	// Machine is the machine ID this status belongs to.
	Machine int `json:"machine,omitempty"`
	// Running indicates whether the machine is currently running.
	Running bool `json:"running"`
	// Migratable indicates whether the machine can be migrated.
	Migratable bool `json:"migratable"`
	// Status is the machine status string (running, stopped, etc.).
	Status string `json:"status,omitempty"`
	// StatusInfo provides additional status details.
	StatusInfo string `json:"status_info,omitempty"`
	// State is the overall state (online, offline, warning, error).
	State string `json:"state,omitempty"`
	// PowerState indicates the power state of the machine.
	PowerState bool `json:"powerstate"`
	// Node is the node ID where the machine is running.
	Node FlexInt `json:"node,omitempty"`
	// NodeName is the display name of the node.
	NodeName string `json:"node_name,omitempty"`
	// MigratedNode is the node ID after migration.
	MigratedNode FlexInt `json:"migrated_node,omitempty"`
	// MigrationDestination is the target node for in-progress migration.
	MigrationDestination FlexInt `json:"migration_destination,omitempty"`
	// Started is the timestamp when the machine was started (Unix epoch).
	Started int64 `json:"started,omitempty"`
	// LocalTime is the local time reported by the machine (Unix epoch).
	LocalTime int64 `json:"local_time,omitempty"`
	// LastUpdate is the timestamp of the last status update (Unix epoch).
	LastUpdate int64 `json:"last_update,omitempty"`
	// RunningCores is the number of CPU cores allocated to the running machine.
	RunningCores int `json:"running_cores,omitempty"`
	// RunningRAM is the RAM in MB allocated to the running machine.
	RunningRAM int `json:"running_ram,omitempty"`
	// AgentVersion is the guest agent version string.
	AgentVersion string `json:"agent_version,omitempty"`
	// AgentFeatures contains guest agent supported features.
	AgentFeatures json.RawMessage `json:"agent_features,omitempty"`
	// AgentGuestInfo contains guest OS information reported by the agent,
	// including network configuration, filesystem info, memory, and hostname.
	AgentGuestInfo *GuestInfo `json:"agent_guest_info,omitempty"`
}

// GuestInfo contains guest OS information reported by the VergeOS guest agent.
type GuestInfo struct {
	// OSInfo contains operating system details.
	OSInfo *GuestOSInfo `json:"osinfo,omitempty"`
	// Network contains network interface information from the guest agent.
	Network []GuestNetworkInterface `json:"network,omitempty"`
	// FSInfo contains filesystem information from the guest agent.
	FSInfo []GuestFSInfo `json:"fsinfo,omitempty"`
	// MemInfo contains memory usage information.
	MemInfo *GuestMemInfo `json:"meminfo,omitempty"`
	// Hostname is the guest hostname.
	Hostname string `json:"hostname,omitempty"`
	// LastRefresh is the timestamp of the last agent refresh (Unix epoch).
	LastRefresh int64 `json:"last_refresh,omitempty"`
}

// GuestOSInfo contains operating system information from the guest agent.
type GuestOSInfo struct {
	Name          string `json:"name,omitempty"`
	PrettyName    string `json:"pretty-name,omitempty"`
	Version       string `json:"version,omitempty"`
	ID            string `json:"id,omitempty"`
	KernelRelease string `json:"kernel-release,omitempty"`
	KernelVersion string `json:"kernel-version,omitempty"`
	Machine       string `json:"machine,omitempty"`
}

// GuestNetworkInterface contains network interface information from the guest agent.
type GuestNetworkInterface struct {
	// Name is the interface name (e.g., "enp1s1", "eth0").
	Name string `json:"name,omitempty"`
	// HardwareAddress is the MAC address.
	HardwareAddress string `json:"hardware-address,omitempty"`
	// IPAddresses contains the IP addresses assigned to this interface.
	IPAddresses []GuestIPAddress `json:"ip-addresses,omitempty"`
}

// GuestIPAddress represents an IP address on a guest network interface.
type GuestIPAddress struct {
	// Type is the address type ("ipv4" or "ipv6").
	Type string `json:"ip-address-type,omitempty"`
	// Address is the IP address string.
	Address string `json:"ip-address,omitempty"`
	// Prefix is the subnet prefix length.
	Prefix int `json:"prefix,omitempty"`
}

// GuestFSInfo contains filesystem information from the guest agent.
type GuestFSInfo struct {
	// Name is the device name.
	Name string `json:"name,omitempty"`
	// Type is the filesystem type (e.g., "ext4").
	Type string `json:"type,omitempty"`
	// TotalBytes is the total filesystem size in bytes.
	TotalBytes int64 `json:"total-bytes,omitempty"`
	// UsedBytes is the used space in bytes.
	UsedBytes int64 `json:"used-bytes,omitempty"`
	// MountPoint is the mount point path.
	MountPoint string `json:"mountpoint,omitempty"`
}

// GuestMemInfo contains memory information from the guest agent.
type GuestMemInfo struct {
	// RAMTotal is the total RAM in MB.
	RAMTotal int `json:"ram_total,omitempty"`
	// RAMUsed is the used RAM in MB.
	RAMUsed int `json:"ram_used,omitempty"`
	// VRAMTotal is the total VRAM in MB.
	VRAMTotal int `json:"vram_total,omitempty"`
	// VRAMUsed is the used VRAM in MB.
	VRAMUsed int `json:"vram_used,omitempty"`
}

// machineStatusListFields are the fields to request when listing machine status.
const machineStatusListFields = "$key,machine,running,migratable,status,status_info,state,powerstate,node,node#name as node_name,migrated_node,migration_destination,started,local_time,last_update,running_cores,running_ram,agent_version,agent_features,agent_guest_info"

// machineStatusGetFields are the fields to request when getting a single machine status.
const machineStatusGetFields = machineStatusListFields
