package vergeos

import (
	"encoding/json"
	"time"
)

// VM import statuses reported by VergeOS.
const (
	VMImportStatusInitializing = "initializing"
	VMImportStatusImporting    = "importing"
	VMImportStatusComplete     = "complete"
	VMImportStatusAborted      = "aborted"
	VMImportStatusError        = "error"
	VMImportStatusWarning      = "warning"
)

// VM import log levels.
const (
	VMImportLogLevelMessage  = "message"
	VMImportLogLevelWarning  = "warning"
	VMImportLogLevelError    = "error"
	VMImportLogLevelCritical = "critical"
	VMImportLogLevelDebug    = "debug"
	VMImportLogLevelSummary  = "summary"
)

const (
	defaultVMImportWaitTimeout  = 10 * time.Minute
	defaultVMImportWaitInterval = 5 * time.Second
)

// VMImport is one vm_imports row.
//
// The key is a 40-character hex string. VergeOS keeps the row after the VM
// is deleted and does not require names to be unique, so the same name can
// appear on more than one row. GetByName refuses that. DeleteByName removes
// every finished row with the name.
type VMImport struct {
	// Key is the import's primary key.
	Key string `json:"$key,omitempty"`
	// ID is the same hex string as Key.
	ID string `json:"id,omitempty"`
	// Name is the VM name the import creates.
	Name string `json:"name,omitempty"`
	// VM is the created VM $key once the platform has assigned one.
	VM *FlexInt `json:"vm,omitempty"`
	// UUID is the source VM UUID.
	UUID string `json:"uuid,omitempty"`
	// File is the media-catalog file $key when the source is a file.
	File *FlexInt `json:"file,omitempty"`
	// Volume is the NAS volume key when the source is a volume.
	Volume string `json:"volume,omitempty"`
	// VolumePath is the path of the image inside Volume.
	VolumePath string `json:"volume_path,omitempty"`
	// SharedObject is the shared-object key for a tenant import.
	SharedObject *FlexInt `json:"shared_object,omitempty"`
	// Status is initializing, importing, complete, aborted, error, or warning.
	Status string `json:"status,omitempty"`
	// StatusInfo is the platform's detail for Status.
	StatusInfo string `json:"status_info,omitempty"`
	// Importing reports that the import is still running.
	Importing bool `json:"importing,omitempty"`
	// Aborted reports that the import was aborted.
	Aborted bool `json:"aborted,omitempty"`
	// FailedDriveCount is how many drives the platform has already failed.
	FailedDriveCount int `json:"failed_drive_count,omitempty"`
	// PreserveMACs keeps source MAC addresses.
	PreserveMACs bool `json:"preserve_macs,omitempty"`
	// PreserveDriveFormat keeps the source disk format.
	PreserveDriveFormat bool `json:"preserve_drive_format,omitempty"`
	// PreferredTier is the storage tier (1-5).
	PreferredTier string `json:"preferred_tier,omitempty"`
	// NoOpticalDrives skips optical drives from an OVA.
	NoOpticalDrives bool `json:"no_optical_drives,omitempty"`
	// OverrideDriveInterface replaces the source drive interface.
	OverrideDriveInterface string `json:"override_drive_interface,omitempty"`
	// OverrideNICInterface replaces the source NIC interface.
	OverrideNICInterface string `json:"override_nic_interface,omitempty"`
	// CleanupOnDelete removes the source path when the import row is deleted.
	CleanupOnDelete bool `json:"cleanup_on_delete,omitempty"`
	// Timestamp is when the row was created, as VergeOS reported it.
	Timestamp int64 `json:"timestamp,omitempty"`
	// Modified is when the row last changed, as VergeOS reported it.
	Modified int64 `json:"modified,omitempty"`
	// AlreadyExisted is set by Create when a VM with this name was already
	// present and Create did not post another import. Key is empty when
	// more than one import row has the name, or when the VM has no import row.
	AlreadyExisted bool `json:"-"`
}

// UnmarshalJSON accepts null foreign keys and ids that arrive as numbers or strings.
func (v *VMImport) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var err error
	if v.Key, err = jsonString(raw["$key"]); err != nil {
		return fieldErr("$key", err)
	}
	if v.ID, err = jsonString(raw["id"]); err != nil {
		return fieldErr("id", err)
	}
	if v.Name, err = jsonString(raw["name"]); err != nil {
		return fieldErr("name", err)
	}
	if v.VM, err = jsonFlexIntPtr(raw["vm"]); err != nil {
		return fieldErr("vm", err)
	}
	if v.UUID, err = jsonString(raw["uuid"]); err != nil {
		return fieldErr("uuid", err)
	}
	if v.File, err = jsonFlexIntPtr(raw["file"]); err != nil {
		return fieldErr("file", err)
	}
	if v.Volume, err = jsonKey(raw["volume"]); err != nil {
		return fieldErr("volume", err)
	}
	if v.VolumePath, err = jsonString(raw["volume_path"]); err != nil {
		return fieldErr("volume_path", err)
	}
	if v.SharedObject, err = jsonFlexIntPtr(raw["shared_object"]); err != nil {
		return fieldErr("shared_object", err)
	}
	if v.Status, err = jsonString(raw["status"]); err != nil {
		return fieldErr("status", err)
	}
	if v.StatusInfo, err = jsonString(raw["status_info"]); err != nil {
		return fieldErr("status_info", err)
	}
	if v.Importing, err = jsonBool(raw["importing"]); err != nil {
		return fieldErr("importing", err)
	}
	if v.Aborted, err = jsonBool(raw["aborted"]); err != nil {
		return fieldErr("aborted", err)
	}
	count, err := jsonInt64(raw["failed_drive_count"])
	if err != nil {
		return fieldErr("failed_drive_count", err)
	}
	v.FailedDriveCount = int(count)
	if v.PreserveMACs, err = jsonBool(raw["preserve_macs"]); err != nil {
		return fieldErr("preserve_macs", err)
	}
	if v.PreserveDriveFormat, err = jsonBool(raw["preserve_drive_format"]); err != nil {
		return fieldErr("preserve_drive_format", err)
	}
	if v.PreferredTier, err = jsonString(raw["preferred_tier"]); err != nil {
		return fieldErr("preferred_tier", err)
	}
	if v.NoOpticalDrives, err = jsonBool(raw["no_optical_drives"]); err != nil {
		return fieldErr("no_optical_drives", err)
	}
	if v.OverrideDriveInterface, err = jsonString(raw["override_drive_interface"]); err != nil {
		return fieldErr("override_drive_interface", err)
	}
	if v.OverrideNICInterface, err = jsonString(raw["override_nic_interface"]); err != nil {
		return fieldErr("override_nic_interface", err)
	}
	if v.CleanupOnDelete, err = jsonBool(raw["cleanup_on_delete"]); err != nil {
		return fieldErr("cleanup_on_delete", err)
	}
	if v.Timestamp, err = jsonInt64(raw["timestamp"]); err != nil {
		return fieldErr("timestamp", err)
	}
	if v.Modified, err = jsonInt64(raw["modified"]); err != nil {
		return fieldErr("modified", err)
	}
	return nil
}

// VMImportCreateRequest creates a vm_imports row from one source.
//
// Set File, URL, Volume, or SharedObject. File is a media-catalog file $key
// (an OVA, OVF, or a disk image such as qcow2, vmdk, vhd, or raw). URL is
// downloaded into the media catalog and then imported. Volume is a NAS
// volume key, with VolumePath naming the image on that volume.
//
// Importing defaults to true, so the platform starts the import as part of
// the create. Set it to false to create the row and call Start later.
//
// Create is idempotent for the VM name. When a non-snapshot VM with Name
// already exists, Create does not post another import: the platform rejects
// a duplicate VM name. The returned row has AlreadyExisted set.
type VMImportCreateRequest struct {
	// Name is the VM name (required).
	Name string
	// File is an existing media-catalog file $key.
	File *int
	// URL is an http or https address of an OVA or disk image.
	// Create downloads it into the media catalog, then imports that file.
	URL string
	// FileName is the catalog name used for URL. The default is the URL's file name.
	FileName string
	// Volume is a NAS volume key.
	Volume string
	// VolumePath is the image path on Volume.
	VolumePath string
	// SharedObject is a shared-object key.
	SharedObject *int
	// PreserveMACs keeps source MAC addresses. Nil leaves the platform default.
	PreserveMACs *bool
	// PreserveDriveFormat keeps the source disk format. Nil leaves the platform default.
	PreserveDriveFormat *bool
	// PreferredTier is the storage tier (1-5).
	PreferredTier string
	// NoOpticalDrives skips optical drives from an OVA.
	NoOpticalDrives *bool
	// OverrideDriveInterface replaces the source drive interface.
	OverrideDriveInterface string
	// OverrideNICInterface replaces the source NIC interface.
	OverrideNICInterface string
	// CleanupOnDelete removes the source path when the import row is deleted.
	CleanupOnDelete *bool
	// Importing starts the import during create. Nil means true.
	Importing *bool
}

// VMImportUpdateRequest updates an import row that has not been started,
// or records options the platform still accepts. Nil fields are left unchanged.
type VMImportUpdateRequest struct {
	Name                   *string `json:"name,omitempty"`
	PreserveMACs           *bool   `json:"preserve_macs,omitempty"`
	PreserveDriveFormat    *bool   `json:"preserve_drive_format,omitempty"`
	PreferredTier          *string `json:"preferred_tier,omitempty"`
	NoOpticalDrives        *bool   `json:"no_optical_drives,omitempty"`
	OverrideDriveInterface *string `json:"override_drive_interface,omitempty"`
	OverrideNICInterface   *string `json:"override_nic_interface,omitempty"`
}

// VMImportWaitOptions configures one Wait call.
// Zero Timeout and PollInterval use the defaults: 10 minutes, polled every 5 seconds.
type VMImportWaitOptions struct {
	Timeout      time.Duration
	PollInterval time.Duration
}

// VMImportLog is one vm_import_logs row.
type VMImportLog struct {
	// Key is the log row id.
	Key FlexInt `json:"$key,omitempty"`
	// VMImport is the parent import key.
	VMImport string `json:"vm_import,omitempty"`
	// Level is message, warning, error, critical, debug, or summary.
	Level string `json:"level,omitempty"`
	// Text is the log line.
	Text string `json:"text,omitempty"`
	// Timestamp is when the line was written, as VergeOS reported it.
	Timestamp int64 `json:"timestamp,omitempty"`
	// User is who the platform recorded for the line.
	User string `json:"user,omitempty"`
}

// UnmarshalJSON accepts a log key and user that arrive as numbers or strings.
func (l *VMImportLog) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	key, err := jsonFlexIntPtr(raw["$key"])
	if err != nil {
		return fieldErr("$key", err)
	}
	if key != nil {
		l.Key = *key
	}
	if l.VMImport, err = jsonKey(raw["vm_import"]); err != nil {
		return fieldErr("vm_import", err)
	}
	if l.Level, err = jsonString(raw["level"]); err != nil {
		return fieldErr("level", err)
	}
	if l.Text, err = jsonString(raw["text"]); err != nil {
		return fieldErr("text", err)
	}
	if l.Timestamp, err = jsonInt64(raw["timestamp"]); err != nil {
		return fieldErr("timestamp", err)
	}
	if l.User, err = jsonString(raw["user"]); err != nil {
		return fieldErr("user", err)
	}
	return nil
}

const vmImportFields = "$key,id,name,vm,uuid,file,volume,volume_path,shared_object," +
	"status,status_info,importing,aborted,failed_drive_count,preserve_macs,preserve_drive_format," +
	"preferred_tier,no_optical_drives,override_drive_interface,override_nic_interface," +
	"cleanup_on_delete,timestamp,modified"

const vmImportLogFields = "$key,vm_import,level,text,timestamp,user"
