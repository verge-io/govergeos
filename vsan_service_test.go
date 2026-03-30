package vergeos

import (
	"context"
	"net/http"
	"testing"
)

// --- StorageTierService ---

func TestStorageTierService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/storage_tiers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []StorageTier{
				{Key: 0, Tier: 0, Description: "SSD", Capacity: 1000000},
				{Key: 1, Tier: 1, Description: "HDD", Capacity: 5000000},
			})
		},
	}))

	tiers, err := client.StorageTiers.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tiers) != 2 {
		t.Fatalf("expected 2 tiers, got %d", len(tiers))
	}
	if tiers[0].Description != "SSD" {
		t.Errorf("expected description 'SSD', got %q", tiers[0].Description)
	}
}

func TestStorageTierService_List_WithFields(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/storage_tiers": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != "all" {
				t.Errorf("expected fields 'all', got %q", fields)
			}
			jsonResponse(w, 200, []StorageTier{{Key: 0}})
		},
	}))

	_, err := client.StorageTiers.List(context.Background(), WithFields("all"))
	if err != nil {
		t.Fatalf("List with fields failed: %v", err)
	}
}

func TestStorageTierService_List_DefaultFields(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/storage_tiers": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != storageTierListFields {
				t.Errorf("expected default fields, got %q", fields)
			}
			jsonResponse(w, 200, []StorageTier{})
		},
	}))

	_, err := client.StorageTiers.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestStorageTierService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/storage_tiers/0": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, StorageTier{Key: 0, Tier: 0, Description: "SSD", Capacity: 1000000})
		},
	}))

	tier, err := client.StorageTiers.Get(context.Background(), 0)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if tier.Description != "SSD" {
		t.Errorf("expected description 'SSD', got %q", tier.Description)
	}
	if tier.Capacity != 1000000 {
		t.Errorf("expected capacity 1000000, got %d", tier.Capacity)
	}
}

func TestStorageTierService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/storage_tiers/99": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.StorageTiers.Get(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- ClusterTierService ---

func TestClusterTierService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_tiers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ClusterTier{
				{Key: FlexInt(1), Tier: 0, Description: "SSD"},
				{Key: FlexInt(2), Tier: 1, Description: "HDD"},
			})
		},
	}))

	tiers, err := client.ClusterTiers.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tiers) != 2 {
		t.Fatalf("expected 2 tiers, got %d", len(tiers))
	}
}

func TestClusterTierService_ListByCluster(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_tiers": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "cluster eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []ClusterTier{{Key: FlexInt(1), Tier: 0}})
		},
	}))

	tiers, err := client.ClusterTiers.ListByCluster(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByCluster failed: %v", err)
	}
	if len(tiers) != 1 {
		t.Fatalf("expected 1 tier, got %d", len(tiers))
	}
}

func TestClusterTierService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_tiers/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, ClusterTier{Key: FlexInt(1), Tier: 0, Description: "SSD"})
		},
	}))

	tier, err := client.ClusterTiers.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if tier.Description != "SSD" {
		t.Errorf("expected description 'SSD', got %q", tier.Description)
	}
}

func TestClusterTierService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_tiers/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.ClusterTiers.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterTierService_GetByClusterAndTier(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_tiers": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "cluster eq 5 and tier eq 2" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []ClusterTier{{Key: FlexInt(10), Tier: 2}})
		},
	}))

	tier, err := client.ClusterTiers.GetByClusterAndTier(context.Background(), 5, 2)
	if err != nil {
		t.Fatalf("GetByClusterAndTier failed: %v", err)
	}
	if tier.Tier != 2 {
		t.Errorf("expected tier 2, got %d", tier.Tier)
	}
}

func TestClusterTierService_GetByClusterAndTier_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_tiers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ClusterTier{})
		},
	}))

	_, err := client.ClusterTiers.GetByClusterAndTier(context.Background(), 99, 99)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- MachineDrivePhysService ---

func TestMachineDrivePhysService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineDrivePhys{
				{Key: FlexInt(1), Serial: "ABC123"},
				{Key: FlexInt(2), Serial: "DEF456"},
			})
		},
	}))

	drives, err := client.MachineDrivePhys.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(drives) != 2 {
		t.Fatalf("expected 2 drives, got %d", len(drives))
	}
	if drives[0].Serial != "ABC123" {
		t.Errorf("expected serial 'ABC123', got %q", drives[0].Serial)
	}
}

func TestMachineDrivePhysService_ListByVSANTier(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vsan_tier eq 0" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineDrivePhys{{Key: FlexInt(1)}})
		},
	}))

	drives, err := client.MachineDrivePhys.ListByVSANTier(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListByVSANTier failed: %v", err)
	}
	if len(drives) != 1 {
		t.Fatalf("expected 1 drive, got %d", len(drives))
	}
}

func TestMachineDrivePhysService_ListSpares(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "spare eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineDrivePhys{{Key: FlexInt(3)}})
		},
	}))

	drives, err := client.MachineDrivePhys.ListSpares(context.Background())
	if err != nil {
		t.Fatalf("ListSpares failed: %v", err)
	}
	if len(drives) != 1 {
		t.Fatalf("expected 1 drive, got %d", len(drives))
	}
}

func TestMachineDrivePhysService_ListWithWarnings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for warnings query")
			}
			jsonResponse(w, 200, []MachineDrivePhys{{Key: FlexInt(1)}})
		},
	}))

	drives, err := client.MachineDrivePhys.ListWithWarnings(context.Background())
	if err != nil {
		t.Fatalf("ListWithWarnings failed: %v", err)
	}
	if len(drives) != 1 {
		t.Fatalf("expected 1 drive, got %d", len(drives))
	}
}

func TestMachineDrivePhysService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, MachineDrivePhys{Key: FlexInt(1), Serial: "ABC123"})
		},
	}))

	drive, err := client.MachineDrivePhys.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if drive.Serial != "ABC123" {
		t.Errorf("expected serial 'ABC123', got %q", drive.Serial)
	}
}

func TestMachineDrivePhysService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.MachineDrivePhys.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMachineDrivePhysService_GetByParentDrive(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "parent_drive eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineDrivePhys{{Key: 1, Serial: "XYZ"}})
		},
	}))

	drive, err := client.MachineDrivePhys.GetByParentDrive(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetByParentDrive failed: %v", err)
	}
	if drive.Serial != "XYZ" {
		t.Errorf("expected serial 'XYZ', got %q", drive.Serial)
	}
}

func TestMachineDrivePhysService_GetByParentDrive_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drive_phys": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineDrivePhys{})
		},
	}))

	_, err := client.MachineDrivePhys.GetByParentDrive(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- ClusterStatsHistoryService ---

func TestClusterStatsHistoryService_ListShort(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ClusterStatsHistory{
				{Key: FlexInt(1), TotalNodes: 3, OnlineNodes: 3},
				{Key: FlexInt(2), TotalNodes: 3, OnlineNodes: 2},
			})
		},
	}))

	stats, err := client.ClusterStatsHistory.ListShort(context.Background())
	if err != nil {
		t.Fatalf("ListShort failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
}

func TestClusterStatsHistoryService_ListShortByCluster(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "cluster eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []ClusterStatsHistory{{Key: FlexInt(1)}})
		},
	}))

	stats, err := client.ClusterStatsHistory.ListShortByCluster(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListShortByCluster failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stats, got %d", len(stats))
	}
}

func TestClusterStatsHistoryService_ListLong(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_long": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ClusterStatsHistory{
				{Key: FlexInt(1), TotalNodes: 4},
			})
		},
	}))

	stats, err := client.ClusterStatsHistory.ListLong(context.Background())
	if err != nil {
		t.Fatalf("ListLong failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stats, got %d", len(stats))
	}
}

func TestClusterStatsHistoryService_ListLongByCluster(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_long": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "cluster eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []ClusterStatsHistory{{Key: FlexInt(1)}})
		},
	}))

	stats, err := client.ClusterStatsHistory.ListLongByCluster(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListLongByCluster failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
}

func TestClusterStatsHistoryService_GetLatestShort(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			sort := r.URL.Query().Get("sort")
			if sort != "-timestamp" {
				t.Errorf("expected sort '-timestamp', got %q", sort)
			}
			limit := r.URL.Query().Get("limit")
			if limit != "1" {
				t.Errorf("expected limit '1', got %q", limit)
			}
			jsonResponse(w, 200, []ClusterStatsHistory{{Key: FlexInt(99), TotalNodes: 3}})
		},
	}))

	stat, err := client.ClusterStatsHistory.GetLatestShort(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetLatestShort failed: %v", err)
	}
	if stat.TotalNodes != 3 {
		t.Errorf("expected TotalNodes 3, got %d", stat.TotalNodes)
	}
}

func TestClusterStatsHistoryService_GetLatestShort_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []ClusterStatsHistory{})
		},
	}))

	_, err := client.ClusterStatsHistory.GetLatestShort(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterStatsHistoryService_GetShort(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_short/42": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, ClusterStatsHistory{Key: FlexInt(42), TotalNodes: 3})
		},
	}))

	stat, err := client.ClusterStatsHistory.GetShort(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetShort failed: %v", err)
	}
	if stat.Key != 42 {
		t.Errorf("expected key 42, got %d", stat.Key)
	}
}

func TestClusterStatsHistoryService_GetShort_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_short/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.ClusterStatsHistory.GetShort(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterStatsHistoryService_GetLong(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_long/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, ClusterStatsHistory{Key: FlexInt(7), OnlineNodes: 2})
		},
	}))

	stat, err := client.ClusterStatsHistory.GetLong(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetLong failed: %v", err)
	}
	if stat.OnlineNodes != 2 {
		t.Errorf("expected OnlineNodes 2, got %d", stat.OnlineNodes)
	}
}

func TestClusterStatsHistoryService_GetLong_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cluster_stats_history_long/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.ClusterStatsHistory.GetLong(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
