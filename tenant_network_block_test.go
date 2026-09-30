package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTenantNetworkBlock_TenantKey(t *testing.T) {
	if got := (TenantNetworkBlock{Owner: "tenants/12"}).TenantKey(); got != 12 {
		t.Fatalf("expected tenant 12, got %d", got)
	}
	if got := (TenantNetworkBlock{Owner: "vnets/4"}).TenantKey(); got != 0 {
		t.Fatalf("expected tenant 0, got %d", got)
	}
}

func TestTenantNetworkBlockService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); fields != tenantNetworkBlockListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []TenantNetworkBlock{
				{Key: 1, VNet: 10, CIDR: "192.168.100.0/24", Owner: "tenants/7", NetworkName: "external"},
			})
		},
	}))

	blocks, err := client.TenantNetworkBlocks.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].CIDR != "192.168.100.0/24" || blocks[0].TenantKey() != 7 {
		t.Fatalf("unexpected block: %+v", blocks[0])
	}
}

func TestTenantNetworkBlockService_ListByTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "owner eq 'tenants/10'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantNetworkBlock{{Key: 1, Owner: "tenants/10", CIDR: "10.1.0.0/24"}})
		},
	}))

	blocks, err := client.TenantNetworkBlocks.ListByTenant(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(blocks) != 1 || blocks[0].TenantKey() != 10 {
		t.Fatalf("unexpected blocks: %+v", blocks)
	}
}

func TestTenantNetworkBlockService_ListByTenant_Invalid(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.TenantNetworkBlocks.ListByTenant(context.Background(), 0)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantNetworkBlockService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNetworkBlock{Key: 1, VNet: 10, CIDR: "192.168.100.0/24", Owner: "tenants/7"})
		},
	}))

	block, err := client.TenantNetworkBlocks.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if block.CIDR != "192.168.100.0/24" || int(block.VNet) != 10 {
		t.Fatalf("unexpected block: %+v", block)
	}
}

func TestTenantNetworkBlockService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs/999": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}))

	_, err := client.TenantNetworkBlocks.Get(context.Background(), 999)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestTenantNetworkBlockService_GetByTenantAndCIDR(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "cidr eq '192.168.100.0/24' and owner eq 'tenants/7'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantNetworkBlock{{Key: 4, Owner: "tenants/7", CIDR: "192.168.100.0/24"}})
		},
		"GET /api/v4/vnet_cidrs/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNetworkBlock{Key: 4, VNet: 10, Owner: "tenants/7", CIDR: "192.168.100.0/24"})
		},
	}))

	block, err := client.TenantNetworkBlocks.GetByTenantAndCIDR(context.Background(), 7, "192.168.100.0/24")
	if err != nil {
		t.Fatalf("GetByTenantAndCIDR failed: %v", err)
	}
	if int(block.Key) != 4 {
		t.Fatalf("expected key 4, got %d", int(block.Key))
	}
}

func TestTenantNetworkBlockService_Create_PendingFirewall(t *testing.T) {
	var applied bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_cidrs": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["vnet"] != float64(10) || body["cidr"] != "192.168.100.0/24" || body["owner"] != "tenants/7" {
				t.Fatalf("unexpected body: %+v", body)
			}
			if _, ok := body["tenant"]; ok {
				t.Fatal("tenant must be sent as owner, not its own field")
			}
			jsonResponse(w, 200, map[string]any{"$key": 5})
		},
		"GET /api/v4/vnet_cidrs/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNetworkBlock{Key: 5, VNet: 10, CIDR: "192.168.100.0/24", Owner: "tenants/7", NetworkName: "external"})
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedFWApply: true})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			applied = true
			w.WriteHeader(http.StatusOK)
		},
	}))

	block, status, err := client.TenantNetworkBlocks.Create(context.Background(), &TenantNetworkBlockCreateRequest{
		Tenant:      7,
		VNet:        10,
		CIDR:        "192.168.100.0/24",
		Description: "customer block",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(block.Key) != 5 || block.TenantKey() != 7 {
		t.Fatalf("unexpected block: %+v", block)
	}
	if status == nil || status.NetworkID != 10 || !status.Pending || status.Applied {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
	if applied {
		t.Fatal("rules were applied without WithApplyParentFirewall")
	}
}

func TestTenantNetworkBlockService_Create_ApplyFirewall(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_cidrs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 5})
		},
		"GET /api/v4/vnet_cidrs/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNetworkBlock{Key: 5, VNet: 10, CIDR: "192.168.100.0/24", Owner: "tenants/7"})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			var body vnetAction
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode action: %v", err)
			}
			if body.VNet != 10 || body.Action != "refresh" {
				t.Fatalf("unexpected action: %+v", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedFWApply: false})
		},
	}))

	_, status, err := client.TenantNetworkBlocks.Create(context.Background(), &TenantNetworkBlockCreateRequest{
		Tenant: 7,
		VNet:   10,
		CIDR:   "192.168.100.0/24",
	}, WithApplyParentFirewall())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if status == nil || !status.Applied || status.Pending || status.NetworkID != 10 {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
}

func TestTenantNetworkBlockService_Create_ApplyFailure(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_cidrs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 5})
		},
		"GET /api/v4/vnet_cidrs/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNetworkBlock{Key: 5, VNet: 10, CIDR: "192.168.100.0/24", Owner: "tenants/7"})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedFWApply: true})
		},
	}))

	block, status, err := client.TenantNetworkBlocks.Create(context.Background(), &TenantNetworkBlockCreateRequest{
		Tenant: 7,
		VNet:   10,
		CIDR:   "192.168.100.0/24",
	}, WithApplyParentFirewall())
	if err == nil {
		t.Fatal("expected apply error")
	}
	if block == nil || int(block.Key) != 5 {
		t.Fatalf("expected created block with the error, got %+v", block)
	}
	if status == nil || status.Applied || !status.Pending || status.NetworkID != 10 {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
}

func TestTenantNetworkBlockService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	ctx := context.Background()

	cases := []struct {
		name string
		req  *TenantNetworkBlockCreateRequest
	}{
		{name: "nil", req: nil},
		{name: "tenant", req: &TenantNetworkBlockCreateRequest{VNet: 10, CIDR: "192.168.100.0/24"}},
		{name: "vnet", req: &TenantNetworkBlockCreateRequest{Tenant: 7, CIDR: "192.168.100.0/24"}},
		{name: "cidr", req: &TenantNetworkBlockCreateRequest{Tenant: 7, VNet: 10, CIDR: "not-a-cidr"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := client.TenantNetworkBlocks.Create(ctx, tc.req)
			if !IsValidationError(err) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

func TestTenantNetworkBlockService_Delete_PendingFirewall(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNetworkBlock{Key: 5, VNet: 10, Owner: "tenants/7"})
		},
		"DELETE /api/v4/vnet_cidrs/5": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedFWApply: true})
		},
	}))

	status, err := client.TenantNetworkBlocks.Delete(context.Background(), 5)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if status == nil || status.NetworkID != 10 || !status.Pending || status.Applied {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
}

func TestTenantNetworkBlockService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_cidrs/999": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}))

	_, err := client.TenantNetworkBlocks.Delete(context.Background(), 999)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
