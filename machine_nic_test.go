package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestMachineNICService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineNIC{
				{Key: 1, Machine: 10, Name: "eno1"},
				{Key: 2, Machine: 10, Name: "eno2"},
			})
		},
	}))

	nics, err := client.MachineNICs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nics) != 2 {
		t.Fatalf("expected 2 NICs, got %d", len(nics))
	}
	if nics[0].Name != "eno1" {
		t.Errorf("expected name 'eno1', got %q", nics[0].Name)
	}
}

func TestMachineNICService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineNIC{})
		},
	}))

	nics, err := client.MachineNICs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nics) != 0 {
		t.Fatalf("expected 0 NICs, got %d", len(nics))
	}
}

func TestMachineNICService_ListByMachine(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineNIC{
				{Key: 1, Machine: 10, Name: "eno1",
					Stats:  &MachineNICStats{TxBytes: 1048576, RxBytes: 2097152},
					Status: &MachineNICStatus{Status: "up", Speed: 10000},
				},
			})
		},
	}))

	nics, err := client.MachineNICs.ListByMachine(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByMachine failed: %v", err)
	}
	if len(nics) != 1 {
		t.Fatalf("expected 1 NIC, got %d", len(nics))
	}
	if nics[0].Stats == nil {
		t.Fatal("expected Stats to be populated")
	}
	if nics[0].Stats.TxBytes != 1048576 {
		t.Errorf("expected TxBytes 1048576, got %d", nics[0].Stats.TxBytes)
	}
	if nics[0].Status == nil {
		t.Fatal("expected Status to be populated")
	}
	if nics[0].Status.Speed != 10000 {
		t.Errorf("expected Speed 10000, got %d", nics[0].Status.Speed)
	}
}

func TestMachineNICService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, MachineNIC{
				Key: 1, Machine: 10, Name: "eno1",
				Stats:  &MachineNICStats{TxPckts: 50000, RxPckts: 80000},
				Status: &MachineNICStatus{Status: "up", Speed: 1000},
			})
		},
	}))

	nic, err := client.MachineNICs.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if nic.Name != "eno1" {
		t.Errorf("expected name 'eno1', got %q", nic.Name)
	}
	if nic.Stats.RxPckts != 80000 {
		t.Errorf("expected RxPckts 80000, got %d", nic.Stats.RxPckts)
	}
}

func TestMachineNICService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.MachineNICs.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
