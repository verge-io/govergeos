package vergeos

// TenantSnapshot represents a VergeOS tenant snapshot.
// Tenant snapshots capture the state of a tenant at a point in time.
type TenantSnapshot struct {
	// Key is the unique identifier for the snapshot.
	Key FlexInt `json:"$key,omitempty"`
	// Tenant is the parent tenant ID.
	Tenant FlexInt `json:"tenant,omitempty"`
	// Name is the snapshot name (readonly after create; may be server-assigned).
	Name string `json:"name,omitempty"`
	// Description is an optional description of the snapshot.
	Description string `json:"description,omitempty"`
	// Profile is the snapshot profile name used to create this snapshot.
	Profile string `json:"profile,omitempty"`
	// Period is the profile period name.
	Period string `json:"period,omitempty"`
	// MinSnapshots is the minimum number of snapshots to retain.
	MinSnapshots int `json:"min_snapshots,omitempty"`
	// Created is the creation timestamp (Unix epoch, readonly).
	Created int64 `json:"created,omitempty"`
	// Expires is the expiration timestamp (Unix epoch, 0 = never).
	Expires int64 `json:"expires,omitempty"`
	// Type is the snapshot coverage: full, partial_include, or partial_exclude.
	Type string `json:"type,omitempty"`
}

// TenantSnapshotCreateRequest is the request body for creating a tenant snapshot.
// Creation is a POST on /tenant_snapshots (not a tenant_snapshot_actions action).
type TenantSnapshotCreateRequest struct {
	// Tenant is the parent tenant ID (required).
	Tenant int `json:"tenant"`
	// Name is the snapshot name. VergeOS marks name as objectname required,readonly;
	// when omitted the platform may assign one.
	Name string `json:"name,omitempty"`
	// Profile is an optional snapshot profile name.
	Profile string `json:"profile,omitempty"`
	// Period is an optional profile period name.
	Period string `json:"period,omitempty"`
	// MinSnapshots is the minimum number of snapshots to retain (default 0).
	MinSnapshots *int `json:"min_snapshots,omitempty"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Expires is the expiration timestamp (Unix epoch, 0 = never).
	Expires *int64 `json:"expires,omitempty"`
	// Type is the snapshot coverage: full, partial_include, or partial_exclude.
	// Defaults to full when omitted.
	Type string `json:"type,omitempty"`
}

// TenantSnapshotUpdateRequest is the request body for updating a tenant snapshot.
type TenantSnapshotUpdateRequest struct {
	// Description is the snapshot description.
	Description *string `json:"description,omitempty"`
	// Expires is the expiration timestamp (0 = never expires).
	Expires *int64 `json:"expires,omitempty"`
}

// Tenant snapshot type values (tenant_snapshots.type).
const (
	TenantSnapshotTypeFull           = "full"
	TenantSnapshotTypePartialInclude = "partial_include"
	TenantSnapshotTypePartialExclude = "partial_exclude"
)

// tenantSnapshotListFields are the fields to request when listing tenant snapshots.
const tenantSnapshotListFields = "$key,tenant,name,description,profile,period,min_snapshots,created,expires,type"

// tenantSnapshotGetFields are the fields to request when getting a single tenant snapshot.
const tenantSnapshotGetFields = tenantSnapshotListFields
