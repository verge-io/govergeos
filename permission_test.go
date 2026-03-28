package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPermissionService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Permission{
				{Key: 1, Identity: 10, Table: "vms", Row: 100, Read: true},
				{Key: 2, Identity: 20, Table: "vnets", Row: 200, Read: true, Modify: true},
			})
		},
	}))

	perms, err := client.Permissions.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(perms))
	}
	if !perms[0].Read {
		t.Error("expected first permission to have read=true")
	}
}

func TestPermissionService_ListByIdentity(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "identity eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Permission{{Key: 1, Identity: 10, Table: "vms"}})
		},
	}))

	perms, err := client.Permissions.ListByIdentity(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByIdentity failed: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(perms))
	}
}

func TestPermissionService_ListByTable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "table eq 'vms'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Permission{{Key: 1, Table: "vms"}})
		},
	}))

	perms, err := client.Permissions.ListByTable(context.Background(), "vms")
	if err != nil {
		t.Fatalf("ListByTable failed: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(perms))
	}
}

func TestPermissionService_ListByResource(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "table eq 'vms' and row eq 100" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Permission{{Key: 1, Table: "vms", Row: 100}})
		},
	}))

	perms, err := client.Permissions.ListByResource(context.Background(), "vms", 100)
	if err != nil {
		t.Fatalf("ListByResource failed: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(perms))
	}
}

func TestPermissionService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, Read: true, Modify: true})
		},
	}))

	perm, err := client.Permissions.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !perm.Modify {
		t.Error("expected modify=true")
	}
	if perm.Table != "vms" {
		t.Errorf("expected table 'vms', got %q", perm.Table)
	}
}

func TestPermissionService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Permissions.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestPermissionService_GetByIdentityAndResource(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Permission{{Key: 1, Identity: 10, Table: "vms", Row: 100}})
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, Read: true})
		},
	}))

	perm, err := client.Permissions.GetByIdentityAndResource(context.Background(), 10, "vms", 100)
	if err != nil {
		t.Fatalf("GetByIdentityAndResource failed: %v", err)
	}
	if int(perm.Identity) != 10 {
		t.Errorf("expected identity 10, got %d", perm.Identity)
	}
}

func TestPermissionService_GetByIdentityAndResource_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Permission{})
		},
	}))

	_, err := client.Permissions.GetByIdentityAndResource(context.Background(), 10, "vms", 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestPermissionService_Create(t *testing.T) {
	readTrue := true
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Identity != 10 {
				t.Errorf("expected identity 10, got %d", req.Identity)
			}
			if req.Table != "vms" {
				t.Errorf("expected table 'vms', got %q", req.Table)
			}
			if req.Row != 100 {
				t.Errorf("expected row 100, got %d", req.Row)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, Read: true})
		},
	}))

	perm, err := client.Permissions.Create(context.Background(), &PermissionCreateRequest{
		Identity: 10,
		Table:    "vms",
		Row:      100,
		Read:     &readTrue,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !perm.Read {
		t.Error("expected read=true")
	}
}

func TestPermissionService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Permissions.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPermissionService_Create_MissingIdentity(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Permissions.Create(context.Background(), &PermissionCreateRequest{
		Table: "vms",
		Row:   100,
	})
	if err == nil {
		t.Fatal("expected error for missing identity")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPermissionService_Create_MissingTable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Permissions.Create(context.Background(), &PermissionCreateRequest{
		Identity: 10,
		Row:      100,
	})
	if err == nil {
		t.Fatal("expected error for missing table")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPermissionService_Create_MissingRow(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Permissions.Create(context.Background(), &PermissionCreateRequest{
		Identity: 10,
		Table:    "vms",
	})
	if err == nil {
		t.Fatal("expected error for missing row")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPermissionService_Update(t *testing.T) {
	modifyTrue := true
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Modify == nil || !*req.Modify {
				t.Error("expected modify=true")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, Read: true, Modify: true})
		},
	}))

	perm, err := client.Permissions.Update(context.Background(), 1, &PermissionUpdateRequest{Modify: &modifyTrue})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !perm.Modify {
		t.Error("expected modify=true after update")
	}
}

func TestPermissionService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Permissions.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestPermissionService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/permissions/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	modifyTrue := true
	_, err := client.Permissions.Update(context.Background(), 999, &PermissionUpdateRequest{Modify: &modifyTrue})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestPermissionService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Permissions.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestPermissionService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/permissions/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Permissions.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestPermissionService_Grant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Identity != 10 {
				t.Errorf("expected identity 10, got %d", req.Identity)
			}
			if req.Read == nil || !*req.Read {
				t.Error("expected read=true")
			}
			if req.Modify == nil || !*req.Modify {
				t.Error("expected modify=true")
			}
			if req.Delete == nil || *req.Delete {
				t.Error("expected delete=false")
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, List: true, Read: true, Modify: true})
		},
	}))

	perm, err := client.Permissions.Grant(context.Background(), 10, "vms", 100, true, true, false)
	if err != nil {
		t.Fatalf("Grant failed: %v", err)
	}
	if !perm.Read || !perm.Modify {
		t.Error("expected read and modify to be true")
	}
}

func TestPermissionService_GrantReadOnly(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Read == nil || !*req.Read {
				t.Error("expected read=true")
			}
			if req.Modify != nil && *req.Modify {
				t.Error("expected modify=false")
			}
			if req.Delete != nil && *req.Delete {
				t.Error("expected delete=false")
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, List: true, Read: true})
		},
	}))

	perm, err := client.Permissions.GrantReadOnly(context.Background(), 10, "vms", 100)
	if err != nil {
		t.Fatalf("GrantReadOnly failed: %v", err)
	}
	if !perm.Read {
		t.Error("expected read=true")
	}
	if perm.Modify {
		t.Error("expected modify=false for read-only")
	}
}

func TestPermissionService_GrantFullAccess(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Read == nil || !*req.Read {
				t.Error("expected read=true")
			}
			if req.Create == nil || !*req.Create {
				t.Error("expected create=true")
			}
			if req.Modify == nil || !*req.Modify {
				t.Error("expected modify=true")
			}
			if req.Delete == nil || !*req.Delete {
				t.Error("expected delete=true")
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100, List: true, Read: true, Create: true, Modify: true, Delete: true})
		},
	}))

	perm, err := client.Permissions.GrantFullAccess(context.Background(), 10, "vms", 100)
	if err != nil {
		t.Fatalf("GrantFullAccess failed: %v", err)
	}
	if !perm.Read || !perm.Create || !perm.Modify || !perm.Delete {
		t.Error("expected all permissions to be true for full access")
	}
}

func TestPermissionService_Revoke(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Permission{{Key: 1, Identity: 10, Table: "vms", Row: 100}})
		},
		"GET /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 1, Identity: 10, Table: "vms", Row: 100})
		},
		"DELETE /api/v4/permissions/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Permissions.Revoke(context.Background(), 10, "vms", 100)
	if err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}
}

func TestPermissionService_Revoke_NotExists(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Permission{})
		},
	}))

	// Revoke on nonexistent permission should not error (idempotent).
	// GetByIdentityAndResource returns NotFoundError, which Revoke treats as success.
	err := client.Permissions.Revoke(context.Background(), 10, "vms", 999)
	if err != nil {
		t.Fatalf("Revoke of nonexistent permission should not error, got: %v", err)
	}
}
