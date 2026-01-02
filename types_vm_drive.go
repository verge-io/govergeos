package vergeos

// VMDrive represents a virtual disk attached to a VM.
type VMDrive struct {
	// ID is the unique identifier for the drive.
	ID FlexInt `json:"$key,omitempty"`
	// Machine is the machine reference ID.
	Machine int `json:"machine,omitempty"`
	// Name is the drive name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Interface is the disk interface. Valid values:
	//   - virtio-scsi (default, recommended)
	//   - virtio (legacy paravirtual)
	//   - scsi
	//   - ide
	//   - sata
	//   - nvme
	//   - usb
	//   - pflash
	//   - direct
	//   - tpm_state
	Interface string `json:"interface,omitempty"`
	// Media is the media type (disk, cdrom, import).
	Media string `json:"media,omitempty"`
	// MediaSource is the media source ID (for cdrom/import).
	MediaSource int `json:"media_source,omitempty"`
	// SizeGB is the disk size in GB.
	SizeGB int64 `json:"-"`
	// sizeBytes is the internal disk size in bytes.
	SizeBytes int64 `json:"disksize,omitempty"`
	// PreferredTier is the preferred storage tier.
	PreferredTier string `json:"preferred_tier,omitempty"`
	// Enabled indicates whether the drive is enabled.
	Enabled bool `json:"enabled"`
	// ReadOnly indicates whether the drive is read-only.
	ReadOnly bool `json:"readonly,omitempty"`
	// Serial is the drive serial number.
	Serial string `json:"serial,omitempty"`
	// Asset is the asset tag.
	Asset string `json:"asset,omitempty"`
	// OrderID is the boot order ID.
	OrderID int `json:"orderid,omitempty"`
	// PreserveDriveFormat indicates whether to preserve the drive format.
	PreserveDriveFormat bool `json:"preserve_drive_format,omitempty"`
	// PowerState is the drive power state ("online" or "offline").
	PowerState string `json:"powerState,omitempty"`
	// Status is the drive status (e.g., "importing").
	Status string `json:"status,omitempty"`
}

// VMDriveCreateRequest is the request body for creating a drive.
type VMDriveCreateRequest struct {
	// Machine is the VM's machine ID.
	Machine int `json:"machine"`
	// Name is the drive name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Interface is the disk interface. Valid values: virtio-scsi (default), virtio, scsi, ide, sata, nvme, usb.
	Interface string `json:"interface,omitempty"`
	// Media is the media type (disk, cdrom, import).
	Media string `json:"media,omitempty"`
	// MediaSource is the media source ID (for cdrom/import).
	MediaSource int `json:"media_source,omitempty"`
	// SizeGB is the disk size in GB.
	SizeGB int64 `json:"-"`
	// sizeBytes is the internal disk size in bytes (set from SizeGB).
	SizeBytes int64 `json:"disksize,omitempty"`
	// PreferredTier is the preferred storage tier.
	PreferredTier string `json:"preferred_tier,omitempty"`
	// Enabled indicates whether the drive is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// ReadOnly indicates whether the drive is read-only.
	ReadOnly *bool `json:"readonly,omitempty"`
	// Serial is the drive serial number.
	Serial string `json:"serial,omitempty"`
	// Asset is the asset tag.
	Asset string `json:"asset,omitempty"`
	// OrderID is the boot order ID.
	OrderID *int `json:"orderid,omitempty"`
	// PreserveDriveFormat indicates whether to preserve the drive format.
	PreserveDriveFormat *bool `json:"preserve_drive_format,omitempty"`
}

// VMDriveUpdateRequest is the request body for updating a drive.
type VMDriveUpdateRequest struct {
	// Name is the drive name.
	Name *string `json:"name,omitempty"`
	// Description is the description.
	Description *string `json:"description,omitempty"`
	// Interface is the disk interface. Valid values: virtio-scsi (default), virtio, scsi, ide, sata, nvme, usb.
	Interface *string `json:"interface,omitempty"`
	// SizeGB is the disk size in GB (can only increase).
	SizeGB *int64 `json:"-"`
	// sizeBytes is the internal disk size in bytes.
	SizeBytes *int64 `json:"disksize,omitempty"`
	// PreferredTier is the preferred storage tier.
	PreferredTier *string `json:"preferred_tier,omitempty"`
	// Enabled indicates whether the drive is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// ReadOnly indicates whether the drive is read-only.
	ReadOnly *bool `json:"readonly,omitempty"`
	// Serial is the drive serial number.
	Serial *string `json:"serial,omitempty"`
	// Asset is the asset tag.
	Asset *string `json:"asset,omitempty"`
	// OrderID is the boot order ID.
	OrderID *int `json:"orderid,omitempty"`
	// PreserveDriveFormat indicates whether to preserve the drive format.
	PreserveDriveFormat *bool `json:"preserve_drive_format,omitempty"`
}

// driveListFields are the fields to request when listing drives.
const driveListFields = "$key,machine,name,description,disksize,interface,media,media_source,preferred_tier,enabled,readonly,serial,preserve_drive_format,asset,orderid"

// driveGetFields are the fields to request when getting a single drive (includes power state).
const driveGetFields = driveListFields + ",status#status as powerState"

// Conversion constants for disk size
const (
	// bytesPerGB is the number of bytes in a gigabyte.
	bytesPerGB = 1024 * 1024 * 1024
)
