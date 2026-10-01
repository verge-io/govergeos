package vergeos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// errSettingsShape is returned when an auth source settings document is
// not a JSON object. The raw payload is not included: it can contain a
// client secret.
var errSettingsShape = errors.New("settings must be a JSON object")

// AuthSourceService handles external authentication source operations.
type AuthSourceService struct {
	client *Client
}

// List returns auth sources, with optional filtering and pagination.
// The default field set does not include settings.
func (s *AuthSourceService) List(ctx context.Context, opts ...ListOption) ([]AuthSource, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = authSourceListFields
	}

	params := options.toQueryParams()

	var sources []AuthSource
	if err := s.client.get(ctx, "/auth_sources", params, &sources); err != nil {
		return nil, err
	}

	return sources, nil
}

// Get returns an auth source by ID, including settings other than
// client_secret.
func (s *AuthSourceService) Get(ctx context.Context, id int) (*AuthSource, error) {
	params := url.Values{}
	params.Set("fields", authSourceGetFields)

	var source AuthSource
	endpoint := fmt.Sprintf("/auth_sources/%d", id)
	if err := s.client.get(ctx, endpoint, params, &source); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "AuthSource", ID: id}
		}
		return nil, err
	}

	return &source, nil
}

// GetByName returns the auth source with this exact name.
func (s *AuthSourceService) GetByName(ctx context.Context, name string) (*AuthSource, error) {
	sources, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, &NotFoundError{Resource: "AuthSource", ID: name}
	}
	if err := requireUniqueName("AuthSource", name, sources, func(source AuthSource) any { return source.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("AuthSource", name, sources[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(sources[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("AuthSource", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates an auth source and returns it.
// client_secret in Settings is sent and is not present on the returned source.
func (s *AuthSourceService) Create(ctx context.Context, req *AuthSourceCreateRequest) (*AuthSource, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.Driver == "" {
		return nil, &ValidationError{Field: "driver", Message: "driver is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/auth_sources", req, &resp); err != nil {
		return nil, err
	}

	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Update updates an auth source and returns it.
//
// When Settings is non-nil, Update reads the stored settings and shallow-merges
// the caller's keys on top before sending. The API replaces the settings
// object, so a partial document would otherwise delete every omitted key,
// including client_secret. Omitted keys are kept, including keys the server
// added (such as debug). The read and the write are not atomic.
//
// A nil Settings leaves the stored document unchanged. An empty non-nil
// Settings still rewrites the stored document with its current keys.
// client_secret is not present on the returned source.
func (s *AuthSourceService) Update(ctx context.Context, id int, req *AuthSourceUpdateRequest) (*AuthSource, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	body := *req
	if req.Settings != nil {
		current, err := s.readSettings(ctx, id)
		if err != nil {
			return nil, err
		}
		body.Settings = mergeAuthSourceSettings(current, req.Settings)
	}

	endpoint := fmt.Sprintf("/auth_sources/%d", id)
	if err := s.client.put(ctx, endpoint, &body, nil); err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete deletes an auth source.
// The API rejects the delete while users or OIDC applications still reference it.
func (s *AuthSourceService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/auth_sources/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "AuthSource", ID: id}
		}
		return err
	}
	return nil
}

// readSettings loads the stored settings document, including client_secret,
// so Update can merge a partial change without dropping keys.
func (s *AuthSourceService) readSettings(ctx context.Context, id int) (map[string]any, error) {
	params := url.Values{}
	params.Set("fields", authSourceSettingsFields)

	var raw struct {
		Settings json.RawMessage `json:"settings"`
	}
	endpoint := fmt.Sprintf("/auth_sources/%d", id)
	if err := s.client.get(ctx, endpoint, params, &raw); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "AuthSource", ID: id}
		}
		return nil, err
	}

	settings, err := decodeSettingsObject(raw.Settings)
	if err != nil {
		return nil, &ValidationError{Field: "settings", Message: err.Error()}
	}
	if settings == nil {
		settings = map[string]any{}
	}
	return settings, nil
}

// mergeAuthSourceSettings overlays patch onto current. Keys only in current
// are kept. The caller's map is not modified.
func mergeAuthSourceSettings(current map[string]any, patch AuthSourceSettings) AuthSourceSettings {
	merged := make(AuthSourceSettings, len(current)+len(patch))
	for k, v := range current {
		merged[k] = v
	}
	for k, v := range patch {
		merged[k] = v
	}
	return merged
}
