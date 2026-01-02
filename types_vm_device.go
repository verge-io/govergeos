package vergeos

// VMDevice represents a device attached to a VM.
type VMDevice struct {
	// ID is the unique identifier for the device.
	ID FlexInt `json:"$key,omitempty"`
	// Machine is the machine reference ID.
	Machine int `json:"machine,omitempty"`
	// Type is the device type (tpm, node_usb_devices, node_pci_devices, node_nvidia_vgpu_devices).
	Type string `json:"type"`
	// Name is the device name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// ResourceGroup is the resource group.
	ResourceGroup string `json:"resource_group,omitempty"`
	// Enabled indicates whether the device is enabled.
	Enabled bool `json:"enabled"`
	// Status is the device status.
	Status int `json:"status,omitempty"`

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
	// ID is the settings ID.
	ID int `json:"$key,omitempty"`
	// MachineDevice is the parent device ID.
	MachineDevice int `json:"machine_device,omitempty"`
	// GuestReset indicates whether guest reset is allowed.
	GuestReset bool `json:"guest_reset,omitempty"`
	// GuestResetsAll indicates whether guest can reset all devices.
	GuestResetsAll bool `json:"guest_resets_all,omitempty"`
}

// TPMDeviceSettings contains TPM device-specific settings.
type TPMDeviceSettings struct {
	// ID is the settings ID.
	ID int `json:"$key,omitempty"`
	// MachineDevice is the parent device ID.
	MachineDevice int `json:"machine_device,omitempty"`
	// Model is the TPM model.
	Model string `json:"model,omitempty"`
	// Version is the TPM version.
	Version string `json:"version,omitempty"`
}

// VGPUDeviceSettings contains NVIDIA vGPU device-specific settings.
type VGPUDeviceSettings struct {
	// ID is the settings ID.
	ID int `json:"$key,omitempty"`
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
	// Machine is the VM's machine ID.
	Machine int `json:"machine"`
	// Type is the device type (tpm, node_usb_devices, node_pci_devices, node_nvidia_vgpu_devices).
	Type string `json:"type"`
	// Name is the device name.
	Name string `json:"name"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// ResourceGroup is the resource group.
	ResourceGroup string `json:"resource_group,omitempty"`
	// Enabled indicates whether the device is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// USB settings (only for USB devices)
	USBSettings *USBDeviceSettings `json:"-"`
	// TPM settings (only for TPM devices)
	TPMSettings *TPMDeviceSettings `json:"-"`
	// vGPU settings (only for vGPU devices)
	VGPUSettings *VGPUDeviceSettings `json:"-"`
}

// VMDeviceUpdateRequest is the request body for updating a device.
type VMDeviceUpdateRequest struct {
	// Name is the device name.
	Name *string `json:"name,omitempty"`
	// Description is the description.
	Description *string `json:"description,omitempty"`
	// ResourceGroup is the resource group.
	ResourceGroup *string `json:"resource_group,omitempty"`
	// Enabled indicates whether the device is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// USB settings (only for USB devices)
	USBSettings *USBDeviceSettings `json:"-"`
	// TPM settings (only for TPM devices)
	TPMSettings *TPMDeviceSettings `json:"-"`
	// vGPU settings (only for vGPU devices)
	VGPUSettings *VGPUDeviceSettings `json:"-"`
}

// Device type constants
const (
	DeviceTypeUSB  = "node_usb_devices"
	DeviceTypeTPM  = "tpm"
	DeviceTypePCI  = "node_pci_devices"
	DeviceTypeVGPU = "node_nvidia_vgpu_devices"
)

// deviceListFields are the fields to request when listing devices.
const deviceListFields = "$key,machine,type,name,description,enabled,resource_group,status"
