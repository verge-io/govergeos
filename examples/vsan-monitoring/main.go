// Example: VSAN Monitoring
//
// This example demonstrates how to use the VSAN monitoring services
// to collect storage metrics for Prometheus exporters and monitoring dashboards.
//
// Services used:
// - StorageTiers: System-wide storage tier capacity and usage
// - ClusterTiers: Cluster-specific tier status, redundancy, encryption
// - MachineDrivePhys: Physical drive metrics (temperature, wear, SMART)
// - ClusterStatsHistory: Historical cluster resource usage
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/macstadium/govergeos"
)

func main() {
	// Create client from environment variables
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(os.Getenv("VERGEOS_HOST")),
		vergeos.WithCredentials(os.Getenv("VERGEOS_USERNAME"), os.Getenv("VERGEOS_PASSWORD")),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Storage Tiers - System-wide capacity and deduplication
	fmt.Println("=== Storage Tiers (System-wide) ===")
	storageTiers, err := client.StorageTiers.List(ctx)
	if err != nil {
		log.Printf("Failed to list storage tiers: %v", err)
	} else {
		for _, tier := range storageTiers {
			fmt.Printf("Tier %d:\n", tier.Tier)
			fmt.Printf("  Capacity: %s\n", formatBytes(tier.Capacity))
			fmt.Printf("  Used: %s (%.1f%%)\n", formatBytes(tier.Used), float64(tier.UsedPct))
			fmt.Printf("  Allocated: %s\n", formatBytes(tier.Allocated))
			fmt.Printf("  Dedupe Ratio: %.2fx\n", float64(tier.DedupeRatio)/100.0)
			if tier.Stats != nil {
				fmt.Printf("  I/O: %d rops, %d wops, %s/s read, %s/s write\n",
					tier.Stats.Rops, tier.Stats.Wops,
					formatBytes(tier.Stats.Rbps), formatBytes(tier.Stats.Wbps))
			}
			fmt.Println()
		}
	}

	// Example 2: Cluster Tiers - Per-cluster status and redundancy
	fmt.Println("=== Cluster Tiers ===")
	clusterTiers, err := client.ClusterTiers.List(ctx)
	if err != nil {
		log.Printf("Failed to list cluster tiers: %v", err)
	} else {
		for _, tier := range clusterTiers {
			fmt.Printf("Cluster %d, Tier %d:\n", tier.Cluster, tier.Tier)
			if tier.Status != nil {
				fmt.Printf("  Status: %s (state: %s)\n", tier.Status.Status, tier.Status.State)
				fmt.Printf("  Capacity: %s\n", formatBytes(tier.Status.Capacity))
				fmt.Printf("  Used: %s (%.1f%%)\n", formatBytes(tier.Status.Used), float64(tier.Status.UsedPct))
				fmt.Printf("  Redundant: %t\n", tier.Status.Redundant)
				fmt.Printf("  Encrypted: %t\n", tier.Status.Encrypted)
				fmt.Printf("  Transaction: %d\n", tier.Status.Transaction)
				fmt.Printf("  Repairs: %d\n", tier.Status.Repairs)
				fmt.Printf("  Bad Drives: %.0f\n", tier.Status.BadDrives)
				fmt.Printf("  Last Walk: %dms\n", tier.Status.LastWalkTimeMs)
			}
			fmt.Println()
		}
	}

	// Example 3: Physical Drive Metrics - For drive-level monitoring
	fmt.Println("=== Physical Drive Metrics ===")
	drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(10))
	if err != nil {
		log.Printf("Failed to list physical drives: %v", err)
	} else {
		for _, drive := range drives {
			fmt.Printf("Drive %d (VSAN ID: %d, Tier: %d):\n", int(drive.Key), drive.VSANDriveID, drive.VSANTier)
			fmt.Printf("  Model: %s\n", drive.Model)
			fmt.Printf("  Size: %s\n", formatBytes(drive.Size))
			fmt.Printf("  Temperature: %d°C (warning: %t)\n", drive.Temp, drive.TempWarn)
			fmt.Printf("  Wear Level: %d%% (warning: %t)\n", drive.WearLevel, drive.WearLevelWarn)
			fmt.Printf("  Power-on Hours: %d (warning: %t)\n", drive.Hours, drive.HoursWarn)
			fmt.Printf("  Reallocated Sectors: %d (warning: %t)\n", drive.ReallocSectors, drive.ReallocSectorsWarn)
			if drive.VSANTier >= 0 {
				fmt.Printf("  VSAN Used: %s / %s\n", formatBytes(drive.VSANUsed), formatBytes(drive.VSANMax))
				fmt.Printf("  VSAN Read Errors: %d\n", drive.VSANReadErrors)
				fmt.Printf("  VSAN Write Errors: %d\n", drive.VSANWriteErrors)
				fmt.Printf("  VSAN Avg Latency: %d µs\n", drive.VSANAvgLatency)
			}
			fmt.Println()
		}
	}

	// Example 4: Check for drives with warnings
	fmt.Println("=== Drives with Warnings ===")
	warningDrives, err := client.MachineDrivePhys.ListWithWarnings(ctx)
	if err != nil {
		log.Printf("Failed to list drives with warnings: %v", err)
	} else if len(warningDrives) == 0 {
		fmt.Println("No drives with warnings - all healthy!")
	} else {
		for _, drive := range warningDrives {
			fmt.Printf("Drive %d: ", int(drive.Key))
			warnings := []string{}
			if drive.TempWarn {
				warnings = append(warnings, fmt.Sprintf("temp=%d°C", drive.Temp))
			}
			if drive.WearLevelWarn {
				warnings = append(warnings, fmt.Sprintf("wear=%d%%", drive.WearLevel))
			}
			if drive.HoursWarn {
				warnings = append(warnings, fmt.Sprintf("hours=%d", drive.Hours))
			}
			if drive.ReallocSectorsWarn {
				warnings = append(warnings, fmt.Sprintf("realloc=%d", drive.ReallocSectors))
			}
			fmt.Println(warnings)
		}
	}
	fmt.Println()

	// Example 5: Cluster Statistics History - For RAM usage trends
	fmt.Println("=== Cluster Stats History (Latest) ===")
	clusters, err := client.Clusters.List(ctx)
	if err != nil {
		log.Printf("Failed to list clusters: %v", err)
	} else {
		for _, cluster := range clusters {
			stats, err := client.ClusterStatsHistory.GetLatestShort(ctx, int(cluster.Key))
			if err != nil {
				log.Printf("  No recent stats for cluster %s: %v", cluster.Name, err)
				continue
			}
			fmt.Printf("Cluster %s (ID: %d):\n", cluster.Name, int(cluster.Key))
			fmt.Printf("  Timestamp: %s\n", time.Unix(int64(stats.Timestamp), 0).Format(time.RFC3339))
			fmt.Printf("  Nodes: %d/%d online\n", stats.OnlineNodes, stats.TotalNodes)
			fmt.Printf("  Running Machines: %d\n", stats.RunningMachines)
			fmt.Printf("  RAM: %d/%d MB used (online: %d MB)\n", stats.UsedRAM, stats.TotalRAM, stats.OnlineRAM)
			fmt.Printf("  Cores: %d/%d used (online: %d)\n", stats.UsedCores, stats.TotalCores, stats.OnlineCores)
			fmt.Printf("  Physical RAM Used: %s\n", formatBytes(stats.PhysRAMUsed))
			if stats.GPUsTotal > 0 {
				fmt.Printf("  GPUs: %d/%d (idle: %d)\n", stats.GPUs, stats.GPUsTotal, stats.GPUsIdle)
			}
			fmt.Println()
		}
	}

	// Example 6: Historical data for graphing
	fmt.Println("=== Recent Stats History (Last 5 Records) ===")
	if len(clusters) > 0 {
		recentStats, err := client.ClusterStatsHistory.ListShortByCluster(ctx, int(clusters[0].Key),
			vergeos.WithSort("-timestamp"),
			vergeos.WithLimit(5))
		if err != nil {
			log.Printf("Failed to get recent history: %v", err)
		} else {
			fmt.Printf("Cluster %s recent history:\n", clusters[0].Name)
			for _, s := range recentStats {
				ts := time.Unix(int64(s.Timestamp), 0)
				fmt.Printf("  %s: RAM %d/%d MB, Cores %d/%d, Machines %d\n",
					ts.Format("15:04:05"),
					s.UsedRAM, s.OnlineRAM,
					s.UsedCores, s.OnlineCores,
					s.RunningMachines)
			}
		}
	}
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
