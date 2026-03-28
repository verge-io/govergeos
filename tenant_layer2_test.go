package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTenantLayer2NetworkService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantLayer2Network{
				{Key: 1, Tenant: 10, VNet: 100, Enabled: true},
				{Key: 2, Tenant: 20, VNet: 200, Enabled: false},
			})
		},
	}))

	networks, err := client.TenantLayer2Networks.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(networks) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(networks))
	}
	if !networks[0].Enabled {
		t.Error("expected first network to be enabled")
	}
}

func TestTenantLayer2NetworkService_ListByTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "tenant eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantLayer2Network{{Key: 1, Tenant: 10, VNet: 100}})
		},
	}))

	networks, err := client.TenantLayer2Networks.ListByTenant(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(networks))
	}
}

func TestTenantLayer2NetworkService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 100" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantLayer2Network{{Key: 1, Tenant: 10, VNet: 100}})
		},
	}))

	networks, err := client.TenantLayer2Networks.ListByNetwork(context.Background(), 100)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(networks))
	}
}

func TestTenantLayer2NetworkService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 1, Tenant: 10, VNet: 100, Enabled: true})
		},
	}))

	network, err := client.TenantLayer2Networks.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if int(network.Tenant) != 10 {
		t.Errorf("expected tenant 10, got %d", int(network.Tenant))
	}
	if int(network.VNet) != 100 {
		t.Errorf("expected vnet 100, got %d", int(network.VNet))
	}
}

func TestTenantLayer2NetworkService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TenantLayer2Networks.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_GetByTenantAndNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantLayer2Network{{Key: 1, Tenant: 10, VNet: 100}})
		},
		"GET /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 1, Tenant: 10, VNet: 100, Enabled: true})
		},
	}))

	network, err := client.TenantLayer2Networks.GetByTenantAndNetwork(context.Background(), 10, 100)
	if err != nil {
		t.Fatalf("GetByTenantAndNetwork failed: %v", err)
	}
	if int(network.Key) != 1 {
		t.Errorf("expected key 1, got %d", int(network.Key))
	}
}

func TestTenantLayer2NetworkService_GetByTenantAndNetwork_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantLayer2Network{})
		},
	}))

	_, err := client.TenantLayer2Networks.GetByTenantAndNetwork(context.Background(), 10, 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			var req TenantLayer2NetworkCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Tenant != 10 {
				t.Errorf("expected tenant 10, got %d", req.Tenant)
			}
			if req.VNet != 100 {
				t.Errorf("expected vnet 100, got %d", req.VNet)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/tenant_layer2_vnets/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 5, Tenant: 10, VNet: 100, Enabled: true})
		},
	}))

	enabled := true
	network, err := client.TenantLayer2Networks.Create(context.Background(), &TenantLayer2NetworkCreateRequest{
		Tenant:  10,
		VNet:    100,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(network.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(network.Key))
	}
}

func TestTenantLayer2NetworkService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantLayer2Networks.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_Create_MissingTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantLayer2Networks.Create(context.Background(), &TenantLayer2NetworkCreateRequest{
		VNet: 100,
	})
	if err == nil {
		t.Fatal("expected error for missing tenant")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantLayer2Networks.Create(context.Background(), &TenantLayer2NetworkCreateRequest{
		Tenant: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_Update(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantLayer2NetworkUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled to be false")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 1, Tenant: 10, VNet: 100, Enabled: false})
		},
	}))

	enabled := false
	network, err := client.TenantLayer2Networks.Update(context.Background(), 1, &TenantLayer2NetworkUpdateRequest{Enabled: &enabled})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if network.Enabled {
		t.Error("expected enabled to be false")
	}
}

func TestTenantLayer2NetworkService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantLayer2Networks.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TenantLayer2Networks.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTenantLayer2NetworkService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenant_layer2_vnets/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.TenantLayer2Networks.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantLayer2NetworkService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantLayer2NetworkUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to be true")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 1, Enabled: true})
		},
	}))

	err := client.TenantLayer2Networks.Enable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestTenantLayer2NetworkService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantLayer2NetworkUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled to be false")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_layer2_vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 1, Enabled: false})
		},
	}))

	err := client.TenantLayer2Networks.Disable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

func TestTenantLayer2NetworkService_Assign(t *testing.T) {
	// Test assign when assignment doesn't exist yet
	callCount := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			// First call: GetByTenantAndNetwork returns empty (not found)
			// Second call (if any): also empty
			jsonResponse(w, 200, []TenantLayer2Network{})
		},
		"POST /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/tenant_layer2_vnets/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 5, Tenant: 10, VNet: 100, Enabled: true})
		},
	}))

	network, err := client.TenantLayer2Networks.Assign(context.Background(), 10, 100)
	if err != nil {
		t.Fatalf("Assign failed: %v", err)
	}
	if int(network.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(network.Key))
	}
	if callCount != 1 {
		t.Errorf("expected 1 create call, got %d", callCount)
	}
}

func TestTenantLayer2NetworkService_Assign_AlreadyExists(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantLayer2Network{{Key: 3, Tenant: 10, VNet: 100}})
		},
		"GET /api/v4/tenant_layer2_vnets/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 3, Tenant: 10, VNet: 100, Enabled: true})
		},
	}))

	network, err := client.TenantLayer2Networks.Assign(context.Background(), 10, 100)
	if err != nil {
		t.Fatalf("Assign failed: %v", err)
	}
	if int(network.Key) != 3 {
		t.Errorf("expected existing key 3, got %d", int(network.Key))
	}
}

func TestTenantLayer2NetworkService_Unassign(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantLayer2Network{{Key: 3, Tenant: 10, VNet: 100}})
		},
		"GET /api/v4/tenant_layer2_vnets/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantLayer2Network{Key: 3, Tenant: 10, VNet: 100})
		},
		"DELETE /api/v4/tenant_layer2_vnets/3": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TenantLayer2Networks.Unassign(context.Background(), 10, 100)
	if err != nil {
		t.Fatalf("Unassign failed: %v", err)
	}
}

func TestTenantLayer2NetworkService_Unassign_AlreadyUnassigned(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_layer2_vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantLayer2Network{})
		},
	}))

	err := client.TenantLayer2Networks.Unassign(context.Background(), 10, 999)
	if err != nil {
		t.Fatalf("expected no error for already unassigned, got: %v", err)
	}
}
