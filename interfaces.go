package vergeos

import "context"

// This file defines interfaces for all services to enable mock testing and dependency injection.
// See ADR-012 in DECISIONS.md for design rationale.

// VMServiceInterface defines the interface for VM operations.
type VMServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]VM, error)
	Get(ctx context.Context, id int) (*VM, error)
	Create(ctx context.Context, req *VMCreateRequest) (*VM, error)
	Update(ctx context.Context, id int, req *VMUpdateRequest) (*VM, error)
	Delete(ctx context.Context, id int) error
	PowerOn(ctx context.Context, id int) error
	PowerOff(ctx context.Context, id int) error
	Reset(ctx context.Context, id int) error
	GuestReboot(ctx context.Context, id int) error
	GuestShutdown(ctx context.Context, id int) error
	Clone(ctx context.Context, id int, opts *VMCloneOptions) error
	Snapshot(ctx context.Context, id int, opts *VMSnapshotOptions) error
}

// VMNICServiceInterface defines the interface for VM NIC operations.
type VMNICServiceInterface interface {
	List(ctx context.Context, vmID int) ([]VMNIC, error)
	Get(ctx context.Context, nicID int) (*VMNIC, error)
	Create(ctx context.Context, vmID int, req *VMNICCreateRequest) (*VMNIC, error)
	Update(ctx context.Context, nicID int, req *VMNICUpdateRequest) (*VMNIC, error)
	Delete(ctx context.Context, nicID int) error
}

// VMDriveServiceInterface defines the interface for VM Drive operations.
type VMDriveServiceInterface interface {
	List(ctx context.Context, vmID int) ([]VMDrive, error)
	Get(ctx context.Context, driveID int) (*VMDrive, error)
	Create(ctx context.Context, vmID int, req *VMDriveCreateRequest) (*VMDrive, error)
	Update(ctx context.Context, driveID int, req *VMDriveUpdateRequest) (*VMDrive, error)
	Delete(ctx context.Context, driveID int) error
}

// VMDeviceServiceInterface defines the interface for VM Device operations.
type VMDeviceServiceInterface interface {
	List(ctx context.Context, vmID int) ([]VMDevice, error)
	Get(ctx context.Context, deviceID int) (*VMDevice, error)
	Create(ctx context.Context, vmID int, req *VMDeviceCreateRequest) (*VMDevice, error)
	Update(ctx context.Context, deviceID int, req *VMDeviceUpdateRequest) (*VMDevice, error)
	Delete(ctx context.Context, deviceID int) error
}

// NetworkServiceInterface defines the interface for Network operations.
type NetworkServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Network, error)
	Get(ctx context.Context, id int) (*Network, error)
	Create(ctx context.Context, req *NetworkCreateRequest) (*Network, error)
	Update(ctx context.Context, id int, req *NetworkUpdateRequest) (*Network, error)
	Delete(ctx context.Context, id int) error
	PowerOn(ctx context.Context, id int) error
	PowerOff(ctx context.Context, id int) error
	Kill(ctx context.Context, id int) error
	Reset(ctx context.Context, id int, applyFirewall bool) error
	ApplyRules(ctx context.Context, id int) error
	ApplyDNS(ctx context.Context, id int) error
}

// UserServiceInterface defines the interface for User operations.
type UserServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]User, error)
	Get(ctx context.Context, id int) (*User, error)
	GetByName(ctx context.Context, name string) (*User, error)
	Create(ctx context.Context, req *UserCreateRequest) (*User, error)
	Update(ctx context.Context, id int, req *UserUpdateRequest) (*User, error)
	Delete(ctx context.Context, id int) error
	Enable(ctx context.Context, id int) error
	Disable(ctx context.Context, id int) error
}

// MemberServiceInterface defines the interface for Member operations.
type MemberServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Member, error)
	ListByGroup(ctx context.Context, groupID int) ([]Member, error)
	Get(ctx context.Context, id int) (*Member, error)
	Create(ctx context.Context, req *MemberCreateRequest) (*Member, error)
	Update(ctx context.Context, id int, req *MemberUpdateRequest) (*Member, error)
	Delete(ctx context.Context, id int) error
	Add(ctx context.Context, groupID int, member string) (*Member, error)
	Remove(ctx context.Context, groupID int, member string) error
}

// CloudInitServiceInterface defines the interface for CloudInit operations.
type CloudInitServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]CloudInitFile, error)
	Get(ctx context.Context, id int) (*CloudInitFile, error)
	GetByName(ctx context.Context, name string) (*CloudInitFile, error)
	Create(ctx context.Context, req *CloudInitFileCreateRequest) (*CloudInitFile, error)
	Update(ctx context.Context, id int, req *CloudInitFileUpdateRequest) (*CloudInitFile, error)
	Delete(ctx context.Context, id int) error
}

// NodeServiceInterface defines the interface for Node operations.
type NodeServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Node, error)
	ListPhysical(ctx context.Context, opts ...ListOption) ([]Node, error)
	Get(ctx context.Context, id int) (*Node, error)
	GetByName(ctx context.Context, name string) (*Node, error)
	GetDashboard(ctx context.Context, id int) (*Node, error)
	EnableMaintenance(ctx context.Context, id int) error
	DisableMaintenance(ctx context.Context, id int) error
	MaintenanceReboot(ctx context.Context, id int) error
	ClearPStore(ctx context.Context, id int) error
}

// ClusterServiceInterface defines the interface for Cluster operations.
type ClusterServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Cluster, error)
	Get(ctx context.Context, id int) (*Cluster, error)
	GetByName(ctx context.Context, name string) (*Cluster, error)
	GetStatus(ctx context.Context, id int) (*ClusterStatus, error)
}

// GroupServiceInterface defines the interface for Group operations.
type GroupServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Group, error)
	Get(ctx context.Context, id int) (*Group, error)
	GetByName(ctx context.Context, name string) (*Group, error)
}

// MediaSourceServiceInterface defines the interface for MediaSource operations.
type MediaSourceServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]MediaSource, error)
	Get(ctx context.Context, id int) (*MediaSource, error)
	GetByName(ctx context.Context, name string) (*MediaSource, error)
	ListISOs(ctx context.Context, opts ...ListOption) ([]MediaSource, error)
}

// ResourceGroupServiceInterface defines the interface for ResourceGroup operations.
type ResourceGroupServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]ResourceGroup, error)
	Get(ctx context.Context, id int) (*ResourceGroup, error)
	GetByName(ctx context.Context, name string) (*ResourceGroup, error)
}

// SettingsServiceInterface defines the interface for Settings operations.
type SettingsServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Setting, error)
	Get(ctx context.Context, id int) (*Setting, error)
	GetByKey(ctx context.Context, key string) (*Setting, error)
	GetValue(ctx context.Context, key string) (string, error)
	GetCloudName(ctx context.Context) (string, error)
}

// SystemServiceInterface defines the interface for System operations.
type SystemServiceInterface interface {
	GetInfo(ctx context.Context) (*SystemInfo, error)
	GetVersion(ctx context.Context) (string, error)
}

// SchemaServiceInterface defines the interface for Schema operations.
type SchemaServiceInterface interface {
	GetTableSchema(ctx context.Context, resource string) (*TableSchema, error)
	GetValidValues(ctx context.Context, resource, field string) (map[string]string, error)
	GetVMMachineTypes(ctx context.Context) (map[string]string, error)
	GetVMOSFamilies(ctx context.Context) (map[string]string, error)
}

// TagServiceInterface defines the interface for Tag operations.
type TagServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Tag, error)
	Get(ctx context.Context, id int) (*Tag, error)
	GetByName(ctx context.Context, name string) (*Tag, error)
	ListByCategory(ctx context.Context, categoryID int) ([]Tag, error)
}

// TagMemberServiceInterface defines the interface for TagMember operations.
type TagMemberServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]TagMember, error)
	ListByTag(ctx context.Context, tagID int) ([]TagMember, error)
	ListByMember(ctx context.Context, member string) ([]TagMember, error)
	Get(ctx context.Context, id int) (*TagMember, error)
	Create(ctx context.Context, req *TagMemberCreateRequest) (*TagMember, error)
	Update(ctx context.Context, id int, req *TagMemberUpdateRequest) (*TagMember, error)
	Delete(ctx context.Context, id int) error
	Assign(ctx context.Context, tagID int, member string) (*TagMember, error)
	Unassign(ctx context.Context, tagID int, member string) error
}

// VolumeServiceInterface defines the interface for Volume operations.
// Note: Unlike other resources, volumes use SHA1 hash strings as IDs instead of integers.
type VolumeServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Volume, error)
	ListByService(ctx context.Context, serviceID int) ([]Volume, error)
	Get(ctx context.Context, id string) (*Volume, error)
	GetByName(ctx context.Context, serviceID int, name string) (*Volume, error)
	Create(ctx context.Context, req *VolumeCreateRequest) (*Volume, error)
	Update(ctx context.Context, id string, req *VolumeUpdateRequest) (*Volume, error)
	Delete(ctx context.Context, id string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	Reset(ctx context.Context, id string) error
}

// VNetRuleServiceInterface defines the interface for network firewall rule operations.
type VNetRuleServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]VNetRule, error)
	ListByNetwork(ctx context.Context, vnetID int, opts ...ListOption) ([]VNetRule, error)
	Get(ctx context.Context, id int) (*VNetRule, error)
	GetByName(ctx context.Context, vnetID int, name string) (*VNetRule, error)
	Create(ctx context.Context, req *VNetRuleCreateRequest) (*VNetRule, error)
	Update(ctx context.Context, id int, req *VNetRuleUpdateRequest) (*VNetRule, error)
	Delete(ctx context.Context, id int) error
	Enable(ctx context.Context, id int, apply bool, forceApply bool) error
	Disable(ctx context.Context, id int, apply bool, forceApply bool) error
}

// VNetRuleAliasServiceInterface defines the interface for network rule alias operations.
type VNetRuleAliasServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]VNetRuleAlias, error)
	Get(ctx context.Context, id int) (*VNetRuleAlias, error)
	GetByName(ctx context.Context, name string) (*VNetRuleAlias, error)
	Create(ctx context.Context, req *VNetRuleAliasCreateRequest) (*VNetRuleAlias, error)
	Update(ctx context.Context, id int, req *VNetRuleAliasUpdateRequest) (*VNetRuleAlias, error)
	Delete(ctx context.Context, id int) error
}

// TenantServiceInterface defines the interface for Tenant operations.
type TenantServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Tenant, error)
	Get(ctx context.Context, id int) (*Tenant, error)
	GetByName(ctx context.Context, name string) (*Tenant, error)
	Create(ctx context.Context, req *TenantCreateRequest) (*Tenant, error)
	Update(ctx context.Context, id int, req *TenantUpdateRequest) (*Tenant, error)
	Delete(ctx context.Context, id int) error
	PowerOn(ctx context.Context, id int) error
	PowerOnWithNode(ctx context.Context, id int, preferredNode int) error
	PowerOff(ctx context.Context, id int) error
	Reset(ctx context.Context, id int) error
	Clone(ctx context.Context, id int, opts *TenantCloneOptions) error
	IsolateOn(ctx context.Context, id int) error
	IsolateOff(ctx context.Context, id int) error
}

// TenantNodeServiceInterface defines the interface for TenantNode operations.
type TenantNodeServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]TenantNode, error)
	ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantNode, error)
	Get(ctx context.Context, id int) (*TenantNode, error)
	GetByName(ctx context.Context, tenantID int, name string) (*TenantNode, error)
	Create(ctx context.Context, req *TenantNodeCreateRequest) (*TenantNode, error)
	Update(ctx context.Context, id int, req *TenantNodeUpdateRequest) (*TenantNode, error)
	Delete(ctx context.Context, id int) error
	PowerOn(ctx context.Context, id int) error
	PowerOff(ctx context.Context, id int) error
	Reset(ctx context.Context, id int) error
	Kill(ctx context.Context, id int) error
	Migrate(ctx context.Context, id int, targetNode int) error
}

// TenantStorageServiceInterface defines the interface for TenantStorage operations.
type TenantStorageServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]TenantStorage, error)
	ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantStorage, error)
	Get(ctx context.Context, id int) (*TenantStorage, error)
	Create(ctx context.Context, req *TenantStorageCreateRequest) (*TenantStorage, error)
	Update(ctx context.Context, id int, req *TenantStorageUpdateRequest) (*TenantStorage, error)
	Delete(ctx context.Context, id int) error
}

// SnapshotProfileServiceInterface defines the interface for SnapshotProfile operations.
type SnapshotProfileServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]SnapshotProfile, error)
	Get(ctx context.Context, id int) (*SnapshotProfile, error)
	GetByName(ctx context.Context, name string) (*SnapshotProfile, error)
	Create(ctx context.Context, req *SnapshotProfileCreateRequest) (*SnapshotProfile, error)
	Update(ctx context.Context, id int, req *SnapshotProfileUpdateRequest) (*SnapshotProfile, error)
	Delete(ctx context.Context, id int) error
}

// SnapshotProfilePeriodServiceInterface defines the interface for SnapshotProfilePeriod operations.
type SnapshotProfilePeriodServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]SnapshotProfilePeriod, error)
	ListByProfile(ctx context.Context, profileID int, opts ...ListOption) ([]SnapshotProfilePeriod, error)
	Get(ctx context.Context, id int) (*SnapshotProfilePeriod, error)
	GetByName(ctx context.Context, profileID int, name string) (*SnapshotProfilePeriod, error)
	Create(ctx context.Context, req *SnapshotProfilePeriodCreateRequest) (*SnapshotProfilePeriod, error)
	Update(ctx context.Context, id int, req *SnapshotProfilePeriodUpdateRequest) (*SnapshotProfilePeriod, error)
	Delete(ctx context.Context, id int) error
}

// AlarmServiceInterface defines the interface for Alarm operations.
type AlarmServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Alarm, error)
	ListActive(ctx context.Context, opts ...ListOption) ([]Alarm, error)
	ListByOwner(ctx context.Context, owner string, opts ...ListOption) ([]Alarm, error)
	ListByLevel(ctx context.Context, level string, opts ...ListOption) ([]Alarm, error)
	ListByAlarmType(ctx context.Context, alarmTypeKey string, opts ...ListOption) ([]Alarm, error)
	Get(ctx context.Context, id int) (*Alarm, error)
	Update(ctx context.Context, id int, req *AlarmUpdateRequest) (*Alarm, error)
	Snooze(ctx context.Context, id int, until int64) error
	Unsnooze(ctx context.Context, id int) error
	Resolve(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}

// AlarmTypeServiceInterface defines the interface for AlarmType operations.
// Alarm types are read-only reference data. Note: Uses string keys, not integers.
type AlarmTypeServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]AlarmType, error)
	Get(ctx context.Context, key string) (*AlarmType, error)
	ListByLevel(ctx context.Context, level string, opts ...ListOption) ([]AlarmType, error)
}

// TaskServiceInterface defines the interface for Task operations.
type TaskServiceInterface interface {
	List(ctx context.Context, opts ...ListOption) ([]Task, error)
	ListRunning(ctx context.Context, opts ...ListOption) ([]Task, error)
	ListByOwner(ctx context.Context, owner string, opts ...ListOption) ([]Task, error)
	ListEnabled(ctx context.Context, opts ...ListOption) ([]Task, error)
	Get(ctx context.Context, id int) (*Task, error)
	GetByID(ctx context.Context, taskID string) (*Task, error)
	GetByName(ctx context.Context, owner, name string) (*Task, error)
	Create(ctx context.Context, req *TaskCreateRequest) (*Task, error)
	Update(ctx context.Context, id int, req *TaskUpdateRequest) (*Task, error)
	Delete(ctx context.Context, id int) error
	Execute(ctx context.Context, id int, opts *TaskExecuteOptions) error
	Enable(ctx context.Context, id int) error
	Disable(ctx context.Context, id int) error
}

// Compile-time verification that concrete types satisfy their interfaces.
var (
	_ VMServiceInterface                    = (*VMService)(nil)
	_ VMNICServiceInterface                 = (*VMNICService)(nil)
	_ VMDriveServiceInterface               = (*VMDriveService)(nil)
	_ VMDeviceServiceInterface              = (*VMDeviceService)(nil)
	_ NetworkServiceInterface               = (*NetworkService)(nil)
	_ UserServiceInterface                  = (*UserService)(nil)
	_ MemberServiceInterface                = (*MemberService)(nil)
	_ CloudInitServiceInterface             = (*CloudInitService)(nil)
	_ NodeServiceInterface                  = (*NodeService)(nil)
	_ ClusterServiceInterface               = (*ClusterService)(nil)
	_ GroupServiceInterface                 = (*GroupService)(nil)
	_ MediaSourceServiceInterface           = (*MediaSourceService)(nil)
	_ ResourceGroupServiceInterface         = (*ResourceGroupService)(nil)
	_ SettingsServiceInterface              = (*SettingsService)(nil)
	_ SystemServiceInterface                = (*SystemService)(nil)
	_ SchemaServiceInterface                = (*SchemaService)(nil)
	_ TagServiceInterface                   = (*TagService)(nil)
	_ TagMemberServiceInterface             = (*TagMemberService)(nil)
	_ VolumeServiceInterface                = (*VolumeService)(nil)
	_ VNetRuleServiceInterface              = (*VNetRuleService)(nil)
	_ VNetRuleAliasServiceInterface         = (*VNetRuleAliasService)(nil)
	_ TenantServiceInterface                = (*TenantService)(nil)
	_ TenantNodeServiceInterface            = (*TenantNodeService)(nil)
	_ TenantStorageServiceInterface         = (*TenantStorageService)(nil)
	_ SnapshotProfileServiceInterface       = (*SnapshotProfileService)(nil)
	_ SnapshotProfilePeriodServiceInterface = (*SnapshotProfilePeriodService)(nil)
	_ AlarmServiceInterface                 = (*AlarmService)(nil)
	_ AlarmTypeServiceInterface             = (*AlarmTypeService)(nil)
	_ TaskServiceInterface                  = (*TaskService)(nil)
)
