package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

const (
	// Power action constants
	vmActionPowerOn          = "poweron"
	vmActionPowerOff         = "poweroff"
	vmActionReset            = "reset"
	vmActionKill             = "kill"
	vmActionClone            = "clone"
	vmActionSnapshot         = "quiesce_snapshot"
	vmActionGuestReboot      = "guestreboot"
	vmActionRefresh          = "refresh"
	vmActionHibernate        = "hibernate"
	vmActionChangeCD         = "changecd"
	vmActionChangeNet        = "changenet"
	vmActionPaste            = "paste"
	vmActionRestore          = "restore"
	vmActionRecoverCloudSnap = "recover_cloudsnapshot"
	vmActionExecute          = "execute"
	vmActionFSyncStrict      = "fsync_strict"
	vmActionEraseDrive       = "erase_drive"

	// Polling configuration
	powerStateMaxRetries   = 30
	powerStatePollInterval = 5 * time.Second
)

// VMService handles VM operations.
type VMService struct {
	client *Client
}

// List returns all VMs, with optional filtering and pagination.
func (s *VMService) List(ctx context.Context, opts ...ListOption) ([]VM, error) {
	options := applyListOptions(opts)

	// Use VM-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = vmListFields
	}

	params := options.toQueryParams()

	var vms []VM
	if err := s.client.get(ctx, "/vms", params, &vms); err != nil {
		return nil, err
	}

	return vms, nil
}

// Get returns a single VM by ID.
func (s *VMService) Get(ctx context.Context, id int) (*VM, error) {
	params := url.Values{}
	params.Set("fields", vmGetFields)

	var vm VM
	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.get(ctx, endpoint, params, &vm); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VM", ID: id}
		}
		return nil, err
	}

	return &vm, nil
}

// Create creates a new VM and returns the created VM.
func (s *VMService) Create(ctx context.Context, req *VMCreateRequest) (*VM, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.CPUCores <= 0 {
		return nil, &ValidationError{Field: "cpu_cores", Message: "cpu_cores must be positive"}
	}
	if req.RAM <= 0 {
		return nil, &ValidationError{Field: "ram", Message: "ram must be positive"}
	}

	// Set defaults
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vms", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created VM's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created VM
	return s.Get(ctx, id)
}

// Update updates a VM and returns the updated VM.
func (s *VMService) Update(ctx context.Context, id int, req *VMUpdateRequest) (*VM, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VM", ID: id}
		}
		return nil, err
	}

	// Read back the updated VM
	return s.Get(ctx, id)
}

// Delete deletes a VM.
func (s *VMService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VM", ID: id}
		}
		return err
	}
	return nil
}

// PowerOn powers on a VM and waits for it to start.
func (s *VMService) PowerOn(ctx context.Context, id int) error {
	// Get current state
	vm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Already running
	if vm.PowerState {
		return nil
	}

	// Send power on action
	action := vmAction{
		VM:     id,
		Action: vmActionPowerOn,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power on VM %d: %w", id, err)
	}

	// Wait for VM to start
	return s.waitForPowerState(ctx, id, true)
}

// PowerOff powers off a VM and waits for it to stop.
func (s *VMService) PowerOff(ctx context.Context, id int) error {
	// Get current state
	vm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Already stopped
	if !vm.PowerState {
		return nil
	}

	// Send kill action
	action := vmAction{
		VM:     id,
		Action: vmActionKill,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power off VM %d: %w", id, err)
	}

	// Wait for VM to stop
	return s.waitForPowerState(ctx, id, false)
}

// waitForPowerState waits for a VM to reach the desired power state.
func (s *VMService) waitForPowerState(ctx context.Context, id int, desiredState bool) error {
	stateStr := "stopped"
	if desiredState {
		stateStr = "running"
	}

	for i := 0; i < powerStateMaxRetries; i++ {
		vm, err := s.Get(ctx, id)
		if err != nil {
			return err
		}

		if vm.PowerState == desiredState {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(powerStatePollInterval):
		}
	}

	return &TimeoutError{Resource: "VM", ID: id, Action: "become " + stateStr}
}

// Reset sends a reset signal to a running VM (equivalent to pressing the reset button).
func (s *VMService) Reset(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionReset,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to reset VM %d: %w", id, err)
	}
	return nil
}

// GuestReboot sends a reboot request to the guest OS via ACPI.
func (s *VMService) GuestReboot(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionGuestReboot,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to guest reboot VM %d: %w", id, err)
	}
	return nil
}

// GuestShutdown sends a graceful shutdown request to the guest OS via ACPI (poweroff action).
func (s *VMService) GuestShutdown(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionPowerOff,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to guest shutdown VM %d: %w", id, err)
	}
	return nil
}

// VMCloneOptions contains options for cloning a VM.
type VMCloneOptions struct {
	// Name is the name for the cloned VM. Defaults to "${NAME}_${TIMESTAMP}".
	Name string
	// PreserveMACs indicates whether to preserve MAC addresses from the source VM.
	PreserveMACs bool
}

// Clone creates a copy of a VM.
func (s *VMService) Clone(ctx context.Context, id int, opts *VMCloneOptions) error {
	params := map[string]any{}
	if opts != nil {
		if opts.Name != "" {
			params["name"] = opts.Name
		}
		params["preserve_macs"] = opts.PreserveMACs
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionClone,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to clone VM %d: %w", id, err)
	}
	return nil
}

// VMSnapshotOptions contains options for taking a VM snapshot.
type VMSnapshotOptions struct {
	// Retention is the snapshot retention duration in seconds. Defaults to 86400 (24 hours).
	Retention int
	// Quiesce indicates whether to quiesce the filesystem before snapshot (requires guest agent).
	Quiesce bool
}

// Snapshot takes a snapshot of a VM.
func (s *VMService) Snapshot(ctx context.Context, id int, opts *VMSnapshotOptions) error {
	params := map[string]any{}
	if opts != nil {
		if opts.Retention > 0 {
			params["retention"] = opts.Retention
		}
		params["quiesce"] = opts.Quiesce
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionSnapshot,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to snapshot VM %d: %w", id, err)
	}
	return nil
}

// VMMigrateOptions contains options for migrating a VM to another node.
type VMMigrateOptions struct {
	// TargetNode is the destination node ID (required).
	TargetNode int
	// Live indicates whether to perform a live migration (default: true).
	Live *bool
}

// Migrate migrates a VM to another node.
// Live migration moves the VM without downtime if Live is true (default).
func (s *VMService) Migrate(ctx context.Context, id int, opts *VMMigrateOptions) error {
	if opts == nil {
		return &ValidationError{Message: "migrate options are required"}
	}
	if opts.TargetNode <= 0 {
		return &ValidationError{Field: "target_node", Message: "target_node is required"}
	}

	params := map[string]any{
		"node": opts.TargetNode,
	}

	// Default to live migration
	if opts.Live == nil || *opts.Live {
		params["method"] = "live"
	} else {
		params["method"] = "auto"
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: "migrate",
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to migrate VM %d: %w", id, err)
	}
	return nil
}

// GetConsoleURL returns the console URL for a VM.
func (s *VMService) GetConsoleURL(ctx context.Context, id int) (string, error) {
	vm, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}

	if !vm.PowerState {
		return "", fmt.Errorf("vergeos: VM %d is not running", id)
	}

	// Console URL format depends on the console type
	consoleURL := fmt.Sprintf("%s/ui/#/main/vms/%d/console", s.client.baseURL, id)
	return consoleURL, nil
}

// Refresh triggers a guest-agent and status refresh for a VM.
func (s *VMService) Refresh(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionRefresh,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to refresh VM %d: %w", id, err)
	}
	return nil
}

// Hibernate sends an ACPI hibernate signal to a VM.
// The guest OS must support ACPI hibernate.
func (s *VMService) Hibernate(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionHibernate,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to hibernate VM %d: %w", id, err)
	}
	return nil
}

// VMChangeCDOptions contains options for changing the CD/ISO attached to a VM.
type VMChangeCDOptions struct {
	// Drive is the drive name or index to change.
	Drive string `json:"drive,omitempty"`
	// Media is the path to the ISO file, or empty to eject.
	Media string `json:"media,omitempty"`
}

// ChangeCD changes the CD/ISO attached to a VM.
func (s *VMService) ChangeCD(ctx context.Context, id int, opts *VMChangeCDOptions) error {
	params := map[string]any{}
	if opts != nil {
		if opts.Drive != "" {
			params["drive"] = opts.Drive
		}
		if opts.Media != "" {
			params["media"] = opts.Media
		}
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionChangeCD,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to change CD on VM %d: %w", id, err)
	}
	return nil
}

// VMChangeNetOptions contains options for changing the network attached to a VM.
type VMChangeNetOptions struct {
	// NIC is the NIC name or index to change.
	NIC string `json:"nic,omitempty"`
	// Network is the target network ID.
	Network int `json:"network,omitempty"`
}

// ChangeNet changes the network attached to a VM NIC.
func (s *VMService) ChangeNet(ctx context.Context, id int, opts *VMChangeNetOptions) error {
	params := map[string]any{}
	if opts != nil {
		if opts.NIC != "" {
			params["nic"] = opts.NIC
		}
		if opts.Network > 0 {
			params["network"] = opts.Network
		}
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionChangeNet,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to change network on VM %d: %w", id, err)
	}
	return nil
}

// VMPasteOptions contains options for pasting text to a VM console.
type VMPasteOptions struct {
	// Text is the text to paste.
	Text string `json:"text"`
}

// Paste sends text to a VM's console.
func (s *VMService) Paste(ctx context.Context, id int, opts *VMPasteOptions) error {
	params := map[string]any{}
	if opts != nil {
		params["text"] = opts.Text
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionPaste,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to paste to VM %d: %w", id, err)
	}
	return nil
}

// VMRestoreOptions contains options for restoring a VM from a snapshot.
type VMRestoreOptions struct {
	// Snapshot is the snapshot reference to restore from.
	Snapshot string `json:"snapshot,omitempty"`
	// Name is the name for the restored VM (optional).
	Name string `json:"name,omitempty"`
	// PreserveMACs indicates whether to preserve MAC addresses.
	PreserveMACs bool `json:"preserve_macs,omitempty"`
}

// Restore restores a VM from a snapshot.
func (s *VMService) Restore(ctx context.Context, id int, opts *VMRestoreOptions) error {
	params := map[string]any{}
	if opts != nil {
		if opts.Snapshot != "" {
			params["snapshot"] = opts.Snapshot
		}
		if opts.Name != "" {
			params["name"] = opts.Name
		}
		params["preserve_macs"] = opts.PreserveMACs
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionRestore,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to restore VM %d: %w", id, err)
	}
	return nil
}

// RecoverCloudSnapshot recovers a VM from a cloud or system snapshot.
func (s *VMService) RecoverCloudSnapshot(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionRecoverCloudSnap,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to recover cloud snapshot for VM %d: %w", id, err)
	}
	return nil
}

// VMExecuteOptions contains options for executing a command on a VM via guest agent.
type VMExecuteOptions struct {
	// Command is the command to execute.
	Command string `json:"command"`
	// Args are the command arguments.
	Args []string `json:"args,omitempty"`
}

// Execute runs a command on a VM via the QEMU guest agent.
func (s *VMService) Execute(ctx context.Context, id int, opts *VMExecuteOptions) error {
	params := map[string]any{}
	if opts != nil {
		params["command"] = opts.Command
		if len(opts.Args) > 0 {
			params["args"] = opts.Args
		}
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionExecute,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to execute command on VM %d: %w", id, err)
	}
	return nil
}

// FSyncStrict performs a strict filesystem sync on a VM.
func (s *VMService) FSyncStrict(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionFSyncStrict,
		Params: vmActionParams{},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to fsync VM %d: %w", id, err)
	}
	return nil
}

// VMEraseDriveOptions contains options for erasing a drive on a VM.
type VMEraseDriveOptions struct {
	// Drive is the drive name or index to erase.
	Drive string `json:"drive,omitempty"`
}

// EraseDrive erases a drive on a VM.
func (s *VMService) EraseDrive(ctx context.Context, id int, opts *VMEraseDriveOptions) error {
	params := map[string]any{}
	if opts != nil {
		if opts.Drive != "" {
			params["drive"] = opts.Drive
		}
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionEraseDrive,
		Params: params,
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to erase drive on VM %d: %w", id, err)
	}
	return nil
}
