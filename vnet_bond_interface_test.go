package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVNetBondInterfaceDecode(t *testing.T) {
	const raw = `{"$key":8,"parent_vnet":3,"child_vnet":12,"primary":true}`
	var row VNetBondInterface
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatal(err)
	}
	if int(row.Key) != 8 || int(row.ParentVNet) != 3 || int(row.ChildVNet) != 12 || !row.Primary {
		t.Fatalf("row = %+v", row)
	}

	body, err := json.Marshal(VNetBondInterfaceUpdateRequest{Primary: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]any
	if err := json.Unmarshal(body, &keys); err != nil {
		t.Fatal(err)
	}
	if keys["primary"] != true {
		t.Fatalf("update body %s", body)
	}
	if _, ok := keys["parent_vnet"]; ok {
		t.Fatalf("update body includes parent_vnet: %s", body)
	}
	if _, ok := keys["child_vnet"]; ok {
		t.Fatalf("update body includes child_vnet: %s", body)
	}
}

func TestVNetBondInterfaceService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bond_interfaces": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); fields != vnetBondInterfaceListFields {
				t.Errorf("fields %s", fields)
			}
			if r.URL.Query().Get("filter") != "" {
				t.Errorf("filter %s", r.URL.Query().Get("filter"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"$key":"8","parent_vnet":"3","child_vnet":"12","primary":false}]`))
		},
	}))

	rows, err := client.VNetBondInterfaces.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(rows) != 1 || int(rows[0].Key) != 8 || int(rows[0].ParentVNet) != 3 || int(rows[0].ChildVNet) != 12 || rows[0].Primary {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestVNetBondInterfaceService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bond_interfaces": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != vnetBondInterfaceListFields+","+vnetBondInterface262Fields {
				t.Errorf("fields %s", fields)
			}
			if r.URL.Query().Get("filter") != "parent_vnet eq 3" {
				t.Errorf("filter %s", r.URL.Query().Get("filter"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"$key":8,"parent_vnet":3,"child_vnet":12,"primary":true}]`))
		},
	}))
	client.serverVersion = "26.2.2.1"

	rows, err := client.VNetBondInterfaces.ListByNetwork(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(rows) != 1 || int(rows[0].ParentVNet) != 3 || int(rows[0].ChildVNet) != 12 || !rows[0].Primary {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestVNetBondInterfaceService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bond_interfaces/8": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); fields != vnetBondInterfaceListFields {
				t.Errorf("fields %s", fields)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"$key":8,"parent_vnet":3,"child_vnet":12,"primary":true}`))
		},
	}))

	row, err := client.VNetBondInterfaces.Get(context.Background(), 8)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if int(row.Key) != 8 || int(row.ParentVNet) != 3 || int(row.ChildVNet) != 12 || !row.Primary {
		t.Fatalf("row = %+v", row)
	}
}
