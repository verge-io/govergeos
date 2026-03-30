package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
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

func TestUserAPIKeyService_Create(t *testing.T) {
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
