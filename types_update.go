package vergeos

// UpdateSettings represents the system update configuration (singleton, key=1).
type UpdateSettings struct {
	// Key is the unique identifier (always 1).
	Key FlexInt `json:"$key,omitempty"`
	// Source is the update source ID.
	Source int `json:"source,omitempty"`
	// Branch is the update branch ID (FK to update_branches).
	Branch int `json:"branch,omitempty"`
	// BranchName is the branch name (computed field: branch#name).
	BranchName string `json:"branch_name,omitempty"`
	// AutoRefresh indicates whether to automatically refresh update info.
	AutoRefresh bool `json:"auto_refresh,omitempty"`
	// AutoUpdate indicates whether to automatically apply updates.
	AutoUpdate bool `json:"auto_update,omitempty"`
	// RebootRequired indicates whether a reboot is needed after update.
	RebootRequired bool `json:"reboot_required,omitempty"`
	// Installed indicates whether updates have been installed.
	Installed bool `json:"installed,omitempty"`
	// ApplyingUpdates indicates whether an update is being applied.
	ApplyingUpdates bool `json:"applying_updates,omitempty"`
	// SnapshotCloudOnUpdate indicates whether to snapshot before updating.
	SnapshotCloudOnUpdate bool `json:"snapshot_cloud_on_update,omitempty"`
	// SnapshotCloudExpireSeconds is the snapshot expiration time in seconds.
	SnapshotCloudExpireSeconds uint32 `json:"snapshot_cloud_expire_seconds,omitempty"`
	// RestartPendingPackages lists packages that still need a restart.
	// Read-only. VergeOS 26.2. The API may return an array or one delimited string.
	RestartPendingPackages StringList `json:"restart_pending_packages,omitempty"`
}

// UpdateBranch represents an update branch (e.g., "stable-4.13").
type UpdateBranch struct {
	// Key is the unique identifier for this branch.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the branch name (e.g., "stable-4.13").
	Name string `json:"name,omitempty"`
	// Description is the branch description.
	Description string `json:"description,omitempty"`
}

// UpdateSourcePackage represents a package available from an update source.
type UpdateSourcePackage struct {
	// Key is the unique identifier for this package record.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the package name (e.g., "ybos").
	Name string `json:"name,omitempty"`
	// Branch is the branch ID (FK to update_branches).
	Branch int `json:"branch,omitempty"`
	// Source is the update source ID.
	Source int `json:"source,omitempty"`
	// Version is the package version string (e.g., "4.13.1").
	Version string `json:"version,omitempty"`
	// Downloaded indicates whether the package has been downloaded.
	Downloaded bool `json:"downloaded,omitempty"`
}

// Field list constants for update resources
const (
	updateSettingsGetFields       = "$key,source,branch,branch#name as branch_name,auto_refresh,auto_update,reboot_required,installed,applying_updates,snapshot_cloud_on_update,snapshot_cloud_expire_seconds"
	updateSettings262Fields       = "restart_pending_packages"
	updateBranchListFields        = "$key,name,description"
	updateSourcePackageListFields = "$key,name,branch,source,version,downloaded"

	updatePackageListFields  = "$key,name,description,version,branch,branch#name as branch_name,type,optional,created,modified"
	updatePackage262Fields   = "dependencies"
	updateSourceStatusFields = "$key,source,source#name as source_name,status,info,nodes_updated,last_update"
)

// Update action values posted to update_actions.
const (
	// UpdateActionUninstall removes an installed optional package. VergeOS 26.2.
	UpdateActionUninstall = "uninstall"
)

// Update source status values. Uninstalling arrived in VergeOS 26.2.
const (
	UpdateSourceStatusIdle         = "idle"
	UpdateSourceStatusRefreshing   = "refreshing"
	UpdateSourceStatusDownloading  = "downloading"
	UpdateSourceStatusInstalling   = "installing"
	UpdateSourceStatusApplying     = "applying"
	UpdateSourceStatusError        = "error"
	UpdateSourceStatusUninstalling = "uninstalling"
)

// UpdatePackage is one installed package (update_packages).
// The row key is the package name.
type UpdatePackage struct {
	// Key is the package name.
	Key FlexText `json:"$key,omitempty"`
	// Name is the package name.
	Name string `json:"name,omitempty"`
	// Description is the package description.
	Description string `json:"description,omitempty"`
	// Version is the installed version.
	Version string `json:"version,omitempty"`
	// Branch is the branch id.
	Branch int `json:"branch,omitempty"`
	// BranchName is branch#name.
	BranchName string `json:"branch_name,omitempty"`
	// Type is the package type.
	Type string `json:"type,omitempty"`
	// Optional reports that the package can be installed and uninstalled.
	Optional bool `json:"optional,omitempty"`
	// Created is the creation timestamp.
	Created int64 `json:"created,omitempty"`
	// Modified is the last modified timestamp.
	Modified int64 `json:"modified,omitempty"`
	// Dependencies lists the packages this package requires. VergeOS 26.2.
	// The API may return an array or one delimited string.
	Dependencies StringList `json:"dependencies,omitempty"`
}

// UpdateSourceStatus is the current operation on an update source (update_source_status).
type UpdateSourceStatus struct {
	// Key is the status row key.
	Key FlexInt `json:"$key,omitempty"`
	// Source is the parent update source.
	Source int `json:"source,omitempty"`
	// SourceName is source#name.
	SourceName string `json:"source_name,omitempty"`
	// Status is idle, refreshing, downloading, installing, applying, uninstalling, or error.
	Status string `json:"status,omitempty"`
	// Info is the status detail.
	Info string `json:"info,omitempty"`
	// NodesUpdated is how many nodes have been updated.
	NodesUpdated int `json:"nodes_updated,omitempty"`
	// LastUpdate is when the status last changed.
	LastUpdate int64 `json:"last_update,omitempty"`
}
