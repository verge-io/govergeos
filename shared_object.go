package vergeos

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
)

// sharedSnapshotCreateBody is the machine_snapshots row created for a share.
// Shared snapshots do not expire, and created_manually marks the row as a
// user snapshot.
type sharedSnapshotCreateBody struct {
	Machine         int    `json:"machine"`
	Name            string `json:"name"`
	ExpiresType     string `json:"expires_type"`
	CreatedManually bool   `json:"created_manually"`
}

// sharedObjectCreateBody is the shared_objects create payload.
type sharedObjectCreateBody struct {
	Recipient   int    `json:"recipient"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Snapshot    string `json:"snapshot"`
	Description string `json:"description,omitempty"`
}

// sharedObjectActionRequest is the body for POST /shared_object_actions.
type sharedObjectActionRequest struct {
	SharedObject int    `json:"shared_object"`
	Action       string `json:"action"`
}

// SharedObjectService shares VMs with tenants (shared_objects).
type SharedObjectService struct {
	client *Client
}

// List returns shared objects, with optional filtering and pagination.
func (s *SharedObjectService) List(ctx context.Context, opts ...ListOption) ([]SharedObject, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = sharedObjectListFields
	}

	var objects []SharedObject
	if err := s.client.get(ctx, "/shared_objects", options.toQueryParams(), &objects); err != nil {
		return nil, err
	}
	return objects, nil
}

// ListByTenant returns shared objects addressed to one tenant.
func (s *SharedObjectService) ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]SharedObject, error) {
	if tenantID <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	opts = append([]ListOption{WithFilter(fmt.Sprintf("recipient eq %d", tenantID))}, opts...)
	return s.List(ctx, opts...)
}

// ListInbox returns shares that are waiting to be imported.
func (s *SharedObjectService) ListInbox(ctx context.Context, opts ...ListOption) ([]SharedObject, error) {
	opts = append([]ListOption{WithFilter("inbox eq true")}, opts...)
	return s.List(ctx, opts...)
}

// Get returns one shared object by key.
func (s *SharedObjectService) Get(ctx context.Context, id int) (*SharedObject, error) {
	params := url.Values{}
	params.Set("fields", sharedObjectListFields)

	var object SharedObject
	endpoint := fmt.Sprintf("/shared_objects/%d", id)
	if err := s.client.get(ctx, endpoint, params, &object); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "SharedObject", ID: id}
		}
		return nil, err
	}
	return &object, nil
}

// GetByName returns the share with this name for one tenant.
//
// Returns NotFoundError when nothing matches, and AmbiguousNameError when
// more than one share matches.
func (s *SharedObjectService) GetByName(ctx context.Context, tenantID int, name string) (*SharedObject, error) {
	if tenantID <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	objects, err := s.List(ctx, WithFilter(fmt.Sprintf("recipient eq %d and name eq '%s'", tenantID, escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, &NotFoundError{Resource: "SharedObject", ID: name}
	}
	if err := requireUniqueName("SharedObject", name, objects, func(object SharedObject) any { return object.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("SharedObject", name, objects[0].Name, name); err != nil {
		return nil, err
	}
	return &objects[0], nil
}

// Create shares a VM with a tenant.
//
// The VM is snapshotted first (expires_type "never", created_manually true).
// The share stores that snapshot as "machine_snapshots/{key}". When the
// shared object POST fails, the snapshot is deleted.
func (s *SharedObjectService) Create(ctx context.Context, req *SharedObjectCreateRequest) (*SharedObject, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Tenant <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	if req.VM <= 0 && req.VMName == "" {
		return nil, &ValidationError{Field: "vm", Message: "vm or vm name is required"}
	}

	vm, err := s.vmForShare(ctx, req)
	if err != nil {
		return nil, err
	}
	if vm.Machine <= 0 {
		return nil, &ValidationError{Field: "machine", Message: fmt.Sprintf("VM %d has no machine key", vm.Key.Int())}
	}

	name := req.Name
	if name == "" {
		name = vm.Name
	}
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	snapshotName := req.SnapshotName
	if snapshotName == "" {
		suffix, err := sharedSnapshotSuffix()
		if err != nil {
			return nil, err
		}
		snapshotName = fmt.Sprintf("share-%s-%d", name, suffix)
	}

	var snapshot apiResponse
	snapshotBody := sharedSnapshotCreateBody{
		Machine:         vm.Machine,
		Name:            snapshotName,
		ExpiresType:     "never",
		CreatedManually: true,
	}
	if err := s.client.post(ctx, "/machine_snapshots", snapshotBody, &snapshot); err != nil {
		return nil, err
	}
	snapshotKey, err := getKey(snapshot)
	if err != nil {
		return nil, err
	}

	body := sharedObjectCreateBody{
		Recipient:   req.Tenant,
		Type:        "vm",
		Name:        name,
		Snapshot:    fmt.Sprintf("machine_snapshots/%d", snapshotKey),
		Description: req.Description,
	}
	var created apiResponse
	if err := s.client.post(ctx, "/shared_objects", body, &created); err != nil {
		_ = s.client.delete(ctx, fmt.Sprintf("/machine_snapshots/%d", snapshotKey))
		return nil, err
	}
	id, err := getKey(created)
	if err != nil {
		return s.GetByName(ctx, req.Tenant, name)
	}
	return s.Get(ctx, id)
}

// Import starts an import of the share into the recipient tenant.
//
// The action is POST /shared_object_actions with action "import".
func (s *SharedObjectService) Import(ctx context.Context, id int) error {
	return s.action(ctx, id, "import")
}

// Refresh updates the share's snapshot.
//
// The action is POST /shared_object_actions with action "refresh".
func (s *SharedObjectService) Refresh(ctx context.Context, id int) error {
	return s.action(ctx, id, "refresh")
}

// Delete removes the share. A VM the tenant already imported is left in place.
func (s *SharedObjectService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/shared_objects/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "SharedObject", ID: id}
		}
		return err
	}
	return nil
}

func (s *SharedObjectService) action(ctx context.Context, id int, action string) error {
	if id <= 0 {
		return &ValidationError{Field: "shared_object", Message: "shared object key is required"}
	}
	body := sharedObjectActionRequest{SharedObject: id, Action: action}
	if err := s.client.post(ctx, "/shared_object_actions", body, nil); err != nil {
		return fmt.Errorf("vergeos: failed to %s shared object: %w", action, err)
	}
	return nil
}

func (s *SharedObjectService) vmForShare(ctx context.Context, req *SharedObjectCreateRequest) (*VM, error) {
	if req.VM > 0 {
		return s.client.VMs.Get(ctx, req.VM)
	}
	return s.client.VMs.GetByName(ctx, req.VMName)
}

func sharedSnapshotSuffix() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(90000))
	if err != nil {
		return 0, fmt.Errorf("vergeos: failed to name shared snapshot: %w", err)
	}
	return int(n.Int64()) + 10000, nil
}
