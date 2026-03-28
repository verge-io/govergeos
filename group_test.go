package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGroupService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Group{
				{ID: 1, Name: "admins"},
				{ID: 2, Name: "operators"},
			})
		},
	}))

	groups, err := client.Groups.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "admins" {
		t.Errorf("expected name 'admins', got %q", groups[0].Name)
	}
}

func TestGroupService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter parameter")
			}
			jsonResponse(w, 200, []Group{{ID: 1, Name: "admins"}})
		},
	}))

	groups, err := client.Groups.List(context.Background(), WithFilter("enabled eq true"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
}

func TestGroupService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Group{ID: 1, Name: "admins", Description: "Administrator group"})
		},
	}))

	group, err := client.Groups.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if group.Name != "admins" {
		t.Errorf("expected name 'admins', got %q", group.Name)
	}
	if group.Description != "Administrator group" {
		t.Errorf("expected description 'Administrator group', got %q", group.Description)
	}
}

func TestGroupService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Groups.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestGroupService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "name eq 'admins'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Group{{ID: 1, Name: "admins"}})
		},
	}))

	group, err := client.Groups.GetByName(context.Background(), "admins")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if group.Name != "admins" {
		t.Errorf("expected name 'admins', got %q", group.Name)
	}
}

func TestGroupService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Group{})
		},
	}))

	_, err := client.Groups.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestGroupService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
			var req GroupCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "newgroup" {
				t.Errorf("expected name 'newgroup', got %q", req.Name)
			}
			jsonResponse(w, 200, struct {
				Key FlexInt `json:"$key"`
			}{Key: 5})
		},
		"GET /api/v4/groups/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Group{ID: 5, Name: "newgroup"})
		},
	}))

	group, err := client.Groups.Create(context.Background(), &GroupCreateRequest{
		Name: "newgroup",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if group.Name != "newgroup" {
		t.Errorf("expected name 'newgroup', got %q", group.Name)
	}
	if int(group.ID) != 5 {
		t.Errorf("expected ID 5, got %d", int(group.ID))
	}
}

func TestGroupService_Update(t *testing.T) {
	newDesc := "Updated description"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/groups/1": func(w http.ResponseWriter, r *http.Request) {
			var req GroupUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Description == nil || *req.Description != newDesc {
				t.Errorf("expected description %q, got %v", newDesc, req.Description)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/groups/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Group{ID: 1, Name: "admins", Description: newDesc})
		},
	}))

	group, err := client.Groups.Update(context.Background(), 1, &GroupUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if group.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, group.Description)
	}
}

func TestGroupService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/groups/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Groups.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestGroupService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/groups/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Groups.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
