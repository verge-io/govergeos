//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/verge-io/goVergeOS"
)

// TestWave8NASServices tests the NAS services: CIFS shares, NFS shares, and volume browser.
// These tests require:
// - A running NAS service with at least one volume
// - Proper permissions to create/modify shares
func TestWave8NASServices(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Test CIFS Share listing
	t.Run("ListCIFSShares", func(t *testing.T) {
		shares, err := client.VolumeCIFSShares.List(ctx)
		if err != nil {
			t.Logf("CIFS shares listing: %v (may be empty if no NAS configured)", err)
			return
		}
		t.Logf("Found %d CIFS shares", len(shares))
		for _, share := range shares {
			t.Logf("  - %s (volume: %s, enabled: %v)", share.Name, share.Volume, share.Enabled)
		}
	})

	// Test NFS Share listing
	t.Run("ListNFSShares", func(t *testing.T) {
		shares, err := client.VolumeNFSShares.List(ctx)
		if err != nil {
			t.Logf("NFS shares listing: %v (may be empty if no NAS configured)", err)
			return
		}
		t.Logf("Found %d NFS shares", len(shares))
		for _, share := range shares {
			t.Logf("  - %s (volume: %s, enabled: %v)", share.Name, share.Volume, share.Enabled)
		}
	})

	// Test Volume Browser job listing
	t.Run("ListBrowserJobs", func(t *testing.T) {
		jobs, err := client.VolumeBrowser.List(ctx)
		if err != nil {
			t.Logf("Browser jobs listing: %v (may be empty)", err)
			return
		}
		t.Logf("Found %d browser jobs", len(jobs))
		for _, job := range jobs {
			t.Logf("  - %s (volume: %s, query: %s, status: %s)", job.ID, job.Volume, job.Query, job.Status)
		}
	})
}

// TestWave8CIFSShareCRUD tests CIFS share CRUD operations.
// Requires a valid volume ID - skipped if no volumes available.
func TestWave8CIFSShareCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// First, get a volume to create shares on
	volumes, err := client.Volumes.List(ctx)
	if err != nil {
		t.Skipf("Cannot list volumes: %v", err)
	}
	if len(volumes) == 0 {
		t.Skip("No volumes available for testing")
	}

	// Find an enabled local volume (not remote)
	var testVolume *vergeos.Volume
	for _, v := range volumes {
		if v.Enabled && v.RemoteTarget == "" {
			testVolume = &v
			break
		}
	}
	if testVolume == nil {
		t.Skip("No suitable local enabled volume found for testing")
	}
	t.Logf("Using volume: %s (%s)", testVolume.Name, testVolume.ID)

	// Create CIFS share
	shareName := "sdk-test-cifs-" + time.Now().Format("150405")
	t.Run("CreateCIFSShare", func(t *testing.T) {
		readOnly := true
		browseable := true
		share, err := client.VolumeCIFSShares.Create(ctx, &vergeos.VolumeCIFSShareCreateRequest{
			Name:       shareName,
			Volume:     testVolume.ID,
			ReadOnly:   &readOnly,
			Browseable: &browseable,
		})
		if err != nil {
			t.Fatalf("Failed to create CIFS share: %v", err)
		}
		t.Logf("Created CIFS share: %s (%s)", share.Name, share.ID)

		// Cleanup
		defer func() {
			if err := client.VolumeCIFSShares.Delete(ctx, share.ID); err != nil {
				t.Logf("Warning: failed to delete CIFS share: %v", err)
			} else {
				t.Logf("Deleted CIFS share: %s", share.ID)
			}
		}()

		// Verify by getting it back
		t.Run("GetCIFSShare", func(t *testing.T) {
			got, err := client.VolumeCIFSShares.Get(ctx, share.ID)
			if err != nil {
				t.Fatalf("Failed to get CIFS share: %v", err)
			}
			if got.Name != shareName {
				t.Errorf("Name mismatch: got %s, want %s", got.Name, shareName)
			}
			if !got.ReadOnly {
				t.Error("ReadOnly should be true")
			}
		})

		// Update the share
		t.Run("UpdateCIFSShare", func(t *testing.T) {
			comment := "goVergeOS test share"
			updated, err := client.VolumeCIFSShares.Update(ctx, share.ID, &vergeos.VolumeCIFSShareUpdateRequest{
				Comment: &comment,
			})
			if err != nil {
				t.Fatalf("Failed to update CIFS share: %v", err)
			}
			if updated.Comment != comment {
				t.Errorf("Comment mismatch: got %s, want %s", updated.Comment, comment)
			}
		})

		// Get by name
		t.Run("GetCIFSShareByName", func(t *testing.T) {
			got, err := client.VolumeCIFSShares.GetByName(ctx, testVolume.ID, shareName)
			if err != nil {
				t.Fatalf("Failed to get CIFS share by name: %v", err)
			}
			if got.ID != share.ID {
				t.Errorf("ID mismatch: got %s, want %s", got.ID, share.ID)
			}
		})

		// List by volume
		t.Run("ListCIFSSharesByVolume", func(t *testing.T) {
			shares, err := client.VolumeCIFSShares.ListByVolume(ctx, testVolume.ID)
			if err != nil {
				t.Fatalf("Failed to list CIFS shares by volume: %v", err)
			}
			found := false
			for _, s := range shares {
				if s.ID == share.ID {
					found = true
					break
				}
			}
			if !found {
				t.Error("Created share not found in volume's share list")
			}
		})
	})
}

// TestWave8NFSShareCRUD tests NFS share CRUD operations.
// Requires a valid volume ID - skipped if no volumes available.
func TestWave8NFSShareCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// First, get a volume to create shares on
	volumes, err := client.Volumes.List(ctx)
	if err != nil {
		t.Skipf("Cannot list volumes: %v", err)
	}
	if len(volumes) == 0 {
		t.Skip("No volumes available for testing")
	}

	// Find an enabled local volume (not remote)
	var testVolume *vergeos.Volume
	for _, v := range volumes {
		if v.Enabled && v.RemoteTarget == "" {
			testVolume = &v
			break
		}
	}
	if testVolume == nil {
		t.Skip("No suitable local enabled volume found for testing")
	}
	t.Logf("Using volume: %s (%s)", testVolume.Name, testVolume.ID)

	// Create NFS share
	shareName := "sdk-test-nfs-" + time.Now().Format("150405")
	t.Run("CreateNFSShare", func(t *testing.T) {
		dataAccess := vergeos.NFSAccessReadOnly
		allowAll := true
		share, err := client.VolumeNFSShares.Create(ctx, &vergeos.VolumeNFSShareCreateRequest{
			Name:       shareName,
			Volume:     testVolume.ID,
			DataAccess: &dataAccess,
			AllowAll:   &allowAll,
		})
		if err != nil {
			t.Fatalf("Failed to create NFS share: %v", err)
		}
		t.Logf("Created NFS share: %s (%s)", share.Name, share.ID)

		// Cleanup
		defer func() {
			if err := client.VolumeNFSShares.Delete(ctx, share.ID); err != nil {
				t.Logf("Warning: failed to delete NFS share: %v", err)
			} else {
				t.Logf("Deleted NFS share: %s", share.ID)
			}
		}()

		// Verify by getting it back
		t.Run("GetNFSShare", func(t *testing.T) {
			got, err := client.VolumeNFSShares.Get(ctx, share.ID)
			if err != nil {
				t.Fatalf("Failed to get NFS share: %v", err)
			}
			if got.Name != shareName {
				t.Errorf("Name mismatch: got %s, want %s", got.Name, shareName)
			}
		})

		// Update the share
		t.Run("UpdateNFSShare", func(t *testing.T) {
			desc := "goVergeOS test NFS share"
			insecure := true
			updated, err := client.VolumeNFSShares.Update(ctx, share.ID, &vergeos.VolumeNFSShareUpdateRequest{
				Description: &desc,
				Insecure:    &insecure,
			})
			if err != nil {
				t.Fatalf("Failed to update NFS share: %v", err)
			}
			if updated.Description != desc {
				t.Errorf("Description mismatch: got %s, want %s", updated.Description, desc)
			}
			if !updated.Insecure {
				t.Error("Insecure should be true")
			}
		})

		// Get by name
		t.Run("GetNFSShareByName", func(t *testing.T) {
			got, err := client.VolumeNFSShares.GetByName(ctx, testVolume.ID, shareName)
			if err != nil {
				t.Fatalf("Failed to get NFS share by name: %v", err)
			}
			if got.ID != share.ID {
				t.Errorf("ID mismatch: got %s, want %s", got.ID, share.ID)
			}
		})

		// List by volume
		t.Run("ListNFSSharesByVolume", func(t *testing.T) {
			shares, err := client.VolumeNFSShares.ListByVolume(ctx, testVolume.ID)
			if err != nil {
				t.Fatalf("Failed to list NFS shares by volume: %v", err)
			}
			found := false
			for _, s := range shares {
				if s.ID == share.ID {
					found = true
					break
				}
			}
			if !found {
				t.Error("Created share not found in volume's share list")
			}
		})
	})
}

// TestWave8VolumeBrowser tests the volume browser async API.
// Requires a running NAS service VM - skipped if not available.
func TestWave8VolumeBrowser(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// First, get a volume to browse
	volumes, err := client.Volumes.List(ctx)
	if err != nil {
		t.Skipf("Cannot list volumes: %v", err)
	}
	if len(volumes) == 0 {
		t.Skip("No volumes available for testing")
	}

	// Find an enabled local volume (not remote)
	var testVolume *vergeos.Volume
	for _, v := range volumes {
		if v.Enabled && v.RemoteTarget == "" {
			testVolume = &v
			break
		}
	}
	if testVolume == nil {
		t.Skip("No suitable local enabled volume found for testing")
	}
	t.Logf("Using volume: %s (%s)", testVolume.Name, testVolume.ID)

	// Browse root directory
	t.Run("BrowseRootDirectory", func(t *testing.T) {
		entries, err := client.VolumeBrowser.Browse(ctx, testVolume.ID, "", 100)
		if err != nil {
			// The NAS service VM must be running - skip if not
			t.Logf("Browse failed (NAS service may not be running): %v", err)
			t.Skip("Volume browser requires running NAS service VM")
		}
		t.Logf("Found %d entries in root directory", len(entries))
		for i, entry := range entries {
			if i >= 10 {
				t.Logf("  ... and %d more", len(entries)-10)
				break
			}
			t.Logf("  - %s (%s, %d bytes)", entry.Name, entry.Type, entry.Size)
		}
	})

	// Test async job creation and polling manually
	t.Run("CreateAndPollJob", func(t *testing.T) {
		job, err := client.VolumeBrowser.CreateJob(ctx, &vergeos.VolumeBrowserRequest{
			Volume: testVolume.ID,
			Query:  vergeos.VolumeBrowserQueryGetDir,
			Params: &vergeos.VolumeBrowserParams{
				Dir:    "",
				Limit:  10,
				Offset: nil,
				Filter: &vergeos.VolumeBrowserFilter{Extensions: ""},
				Volume: testVolume.ID,
				Sort:   "",
			},
		})
		if err != nil {
			t.Logf("Create job failed (NAS service may not be running): %v", err)
			t.Skip("Volume browser requires running NAS service VM")
		}
		t.Logf("Created browse job: %s", job.ID)

		// Poll for result
		entries, err := client.VolumeBrowser.WaitForResult(ctx, job.ID, 30*time.Second)
		if err != nil {
			t.Fatalf("WaitForResult failed: %v", err)
		}
		t.Logf("Job completed with %d entries", len(entries))
	})
}
