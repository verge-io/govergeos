package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVNetProxyService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); fields != vnetProxyListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []VNetProxy{{
				Key: 1, VNet: 10, NetworkName: "external", ListenAddress: "0.0.0.0", DefaultSelf: true,
			}})
		},
	}))

	proxies, err := client.VNetProxies.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(proxies) != 1 || proxies[0].NetworkName != "external" || int(proxies[0].VNet) != 10 {
		t.Fatalf("unexpected proxies: %+v", proxies)
	}
}

func TestVNetProxyService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "vnet eq 10 and name eq 'edge'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetProxy{{Key: 1, VNet: 10}})
		},
	}))

	proxies, err := client.VNetProxies.ListByNetwork(context.Background(), 10, WithFilter("name eq 'edge'"))
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(proxies) != 1 {
		t.Fatalf("expected 1 proxy, got %d", len(proxies))
	}
}

func TestVNetProxyService_GetByNetwork_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxy{{Key: 1, VNet: 10}, {Key: 2, VNet: 10}})
		},
	}))

	_, err := client.VNetProxies.GetByNetwork(context.Background(), 10)
	if !IsAmbiguousNameError(err) {
		t.Fatalf("expected AmbiguousNameError, got %v", err)
	}
}

func TestVNetProxyService_Create_Defaults(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxy{})
		},
		"POST /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["vnet"] != float64(10) || body["listen_address"] != "0.0.0.0" || body["default_self"] != true {
				t.Fatalf("unexpected body: %+v", body)
			}
			if _, ok := body["name"]; ok {
				t.Fatalf("empty name should be omitted: %+v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 4})
		},
		"GET /api/v4/vnet_proxy/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxy{Key: 4, VNet: 10, ListenAddress: "0.0.0.0", DefaultSelf: true})
		},
	}))

	proxy, err := client.VNetProxies.Create(context.Background(), &VNetProxyCreateRequest{VNet: 10})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(proxy.Key) != 4 || !proxy.DefaultSelf {
		t.Fatalf("unexpected proxy: %+v", proxy)
	}
}

func TestVNetProxyService_Create_AlreadyExists(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxy{{Key: 4, VNet: 10}})
		},
		"POST /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("POST must not run when a proxy already exists")
		},
	}))

	_, err := client.VNetProxies.Create(context.Background(), &VNetProxyCreateRequest{VNet: 10})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestVNetProxyService_Create_InvalidListenAddress(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.VNetProxies.Create(context.Background(), &VNetProxyCreateRequest{
		VNet:          10,
		ListenAddress: "everywhere",
	})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestVNetProxyService_GetOrCreate_Existing(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxy{{Key: 4, VNet: 10, ListenAddress: "0.0.0.0"}})
		},
		"GET /api/v4/vnet_proxy/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxy{Key: 4, VNet: 10, ListenAddress: "0.0.0.0", DefaultSelf: true})
		},
		"POST /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("POST must not run when a proxy already exists")
		},
	}))

	proxy, err := client.VNetProxies.GetOrCreate(context.Background(), &VNetProxyCreateRequest{VNet: 10})
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}
	if int(proxy.Key) != 4 {
		t.Fatalf("unexpected proxy: %+v", proxy)
	}
}

func TestVNetProxyService_Update(t *testing.T) {
	listen := "192.0.2.10"
	defaultSelf := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_proxy/4": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["listen_address"] != listen || body["default_self"] != false {
				t.Fatalf("unexpected body: %+v", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vnet_proxy/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxy{Key: 4, VNet: 10, ListenAddress: listen, DefaultSelf: false})
		},
	}))

	proxy, err := client.VNetProxies.Update(context.Background(), 4, &VNetProxyUpdateRequest{
		ListenAddress: &listen,
		DefaultSelf:   &defaultSelf,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if proxy.ListenAddress != listen || proxy.DefaultSelf {
		t.Fatalf("unexpected proxy: %+v", proxy)
	}
}

func TestVNetProxyService_DeleteByNetwork(t *testing.T) {
	deleted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxy{{Key: 4, VNet: 10}})
		},
		"GET /api/v4/vnet_proxy/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetProxy{Key: 4, VNet: 10})
		},
		"DELETE /api/v4/vnet_proxy/4": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(http.StatusOK)
		},
	}))

	if err := client.VNetProxies.DeleteByNetwork(context.Background(), 10); err != nil {
		t.Fatalf("DeleteByNetwork failed: %v", err)
	}
	if !deleted {
		t.Fatal("proxy was not deleted")
	}
}

func TestVNetProxyService_Exists(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_proxy": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetProxy{})
		},
	}))

	exists, err := client.VNetProxies.Exists(context.Background(), 10)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected no proxy")
	}
}
