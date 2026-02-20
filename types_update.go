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
	// SnapshotCloudOnUpdate indicates whether to snapshot before updating.
	SnapshotCloudOnUpdate bool `json:"snapshot_cloud_on_update,omitempty"`
	// SnapshotCloudExpireSeconds is the snapshot expiration time in seconds.
	SnapshotCloudExpireSeconds uint32 `json:"snapshot_cloud_expire_seconds,omitempty"`
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
	updateSettingsGetFields       = "$key,source,branch,branch#name as branch_name,auto_refresh,auto_update,reboot_required,installed,snapshot_cloud_on_update,snapshot_cloud_expire_seconds"
	updateBranchListFields        = "$key,name,description"
	updateSourcePackageListFields = "$key,name,branch,source,version,downloaded"
)
