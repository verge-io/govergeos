package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// UserAPIKeyService handles user API key operations.
type UserAPIKeyService struct {
	client *Client
}

// List returns all user API keys, with optional filtering and pagination.
func (s *UserAPIKeyService) List(ctx context.Context, opts ...ListOption) ([]UserAPIKey, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = userAPIKeyListFields
	}

	params := options.toQueryParams()

	var keys []UserAPIKey
	if err := s.client.get(ctx, "/user_api_keys", params, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// ListByUser returns all API keys for a specific user.
func (s *UserAPIKeyService) ListByUser(ctx context.Context, userID int, opts ...ListOption) ([]UserAPIKey, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("user eq %d", userID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns a single API key by ID.
func (s *UserAPIKeyService) Get(ctx context.Context, id int) (*UserAPIKey, error) {
	params := url.Values{}
	params.Set("fields", userAPIKeyGetFields)

	var key UserAPIKey
	endpoint := fmt.Sprintf("/user_api_keys/%d", id)
	if err := s.client.get(ctx, endpoint, params, &key); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "UserAPIKey", ID: id}
		}
		return nil, err
	}

	return &key, nil
}

// GetByName returns an API key by name within a user's keys.
func (s *UserAPIKeyService) GetByName(ctx context.Context, userID int, name string) (*UserAPIKey, error) {
	keys, err := s.List(ctx,
		WithFilter(fmt.Sprintf("user eq %d and name eq '%s'", userID, escapeFilterValue(name))),
	)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, &NotFoundError{Resource: "UserAPIKey", ID: name}
	}
	if err := requireUniqueName("UserAPIKey", name, keys, func(k UserAPIKey) any { return k.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("UserAPIKey", name, keys[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(keys[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("UserAPIKey", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a new API key and returns the key together with its secret.
// The secret is only available from this call. VergeOS 26.2 returns it as
// private_key inside the response object. A response that still carries the
// secret in token is accepted. When the response includes neither, Create
// returns an error that includes the new key id. The key row already exists.
func (s *UserAPIKeyService) Create(ctx context.Context, req *UserAPIKeyCreateRequest) (*UserAPIKey, string, error) {
	if req == nil {
		return nil, "", &ValidationError{Message: "create request is required"}
	}
	if req.User == 0 {
		return nil, "", &ValidationError{Field: "user", Message: "user is required"}
	}
	if req.Name == "" {
		return nil, "", &ValidationError{Field: "name", Message: "name is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/user_api_keys", req, &resp); err != nil {
		return nil, "", err
	}

	id, err := getKey(resp)
	if err != nil {
		return nil, "", err
	}

	token := userAPIKeySecret(resp.Response)
	if token == "" {
		return nil, "", fmt.Errorf("vergeos: user API key %d was created but the response did not include the secret", id)
	}

	key, err := s.Get(ctx, id)
	if err != nil {
		return nil, token, err
	}

	return key, token, nil
}

// userAPIKeySecret reads the one-time secret from a user_api_keys create
// response object. private_key is the field VergeOS 26.2 returns. token is
// the older field name.
func userAPIKeySecret(response any) string {
	respMap, ok := response.(map[string]any)
	if !ok {
		return ""
	}
	if secret, ok := respMap["private_key"].(string); ok && secret != "" {
		return secret
	}
	if secret, ok := respMap["token"].(string); ok && secret != "" {
		return secret
	}
	return ""
}

// Update updates an API key and returns the updated key.
func (s *UserAPIKeyService) Update(ctx context.Context, id int, req *UserAPIKeyUpdateRequest) (*UserAPIKey, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/user_api_keys/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Delete deletes an API key.
func (s *UserAPIKeyService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/user_api_keys/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "UserAPIKey", ID: id}
		}
		return err
	}
	return nil
}

// ListExpired returns all expired API keys.
func (s *UserAPIKeyService) ListExpired(ctx context.Context, opts ...ListOption) ([]UserAPIKey, error) {
	// Note: The API auto-deletes expired keys, so this may return empty
	filterOpts := []ListOption{WithFilter("expires gt 0")}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}
