package vergeos

// GPU operating modes. These are the mode values stored on node_gpus.
const (
	// GPUModeNone leaves the GPU unassigned.
	GPUModeNone = "none"
	// GPUModePassthrough assigns the whole GPU to one VM.
	GPUModePassthrough = "gpu"
	// GPUModeNVIDIAvGPU shares the GPU as NVIDIA vGPU instances.
	GPUModeNVIDIAvGPU = "nvidia_vgpu"
)

// NVIDIA vGPU profile type codes.
const (
	// VGPUProfileTypeApps is profile type A, virtual applications.
	VGPUProfileTypeApps = "A"
	// VGPUProfileTypeDesktops is profile type B, virtual desktops.
	VGPUProfileTypeDesktops = "B"
	// VGPUProfileTypeCompute is profile type C, AI, machine learning, and compute.
	VGPUProfileTypeCompute = "C"
	// VGPUProfileTypeWorkstations is profile type Q, virtual workstations.
	VGPUProfileTypeWorkstations = "Q"
)

// VGPUProfileTypeDisplay returns the description pyVergeOS uses for a profile type code.
// An unrecognized code is returned unchanged.
func VGPUProfileTypeDisplay(code string) string {
	switch code {
	case VGPUProfileTypeApps:
		return "Virtual Applications (vApps)"
	case VGPUProfileTypeDesktops:
		return "Virtual Desktops (vPC)"
	case VGPUProfileTypeCompute:
		return "AI/Machine Learning/Training (vCS or vWS)"
	case VGPUProfileTypeWorkstations:
		return "Virtual Workstations (vWS)"
	default:
		return code
	}
}

func validGPUMode(mode string) bool {
	switch mode {
	case GPUModeNone, GPUModePassthrough, GPUModeNVIDIAvGPU:
		return true
	default:
		return false
	}
}

func validVGPUProfileType(code string) bool {
	switch code {
	case VGPUProfileTypeApps, VGPUProfileTypeDesktops, VGPUProfileTypeCompute, VGPUProfileTypeWorkstations:
		return true
	default:
		return false
	}
}

// VGPUProfile is an NVIDIA vGPU profile from nvidia_vgpu_profiles.
//
// Profiles are read-only and come from the NVIDIA driver. A VM vGPU device
// selects one of these profiles.
type VGPUProfile struct {
	// Key is the profile row key.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the profile name (for example "nvidia-256").
	Name string `json:"name,omitempty"`
	// TypeID is the NVIDIA type id.
	TypeID int `json:"type_id,omitempty"`
	// DeviceHex is the vendor:device id (for example "10de:1eb8").
	DeviceHex string `json:"device_hex,omitempty"`
	// NumHeads is the number of display heads.
	NumHeads int `json:"num_heads,omitempty"`
	// FRLConfig is the frame rate limiter configuration.
	FRLConfig int `json:"frl_config,omitempty"`
	// Framebuffer is the framebuffer size (for example "256M").
	Framebuffer string `json:"framebuffer,omitempty"`
	// MaxResolution is the maximum resolution (for example "4096x2160").
	MaxResolution string `json:"max_resolution,omitempty"`
	// MaxInstance is the maximum instances per physical GPU.
	MaxInstance int `json:"max_instance,omitempty"`
	// MaxInstancesPerVM is the maximum instances of this profile on one VM.
	MaxInstancesPerVM int `json:"max_instances_per_vm,omitempty"`
	// PlacementIDs lists placement ids for the profile.
	PlacementIDs string `json:"placement_ids,omitempty"`
	// Location is the profile path on the host.
	Location string `json:"location,omitempty"`
	// ProfileType is the type code: A, B, C, or Q.
	ProfileType string `json:"profile_type,omitempty"`
	// GridLicense is the GRID license the profile requires.
	GridLicense string `json:"grid_license,omitempty"`
	// VirtualFunction reports whether this is a virtual function profile.
	VirtualFunction bool `json:"virtual_function,omitempty"`
	// ProfileFolder is the profile folder name.
	ProfileFolder string `json:"profile_folder,omitempty"`
}

// ProfileTypeDisplay returns the description of ProfileType.
func (p VGPUProfile) ProfileTypeDisplay() string {
	return VGPUProfileTypeDisplay(p.ProfileType)
}

// NodeGPU is a physical GPU configured on a node (node_gpus).
//
// Mode is GPUModeNone, GPUModePassthrough, or GPUModeNVIDIAvGPU.
// NVIDIAVgpuProfile is set when Mode is GPUModeNVIDIAvGPU.
type NodeGPU struct {
	// Key is the GPU row key.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the GPU name (for example "GPU_1").
	Name string `json:"name,omitempty"`
	// Description is the GPU description.
	Description string `json:"description,omitempty"`
	// PCIDevice is the associated PCI device key. Zero means unset.
	PCIDevice int `json:"pci_device,omitempty"`
	// PCIDeviceName is the PCI device name (pci_device#name).
	PCIDeviceName string `json:"pci_device_name,omitempty"`
	// Node is the parent node key.
	Node int `json:"node,omitempty"`
	// NodeName is the parent node name (node#name).
	NodeName string `json:"node_name,omitempty"`
	// Mode is the operating mode.
	Mode string `json:"mode,omitempty"`
	// NVIDIAVgpuProfile is the assigned vGPU profile key. Zero means unset.
	NVIDIAVgpuProfile int `json:"nvidia_vgpu_profile,omitempty"`
	// NVIDIAVgpuProfileDisplay is the display name of the assigned profile.
	NVIDIAVgpuProfileDisplay string `json:"nvidia_vgpu_profile_disp,omitempty"`
	// MaxInstances is how many GPU or vGPU instances this GPU can provide.
	MaxInstances int `json:"max_instances,omitempty"`
	// InstancesCount is the number of instances currently assigned.
	InstancesCount int `json:"instances_count,omitempty"`
	// Modified is the last modification time as a Unix timestamp.
	Modified int64 `json:"modified,omitempty"`
}

// ModeDisplay returns the description pyVergeOS uses for Mode.
// An unrecognized mode is returned unchanged.
func (g NodeGPU) ModeDisplay() string {
	switch g.Mode {
	case GPUModeNone:
		return "None"
	case GPUModePassthrough:
		return "PCI Passthrough"
	case GPUModeNVIDIAvGPU:
		return "NVIDIA vGPU"
	default:
		return g.Mode
	}
}

// IsPassthrough reports whether the GPU is assigned as PCI passthrough.
func (g NodeGPU) IsPassthrough() bool {
	return g.Mode == GPUModePassthrough
}

// IsVGPU reports whether the GPU is shared as NVIDIA vGPU.
func (g NodeGPU) IsVGPU() bool {
	return g.Mode == GPUModeNVIDIAvGPU
}

// IsDisabled reports whether the GPU has no mode assigned.
func (g NodeGPU) IsDisabled() bool {
	return g.Mode == GPUModeNone
}

// NodeGPUUpdateRequest changes a node GPU.
// Nil fields are left unchanged.
type NodeGPUUpdateRequest struct {
	// Name is the GPU name.
	Name *string `json:"name,omitempty"`
	// Description is the GPU description.
	Description *string `json:"description,omitempty"`
	// Mode is GPUModeNone, GPUModePassthrough, or GPUModeNVIDIAvGPU.
	Mode *string `json:"mode,omitempty"`
	// NVIDIAVgpuProfile is the vGPU profile key used with GPUModeNVIDIAvGPU.
	NVIDIAVgpuProfile *int `json:"nvidia_vgpu_profile,omitempty"`
}

// NodeGPUStats is a current or historical utilization row for one node GPU.
//
// Current rows come from node_gpu_stats. History rows come from
// node_gpu_stats_history_short and node_gpu_stats_history_long.
type NodeGPUStats struct {
	// Key is the stats row key.
	Key FlexInt `json:"$key,omitempty"`
	// NodeGPU is the parent GPU key.
	NodeGPU int `json:"node_gpu,omitempty"`
	// GPUsTotal is the number of GPU slots.
	GPUsTotal int `json:"gpus_total,omitempty"`
	// GPUs is the number of GPU slots in use.
	GPUs int `json:"gpus,omitempty"`
	// GPUsIdle is the number of idle GPU slots.
	GPUsIdle int `json:"gpus_idle,omitempty"`
	// VGPUsTotal is the number of vGPU slots.
	VGPUsTotal int `json:"vgpus_total,omitempty"`
	// VGPUs is the number of vGPU slots in use.
	VGPUs int `json:"vgpus,omitempty"`
	// VGPUsIdle is the number of idle vGPU slots.
	VGPUsIdle int `json:"vgpus_idle,omitempty"`
	// Timestamp is the sample time as a Unix timestamp.
	Timestamp int64 `json:"timestamp,omitempty"`
}

// NodeGPUInstance is one GPU or vGPU instance assigned to a VM (node_gpu_instances).
type NodeGPUInstance struct {
	// Key is the instance row key.
	Key FlexInt `json:"$key,omitempty"`
	// Description is the instance description.
	Description string `json:"description,omitempty"`
	// Modified is the last modification time as a Unix timestamp.
	Modified int64 `json:"modified,omitempty"`
	// GPUKey is the parent GPU key (gpu#$key).
	GPUKey int `json:"gpu_key,omitempty"`
	// GPUName is the parent GPU name (gpu#name).
	GPUName string `json:"gpu_name,omitempty"`
	// NodeKey is the node key (gpu#node#$key).
	NodeKey int `json:"node_key,omitempty"`
	// NodeDisplay is the node display name (gpu#node#$display).
	NodeDisplay string `json:"node_display,omitempty"`
	// MachineKey is the VM machine key.
	MachineKey int `json:"machine_key,omitempty"`
	// MachineName is the VM name.
	MachineName string `json:"machine_name,omitempty"`
	// MachineType is the machine type (for example "vm").
	MachineType string `json:"machine_type,omitempty"`
	// MachineTypeDisplay is the machine type display name.
	MachineTypeDisplay string `json:"machine_type_display,omitempty"`
	// MachineDeviceKey is the VM device key.
	MachineDeviceKey int `json:"machine_device_key,omitempty"`
	// MachineDeviceName is the VM device name.
	MachineDeviceName string `json:"machine_device_name,omitempty"`
	// MachineDeviceStatus is the VM device status.
	MachineDeviceStatus string `json:"machine_device_status,omitempty"`
	// PCIDeviceKey is the PCI device key.
	PCIDeviceKey int `json:"pci_device_key,omitempty"`
	// PCIDeviceName is the PCI device name.
	PCIDeviceName string `json:"pci_device_name,omitempty"`
	// Mode is the parent GPU mode.
	Mode string `json:"mode,omitempty"`
	// ModeDisplay is the parent GPU mode display name from the API.
	ModeDisplay string `json:"mode_display,omitempty"`
}

// NodeVGPUDevice is a detected NVIDIA vGPU-capable device (node_nvidia_vgpu_devices).
//
// The platform creates these rows. They are read-only.
type NodeVGPUDevice struct {
	// Key is the device row key.
	Key FlexInt `json:"$key,omitempty"`
	// Node is the parent node key.
	Node int `json:"node,omitempty"`
	// NodeName is the parent node name (node#name).
	NodeName string `json:"node_name,omitempty"`
	// PCIDevice is the associated PCI device key. Zero means unset.
	PCIDevice int `json:"pci_device,omitempty"`
	// Name is the device name.
	Name string `json:"name,omitempty"`
	// Slot is the PCI slot (for example "0000:41:00.0").
	Slot string `json:"slot,omitempty"`
	// Vendor is the vendor name.
	Vendor string `json:"vendor,omitempty"`
	// Device is the device description.
	Device string `json:"device,omitempty"`
	// VendorDeviceHex is the vendor:device id.
	VendorDeviceHex string `json:"vendor_device_hex,omitempty"`
	// Driver is the current driver.
	Driver string `json:"driver,omitempty"`
	// Module is the kernel module.
	Module string `json:"module,omitempty"`
	// NUMA is the NUMA node.
	NUMA string `json:"numa,omitempty"`
	// IOMMUGroup is the IOMMU group.
	IOMMUGroup string `json:"iommu_group,omitempty"`
	// TypeID is the device type id.
	TypeID int `json:"type_id,omitempty"`
	// MaxInstances is the maximum vGPU instances.
	MaxInstances int `json:"max_instances,omitempty"`
	// PhysicalFunction is the SR-IOV physical function.
	PhysicalFunction string `json:"physical_function,omitempty"`
	// VirtFn is the virtual function identifier.
	VirtFn string `json:"virtfn,omitempty"`
	// Fingerprint identifies the device for live migration.
	Fingerprint string `json:"fingerprint,omitempty"`
	// Created is when the device was detected, as a Unix timestamp.
	Created int64 `json:"created,omitempty"`
	// Modified is the last update time as a Unix timestamp.
	Modified int64 `json:"modified,omitempty"`
}

// NodeHostGPUDevice is a detected GPU available for passthrough (node_host_gpu_devices).
//
// The platform creates these rows. They are read-only.
type NodeHostGPUDevice struct {
	// Key is the device row key.
	Key FlexInt `json:"$key,omitempty"`
	// Node is the parent node key.
	Node int `json:"node,omitempty"`
	// NodeName is the parent node name (node#name).
	NodeName string `json:"node_name,omitempty"`
	// PCIDevice is the associated PCI device key. Zero means unset.
	PCIDevice int `json:"pci_device,omitempty"`
	// Name is the device name.
	Name string `json:"name,omitempty"`
	// Slot is the PCI slot.
	Slot string `json:"slot,omitempty"`
	// Vendor is the vendor name.
	Vendor string `json:"vendor,omitempty"`
	// Device is the device description.
	Device string `json:"device,omitempty"`
	// VendorDeviceHex is the vendor:device id.
	VendorDeviceHex string `json:"vendor_device_hex,omitempty"`
	// Driver is the current driver.
	Driver string `json:"driver,omitempty"`
	// Module is the kernel module.
	Module string `json:"module,omitempty"`
	// NUMA is the NUMA node.
	NUMA string `json:"numa,omitempty"`
	// IOMMUGroup is the IOMMU group.
	IOMMUGroup string `json:"iommu_group,omitempty"`
	// TypeID is the device type id.
	TypeID int `json:"type_id,omitempty"`
	// DeviceIndex is the device index.
	DeviceIndex int `json:"device_index,omitempty"`
	// MaxInstances is the maximum instances. Passthrough is typically 1.
	MaxInstances int `json:"max_instances,omitempty"`
	// Fingerprint identifies the device for live migration.
	Fingerprint string `json:"fingerprint,omitempty"`
	// Created is when the device was detected, as a Unix timestamp.
	Created int64 `json:"created,omitempty"`
	// Modified is the last update time as a Unix timestamp.
	Modified int64 `json:"modified,omitempty"`
}

// NodeVGPUProfile is a vGPU profile available on one physical GPU (node_nvidia_vgpu_profiles).
//
// These rows are read-only and follow the hardware and the NVIDIA driver.
// VGPUProfile is the system-wide catalog. This type is the per-GPU view,
// including how many instances are still free.
type NodeVGPUProfile struct {
	// Key is the profile row key.
	Key FlexInt `json:"$key,omitempty"`
	// PhysicalGPU is the physical GPU (PCI device) key.
	PhysicalGPU int `json:"physical_gpu,omitempty"`
	// Name is the profile name (for example "nvidia-256").
	Name string `json:"name,omitempty"`
	// NumHeads is the number of display heads.
	NumHeads int `json:"num_heads,omitempty"`
	// FRLConfig is the frame rate limiter configuration.
	FRLConfig int `json:"frl_config,omitempty"`
	// Framebuffer is the framebuffer size.
	Framebuffer string `json:"framebuffer,omitempty"`
	// MaxResolution is the maximum resolution.
	MaxResolution string `json:"max_resolution,omitempty"`
	// MaxInstance is the maximum instances of this profile on the GPU.
	MaxInstance int `json:"max_instance,omitempty"`
	// AvailableInstances is how many instances are still free.
	AvailableInstances int `json:"available_instances,omitempty"`
	// DeviceAPI is the device API version.
	DeviceAPI string `json:"device_api,omitempty"`
	// ProfileType is the type code: A, B, C, or Q.
	ProfileType string `json:"profile_type,omitempty"`
	// VirtualFunction reports whether this is a virtual function profile.
	VirtualFunction bool `json:"virtual_function,omitempty"`
	// ProfileFolder is the profile folder name.
	ProfileFolder string `json:"profile_folder,omitempty"`
}

// ProfileTypeDisplay returns the description of ProfileType.
func (p NodeVGPUProfile) ProfileTypeDisplay() string {
	return VGPUProfileTypeDisplay(p.ProfileType)
}

const (
	vgpuProfileListFields = "$key,name,type_id,device_hex,num_heads,frl_config,framebuffer,max_resolution,max_instance,max_instances_per_vm,placement_ids,location,profile_type,grid_license,virtual_function,profile_folder"

	nodeGPUListFields = "$key,name,description,pci_device,node,nvidia_vgpu_profile,max_instances,modified,pci_device#name as pci_device_name,node#name as node_name,mode,display(nvidia_vgpu_profile) as nvidia_vgpu_profile_disp,count(instances) as instances_count"

	nodeGPUStatsListFields = "$key,node_gpu,gpus_total,gpus,gpus_idle,vgpus_total,vgpus,vgpus_idle,timestamp"

	nodeGPUInstanceListFields = "$key,description,modified,gpu#$key as gpu_key,gpu#name as gpu_name,gpu#node#$key as node_key,gpu#node#$display as node_display,machine_device#machine#$key as machine_key,machine_device#machine#name as machine_name,machine_device#machine#type as machine_type,machine_device#machine#display(type) as machine_type_display,machine_device#$key as machine_device_key,machine_device#name as machine_device_name,machine_device#status#status as machine_device_status,gpu#pci_device#$key as pci_device_key,gpu#pci_device#name as pci_device_name,gpu#mode as mode,gpu#display(mode) as mode_display"

	nodeVGPUDeviceListFields = "$key,node,pci_device,name,slot,vendor,device,vendor_device_hex,driver,module,numa,iommu_group,type_id,max_instances,physical_function,virtfn,fingerprint,created,modified,node#name as node_name"

	nodeHostGPUDeviceListFields = "$key,node,pci_device,name,slot,vendor,device,vendor_device_hex,driver,module,numa,iommu_group,type_id,device_index,max_instances,fingerprint,created,modified,node#name as node_name"

	nodeVGPUProfileListFields = "$key,physical_gpu,name,num_heads,frl_config,framebuffer,max_resolution,max_instance,available_instances,device_api,profile_type,virtual_function,profile_folder"
)
