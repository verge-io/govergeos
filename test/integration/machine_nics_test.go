//go:build integration

package integration

import (
	"context"
	"testing"
)

// TestMachineNICs tests the MachineNICs service against a live VergeOS API.
func TestMachineNICs(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		nics, err := client.MachineNICs.List(ctx)
		if err != nil {
			t.Fatalf("MachineNICs.List failed: %v", err)
		}
		t.Logf("Found %d machine NICs", len(nics))

		if len(nics) == 0 {
			t.Log("No machine NICs found")
			return
		}

		first := nics[0]
		t.Logf("First NIC: Key=%d, Machine=%d, Name=%q",
			int(first.Key), first.Machine, first.Name)

		if first.Stats != nil {
			t.Logf("NIC stats: TxPckts=%d, RxPckts=%d, TxBytes=%d, RxBytes=%d",
				first.Stats.TxPckts, first.Stats.RxPckts, first.Stats.TxBytes, first.Stats.RxBytes)
		} else {
			t.Log("NIC stats: nil (not expanded)")
		}

		if first.Status != nil {
			t.Logf("NIC status: Status=%q, Speed=%dMbps",
				first.Status.Status, first.Status.Speed)
		} else {
			t.Log("NIC status: nil (not expanded)")
		}

		prettyPrint(t, "Sample MachineNIC", first)
	})

	t.Run("ListByMachine", func(t *testing.T) {
		// Get physical nodes to find a machine ID
		nodes, err := client.Nodes.ListPhysical(ctx)
		if err != nil || len(nodes) == 0 {
			t.Skip("No physical nodes available")
		}

		machineID := nodes[0].Machine
		if machineID == 0 {
			t.Skip("Node has no machine reference")
		}

		nics, err := client.MachineNICs.ListByMachine(ctx, machineID)
		if err != nil {
			t.Fatalf("MachineNICs.ListByMachine(%d) failed: %v", machineID, err)
		}
		t.Logf("Found %d NICs for machine %d", len(nics), machineID)

		for _, nic := range nics {
			status := "unknown"
			if nic.Status != nil {
				status = nic.Status.Status
			}
			t.Logf("  NIC %q: status=%s", nic.Name, status)
		}
	})

	t.Run("Get", func(t *testing.T) {
		nics, err := client.MachineNICs.List(ctx)
		if err != nil || len(nics) == 0 {
			t.Skip("No machine NICs available")
		}

		first := nics[0]
		fetched, err := client.MachineNICs.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("MachineNICs.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("MachineNICs.Get succeeded: Name=%q, Machine=%d", fetched.Name, fetched.Machine)
	})
}
