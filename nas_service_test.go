package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// NASService tests

func TestNASServiceService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASService{
				{Key: 1, Name: "nas1", VM: 100},
				{Key: 2, Name: "nas2", VM: 200},
			})
		},
	}))

	services, err := client.NASServices.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
	if services[0].Name != "nas1" {
		t.Errorf("expected name 'nas1', got %q", services[0].Name)
	}
}

func TestNASServiceService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASService{Key: 1, Name: "nas1", VM: 100, Enabled: true})
		},
	}))

	svc, err := client.NASServices.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if svc.Name != "nas1" {
		t.Errorf("expected name 'nas1', got %q", svc.Name)
	}
	if !svc.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestNASServiceService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.NASServices.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNASServiceService_GetByVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASService{{Key: 1, Name: "nas1", VM: 100}})
		},
		"GET /api/v4/vm_services/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASService{Key: 1, Name: "nas1", VM: 100})
		},
	}))

	svc, err := client.NASServices.GetByVM(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetByVM failed: %v", err)
	}
	if svc.Name != "nas1" {
		t.Errorf("expected name 'nas1', got %q", svc.Name)
	}
}

func TestNASServiceService_GetByVM_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASService{})
		},
	}))

	_, err := client.NASServices.GetByVM(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNASServiceService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASService{{Key: 1, Name: "nas1", VM: 100}})
		},
		"GET /api/v4/vm_services/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASService{Key: 1, Name: "nas1", VM: 100})
		},
	}))

	svc, err := client.NASServices.GetByName(context.Background(), "nas1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if svc.Name != "nas1" {
		t.Errorf("expected name 'nas1', got %q", svc.Name)
	}
}

func TestNASServiceService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_services": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASService{})
		},
	}))

	_, err := client.NASServices.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNASServiceService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_services": func(w http.ResponseWriter, r *http.Request) {
			var req NASServiceCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.VM != 100 {
				t.Errorf("expected vm 100, got %d", req.VM)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/vm_services/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASService{Key: 5, Name: "nas-new", VM: 100})
		},
	}))

	svc, err := client.NASServices.Create(context.Background(), &NASServiceCreateRequest{VM: 100})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if svc.Name != "nas-new" {
		t.Errorf("expected name 'nas-new', got %q", svc.Name)
	}
}

func TestNASServiceService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServices.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceService_Create_MissingVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServices.Create(context.Background(), &NASServiceCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing vm")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceService_Update(t *testing.T) {
	maxImports := 8
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_services/1": func(w http.ResponseWriter, r *http.Request) {
			var req NASServiceUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.MaxImports == nil || *req.MaxImports != maxImports {
				t.Errorf("expected max_imports %d, got %v", maxImports, req.MaxImports)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vm_services/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASService{Key: 1, Name: "nas1", MaxImports: maxImports})
		},
	}))

	svc, err := client.NASServices.Update(context.Background(), 1, &NASServiceUpdateRequest{MaxImports: &maxImports})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if svc.MaxImports != maxImports {
		t.Errorf("expected max_imports %d, got %d", maxImports, svc.MaxImports)
	}
}

func TestNASServiceService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServices.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vm_services/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.NASServices.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNASServiceService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vm_services/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.NASServices.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// NASServiceUser tests

func TestNASServiceUserService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_users": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASServiceUser{
				{ID: "abc123", Name: "user1", Service: 1},
				{ID: "def456", Name: "user2", Service: 1},
			})
		},
	}))

	users, err := client.NASServiceUsers.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "user1" {
		t.Errorf("expected name 'user1', got %q", users[0].Name)
	}
}

func TestNASServiceUserService_ListByService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_users": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "service eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NASServiceUser{{ID: "abc123", Name: "user1", Service: 1}})
		},
	}))

	users, err := client.NASServiceUsers.ListByService(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByService failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestNASServiceUserService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceUser{ID: "abc123", Name: "user1", Enabled: true})
		},
	}))

	user, err := client.NASServiceUsers.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if user.Name != "user1" {
		t.Errorf("expected name 'user1', got %q", user.Name)
	}
	if !user.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestNASServiceUserService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_users/nonexistent": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.NASServiceUsers.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_users": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASServiceUser{{ID: "abc123", Name: "user1", Service: 1}})
		},
		"GET /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceUser{ID: "abc123", Name: "user1", Service: 1})
		},
	}))

	user, err := client.NASServiceUsers.GetByName(context.Background(), 1, "user1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if user.Name != "user1" {
		t.Errorf("expected name 'user1', got %q", user.Name)
	}
}

func TestNASServiceUserService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_users": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NASServiceUser{})
		},
	}))

	_, err := client.NASServiceUsers.GetByName(context.Background(), 1, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_service_users": func(w http.ResponseWriter, r *http.Request) {
			var req NASServiceUserCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Service != 1 {
				t.Errorf("expected service 1, got %d", req.Service)
			}
			if req.Name != "newuser" {
				t.Errorf("expected name 'newuser', got %q", req.Name)
			}
			jsonResponse(w, 200, apiResponse{Key: "sha1hash123"})
		},
		"GET /api/v4/vm_service_users/sha1hash123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceUser{ID: "sha1hash123", Name: "newuser", Service: 1, Enabled: true})
		},
	}))

	user, err := client.NASServiceUsers.Create(context.Background(), &NASServiceUserCreateRequest{
		Service:  1,
		Name:     "newuser",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.Name != "newuser" {
		t.Errorf("expected name 'newuser', got %q", user.Name)
	}
}

func TestNASServiceUserService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServiceUsers.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Create_MissingService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServiceUsers.Create(context.Background(), &NASServiceUserCreateRequest{
		Name:     "user",
		Password: "pass",
	})
	if err == nil {
		t.Fatal("expected error for missing service")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServiceUsers.Create(context.Background(), &NASServiceUserCreateRequest{
		Service:  1,
		Password: "pass",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Create_MissingPassword(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServiceUsers.Create(context.Background(), &NASServiceUserCreateRequest{
		Service: 1,
		Name:    "user",
	})
	if err == nil {
		t.Fatal("expected error for missing password")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Update(t *testing.T) {
	newDisplay := "New Display"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			var req NASServiceUserUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.DisplayName == nil || *req.DisplayName != newDisplay {
				t.Errorf("expected display name %q, got %v", newDisplay, req.DisplayName)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceUser{ID: "abc123", Name: "user1", DisplayName: newDisplay})
		},
	}))

	user, err := client.NASServiceUsers.Update(context.Background(), "abc123", &NASServiceUserUpdateRequest{DisplayName: &newDisplay})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if user.DisplayName != newDisplay {
		t.Errorf("expected display name %q, got %q", newDisplay, user.DisplayName)
	}
}

func TestNASServiceUserService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.NASServiceUsers.Update(context.Background(), "abc123", nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.NASServiceUsers.Delete(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNASServiceUserService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vm_service_users/nonexistent": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.NASServiceUsers.Delete(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNASServiceUserService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			var req NASServiceUserUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to be true")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceUser{ID: "abc123", Name: "user1", Enabled: true})
		},
	}))

	err := client.NASServiceUsers.Enable(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestNASServiceUserService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			var req NASServiceUserUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled to be false")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vm_service_users/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceUser{ID: "abc123", Name: "user1", Enabled: false})
		},
	}))

	err := client.NASServiceUsers.Disable(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}
