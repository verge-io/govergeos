package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestNodeService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nodes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Node{
				{ID: 1, Name: "node1", Physical: true},
				{ID: 2, Name: "node2", Physical: false},
			})
		},
	}))

	nodes, err := client.Nodes.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Name != "node1" {
		t.Errorf("expected name 'node1', got %q", nodes[0].Name)
	}
}

func TestNodeService_ListPhysical(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nodes": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for physical nodes")
			}
			jsonResponse(w, 200, []Node{
				{ID: 1, Name: "node1", Physical: true},
			})
		},
	}))

	nodes, err := client.Nodes.ListPhysical(context.Background())
	if err != nil {
		t.Fatalf("ListPhysical failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if !nodes[0].Physical {
		t.Error("expected physical node")
	}
}

func TestNodeService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nodes/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Node{ID: 1, Name: "node1", Cores: 16, Physical: true})
		},
	}))

	node, err := client.Nodes.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if node.Name != "node1" {
		t.Errorf("expected name 'node1', got %q", node.Name)
	}
	if node.Cores != 16 {
		t.Errorf("expected 16 cores, got %d", node.Cores)
	}
}

func TestNodeService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nodes/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Nodes.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNodeService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nodes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Node{{ID: 1, Name: "node1", Physical: true}})
		},
	}))

	node, err := client.Nodes.GetByName(context.Background(), "node1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if node.Name != "node1" {
		t.Errorf("expected name 'node1', got %q", node.Name)
	}
}

func TestNodeService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nodes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Node{})
		},
	}))

	_, err := client.Nodes.GetByName(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestNodeService_EnableMaintenance(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/nodes/1/enable_maintenance": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("expected empty body, got %v", body)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Nodes.EnableMaintenance(context.Background(), 1)
	if err != nil {
		t.Fatalf("EnableMaintenance failed: %v", err)
	}
}

func TestNodeService_DisableMaintenance(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/nodes/1/disable_maintenance": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("expected empty body, got %v", body)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Nodes.DisableMaintenance(context.Background(), 1)
	if err != nil {
		t.Fatalf("DisableMaintenance failed: %v", err)
	}
}

func TestNodeService_MaintenanceReboot(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/node_actions": func(w http.ResponseWriter, r *http.Request) {
			var body nodeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Action != "maintenance_reboot" {
				t.Errorf("expected action 'maintenance_reboot', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Nodes.MaintenanceReboot(context.Background(), 1)
	if err != nil {
		t.Fatalf("MaintenanceReboot failed: %v", err)
	}
}

func TestNodeService_ClearPStore(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/node_queries": func(w http.ResponseWriter, r *http.Request) {
			var body nodeQueryRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Node != 1 {
				t.Errorf("expected node 1, got %d", body.Node)
			}
			if body.Query != "clear-pstore" {
				t.Errorf("expected query 'clear-pstore', got %q", body.Query)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Nodes.ClearPStore(context.Background(), 1)
	if err != nil {
		t.Fatalf("ClearPStore failed: %v", err)
	}
}
