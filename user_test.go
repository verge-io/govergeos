package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestUserService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []User{
				{Key: 1, Name: "admin", Enabled: true},
				{Key: 2, Name: "operator", Enabled: true},
			})
		},
	}))

	users, err := client.Users.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "admin" {
		t.Errorf("expected name 'admin', got %q", users[0].Name)
	}
}

func TestUserService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter parameter")
			}
			jsonResponse(w, 200, []User{{Key: 1, Name: "admin", Enabled: true}})
		},
	}))

	users, err := client.Users.List(context.Background(), WithFilter("enabled eq true"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestUserService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, User{Key: 1, Name: "admin", Email: "admin@example.com", Enabled: true})
		},
	}))

	user, err := client.Users.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if user.Name != "admin" {
		t.Errorf("expected name 'admin', got %q", user.Name)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("expected email 'admin@example.com', got %q", user.Email)
	}
}

func TestUserService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Users.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "name eq 'admin'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []User{{Key: 1, Name: "admin"}})
		},
	}))

	user, err := client.Users.GetByName(context.Background(), "admin")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if user.Name != "admin" {
		t.Errorf("expected name 'admin', got %q", user.Name)
	}
}

func TestUserService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []User{})
		},
	}))

	_, err := client.Users.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/users": func(w http.ResponseWriter, r *http.Request) {
			var req UserCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "newuser" {
				t.Errorf("expected name 'newuser', got %q", req.Name)
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: float64(3)})
		},
		"GET /api/v4/users/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, User{Key: 3, Name: "newuser", Enabled: true})
		},
	}))

	user, err := client.Users.Create(context.Background(), &UserCreateRequest{
		Name: "newuser",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.Name != "newuser" {
		t.Errorf("expected name 'newuser', got %q", user.Name)
	}
	if int(user.Key) != 3 {
		t.Errorf("expected key 3, got %d", int(user.Key))
	}
}

func TestUserService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Users.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Users.Create(context.Background(), &UserCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserService_Update(t *testing.T) {
	newEmail := "updated@example.com"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/users/1": func(w http.ResponseWriter, r *http.Request) {
			var req UserUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Email == nil || *req.Email != newEmail {
				t.Errorf("expected email %q, got %v", newEmail, req.Email)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/users/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, User{Key: 1, Name: "admin", Email: newEmail})
		},
	}))

	user, err := client.Users.Update(context.Background(), 1, &UserUpdateRequest{Email: &newEmail})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if user.Email != newEmail {
		t.Errorf("expected email %q, got %q", newEmail, user.Email)
	}
}

func TestUserService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Users.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUserService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/users/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	name := "updated"
	_, err := client.Users.Update(context.Background(), 999, &UserUpdateRequest{Name: &name})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/users/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Users.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestUserService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/users/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete of non-existent resource returns NotFoundError
	err := client.Users.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted user")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUserService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/user_actions": func(w http.ResponseWriter, r *http.Request) {
			var body userAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.User != 1 {
				t.Errorf("expected user 1, got %d", body.User)
			}
			if body.Action != "enable" {
				t.Errorf("expected action 'enable', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Users.Enable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestUserService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/user_actions": func(w http.ResponseWriter, r *http.Request) {
			var body userAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.User != 2 {
				t.Errorf("expected user 2, got %d", body.User)
			}
			if body.Action != "disable" {
				t.Errorf("expected action 'disable', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Users.Disable(context.Background(), 2)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}
