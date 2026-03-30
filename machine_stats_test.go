package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestMachineStatsService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineStats{
				{Key: 1, Machine: 10, TotalCPU: 45, RAMUsed: 4096, RAMPct: 50},
				{Key: 2, Machine: 20, TotalCPU: 80, RAMUsed: 7168, RAMPct: 87},
			})
		},
	}))

	stats, err := client.MachineStats.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	if stats[0].TotalCPU != 45 {
		t.Errorf("expected TotalCPU 45, got %d", stats[0].TotalCPU)
	}
}

func TestMachineStatsService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineStats{})
		},
	}))

	stats, err := client.MachineStats.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected 0 stats, got %d", len(stats))
	}
}

func TestMachineStatsService_GetByMachine(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_stats": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineStats{
				{Key: 1, Machine: 10, TotalCPU: 55, UserCPU: 40, SystemCPU: 15},
			})
		},
	}))

	stats, err := client.MachineStats.GetByMachine(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByMachine failed: %v", err)
	}
	if stats.TotalCPU != 55 {
		t.Errorf("expected TotalCPU 55, got %d", stats.TotalCPU)
	}
}

func TestMachineStatsService_GetByMachine_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineStats{})
		},
	}))

	_, err := client.MachineStats.GetByMachine(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMachineStatsService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_stats/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, MachineStats{
				Key: 1, Machine: 10, TotalCPU: 30, CoreTemp: 65, CoreTempTop: 72,
			})
		},
	}))

	stats, err := client.MachineStats.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if stats.CoreTemp != 65 {
		t.Errorf("expected CoreTemp 65, got %d", stats.CoreTemp)
	}
	if stats.CoreTempTop != 72 {
		t.Errorf("expected CoreTempTop 72, got %d", stats.CoreTempTop)
	}
}

func TestMachineStatsService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_stats/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.MachineStats.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
