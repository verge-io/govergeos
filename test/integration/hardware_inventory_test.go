//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestHardwareInventory reads GPU, DIMM, and LLDP inventory.
// It does not change GPU mode.
func TestHardwareInventory(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("VGPUProfiles", func(t *testing.T) {
		profiles, err := client.VGPUProfiles.List(ctx)
		if err != nil {
			t.Fatalf("VGPUProfiles.List failed: %v", err)
		}
		t.Logf("Found %d vGPU profiles", len(profiles))
		if len(profiles) == 0 {
			return
		}
		first := profiles[0]
		got, err := client.VGPUProfiles.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("VGPUProfiles.Get(%d) failed: %v", int(first.Key), err)
		}
		if got.Name != first.Name {
			t.Errorf("Get name %q, list name %q", got.Name, first.Name)
		}
	})

	t.Run("NodeGPUs", func(t *testing.T) {
		gpus, err := client.NodeGPUs.List(ctx)
		if err != nil {
			t.Fatalf("NodeGPUs.List failed: %v", err)
		}
		t.Logf("Found %d node GPUs", len(gpus))
		if len(gpus) == 0 {
			return
		}
		got, err := client.NodeGPUs.Get(ctx, int(gpus[0].Key))
		if err != nil {
			t.Fatalf("NodeGPUs.Get failed: %v", err)
		}
		t.Logf("GPU %s mode %s (%s)", got.Name, got.Mode, got.ModeDisplay())
	})

	t.Run("NodeGPUStats", func(t *testing.T) {
		stats, err := client.NodeGPUStats.List(ctx)
		if err != nil {
			t.Fatalf("NodeGPUStats.List failed: %v", err)
		}
		t.Logf("Found %d GPU stats rows", len(stats))
		history, err := client.NodeGPUStats.ListHistoryShort(ctx, vergeos.WithLimit(1))
		if err != nil {
			t.Fatalf("NodeGPUStats.ListHistoryShort failed: %v", err)
		}
		t.Logf("Found %d short GPU history rows (limited)", len(history))
	})

	t.Run("NodeGPUInstances", func(t *testing.T) {
		instances, err := client.NodeGPUInstances.List(ctx)
		if err != nil {
			t.Fatalf("NodeGPUInstances.List failed: %v", err)
		}
		t.Logf("Found %d GPU instances", len(instances))
	})

	t.Run("NodeVGPUDevices", func(t *testing.T) {
		devices, err := client.NodeVGPUDevices.List(ctx)
		if err != nil {
			t.Fatalf("NodeVGPUDevices.List failed: %v", err)
		}
		t.Logf("Found %d vGPU devices", len(devices))
	})

	t.Run("NodeHostGPUDevices", func(t *testing.T) {
		devices, err := client.NodeHostGPUDevices.List(ctx)
		if err != nil {
			t.Fatalf("NodeHostGPUDevices.List failed: %v", err)
		}
		t.Logf("Found %d host GPU devices", len(devices))
	})

	t.Run("NodeVGPUProfiles", func(t *testing.T) {
		profiles, err := client.NodeVGPUProfiles.List(ctx)
		if err != nil {
			t.Fatalf("NodeVGPUProfiles.List failed: %v", err)
		}
		t.Logf("Found %d per-GPU vGPU profiles", len(profiles))
	})

	t.Run("NodeMemory", func(t *testing.T) {
		dimms, err := client.NodeMemory.List(ctx)
		if err != nil {
			t.Fatalf("NodeMemory.List failed: %v", err)
		}
		t.Logf("Found %d DIMMs", len(dimms))
		if len(dimms) == 0 {
			return
		}
		got, err := client.NodeMemory.Get(ctx, int(dimms[0].Key))
		if err != nil {
			t.Fatalf("NodeMemory.Get failed: %v", err)
		}
		t.Logf("DIMM %s status %s healthy %v", got.Locator, got.Status, got.IsHealthy())
	})

	t.Run("NodeLLDPNeighbors", func(t *testing.T) {
		neighbors, err := client.NodeLLDPNeighbors.List(ctx)
		if err != nil {
			t.Fatalf("NodeLLDPNeighbors.List failed: %v", err)
		}
		t.Logf("Found %d LLDP neighbors", len(neighbors))
		if len(neighbors) == 0 {
			return
		}
		got, err := client.NodeLLDPNeighbors.Get(ctx, int(neighbors[0].Key))
		if err != nil {
			t.Fatalf("NodeLLDPNeighbors.Get failed: %v", err)
		}
		t.Logf("neighbor nic %d chassis %s port %s", got.NIC, got.ChassisName(), got.PortID())
	})
}
