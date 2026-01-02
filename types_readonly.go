package vergeos

// Cluster represents a VergeOS cluster.
type Cluster struct {
	// ID is the unique identifier for the cluster.
	ID FlexInt `json:"$key,omitempty"`
	// Name is the cluster name.
	Name string `json:"name"`
	// Description is the cluster description.
	Description string `json:"description,omitempty"`
	// Enabled indicates whether the cluster is enabled.
	Enabled bool `json:"enabled"`
	// Type is the cluster type.
	Type string `json:"type,omitempty"`
	// Status contains cluster status information.
	Status *ClusterStatus `json:"status,omitempty"`
}

// ClusterStatus contains cluster status information.
type ClusterStatus struct {
	// TotalNodes is the total number of nodes.
	TotalNodes int `json:"total_nodes,omitempty"`
	// OnlineNodes is the number of online nodes.
	OnlineNodes int `json:"online_nodes,omitempty"`
	// OnlineRAM is the total online RAM in MB.
	OnlineRAM int64 `json:"online_ram,omitempty"`
	// OnlineCores is the total online CPU cores.
	OnlineCores int `json:"online_cores,omitempty"`
	// PhysRAMUsed is the physical RAM used in MB.
	PhysRAMUsed int64 `json:"phys_ram_used,omitempty"`
	// RAMPerUnit is the RAM per unit.
	RAMPerUnit int64 `json:"ram_per_unit,omitempty"`
	// CoresPerUnit is the cores per unit.
	CoresPerUnit int `json:"cores_per_unit,omitempty"`
	// TargetRAMPct is the target RAM percentage.
	TargetRAMPct float64 `json:"target_ram_pct,omitempty"`
}

// Node represents a VergeOS node.
type Node struct {
	// ID is the unique identifier for the node.
	ID int `json:"id,omitempty"`
	// Name is the node name.
	Name string `json:"name"`
	// Description is the node description.
	Description string `json:"description,omitempty"`
	// Physical indicates whether this is a physical node.
	Physical bool `json:"physical,omitempty"`
	// Cluster is the cluster ID the node belongs to.
	Cluster int `json:"cluster,omitempty"`
	// Machine is the machine ID.
	Machine int `json:"machine,omitempty"`
	// Model is the hardware model.
	Model string `json:"model,omitempty"`
	// CPU is the CPU description.
	CPU string `json:"cpu,omitempty"`
	// CPUSpeed is the CPU speed.
	CPUSpeed string `json:"cpu_speed,omitempty"`
	// RAM is the total RAM in MB.
	RAM int `json:"ram,omitempty"`
	// Cores is the total CPU cores.
	Cores int `json:"cores,omitempty"`
	// Maintenance indicates whether the node is in maintenance mode.
	Maintenance bool `json:"maintenance,omitempty"`
	// YBVersion is the VergeOS version.
	YBVersion string `json:"yb_version,omitempty"`
	// OSVersion is the OS version.
	OSVersion string `json:"os_version,omitempty"`
}

// NodeStatus contains node status information.
type NodeStatus struct {
	// CPUTemp is the CPU temperature.
	CPUTemp float64 `json:"cpu_temp,omitempty"`
	// CPUUsage is the CPU usage percentage.
	CPUUsage float64 `json:"cpu_usage,omitempty"`
	// MemoryTotal is the total memory in bytes.
	MemoryTotal int64 `json:"memory_total,omitempty"`
	// MemoryUsed is the used memory in bytes.
	MemoryUsed int64 `json:"memory_used,omitempty"`
	// MemoryPct is the memory usage percentage.
	MemoryPct float64 `json:"memory_pct,omitempty"`
}

// Group represents a VergeOS group.
type Group struct {
	// ID is the unique identifier for the group.
	ID FlexInt `json:"$key,omitempty"`
	// Name is the group name.
	Name string `json:"name"`
	// Description is the group description.
	Description string `json:"description,omitempty"`
	// Enabled indicates whether the group is enabled.
	Enabled bool `json:"enabled"`
	// Type is the group type.
	Type string `json:"type,omitempty"`
}

// MediaSource represents a media source (ISO, etc.) in VergeOS.
type MediaSource struct {
	// ID is the unique identifier for the media source.
	ID FlexInt `json:"$key,omitempty"`
	// Name is the file name.
	Name string `json:"name"`
	// Description is the description.
	Description string `json:"description,omitempty"`
	// Type is the file type.
	Type string `json:"type,omitempty"`
	// Size is the file size in bytes.
	Size int64 `json:"size,omitempty"`
	// Path is the file path.
	Path string `json:"path,omitempty"`
}

// ResourceGroup represents a resource group in VergeOS.
type ResourceGroup struct {
	// ID is the unique identifier for the resource group (UUID string).
	ID string `json:"$key,omitempty"`
	// Name is the resource group name.
	Name string `json:"name"`
	// Description is the description.
	Description string `json:"description,omitempty"`
	// Type is the resource group type.
	Type string `json:"type,omitempty"`
	// Enabled indicates whether the resource group is enabled.
	Enabled bool `json:"enabled"`
}

// Setting represents a VergeOS system setting.
type Setting struct {
	// Key is the setting key/name (used as identifier).
	Key string `json:"key,omitempty"`
	// Value is the setting value.
	Value string `json:"value,omitempty"`
	// DefaultValue is the default value.
	DefaultValue string `json:"default_value,omitempty"`
	// Description is the setting description.
	Description string `json:"description,omitempty"`
}

// SystemInfo represents VergeOS system information.
type SystemInfo struct {
	// Name is the API name (e.g., "v4").
	Name string `json:"name,omitempty"`
	// Version is the VergeOS version.
	Version string `json:"version,omitempty"`
	// Hash is the build hash.
	Hash string `json:"hash,omitempty"`
}

// TableSchema represents the schema for an API resource.
type TableSchema struct {
	// Fields contains field definitions.
	Fields map[string]TableField `json:"fields,omitempty"`
}

// TableField represents a field definition in a table schema.
type TableField struct {
	// Type is the field type.
	Type string `json:"type,omitempty"`
	// List contains valid values and their descriptions.
	List map[string]string `json:"list,omitempty"`
	// Required indicates whether the field is required.
	Required bool `json:"required,omitempty"`
}

// Field list constants for read-only resources
const (
	clusterListFields       = "$key,name,description,enabled,type"
	nodeListFields          = "id,name,description,physical,cluster,machine,model,cpu,cpu_speed,ram,cores,maintenance,yb_version,os_version"
	groupListFields         = "$key,name,description,enabled,type"
	mediaSourceListFields   = "$key,name,description,type,size,path"
	resourceGroupListFields = "$key,name,description,type,enabled"
	settingListFields       = "key,value,default_value,description"
)
