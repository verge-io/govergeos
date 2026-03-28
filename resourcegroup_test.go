package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestResourceGroupService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/resource_groups": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ResourceGroup{
				{Name: "compute", Type: "vm", Enabled: true},
				{Name: "storage", Type: "volume", Enabled: true},
			})
		},
	}))

	groups, err := client.ResourceGroups.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "compute" {
		t.Errorf("expected name 'compute', got %q", groups[0].Name)
	}
}

func TestResourceGroupService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/resource_groups/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, ResourceGroup{Name: "compute", Type: "vm", Enabled: true, Description: "Compute resources"})
		},
	}))

	group, err := client.ResourceGroups.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if group.Name != "compute" {
		t.Errorf("expected name 'compute', got %q", group.Name)
	}
	if group.Description != "Compute resources" {
		t.Errorf("expected description 'Compute resources', got %q", group.Description)
	}
}

func TestResourceGroupService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/resource_groups/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.ResourceGroups.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestResourceGroupService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/resource_groups": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "name eq 'compute'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []ResourceGroup{{Name: "compute", Type: "vm"}})
		},
	}))

	group, err := client.ResourceGroups.GetByName(context.Background(), "compute")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if group.Name != "compute" {
		t.Errorf("expected name 'compute', got %q", group.Name)
	}
}

func TestResourceGroupService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/resource_groups": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ResourceGroup{})
		},
	}))

	_, err := client.ResourceGroups.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
