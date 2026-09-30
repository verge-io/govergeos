package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

const (
	// Power action constants
	vmActionPowerOn  = "poweron"
	vmActionPowerOff = "poweroff"
	vmActionReset    = "reset"
	vmActionKill     = "kill"
	vmActionClone    = "clone"
	// vmActionSnapshot asks the guest agent to freeze filesystems. It does
	// not insert a machine_snapshots row; VMSnapshots.Create does that.
	vmActionSnapshot = "quiesce_snapshot"

	// Default VM power-wait budget: 30 polls, 5 seconds apart.
	// WithPowerWait and VMPowerOffOptions replace this for one client or one call.
	defaultPowerWaitTimeout  = 150 * time.Second
	defaultPowerWaitInterval = 5 * time.Second

	// vmSnapshotDefaultRetention is the snapshot lifetime, in seconds, used
	// when VMSnapshotOptions.Retention is not set (24 hours).
	vmSnapshotDefaultRetention = 86400
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

// GetByName returns the VM named name.
//
// The vms table stores snapshots in the same rows as VMs. This lookup
// adds is_snapshot eq false, so a snapshot is not returned. Look up a
// snapshot with VMSnapshots.GetByName.
// Returns NotFoundError when nothing matches, and AmbiguousNameError when
// more than one VM matches.
func (s *VMService) GetByName(ctx context.Context, name string) (*VM, error) {
	vms, err := s.List(ctx, WithFilter(fmt.Sprintf("is_snapshot eq false and name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(vms) == 0 {
		return nil, &NotFoundError{Resource: "VM", ID: name}
	}
	if err := requireUniqueName("VM", name, vms, func(vm VM) any { return vm.ID }); err != nil {
		return nil, err
	}
	if err := requireExactName("VM", name, vms[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, vms[0].ID.Int())
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VM", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
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

	if err := s.postVMPowerAction(ctx, id, vmActionPowerOn); err != nil {
		return fmt.Errorf("vergeos: failed to power on VM %d: %w", id, err)
	}

	// Wait for VM to start
	timeout, interval := s.client.vmPowerWait()
	return s.waitForPowerState(ctx, id, true, timeout, interval)
}

// PowerOff asks the guest to shut down and waits until the VM stops.
//
// It sends the poweroff action. Kill sends kill, which stops the VM
// immediately. GuestShutdown sends poweroff and returns without waiting.
//
// The wait uses the client power-wait timeout and poll interval (150
// seconds and 5 seconds, unless WithPowerWait changed them).
// PowerOffWithOptions overrides that wait for one call. A context
// deadline can still end the wait sooner.
func (s *VMService) PowerOff(ctx context.Context, id int) error {
	return s.PowerOffWithOptions(ctx, id, nil)
}

// VMPowerOffOptions configures one graceful VM shutdown.
type VMPowerOffOptions struct {
	// Timeout is how long to wait for the VM to stop after poweroff.
	// Zero uses the client default (150 seconds, or the WithPowerWait
	// timeout). A longer value extends the wait. The context can still
	// end it sooner.
	Timeout time.Duration

	// PollInterval is how often to read the VM while waiting.
	// Zero uses the client default (5 seconds, or the WithPowerWait
	// interval).
	PollInterval time.Duration

	// ForceAfterTimeout sends kill when the VM is still running at
	// Timeout, then waits again for the same Timeout. A cancelled
	// context does not escalate to kill.
	ForceAfterTimeout bool
}

// PowerOffWithOptions asks the guest to shut down, waits for it to stop,
// and applies opts. A nil opts value is the same as PowerOff.
func (s *VMService) PowerOffWithOptions(ctx context.Context, id int, opts *VMPowerOffOptions) error {
	var requestedTimeout, requestedInterval time.Duration
	force := false
	if opts != nil {
		requestedTimeout = opts.Timeout
		requestedInterval = opts.PollInterval
		force = opts.ForceAfterTimeout
	}
	timeout, interval, err := s.resolvePowerWait(requestedTimeout, requestedInterval)
	if err != nil {
		return err
	}

	vm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if !vm.PowerState {
		return nil
	}

	if err := s.postVMPowerAction(ctx, id, vmActionPowerOff); err != nil {
		return fmt.Errorf("vergeos: failed to power off VM %d: %w", id, err)
	}

	err = s.waitForPowerState(ctx, id, false, timeout, interval)
	if err == nil || !force || !IsTimeoutError(err) {
		return err
	}

	// The guest did not stop in time. Re-read in case it stopped
	// between the deadline and this escalation.
	vm, err = s.Get(ctx, id)
	if err != nil {
		return err
	}
	if !vm.PowerState {
		return nil
	}

	if err := s.postVMPowerAction(ctx, id, vmActionKill); err != nil {
		return fmt.Errorf("vergeos: failed to kill VM %d after power off timed out: %w", id, err)
	}
	if err := s.waitForPowerState(ctx, id, false, timeout, interval); err != nil {
		if IsTimeoutError(err) {
			return &TimeoutError{Resource: "VM", ID: id, Action: "become stopped after kill"}
		}
		return err
	}
	return nil
}

// Kill powers the VM off immediately and waits until it stops.
//
// It sends the kill action, the same hard stop NetworkService.Kill and
// TenantNodeService.Kill send. PowerOff is the guest shutdown. A VM that
// is already stopped is left alone.
//
// The wait uses the client power-wait settings. WithPowerWait changes
// the timeout and poll interval. A context deadline can end the wait sooner.
func (s *VMService) Kill(ctx context.Context, id int) error {
	timeout, interval := s.client.vmPowerWait()

	vm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if !vm.PowerState {
		return nil
	}

	if err := s.postVMPowerAction(ctx, id, vmActionKill); err != nil {
		return fmt.Errorf("vergeos: failed to kill VM %d: %w", id, err)
	}
	return s.waitForPowerState(ctx, id, false, timeout, interval)
}

// resolvePowerWait turns call-level durations into the wait budget.
// Zero selects the client default. Negative values are rejected.
func (s *VMService) resolvePowerWait(timeout, interval time.Duration) (time.Duration, time.Duration, error) {
	if timeout < 0 {
		return 0, 0, &ValidationError{Field: "timeout", Message: "timeout must be >= 0"}
	}
	if interval < 0 {
		return 0, 0, &ValidationError{Field: "poll_interval", Message: "poll_interval must be >= 0"}
	}
	defTimeout, defInterval := s.client.vmPowerWait()
	if timeout == 0 {
		timeout = defTimeout
	}
	if interval == 0 {
		interval = defInterval
	}
	return timeout, interval, nil
}

// postVMPowerAction posts a VM power action with empty params.
func (s *VMService) postVMPowerAction(ctx context.Context, id int, action string) error {
	return s.client.post(ctx, "/vm_actions", vmAction{
		VM:     id,
		Action: action,
		Params: vmActionParams{},
	}, nil)
}

// waitForPowerState waits for a VM to reach the desired power state.
// timeout is the wall-clock budget. interval is the pause between reads.
// The first read happens immediately. A cancelled context ends the wait
// with ctx.Err() and does not report TimeoutError.
func (s *VMService) waitForPowerState(ctx context.Context, id int, desiredState bool, timeout, interval time.Duration) error {
	stateStr := "stopped"
	if desiredState {
		stateStr = "running"
	}

	deadline := time.Now().Add(timeout)
	for {
		vm, err := s.Get(ctx, id)
		if err != nil {
			return err
		}
		if vm.PowerState == desiredState {
			return nil
		}
		// A caller that cancelled wants that error, including when the
		// budget is also exhausted. ForceAfterTimeout keys off TimeoutError.
		if err := ctx.Err(); err != nil {
			return err
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return &TimeoutError{Resource: "VM", ID: id, Action: "become " + stateStr}
		}

		sleep := interval
		if sleep > remaining {
			sleep = remaining
		}
		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
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

// GuestReboot asks the guest OS to reboot cleanly.
// VergeOS accepts this as a reset action with params.graceful set.
func (s *VMService) GuestReboot(ctx context.Context, id int) error {
	action := vmAction{
		VM:     id,
		Action: vmActionReset,
		Params: vmActionParams{Graceful: true},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to guest reboot VM %d: %w", id, err)
	}
	return nil
}

// GuestShutdown sends a graceful shutdown request to the guest OS via ACPI
// (poweroff action) and returns without waiting. PowerOff sends the same
// action and waits until the VM stops.
func (s *VMService) GuestShutdown(ctx context.Context, id int) error {
	if err := s.postVMPowerAction(ctx, id, vmActionPowerOff); err != nil {
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
	// Name is the snapshot name. When empty, a name of the form
	// snapshot-YYYYMMDD-HHMMSS (UTC) is generated.
	Name string
	// Retention is how long to keep the snapshot, in seconds.
	// Zero or negative uses vmSnapshotDefaultRetention (24 hours). The
	// duration is sent as an absolute expires timestamp. machine_snapshots
	// ignores a retention field.
	Retention int
	// Quiesce freezes guest filesystems while the snapshot is taken.
	// It requires a running guest agent in the VM. When false, no quiesce
	// request is sent and the snapshot is still created.
	Quiesce bool
}

// Snapshot creates a VM snapshot and returns it.
//
// id is the VM $key used by Get and the power methods. The row is inserted
// with VMSnapshots.Create (POST /machine_snapshots). Create resolves id to
// the VM's machine key, which is what machine_snapshots stores, along with
// a name and an expires timestamp of now + Retention. The quiesce_snapshot
// VM action does not create that row.
//
// When Quiesce is true, the create request sets quiesce and the
// quiesce_snapshot action is sent afterward so the guest agent can freeze
// filesystems. The action is not sent when Quiesce is false, which is what
// lets a VM without a guest agent still get a snapshot. The action requires
// a guest agent. If the row is created but that action fails, Snapshot
// returns the snapshot together with the error.
func (s *VMService) Snapshot(ctx context.Context, id int, opts *VMSnapshotOptions) (*VMSnapshot, error) {
	retention := vmSnapshotDefaultRetention
	name := ""
	quiesce := false
	if opts != nil {
		if opts.Retention > 0 {
			retention = opts.Retention
		}
		name = opts.Name
		quiesce = opts.Quiesce
	}
	if name == "" {
		name = "snapshot-" + time.Now().UTC().Format("20060102-150405")
	}

	expires := time.Now().Unix() + int64(retention)
	req := &VMSnapshotCreateRequest{
		VM:          id,
		Name:        name,
		ExpiresType: "date",
		Expires:     &expires,
	}
	if quiesce {
		q := true
		req.Quiesce = &q
	}

	snap, err := s.client.VMSnapshots.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("vergeos: failed to snapshot VM %d: %w", id, err)
	}
	if !quiesce {
		return snap, nil
	}

	action := struct {
		VM     int            `json:"vm"`
		Action string         `json:"action"`
		Params map[string]any `json:"params"`
	}{
		VM:     id,
		Action: vmActionSnapshot,
		Params: map[string]any{"quiesce": true},
	}
	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return snap, fmt.Errorf("vergeos: snapshot %d created for VM %d but quiesce failed (guest agent required): %w", int(snap.Key), id, err)
	}
	return snap, nil
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

// GetGuestAgentInfo returns guest agent information for a VM, including
// network interfaces and IP addresses reported by the in-guest qemu-guest-agent.
// Returns nil (with a nil error) when the guest agent has not reported — for
// example, when the VM is powered off, the agent is not installed, or the
// agent has not yet delivered an update.
func (s *VMService) GetGuestAgentInfo(ctx context.Context, id int) (*GuestInfo, error) {
	params := url.Values{}
	params.Set("fields", "dashboard")

	var resp vmDashboardResponse
	endpoint := fmt.Sprintf("/vms/%d", id)
	if err := s.client.get(ctx, endpoint, params, &resp); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VM", ID: id}
		}
		return nil, err
	}

	info := resp.Machine.Status.AgentGuestInfo
	// The API sends agent_guest_info: [] when the agent has not reported;
	// GuestInfo.UnmarshalJSON decodes that to a zero value. Normalize to nil
	// so callers get a single "not reported" signal.
	if info != nil && info.OSInfo == nil && info.Network == nil && info.FSInfo == nil &&
		info.MemInfo == nil && info.Hostname == "" && info.LastRefresh == 0 {
		return nil, nil
	}
	return info, nil
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
