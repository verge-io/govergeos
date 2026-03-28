package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVNetHostService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetHost{
				{Key: 1, VNet: 10, Host: "server1", IP: "192.168.1.10"},
				{Key: 2, VNet: 10, Host: "server2", IP: "192.168.1.11"},
			})
		},
	}))

	hosts, err := client.VNetHosts.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}
	if hosts[0].Host != "server1" {
		t.Errorf("expected host 'server1', got %q", hosts[0].Host)
	}
}

func TestVNetHostService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetHost{{Key: 1, VNet: 10, Host: "server1"}})
		},
	}))

	hosts, err := client.VNetHosts.ListByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
}

func TestVNetHostService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetHost{Key: 1, VNet: 10, Host: "server1", IP: "192.168.1.10", Type: "host"})
		},
	}))

	host, err := client.VNetHosts.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if host.Host != "server1" {
		t.Errorf("expected host 'server1', got %q", host.Host)
	}
	if host.IP != "192.168.1.10" {
		t.Errorf("expected IP '192.168.1.10', got %q", host.IP)
	}
}

func TestVNetHostService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetHosts.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetHostService_GetByHost(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetHost{{Key: 1, VNet: 10, Host: "server1"}})
		},
		"GET /api/v4/vnet_hosts/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetHost{Key: 1, VNet: 10, Host: "server1", IP: "192.168.1.10"})
		},
	}))

	host, err := client.VNetHosts.GetByHost(context.Background(), 10, "server1")
	if err != nil {
		t.Fatalf("GetByHost failed: %v", err)
	}
	if host.IP != "192.168.1.10" {
		t.Errorf("expected IP '192.168.1.10', got %q", host.IP)
	}
}

func TestVNetHostService_GetByHost_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetHost{})
		},
	}))

	_, err := client.VNetHosts.GetByHost(context.Background(), 10, "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetHostService_GetByIP(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetHost{{Key: 1, VNet: 10, IP: "192.168.1.10"}})
		},
		"GET /api/v4/vnet_hosts/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetHost{Key: 1, VNet: 10, Host: "server1", IP: "192.168.1.10"})
		},
	}))

	host, err := client.VNetHosts.GetByIP(context.Background(), 10, "192.168.1.10")
	if err != nil {
		t.Fatalf("GetByIP failed: %v", err)
	}
	if host.Host != "server1" {
		t.Errorf("expected host 'server1', got %q", host.Host)
	}
}

func TestVNetHostService_GetByIP_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetHost{})
		},
	}))

	_, err := client.VNetHosts.GetByIP(context.Background(), 10, "10.0.0.99")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetHostService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_hosts": func(w http.ResponseWriter, r *http.Request) {
			var req VNetHostCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			if req.Host != "db-server" {
				t.Errorf("expected host 'db-server', got %q", req.Host)
			}
			if req.IP != "192.168.1.50" {
				t.Errorf("expected IP '192.168.1.50', got %q", req.IP)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 5, "status": "ok"})
		},
		"GET /api/v4/vnet_hosts/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetHost{Key: 5, VNet: 10, Host: "db-server", IP: "192.168.1.50", Type: "host"})
		},
	}))

	host, err := client.VNetHosts.Create(context.Background(), &VNetHostCreateRequest{
		VNet: 10,
		Host: "db-server",
		IP:   "192.168.1.50",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(host.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(host.Key))
	}
}

func TestVNetHostService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetHosts.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetHostService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetHosts.Create(context.Background(), &VNetHostCreateRequest{Host: "test", IP: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetHostService_Create_MissingHost(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetHosts.Create(context.Background(), &VNetHostCreateRequest{VNet: 10, IP: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetHostService_Create_MissingIP(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetHosts.Create(context.Background(), &VNetHostCreateRequest{VNet: 10, Host: "test"})
	if err == nil {
		t.Fatal("expected error for missing ip")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetHostService_Update(t *testing.T) {
	newIP := "192.168.1.99"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_hosts/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetHostUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.IP == nil || *req.IP != newIP {
				t.Errorf("expected IP %q, got %v", newIP, req.IP)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_hosts/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetHost{Key: 1, Host: "server1", IP: newIP})
		},
	}))

	host, err := client.VNetHosts.Update(context.Background(), 1, &VNetHostUpdateRequest{IP: &newIP})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if host.IP != newIP {
		t.Errorf("expected IP %q, got %q", newIP, host.IP)
	}
}

func TestVNetHostService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetHosts.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetHostService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_hosts/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetHosts.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetHostService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_hosts/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetHosts.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
