package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// --- List ---

func TestNetworkService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Network{
				{ID: 1, Name: "internal", Type: "internal"},
				{ID: 2, Name: "external", Type: "external"},
			})
		},
	}))

	networks, err := client.Networks.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(networks) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(networks))
	}
	if networks[0].Name != "internal" {
		t.Errorf("expected name 'internal', got %q", networks[0].Name)
	}
}

func TestNetworkService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter parameter")
			}
			jsonResponse(w, 200, []Network{{ID: 1, Name: "internal", Type: "internal"}})
		},
	}))

	networks, err := client.Networks.List(context.Background(), WithFilter("type eq 'internal'"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(networks))
	}
}

func TestNetworkService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Network{})
		},
	}))

	networks, err := client.Networks.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(networks) != 0 {
		t.Fatalf("expected 0 networks, got %d", len(networks))
	}
}

// --- Get ---

func TestNetworkService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{
				ID:        1,
				Name:      "internal",
				Type:      "internal",
				Enabled:   true,
				IPAddress: "10.0.0.1",
			})
		},
	}))

	net, err := client.Networks.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if net.Name != "internal" {
		t.Errorf("expected name 'internal', got %q", net.Name)
	}
	if net.IPAddress != "10.0.0.1" {
		t.Errorf("expected IP '10.0.0.1', got %q", net.IPAddress)
	}
}

func TestNetworkService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Networks.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Create ---

func TestNetworkService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "test-net" {
				t.Errorf("expected name 'test-net', got %q", req.Name)
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: 42})
		},
		"GET /api/v4/vnets/42": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{ID: 42, Name: "test-net", Enabled: true})
		},
	}))

	net, err := client.Networks.Create(context.Background(), &NetworkCreateRequest{
		Name: "test-net",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(net.ID) != 42 {
		t.Errorf("expected ID 42, got %d", int(net.ID))
	}
	if net.Name != "test-net" {
		t.Errorf("expected name 'test-net', got %q", net.Name)
	}
}

func TestNetworkService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Networks.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNetworkService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Networks.Create(context.Background(), &NetworkCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// --- Update ---

func TestNetworkService_Update(t *testing.T) {
	newName := "renamed-net"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnets/1": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{ID: 1, Name: newName})
		},
	}))

	net, err := client.Networks.Update(context.Background(), 1, &NetworkUpdateRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if net.Name != newName {
		t.Errorf("expected name %q, got %q", newName, net.Name)
	}
}

func TestNetworkService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Networks.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNetworkService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnets/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "nope"
	_, err := client.Networks.Update(context.Background(), 999, &NetworkUpdateRequest{
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Delete ---

func TestNetworkService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnets/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Networks.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNetworkService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnets/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete returns nil for not-found (idempotent)
	err := client.Networks.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted network")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Kill ---

func TestNetworkService_Kill(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			var body vnetAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.VNet != 5 {
				t.Errorf("expected vnet 5, got %d", body.VNet)
			}
			if body.Action != "killpower" {
				t.Errorf("expected action 'killpower', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Networks.Kill(context.Background(), 5)
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}
}

// --- Reset ---

func TestNetworkService_Reset(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "reset" {
				t.Errorf("expected action 'reset', got %v", body["action"])
			}
			params, ok := body["params"].(map[string]any)
			if !ok {
				t.Fatal("expected params to be a map")
			}
			if params["apply"] != true {
				t.Errorf("expected apply=true, got %v", params["apply"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Networks.Reset(context.Background(), 3, true)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestNetworkService_Reset_NoApply(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			params, _ := body["params"].(map[string]any)
			if params["apply"] != false {
				t.Errorf("expected apply=false, got %v", params["apply"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Networks.Reset(context.Background(), 3, false)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

// --- ApplyRules ---

func TestNetworkService_ApplyRules(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			var body vnetAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.VNet != 7 {
				t.Errorf("expected vnet 7, got %d", body.VNet)
			}
			if body.Action != "refresh" {
				t.Errorf("expected action 'refresh', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Networks.ApplyRules(context.Background(), 7)
	if err != nil {
		t.Fatalf("ApplyRules failed: %v", err)
	}
}

// --- ApplyDNS ---

func TestNetworkService_ApplyDNS(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			var body vnetAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.VNet != 7 {
				t.Errorf("expected vnet 7, got %d", body.VNet)
			}
			if body.Action != "applydns" {
				t.Errorf("expected action 'applydns', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Networks.ApplyDNS(context.Background(), 7)
	if err != nil {
		t.Fatalf("ApplyDNS failed: %v", err)
	}
}

// --- RunQuery ---

func TestNetworkService_RunQuery(t *testing.T) {
	queryID := "abc123def456"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_queries": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkQueryRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			if req.Query != "ping" {
				t.Errorf("expected query 'ping', got %q", req.Query)
			}
			jsonResponse(w, 200, map[string]string{"id": queryID})
		},
		"GET /api/v4/vnet_queries/": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     queryID,
				VNet:   10,
				Query:  "ping",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	q, err := client.Networks.RunQuery(context.Background(), &NetworkQueryRequest{
		VNet:  10,
		Query: NetworkQueryPing,
	})
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}
	if q.ID != queryID {
		t.Errorf("expected query ID %q, got %q", queryID, q.ID)
	}
}

func TestNetworkService_RunQuery_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Networks.RunQuery(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNetworkService_RunQuery_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Networks.RunQuery(context.Background(), &NetworkQueryRequest{
		Query: "ping",
	})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestNetworkService_RunQuery_MissingQuery(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Networks.RunQuery(context.Background(), &NetworkQueryRequest{
		VNet: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing query")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// --- GetQuery ---

func TestNetworkService_GetQuery(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_queries/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     "abc123",
				VNet:   10,
				Query:  "ping",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	q, err := client.Networks.GetQuery(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("GetQuery failed: %v", err)
	}
	if q.Status != NetworkQueryStatusComplete {
		t.Errorf("expected status 'complete', got %q", q.Status)
	}
}

func TestNetworkService_GetQuery_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_queries/nope": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Networks.GetQuery(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Ping ---

func TestNetworkService_Ping(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_queries": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkQueryRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Query != "ping" {
				t.Errorf("expected query 'ping', got %q", req.Query)
			}
			if req.Params["address"] != "8.8.8.8" {
				t.Errorf("expected address '8.8.8.8', got %v", req.Params["address"])
			}
			jsonResponse(w, 200, map[string]string{"id": "ping-001"})
		},
		"GET /api/v4/vnet_queries/": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     "ping-001",
				VNet:   1,
				Query:  "ping",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	q, err := client.Networks.Ping(context.Background(), 1, "8.8.8.8", 4)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if q.Query != "ping" {
		t.Errorf("expected query 'ping', got %q", q.Query)
	}
}

func TestNetworkService_Ping_DefaultCount(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_queries": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkQueryRequest
			json.NewDecoder(r.Body).Decode(&req)
			count, ok := req.Params["count"].(float64) // JSON numbers are float64
			if !ok || int(count) != 4 {
				t.Errorf("expected default count 4, got %v", req.Params["count"])
			}
			jsonResponse(w, 200, map[string]string{"id": "ping-002"})
		},
		"GET /api/v4/vnet_queries/": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     "ping-002",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	_, err := client.Networks.Ping(context.Background(), 1, "10.0.0.1", 0)
	if err != nil {
		t.Fatalf("Ping with default count failed: %v", err)
	}
}

// --- Traceroute ---

func TestNetworkService_Traceroute(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_queries": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkQueryRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Query != "traceroute" {
				t.Errorf("expected query 'traceroute', got %q", req.Query)
			}
			if req.Params["address"] != "1.1.1.1" {
				t.Errorf("expected address '1.1.1.1', got %v", req.Params["address"])
			}
			jsonResponse(w, 200, map[string]string{"id": "trace-001"})
		},
		"GET /api/v4/vnet_queries/": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     "trace-001",
				Query:  "traceroute",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	q, err := client.Networks.Traceroute(context.Background(), 1, "1.1.1.1")
	if err != nil {
		t.Fatalf("Traceroute failed: %v", err)
	}
	if q.Query != "traceroute" {
		t.Errorf("expected query 'traceroute', got %q", q.Query)
	}
}

// --- DNSLookup ---

func TestNetworkService_DNSLookup(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_queries": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkQueryRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Query != "dns" {
				t.Errorf("expected query 'dns', got %q", req.Query)
			}
			if req.Params["hostname"] != "example.com" {
				t.Errorf("expected hostname 'example.com', got %v", req.Params["hostname"])
			}
			jsonResponse(w, 200, map[string]string{"id": "dns-001"})
		},
		"GET /api/v4/vnet_queries/": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     "dns-001",
				Query:  "dns",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	q, err := client.Networks.DNSLookup(context.Background(), 1, "example.com")
	if err != nil {
		t.Fatalf("DNSLookup failed: %v", err)
	}
	if q.Query != "dns" {
		t.Errorf("expected query 'dns', got %q", q.Query)
	}
}

// --- GetDiagnostics ---

func TestNetworkService_GetDiagnostics(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_queries": func(w http.ResponseWriter, r *http.Request) {
			var req NetworkQueryRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Query != "whatsmyip" {
				t.Errorf("expected query 'whatsmyip', got %q", req.Query)
			}
			jsonResponse(w, 200, map[string]string{"id": "diag-001"})
		},
		"GET /api/v4/vnet_queries/": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NetworkQuery{
				ID:     "diag-001",
				Query:  "whatsmyip",
				Status: NetworkQueryStatusComplete,
			})
		},
	}))

	q, err := client.Networks.GetDiagnostics(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetDiagnostics failed: %v", err)
	}
	if q.Query != "whatsmyip" {
		t.Errorf("expected query 'whatsmyip', got %q", q.Query)
	}
}

// --- GetStatistics ---

func TestNetworkService_GetStatistics(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_monitor_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 1" {
				t.Errorf("expected filter 'vnet eq 1', got %q", filter)
			}
			jsonResponse(w, 200, []NetworkMonitorStats{
				{Key: 100, VNet: 1, Quality: 99},
				{Key: 101, VNet: 1, Quality: 95},
			})
		},
	}))

	stats, err := client.Networks.GetStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStatistics failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	if stats[0].Quality != 99 {
		t.Errorf("expected quality 99, got %d", stats[0].Quality)
	}
}

func TestNetworkService_GetStatistics_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_monitor_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NetworkMonitorStats{})
		},
	}))

	stats, err := client.Networks.GetStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStatistics failed: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected 0 stats, got %d", len(stats))
	}
}

// --- GetLatestStatistics ---

func TestNetworkService_GetLatestStatistics(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_monitor_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NetworkMonitorStats{
				{Key: 100, VNet: 1, Quality: 99},
				{Key: 101, VNet: 1, Quality: 95},
			})
		},
	}))

	stat, err := client.Networks.GetLatestStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLatestStatistics failed: %v", err)
	}
	if stat == nil {
		t.Fatal("expected non-nil stat")
	}
	if stat.Quality != 99 {
		t.Errorf("expected quality 99, got %d", stat.Quality)
	}
}

func TestNetworkService_GetLatestStatistics_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_monitor_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NetworkMonitorStats{})
		},
	}))

	stat, err := client.Networks.GetLatestStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLatestStatistics failed: %v", err)
	}
	if stat != nil {
		t.Errorf("expected nil for empty stats, got %+v", stat)
	}
}

// --- PowerOn ---

func TestNetworkService_PowerOn_AlreadyRunning(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{ID: 1, Name: "net", PowerState: true})
		},
	}))

	// Should return nil immediately since already running
	err := client.Networks.PowerOn(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOn (already running) failed: %v", err)
	}
}

// --- PowerOff ---

func TestNetworkService_PowerOff_AlreadyStopped(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{ID: 1, Name: "net", PowerState: false})
		},
	}))

	// Should return nil immediately since already stopped
	err := client.Networks.PowerOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOff (already stopped) failed: %v", err)
	}
}

// --- Action server errors ---

func TestNetworkService_Kill_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.Networks.Kill(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestNetworkService_ApplyRules_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.Networks.ApplyRules(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestNetworkService_ApplyDNS_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.Networks.ApplyDNS(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}
