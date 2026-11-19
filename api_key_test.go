package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// UserAPIKey tests

func TestUserAPIKeyService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []UserAPIKey{
				{Key: 1, Name: "key1", User: 10},
				{Key: 2, Name: "key2", User: 20},
			})
		},
	}))

	keys, err := client.UserAPIKeys.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Name != "key1" {
		t.Errorf("expected name 'key1', got %q", keys[0].Name)
	}
}

func TestUserAPIKeyService_ListByUser(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "user eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []UserAPIKey{{Key: 1, Name: "key1", User: 10}})
		},
	}))

	keys, err := client.UserAPIKeys.ListByUser(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
}

func TestUserAPIKeyService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UserAPIKey{Key: 1, Name: "my-key", User: 10, Description: "test key"})
		},
	}))

	key, err := client.UserAPIKeys.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if key.Name != "my-key" {
		t.Errorf("expected name 'my-key', got %q", key.Name)
	}
	if key.Description != "test key" {
		t.Errorf("expected description 'test key', got %q", key.Description)
	}
}

func TestUserAPIKeyService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.UserAPIKeys.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []UserAPIKey{{Key: 1, Name: "my-key", User: 10}})
		},
		"GET /api/v4/user_api_keys/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UserAPIKey{Key: 1, Name: "my-key", User: 10})
		},
	}))

	key, err := client.UserAPIKeys.GetByName(context.Background(), 10, "my-key")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if key.Name != "my-key" {
		t.Errorf("expected name 'my-key', got %q", key.Name)
	}
}

func TestUserAPIKeyService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []UserAPIKey{})
		},
	}))

	_, err := client.UserAPIKeys.GetByName(context.Background(), 10, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_Create_PrivateKey(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			var req UserAPIKeyCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.User != 10 {
				t.Errorf("expected user 10, got %d", req.User)
			}
			if req.Name != "new-key" {
				t.Errorf("expected name 'new-key', got %q", req.Name)
			}
			// VergeOS 26.2 create body: the secret is response.private_key.
			jsonResponse(w, http.StatusCreated, map[string]any{
				"$key":     "5",
				"$row":     3,
				"dbpath":   "user_api_keys/5",
				"location": "/api/v4/user_api_keys/5",
				"response": map[string]any{"private_key": secret},
			})
		},
		"GET /api/v4/user_api_keys/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UserAPIKey{Key: 5, Name: "new-key", User: 10})
		},
	}))

	key, token, err := client.UserAPIKeys.Create(context.Background(), &UserAPIKeyCreateRequest{
		User: 10,
		Name: "new-key",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if key.Name != "new-key" {
		t.Errorf("expected name 'new-key', got %q", key.Name)
	}
	if int(key.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(key.Key))
	}
	if token != secret {
		t.Errorf("expected secret from private_key, got %q", token)
	}
}

func TestUserAPIKeyService_Create_Token(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, apiResponse{
				Key:      float64(5),
				Response: map[string]any{"token": "secret-token-123"},
			})
		},
		"GET /api/v4/user_api_keys/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UserAPIKey{Key: 5, Name: "new-key", User: 10})
		},
	}))

	key, token, err := client.UserAPIKeys.Create(context.Background(), &UserAPIKeyCreateRequest{
		User: 10,
		Name: "new-key",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if key.Name != "new-key" {
		t.Errorf("expected name 'new-key', got %q", key.Name)
	}
	if token != "secret-token-123" {
		t.Errorf("expected token 'secret-token-123', got %q", token)
	}
}

func TestUserAPIKeyService_Create_PrivateKeyPreferred(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, http.StatusCreated, map[string]any{
				"$key": "5",
				"response": map[string]any{
					"private_key": "from-private-key",
					"token":       "from-token",
				},
			})
		},
		"GET /api/v4/user_api_keys/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UserAPIKey{Key: 5, Name: "new-key", User: 10})
		},
	}))

	_, token, err := client.UserAPIKeys.Create(context.Background(), &UserAPIKeyCreateRequest{
		User: 10,
		Name: "new-key",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if token != "from-private-key" {
		t.Errorf("expected private_key secret, got %q", token)
	}
}

func TestUserAPIKeyService_Create_MissingSecret(t *testing.T) {
	cases := []struct {
		name string
		body any
	}{
		{
			name: "no response object",
			body: map[string]any{"$key": "5"},
		},
		{
			name: "empty response object",
			body: map[string]any{"$key": "5", "response": map[string]any{}},
		},
		{
			name: "empty private_key",
			body: map[string]any{"$key": "5", "response": map[string]any{"private_key": ""}},
		},
		{
			name: "empty token",
			body: map[string]any{"$key": "5", "response": map[string]any{"token": ""}},
		},
		{
			name: "null response",
			body: map[string]any{"$key": "5", "response": nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"POST /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
					jsonResponse(w, http.StatusCreated, body)
				},
			}))

			key, token, err := client.UserAPIKeys.Create(context.Background(), &UserAPIKeyCreateRequest{
				User: 10,
				Name: "new-key",
			})
			if err == nil {
				t.Fatal("expected error when the create response has no secret")
			}
			if token != "" {
				t.Errorf("expected empty token, got %q", token)
			}
			if key != nil {
				t.Errorf("expected nil key, got %+v", key)
			}
			if !strings.Contains(err.Error(), "secret") || !strings.Contains(err.Error(), "5") {
				t.Errorf("expected error to name the missing secret and key id, got %v", err)
			}
		})
	}
}

func TestUserAPIKeyService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, _, err := client.UserAPIKeys.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_Create_MissingUser(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, _, err := client.UserAPIKeys.Create(context.Background(), &UserAPIKeyCreateRequest{
		Name: "key",
	})
	if err == nil {
		t.Fatal("expected error for missing user")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, _, err := client.UserAPIKeys.Create(context.Background(), &UserAPIKeyCreateRequest{
		User: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_Update(t *testing.T) {
	newName := "updated-key"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/user_api_keys/1": func(w http.ResponseWriter, r *http.Request) {
			var req UserAPIKeyUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/user_api_keys/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UserAPIKey{Key: 1, Name: newName, User: 10})
		},
	}))

	key, err := client.UserAPIKeys.Update(context.Background(), 1, &UserAPIKeyUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if key.Name != newName {
		t.Errorf("expected name %q, got %q", newName, key.Name)
	}
}

func TestUserAPIKeyService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.UserAPIKeys.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/user_api_keys/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.UserAPIKeys.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestUserAPIKeyService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/user_api_keys/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.UserAPIKeys.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserAPIKeyService_ListExpired(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/user_api_keys": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "expires gt 0" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []UserAPIKey{{Key: 1, Expires: 1700000000}})
		},
	}))

	keys, err := client.UserAPIKeys.ListExpired(context.Background())
	if err != nil {
		t.Fatalf("ListExpired failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
}
