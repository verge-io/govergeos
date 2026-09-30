package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestNodeMemory_IsHealthy(t *testing.T) {
	if !(NodeMemory{Status: "online"}).IsHealthy() {
		t.Fatal("online DIMM should be healthy")
	}
	for _, status := range []string{"error", "warning", "offline", ""} {
		if (NodeMemory{Status: status}).IsHealthy() {
			t.Fatalf("status %q should not be healthy", status)
		}
	}
}

func TestNodeMemoryService_ListByNode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_memory": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != nodeMemoryListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []NodeMemory{{
				Key: 1, Node: 1, Status: "online", Locator: "DIMM 0",
				BankLocator: "P0 CHANNEL A", Type: "DDR5", Size: "48 GB",
				Speed: "5600 MT/s", Manufacturer: "Micron Technology",
			}})
		},
	}))

	dimms, err := client.NodeMemory.ListByNode(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByNode failed: %v", err)
	}
	if len(dimms) != 1 {
		t.Fatalf("expected 1 DIMM, got %d", len(dimms))
	}
	if dimms[0].Locator != "DIMM 0" || dimms[0].Type != "DDR5" || !dimms[0].IsHealthy() {
		t.Fatalf("unexpected dimm: %+v", dimms[0])
	}
}

func TestNodeMemoryService_ListUnhealthy(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_memory": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "status ne 'online'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeMemory{{
				Key: 5, Node: 2, Status: "error", StatusInfo: "Correctable ECC errors detected",
				Locator: "DIMM 2",
			}})
		},
	}))

	dimms, err := client.NodeMemory.ListUnhealthy(context.Background())
	if err != nil {
		t.Fatalf("ListUnhealthy failed: %v", err)
	}
	if len(dimms) != 1 || dimms[0].IsHealthy() || dimms[0].StatusInfo == "" {
		t.Fatalf("unexpected dimms: %+v", dimms)
	}
}

func TestNodeMemoryService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_memory/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NodeMemory{Key: 1, Locator: "DIMM 0", FormFactor: "SODIMM", PartNumber: "CT48G56C46S5.M16C1"})
		},
	}))

	dimm, err := client.NodeMemory.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if dimm.PartNumber != "CT48G56C46S5.M16C1" || dimm.FormFactor != "SODIMM" {
		t.Fatalf("unexpected dimm: %+v", dimm)
	}
}

func TestNodeMemoryService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_memory/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.NodeMemory.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
