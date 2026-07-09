package vergeos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

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
// Network and FSInfo are normalized to slices, but the VergeOS API may return
// them as either JSON arrays or JSON objects (maps keyed by string) depending
// on the VM and guest agent version. The custom UnmarshalJSON handles both.
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

// UnmarshalJSON handles the VergeOS API inconsistency where network and fsinfo
// can be either a JSON array or a JSON object (map keyed by string).
func (g *GuestInfo) UnmarshalJSON(data []byte) error {
	// The API sends [] (empty array) instead of an object or null when no
	// guest agent is running. Detect that case and return a zero-value.
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] == '[' {
		return nil
	}

	// Use an alias to avoid infinite recursion.
	type guestInfoRaw struct {
		OSInfo      *GuestOSInfo    `json:"osinfo,omitempty"`
		Network     json.RawMessage `json:"network,omitempty"`
		FSInfo      json.RawMessage `json:"fsinfo,omitempty"`
		MemInfo     *GuestMemInfo   `json:"meminfo,omitempty"`
		Hostname    string          `json:"hostname,omitempty"`
		LastRefresh int64           `json:"last_refresh,omitempty"`
	}

	var raw guestInfoRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	g.OSInfo = raw.OSInfo
	g.MemInfo = raw.MemInfo
	g.Hostname = raw.Hostname
	g.LastRefresh = raw.LastRefresh

	if len(raw.Network) > 0 {
		var err error
		g.Network, err = unmarshalFlexList[GuestNetworkInterface](raw.Network)
		if err != nil {
			return fmt.Errorf("unmarshaling network: %w", err)
		}
	}

	if len(raw.FSInfo) > 0 {
		var err error
		g.FSInfo, err = unmarshalFlexList[GuestFSInfo](raw.FSInfo)
		if err != nil {
			return fmt.Errorf("unmarshaling fsinfo: %w", err)
		}
	}

	return nil
}

// unmarshalFlexList unmarshals a JSON value that may be either an array or
// an object (map keyed by string) into a slice. If it's an object, the map
// values are collected into a slice (keys are discarded).
func unmarshalFlexList[T any](data json.RawMessage) ([]T, error) {
	// Try array first.
	var arr []T
	if err := json.Unmarshal(data, &arr); err == nil {
		return arr, nil
	}

	// Try object (map keyed by string).
	var m map[string]T
	if err := json.Unmarshal(data, &m); err == nil {
		result := make([]T, 0, len(m))
		for _, v := range m {
			result = append(result, v)
		}
		return result, nil
	}

	return nil, fmt.Errorf("expected JSON array or object, got: %.50s", data)
}

// IPsForMAC returns the IP addresses of the guest interfaces whose hardware
// address matches mac (case-insensitive). This is the reliable way to find a
// VM's real IP: match on the MAC of a known VM NIC rather than interface
// names, which are OS-dependent ("eth0", "ens1", "Ethernet 2") and drowned
// out by container bridges and veths on busy guests. Returns nil when the
// MAC is not found or g is nil.
func (g *GuestInfo) IPsForMAC(mac string) []GuestIPAddress {
	if g == nil {
		return nil
	}
	var ips []GuestIPAddress
	for _, nic := range g.Network {
		if strings.EqualFold(nic.HardwareAddress, mac) {
			ips = append(ips, nic.IPAddresses...)
		}
	}
	return ips
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
