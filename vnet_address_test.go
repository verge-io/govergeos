package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVNetAddressService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetAddress{
				{Key: 1, VNet: 10, IP: "192.168.1.100", Type: "static"},
				{Key: 2, VNet: 10, IP: "192.168.1.101", Type: "dynamic"},
			})
		},
	}))

	addresses, err := client.VNetAddresses.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(addresses))
	}
	if addresses[0].IP != "192.168.1.100" {
		t.Errorf("expected IP '192.168.1.100', got %q", addresses[0].IP)
	}
}

func TestVNetAddressService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetAddress{{Key: 1, VNet: 10, IP: "192.168.1.100"}})
		},
	}))

	addresses, err := client.VNetAddresses.ListByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(addresses) != 1 {
		t.Fatalf("expected 1 address, got %d", len(addresses))
	}
}

func TestVNetAddressService_ListByType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10 and type eq 'static'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetAddress{{Key: 1, VNet: 10, Type: "static"}})
		},
	}))

	addresses, err := client.VNetAddresses.ListByType(context.Background(), 10, "static")
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}
	if len(addresses) != 1 {
		t.Fatalf("expected 1 address, got %d", len(addresses))
	}
}

func TestVNetAddressService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetAddress{Key: 1, VNet: 10, IP: "192.168.1.100", MAC: "aa:bb:cc:dd:ee:ff", Type: "static"})
		},
	}))

	addr, err := client.VNetAddresses.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if addr.IP != "192.168.1.100" {
		t.Errorf("expected IP '192.168.1.100', got %q", addr.IP)
	}
	if addr.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC 'aa:bb:cc:dd:ee:ff', got %q", addr.MAC)
	}
}

func TestVNetAddressService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetAddresses.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_GetByIP(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetAddress{{Key: 1, VNet: 10, IP: "192.168.1.100"}})
		},
		"GET /api/v4/vnet_addresses/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetAddress{Key: 1, VNet: 10, IP: "192.168.1.100", Type: "static"})
		},
	}))

	addr, err := client.VNetAddresses.GetByIP(context.Background(), 10, "192.168.1.100")
	if err != nil {
		t.Fatalf("GetByIP failed: %v", err)
	}
	if addr.Type != "static" {
		t.Errorf("expected type 'static', got %q", addr.Type)
	}
}

func TestVNetAddressService_GetByIP_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetAddress{})
		},
	}))

	_, err := client.VNetAddresses.GetByIP(context.Background(), 10, "10.0.0.99")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_GetByMAC(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetAddress{{Key: 1, VNet: 10, MAC: "aa:bb:cc:dd:ee:ff"}})
		},
		"GET /api/v4/vnet_addresses/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetAddress{Key: 1, VNet: 10, MAC: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.100"})
		},
	}))

	addr, err := client.VNetAddresses.GetByMAC(context.Background(), 10, "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("GetByMAC failed: %v", err)
	}
	if addr.IP != "192.168.1.100" {
		t.Errorf("expected IP '192.168.1.100', got %q", addr.IP)
	}
}

func TestVNetAddressService_GetByMAC_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetAddress{})
		},
	}))

	_, err := client.VNetAddresses.GetByMAC(context.Background(), 10, "00:00:00:00:00:00")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_addresses": func(w http.ResponseWriter, r *http.Request) {
			var req VNetAddressCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			if req.Type != "static" {
				t.Errorf("expected type 'static', got %q", req.Type)
			}
			jsonResponse(w, 200, map[string]any{"$key": 5, "status": "ok"})
		},
		"GET /api/v4/vnet_addresses/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetAddress{Key: 5, VNet: 10, IP: "192.168.1.200", Type: "static"})
		},
	}))

	addr, err := client.VNetAddresses.Create(context.Background(), &VNetAddressCreateRequest{
		VNet: 10,
		IP:   "192.168.1.200",
		Type: "static",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(addr.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(addr.Key))
	}
}

func TestVNetAddressService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetAddresses.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetAddresses.Create(context.Background(), &VNetAddressCreateRequest{Type: "static"})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_Create_MissingType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetAddresses.Create(context.Background(), &VNetAddressCreateRequest{VNet: 10})
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_Update(t *testing.T) {
	newIP := "192.168.1.201"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_addresses/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetAddressUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.IP == nil || *req.IP != newIP {
				t.Errorf("expected IP %q, got %v", newIP, req.IP)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_addresses/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetAddress{Key: 1, IP: newIP})
		},
	}))

	addr, err := client.VNetAddresses.Update(context.Background(), 1, &VNetAddressUpdateRequest{IP: &newIP})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if addr.IP != newIP {
		t.Errorf("expected IP %q, got %q", newIP, addr.IP)
	}
}

func TestVNetAddressService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetAddresses.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetAddressService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_addresses/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetAddresses.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetAddressService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_addresses/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetAddresses.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
