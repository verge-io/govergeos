package vergeos

import "strconv"

// SharedObject is a VM shared from a parent system to a tenant (shared_objects).
//
// The tenant imports the share to make its own copy. Recipient is the tenant
// key. ID is the shared resource path, such as "vms/123", and is separate
// from Key.
type SharedObject struct {
	// Key is the shared object row key.
	Key FlexInt `json:"$key,omitempty"`
	// Recipient is the tenant that receives the share.
	Recipient FlexInt `json:"recipient,omitempty"`
	// RecipientName is the recipient tenant name.
	RecipientName string `json:"recipient_name,omitempty"`
	// Type is the shared resource type. Sharing a VM sends "vm".
	Type string `json:"type,omitempty"`
	// Name is the share name.
	Name string `json:"name,omitempty"`
	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Created is the creation timestamp (Unix epoch).
	Created int64 `json:"created,omitempty"`
	// Inbox reports whether the share is waiting to be imported.
	Inbox bool `json:"inbox"`
	// Snapshot is the machine snapshot path, such as "machine_snapshots/14".
	Snapshot string `json:"snapshot,omitempty"`
	// ID is the shared resource path, such as "vms/123".
	ID string `json:"id,omitempty"`
}

// SnapshotKey returns the machine snapshot key stored in Snapshot.
//
// The path is "machine_snapshots/{key}". A value without that slash returns false.
func (o SharedObject) SnapshotKey() (int, bool) {
	slash := -1
	for i := len(o.Snapshot) - 1; i >= 0; i-- {
		if o.Snapshot[i] == '/' {
			slash = i
			break
		}
	}
	if slash < 0 || slash == len(o.Snapshot)-1 {
		return 0, false
	}
	key, err := strconv.Atoi(o.Snapshot[slash+1:])
	if err != nil {
		return 0, false
	}
	return key, true
}

// SharedObjectCreateRequest shares a VM with a tenant.
//
// Create snapshots the VM, then posts the shared object. The snapshot does
// not expire. Tenant is required. Set VM or VMName. An empty Name uses the
// VM name. An empty SnapshotName is generated.
type SharedObjectCreateRequest struct {
	// Tenant is the recipient tenant key.
	Tenant int
	// VM is the VM $key to share. VMName is used when VM is 0.
	VM int
	// VMName is the VM name to share. VM is used when both are set.
	VMName string
	// Name is the shared object name. Empty uses the VM name.
	Name string
	// Description is an optional description.
	Description string
	// SnapshotName names the machine snapshot. Empty generates a name.
	SnapshotName string
}

const sharedObjectListFields = "$key,recipient,recipient#name as recipient_name,type,name,description,created,inbox,snapshot,id"
