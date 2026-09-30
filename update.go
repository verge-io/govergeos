package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

const (
	// Check posts refresh. update_actions has no "check" value.
	updateActionRefresh  = "refresh"
	updateActionDownload = "download"
	updateActionInstall  = "install"
	// UpdateAll posts all, which downloads, installs, and starts the rolling reboot.
	updateActionAll = "all"
)

// updateActionRequest is the body for POST /update_actions.
// Force is sent only for the all action. A false value is explicit.
type updateActionRequest struct {
	Source int    `json:"source"`
	Action string `json:"action"`
	Force  *bool  `json:"force,omitempty"`
}

// UpdateSettingsService handles the update settings singleton (key=1)
// and the system-wide update actions.
type UpdateSettingsService struct {
	client *Client
}

// Get retrieves the update settings (singleton, key=1).
// The BranchName field is populated via computed field alias (branch#name).
func (s *UpdateSettingsService) Get(ctx context.Context) (*UpdateSettings, error) {
	params := url.Values{}
	params.Set("fields", updateSettingsGetFields)

	var settings UpdateSettings
	if err := s.client.get(ctx, "/update_settings/1", params, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// Check asks the configured update source for new packages.
// The API action is refresh. This does not download, install, or reboot.
func (s *UpdateSettingsService) Check(ctx context.Context) error {
	return s.postUpdateAction(ctx, updateActionRefresh, "check for updates", nil)
}

// Download downloads packages the configured source offers.
// This does not install them or reboot nodes.
func (s *UpdateSettingsService) Download(ctx context.Context) error {
	return s.postUpdateAction(ctx, updateActionDownload, "download updates", nil)
}

// Install installs packages that have already been downloaded.
// Nodes keep running. A reboot is still required to bring the update into service.
func (s *UpdateSettingsService) Install(ctx context.Context) error {
	return s.postUpdateAction(ctx, updateActionInstall, "install updates", nil)
}

// UpdateAll downloads available packages, installs them, and starts the
// platform rolling reboot. Nodes restart one at a time after workloads migrate.
// force allows a node to reboot workloads that cannot be migrated.
// The API action is all, and force is always sent.
func (s *UpdateSettingsService) UpdateAll(ctx context.Context, force bool) error {
	return s.postUpdateAction(ctx, updateActionAll, "apply updates", &force)
}

// postUpdateAction reads the configured source and posts one update_actions row.
func (s *UpdateSettingsService) postUpdateAction(ctx context.Context, action, what string, force *bool) error {
	settings, err := s.Get(ctx)
	if err != nil {
		return err
	}
	if settings.Source == 0 {
		return &ValidationError{Field: "source", Message: "no update source configured in settings"}
	}

	body := updateActionRequest{
		Source: settings.Source,
		Action: action,
		Force:  force,
	}
	if err := s.client.post(ctx, "/update_actions", body, nil); err != nil {
		return fmt.Errorf("vergeos: failed to %s: %w", what, err)
	}
	return nil
}

// UpdateBranchService handles update branch read operations.
type UpdateBranchService struct {
	client *Client
}

// List returns all update branches.
func (s *UpdateBranchService) List(ctx context.Context, opts ...ListOption) ([]UpdateBranch, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = updateBranchListFields
	}

	params := options.toQueryParams()

	var branches []UpdateBranch
	if err := s.client.get(ctx, "/update_branches", params, &branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// Get retrieves a specific update branch by key.
func (s *UpdateBranchService) Get(ctx context.Context, id int) (*UpdateBranch, error) {
	params := url.Values{}
	params.Set("fields", updateBranchListFields)

	var branch UpdateBranch
	endpoint := fmt.Sprintf("/update_branches/%d", id)
	if err := s.client.get(ctx, endpoint, params, &branch); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "UpdateBranch", ID: id}
		}
		return nil, err
	}

	return &branch, nil
}

// UpdateSourcePackageService handles update source package read operations.
type UpdateSourcePackageService struct {
	client *Client
}

// List returns all update source packages, with optional filtering.
func (s *UpdateSourcePackageService) List(ctx context.Context, opts ...ListOption) ([]UpdateSourcePackage, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = updateSourcePackageListFields
	}

	params := options.toQueryParams()

	var packages []UpdateSourcePackage
	if err := s.client.get(ctx, "/update_source_packages", params, &packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// ListByBranchAndSource returns packages for a specific branch and source.
func (s *UpdateSourcePackageService) ListByBranchAndSource(ctx context.Context, branch, source int) ([]UpdateSourcePackage, error) {
	return s.List(ctx, WithFilter(fmt.Sprintf("branch eq %d and source eq %d", branch, source)))
}

// Get retrieves a specific update source package by key.
func (s *UpdateSourcePackageService) Get(ctx context.Context, id int) (*UpdateSourcePackage, error) {
	params := url.Values{}
	params.Set("fields", updateSourcePackageListFields)

	var pkg UpdateSourcePackage
	endpoint := fmt.Sprintf("/update_source_packages/%d", id)
	if err := s.client.get(ctx, endpoint, params, &pkg); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "UpdateSourcePackage", ID: id}
		}
		return nil, err
	}

	return &pkg, nil
}
