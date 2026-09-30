package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTenantExternalIP_TenantKey(t *testing.T) {
	if got := (TenantExternalIP{Owner: "tenants/3"}).TenantKey(); got != 3 {
		t.Fatalf("expected tenant 3, got %d", got)
	}
	if got := (TenantExternalIP{Owner: "tenants/nope"}).TenantKey(); got != 0 {
		t.Fatalf("expected tenant 0, got %d", got)
	}
}

func TestTenantExternalIPService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "type eq 'virtual' and owner bw 'tenants/'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != tenantExternalIPListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []TenantExternalIP{
				{Key: 1, VNet: 10, IP: "203.0.113.50", Type: "virtual", Owner: "tenants/7"},
			})
		},
	}))

	addresses, err := client.TenantExternalIPs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(addresses) != 1 || addresses[0].IP != "203.0.113.50" || addresses[0].TenantKey() != 7 {
		t.Fatalf("unexpected addresses: %+v", addresses)
	}
}

func TestTenantExternalIPService_ListByTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "type eq 'virtual' and owner bw 'tenants/' and owner eq 'tenants/7'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantExternalIP{{Key: 1, IP: "203.0.113.50", Type: "virtual", Owner: "tenants/7"}})
		},
	}))

	addresses, err := client.TenantExternalIPs.ListByTenant(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(addresses) != 1 {
		t.Fatalf("expected 1 address, got %d", len(addresses))
	}
}

func TestTenantExternalIPService_Get_NotTenantVirtual(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantExternalIP{Key: 8, VNet: 10, IP: "192.168.1.20", Type: "static", Owner: "vnets/10"})
		},
	}))

	_, err := client.TenantExternalIPs.Get(context.Background(), 8)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestTenantExternalIPService_GetByTenantAndIP(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "type eq 'virtual' and owner bw 'tenants/' and ip eq '203.0.113.50' and owner eq 'tenants/7'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantExternalIP{{Key: 9, IP: "203.0.113.50", Type: "virtual", Owner: "tenants/7"}})
		},
		"GET /api/v4/vnet_addresses/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantExternalIP{Key: 9, VNet: 10, IP: "203.0.113.50", Type: "virtual", Owner: "tenants/7", NetworkName: "external"})
		},
	}))

	address, err := client.TenantExternalIPs.GetByTenantAndIP(context.Background(), 7, "203.0.113.50")
	if err != nil {
		t.Fatalf("GetByTenantAndIP failed: %v", err)
	}
	if int(address.Key) != 9 || address.NetworkName != "external" {
		t.Fatalf("unexpected address: %+v", address)
	}
}

func TestTenantExternalIPService_Create_PendingFirewall(t *testing.T) {
	var applied bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["vnet"] != float64(10) || body["ip"] != "203.0.113.50" || body["type"] != "virtual" || body["owner"] != "tenants/7" {
				t.Fatalf("unexpected body: %+v", body)
			}
			if body["hostname"] != "edge" {
				t.Fatalf("expected hostname, got %+v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 9})
		},
		"GET /api/v4/vnet_addresses/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantExternalIP{Key: 9, VNet: 10, IP: "203.0.113.50", Type: "virtual", Owner: "tenants/7"})
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedFWApply: true})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			applied = true
			w.WriteHeader(http.StatusOK)
		},
	}))

	address, status, err := client.TenantExternalIPs.Create(context.Background(), &TenantExternalIPCreateRequest{
		Tenant:   7,
		VNet:     10,
		IP:       "203.0.113.50",
		Hostname: "edge",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if address.TenantKey() != 7 || address.IP != "203.0.113.50" {
		t.Fatalf("unexpected address: %+v", address)
	}
	if status == nil || status.NetworkID != 10 || !status.Pending || status.Applied {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
	if applied {
		t.Fatal("rules were applied without WithApplyParentFirewall")
	}
}

func TestTenantExternalIPService_Create_ApplyFirewall(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 9})
		},
		"GET /api/v4/vnet_addresses/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantExternalIP{Key: 9, VNet: 10, IP: "203.0.113.50", Type: "virtual", Owner: "tenants/7"})
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

	_, status, err := client.TenantExternalIPs.Create(context.Background(), &TenantExternalIPCreateRequest{
		Tenant: 7,
		VNet:   10,
		IP:     "203.0.113.50",
	}, WithApplyParentFirewall())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if status == nil || !status.Applied || status.Pending {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
}

func TestTenantExternalIPService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	ctx := context.Background()

	cases := []struct {
		name string
		req  *TenantExternalIPCreateRequest
	}{
		{name: "nil", req: nil},
		{name: "tenant", req: &TenantExternalIPCreateRequest{VNet: 10, IP: "203.0.113.50"}},
		{name: "vnet", req: &TenantExternalIPCreateRequest{Tenant: 7, IP: "203.0.113.50"}},
		{name: "ip", req: &TenantExternalIPCreateRequest{Tenant: 7, VNet: 10, IP: "not-an-ip"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := client.TenantExternalIPs.Create(ctx, tc.req)
			if !IsValidationError(err) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

func TestTenantExternalIPService_Delete_PendingFirewall(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantExternalIP{Key: 9, VNet: 10, Type: "virtual", Owner: "tenants/7", IP: "203.0.113.50"})
		},
		"DELETE /api/v4/vnet_addresses/9": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedFWApply: true})
		},
	}))

	status, err := client.TenantExternalIPs.Delete(context.Background(), 9)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if status == nil || status.NetworkID != 10 || !status.Pending || status.Applied {
		t.Fatalf("unexpected firewall status: %+v", status)
	}
}

func TestTenantExternalIPService_Delete_NotExternal(t *testing.T) {
	var deleted bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantExternalIP{Key: 8, Type: "dynamic", Owner: "vnets/10", IP: "192.168.1.20"})
		},
		"DELETE /api/v4/vnet_addresses/8": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(http.StatusOK)
		},
	}))

	_, err := client.TenantExternalIPs.Delete(context.Background(), 8)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if deleted {
		t.Fatal("deleted an address that is not a tenant external IP")
	}
}
