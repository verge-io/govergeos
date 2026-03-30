//go:build integration

package integration

import (
	"context"
	"testing"
)

// TestNodes tests the Nodes service against a live VergeOS API.
func TestNodes(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		nodes, err := client.Nodes.List(ctx)
		if err != nil {
			t.Fatalf("Nodes.List failed: %v", err)
		}
		t.Logf("Found %d nodes", len(nodes))

		if len(nodes) == 0 {
			t.Skip("No nodes found")
		}

		first := nodes[0]
		t.Logf("First node: ID=%d, Name=%q, Cluster=%d, Physical=%v, Cores=%d, RAM=%d",
			first.ID, first.Name, first.Cluster, first.Physical, first.Cores, first.RAM)

		// Validate vm_stats_totals is populated
		if first.VMStatsTotals == nil {
			t.Error("VMStatsTotals is nil — running_machines aggregate field expression may not be working")
		} else {
			t.Logf("VMStatsTotals: RunningCores=%d, RunningRAM=%d",
				first.VMStatsTotals.RunningCores, first.VMStatsTotals.RunningRAM)

			if first.VMStatsTotals.RunningCores == 0 && first.VMStatsTotals.RunningRAM == 0 {
				t.Log("Warning: both RunningCores and RunningRAM are 0 — node may have no running VMs")
			}
		}

		prettyPrint(t, "Sample Node", first)
	})

	t.Run("ListPhysical", func(t *testing.T) {
		physNodes, err := client.Nodes.ListPhysical(ctx)
		if err != nil {
			t.Fatalf("Nodes.ListPhysical failed: %v", err)
		}
		t.Logf("Found %d physical nodes", len(physNodes))

		for _, n := range physNodes {
			if !n.Physical {
				t.Errorf("Node %q (ID=%d) returned by ListPhysical but Physical=%v", n.Name, n.ID, n.Physical)
			}
			if n.VMStatsTotals != nil {
				t.Logf("Node %q: RunningCores=%d, RunningRAM=%d",
					n.Name, n.VMStatsTotals.RunningCores, n.VMStatsTotals.RunningRAM)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		nodes, err := client.Nodes.List(ctx)
		if err != nil || len(nodes) == 0 {
			t.Skip("No nodes available")
		}

		first := nodes[0]
		fetched, err := client.Nodes.Get(ctx, first.ID)
		if err != nil {
			t.Fatalf("Nodes.Get(%d) failed: %v", first.ID, err)
		}
		t.Logf("Nodes.Get succeeded: ID=%d, Name=%q", fetched.ID, fetched.Name)

		if fetched.VMStatsTotals != nil {
			t.Logf("VMStatsTotals from Get: RunningCores=%d, RunningRAM=%d",
				fetched.VMStatsTotals.RunningCores, fetched.VMStatsTotals.RunningRAM)
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		nodes, err := client.Nodes.List(ctx)
		if err != nil || len(nodes) == 0 {
			t.Skip("No nodes available")
		}

		first := nodes[0]
		fetched, err := client.Nodes.GetByName(ctx, first.Name)
		if err != nil {
			t.Fatalf("Nodes.GetByName(%q) failed: %v", first.Name, err)
		}
		t.Logf("Nodes.GetByName succeeded: ID=%d, Name=%q", fetched.ID, fetched.Name)
	})

	t.Run("VMStatsTotalsAllNodes", func(t *testing.T) {
		nodes, err := client.Nodes.ListPhysical(ctx)
		if err != nil {
			t.Fatalf("Nodes.ListPhysical failed: %v", err)
		}

		totalCores := 0
		totalRAM := 0
		for _, n := range nodes {
			if n.VMStatsTotals != nil {
				totalCores += n.VMStatsTotals.RunningCores
				totalRAM += n.VMStatsTotals.RunningRAM
				t.Logf("vergeos_node_running_cores{node=%q} %d", n.Name, n.VMStatsTotals.RunningCores)
				t.Logf("vergeos_node_running_ram{node=%q} %d", n.Name, n.VMStatsTotals.RunningRAM)
			}
		}
		t.Logf("Totals across all nodes: RunningCores=%d, RunningRAM=%dMB", totalCores, totalRAM)

		if totalCores == 0 {
			t.Error("Total RunningCores across all nodes is 0 — expected at least some running VMs")
		}
	})
}
