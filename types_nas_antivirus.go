package vergeos

// NASServiceAntivirus is the antivirus configuration for one NAS service
// (vm_service_antivirus).
//
// VergeOS creates the row with the NAS service. The service needs 8GB of RAM
// or more before antivirus can run. NASService.Antivirus is this row's key.
type NASServiceAntivirus struct {
	// Key is the antivirus settings row key.
	Key FlexInt `json:"$key,omitempty"`
	// Service is the parent NAS service key.
	Service FlexInt `json:"service,omitempty"`
	// ServiceDisplay is the parent service display name.
	ServiceDisplay string `json:"service_display,omitempty"`
	// Enabled reports whether antivirus support is turned on.
	Enabled bool `json:"enabled"`
	// MaxRecursion is the directory recursion limit (0-100).
	MaxRecursion int `json:"max_recursion,omitempty"`
	// DatabasePrivateMirror is a private mirror URL for virus definitions.
	DatabasePrivateMirror string `json:"database_private_mirror,omitempty"`
	// DatabaseLocation is the virus database directory.
	DatabaseLocation string `json:"database_location,omitempty"`
	// DatabaseUpdatesEnabled reports whether definition updates are on.
	DatabaseUpdatesEnabled bool `json:"database_updates_enabled"`
}

// NASServiceAntivirusUpdateRequest updates one NAS service antivirus row.
//
// Nil fields are left unchanged. An empty request reads the row back and
// does not PUT.
type NASServiceAntivirusUpdateRequest struct {
	// Enabled turns antivirus support on or off.
	Enabled *bool `json:"enabled,omitempty"`
	// MaxRecursion is the directory recursion limit (0-100).
	MaxRecursion *int `json:"max_recursion,omitempty"`
	// DatabasePrivateMirror is a private mirror URL for virus definitions.
	DatabasePrivateMirror *string `json:"database_private_mirror,omitempty"`
	// DatabaseLocation is the virus database directory.
	DatabaseLocation *string `json:"database_location,omitempty"`
	// DatabaseUpdatesEnabled turns automatic definition updates on or off.
	DatabaseUpdatesEnabled *bool `json:"database_updates_enabled,omitempty"`
}

const nasServiceAntivirusListFields = "$key,service,service#$display as service_display,enabled,max_recursion,database_private_mirror,database_location,database_updates_enabled"
