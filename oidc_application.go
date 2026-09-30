package vergeos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// OIDCApplicationService handles OIDC applications where VergeOS is the
// identity provider.
type OIDCApplicationService struct {
	client *Client
}

// List returns OIDC applications, with optional filtering and pagination.
// client_secret is not requested and is discarded if a response includes it.
func (s *OIDCApplicationService) List(ctx context.Context, opts ...ListOption) ([]OIDCApplication, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = oidcApplicationListFields
	}

	params := options.toQueryParams()

	var apps []OIDCApplication
	if err := s.client.get(ctx, "/oidc_applications", params, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

// Get returns an OIDC application by ID.
// client_secret is not requested and is discarded if a response includes it.
func (s *OIDCApplicationService) Get(ctx context.Context, id int) (*OIDCApplication, error) {
	app, _, err := s.get(ctx, id, oidcApplicationGetFields)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// GetByName returns the OIDC application with this exact name.
func (s *OIDCApplicationService) GetByName(ctx context.Context, name string) (*OIDCApplication, error) {
	apps, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(apps) == 0 {
		return nil, &NotFoundError{Resource: "OIDCApplication", ID: name}
	}
	if err := requireUniqueName("OIDCApplication", name, apps, func(app OIDCApplication) any { return app.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("OIDCApplication", name, apps[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(apps[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("OIDCApplication", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates an OIDC application.
//
// VergeOS generates client_id and client_secret. The secret is returned
// separately and is not stored on the application. Printing the returned
// WriteOnlySecret redacts it; call Value to read it. The secret is not
// available from later Get or List calls.
func (s *OIDCApplicationService) Create(ctx context.Context, req *OIDCApplicationCreateRequest) (*OIDCApplication, WriteOnlySecret, error) {
	if req == nil {
		return nil, WriteOnlySecret{}, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, WriteOnlySecret{}, &ValidationError{Field: "name", Message: "name is required"}
	}

	body := *req
	applyOIDCCreateDefaults(&body)

	var raw json.RawMessage
	if err := s.client.post(ctx, "/oidc_applications", &body, &raw); err != nil {
		return nil, WriteOnlySecret{}, err
	}

	id, secretFromPost, err := parseOIDCCreateResponse(raw)
	if err != nil {
		return nil, NewWriteOnlySecret(secretFromPost), err
	}

	app, secret, err := s.get(ctx, id, oidcApplicationSecretFields)
	if secret.Value() == "" && secretFromPost != "" {
		secret = NewWriteOnlySecret(secretFromPost)
	}
	if err != nil {
		return nil, secret, err
	}
	return app, secret, nil
}

// Update updates an OIDC application and returns it.
// client_id and client_secret cannot be changed.
func (s *OIDCApplicationService) Update(ctx context.Context, id int, req *OIDCApplicationUpdateRequest) (*OIDCApplication, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/oidc_applications/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "OIDCApplication", ID: id}
		}
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete deletes an OIDC application.
func (s *OIDCApplicationService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/oidc_applications/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "OIDCApplication", ID: id}
		}
		return err
	}
	return nil
}

// get reads one application. fields must not be logged with a response body.
// A client_secret in the payload is removed before the application is decoded
// and returned separately so callers of Get can ignore it.
func (s *OIDCApplicationService) get(ctx context.Context, id int, fields string) (*OIDCApplication, WriteOnlySecret, error) {
	params := url.Values{}
	params.Set("fields", fields)

	var raw json.RawMessage
	endpoint := fmt.Sprintf("/oidc_applications/%d", id)
	if err := s.client.get(ctx, endpoint, params, &raw); err != nil {
		if IsNotFoundError(err) {
			return nil, WriteOnlySecret{}, &NotFoundError{Resource: "OIDCApplication", ID: id}
		}
		return nil, WriteOnlySecret{}, err
	}

	app, secret, err := decodeOIDCApplication(raw)
	if err != nil {
		return nil, secret, err
	}
	return app, secret, nil
}

// applyOIDCCreateDefaults fills the values the API expects when the caller
// left them unset. force_auth_source and map_user stay nil so they encode
// as JSON null: 0 is not a valid row and the field is required.
func applyOIDCCreateDefaults(req *OIDCApplicationCreateRequest) {
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}
	if req.ScopeProfile == nil {
		scope := true
		req.ScopeProfile = &scope
	}
	if req.ScopeEmail == nil {
		scope := true
		req.ScopeEmail = &scope
	}
	if req.ScopeGroups == nil {
		scope := true
		req.ScopeGroups = &scope
	}
}

// parseOIDCCreateResponse reads the new row ID and any client secret the
// create response included. The raw body is not attached to errors.
func parseOIDCCreateResponse(raw json.RawMessage) (int, string, error) {
	var resp apiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, "", errors.New("vergeos: failed to decode create response")
	}
	id, err := getKey(resp)
	if err != nil {
		return 0, "", err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return id, "", nil
	}
	secret := secretFromRaw(fields[clientSecretField])
	if secret == "" {
		secret = secretFromRawObject(fields["response"])
	}
	return id, secret, nil
}

// decodeOIDCApplication decodes one application and lifts client_secret out
// of the payload before the application is filled, so the struct field stays
// empty even if UnmarshalJSON were to change.
func decodeOIDCApplication(data []byte) (*OIDCApplication, WriteOnlySecret, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, WriteOnlySecret{}, errors.New("vergeos: failed to decode OIDC application")
	}
	secret := secretFromRaw(fields[clientSecretField])
	delete(fields, clientSecretField)

	stripped, err := json.Marshal(fields)
	if err != nil {
		return nil, NewWriteOnlySecret(secret), fmt.Errorf("vergeos: failed to decode OIDC application: %w", err)
	}
	var app OIDCApplication
	if err := json.Unmarshal(stripped, &app); err != nil {
		return nil, NewWriteOnlySecret(secret), fmt.Errorf("vergeos: failed to decode OIDC application: %w", err)
	}
	return &app, NewWriteOnlySecret(secret), nil
}

// secretFromRaw reads a JSON string. Any other shape, including null, is
// an empty secret. The raw value is not included in an error.
func secretFromRaw(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var secret string
	if err := json.Unmarshal(raw, &secret); err != nil {
		return ""
	}
	return secret
}

// secretFromRawObject reads client_secret from a nested JSON object.
func secretFromRawObject(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return ""
	}
	return secretFromRaw(fields[clientSecretField])
}
