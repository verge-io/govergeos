//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestMachineDriveStats tests the MachineDriveStats service against a live VergeOS API.
func TestMachineDriveStats(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		stats, err := client.MachineDriveStats.List(ctx)
		if err != nil {
			t.Fatalf("MachineDriveStats.List failed: %v", err)
		}
		t.Logf("Found %d machine drive stats records", len(stats))

		if len(stats) == 0 {
			t.Log("No machine drive stats found")
			return
		}

		first := stats[0]
		t.Logf("First drive stats: Key=%d, ParentDrive=%d, Physical=%v",
			int(first.Key), first.ParentDrive, first.Physical)
		t.Logf("I/O totals: Reads=%d, Writes=%d, ReadBytes=%d, WriteBytes=%d",
			first.Reads, first.Writes, first.ReadBytes, first.WriteBytes)
		t.Logf("I/O rates: Rops=%d/s, Wops=%d/s, Rbps=%d/s, Wbps=%d/s",
			first.Rops, first.Wops, first.Rbps, first.Wbps)
		t.Logf("Performance: ServiceTime=%.2fms, Util=%.2f%%",
			first.ServiceTime, first.Util)

		prettyPrint(t, "Sample MachineDriveStats", first)
	})

	t.Run("ListPhysical", func(t *testing.T) {
		stats, err := client.MachineDriveStats.ListPhysical(ctx)
		if err != nil {
			t.Fatalf("MachineDriveStats.ListPhysical failed: %v", err)
		}
		t.Logf("Found %d physical drive stats records", len(stats))
	})

	t.Run("GetByDrive", func(t *testing.T) {
		// Get a drive to look up stats for
		drives, err := client.MachineDrivePhys.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(drives) == 0 {
			t.Skip("No physical drives available")
		}

		driveID := drives[0].ParentDrive
		if driveID == 0 {
			t.Skip("Drive has no parent_drive reference")
		}

		stats, err := client.MachineDriveStats.GetByDrive(ctx, driveID)
		if err != nil {
			if vergeos.IsNotFoundError(err) {
				t.Logf("No stats for drive %d", driveID)
				return
			}
			t.Fatalf("MachineDriveStats.GetByDrive(%d) failed: %v", driveID, err)
		}
		t.Logf("MachineDriveStats.GetByDrive(%d): Reads=%d, Writes=%d, Util=%.2f%%",
			driveID, stats.Reads, stats.Writes, stats.Util)
	})

	t.Run("Get", func(t *testing.T) {
		stats, err := client.MachineDriveStats.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(stats) == 0 {
			t.Skip("No machine drive stats available")
		}

		first := stats[0]
		fetched, err := client.MachineDriveStats.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("MachineDriveStats.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("MachineDriveStats.Get succeeded: ParentDrive=%d, Reads=%d", fetched.ParentDrive, fetched.Reads)
	})
}
