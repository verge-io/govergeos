//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave14VSANMonitoring tests the Wave 14 VSAN monitoring services
// (StorageTiers, ClusterTiers, MachineDrivePhys, ClusterStatsHistory)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave14
func TestWave14VSANMonitoring(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("StorageTiers", func(t *testing.T) {
		testStorageTiers(t, ctx, client)
	})

	t.Run("ClusterTiers", func(t *testing.T) {
		testClusterTiers(t, ctx, client)
	})

	t.Run("MachineDrivePhys", func(t *testing.T) {
		testMachineDrivePhys(t, ctx, client)
	})

	t.Run("ClusterStatsHistory", func(t *testing.T) {
		testClusterStatsHistory(t, ctx, client)
	})
}

func testStorageTiers(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing StorageTiers service...")

	// List all storage tiers
	tiers, err := client.StorageTiers.List(ctx)
	if err != nil {
		t.Fatalf("StorageTiers.List failed: %v", err)
	}

	t.Logf("Found %d storage tiers", len(tiers))

	if len(tiers) == 0 {
		t.Log("No storage tiers found - system may not have VSAN configured")
		return
	}

	// Log first tier to verify field mapping
	first := tiers[0]
	t.Logf("First storage tier: Tier=%d, Capacity=%d, Used=%d, UsedPct=%d%%, DedupeRatio=%.2fx",
		first.Tier, first.Capacity, first.Used, first.UsedPct, float64(first.DedupeRatio)/100.0)

	// Verify essential fields
	if first.Capacity == 0 {
		t.Log("Warning: StorageTier.Capacity is 0 - tier may be empty")
	}

	// Test Get by tier number
	fetched, err := client.StorageTiers.Get(ctx, first.Tier)
	if err != nil {
		t.Errorf("StorageTiers.Get(%d) failed: %v", first.Tier, err)
	} else {
		t.Logf("StorageTiers.Get succeeded: Tier=%d, Allocated=%d", fetched.Tier, fetched.Allocated)
	}

	// Verify stats are populated
	if first.Stats != nil {
		t.Logf("Storage tier stats: Reads=%d, Writes=%d, Rops=%d/s, Wops=%d/s, Rbps=%d/s, Wbps=%d/s",
			first.Stats.Reads, first.Stats.Writes, first.Stats.Rops, first.Stats.Wops,
			first.Stats.Rbps, first.Stats.Wbps)
	} else {
		t.Log("Warning: StorageTier.Stats is nil")
	}

	// Pretty print first tier for field verification
	prettyPrint(t, "Sample StorageTier", first)
}

func testClusterTiers(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing ClusterTiers service...")

	// List all cluster tiers
	tiers, err := client.ClusterTiers.List(ctx)
	if err != nil {
		t.Fatalf("ClusterTiers.List failed: %v", err)
	}

	t.Logf("Found %d cluster tiers", len(tiers))

	if len(tiers) == 0 {
		t.Log("No cluster tiers found - system may not have VSAN configured")
		return
	}

	// Log first tier to verify field mapping
	first := tiers[0]
	t.Logf("First cluster tier: Key=%d, Cluster=%d, Tier=%d",
		int(first.Key), first.Cluster, first.Tier)

	// Verify status is populated
	if first.Status != nil {
		t.Logf("Cluster tier status: Status=%q, State=%q, Redundant=%v, Encrypted=%v",
			first.Status.Status, first.Status.State, first.Status.Redundant, first.Status.Encrypted)
		t.Logf("Cluster tier metrics: Capacity=%d, Used=%d, UsedPct=%d%%, Transaction=%d, Repairs=%d",
			first.Status.Capacity, first.Status.Used, first.Status.UsedPct,
			first.Status.Transaction, first.Status.Repairs)
		t.Logf("Cluster tier health: BadDrives=%.0f, Working=%v, LastWalkTimeMs=%d",
			first.Status.BadDrives, first.Status.Working, first.Status.LastWalkTimeMs)
	} else {
		t.Log("Warning: ClusterTier.Status is nil")
	}

	// Test Get by ID
	fetched, err := client.ClusterTiers.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("ClusterTiers.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("ClusterTiers.Get succeeded: Key=%d", int(fetched.Key))
	}

	// Test ListByCluster
	if first.Cluster > 0 {
		clusterTiers, err := client.ClusterTiers.ListByCluster(ctx, first.Cluster)
		if err != nil {
			t.Errorf("ClusterTiers.ListByCluster(%d) failed: %v", first.Cluster, err)
		} else {
			t.Logf("Found %d tiers for cluster %d", len(clusterTiers), first.Cluster)
		}

		// Test GetByClusterAndTier
		byClusterAndTier, err := client.ClusterTiers.GetByClusterAndTier(ctx, first.Cluster, first.Tier)
		if err != nil {
			t.Errorf("ClusterTiers.GetByClusterAndTier(%d, %d) failed: %v", first.Cluster, first.Tier, err)
		} else {
			t.Logf("ClusterTiers.GetByClusterAndTier succeeded: Key=%d", int(byClusterAndTier.Key))
		}
	}

	// Pretty print first tier for field verification
	prettyPrint(t, "Sample ClusterTier", first)
}

func testMachineDrivePhys(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing MachineDrivePhys service...")

	// List all physical drives (limit to 10 for performance)
	drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(10))
	if err != nil {
		t.Fatalf("MachineDrivePhys.List failed: %v", err)
	}

	t.Logf("Found %d physical drives (limited to 10)", len(drives))

	if len(drives) == 0 {
		t.Log("No physical drives found - may be virtual environment without physical VSAN drives")
		return
	}

	// Log first drive to verify field mapping
	first := drives[0]
	t.Logf("First physical drive: Key=%d, ParentDrive=%d, Model=%q, Serial=%q",
		int(first.Key), first.ParentDrive, first.Model, first.Serial)
	t.Logf("Drive metrics: Temp=%d°C, WearLevel=%d%%, Hours=%d, Size=%d",
		first.Temp, first.WearLevel, first.Hours, first.Size)
	t.Logf("Drive warnings: TempWarn=%v, WearLevelWarn=%v, HoursWarn=%v, ReallocSectorsWarn=%v",
		first.TempWarn, first.WearLevelWarn, first.HoursWarn, first.ReallocSectorsWarn)

	// Log VSAN-specific fields
	if first.VSANTier >= 0 {
		t.Logf("VSAN metrics: DriveID=%d, Tier=%d, Used=%d, Max=%d",
			first.VSANDriveID, first.VSANTier, first.VSANUsed, first.VSANMax)
		t.Logf("VSAN errors: ReadErrors=%d, WriteErrors=%d, AvgLatency=%dµs, MaxLatency=%dµs",
			first.VSANReadErrors, first.VSANWriteErrors, first.VSANAvgLatency, first.VSANMaxLatency)
	} else {
		t.Logf("Drive is not in VSAN (VSANTier=%d)", first.VSANTier)
	}

	// Test Get by ID
	fetched, err := client.MachineDrivePhys.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("MachineDrivePhys.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("MachineDrivePhys.Get succeeded: Model=%q, Path=%q", fetched.Model, fetched.Path)
	}

	// Test GetByParentDrive
	if first.ParentDrive > 0 {
		byParent, err := client.MachineDrivePhys.GetByParentDrive(ctx, first.ParentDrive)
		if err != nil {
			t.Errorf("MachineDrivePhys.GetByParentDrive(%d) failed: %v", first.ParentDrive, err)
		} else {
			t.Logf("MachineDrivePhys.GetByParentDrive succeeded: Key=%d", int(byParent.Key))
		}
	}

	// Test ListByVSANTier if drive is in VSAN
	if first.VSANTier >= 0 {
		tierDrives, err := client.MachineDrivePhys.ListByVSANTier(ctx, int(first.VSANTier))
		if err != nil {
			t.Errorf("MachineDrivePhys.ListByVSANTier(%d) failed: %v", first.VSANTier, err)
		} else {
			t.Logf("Found %d drives in VSAN tier %d", len(tierDrives), first.VSANTier)
		}
	}

	// Test ListSpares
	spares, err := client.MachineDrivePhys.ListSpares(ctx)
	if err != nil {
		t.Errorf("MachineDrivePhys.ListSpares failed: %v", err)
	} else {
		t.Logf("Found %d spare drives", len(spares))
	}

	// Test ListWithWarnings
	warnings, err := client.MachineDrivePhys.ListWithWarnings(ctx)
	if err != nil {
		t.Errorf("MachineDrivePhys.ListWithWarnings failed: %v", err)
	} else {
		t.Logf("Found %d drives with warnings", len(warnings))
		if len(warnings) > 0 {
			for _, w := range warnings {
				t.Logf("  Drive %d: TempWarn=%v, WearWarn=%v, HoursWarn=%v",
					int(w.Key), w.TempWarn, w.WearLevelWarn, w.HoursWarn)
			}
		}
	}

	// Pretty print first drive for field verification
	prettyPrint(t, "Sample MachineDrivePhys", first)
}

func testClusterStatsHistory(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing ClusterStatsHistory service...")

	// First get clusters to test with
	clusters, err := client.Clusters.List(ctx)
	if err != nil {
		t.Fatalf("Clusters.List failed: %v", err)
	}

	if len(clusters) == 0 {
		t.Skip("No clusters found - skipping ClusterStatsHistory tests")
	}

	clusterID := int(clusters[0].Key)
	t.Logf("Testing with cluster %d (%s)", clusterID, clusters[0].Name)

	// Test ListShort
	shortStats, err := client.ClusterStatsHistory.ListShort(ctx, vergeos.WithLimit(5))
	if err != nil {
		t.Fatalf("ClusterStatsHistory.ListShort failed: %v", err)
	}

	t.Logf("Found %d short-term history records (limited to 5)", len(shortStats))

	if len(shortStats) == 0 {
		t.Log("No short-term history records found")
	} else {
		first := shortStats[0]
		t.Logf("First short-term record: Key=%d, Cluster=%d, Timestamp=%d",
			int(first.Key), first.Cluster, first.Timestamp)
		t.Logf("Cluster resources: Nodes=%d/%d online, Machines=%d running",
			first.OnlineNodes, first.TotalNodes, first.RunningMachines)
		t.Logf("RAM (MB): %d/%d used (online: %d)", first.UsedRAM, first.TotalRAM, first.OnlineRAM)
		t.Logf("Cores: %d/%d used (online: %d)", first.UsedCores, first.TotalCores, first.OnlineCores)
		t.Logf("Physical: RAM=%d bytes, VRAM=%d bytes, CPU=%d%%",
			first.PhysRAMUsed, first.PhysVRAMUsed, first.PhysTotalCPU)

		if first.GPUsTotal > 0 {
			t.Logf("GPUs: %d/%d (idle: %d), vGPUs: %d/%d (idle: %d)",
				first.GPUs, first.GPUsTotal, first.GPUsIdle,
				first.VGPUs, first.VGPUsTotal, first.VGPUsIdle)
		}

		// Pretty print first record
		prettyPrint(t, "Sample ClusterStatsHistory (short)", first)
	}

	// Test ListShortByCluster
	clusterShortStats, err := client.ClusterStatsHistory.ListShortByCluster(ctx, clusterID, vergeos.WithLimit(5))
	if err != nil {
		t.Errorf("ClusterStatsHistory.ListShortByCluster(%d) failed: %v", clusterID, err)
	} else {
		t.Logf("Found %d short-term records for cluster %d", len(clusterShortStats), clusterID)
	}

	// Test GetLatestShort
	latest, err := client.ClusterStatsHistory.GetLatestShort(ctx, clusterID)
	if err != nil {
		if vergeos.IsNotFoundError(err) {
			t.Logf("No recent stats for cluster %d (this is normal for new/idle clusters)", clusterID)
		} else {
			t.Errorf("ClusterStatsHistory.GetLatestShort(%d) failed: %v", clusterID, err)
		}
	} else {
		t.Logf("Latest short-term stats for cluster %d: Timestamp=%d, PhysRAMUsed=%d",
			clusterID, latest.Timestamp, latest.PhysRAMUsed)
	}

	// Test ListLong
	longStats, err := client.ClusterStatsHistory.ListLong(ctx, vergeos.WithLimit(5))
	if err != nil {
		t.Errorf("ClusterStatsHistory.ListLong failed: %v", err)
	} else {
		t.Logf("Found %d long-term history records (limited to 5)", len(longStats))
	}

	// Test ListLongByCluster
	clusterLongStats, err := client.ClusterStatsHistory.ListLongByCluster(ctx, clusterID, vergeos.WithLimit(5))
	if err != nil {
		t.Errorf("ClusterStatsHistory.ListLongByCluster(%d) failed: %v", clusterID, err)
	} else {
		t.Logf("Found %d long-term records for cluster %d", len(clusterLongStats), clusterID)
	}

	// Test GetShort if we have records
	if len(shortStats) > 0 {
		record, err := client.ClusterStatsHistory.GetShort(ctx, int(shortStats[0].Key))
		if err != nil {
			t.Errorf("ClusterStatsHistory.GetShort(%d) failed: %v", int(shortStats[0].Key), err)
		} else {
			t.Logf("ClusterStatsHistory.GetShort succeeded: Timestamp=%d", record.Timestamp)
		}
	}

	// Test GetLong if we have records
	if len(longStats) > 0 {
		record, err := client.ClusterStatsHistory.GetLong(ctx, int(longStats[0].Key))
		if err != nil {
			t.Errorf("ClusterStatsHistory.GetLong(%d) failed: %v", int(longStats[0].Key), err)
		} else {
			t.Logf("ClusterStatsHistory.GetLong succeeded: Timestamp=%d", record.Timestamp)
		}
	}
}

// TestVSANMonitoringMetrics demonstrates collecting metrics suitable for Prometheus exporters.
// This test verifies the data can be collected in a format useful for monitoring.
func TestVSANMonitoringMetrics(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Collecting VSAN monitoring metrics...")

	// Collect storage tier metrics
	t.Run("StorageTierMetrics", func(t *testing.T) {
		tiers, err := client.StorageTiers.List(ctx)
		if err != nil {
			t.Fatalf("Failed to collect storage tier metrics: %v", err)
		}

		for _, tier := range tiers {
			// These are the metrics the exporter would expose
			t.Logf("vergeos_vsan_tier_capacity{tier=\"%d\"} %d", tier.Tier, tier.Capacity)
			t.Logf("vergeos_vsan_tier_used{tier=\"%d\"} %d", tier.Tier, tier.Used)
			t.Logf("vergeos_vsan_tier_allocated{tier=\"%d\"} %d", tier.Tier, tier.Allocated)
			t.Logf("vergeos_vsan_tier_used_pct{tier=\"%d\"} %d", tier.Tier, tier.UsedPct)
			t.Logf("vergeos_vsan_tier_dedupe_ratio{tier=\"%d\"} %.2f", tier.Tier, float64(tier.DedupeRatio)/100.0)
			if tier.Stats != nil {
				t.Logf("vergeos_vsan_tier_read_ops{tier=\"%d\"} %d", tier.Tier, tier.Stats.Rops)
				t.Logf("vergeos_vsan_tier_write_ops{tier=\"%d\"} %d", tier.Tier, tier.Stats.Wops)
				t.Logf("vergeos_vsan_tier_read_bytes{tier=\"%d\"} %d", tier.Tier, tier.Stats.Rbps)
				t.Logf("vergeos_vsan_tier_write_bytes{tier=\"%d\"} %d", tier.Tier, tier.Stats.Wbps)
			}
		}
	})

	// Collect cluster tier metrics
	t.Run("ClusterTierMetrics", func(t *testing.T) {
		tiers, err := client.ClusterTiers.List(ctx)
		if err != nil {
			t.Fatalf("Failed to collect cluster tier metrics: %v", err)
		}

		for _, tier := range tiers {
			labels := map[string]interface{}{"cluster": tier.Cluster, "tier": tier.Tier}
			if tier.Status != nil {
				t.Logf("vergeos_cluster_tier_redundant{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster, tier.Tier, boolToInt(tier.Status.Redundant))
				t.Logf("vergeos_cluster_tier_encrypted{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster, tier.Tier, boolToInt(tier.Status.Encrypted))
				t.Logf("vergeos_cluster_tier_transaction{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster, tier.Tier, tier.Status.Transaction)
				t.Logf("vergeos_cluster_tier_repairs{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster, tier.Tier, tier.Status.Repairs)
				t.Logf("vergeos_cluster_tier_bad_drives{cluster=\"%d\",tier=\"%d\"} %.0f",
					tier.Cluster, tier.Tier, tier.Status.BadDrives)
				t.Logf("vergeos_cluster_tier_walk_time_ms{cluster=\"%d\",tier=\"%d\"} %d",
					tier.Cluster, tier.Tier, tier.Status.LastWalkTimeMs)
			}
			_ = labels // Use labels variable
		}
	})

	// Collect drive metrics
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
			t.Logf("vergeos_drive_vsan_used{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.VSANUsed)
			t.Logf("vergeos_drive_vsan_max{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.VSANMax)
			t.Logf("vergeos_drive_vsan_read_errors{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.VSANReadErrors)
			t.Logf("vergeos_drive_vsan_write_errors{drive=\"%d\",tier=\"%d\"} %d",
				int(drive.Key), drive.VSANTier, drive.VSANWriteErrors)
		}
	})

	// Collect cluster stats history metrics
	t.Run("ClusterHistoryMetrics", func(t *testing.T) {
		clusters, err := client.Clusters.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list clusters: %v", err)
		}

		for _, cluster := range clusters {
			clusterID := int(cluster.Key)
			stats, err := client.ClusterStatsHistory.GetLatestShort(ctx, clusterID)
			if err != nil {
				if !vergeos.IsNotFoundError(err) {
					t.Errorf("Failed to get stats for cluster %d: %v", clusterID, err)
				}
				continue
			}

			t.Logf("vergeos_cluster_phys_ram_used{cluster=\"%d\"} %d", clusterID, stats.PhysRAMUsed)
			t.Logf("vergeos_cluster_phys_vram_used{cluster=\"%d\"} %d", clusterID, stats.PhysVRAMUsed)
			t.Logf("vergeos_cluster_running_machines{cluster=\"%d\"} %d", clusterID, stats.RunningMachines)
			t.Logf("vergeos_cluster_online_nodes{cluster=\"%d\"} %d", clusterID, stats.OnlineNodes)
			t.Logf("vergeos_cluster_total_nodes{cluster=\"%d\"} %d", clusterID, stats.TotalNodes)
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
