package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

// VMImportService creates, reads, and deletes VM import jobs.
type VMImportService struct {
	client *Client
}

// List returns VM imports.
func (s *VMImportService) List(ctx context.Context, opts ...ListOption) ([]VMImport, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vmImportFields
	}
	var rows []VMImport
	if err := s.client.get(ctx, "/vm_imports", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// Get returns the import with this hex key.
func (s *VMImportService) Get(ctx context.Context, id string) (*VMImport, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	return getHexResource[VMImport](ctx, s.client, "/vm_imports", "VMImport", id, vmImportFields, func(row VMImport) (string, string) {
		return row.Key, row.ID
	})
}

// GetByName returns the import with this exact name.
// Two imports with the same name is an AmbiguousNameError. VergeOS allows that.
// Use List to see every row, and DeleteByName to remove the finished ones.
func (s *VMImportService) GetByName(ctx context.Context, name string) (*VMImport, error) {
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	rows, err := s.listByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "VMImport", ID: name}
	}
	if err := requireUniqueName("VMImport", name, rows, func(row VMImport) any { return importKey(row) }); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, importKey(rows[0]))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VMImport", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create imports a VM from a media-catalog file, a URL, a NAS volume path, or a shared object.
//
// A URL is stored as a media-catalog file first. OVA, OVF, and disk images
// (qcow2, vmdk, vhd, raw) use the same call. Importing defaults to true.
//
// When a non-snapshot VM with the requested name already exists, Create
// returns that import instead of posting another one. AlreadyExisted is set.
// A second post would be rejected because the VM name is taken. When more
// than one import row has the name, the returned Key is empty: Create will
// not pick one. DeleteByName removes the finished rows.
func (s *VMImportService) Create(ctx context.Context, req *VMImportCreateRequest) (*VMImport, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if err := validateImportSource(req); err != nil {
		return nil, err
	}

	existing, err := s.existingVM(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.adoptExisting(ctx, existing)
	}

	body, err := s.importBody(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vm_imports", body, &resp); err != nil {
		if IsConflictError(err) {
			existing, lookupErr := s.existingVM(ctx, req.Name)
			if lookupErr != nil {
				return nil, lookupErr
			}
			if existing != nil {
				return s.adoptExisting(ctx, existing)
			}
		}
		return nil, err
	}
	key, err := getStringKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, key)
}

// Update changes options on an import row and returns the row.
func (s *VMImportService) Update(ctx context.Context, id string, req *VMImportUpdateRequest) (*VMImport, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	endpoint := "/vm_imports/" + url.PathEscape(id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMImport", ID: id}
		}
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete removes one import row by key. The VM, if the import created one, is left in place.
func (s *VMImportService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	endpoint := "/vm_imports/" + url.PathEscape(id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMImport", ID: id}
		}
		return err
	}
	return nil
}

// DeleteByName removes every finished import row with this name.
//
// No matching row is success. One row is deleted even when that import is
// still running. Two or more rows are all deleted when every one of them
// has finished. When more than one row shares the name and any of them is
// still importing, DeleteByName returns VMImportInProgressError and deletes
// nothing, so a cleanup cannot cancel a live import that happens to share
// the name. Delete removes one row by key, including a live one.
func (s *VMImportService) DeleteByName(ctx context.Context, name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	rows, err := s.listByName(ctx, name)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	if len(rows) > 1 {
		var live []string
		for _, row := range rows {
			if vmImportInProgress(row) {
				live = append(live, importKey(row))
			}
		}
		if len(live) > 0 {
			return &VMImportInProgressError{Name: name, Keys: live}
		}
	}
	for _, row := range rows {
		if err := s.Delete(ctx, importKey(row)); err != nil {
			return err
		}
	}
	return nil
}

// Start begins an import that was created with Importing set to false.
func (s *VMImportService) Start(ctx context.Context, id string) error {
	return s.importAction(ctx, id, "import")
}

// Abort stops an import that is still running.
func (s *VMImportService) Abort(ctx context.Context, id string) error {
	return s.importAction(ctx, id, "abort")
}

// Wait polls until the import finishes.
//
// Status error, status aborted, the aborted flag, and a non-zero failed
// drive count are returned immediately as VMImportFailedError. The error
// includes status_info and the error log lines. A failed import does not
// consume the rest of Timeout. Status complete and warning are success.
//
// Timeout defaults to 10 minutes and the poll interval to 5 seconds.
// A cancelled context returns ctx.Err().
func (s *VMImportService) Wait(ctx context.Context, id string, opts *VMImportWaitOptions) (*VMImport, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	var requestedTimeout, requestedInterval time.Duration
	if opts != nil {
		requestedTimeout = opts.Timeout
		requestedInterval = opts.PollInterval
	}
	timeout, interval, err := resolveWait(requestedTimeout, requestedInterval, defaultVMImportWaitTimeout, defaultVMImportWaitInterval)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	for {
		imp, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if imp.failed() {
			return nil, s.failure(ctx, imp)
		}
		if imp.succeeded() {
			return imp, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, &TimeoutError{Resource: "VMImport", Key: id, Action: "complete"}
		}
		sleep := interval
		if sleep > remaining {
			sleep = remaining
		}
		if err := waitInterval(ctx, sleep); err != nil {
			return nil, err
		}
	}
}

// Logs returns the log lines for one import, newest behavior depending on the caller's sort.
func (s *VMImportService) Logs(ctx context.Context, importID string, opts ...ListOption) ([]VMImportLog, error) {
	return s.client.VMImportLogs.ListByImport(ctx, importID, opts...)
}

func (s *VMImportService) importAction(ctx context.Context, id, action string) error {
	if id == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	endpoint := fmt.Sprintf("/vm_imports/%s?action=%s", url.PathEscape(id), url.QueryEscape(action))
	if err := s.client.put(ctx, endpoint, struct{}{}, nil); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMImport", ID: id}
		}
		return err
	}
	return nil
}

func (s *VMImportService) listByName(ctx context.Context, name string) ([]VMImport, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	exact := make([]VMImport, 0, len(rows))
	for _, row := range rows {
		if row.Name == name {
			exact = append(exact, row)
		}
	}
	return exact, nil
}

func (s *VMImportService) existingVM(ctx context.Context, name string) (*VM, error) {
	vm, err := s.client.VMs.GetByName(ctx, name)
	if IsNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (s *VMImportService) adoptExisting(ctx context.Context, vm *VM) (*VMImport, error) {
	rows, err := s.listByName(ctx, vm.Name)
	if err != nil {
		return nil, err
	}
	vmID := FlexInt(vm.Key)
	if len(rows) != 1 {
		return &VMImport{
			Name:           vm.Name,
			VM:             &vmID,
			Status:         VMImportStatusComplete,
			AlreadyExisted: true,
		}, nil
	}
	got, err := s.Get(ctx, importKey(rows[0]))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VMImport", vm.Name, got.Name, vm.Name); err != nil {
		return nil, err
	}
	got.AlreadyExisted = true
	if got.VM == nil {
		got.VM = &vmID
	}
	return got, nil
}

func (s *VMImportService) importBody(ctx context.Context, req *VMImportCreateRequest) (map[string]any, error) {
	importing := true
	if req.Importing != nil {
		importing = *req.Importing
	}
	body := map[string]any{
		"name":      req.Name,
		"importing": importing,
	}
	switch {
	case req.File != nil:
		body["file"] = *req.File
	case req.URL != "":
		file, err := s.fileFromURL(ctx, req)
		if err != nil {
			return nil, err
		}
		body["file"] = int(file.Key)
	case req.Volume != "":
		body["volume"] = req.Volume
		if req.VolumePath != "" {
			body["volume_path"] = req.VolumePath
		}
	case req.SharedObject != nil:
		body["shared_object"] = *req.SharedObject
	}
	if req.PreserveMACs != nil {
		body["preserve_macs"] = *req.PreserveMACs
	}
	if req.PreserveDriveFormat != nil {
		body["preserve_drive_format"] = *req.PreserveDriveFormat
	}
	if req.PreferredTier != "" {
		body["preferred_tier"] = req.PreferredTier
	}
	if req.NoOpticalDrives != nil {
		body["no_optical_drives"] = *req.NoOpticalDrives
	}
	if req.OverrideDriveInterface != "" {
		body["override_drive_interface"] = req.OverrideDriveInterface
	}
	if req.OverrideNICInterface != "" {
		body["override_nic_interface"] = req.OverrideNICInterface
	}
	if req.CleanupOnDelete != nil {
		body["cleanup_on_delete"] = *req.CleanupOnDelete
	}
	return body, nil
}

func (s *VMImportService) fileFromURL(ctx context.Context, req *VMImportCreateRequest) (*File, error) {
	name := req.FileName
	if name == "" {
		var err error
		name, err = fileNameFromURL(req.URL)
		if err != nil {
			return nil, err
		}
	}
	fileReq := &FileCreateRequest{
		Name:          name,
		Description:   "Downloaded for VM import " + req.Name,
		URL:           req.URL,
		PreferredTier: req.PreferredTier,
	}
	return s.client.Files.Create(ctx, fileReq)
}

func (s *VMImportService) failure(ctx context.Context, imp *VMImport) error {
	status := imp.Status
	if status == "" && imp.Aborted {
		status = VMImportStatusAborted
	}
	if status == "" && imp.FailedDriveCount > 0 {
		status = VMImportStatusError
	}
	return &VMImportFailedError{
		Key:          importKey(*imp),
		Name:         imp.Name,
		Status:       status,
		StatusInfo:   imp.StatusInfo,
		FailedDrives: imp.FailedDriveCount,
		LogLines:     s.errorLogLines(ctx, importKey(*imp)),
	}
}

func (s *VMImportService) errorLogLines(ctx context.Context, importID string) []string {
	if importID == "" {
		return nil
	}
	logs, err := s.Logs(ctx, importID, WithFilter("((level eq 'error') or (level eq 'critical'))"), WithLimit(10))
	if err != nil {
		return nil
	}
	lines := make([]string, 0, len(logs))
	for _, log := range logs {
		text := strings.TrimSpace(log.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	return lines
}

func validateImportSource(req *VMImportCreateRequest) error {
	sources := 0
	if req.File != nil {
		sources++
	}
	if req.URL != "" {
		sources++
	}
	if req.Volume != "" {
		sources++
	}
	if req.SharedObject != nil {
		sources++
	}
	if sources != 1 {
		return &ValidationError{Message: "set one of file, url, volume, or shared_object"}
	}
	if req.File != nil && *req.File <= 0 {
		return &ValidationError{Field: "file", Message: "file must be a positive id"}
	}
	if req.SharedObject != nil && *req.SharedObject <= 0 {
		return &ValidationError{Field: "shared_object", Message: "shared_object must be a positive id"}
	}
	if req.VolumePath != "" && req.Volume == "" {
		return &ValidationError{Field: "volume_path", Message: "volume_path requires volume"}
	}
	if req.URL != "" && req.FileName == "" {
		if _, err := fileNameFromURL(req.URL); err != nil {
			return err
		}
	}
	if req.URL != "" {
		if err := validateImportURL(req.URL); err != nil {
			return err
		}
	}
	return nil
}

func validateImportURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return &ValidationError{Field: "url", Message: "url must be an absolute http or https URL"}
	}
	return nil
}

func fileNameFromURL(raw string) (string, error) {
	if err := validateImportURL(raw); err != nil {
		return "", err
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", &ValidationError{Field: "url", Message: "url must be an absolute http or https URL"}
	}
	if strings.HasSuffix(u.Path, "/") || u.Path == "" {
		return "", &ValidationError{Field: "url", Message: "url path must name a file, or set file_name"}
	}
	name, err := url.PathUnescape(path.Base(u.Path))
	if err != nil || name == "" || name == "." || name == "/" || strings.Contains(name, "/") {
		return "", &ValidationError{Field: "url", Message: "url path must name a file, or set file_name"}
	}
	return name, nil
}

func importKey(row VMImport) string {
	if row.Key != "" {
		return row.Key
	}
	return row.ID
}

func (v *VMImport) failed() bool {
	if v == nil {
		return false
	}
	if v.Aborted {
		return true
	}
	switch strings.ToLower(v.Status) {
	case VMImportStatusError, VMImportStatusAborted:
		return true
	case VMImportStatusComplete, VMImportStatusWarning:
		return false
	}
	return v.FailedDriveCount > 0
}

func (v *VMImport) succeeded() bool {
	if v == nil || v.failed() {
		return false
	}
	switch strings.ToLower(v.Status) {
	case VMImportStatusComplete, VMImportStatusWarning:
		return true
	}
	return false
}

func vmImportInProgress(v VMImport) bool {
	if v.Aborted {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(v.Status)) {
	case VMImportStatusComplete, VMImportStatusError, VMImportStatusAborted, VMImportStatusWarning:
		return false
	case VMImportStatusImporting, VMImportStatusInitializing, "pending", "queued", "running", "in_progress", "processing", "new":
		return true
	}
	return v.Importing
}

// VMImportLogService reads vm_import_logs.
type VMImportLogService struct {
	client *Client
}

// List returns import log lines.
func (s *VMImportLogService) List(ctx context.Context, opts ...ListOption) ([]VMImportLog, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vmImportLogFields
	}
	var rows []VMImportLog
	if err := s.client.get(ctx, "/vm_import_logs", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByImport returns the log lines for one import.
func (s *VMImportLogService) ListByImport(ctx context.Context, importID string, opts ...ListOption) ([]VMImportLog, error) {
	if importID == "" {
		return nil, &ValidationError{Field: "import_id", Message: "import_id is required"}
	}
	opts = append([]ListOption{WithFilter(fmt.Sprintf("vm_import eq '%s'", escapeFilterValue(importID)))}, opts...)
	return s.List(ctx, opts...)
}

// Get returns one log line by row id.
func (s *VMImportLogService) Get(ctx context.Context, id int) (*VMImportLog, error) {
	if id <= 0 {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	params := url.Values{}
	params.Set("fields", vmImportLogFields)
	var row VMImportLog
	endpoint := fmt.Sprintf("/vm_import_logs/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMImportLog", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

func resolveWait(timeout, interval, defTimeout, defInterval time.Duration) (time.Duration, time.Duration, error) {
	if timeout < 0 {
		return 0, 0, &ValidationError{Field: "timeout", Message: "timeout must be >= 0"}
	}
	if interval < 0 {
		return 0, 0, &ValidationError{Field: "poll_interval", Message: "poll_interval must be >= 0"}
	}
	if timeout == 0 {
		timeout = defTimeout
	}
	if interval == 0 {
		interval = defInterval
	}
	return timeout, interval, nil
}

func waitInterval(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
