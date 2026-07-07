//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/macstadium/govergeos"
)

// TestStorageTiers tests the StorageTiers service against a live VergeOS API.
func TestStorageTiers(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		tiers, err := client.StorageTiers.List(ctx)
		if err != nil {
			t.Fatalf("StorageTiers.List failed: %v", err)
		}
		t.Logf("Found %d storage tiers", len(tiers))

		if len(tiers) == 0 {
			t.Log("No storage tiers found - system may not have VSAN configured")
			return
		}

		first := tiers[0]
		t.Logf("First storage tier: Tier=%d, Capacity=%d, Used=%d, UsedPct=%d%%, DedupeRatio=%.2fx",
			first.Tier, first.Capacity, first.Used, first.UsedPct, float64(first.DedupeRatio)/100.0)

		if first.Stats != nil {
			t.Logf("Storage tier stats: Reads=%d, Writes=%d, Rops=%d/s, Wops=%d/s, Rbps=%d/s, Wbps=%d/s",
				first.Stats.Reads, first.Stats.Writes, first.Stats.Rops, first.Stats.Wops,
				first.Stats.Rbps, first.Stats.Wbps)
		}
		prettyPrint(t, "Sample StorageTier", first)
	})

	t.Run("Get", func(t *testing.T) {
		tiers, err := client.StorageTiers.List(ctx)
		if err != nil || len(tiers) == 0 {
			t.Skip("No storage tiers available")
		}

		first := tiers[0]
		fetched, err := client.StorageTiers.Get(ctx, first.Tier)
		if err != nil {
			t.Fatalf("StorageTiers.Get(%d) failed: %v", first.Tier, err)
		}
		t.Logf("StorageTiers.Get succeeded: Tier=%d, Allocated=%d", fetched.Tier, fetched.Allocated)
	})
}

// TestClusterTiers tests the ClusterTiers service against a live VergeOS API.
func TestClusterTiers(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		tiers, err := client.ClusterTiers.List(ctx)
		if err != nil {
			t.Fatalf("ClusterTiers.List failed: %v", err)
		}
		t.Logf("Found %d cluster tiers", len(tiers))

		if len(tiers) == 0 {
			t.Log("No cluster tiers found - system may not have VSAN configured")
			return
		}

		first := tiers[0]
		t.Logf("First cluster tier: Key=%d, Cluster=%d, Tier=%d",
			first.Key.Int(), first.Cluster.Int(), first.Tier)

		if first.Status != nil {
			t.Logf("Cluster tier status: Status=%q, State=%q, Redundant=%v, Encrypted=%v",
				first.Status.Status, first.Status.State, first.Status.Redundant, first.Status.Encrypted)
			t.Logf("Cluster tier metrics: Capacity=%d, Used=%d, UsedPct=%d%%, Transaction=%d, Repairs=%d",
				first.Status.Capacity, first.Status.Used, first.Status.UsedPct,
				first.Status.Transaction, first.Status.Repairs)
		}

		// Validate nodes_online and drives_online are populated
		onlineNodes := first.CountOnlineNodes()
		onlineDrives := first.CountOnlineDrives()
		t.Logf("Online counts: Nodes=%d, Drives=%d", onlineNodes, onlineDrives)

		if first.NodesOnline == nil {
			t.Error("NodesOnline is nil — computed field expression may not be working")
		}
		if len(first.DrivesOnline) == 0 {
			t.Error("DrivesOnline is empty — computed field expression may not be working")
		}
		if onlineNodes == 0 {
			t.Error("CountOnlineNodes returned 0 — expected at least 1 online node")
		}
		if onlineDrives == 0 {
			t.Error("CountOnlineDrives returned 0 — expected at least 1 online drive")
		}

		prettyPrint(t, "Sample ClusterTier", first)
	})

	t.Run("Get", func(t *testing.T) {
		tiers, err := client.ClusterTiers.List(ctx)
		if err != nil || len(tiers) == 0 {
			t.Skip("No cluster tiers available")
		}

		first := tiers[0]
		fetched, err := client.ClusterTiers.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("ClusterTiers.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("ClusterTiers.Get succeeded: Key=%d", int(fetched.Key))
	})

	t.Run("ListByCluster", func(t *testing.T) {
		tiers, err := client.ClusterTiers.List(ctx)
		if err != nil || len(tiers) == 0 {
			t.Skip("No cluster tiers available")
		}

		first := tiers[0]
		if first.Cluster.Int() == 0 {
			t.Skip("Cluster tier has no cluster reference")
		}

		clusterTiers, err := client.ClusterTiers.ListByCluster(ctx, first.Cluster.Int())
		if err != nil {
			t.Fatalf("ClusterTiers.ListByCluster(%d) failed: %v", first.Cluster.Int(), err)
		}
		t.Logf("Found %d tiers for cluster %d", len(clusterTiers), first.Cluster.Int())
	})

	t.Run("GetByClusterAndTier", func(t *testing.T) {
		tiers, err := client.ClusterTiers.List(ctx)
		if err != nil || len(tiers) == 0 {
			t.Skip("No cluster tiers available")
		}

		first := tiers[0]
		if first.Cluster.Int() == 0 {
			t.Skip("Cluster tier has no cluster reference")
		}

		byClusterAndTier, err := client.ClusterTiers.GetByClusterAndTier(ctx, first.Cluster.Int(), first.Tier)
		if err != nil {
			t.Fatalf("ClusterTiers.GetByClusterAndTier(%d, %d) failed: %v", first.Cluster.Int(), first.Tier, err)
		}
		t.Logf("ClusterTiers.GetByClusterAndTier succeeded: Key=%d", int(byClusterAndTier.Key))
	})
}

// TestMachineDrivePhys tests the MachineDrivePhys service against a live VergeOS API.
func TestMachineDrivePhys(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(10))
		if err != nil {
			t.Fatalf("MachineDrivePhys.List failed: %v", err)
		}
		t.Logf("Found %d physical drives (limited to 10)", len(drives))

		if len(drives) == 0 {
			t.Log("No physical drives found - may be virtual environment without physical VSAN drives")
			return
		}

		first := drives[0]
		t.Logf("First physical drive: Key=%d, ParentDrive=%d, Model=%q, Serial=%q",
			int(first.Key), first.ParentDrive, first.Model, first.Serial)
		t.Logf("Drive metrics: Temp=%d°C, WearLevel=%d%%, Hours=%d, Size=%d",
			first.Temp, first.WearLevel, first.Hours, first.Size)

		if first.VSANTier >= 0 {
			t.Logf("VSAN metrics: DriveID=%d, Tier=%d, Used=%d, Max=%d",
				first.VSANDriveID, first.VSANTier, first.VSANUsed, first.VSANMax)
		}
		prettyPrint(t, "Sample MachineDrivePhys", first)
	})

	t.Run("Get", func(t *testing.T) {
		drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(drives) == 0 {
			t.Skip("No physical drives available")
		}

		first := drives[0]
		fetched, err := client.MachineDrivePhys.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("MachineDrivePhys.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("MachineDrivePhys.Get succeeded: Model=%q, Path=%q", fetched.Model, fetched.Path)
	})

	t.Run("GetByParentDrive", func(t *testing.T) {
		drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(drives) == 0 {
			t.Skip("No physical drives available")
		}

		first := drives[0]
		if first.ParentDrive == 0 {
			t.Skip("Drive has no parent drive reference")
		}

		byParent, err := client.MachineDrivePhys.GetByParentDrive(ctx, first.ParentDrive)
		if err != nil {
			t.Fatalf("MachineDrivePhys.GetByParentDrive(%d) failed: %v", first.ParentDrive, err)
		}
		t.Logf("MachineDrivePhys.GetByParentDrive succeeded: Key=%d", int(byParent.Key))
	})

	t.Run("ListByVSANTier", func(t *testing.T) {
		drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(drives) == 0 {
			t.Skip("No physical drives available")
		}

		first := drives[0]
		if first.VSANTier < 0 {
			t.Skip("Drive is not in VSAN")
		}

		tierDrives, err := client.MachineDrivePhys.ListByVSANTier(ctx, int(first.VSANTier))
		if err != nil {
			t.Fatalf("MachineDrivePhys.ListByVSANTier(%d) failed: %v", first.VSANTier, err)
		}
		t.Logf("Found %d drives in VSAN tier %d", len(tierDrives), first.VSANTier)
	})

	t.Run("ListSpares", func(t *testing.T) {
		spares, err := client.MachineDrivePhys.ListSpares(ctx)
		if err != nil {
			t.Fatalf("MachineDrivePhys.ListSpares failed: %v", err)
		}
		t.Logf("Found %d spare drives", len(spares))
	})

	t.Run("ListWithWarnings", func(t *testing.T) {
		warnings, err := client.MachineDrivePhys.ListWithWarnings(ctx)
		if err != nil {
			t.Fatalf("MachineDrivePhys.ListWithWarnings failed: %v", err)
		}
		t.Logf("Found %d drives with warnings", len(warnings))

		for _, w := range warnings {
			t.Logf("  Drive %d: TempWarn=%v, WearWarn=%v, HoursWarn=%v",
				int(w.Key), w.TempWarn, w.WearLevelWarn, w.HoursWarn)
		}
	})
}

// TestClusterStatsHistory tests the ClusterStatsHistory service against a live VergeOS API.
func TestClusterStatsHistory(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Get clusters for testing
	clusters, err := client.Clusters.List(ctx)
	if err != nil {
		t.Fatalf("Clusters.List failed: %v", err)
	}
	if len(clusters) == 0 {
		t.Skip("No clusters found - skipping ClusterStatsHistory tests")
	}
	clusterID := int(clusters[0].Key)
	t.Logf("Testing with cluster %d (%s)", clusterID, clusters[0].Name)

	t.Run("ListShort", func(t *testing.T) {
		shortStats, err := client.ClusterStatsHistory.ListShort(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("ClusterStatsHistory.ListShort failed: %v", err)
		}
		t.Logf("Found %d short-term history records (limited to 5)", len(shortStats))

		if len(shortStats) > 0 {
			first := shortStats[0]
			t.Logf("First short-term record: Key=%d, Cluster=%d, Timestamp=%d",
				int(first.Key), first.Cluster, first.Timestamp)
			t.Logf("Cluster resources: Nodes=%d/%d online, Machines=%d running",
				first.OnlineNodes, first.TotalNodes, first.RunningMachines)
			prettyPrint(t, "Sample ClusterStatsHistory (short)", first)
		}
	})

	t.Run("ListShortByCluster", func(t *testing.T) {
		clusterShortStats, err := client.ClusterStatsHistory.ListShortByCluster(ctx, clusterID, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("ClusterStatsHistory.ListShortByCluster(%d) failed: %v", clusterID, err)
		}
		t.Logf("Found %d short-term records for cluster %d", len(clusterShortStats), clusterID)
	})

	t.Run("GetLatestShort", func(t *testing.T) {
		latest, err := client.ClusterStatsHistory.GetLatestShort(ctx, clusterID)
		if err != nil {
			if vergeos.IsNotFoundError(err) {
				t.Logf("No recent stats for cluster %d (this is normal for new/idle clusters)", clusterID)
			} else {
				t.Fatalf("ClusterStatsHistory.GetLatestShort(%d) failed: %v", clusterID, err)
			}
		} else {
			t.Logf("Latest short-term stats for cluster %d: Timestamp=%d, PhysRAMUsed=%d",
				clusterID, latest.Timestamp, latest.PhysRAMUsed)
		}
	})

	t.Run("ListLong", func(t *testing.T) {
		longStats, err := client.ClusterStatsHistory.ListLong(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("ClusterStatsHistory.ListLong failed: %v", err)
		}
		t.Logf("Found %d long-term history records (limited to 5)", len(longStats))
	})

	t.Run("ListLongByCluster", func(t *testing.T) {
		clusterLongStats, err := client.ClusterStatsHistory.ListLongByCluster(ctx, clusterID, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("ClusterStatsHistory.ListLongByCluster(%d) failed: %v", clusterID, err)
		}
		t.Logf("Found %d long-term records for cluster %d", len(clusterLongStats), clusterID)
	})
}

// TestVSANMonitoringMetrics demonstrates collecting metrics suitable for Prometheus exporters.
func TestVSANMonitoringMetrics(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Collecting VSAN monitoring metrics...")

	t.Run("StorageTierMetrics", func(t *testing.T) {
		tiers, err := client.StorageTiers.List(ctx)
		if err != nil {
			t.Fatalf("Failed to collect storage tier metrics: %v", err)
		}

		for _, tier := range tiers {
			t.Logf("vergeos_vsan_tier_capacity{tier=\"%d\"} %d", tier.Tier, tier.Capacity)
			t.Logf("vergeos_vsan_tier_used{tier=\"%d\"} %d", tier.Tier, tier.Used)
			t.Logf("vergeos_vsan_tier_allocated{tier=\"%d\"} %d", tier.Tier, tier.Allocated)
			t.Logf("vergeos_vsan_tier_used_pct{tier=\"%d\"} %d", tier.Tier, tier.UsedPct)
			t.Logf("vergeos_vsan_tier_dedupe_ratio{tier=\"%d\"} %.2f", tier.Tier, float64(tier.DedupeRatio)/100.0)
			if tier.Stats != nil {
				t.Logf("vergeos_vsan_tier_read_ops{tier=\"%d\"} %d", tier.Tier, tier.Stats.Rops)
				t.Logf("vergeos_vsan_tier_write_ops{tier=\"%d\"} %d", tier.Tier, tier.Stats.Wops)
			}
		}
	})

	t.Run("ClusterTierMetrics", func(t *testing.T) {
		tiers, err := client.ClusterTiers.List(ctx)
		if err != nil {
			t.Fatalf("Failed to collect cluster tier metrics: %v", err)
		}

		for _, tier := range tiers {
			if tier.Status != nil {
				t.Logf("vergeos_cluster_tier_redundant{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster.Int(), tier.Tier, boolToInt(tier.Status.Redundant))
				t.Logf("vergeos_cluster_tier_encrypted{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster.Int(), tier.Tier, boolToInt(tier.Status.Encrypted))
				t.Logf("vergeos_cluster_tier_transaction{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster.Int(), tier.Tier, tier.Status.Transaction)
				t.Logf("vergeos_cluster_tier_repairs{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster.Int(), tier.Tier, tier.Status.Repairs)
			}
		}
	})

	t.Run("DriveMetrics", func(t *testing.T) {
		drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Failed to collect drive metrics: %v", err)
		}

		for _, drive := range drives {
			if drive.VSANTier < 0 {
				continue // Skip non-VSAN drives
			}

			t.Logf("vergeos_drive_temperature{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.Temp)
			t.Logf("vergeos_drive_wear_level{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.WearLevel)
			t.Logf("vergeos_drive_hours{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.Hours)
		}
	})
}

// boolToInt converts a boolean to 0 or 1 for Prometheus gauge metrics
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
