package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// VMSnapshotService handles VM snapshot operations.
// VM Snapshots are point-in-time copies of a VM's state.
type VMSnapshotService struct {
	client *Client
}

// List returns all VM snapshots, with optional filtering and pagination.
func (s *VMSnapshotService) List(ctx context.Context, opts ...ListOption) ([]VMSnapshot, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = vmSnapshotListFields
	}

	params := options.toQueryParams()

	var snapshots []VMSnapshot
	if err := s.client.get(ctx, "/machine_snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

// ListByVM returns all snapshots for a specific VM.
// vmID is the VM $key (VM.Key). Snapshots are stored against the machine key,
// so the VM is resolved before the machine_snapshots filter is applied.
func (s *VMSnapshotService) ListByVM(ctx context.Context, vmID int, opts ...ListOption) ([]VMSnapshot, error) {
	machine, err := s.client.machineKeyForVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = vmSnapshotListFields
	}

	// Add machine filter
	if options.Filter != "" {
		options.Filter = fmt.Sprintf("(%s) and machine eq %d", options.Filter, machine)
	} else {
		options.Filter = fmt.Sprintf("machine eq %d", machine)
	}

	params := options.toQueryParams()

	var snapshots []VMSnapshot
	if err := s.client.get(ctx, "/machine_snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

// ListExpiring returns snapshots expiring within the specified number of days.
func (s *VMSnapshotService) ListExpiring(ctx context.Context, days int, opts ...ListOption) ([]VMSnapshot, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = vmSnapshotListFields
	}

	// Add expiring filter using API timestamp math
	if options.Filter != "" {
		options.Filter = fmt.Sprintf("(%s) and expires lt {$add({$now},%d)} and expires gt 0", options.Filter, days*86400)
	} else {
		options.Filter = fmt.Sprintf("expires lt {$add({$now},%d)} and expires gt 0", days*86400)
	}

	params := options.toQueryParams()

	var snapshots []VMSnapshot
	if err := s.client.get(ctx, "/machine_snapshots", params, &snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}

// Get returns a single VM snapshot by ID.
func (s *VMSnapshotService) Get(ctx context.Context, id int) (*VMSnapshot, error) {
	params := url.Values{}
	params.Set("fields", vmSnapshotGetFields)

	var snapshot VMSnapshot
	endpoint := fmt.Sprintf("/machine_snapshots/%d", id)
	if err := s.client.get(ctx, endpoint, params, &snapshot); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMSnapshot", ID: id}
		}
		return nil, err
	}

	return &snapshot, nil
}

// GetByName returns a VM snapshot by name within a specific VM.
// vmID is the VM $key (VM.Key).
func (s *VMSnapshotService) GetByName(ctx context.Context, vmID int, name string) (*VMSnapshot, error) {
	snapshots, err := s.ListByVM(ctx, vmID, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, &NotFoundError{Resource: "VMSnapshot", ID: name}
	}
	if err := requireUniqueName("VMSnapshot", name, snapshots, func(snap VMSnapshot) any { return snap.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("VMSnapshot", name, snapshots[0].Name, name); err != nil {
		return nil, err
	}

	return &snapshots[0], nil
}

// vmSnapshotCreateBody is the machine_snapshots create payload.
// Machine is the machine key, resolved from VMSnapshotCreateRequest.VM.
type vmSnapshotCreateBody struct {
	Machine     int    `json:"machine"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ExpiresType string `json:"expires_type,omitempty"`
	Expires     *int64 `json:"expires,omitempty"`
	Quiesce     *bool  `json:"quiesce,omitempty"`
}

// Create creates a new VM snapshot and returns the created snapshot.
// req.VM is the VM $key. The machine key is resolved and posted as "machine".
func (s *VMSnapshotService) Create(ctx context.Context, req *VMSnapshotCreateRequest) (*VMSnapshot, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.VM <= 0 {
		return nil, &ValidationError{Field: "vm", Message: "VM ID is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	machine, err := s.client.machineKeyForVM(ctx, req.VM)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if req.ExpiresType == "" {
		req.ExpiresType = "date"
	}

	body := vmSnapshotCreateBody{
		Machine:     machine,
		Name:        req.Name,
		Description: req.Description,
		ExpiresType: req.ExpiresType,
		Expires:     req.Expires,
		Quiesce:     req.Quiesce,
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/machine_snapshots", body, &resp); err != nil {
		return nil, err
	}

	// Extract the created snapshot's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created snapshot
	return s.Get(ctx, id)
}

// Update updates a VM snapshot and returns the updated snapshot.
func (s *VMSnapshotService) Update(ctx context.Context, id int, req *VMSnapshotUpdateRequest) (*VMSnapshot, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/machine_snapshots/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMSnapshot", ID: id}
		}
		return nil, err
	}

	// Read back the updated snapshot
	return s.Get(ctx, id)
}

// Delete deletes a VM snapshot.
func (s *VMSnapshotService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/machine_snapshots/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMSnapshot", ID: id}
		}
		return err
	}
	return nil
}

// Restore restores a VM from a snapshot.
// The VM will be reverted to the state at the time of the snapshot.
//
// snap_machine is a machine key. vm_actions wants the snapshot VM's $key
// (the vms row with that machine and is_snapshot true). Posting the machine
// key returns 404, or restores a different VM when that key collides.
func (s *VMSnapshotService) Restore(ctx context.Context, id int, opts *VMSnapshotRestoreOptions) error {
	snapshot, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	snapMachine := int(snapshot.SnapMachine)
	if snapMachine <= 0 {
		return &ValidationError{Field: "snap_machine", Message: fmt.Sprintf("snapshot %d has no snap_machine", id)}
	}

	vmKey, err := s.client.vmKeyForMachine(ctx, snapMachine, true)
	if err != nil {
		return fmt.Errorf("vergeos: failed to resolve snapshot VM for snapshot %d: %w", id, err)
	}

	params := map[string]any{}
	if opts != nil {
		params["poweron"] = opts.PowerOn
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     vmKey,
		Action: "restore",
		Params: params,
	}
	action.Params["snapshot"] = id

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to restore VM snapshot %d: %w", id, err)
	}
	return nil
}

// SetNeverExpires sets a snapshot to never expire.
// VergeOS stores that as Expires 0. expires_type is a write argument and is not returned.
func (s *VMSnapshotService) SetNeverExpires(ctx context.Context, id int) (*VMSnapshot, error) {
	expiresType := "never"
	return s.Update(ctx, id, &VMSnapshotUpdateRequest{
		ExpiresType: &expiresType,
	})
}

// SetExpires sets the expiration timestamp for a snapshot.
func (s *VMSnapshotService) SetExpires(ctx context.Context, id int, expires int64) (*VMSnapshot, error) {
	expiresType := "date"
	return s.Update(ctx, id, &VMSnapshotUpdateRequest{
		ExpiresType: &expiresType,
		Expires:     &expires,
	})
}
