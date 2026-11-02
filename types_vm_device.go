package vergeos

import (
	"encoding/json"
	"strconv"
)

// DeviceResourceGroup is the resource_groups key stored on a machine device.
// The key is a UUID string. A JSON number still decodes, and a decimal value
// is written back as a number so an integer key posts the way it used to.
type DeviceResourceGroup string

// UnmarshalJSON accepts a UUID string, a JSON number, or null.
func (g *DeviceResourceGroup) UnmarshalJSON(data []byte) error {
	var text FlexText
	if err := text.UnmarshalJSON(data); err != nil {
		return err
	}
	*g = DeviceResourceGroup(text)
	return nil
}

// MarshalJSON writes a UUID as a string and a decimal key as a number.
func (g DeviceResourceGroup) MarshalJSON() ([]byte, error) {
	s := string(g)
	if s == "" {
		return []byte(`""`), nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && strconv.FormatInt(n, 10) == s {
		return json.Marshal(n)
	}
	return json.Marshal(s)
}

// String returns the key text.
func (g DeviceResourceGroup) String() string {
	return string(g)
}

// VMDevice represents a device attached to a machine.
// VMDevices and TenantNodeDevices both use this row (machine_devices).
type VMDevice struct {
	// Key is the unique identifier for the device ($key).
	Key FlexInt `json:"$key,omitempty"`
	// Machine is the machine reference ID.
	Machine int `json:"machine,omitempty"`
	// MachineType is the machine kind stored on the device (vm, node, vnet, tenant).
	// DeviceMachineTypeTenant is the tenant node value.
	MachineType string `json:"machine_type,omitempty"`
	// OrderID is the device order (0-64).
	OrderID int `json:"orderid,omitempty"`
	// Type is the device type. See the DeviceType* constants.
	Type string `json:"type"`
	// Name is the device name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// ResourceGroup is the resource group key (a UUID). A JSON number still decodes.
	ResourceGroup DeviceResourceGroup `json:"resource_group,omitempty"`
	// UUID is the device UUID.
	UUID string `json:"uuid,omitempty"`
	// Enabled indicates whether the device is enabled.
	Enabled bool `json:"enabled"`
	// Optional allows the machine to start without this device if unavailable.
	Optional bool `json:"optional,omitempty"`
	// Asset is the asset tag (used for recipe/snapshot identification).
	Asset string `json:"asset,omitempty"`
	// Created is the creation timestamp (Unix epoch).
	Created int64 `json:"created,omitempty"`
	// Modified is the last modification timestamp (Unix epoch).
	Modified int64 `json:"modified,omitempty"`

	// Settings contains device-specific settings based on Type.
	// For USB devices: USBSettings
	// For TPM devices: TPMSettings
	// For vGPU devices: VGPUSettings
	USBSettings  *USBDeviceSettings  `json:"-"`
	TPMSettings  *TPMDeviceSettings  `json:"-"`
	VGPUSettings *VGPUDeviceSettings `json:"-"`
}

// USBDeviceSettings contains USB device-specific settings.
type USBDeviceSettings struct {
	// Key is the settings row key ($key).
	Key int `json:"$key,omitempty"`
	// MachineDevice is the parent device ID.
	MachineDevice int `json:"machine_device,omitempty"`
	// GuestReset indicates whether guest reset is allowed.
	GuestReset bool `json:"guest_reset,omitempty"`
	// GuestResetsAll indicates whether guest can reset all devices.
	GuestResetsAll bool `json:"guest_resets_all,omitempty"`
}

// TPMDeviceSettings contains TPM device-specific settings.
type TPMDeviceSettings struct {
	// Key is the settings row key ($key).
	Key int `json:"$key,omitempty"`
	// MachineDevice is the parent device ID.
	MachineDevice int `json:"machine_device,omitempty"`
	// Model is the TPM model.
	Model string `json:"model,omitempty"`
	// Version is the TPM version.
	Version string `json:"version,omitempty"`
}

// VGPUDeviceSettings contains NVIDIA vGPU device-specific settings.
type VGPUDeviceSettings struct {
	// Key is the settings row key ($key).
	Key int `json:"$key,omitempty"`
	// MachineDevice is the parent device ID.
	MachineDevice int `json:"machine_device,omitempty"`
	// ProfileType is the vGPU profile type.
	ProfileType string `json:"profile_type,omitempty"`
	// FrameRateLimiter is the frame rate limiter setting.
	FrameRateLimiter int `json:"frame_rate_limiter,omitempty"`
	// DisableVNC indicates whether VNC is disabled.
	DisableVNC bool `json:"disable_vnc,omitempty"`
	// EnableUVM indicates whether UVM is enabled.
	EnableUVM bool `json:"enable_uvm,omitempty"`
	// EnableDebugging indicates whether debugging is enabled.
	EnableDebugging bool `json:"enable_debugging,omitempty"`
	// EnableProfiling indicates whether profiling is enabled.
	EnableProfiling bool `json:"enable_profiling,omitempty"`
}

// VMDeviceCreateRequest is the request body for creating a device.
type VMDeviceCreateRequest struct {
	// Machine is the machine key (VM.Machine) written to machine_devices.
	// Create sets this from the VM $key argument; callers do not need to set it.
	Machine int `json:"machine"`
	// OrderID is the device order (0-64, auto-assigned if not specified).
	OrderID *int `json:"orderid,omitempty"`
	// Type is the device type. See the DeviceType* constants.
	Type string `json:"type"`
	// Name is the device name (auto-generated if blank).
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// ResourceGroup is the resource group key (a UUID). A decimal key is posted as a JSON number.
	ResourceGroup DeviceResourceGroup `json:"resource_group,omitempty"`
	// UUID is the device UUID (leave blank for system-generated).
	UUID string `json:"uuid,omitempty"`
	// Enabled indicates whether the device is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// Optional allows the machine to start without this device if unavailable.
	Optional *bool `json:"optional,omitempty"`
	// Asset is the asset tag.
	Asset string `json:"asset,omitempty"`
	// Count creates X copies of this device (1-16).
	Count *int `json:"count,omitempty"`

	// USB settings (only for USB devices)
	USBSettings *USBDeviceSettings `json:"-"`
	// TPM settings (only for TPM devices)
	TPMSettings *TPMDeviceSettings `json:"-"`
	// vGPU settings (only for vGPU devices)
	VGPUSettings *VGPUDeviceSettings `json:"-"`
}

// VMDeviceUpdateRequest is the request body for updating a device.
type VMDeviceUpdateRequest struct {
	// OrderID is the device order (0-64).
	OrderID *int `json:"orderid,omitempty"`
	// Name is the device name.
	Name *string `json:"name,omitempty"`
	// Description is the description.
	Description *string `json:"description,omitempty"`
	// ResourceGroup is the resource group key (a UUID). A decimal key is posted as a JSON number.
	ResourceGroup *DeviceResourceGroup `json:"resource_group,omitempty"`
	// UUID is the device UUID.
	UUID *string `json:"uuid,omitempty"`
	// Enabled indicates whether the device is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// Optional allows the machine to start without this device if unavailable.
	Optional *bool `json:"optional,omitempty"`
	// Asset is the asset tag.
	Asset *string `json:"asset,omitempty"`

	// USB settings (only for USB devices)
	USBSettings *USBDeviceSettings `json:"-"`
	// TPM settings (only for TPM devices)
	TPMSettings *TPMDeviceSettings `json:"-"`
	// vGPU settings (only for vGPU devices)
	VGPUSettings *VGPUDeviceSettings `json:"-"`
}

// Device type values stored on machine_devices.type.
// pyVergeOS creates the node_* values. gpu and nvidia_vgpu remain in the
// type list and are the older names.
const (
	DeviceTypeUSB        = "node_usb_devices"
	DeviceTypeTPM        = "tpm"
	DeviceTypePCI        = "node_pci_devices"
	DeviceTypeVGPU       = "node_nvidia_vgpu_devices"
	DeviceTypeHostGPU    = "node_host_gpu_devices"
	DeviceTypeSRIOVNIC   = "node_sriov_nic_devices"
	DeviceTypeGPULegacy  = "gpu"
	DeviceTypeVGPULegacy = "nvidia_vgpu"
)

// DeviceMachineTypeTenant is machine_devices.machine_type for a tenant node.
// The schema lists the value as tenant and the label as "Tenant Node".
const DeviceMachineTypeTenant = "tenant"

// deviceListFields are the fields to request when listing devices.
const deviceListFields = "$key,machine,machine_type,orderid,type,name,description,resource_group,uuid,enabled,optional,asset,created,modified"
