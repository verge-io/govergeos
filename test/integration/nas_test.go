//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestNASServices tests the NASServices service against a live VergeOS API.
func TestNASServices(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("List", func(t *testing.T) {
		services, err := client.NASServices.List(ctx)
		if err != nil {
			t.Fatalf("NASServices.List failed: %v", err)
		}
		t.Logf("Found %d NAS services", len(services))

		if len(services) == 0 {
			t.Log("No NAS services found - skipping Get tests")
			return
		}

		first := services[0]
		prettyPrint(t, "First NAS Service", first)
	})

	t.Run("Get", func(t *testing.T) {
		services, err := client.NASServices.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(services) == 0 {
			t.Skip("No NAS services available")
		}

		first := services[0]
		service, err := client.NASServices.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("NASServices.Get failed: %v", err)
		}
		if int(service.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", service.Key, first.Key)
		}
	})

	t.Run("GetByVM", func(t *testing.T) {
		services, err := client.NASServices.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(services) == 0 {
			t.Skip("No NAS services available")
		}

		first := services[0]
		if int(first.VM) == 0 {
			t.Skip("NAS service has no VM reference")
		}

		service, err := client.NASServices.GetByVM(ctx, int(first.VM))
		if err != nil {
			t.Fatalf("NASServices.GetByVM failed: %v", err)
		}
		if int(service.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", service.Key, first.Key)
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		services, err := client.NASServices.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(services) == 0 {
			t.Skip("No NAS services available")
		}

		first := services[0]
		service, err := client.NASServices.GetByName(ctx, first.Name)
		if err != nil {
			t.Fatalf("NASServices.GetByName failed: %v", err)
		}
		if int(service.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", service.Key, first.Key)
		}
	})
}

// TestNASServiceUsers tests the NASServiceUsers service against a live VergeOS API.
func TestNASServiceUsers(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("List", func(t *testing.T) {
		users, err := client.NASServiceUsers.List(ctx)
		if err != nil {
			t.Fatalf("NASServiceUsers.List failed: %v", err)
		}
		t.Logf("Found %d NAS service users", len(users))

		if len(users) > 0 {
			prettyPrint(t, "First NAS Service User", users[0])
		}
	})

	t.Run("ListByService", func(t *testing.T) {
		services, err := client.NASServices.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(services) == 0 {
			t.Skip("No NAS services available")
		}

		serviceID := int(services[0].Key)
		serviceUsers, err := client.NASServiceUsers.ListByService(ctx, serviceID)
		if err != nil {
			t.Fatalf("NASServiceUsers.ListByService failed: %v", err)
		}
		t.Logf("Found %d users for service %d", len(serviceUsers), serviceID)
	})
}

// TestVolumeSyncs tests the VolumeSyncs service against a live VergeOS API.
func TestVolumeSyncs(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("List", func(t *testing.T) {
		syncs, err := client.VolumeSyncs.List(ctx)
		if err != nil {
			t.Fatalf("VolumeSyncs.List failed: %v", err)
		}
		t.Logf("Found %d volume syncs", len(syncs))

		if len(syncs) > 0 {
			prettyPrint(t, "First Volume Sync", syncs[0])
		}
	})

	t.Run("Get", func(t *testing.T) {
		syncs, err := client.VolumeSyncs.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(syncs) == 0 {
			t.Skip("No volume syncs available")
		}

		first := syncs[0]
		sync, err := client.VolumeSyncs.Get(ctx, first.ID)
		if err != nil {
			t.Fatalf("VolumeSyncs.Get failed: %v", err)
		}
		if sync.ID != first.ID {
			t.Errorf("ID mismatch: got %s, want %s", sync.ID, first.ID)
		}
	})

	t.Run("ListEnabled", func(t *testing.T) {
		enabled, err := client.VolumeSyncs.ListEnabled(ctx)
		if err != nil {
			t.Fatalf("VolumeSyncs.ListEnabled failed: %v", err)
		}
		t.Logf("Found %d enabled volume syncs", len(enabled))
	})

	t.Run("ListByService", func(t *testing.T) {
		services, err := client.NASServices.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(services) == 0 {
			t.Skip("No NAS services available")
		}

		serviceID := int(services[0].Key)
		serviceSyncs, err := client.VolumeSyncs.ListByService(ctx, serviceID)
		if err != nil {
			t.Fatalf("VolumeSyncs.ListByService failed: %v", err)
		}
		t.Logf("Found %d syncs for service %d", len(serviceSyncs), serviceID)
	})
}

// TestVolumeSnapshots tests the VolumeSnapshots service against a live VergeOS API.
func TestVolumeSnapshots(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("List", func(t *testing.T) {
		snapshots, err := client.VolumeSnapshots.List(ctx)
		if err != nil {
			t.Fatalf("VolumeSnapshots.List failed: %v", err)
		}
		t.Logf("Found %d volume snapshots", len(snapshots))

		if len(snapshots) > 0 {
			prettyPrint(t, "First Volume Snapshot", snapshots[0])
		}
	})

	t.Run("Get", func(t *testing.T) {
		snapshots, err := client.VolumeSnapshots.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(snapshots) == 0 {
			t.Skip("No volume snapshots available")
		}

		first := snapshots[0]
		snapshot, err := client.VolumeSnapshots.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("VolumeSnapshots.Get failed: %v", err)
		}
		if int(snapshot.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", snapshot.Key, first.Key)
		}
	})

	t.Run("ListExpiring", func(t *testing.T) {
		expiring, err := client.VolumeSnapshots.ListExpiring(ctx, 30)
		if err != nil {
			t.Fatalf("VolumeSnapshots.ListExpiring failed: %v", err)
		}
		t.Logf("Found %d volume snapshots expiring within 30 days", len(expiring))
	})

	t.Run("ListManual", func(t *testing.T) {
		manual, err := client.VolumeSnapshots.ListManual(ctx)
		if err != nil {
			t.Fatalf("VolumeSnapshots.ListManual failed: %v", err)
		}
		t.Logf("Found %d manually created volume snapshots", len(manual))
	})

	t.Run("ListByVolume", func(t *testing.T) {
		volumes, err := client.Volumes.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(volumes) == 0 {
			t.Skip("No volumes available")
		}

		volID := int(volumes[0].Service)
		volSnapshots, err := client.VolumeSnapshots.ListByVolume(ctx, volID)
		if err != nil {
			t.Fatalf("VolumeSnapshots.ListByVolume failed: %v", err)
		}
		t.Logf("Found %d snapshots for volume service %d", len(volSnapshots), volID)
	})
}

// TestCIFSSharesList tests the VolumeCIFSShares service.
func TestCIFSSharesList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Log("Testing VolumeCIFSShares service...")

	shares, err := client.VolumeCIFSShares.List(ctx)
	if err != nil {
		t.Logf("CIFS shares listing: %v (may be empty if no NAS configured)", err)
		return
	}

	t.Logf("Found %d CIFS shares", len(shares))

	if len(shares) == 0 {
		t.Log("No CIFS shares found - this is normal if NAS is not configured")
		return
	}

	// Log first share to verify field mapping
	first := shares[0]
	t.Logf("First CIFS share: ID=%s, Name=%q, Volume=%s, Enabled=%v",
		first.ID, first.Name, first.Volume, first.Enabled)

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VolumeCIFSShares.Get(ctx, first.ID)
		if err != nil {
			t.Errorf("VolumeCIFSShares.Get(%s) failed: %v", first.ID, err)
		} else {
			t.Logf("VolumeCIFSShares.Get succeeded: Name=%q, Comment=%q", fetched.Name, fetched.Comment)
		}
	})

	// Test ListByVolume
	t.Run("ListByVolume", func(t *testing.T) {
		if first.Volume == "" {
			t.Skip("No Volume ID available")
		}
		volShares, err := client.VolumeCIFSShares.ListByVolume(ctx, first.Volume)
		if err != nil {
			t.Errorf("VolumeCIFSShares.ListByVolume failed: %v", err)
		} else {
			t.Logf("Found %d CIFS shares in volume %s", len(volShares), first.Volume)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.Volume == "" {
			t.Skip("No name or Volume ID available")
		}
		byName, err := client.VolumeCIFSShares.GetByName(ctx, first.Volume, first.Name)
		if err != nil {
			t.Errorf("VolumeCIFSShares.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: ID=%s", byName.ID)
		}
	})

	prettyPrint(t, "Sample VolumeCIFSShare", first)
}

// TestNFSSharesList tests the VolumeNFSShares service.
func TestNFSSharesList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Log("Testing VolumeNFSShares service...")

	shares, err := client.VolumeNFSShares.List(ctx)
	if err != nil {
		t.Logf("NFS shares listing: %v (may be empty if no NAS configured)", err)
		return
	}

	t.Logf("Found %d NFS shares", len(shares))

	if len(shares) == 0 {
		t.Log("No NFS shares found - this is normal if NAS is not configured")
		return
	}

	// Log first share to verify field mapping
	first := shares[0]
	t.Logf("First NFS share: ID=%s, Name=%q, Volume=%s, Enabled=%v",
		first.ID, first.Name, first.Volume, first.Enabled)

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.VolumeNFSShares.Get(ctx, first.ID)
		if err != nil {
			t.Errorf("VolumeNFSShares.Get(%s) failed: %v", first.ID, err)
		} else {
			t.Logf("VolumeNFSShares.Get succeeded: Name=%q, Description=%q", fetched.Name, fetched.Description)
		}
	})

	// Test ListByVolume
	t.Run("ListByVolume", func(t *testing.T) {
		if first.Volume == "" {
			t.Skip("No Volume ID available")
		}
		volShares, err := client.VolumeNFSShares.ListByVolume(ctx, first.Volume)
		if err != nil {
			t.Errorf("VolumeNFSShares.ListByVolume failed: %v", err)
		} else {
			t.Logf("Found %d NFS shares in volume %s", len(volShares), first.Volume)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.Volume == "" {
			t.Skip("No name or Volume ID available")
		}
		byName, err := client.VolumeNFSShares.GetByName(ctx, first.Volume, first.Name)
		if err != nil {
			t.Errorf("VolumeNFSShares.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: ID=%s", byName.ID)
		}
	})

	prettyPrint(t, "Sample VolumeNFSShare", first)
}

// TestVolumeBrowserList tests the VolumeBrowser service.
func TestVolumeBrowserList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Log("Testing VolumeBrowser service...")

	jobs, err := client.VolumeBrowser.List(ctx)
	if err != nil {
		t.Logf("Browser jobs listing: %v (may be empty)", err)
		return
	}

	t.Logf("Found %d browser jobs", len(jobs))

	for i, job := range jobs {
		if i >= 3 {
			t.Logf("... and %d more", len(jobs)-3)
			break
		}
		t.Logf("  - %s (volume: %s, query: %s, status: %s)", job.ID, job.Volume, job.Query, job.Status)
	}
}

// TestCIFSSharesCRUD tests CIFS share CRUD operations.
// Requires a valid volume ID - skipped if no volumes available.
func TestCIFSSharesCRUD(t *testing.T) {
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

	// Read
	got, err := client.VolumeCIFSShares.Get(ctx, share.ID)
	if err != nil {
		t.Fatalf("Failed to get CIFS share: %v", err)
	}
	if got.Name != shareName {
		t.Errorf("Name mismatch: got %s, want %s", got.Name, shareName)
	}

	// Update
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
	t.Logf("Updated CIFS share comment to: %q", updated.Comment)

	t.Log("CIFS Share CRUD test completed")
}

// TestNFSSharesCRUD tests NFS share CRUD operations.
// Requires a valid volume ID - skipped if no volumes available.
func TestNFSSharesCRUD(t *testing.T) {
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

	// Read
	got, err := client.VolumeNFSShares.Get(ctx, share.ID)
	if err != nil {
		t.Fatalf("Failed to get NFS share: %v", err)
	}
	if got.Name != shareName {
		t.Errorf("Name mismatch: got %s, want %s", got.Name, shareName)
	}

	// Update
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
	t.Logf("Updated NFS share description to: %q", updated.Description)

	t.Log("NFS Share CRUD test completed")
}

// TestVolumeBrowserBrowse tests the volume browser async API.
// Requires a running NAS service VM - skipped if not available.
func TestVolumeBrowserBrowse(t *testing.T) {
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
	entries, err := client.VolumeBrowser.Browse(ctx, testVolume.ID, "", 100)
	if err != nil {
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
}
