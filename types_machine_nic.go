package vergeos

// MachineNIC represents a physical network interface on a machine.
type MachineNIC struct {
	// Key is the unique identifier for this NIC record.
	Key FlexInt `json:"$key,omitempty"`
	// Machine is the parent machine ID.
	Machine int `json:"machine,omitempty"`
	// Name is the interface name (e.g., "eno1", "eth0").
	Name string `json:"name,omitempty"`
	// Stats contains the NIC traffic statistics (populated via stats[all] expansion).
	Stats *MachineNICStats `json:"stats,omitempty"`
	// Status contains the NIC link status (populated via status[all] expansion).
	Status *MachineNICStatus `json:"status,omitempty"`
}

// MachineNICStats represents traffic counters for a machine NIC.
type MachineNICStats struct {
	// Key is the unique identifier for this stats record.
	Key FlexInt `json:"$key,omitempty"`
	// TxPckts is the total transmitted packets.
	TxPckts uint64 `json:"tx_pckts,omitempty"`
	// RxPckts is the total received packets.
	RxPckts uint64 `json:"rx_pckts,omitempty"`
	// TxBytes is the total transmitted bytes.
	TxBytes uint64 `json:"tx_bytes,omitempty"`
	// RxBytes is the total received bytes.
	RxBytes uint64 `json:"rx_bytes,omitempty"`
}

// MachineNICStatus represents the link status of a machine NIC.
type MachineNICStatus struct {
	// Key is the unique identifier for this status record.
	Key FlexInt `json:"$key,omitempty"`
	// Status is the link state ("up", "down", "unknown", "lowerlayerdown").
	Status string `json:"status,omitempty"`
	// Speed is the link speed in Mbps.
	Speed uint32 `json:"speed,omitempty"`
}

// Field list constants for machine NICs
const (
	machineNICListFields = "$key,machine,name,stats[all],status[all]"
)
