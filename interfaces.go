package vergeos

import "context"

// VMServiceInterface defines the interface for VM operations.
// This interface enables mocking for testing purposes.
type VMServiceInterface interface {
	// List returns all VMs, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]VM, error)
	// Get returns a single VM by ID.
	Get(ctx context.Context, id int) (*VM, error)
	// Create creates a new VM and returns the created VM.
	Create(ctx context.Context, req *VMCreateRequest) (*VM, error)
	// Update updates a VM and returns the updated VM.
	Update(ctx context.Context, id int, req *VMUpdateRequest) (*VM, error)
	// Delete deletes a VM.
	Delete(ctx context.Context, id int) error
	// PowerOn powers on a VM and waits for it to start.
	PowerOn(ctx context.Context, id int) error
	// PowerOff powers off a VM and waits for it to stop.
	PowerOff(ctx context.Context, id int) error
	// Reset sends a reset signal to a running VM.
	Reset(ctx context.Context, id int) error
	// GuestReboot sends a reboot request to the guest OS via ACPI.
	GuestReboot(ctx context.Context, id int) error
	// GuestShutdown sends a graceful shutdown request to the guest OS via ACPI.
	GuestShutdown(ctx context.Context, id int) error
	// Clone creates a copy of a VM.
	Clone(ctx context.Context, id int, opts *VMCloneOptions) error
	// Snapshot takes a snapshot of a VM.
	Snapshot(ctx context.Context, id int, opts *VMSnapshotOptions) error
}

// VMNICServiceInterface defines the interface for VM NIC operations.
type VMNICServiceInterface interface {
	// List returns all NICs for a VM.
	List(ctx context.Context, vmID int) ([]VMNIC, error)
	// Get returns a single NIC by ID.
	Get(ctx context.Context, nicID int) (*VMNIC, error)
	// Create creates a new NIC and returns the created NIC.
	Create(ctx context.Context, vmID int, req *VMNICCreateRequest) (*VMNIC, error)
	// Update updates a NIC and returns the updated NIC.
	Update(ctx context.Context, nicID int, req *VMNICUpdateRequest) (*VMNIC, error)
	// Delete deletes a NIC.
	Delete(ctx context.Context, nicID int) error
}

// VMDriveServiceInterface defines the interface for VM drive operations.
type VMDriveServiceInterface interface {
	// List returns all drives for a VM.
	List(ctx context.Context, vmID int) ([]VMDrive, error)
	// Get returns a single drive by ID.
	Get(ctx context.Context, driveID int) (*VMDrive, error)
	// Create creates a new drive and returns the created drive.
	Create(ctx context.Context, vmID int, req *VMDriveCreateRequest) (*VMDrive, error)
	// Update updates a drive and returns the updated drive.
	Update(ctx context.Context, driveID int, req *VMDriveUpdateRequest) (*VMDrive, error)
	// Delete deletes a drive.
	Delete(ctx context.Context, driveID int) error
}

// VMDeviceServiceInterface defines the interface for VM device operations.
type VMDeviceServiceInterface interface {
	// List returns all devices for a VM.
	List(ctx context.Context, vmID int) ([]VMDevice, error)
	// Get returns a single device by ID.
	Get(ctx context.Context, deviceID int) (*VMDevice, error)
	// Create creates a new device and returns the created device.
	Create(ctx context.Context, vmID int, req *VMDeviceCreateRequest) (*VMDevice, error)
	// Update updates a device and returns the updated device.
	Update(ctx context.Context, deviceID int, req *VMDeviceUpdateRequest) (*VMDevice, error)
	// Delete deletes a device.
	Delete(ctx context.Context, deviceID int) error
}

// NetworkServiceInterface defines the interface for network operations.
type NetworkServiceInterface interface {
	// List returns all networks, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]Network, error)
	// Get returns a single network by ID.
	Get(ctx context.Context, id int) (*Network, error)
	// Create creates a new network and returns the created network.
	Create(ctx context.Context, req *NetworkCreateRequest) (*Network, error)
	// Update updates a network and returns the updated network.
	Update(ctx context.Context, id int, req *NetworkUpdateRequest) (*Network, error)
	// Delete deletes a network.
	Delete(ctx context.Context, id int) error
	// PowerOn powers on a network and waits for it to start.
	PowerOn(ctx context.Context, id int) error
	// PowerOff powers off a network and waits for it to stop.
	PowerOff(ctx context.Context, id int) error
}

// UserServiceInterface defines the interface for user operations.
type UserServiceInterface interface {
	// List returns all users, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]User, error)
	// Get returns a single user by ID.
	Get(ctx context.Context, id int) (*User, error)
	// GetByName returns a user by username.
	GetByName(ctx context.Context, name string) (*User, error)
	// Create creates a new user and returns the created user.
	Create(ctx context.Context, req *UserCreateRequest) (*User, error)
	// Update updates a user and returns the updated user.
	Update(ctx context.Context, id int, req *UserUpdateRequest) (*User, error)
	// Delete deletes a user.
	Delete(ctx context.Context, id int) error
}

// MemberServiceInterface defines the interface for group membership operations.
type MemberServiceInterface interface {
	// List returns all members, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]Member, error)
	// ListByGroup returns all members of a specific group.
	ListByGroup(ctx context.Context, groupID int) ([]Member, error)
	// Get returns a single member by ID.
	Get(ctx context.Context, id int) (*Member, error)
	// Create creates a new membership and returns the created member.
	Create(ctx context.Context, req *MemberCreateRequest) (*Member, error)
	// Update updates a membership and returns the updated member.
	Update(ctx context.Context, id int, req *MemberUpdateRequest) (*Member, error)
	// Delete deletes a membership.
	Delete(ctx context.Context, id int) error
	// Add is a convenience method to add a member to a group.
	Add(ctx context.Context, groupID int, member string) (*Member, error)
	// Remove is a convenience method to remove a member from a group.
	Remove(ctx context.Context, groupID int, member string) error
}

// CloudInitServiceInterface defines the interface for cloud-init file operations.
type CloudInitServiceInterface interface {
	// List returns all cloud-init files, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]CloudInitFile, error)
	// Get returns a single cloud-init file by ID.
	Get(ctx context.Context, id int) (*CloudInitFile, error)
	// GetByName returns a cloud-init file by name.
	GetByName(ctx context.Context, name string) (*CloudInitFile, error)
	// Create creates a new cloud-init file and returns the created file.
	Create(ctx context.Context, req *CloudInitFileCreateRequest) (*CloudInitFile, error)
	// Update updates a cloud-init file and returns the updated file.
	Update(ctx context.Context, id int, req *CloudInitFileUpdateRequest) (*CloudInitFile, error)
	// Delete deletes a cloud-init file.
	Delete(ctx context.Context, id int) error
}

// ClusterServiceInterface defines the interface for cluster read operations.
type ClusterServiceInterface interface {
	// List returns all clusters, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]Cluster, error)
	// Get returns a single cluster by ID.
	Get(ctx context.Context, id int) (*Cluster, error)
	// GetByName returns a cluster by name.
	GetByName(ctx context.Context, name string) (*Cluster, error)
	// GetStatus returns detailed status for a cluster.
	GetStatus(ctx context.Context, id int) (*ClusterStatus, error)
}

// NodeServiceInterface defines the interface for node read operations.
type NodeServiceInterface interface {
	// List returns all nodes, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]Node, error)
	// ListPhysical returns all physical nodes.
	ListPhysical(ctx context.Context, opts ...ListOption) ([]Node, error)
	// Get returns a single node by ID.
	Get(ctx context.Context, id int) (*Node, error)
	// GetByName returns a node by name.
	GetByName(ctx context.Context, name string) (*Node, error)
	// GetDashboard returns detailed dashboard information for a node.
	GetDashboard(ctx context.Context, id int) (*Node, error)
}

// GroupServiceInterface defines the interface for group read operations.
type GroupServiceInterface interface {
	// List returns all groups, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]Group, error)
	// Get returns a single group by ID.
	Get(ctx context.Context, id int) (*Group, error)
	// GetByName returns a group by name.
	GetByName(ctx context.Context, name string) (*Group, error)
}

// MediaSourceServiceInterface defines the interface for media source read operations.
type MediaSourceServiceInterface interface {
	// List returns all media sources, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]MediaSource, error)
	// Get returns a single media source by ID.
	Get(ctx context.Context, id int) (*MediaSource, error)
	// GetByName returns a media source by name.
	GetByName(ctx context.Context, name string) (*MediaSource, error)
	// ListISOs returns all ISO media sources.
	ListISOs(ctx context.Context, opts ...ListOption) ([]MediaSource, error)
}

// ResourceGroupServiceInterface defines the interface for resource group read operations.
type ResourceGroupServiceInterface interface {
	// List returns all resource groups, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]ResourceGroup, error)
	// Get returns a single resource group by ID.
	Get(ctx context.Context, id int) (*ResourceGroup, error)
	// GetByName returns a resource group by name.
	GetByName(ctx context.Context, name string) (*ResourceGroup, error)
}

// SettingsServiceInterface defines the interface for system settings read operations.
type SettingsServiceInterface interface {
	// List returns all settings, with optional filtering and pagination.
	List(ctx context.Context, opts ...ListOption) ([]Setting, error)
	// Get returns a single setting by ID.
	Get(ctx context.Context, id int) (*Setting, error)
	// GetByKey returns a setting by key name.
	GetByKey(ctx context.Context, key string) (*Setting, error)
	// GetValue returns the value of a setting by key name.
	GetValue(ctx context.Context, key string) (string, error)
	// GetCloudName returns the cloud_name setting value.
	GetCloudName(ctx context.Context) (string, error)
}

// SystemServiceInterface defines the interface for system information operations.
type SystemServiceInterface interface {
	// GetInfo returns system version information.
	GetInfo(ctx context.Context) (*SystemInfo, error)
	// GetVersion returns the VergeOS version string.
	GetVersion(ctx context.Context) (string, error)
}

// SchemaServiceInterface defines the interface for schema discovery operations.
type SchemaServiceInterface interface {
	// GetTableSchema returns the schema for a resource type.
	GetTableSchema(ctx context.Context, resource string) (*TableSchema, error)
	// GetValidValues returns the valid values for a field in a resource type.
	GetValidValues(ctx context.Context, resource, field string) (map[string]string, error)
	// GetVMMachineTypes returns valid machine types for VMs.
	GetVMMachineTypes(ctx context.Context) (map[string]string, error)
	// GetVMOSFamilies returns valid OS families for VMs.
	GetVMOSFamilies(ctx context.Context) (map[string]string, error)
}

// Compile-time interface implementation checks.
var (
	_ VMServiceInterface            = (*VMService)(nil)
	_ VMNICServiceInterface         = (*VMNICService)(nil)
	_ VMDriveServiceInterface       = (*VMDriveService)(nil)
	_ VMDeviceServiceInterface      = (*VMDeviceService)(nil)
	_ NetworkServiceInterface       = (*NetworkService)(nil)
	_ UserServiceInterface          = (*UserService)(nil)
	_ MemberServiceInterface        = (*MemberService)(nil)
	_ CloudInitServiceInterface     = (*CloudInitService)(nil)
	_ ClusterServiceInterface       = (*ClusterService)(nil)
	_ NodeServiceInterface          = (*NodeService)(nil)
	_ GroupServiceInterface         = (*GroupService)(nil)
	_ MediaSourceServiceInterface   = (*MediaSourceService)(nil)
	_ ResourceGroupServiceInterface = (*ResourceGroupService)(nil)
	_ SettingsServiceInterface      = (*SettingsService)(nil)
	_ SystemServiceInterface        = (*SystemService)(nil)
	_ SchemaServiceInterface        = (*SchemaService)(nil)
)
