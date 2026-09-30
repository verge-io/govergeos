package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

const (
	// Microsoft 2023 Secure Boot KEK support was introduced in VergeOS 26.1.5.
	ms2023KEKMinimumMajor = 26
	ms2023KEKMinimumMinor = 1
	ms2023KEKMinimumPatch = 5

	// Drive hotplug action
	vmActionHotplugDrive = "hotplugdrive"

	// Drive power state polling
	driveUnplugMaxRetries   = 5
	driveUnplugPollInterval = 5 * time.Second

	// Import status polling
	importMaxRetries   = 10
	importPollInterval = 5 * time.Second
)

// VMDriveService handles VM drive operations.
type VMDriveService struct {
	client *Client
}

func (s *VMDriveService) listFields() string {
	fields := driveListFields
	if isVersionAtLeast(s.client.serverVersion, ms2023KEKMinimumMajor, ms2023KEKMinimumMinor, ms2023KEKMinimumPatch) {
		fields += "," + driveMS2023KEKField
	}
	return fields
}

func (s *VMDriveService) getFields() string {
	fields := driveGetFields
	if isVersionAtLeast(s.client.serverVersion, ms2023KEKMinimumMajor, ms2023KEKMinimumMinor, ms2023KEKMinimumPatch) {
		fields += "," + driveMS2023KEKField
	}
	return fields
}

// List returns all drives for a VM.
// vmID is the VM $key (VM.Key). Drives are stored against the machine key,
// which is resolved before querying machine_drives.
func (s *VMDriveService) List(ctx context.Context, vmID int) ([]VMDrive, error) {
	machine, err := s.client.machineKeyForVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("fields", s.listFields())
	params.Set("filter", fmt.Sprintf("machine eq %d", machine))

	var drives []VMDrive
	if err := s.client.get(ctx, "/machine_drives", params, &drives); err != nil {
		return nil, err
	}

	// Convert bytes to GB
	for i := range drives {
		drives[i].SizeGB = drives[i].SizeBytes / bytesPerGB
	}

	return drives, nil
}

// ListAll returns all machine drives across all VMs.
func (s *VMDriveService) ListAll(ctx context.Context, opts ...ListOption) ([]VMDrive, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = s.listFields()
	}
	params := options.toQueryParams()

	var drives []VMDrive
	if err := s.client.get(ctx, "/machine_drives", params, &drives); err != nil {
		return nil, err
	}

	// Convert bytes to GB
	for i := range drives {
		drives[i].SizeGB = drives[i].SizeBytes / bytesPerGB
	}

	return drives, nil
}

// Get returns a single drive by ID.
func (s *VMDriveService) Get(ctx context.Context, driveID int) (*VMDrive, error) {
	params := url.Values{}
	params.Set("fields", s.getFields())

	var drive VMDrive
	endpoint := fmt.Sprintf("/machine_drives/%d", driveID)
	if err := s.client.get(ctx, endpoint, params, &drive); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMDrive", ID: driveID}
		}
		return nil, err
	}

	// Convert bytes to GB
	drive.SizeGB = drive.SizeBytes / bytesPerGB

	return &drive, nil
}

// GetByName returns a drive by name within a specific VM.
// vmID is the VM $key (VM.Key). Drive names are scoped to a machine
// (every VM can have a "disk0"), so the machine key is resolved first.
// Returns NotFoundError if no drive with the given name exists on the VM.
func (s *VMDriveService) GetByName(ctx context.Context, vmID int, name string) (*VMDrive, error) {
	machine, err := s.client.machineKeyForVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	drives, err := s.ListAll(ctx, WithFilter(fmt.Sprintf("machine eq %d and name eq '%s'", machine, escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(drives) == 0 {
		return nil, &NotFoundError{Resource: "VMDrive", ID: name}
	}
	if err := requireUniqueName("VMDrive", name, drives, func(d VMDrive) any { return d.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("VMDrive", name, drives[0].Name, name); err != nil {
		return nil, err
	}
	// Fetch full details (Get returns additional fields like power state and status).
	got, err := s.Get(ctx, drives[0].Key.Int())
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VMDrive", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a new drive and returns the created drive.
// vmID is the VM $key (VM.Key). It is resolved to the machine key stored on the drive.
// For import media, this method waits for the import to complete.
func (s *VMDriveService) Create(ctx context.Context, vmID int, req *VMDriveCreateRequest) (*VMDrive, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	machine, err := s.client.machineKeyForVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	req.Machine = machine

	// Convert GB to bytes
	if req.SizeGB > 0 {
		req.SizeBytes = req.SizeGB * bytesPerGB
	}

	// Set defaults
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/machine_drives", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created drive's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created drive
	drive, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// If this is an import, wait for it to complete
	if req.Media == "import" {
		drive, err = s.waitForImport(ctx, id, req.SizeGB)
		if err != nil {
			return nil, err
		}
	}

	return drive, nil
}

// waitForImport waits for a drive import to complete and handles resizing if needed.
func (s *VMDriveService) waitForImport(ctx context.Context, driveID int, targetSizeGB int64) (*VMDrive, error) {
	// Initial wait before first poll
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
	}

	for i := 0; i < importMaxRetries; i++ {
		drive, err := s.Get(ctx, driveID)
		if err != nil {
			return nil, err
		}

		// Check if import is complete
		if drive.Status != "importing" {
			// If target size differs from actual size, resize the drive
			if targetSizeGB > 0 && drive.SizeGB != targetSizeGB {
				targetSizeBytes := targetSizeGB * bytesPerGB
				updateReq := &VMDriveUpdateRequest{
					SizeBytes: &targetSizeBytes,
				}
				return s.Update(ctx, driveID, updateReq)
			}
			return drive, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(importPollInterval):
		}
	}

	return nil, &TimeoutError{Resource: "VMDrive", ID: driveID, Action: "complete import"}
}

// Update updates a drive and returns the updated drive.
func (s *VMDriveService) Update(ctx context.Context, driveID int, req *VMDriveUpdateRequest) (*VMDrive, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	// Convert GB to bytes if specified
	if req.SizeGB != nil && *req.SizeGB > 0 {
		sizeBytes := *req.SizeGB * bytesPerGB
		req.SizeBytes = &sizeBytes
	}

	endpoint := fmt.Sprintf("/machine_drives/%d", driveID)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMDrive", ID: driveID}
		}
		return nil, err
	}

	// Read back the updated drive
	return s.Get(ctx, driveID)
}

// ApplyUniversalVars applies the Microsoft 2023 Secure Boot keys to an EFI drive.
// VergeOS requires the drive to belong to an offline secure-boot VM.
func (s *VMDriveService) ApplyUniversalVars(ctx context.Context, driveID int) error {
	endpoint := fmt.Sprintf("/machine_drives/%d/apply_universal_vars", driveID)
	return s.client.post(ctx, endpoint, struct{}{}, nil)
}

// Delete deletes a drive.
// If the drive is currently attached (VM running), it will be hot-unplugged first.
func (s *VMDriveService) Delete(ctx context.Context, driveID int) error {
	// Get drive to check power state and get VM ID
	drive, err := s.Get(ctx, driveID)
	if err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMDrive", ID: driveID}
		}
		return err
	}

	// If drive is online, hot-unplug it first.
	// drive.Machine is a machine key; vm_actions wants the VM $key.
	if drive.PowerState != "" && drive.PowerState != "offline" {
		vmKey, err := s.client.vmKeyForMachine(ctx, drive.Machine, false)
		if err != nil {
			return fmt.Errorf("vergeos: failed to resolve VM for machine %d: %w", drive.Machine, err)
		}
		if err := s.hotUnplug(ctx, vmKey, driveID); err != nil {
			return fmt.Errorf("vergeos: failed to unplug drive before deletion: %w", err)
		}
	}

	endpoint := fmt.Sprintf("/machine_drives/%d", driveID)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMDrive", ID: driveID}
		}
		return err
	}

	return nil
}

// HotplugDrive hot-plugs a drive into a running VM, making it visible to the guest OS.
// vmID is the VM's $key (not the internal machine reference).
// The drive must already be assigned to the VM (machine field set) before calling this.
func (s *VMDriveService) HotplugDrive(ctx context.Context, vmID, driveID int) error {
	action := vmAction{
		VM:     vmID,
		Action: vmActionHotplugDrive,
		Params: vmActionParams{
			Device: fmt.Sprintf("%d", driveID),
		},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return err
	}

	// Wait for drive to come online
	for i := 0; i < driveUnplugMaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(driveUnplugPollInterval):
		}

		drive, err := s.Get(ctx, driveID)
		if err != nil {
			return err
		}

		if drive.PowerState == "online" {
			return nil
		}
	}

	return &TimeoutError{Resource: "VMDrive", ID: driveID, Action: "hotplug"}
}

// HotUnplugDrive hot-unplugs a drive from a running VM.
// vmID is the VM's $key (not the internal machine reference).
func (s *VMDriveService) HotUnplugDrive(ctx context.Context, vmID, driveID int) error {
	action := vmAction{
		VM:     vmID,
		Action: vmActionHotplugDrive,
		Params: vmActionParams{
			Device: fmt.Sprintf("%d", driveID),
			Unplug: true,
		},
	}

	if err := s.client.post(ctx, "/vm_actions", action, nil); err != nil {
		return err
	}

	// Wait for drive to be unplugged
	for i := 0; i < driveUnplugMaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(driveUnplugPollInterval):
		}

		drive, err := s.Get(ctx, driveID)
		if err != nil {
			if IsNotFoundError(err) {
				return nil // Drive was removed
			}
			return err
		}

		if drive.PowerState == "offline" || drive.PowerState == "" {
			return nil
		}
	}

	return &TimeoutError{Resource: "VMDrive", ID: driveID, Action: "unplug"}
}

// hotUnplug is the internal wrapper used by Delete.
func (s *VMDriveService) hotUnplug(ctx context.Context, vmID, driveID int) error {
	return s.HotUnplugDrive(ctx, vmID, driveID)
}
