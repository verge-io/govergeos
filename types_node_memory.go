package vergeos

// NodeMemory is one physical memory module on a node (node_memory).
//
// Rows are discovered by the platform and are read-only. Status is
// "online", "error", "warning", or "offline".
type NodeMemory struct {
	// Key is the DIMM row key.
	Key FlexInt `json:"$key,omitempty"`
	// Node is the parent node key.
	Node int `json:"node,omitempty"`
	// Status is the DIMM status.
	Status string `json:"status,omitempty"`
	// StatusInfo is extra status detail.
	StatusInfo string `json:"status_info,omitempty"`
	// Label is the DIMM label.
	Label string `json:"label,omitempty"`
	// Locator is the slot locator (for example "DIMM 0").
	Locator string `json:"locator,omitempty"`
	// BankLocator is the bank locator (for example "P0 CHANNEL A").
	BankLocator string `json:"bank_locator,omitempty"`
	// Type is the memory type (for example "DDR5").
	Type string `json:"type,omitempty"`
	// TypeDetail is the memory type detail.
	TypeDetail string `json:"type_detail,omitempty"`
	// Size is the module size (for example "48 GB").
	Size string `json:"size,omitempty"`
	// Speed is the memory speed (for example "5600 MT/s").
	Speed string `json:"speed,omitempty"`
	// ConfiguredMemorySpeed is the configured speed.
	ConfiguredMemorySpeed string `json:"configured_memory_speed,omitempty"`
	// FormFactor is the form factor (for example "DIMM" or "SODIMM").
	FormFactor string `json:"form_factor,omitempty"`
	// Manufacturer is the module manufacturer.
	Manufacturer string `json:"manufacturer,omitempty"`
	// SerialNumber is the module serial number.
	SerialNumber string `json:"serial_number,omitempty"`
	// PartNumber is the part number.
	PartNumber string `json:"part_number,omitempty"`
	// AssetTag is the asset tag.
	AssetTag string `json:"asset_tag,omitempty"`
	// Rank is the memory rank.
	Rank string `json:"rank,omitempty"`
	// DataWidth is the data width.
	DataWidth string `json:"data_width,omitempty"`
	// TotalWidth is the total width, including ECC bits when present.
	TotalWidth string `json:"total_width,omitempty"`
	// MemoryTechnology is the technology (for example "DRAM").
	MemoryTechnology string `json:"memory_technology,omitempty"`
	// Modified is the last update time as a Unix timestamp.
	Modified int64 `json:"modified,omitempty"`
}

// IsHealthy reports whether the DIMM status is online.
func (m NodeMemory) IsHealthy() bool {
	return m.Status == "online"
}

const nodeMemoryListFields = "$key,node,status,status_info,label,locator,bank_locator,type,type_detail,size,speed,configured_memory_speed,form_factor,manufacturer,serial_number,part_number,asset_tag,rank,data_width,total_width,memory_technology,modified"
