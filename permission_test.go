package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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

func TestPermissionService_NegativeRow(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, client *Client) error
	}{
		{
			name: "Create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Permissions.Create(ctx, &PermissionCreateRequest{
					Identity: 10,
					Table:    "vms",
					Row:      -1,
				})
				return err
			},
		},
		{
			name: "GrantReadOnly",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Permissions.GrantReadOnly(ctx, 10, "vms", -1)
				return err
			},
		},
		{
			name: "GrantFullAccess",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Permissions.GrantFullAccess(ctx, 10, "vms", -1)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posted := false
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
					posted = true
					w.WriteHeader(http.StatusCreated)
				},
			}))

			err := tt.call(context.Background(), client)
			if err == nil {
				t.Fatal("expected error for negative row")
			}
			if !IsValidationError(err) {
				t.Fatalf("expected ValidationError, got %T: %v", err, err)
			}
			ve := err.(*ValidationError)
			if ve.Field != "row" {
				t.Errorf("expected field row, got %q", ve.Field)
			}
			if posted {
				t.Error("negative row was posted")
			}
		})
	}
}

func TestPermissionService_TableWideRowZero(t *testing.T) {
	readTrue := true
	tests := []struct {
		name string
		call func(ctx context.Context, client *Client) error
	}{
		{
			name: "Create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Permissions.Create(ctx, &PermissionCreateRequest{
					Identity: 11,
					Table:    "webhook_urls",
					Row:      0,
					List:     &readTrue,
					Read:     &readTrue,
				})
				return err
			},
		},
		{
			name: "GrantReadOnly",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Permissions.GrantReadOnly(ctx, 11, "webhook_urls", 0)
				return err
			},
		},
		{
			name: "GrantFullAccess",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Permissions.GrantFullAccess(ctx, 11, "webhook_urls", 0)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("read body: %v", err)
					}
					if !strings.Contains(string(body), `"row":0`) {
						t.Fatalf("body %s does not contain \"row\":0", body)
					}
					var raw map[string]json.RawMessage
					if err := json.Unmarshal(body, &raw); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					if string(raw["row"]) != "0" {
						t.Fatalf("row JSON = %s, want 0", raw["row"])
					}
					jsonResponse(w, http.StatusCreated, map[string]any{"$key": 23})
				},
				"GET /api/v4/permissions/23": func(w http.ResponseWriter, r *http.Request) {
					jsonResponse(w, 200, Permission{Key: 23, Identity: 11, Table: "webhook_urls", Row: 0, List: true, Read: true})
				},
			}))

			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("%s row 0: %v", tt.name, err)
			}
		})
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
	apiErr, ok := err.(*APIError)
	if IsNotFoundError(err) || !ok || apiErr.StatusCode != 404 {
		t.Errorf("expected API 404, not a missing resource, got %T: %v", err, err)
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

func TestPermission_IdentityJSONString(t *testing.T) {
	var perm Permission
	if err := json.Unmarshal([]byte(`{"$key":22,"identity":"11","table":"users","row":11}`), &perm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if perm.Key != 22 || perm.Identity != 11 || perm.Row != 11 || perm.Table != "users" {
		t.Fatalf("got key %d identity %d table %s row %d", perm.Key, perm.Identity, perm.Table, perm.Row)
	}
}

func TestPermissionService_Grant_ResolvedIdentity(t *testing.T) {
	// User key 11 owns identity 16. Identity 11 belongs to a different user.
	// The grant has to post 16.
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/users/11": func(w http.ResponseWriter, r *http.Request) {
			if !fieldListHas(r.URL.Query().Get("fields"), "identity") {
				t.Errorf("fields %q do not request identity", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, json.RawMessage(`{"$key":11,"name":"bugtest-g2-u5","identity":"16"}`))
		},
		"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.Identity != 16 {
				t.Errorf("posted identity %d, want 16", req.Identity)
			}
			if req.Table != "users" || req.Row != 11 {
				t.Errorf("posted table %s row %d", req.Table, req.Row)
			}
			jsonResponse(w, 201, map[string]any{"$key": 22})
		},
		"GET /api/v4/permissions/22": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, json.RawMessage(`{"$key":22,"identity":"16","table":"users","row":11,"list":true,"read":true}`))
		},
	}))

	ctx := context.Background()
	identity, err := client.Users.IdentityOf(ctx, 11)
	if err != nil {
		t.Fatalf("IdentityOf: %v", err)
	}
	perm, err := client.Permissions.GrantReadOnly(ctx, identity, "users", 11)
	if err != nil {
		t.Fatalf("GrantReadOnly: %v", err)
	}
	if perm.Identity != 16 {
		t.Fatalf("stored identity %d, want 16", perm.Identity)
	}
}

func TestPermissionService_Grant_GroupIdentity(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, json.RawMessage(`{"$key":4,"name":"operators","identity":9}`))
		},
		"POST /api/v4/permissions": func(w http.ResponseWriter, r *http.Request) {
			var req PermissionCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.Identity != 9 {
				t.Errorf("posted identity %d, want 9", req.Identity)
			}
			jsonResponse(w, 200, map[string]any{"$key": 3})
		},
		"GET /api/v4/permissions/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Permission{Key: 3, Identity: 9, Table: "vms", Row: 100, List: true, Read: true, Create: true, Modify: true, Delete: true})
		},
	}))

	ctx := context.Background()
	identity, err := client.Groups.IdentityOf(ctx, 4)
	if err != nil {
		t.Fatalf("IdentityOf: %v", err)
	}
	perm, err := client.Permissions.GrantFullAccess(ctx, identity, "vms", 100)
	if err != nil {
		t.Fatalf("GrantFullAccess: %v", err)
	}
	if perm.Identity != identity {
		t.Fatalf("stored identity %d, want %d", perm.Identity, identity)
	}
}
