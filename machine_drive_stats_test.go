package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestMachineDriveStatsService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineDriveStats{
				{Key: 1, ParentDrive: 100, Reads: 5000, Writes: 3000, Physical: true},
				{Key: 2, ParentDrive: 200, Reads: 1200, Writes: 800, Physical: false},
			})
		},
	}))

	stats, err := client.MachineDriveStats.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
	if stats[0].Reads != 5000 {
		t.Errorf("expected Reads 5000, got %d", stats[0].Reads)
	}
}

func TestMachineDriveStatsService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineDriveStats{})
		},
	}))

	stats, err := client.MachineDriveStats.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected 0 stats, got %d", len(stats))
	}
}

func TestMachineDriveStatsService_ListPhysical(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "physical eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineDriveStats{
				{Key: 1, ParentDrive: 100, Physical: true, Rops: 150, Wops: 80},
			})
		},
	}))

	stats, err := client.MachineDriveStats.ListPhysical(context.Background())
	if err != nil {
		t.Fatalf("ListPhysical failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if !stats[0].Physical {
		t.Error("expected physical drive stat")
	}
}

func TestMachineDriveStatsService_GetByDrive(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "parent_drive eq 100" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineDriveStats{
				{Key: 1, ParentDrive: 100, ReadBytes: 1048576, WriteBytes: 524288, ServiceTime: 1.5, Util: 35.2},
			})
		},
	}))

	stats, err := client.MachineDriveStats.GetByDrive(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetByDrive failed: %v", err)
	}
	if stats.ParentDrive != 100 {
		t.Errorf("expected ParentDrive 100, got %d", stats.ParentDrive)
	}
	if stats.ServiceTime != 1.5 {
		t.Errorf("expected ServiceTime 1.5, got %f", stats.ServiceTime)
	}
}

func TestMachineDriveStatsService_GetByDrive_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineDriveStats{})
		},
	}))

	_, err := client.MachineDriveStats.GetByDrive(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMachineDriveStatsService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, MachineDriveStats{
				Key: 1, ParentDrive: 100, Rbps: 52428800, Wbps: 26214400,
			})
		},
	}))

	stats, err := client.MachineDriveStats.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if stats.Rbps != 52428800 {
		t.Errorf("expected Rbps 52428800, got %d", stats.Rbps)
	}
}

func TestMachineDriveStatsService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_stats/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.MachineDriveStats.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
