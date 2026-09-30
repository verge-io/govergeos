package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestNodeLLDPNeighbor_Names(t *testing.T) {
	raw := []byte(`{
		"$key": 1,
		"node": 5,
		"nic": 42,
		"rid": "1",
		"via": "LLDP",
		"age": "0 day, 00:05:30",
		"chassis": {"name": "switch01.example.com", "ChassisID": "aa:bb:cc:dd:ee:ff"},
		"port": {"PortID": "Ethernet1/1", "descr": "Server Port 1"},
		"vlan": {"vlan-id": 100, "pvid": "yes"},
		"other": {"mfs": 9216}
	}`)
	var neighbor NodeLLDPNeighbor
	if err := json.Unmarshal(raw, &neighbor); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if neighbor.Node != 5 || neighbor.NIC != 42 || neighbor.RemoteID != "1" || neighbor.Via != "LLDP" {
		t.Fatalf("unexpected neighbor: %+v", neighbor)
	}
	if neighbor.ChassisName() != "switch01.example.com" {
		t.Errorf("chassis name %q", neighbor.ChassisName())
	}
	if neighbor.PortID() != "Ethernet1/1" {
		t.Errorf("port id %q", neighbor.PortID())
	}
	if neighbor.VLAN["vlan-id"] != float64(100) {
		t.Errorf("vlan %#v", neighbor.VLAN["vlan-id"])
	}

	fallback := NodeLLDPNeighbor{Chassis: JSONObject{"ChassisID": "aa:bb:cc:dd:ee:ff"}}
	if fallback.ChassisName() != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("fallback chassis name %q", fallback.ChassisName())
	}
	if (NodeLLDPNeighbor{}).ChassisName() != "" || (NodeLLDPNeighbor{}).PortID() != "" {
		t.Fatal("empty neighbor should have no chassis or port id")
	}
}

func TestNodeLLDPNeighborService_ListByNode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_lldp_neighbors": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != nodeLLDPNeighborListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []NodeLLDPNeighbor{{
				Key: 1, Node: 5, NIC: 42, Via: "LLDP",
				Chassis: JSONObject{"name": "switch01.example.com"},
				Port:    JSONObject{"PortID": "Ethernet1/1"},
			}})
		},
	}))

	neighbors, err := client.NodeLLDPNeighbors.ListByNode(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByNode failed: %v", err)
	}
	if len(neighbors) != 1 {
		t.Fatalf("expected 1 neighbor, got %d", len(neighbors))
	}
	if neighbors[0].ChassisName() != "switch01.example.com" || neighbors[0].PortID() != "Ethernet1/1" {
		t.Fatalf("unexpected neighbor: %+v", neighbors[0])
	}
}

func TestNodeLLDPNeighborService_ListByNIC(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_lldp_neighbors": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node eq 5 and nic eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeLLDPNeighbor{})
		},
	}))

	neighbors, err := client.NodeLLDPNeighbors.ListByNIC(context.Background(), 42, WithFilter("node eq 5"))
	if err != nil {
		t.Fatalf("ListByNIC failed: %v", err)
	}
	if len(neighbors) != 0 {
		t.Fatalf("expected empty list, got %d", len(neighbors))
	}
}

func TestNodeLLDPNeighborService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_lldp_neighbors/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.NodeLLDPNeighbors.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
