package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVNetProxyTenantService_ListByProxy(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy_tenants": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); fields != vnetProxyTenantListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			if filter := r.URL.Query().Get("filter"); filter != "proxy eq 3" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetProxyTenant{{
				Key: 8, Proxy: 3, Tenant: 7, TenantName: "customer-a", FQDN: "customer-a.example.com",
			}})
		},
	}))

	rows, err := client.VNetProxyTenants.ListByProxy(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListByProxy failed: %v", err)
	}
	if len(rows) != 1 || rows[0].FQDN != "customer-a.example.com" || rows[0].TenantName != "customer-a" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestVNetProxyTenantService_GetByFQDN(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy_tenants": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != `proxy eq 3 and fqdn eq 'customer-a.example.com'` {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetProxyTenant{{Key: 8, Proxy: 3, FQDN: "customer-a.example.com"}})
		},
		"GET /api/v4/vnet_proxy_tenants/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxyTenant{
				Key: 8, Proxy: 3, Tenant: 7, TenantName: "customer-a", FQDN: "customer-a.example.com",
			})
		},
	}))

	row, err := client.VNetProxyTenants.GetByFQDN(context.Background(), 3, "customer-a.example.com")
	if err != nil {
		t.Fatalf("GetByFQDN failed: %v", err)
	}
	if int(row.Tenant) != 7 || row.TenantName != "customer-a" {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestVNetProxyTenantService_GetByFQDN_EscapesQuote(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy_tenants": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != `proxy eq 3 and fqdn eq 'cust\'omer.example.com'` {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetProxyTenant{})
		},
	}))

	_, err := client.VNetProxyTenants.GetByFQDN(context.Background(), 3, "cust'omer.example.com")
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestVNetProxyTenantService_GetByTenant_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy_tenants": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxyTenant{
				{Key: 8, Proxy: 3, Tenant: 7, FQDN: "a.example.com"},
				{Key: 9, Proxy: 3, Tenant: 7, FQDN: "b.example.com"},
			})
		},
	}))

	_, err := client.VNetProxyTenants.GetByTenant(context.Background(), 3, 7)
	if !IsAmbiguousNameError(err) {
		t.Fatalf("expected AmbiguousNameError, got %v", err)
	}
}

func TestVNetProxyTenantService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_proxy_tenants": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["proxy"] != float64(3) || body["tenant"] != float64(7) || body["fqdn"] != "customer-a.example.com" {
				t.Fatalf("unexpected body: %+v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 8})
		},
		"GET /api/v4/vnet_proxy_tenants/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxyTenant{Key: 8, Proxy: 3, Tenant: 7, FQDN: "customer-a.example.com"})
		},
	}))

	row, err := client.VNetProxyTenants.Create(context.Background(), &VNetProxyTenantCreateRequest{
		Proxy:  3,
		Tenant: 7,
		FQDN:   " customer-a.example.com ",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if row.FQDN != "customer-a.example.com" || int(row.Key) != 8 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestVNetProxyTenantService_UpdateAndDelete(t *testing.T) {
	fqdn := "new.example.com"
	deleted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_proxy_tenants/8": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["fqdn"] != fqdn {
				t.Fatalf("unexpected body: %+v", body)
			}
			if _, ok := body["proxy"]; ok {
				t.Fatalf("proxy must stay on the row: %+v", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vnet_proxy_tenants/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxyTenant{Key: 8, Proxy: 3, Tenant: 7, FQDN: fqdn})
		},
		"DELETE /api/v4/vnet_proxy_tenants/8": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(http.StatusOK)
		},
	}))

	row, err := client.VNetProxyTenants.Update(context.Background(), 8, &VNetProxyTenantUpdateRequest{FQDN: &fqdn})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if row.FQDN != fqdn {
		t.Fatalf("unexpected row: %+v", row)
	}
	if err := client.VNetProxyTenants.Delete(context.Background(), 8); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !deleted {
		t.Fatal("mapping was not deleted")
	}
}

func TestVNetProxyTenantService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.VNetProxyTenants.Create(context.Background(), &VNetProxyTenantCreateRequest{Proxy: 3, Tenant: 7})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}
