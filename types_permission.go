package vergeos

// Permission represents a VergeOS permission granting an identity access to a resource.
// Permissions are row-based. The grantee is an identities-table key, the same
// value as User.Identity or Group.Identity. Row 0 covers the whole table.
// A positive row is one resource instance.
type Permission struct {
	// Key is the unique identifier for this permission.
	Key FlexInt `json:"$key,omitempty"`

	// Identity is the identities-table key this permission is granted to.
	Identity IdentityKey `json:"identity,omitempty"`
	// IdentityDisplay is the display name of the identity (read-only).
	IdentityDisplay string `json:"identity_display,omitempty"`
	// IdentityOwner is the owner of the identity (read-only).
	IdentityOwner string `json:"identity_owner,omitempty"`

	// Table is the resource type (e.g., "vms", "networks", "volumes").
	Table string `json:"table,omitempty"`
	// TableID is the internal ID of the target table (set automatically).
	TableID int64 `json:"tableid,omitempty"`
	// Row is the resource id. 0 covers the whole table.
	Row int64 `json:"row,omitempty"`
	// RowDisplay is the display name of the resource (read-only).
	RowDisplay string `json:"rowdisplay,omitempty"`

	// Permission flags
	// List grants permission to see the resource in listings.
	List bool `json:"list"`
	// Read grants permission to view/read the resource details.
	Read bool `json:"read"`
	// Create grants permission to create child/related resources.
	Create bool `json:"create"`
	// Modify grants permission to update the resource.
	Modify bool `json:"modify"`
	// Delete grants permission to delete the resource.
	Delete bool `json:"delete"`
}

// PermissionCreateRequest is the request body for creating a permission.
type PermissionCreateRequest struct {
	// Identity is the identities-table key to grant (required).
	// Use User.Identity, Group.Identity, Users.IdentityOf, or Groups.IdentityOf.
	// A user row key or a group row key is a different number.
	Identity IdentityKey `json:"identity"`
	// Table is the resource type (required, e.g., "vms", "networks").
	Table string `json:"table"`
	// Row is the resource id. 0 covers the whole table and is sent in the JSON body.
	// A negative value is rejected.
	Row int64 `json:"row"`

	// Permission flags (all default to false if not specified)
	// List grants permission to see the resource in listings.
	List *bool `json:"list,omitempty"`
	// Read grants permission to view/read the resource details.
	// Note: Setting read=true automatically sets list=true.
	Read *bool `json:"read,omitempty"`
	// Create grants permission to create child/related resources.
	Create *bool `json:"create,omitempty"`
	// Modify grants permission to update the resource.
	Modify *bool `json:"modify,omitempty"`
	// Delete grants permission to delete the resource.
	Delete *bool `json:"delete,omitempty"`
}

// PermissionUpdateRequest is the request body for updating a permission.
type PermissionUpdateRequest struct {
	// Permission flags
	// List grants permission to see the resource in listings.
	List *bool `json:"list,omitempty"`
	// Read grants permission to view/read the resource details.
	Read *bool `json:"read,omitempty"`
	// Create grants permission to create child/related resources.
	Create *bool `json:"create,omitempty"`
	// Modify grants permission to update the resource.
	Modify *bool `json:"modify,omitempty"`
	// Delete grants permission to delete the resource.
	Delete *bool `json:"delete,omitempty"`
}

// permissionListFields are the fields to request when listing permissions.
const permissionListFields = "$key,identity,identity_display,identity_owner,table,tableid,row,rowdisplay,list,read,create,modify,delete"

// permissionGetFields are the fields to request when getting a single permission.
const permissionGetFields = permissionListFields

// Common resource table names for use with permissions.
const (
	PermissionTableVMs       = "vms"
	PermissionTableNetworks  = "vnets"
	PermissionTableVolumes   = "volumes"
	PermissionTableTenants   = "tenants"
	PermissionTableUsers     = "users"
	PermissionTableGroups    = "groups"
	PermissionTableFiles     = "media_images"
	PermissionTableClusters  = "clusters"
	PermissionTableNodes     = "machines"
	PermissionTableSnapshots = "snapshots"
)
