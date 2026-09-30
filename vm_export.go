package vergeos

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// VMExportService exports VMs onto a NAS volume.
type VMExportService struct {
	client *Client
}

// List returns volume VM-export configurations.
func (s *VMExportService) List(ctx context.Context, opts ...ListOption) ([]VMExport, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vmExportFields
	}
	var rows []VMExport
	if err := s.client.get(ctx, "/volume_vm_exports", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// Get returns one export configuration by row id.
func (s *VMExportService) Get(ctx context.Context, id int) (*VMExport, error) {
	if id <= 0 {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	params := url.Values{}
	params.Set("fields", vmExportFields)
	var row VMExport
	endpoint := fmt.Sprintf("/volume_vm_exports/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMExport", ID: id}
		}
		return nil, err
	}
	if row.Key.Int() != 0 && row.Key.Int() != id {
		return nil, &NotFoundError{Resource: "VMExport", ID: id}
	}
	return &row, nil
}

// GetByVolume returns the export configuration for a NAS volume key.
// Two rows for the same volume is an AmbiguousNameError.
func (s *VMExportService) GetByVolume(ctx context.Context, volumeID string) (*VMExport, error) {
	if volumeID == "" {
		return nil, &ValidationError{Field: "volume", Message: "volume is required"}
	}
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("volume eq '%s'", escapeFilterValue(volumeID))))
	if err != nil {
		return nil, err
	}
	exact := make([]VMExport, 0, len(rows))
	for _, row := range rows {
		if row.Volume == volumeID {
			exact = append(exact, row)
		}
	}
	if len(exact) == 0 {
		return nil, &NotFoundError{Resource: "VMExport", ID: volumeID}
	}
	if err := requireUniqueName("VMExport", volumeID, exact, func(row VMExport) any { return row.Key.Int() }); err != nil {
		return nil, err
	}
	return s.Get(ctx, exact[0].Key.Int())
}

// Create creates the export configuration for a NAS volume.
func (s *VMExportService) Create(ctx context.Context, req *VMExportCreateRequest) (*VMExport, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Volume == "" {
		return nil, &ValidationError{Field: "volume", Message: "volume is required"}
	}
	if err := validateMaxExports(req.MaxExports); err != nil {
		return nil, err
	}
	body := map[string]any{"volume": req.Volume}
	if req.Quiesced != nil {
		body["quiesced"] = *req.Quiesced
	}
	if req.CreateCurrent != nil {
		body["create_current"] = *req.CreateCurrent
	}
	if req.MaxExports != nil {
		body["max_exports"] = *req.MaxExports
	}
	var resp apiResponse
	if err := s.client.post(ctx, "/volume_vm_exports", body, &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update changes an export configuration and returns the row.
func (s *VMExportService) Update(ctx context.Context, id int, req *VMExportUpdateRequest) (*VMExport, error) {
	if id <= 0 {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if err := validateMaxExports(req.MaxExports); err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("/volume_vm_exports/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VMExport", ID: id}
		}
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete removes the export configuration. Files already written to the volume are left in place.
func (s *VMExportService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	endpoint := fmt.Sprintf("/volume_vm_exports/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMExport", ID: id}
		}
		return err
	}
	return nil
}

// Run ensures the volume has an export configuration and starts an export.
// It returns when the platform has accepted the run. Call Wait to block
// until the run finishes, including a failure the platform has already recorded.
func (s *VMExportService) Run(ctx context.Context, req *VMExportRunRequest) (*VMExport, error) {
	if req == nil {
		return nil, &ValidationError{Message: "run request is required"}
	}
	if req.Volume == "" {
		return nil, &ValidationError{Field: "volume", Message: "volume is required"}
	}
	if err := validateMaxExports(req.MaxExports); err != nil {
		return nil, err
	}

	row, err := s.GetByVolume(ctx, req.Volume)
	if IsNotFoundError(err) {
		row, err = s.Create(ctx, &VMExportCreateRequest{
			Volume:        req.Volume,
			Quiesced:      req.Quiesced,
			CreateCurrent: req.CreateCurrent,
			MaxExports:    req.MaxExports,
		})
	} else if err == nil {
		row, err = s.applyExportSettings(ctx, row, req)
	}
	if err != nil {
		return nil, err
	}
	if err := s.Start(ctx, row.Key.Int(), &VMExportStart{Name: req.Name, VMs: req.VMs}); err != nil {
		return nil, err
	}
	return s.Get(ctx, row.Key.Int())
}

// Start starts an export on an existing configuration.
func (s *VMExportService) Start(ctx context.Context, id int, run *VMExportStart) error {
	return s.exportAction(ctx, id, "start_export", run)
}

// Stop stops an export that is still building.
func (s *VMExportService) Stop(ctx context.Context, id int) error {
	return s.exportAction(ctx, id, "stop_export", nil)
}

// Cleanup removes old export generations past MaxExports.
func (s *VMExportService) Cleanup(ctx context.Context, id int) error {
	return s.exportAction(ctx, id, "cleanup", nil)
}

// Wait polls until the export leaves the building and cleaning states.
//
// Status error is returned immediately as VMExportFailedError. After the
// run is idle, the newest statistics row is read, and a row that recorded
// errors is also VMExportFailedError. A failed export does not consume the
// rest of Timeout.
//
// Timeout defaults to 1 hour and the poll interval to 5 seconds.
// A cancelled context returns ctx.Err().
func (s *VMExportService) Wait(ctx context.Context, id int, opts *VMExportWaitOptions) (*VMExport, error) {
	if id <= 0 {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	var requestedTimeout, requestedInterval time.Duration
	if opts != nil {
		requestedTimeout = opts.Timeout
		requestedInterval = opts.PollInterval
	}
	timeout, interval, err := resolveWait(requestedTimeout, requestedInterval, defaultVMExportWaitTimeout, defaultVMExportWaitInterval)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	for {
		row, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if row.Status == VMExportStatusError {
			return nil, &VMExportFailedError{ID: id, Status: row.Status, StatusInfo: row.StatusInfo}
		}
		if row.Status != "" && row.Status != VMExportStatusBuilding && row.Status != VMExportStatusCleaning {
			return s.finishExport(ctx, row)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, &TimeoutError{Resource: "VMExport", ID: id, Action: "finish"}
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

// Stats returns the statistics rows for one export configuration.
func (s *VMExportService) Stats(ctx context.Context, id int, opts ...ListOption) ([]VMExportStat, error) {
	if id <= 0 {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	opts = append([]ListOption{WithFilter(fmt.Sprintf("volume_vm_exports eq %d", id))}, opts...)
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vmExportStatFields
	}
	var rows []VMExportStat
	if err := s.client.get(ctx, "/volume_vm_export_stats", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *VMExportService) finishExport(ctx context.Context, row *VMExport) (*VMExport, error) {
	stats, err := s.Stats(ctx, row.Key.Int())
	if err != nil {
		return nil, err
	}
	latest := newestExportStat(stats)
	if latest != nil && latest.Errors > 0 {
		return nil, &VMExportFailedError{
			ID:              row.Key.Int(),
			Status:          row.Status,
			StatusInfo:      row.StatusInfo,
			Errors:          latest.Errors,
			VirtualMachines: latest.VirtualMachines,
			FileName:        latest.FileName,
		}
	}
	return row, nil
}

func (s *VMExportService) applyExportSettings(ctx context.Context, row *VMExport, req *VMExportRunRequest) (*VMExport, error) {
	update := &VMExportUpdateRequest{}
	changed := false
	if req.Quiesced != nil && *req.Quiesced != row.Quiesced {
		update.Quiesced = req.Quiesced
		changed = true
	}
	if req.CreateCurrent != nil && *req.CreateCurrent != row.CreateCurrent {
		update.CreateCurrent = req.CreateCurrent
		changed = true
	}
	if req.MaxExports != nil && *req.MaxExports != row.MaxExports {
		update.MaxExports = req.MaxExports
		changed = true
	}
	if !changed {
		return row, nil
	}
	return s.Update(ctx, row.Key.Int(), update)
}

func (s *VMExportService) exportAction(ctx context.Context, id int, action string, run *VMExportStart) error {
	if id <= 0 {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	body := map[string]any{
		"volume_vm_export": id,
		"action":           action,
	}
	if run != nil && (run.Name != "" || len(run.VMs) > 0) {
		params := map[string]any{}
		if run.Name != "" {
			params["name"] = run.Name
		}
		if len(run.VMs) > 0 {
			params["vms"] = run.VMs
		}
		body["params"] = params
	}
	if err := s.client.post(ctx, "/volume_vm_export_actions", body, nil); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VMExport", ID: id}
		}
		return err
	}
	return nil
}

func newestExportStat(rows []VMExportStat) *VMExportStat {
	if len(rows) == 0 {
		return nil
	}
	best := 0
	for i := range rows {
		if rows[i].Timestamp >= rows[best].Timestamp {
			best = i
		}
	}
	return &rows[best]
}

func validateMaxExports(maxExports *int) error {
	if maxExports == nil {
		return nil
	}
	if *maxExports < 1 || *maxExports > 100 {
		return &ValidationError{Field: "max_exports", Message: "max_exports must be from 1 to 100"}
	}
	return nil
}
