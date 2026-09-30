package vergeos

import (
	"encoding/json"
	"time"
)

// VM export statuses reported by VergeOS on volume_vm_exports.
const (
	VMExportStatusIdle     = "idle"
	VMExportStatusBuilding = "building"
	VMExportStatusError    = "error"
	VMExportStatusCleaning = "cleaning"
)

const (
	defaultVMExportWaitTimeout  = time.Hour
	defaultVMExportWaitInterval = 5 * time.Second
)

// VMExport is the VM-export configuration for one NAS volume.
//
// There is one row per volume. The key is an integer row id. Volume is the
// NAS volume's SHA1 key. An export writes VM images onto that volume.
type VMExport struct {
	// Key is the export row id.
	Key FlexInt `json:"$key,omitempty"`
	// Volume is the NAS volume key the files are written to.
	Volume string `json:"volume,omitempty"`
	// VolumeName is the volume's name.
	VolumeName string `json:"volume_name,omitempty"`
	// Quiesced reports whether guests are frozen during the export.
	Quiesced bool `json:"quiesced,omitempty"`
	// Status is idle, building, error, or cleaning.
	Status string `json:"status,omitempty"`
	// StatusInfo is the platform's detail for Status.
	StatusInfo string `json:"status_info,omitempty"`
	// CreateCurrent keeps a current folder pointing at the latest export.
	CreateCurrent bool `json:"create_current,omitempty"`
	// MaxExports is how many generations to keep (1-100).
	MaxExports int `json:"max_exports,omitempty"`
}

// UnmarshalJSON accepts a null volume and bools that arrive as 0/1 or strings.
func (v *VMExport) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	key, err := jsonFlexIntPtr(raw["$key"])
	if err != nil {
		return fieldErr("$key", err)
	}
	if key != nil {
		v.Key = *key
	}
	if v.Volume, err = jsonKey(raw["volume"]); err != nil {
		return fieldErr("volume", err)
	}
	if v.VolumeName, err = jsonString(raw["volume_name"]); err != nil {
		return fieldErr("volume_name", err)
	}
	if v.Quiesced, err = jsonBool(raw["quiesced"]); err != nil {
		return fieldErr("quiesced", err)
	}
	if v.Status, err = jsonString(raw["status"]); err != nil {
		return fieldErr("status", err)
	}
	if v.StatusInfo, err = jsonString(raw["status_info"]); err != nil {
		return fieldErr("status_info", err)
	}
	if v.CreateCurrent, err = jsonBool(raw["create_current"]); err != nil {
		return fieldErr("create_current", err)
	}
	maxExports, err := jsonInt64(raw["max_exports"])
	if err != nil {
		return fieldErr("max_exports", err)
	}
	v.MaxExports = int(maxExports)
	return nil
}

// VMExportCreateRequest creates the export configuration for a NAS volume.
// Nil option pointers are omitted so the platform default applies.
type VMExportCreateRequest struct {
	// Volume is the NAS volume key (required).
	Volume string
	// Quiesced freezes guest filesystems during the export.
	Quiesced *bool
	// CreateCurrent keeps a current folder for the latest export.
	CreateCurrent *bool
	// MaxExports is how many generations to keep, from 1 to 100.
	MaxExports *int
}

// VMExportUpdateRequest updates an export configuration. Nil fields are left unchanged.
type VMExportUpdateRequest struct {
	Quiesced      *bool `json:"quiesced,omitempty"`
	CreateCurrent *bool `json:"create_current,omitempty"`
	MaxExports    *int  `json:"max_exports,omitempty"`
}

// VMExportStart is the payload for one export run.
// An empty VMs list exports the VMs the volume is configured to include.
type VMExportStart struct {
	// Name is the folder name for this run.
	Name string
	// VMs are VM $keys to export.
	VMs []int
}

// VMExportRunRequest ensures the volume's export configuration and starts a run.
type VMExportRunRequest struct {
	// Volume is the NAS volume key (required).
	Volume string
	// VMs are VM $keys to export. Empty exports the volume's configured set.
	VMs []int
	// Name is the folder name for this run.
	Name string
	// Quiesced, CreateCurrent, and MaxExports are applied to the configuration.
	Quiesced      *bool
	CreateCurrent *bool
	MaxExports    *int
}

// VMExportWaitOptions configures one Wait call.
// Zero Timeout and PollInterval use the defaults: 1 hour, polled every 5 seconds.
type VMExportWaitOptions struct {
	Timeout      time.Duration
	PollInterval time.Duration
}

// VMExportStat is one volume_vm_export_stats row, the result of a finished run.
type VMExportStat struct {
	Key             FlexInt `json:"$key,omitempty"`
	Export          FlexInt `json:"volume_vm_exports,omitempty"`
	Duration        int     `json:"duration,omitempty"`
	VirtualMachines int     `json:"virtual_machines,omitempty"`
	ExportSuccess   int     `json:"export_success,omitempty"`
	Errors          int     `json:"errors,omitempty"`
	Quiesced        bool    `json:"quiesced,omitempty"`
	SizeBytes       int64   `json:"size_bytes,omitempty"`
	FileName        string  `json:"file_name,omitempty"`
	Timestamp       int64   `json:"timestamp,omitempty"`
}

// UnmarshalJSON accepts null counters.
func (s *VMExportStat) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	key, err := jsonFlexIntPtr(raw["$key"])
	if err != nil {
		return fieldErr("$key", err)
	}
	if key != nil {
		s.Key = *key
	}
	exportID, err := jsonFlexIntPtr(raw["volume_vm_exports"])
	if err != nil {
		return fieldErr("volume_vm_exports", err)
	}
	if exportID != nil {
		s.Export = *exportID
	}
	n, err := jsonInt64(raw["duration"])
	if err != nil {
		return fieldErr("duration", err)
	}
	s.Duration = int(n)
	if n, err = jsonInt64(raw["virtual_machines"]); err != nil {
		return fieldErr("virtual_machines", err)
	}
	s.VirtualMachines = int(n)
	if n, err = jsonInt64(raw["export_success"]); err != nil {
		return fieldErr("export_success", err)
	}
	s.ExportSuccess = int(n)
	if n, err = jsonInt64(raw["errors"]); err != nil {
		return fieldErr("errors", err)
	}
	s.Errors = int(n)
	if s.Quiesced, err = jsonBool(raw["quiesced"]); err != nil {
		return fieldErr("quiesced", err)
	}
	if s.SizeBytes, err = jsonInt64(raw["size_bytes"]); err != nil {
		return fieldErr("size_bytes", err)
	}
	if s.FileName, err = jsonString(raw["file_name"]); err != nil {
		return fieldErr("file_name", err)
	}
	if s.Timestamp, err = jsonInt64(raw["timestamp"]); err != nil {
		return fieldErr("timestamp", err)
	}
	return nil
}

const vmExportFields = "$key,volume,volume#name as volume_name,quiesced,status,status_info,create_current,max_exports"

const vmExportStatFields = "$key,volume_vm_exports,duration,virtual_machines,export_success,errors,quiesced,size_bytes,file_name,timestamp"
