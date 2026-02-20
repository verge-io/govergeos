//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestMachineStats tests the MachineStats service against a live VergeOS API.
func TestMachineStats(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		stats, err := client.MachineStats.List(ctx)
		if err != nil {
			t.Fatalf("MachineStats.List failed: %v", err)
		}
		t.Logf("Found %d machine stats records", len(stats))

		if len(stats) == 0 {
			t.Log("No machine stats found")
			return
		}

		first := stats[0]
		t.Logf("First machine stats: Key=%d, Machine=%d, TotalCPU=%d%%, RAMUsed=%dMB, RAMPct=%d%%",
			int(first.Key), first.Machine, first.TotalCPU, first.RAMUsed, first.RAMPct)
		t.Logf("CPU breakdown: User=%d%%, System=%d%%, IOWait=%d%%",
			first.UserCPU, first.SystemCPU, first.IOWaitCPU)
		t.Logf("Temperature: CoreTemp=%d°C, CoreTempTop=%d°C",
			first.CoreTemp, first.CoreTempTop)

		usages, err := first.GetCoreUsages()
		if err != nil {
			t.Logf("Could not parse core usages: %v", err)
		} else if len(usages) > 0 {
			t.Logf("Per-core usages (%d cores): first=%.1f%%", len(usages), usages[0])
		}

		prettyPrint(t, "Sample MachineStats", first)
	})

	t.Run("GetByMachine", func(t *testing.T) {
		// Get physical nodes to find a machine ID
		nodes, err := client.Nodes.ListPhysical(ctx)
		if err != nil || len(nodes) == 0 {
			t.Skip("No physical nodes available")
		}

		machineID := nodes[0].Machine
		if machineID == 0 {
			t.Skip("Node has no machine reference")
		}

		stats, err := client.MachineStats.GetByMachine(ctx, machineID)
		if err != nil {
			if vergeos.IsNotFoundError(err) {
				t.Logf("No stats for machine %d (normal for some node types)", machineID)
				return
			}
			t.Fatalf("MachineStats.GetByMachine(%d) failed: %v", machineID, err)
		}
		t.Logf("MachineStats.GetByMachine(%d): TotalCPU=%d%%, RAMUsed=%dMB, CoreTemp=%d°C",
			machineID, stats.TotalCPU, stats.RAMUsed, stats.CoreTemp)
	})

	t.Run("Get", func(t *testing.T) {
		stats, err := client.MachineStats.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(stats) == 0 {
			t.Skip("No machine stats available")
		}

		first := stats[0]
		fetched, err := client.MachineStats.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("MachineStats.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("MachineStats.Get succeeded: Machine=%d, TotalCPU=%d%%", fetched.Machine, fetched.TotalCPU)
	})
}
