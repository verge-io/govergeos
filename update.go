package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// UpdateSettingsService handles update settings read operations.
// Update settings is a singleton resource (key=1).
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
