package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// TenantService tests
// ---------------------------------------------------------------------------

func TestTenantService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tenant{
				{Key: 1, Name: "tenant-a"},
				{Key: 2, Name: "tenant-b"},
			})
		},
	}))

	tenants, err := client.Tenants.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tenants) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(tenants))
	}
	if tenants[0].Name != "tenant-a" {
		t.Errorf("expected name 'tenant-a', got %q", tenants[0].Name)
	}
}

func TestTenantService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 1, Name: "tenant-a", Description: "first"})
		},
	}))

	tenant, err := client.Tenants.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if tenant.Name != "tenant-a" {
		t.Errorf("expected name 'tenant-a', got %q", tenant.Name)
	}
	if tenant.Description != "first" {
		t.Errorf("expected description 'first', got %q", tenant.Description)
	}
}

func TestTenantService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Tenants.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tenant{{Key: 1, Name: "tenant-a"}})
		},
		"GET /api/v4/tenants/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 1, Name: "tenant-a", Description: "full"})
		},
	}))

	tenant, err := client.Tenants.GetByName(context.Background(), "tenant-a")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if tenant.Description != "full" {
		t.Errorf("expected description 'full', got %q", tenant.Description)
	}
}

func TestTenantService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tenant{})
		},
	}))

	_, err := client.Tenants.GetByName(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenants": func(w http.ResponseWriter, r *http.Request) {
			var req TenantCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "new-tenant" {
				t.Errorf("expected name 'new-tenant', got %q", req.Name)
			}
			jsonResponse(w, 200, apiResponse{Key: 10})
		},
		"GET /api/v4/tenants/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 10, Name: "new-tenant"})
		},
	}))

	tenant, err := client.Tenants.Create(context.Background(), &TenantCreateRequest{
		Name: "new-tenant",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if tenant.Name != "new-tenant" {
		t.Errorf("expected name 'new-tenant', got %q", tenant.Name)
	}
}

func TestTenantService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tenants.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantService_Create_EmptyName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tenants.Create(context.Background(), &TenantCreateRequest{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantService_Update(t *testing.T) {
	newDesc := "updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenants/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Description == nil || *req.Description != newDesc {
				t.Errorf("expected description %q, got %v", newDesc, req.Description)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenants/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 1, Name: "tenant-a", Description: newDesc})
		},
	}))

	tenant, err := client.Tenants.Update(context.Background(), 1, &TenantUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if tenant.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, tenant.Description)
	}
}

func TestTenantService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tenants.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenants/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "nope"
	_, err := client.Tenants.Update(context.Background(), 999, &TenantUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenants/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTenantService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenants/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Tenants.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantService_PowerOn(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "poweron" {
				t.Errorf("expected action 'poweron', got %q", body.Action)
			}
			if body.Tenant != 1 {
				t.Errorf("expected tenant 1, got %d", body.Tenant)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.PowerOn(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOn failed: %v", err)
	}
}

func TestTenantService_PowerOnWithNode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "poweron" {
				t.Errorf("expected action 'poweron', got %q", body.Action)
			}
			if body.Params == nil {
				t.Fatal("expected params with preferred_node")
			}
			if pn, ok := body.Params["preferred_node"].(float64); !ok || int(pn) != 5 {
				t.Errorf("expected preferred_node 5, got %v", body.Params["preferred_node"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.PowerOnWithNode(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("PowerOnWithNode failed: %v", err)
	}
}

func TestTenantService_PowerOff(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "poweroff" {
				t.Errorf("expected action 'poweroff', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.PowerOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOff failed: %v", err)
	}
}

func TestTenantService_Reset(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "reset" {
				t.Errorf("expected action 'reset', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.Reset(context.Background(), 1)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestTenantService_Clone(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "clone" {
				t.Errorf("expected action 'clone', got %q", body.Action)
			}
			if body.Params == nil {
				t.Fatal("expected params for clone")
			}
			if name, ok := body.Params["name"].(string); !ok || name != "cloned" {
				t.Errorf("expected clone name 'cloned', got %v", body.Params["name"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.Clone(context.Background(), 1, &TenantCloneOptions{Name: "cloned"})
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
}

func TestTenantService_Clone_NilOptions(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.Tenants.Clone(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil options")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantService_Clone_EmptyName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.Tenants.Clone(context.Background(), 1, &TenantCloneOptions{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantService_IsolateOn(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "isolateon" {
				t.Errorf("expected action 'isolateon', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.IsolateOn(context.Background(), 1)
	if err != nil {
		t.Fatalf("IsolateOn failed: %v", err)
	}
}

func TestTenantService_IsolateOff(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "isolateoff" {
				t.Errorf("expected action 'isolateoff', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tenants.IsolateOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("IsolateOff failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TenantNodeService tests
// ---------------------------------------------------------------------------

func TestTenantNodeService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantNode{
				{Key: 1, Name: "node-1", CPUCores: 4, RAM: 8192},
				{Key: 2, Name: "node-2", CPUCores: 8, RAM: 16384},
			})
		},
	}))

	nodes, err := client.TenantNodes.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Name != "node-1" {
		t.Errorf("expected name 'node-1', got %q", nodes[0].Name)
	}
}

func TestTenantNodeService_ListByTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for tenant")
			}
			jsonResponse(w, 200, []TenantNode{{Key: 1, Tenant: 5, Name: "node-1"}})
		},
	}))

	nodes, err := client.TenantNodes.ListByTenant(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
}

func TestTenantNodeService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNode{Key: 1, Name: "node-1", CPUCores: 4, RAM: 8192})
		},
	}))

	node, err := client.TenantNodes.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if node.CPUCores != 4 {
		t.Errorf("expected 4 cores, got %d", node.CPUCores)
	}
}

func TestTenantNodeService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TenantNodes.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantNode{{Key: 1, Tenant: 5, Name: "node-1"}})
		},
		"GET /api/v4/tenant_nodes/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNode{Key: 1, Tenant: 5, Name: "node-1", CPUCores: 4})
		},
	}))

	node, err := client.TenantNodes.GetByName(context.Background(), 5, "node-1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if node.CPUCores != 4 {
		t.Errorf("expected 4 cores, got %d", node.CPUCores)
	}
}

func TestTenantNodeService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantNode{})
		},
	}))

	_, err := client.TenantNodes.GetByName(context.Background(), 5, "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_nodes": func(w http.ResponseWriter, r *http.Request) {
			var req TenantNodeCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Tenant != 5 {
				t.Errorf("expected tenant 5, got %d", req.Tenant)
			}
			if req.CPUCores != 4 {
				t.Errorf("expected 4 cores, got %d", req.CPUCores)
			}
			jsonResponse(w, 200, apiResponse{Key: 10})
		},
		"GET /api/v4/tenant_nodes/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNode{Key: 10, Tenant: 5, Name: "node-new", CPUCores: 4, RAM: 4096})
		},
	}))

	node, err := client.TenantNodes.Create(context.Background(), &TenantNodeCreateRequest{
		Tenant:   5,
		CPUCores: 4,
		RAM:      4096,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if node.Name != "node-new" {
		t.Errorf("expected name 'node-new', got %q", node.Name)
	}
}

func TestTenantNodeService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantNodes.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Create_MissingTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantNodes.Create(context.Background(), &TenantNodeCreateRequest{
		CPUCores: 4,
		RAM:      4096,
	})
	if err == nil {
		t.Fatal("expected error for missing tenant")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Create_MissingCPUCores(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantNodes.Create(context.Background(), &TenantNodeCreateRequest{
		Tenant: 5,
		RAM:    4096,
	})
	if err == nil {
		t.Fatal("expected error for missing cpu_cores")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Create_InsufficientRAM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantNodes.Create(context.Background(), &TenantNodeCreateRequest{
		Tenant:   5,
		CPUCores: 4,
		RAM:      1024,
	})
	if err == nil {
		t.Fatal("expected error for insufficient RAM")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Update(t *testing.T) {
	newDesc := "updated node"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_nodes/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_nodes/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNode{Key: 1, Name: "node-1", Description: newDesc})
		},
	}))

	node, err := client.TenantNodes.Update(context.Background(), 1, &TenantNodeUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if node.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, node.Description)
	}
}

func TestTenantNodeService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantNodes.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_nodes/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newDesc := "nope"
	_, err := client.TenantNodes.Update(context.Background(), 999, &TenantNodeUpdateRequest{Description: &newDesc})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenant_nodes/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTenantNodeService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenant_nodes/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.TenantNodes.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantNodeService_PowerOn(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantNodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "poweron" {
				t.Errorf("expected action 'poweron', got %q", body.Action)
			}
			if body.TenantNode != 1 {
				t.Errorf("expected tenant_node 1, got %d", body.TenantNode)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.PowerOn(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOn failed: %v", err)
	}
}

func TestTenantNodeService_PowerOff(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantNodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "poweroff" {
				t.Errorf("expected action 'poweroff', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.PowerOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOff failed: %v", err)
	}
}

func TestTenantNodeService_Reset(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantNodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "reset" {
				t.Errorf("expected action 'reset', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.Reset(context.Background(), 1)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestTenantNodeService_Kill(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantNodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "kill" {
				t.Errorf("expected action 'kill', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.Kill(context.Background(), 1)
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}
}

func TestTenantNodeService_Migrate(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantNodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "migrate" {
				t.Errorf("expected action 'migrate', got %q", body.Action)
			}
			if body.Params == nil {
				t.Fatal("expected params with target node")
			}
			if node, ok := body.Params["node"].(float64); !ok || int(node) != 3 {
				t.Errorf("expected target node 3, got %v", body.Params["node"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.Migrate(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
}

func TestTenantNodeService_Migrate_NoTarget(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body tenantNodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "migrate" {
				t.Errorf("expected action 'migrate', got %q", body.Action)
			}
			if body.Params != nil {
				t.Error("expected nil params when no target node specified")
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodes.Migrate(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
}
