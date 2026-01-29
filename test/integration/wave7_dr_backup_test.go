//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestWave7DRBackup tests the Wave 7 DR & Backup services (Sites, SiteSyncs, CloudSnapshots)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave7
func TestWave7DRBackup(t *testing.T) {
	client := setupTestClientWave7(t)
	ctx := context.Background()

	t.Run("Sites", func(t *testing.T) {
		testSites(t, ctx, client)
	})

	t.Run("SiteSyncsIncoming", func(t *testing.T) {
		testSiteSyncsIncoming(t, ctx, client)
	})

	t.Run("SiteSyncsOutgoing", func(t *testing.T) {
		testSiteSyncsOutgoing(t, ctx, client)
	})

	t.Run("CloudSnapshots", func(t *testing.T) {
		testCloudSnapshots(t, ctx, client)
	})

	t.Run("CloudSnapshotVMs", func(t *testing.T) {
		testCloudSnapshotVMs(t, ctx, client)
	})

	t.Run("CloudSnapshotTenants", func(t *testing.T) {
		testCloudSnapshotTenants(t, ctx, client)
	})
}

func testSites(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Sites service...")

	// List all sites
	sites, err := client.Sites.List(ctx)
	if err != nil {
		t.Fatalf("Sites.List failed: %v", err)
	}

	t.Logf("Found %d sites", len(sites))

	if len(sites) == 0 {
		t.Log("No sites found - this is normal if DR/backup sites are not configured")
		return
	}

	// Log first site to verify field mapping
	first := sites[0]
	t.Logf("First site: Key=%d, Name=%q, URL=%q, Status=%q, AuthStatus=%q, Enabled=%v",
		int(first.Key), first.Name, first.URL, first.Status, first.AuthenticationStatus, first.Enabled)

	// Verify ID is a 40-character SHA1 hash
	if first.ID != "" && len(first.ID) != 40 {
		t.Logf("Warning: Site.ID length is %d, expected 40 (SHA1 hash)", len(first.ID))
	}

	// Test Get by ID
	fetched, err := client.Sites.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("Sites.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("Sites.Get succeeded: Name=%q, Domain=%q, City=%q, Country=%q",
			fetched.Name, fetched.Domain, fetched.City, fetched.Country)
	}

	// Test GetByName
	if first.Name != "" {
		byName, err := client.Sites.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("Sites.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Test GetBySiteID if we have a SHA1 ID
	if first.ID != "" {
		bySiteID, err := client.Sites.GetBySiteID(ctx, first.ID)
		if err != nil {
			t.Errorf("Sites.GetBySiteID failed: %v", err)
		} else {
			t.Logf("GetBySiteID succeeded: Key=%d", int(bySiteID.Key))
		}
	}

	// Pretty print first site for field verification
	prettyPrintWave7(t, "Sample Site", first)
}

func testSiteSyncsIncoming(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing SiteSyncsIncoming service...")

	// List all incoming syncs
	syncs, err := client.SiteSyncsIncoming.List(ctx)
	if err != nil {
		t.Fatalf("SiteSyncsIncoming.List failed: %v", err)
	}

	t.Logf("Found %d incoming syncs", len(syncs))

	if len(syncs) == 0 {
		t.Log("No incoming syncs found - this is normal if site sync is not configured")
		return
	}

	// Log first sync to verify field mapping
	first := syncs[0]
	t.Logf("First incoming sync: Key=%d, Name=%q, Site=%d, Status=%q, State=%q, Enabled=%v",
		int(first.Key), first.Name, int(first.Site), first.Status, first.State, first.Enabled)

	// Verify SyncID is a 40-character SHA1 hash if present
	if first.SyncID != "" && len(first.SyncID) != 40 {
		t.Logf("Warning: SiteSyncIncoming.SyncID length is %d, expected 40 (SHA1 hash)", len(first.SyncID))
	}

	// Test Get by ID
	fetched, err := client.SiteSyncsIncoming.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("SiteSyncsIncoming.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("SiteSyncsIncoming.Get succeeded: Name=%q, PublicIP=%q, ForceTier=%q",
			fetched.Name, fetched.PublicIP, fetched.ForceTier)
	}

	// Test ListBySite
	if first.Site > 0 {
		siteSyncs, err := client.SiteSyncsIncoming.ListBySite(ctx, int(first.Site))
		if err != nil {
			t.Errorf("SiteSyncsIncoming.ListBySite failed: %v", err)
		} else {
			t.Logf("Found %d incoming syncs for site %d", len(siteSyncs), int(first.Site))
		}
	}

	// Test GetByName
	if first.Name != "" && first.Site > 0 {
		byName, err := client.SiteSyncsIncoming.GetByName(ctx, int(first.Site), first.Name)
		if err != nil {
			t.Errorf("SiteSyncsIncoming.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Test GetBySyncID if we have a SHA1 ID
	if first.SyncID != "" {
		bySyncID, err := client.SiteSyncsIncoming.GetBySyncID(ctx, first.SyncID)
		if err != nil {
			t.Errorf("SiteSyncsIncoming.GetBySyncID failed: %v", err)
		} else {
			t.Logf("GetBySyncID succeeded: Key=%d", int(bySyncID.Key))
		}
	}

	// Pretty print first sync for field verification
	prettyPrintWave7(t, "Sample SiteSyncIncoming", first)
}

func testSiteSyncsOutgoing(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing SiteSyncsOutgoing service...")

	// List all outgoing syncs
	syncs, err := client.SiteSyncsOutgoing.List(ctx)
	if err != nil {
		t.Fatalf("SiteSyncsOutgoing.List failed: %v", err)
	}

	t.Logf("Found %d outgoing syncs", len(syncs))

	if len(syncs) == 0 {
		t.Log("No outgoing syncs found - this is normal if site sync is not configured")
		return
	}

	// Log first sync to verify field mapping
	first := syncs[0]
	t.Logf("First outgoing sync: Key=%d, Name=%q, Site=%d, URL=%q, Status=%q, State=%q, Enabled=%v",
		int(first.Key), first.Name, int(first.Site), first.URL, first.Status, first.State, first.Enabled)

	// Test Get by ID
	fetched, err := client.SiteSyncsOutgoing.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("SiteSyncsOutgoing.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("SiteSyncsOutgoing.Get succeeded: Name=%q, DestinationTier=%q, Encryption=%v, Compression=%v",
			fetched.Name, fetched.DestinationTier, fetched.Encryption, fetched.Compression)
	}

	// Test ListBySite
	if first.Site > 0 {
		siteSyncs, err := client.SiteSyncsOutgoing.ListBySite(ctx, int(first.Site))
		if err != nil {
			t.Errorf("SiteSyncsOutgoing.ListBySite failed: %v", err)
		} else {
			t.Logf("Found %d outgoing syncs for site %d", len(siteSyncs), int(first.Site))
		}
	}

	// Test GetByName
	if first.Name != "" && first.Site > 0 {
		byName, err := client.SiteSyncsOutgoing.GetByName(ctx, int(first.Site), first.Name)
		if err != nil {
			t.Errorf("SiteSyncsOutgoing.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first sync for field verification
	prettyPrintWave7(t, "Sample SiteSyncOutgoing", first)
}

func testCloudSnapshots(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing CloudSnapshots service...")

	// List all cloud snapshots
	snapshots, err := client.CloudSnapshots.List(ctx)
	if err != nil {
		t.Fatalf("CloudSnapshots.List failed: %v", err)
	}

	t.Logf("Found %d cloud snapshots", len(snapshots))

	if len(snapshots) == 0 {
		t.Log("No cloud snapshots found - this is normal if no system snapshots have been taken")
		return
	}

	// Log first snapshot to verify field mapping
	first := snapshots[0]
	t.Logf("First snapshot: Key=%d, Name=%q, Status=%q, Expires=%d, Immutable=%v, Private=%v",
		int(first.Key), first.Name, first.Status, first.Expires, first.Immutable, first.Private)

	// Test Get by ID
	fetched, err := client.CloudSnapshots.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("CloudSnapshots.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("CloudSnapshots.Get succeeded: Name=%q, Description=%q, Provider=%v, RemoteSync=%v",
			fetched.Name, fetched.Description, fetched.Provider, fetched.RemoteSync)
	}

	// Test GetByName
	if first.Name != "" {
		byName, err := client.CloudSnapshots.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("CloudSnapshots.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Test ListExpiring
	expiring, err := client.CloudSnapshots.ListExpiring(ctx)
	if err != nil {
		t.Errorf("CloudSnapshots.ListExpiring failed: %v", err)
	} else {
		t.Logf("Found %d expiring cloud snapshots", len(expiring))
	}

	// Test ListLocal
	local, err := client.CloudSnapshots.ListLocal(ctx)
	if err != nil {
		t.Errorf("CloudSnapshots.ListLocal failed: %v", err)
	} else {
		t.Logf("Found %d local cloud snapshots (non-provider)", len(local))
	}

	// Test ListByProfile if we have a profile
	if first.SnapshotProfile > 0 {
		profileSnaps, err := client.CloudSnapshots.ListByProfile(ctx, int(first.SnapshotProfile))
		if err != nil {
			t.Errorf("CloudSnapshots.ListByProfile failed: %v", err)
		} else {
			t.Logf("Found %d snapshots for profile %d", len(profileSnaps), int(first.SnapshotProfile))
		}
	}

	// Pretty print first snapshot for field verification
	prettyPrintWave7(t, "Sample CloudSnapshot", first)
}

func testCloudSnapshotVMs(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing CloudSnapshotVMs service...")

	// List all cloud snapshot VMs
	vms, err := client.CloudSnapshotVMs.List(ctx)
	if err != nil {
		t.Fatalf("CloudSnapshotVMs.List failed: %v", err)
	}

	t.Logf("Found %d cloud snapshot VMs", len(vms))

	if len(vms) == 0 {
		t.Log("No cloud snapshot VMs found - this is normal if snapshots haven't been scanned for VMs")
		return
	}

	// Log first VM to verify field mapping
	first := vms[0]
	t.Logf("First snapshot VM: Key=%d, Name=%q, VM=%d, CloudSnapshot=%d, CPUCores=%d, RAM=%d",
		int(first.Key), first.Name, int(first.VM), int(first.CloudSnapshot), first.CPUCores, first.RAM)

	// Test Get by ID
	fetched, err := client.CloudSnapshotVMs.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("CloudSnapshotVMs.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("CloudSnapshotVMs.Get succeeded: Name=%q, Description=%q, OSFamily=%q",
			fetched.Name, fetched.Description, fetched.OSFamily)
	}

	// Test ListBySnapshot
	if first.CloudSnapshot > 0 {
		snapVMs, err := client.CloudSnapshotVMs.ListBySnapshot(ctx, int(first.CloudSnapshot))
		if err != nil {
			t.Errorf("CloudSnapshotVMs.ListBySnapshot failed: %v", err)
		} else {
			t.Logf("Found %d VMs in snapshot %d", len(snapVMs), int(first.CloudSnapshot))
		}
	}

	// Pretty print first VM for field verification
	prettyPrintWave7(t, "Sample CloudSnapshotVM", first)
}

func testCloudSnapshotTenants(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing CloudSnapshotTenants service...")

	// List all cloud snapshot tenants
	tenants, err := client.CloudSnapshotTenants.List(ctx)
	if err != nil {
		t.Fatalf("CloudSnapshotTenants.List failed: %v", err)
	}

	t.Logf("Found %d cloud snapshot tenants", len(tenants))

	if len(tenants) == 0 {
		t.Log("No cloud snapshot tenants found - this is normal if snapshots haven't been scanned for tenants")
		return
	}

	// Log first tenant to verify field mapping
	first := tenants[0]
	t.Logf("First snapshot tenant: Key=%d, Name=%q, Tenant=%d, CloudSnapshot=%d",
		int(first.Key), first.Name, int(first.Tenant), int(first.CloudSnapshot))

	// Test Get by ID
	fetched, err := client.CloudSnapshotTenants.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("CloudSnapshotTenants.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("CloudSnapshotTenants.Get succeeded: Name=%q, Description=%q",
			fetched.Name, fetched.Description)
	}

	// Test ListBySnapshot
	if first.CloudSnapshot > 0 {
		snapTenants, err := client.CloudSnapshotTenants.ListBySnapshot(ctx, int(first.CloudSnapshot))
		if err != nil {
			t.Errorf("CloudSnapshotTenants.ListBySnapshot failed: %v", err)
		} else {
			t.Logf("Found %d tenants in snapshot %d", len(snapTenants), int(first.CloudSnapshot))
		}
	}

	// Pretty print first tenant for field verification
	prettyPrintWave7(t, "Sample CloudSnapshotTenant", first)
}

// TestWave7CloudSnapshotCRUD tests Create/Update/Delete operations for cloud snapshots.
// This creates a test snapshot to verify CRUD operations work correctly.
//
// WARNING: This test creates a real cloud snapshot which captures system state.
// The snapshot is deleted after the test, but be aware of storage implications.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave7CloudSnapshotCRUD
func TestWave7CloudSnapshotCRUD(t *testing.T) {
	client := setupTestClientWave7(t)
	ctx := context.Background()

	t.Log("Testing CloudSnapshot CRUD operations...")

	// Create a cloud snapshot
	retention := 60 // 1 minute retention for test
	minSnaps := 1
	snapshot, err := client.CloudSnapshots.Create(ctx, &vergeos.CloudSnapshotCreateRequest{
		Name:         "sdk-wave7-test-snapshot",
		Description:  "goVergeOS integration test snapshot - safe to delete",
		Retention:    &retention,
		MinSnapshots: &minSnaps,
	})
	if err != nil {
		t.Fatalf("CloudSnapshots.Create failed: %v", err)
	}
	snapshotID := int(snapshot.Key)
	t.Logf("Created cloud snapshot: [%d] %s", snapshotID, snapshot.Name)

	// Cleanup: delete test snapshot when done
	defer func() {
		t.Log("Cleaning up: deleting test snapshot...")
		if err := client.CloudSnapshots.Delete(ctx, snapshotID); err != nil {
			t.Logf("Warning: failed to delete test snapshot: %v", err)
		} else {
			t.Log("Test snapshot deleted successfully")
		}
	}()

	// Read
	snapshot, err = client.CloudSnapshots.Get(ctx, snapshotID)
	if err != nil {
		t.Fatalf("CloudSnapshots.Get failed: %v", err)
	}
	t.Logf("Read snapshot: [%d] %s Status=%q Expires=%d", snapshotID, snapshot.Name, snapshot.Status, snapshot.Expires)

	// Update
	newDesc := "Updated goVergeOS test snapshot"
	snapshot, err = client.CloudSnapshots.Update(ctx, snapshotID, &vergeos.CloudSnapshotUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Errorf("CloudSnapshots.Update failed: %v", err)
	} else if snapshot.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, snapshot.Description)
	} else {
		t.Logf("Updated snapshot description to: %q", snapshot.Description)
	}

	// Test Refresh action
	if err := client.CloudSnapshots.Refresh(ctx, snapshotID); err != nil {
		t.Logf("CloudSnapshots.Refresh: %v (may not be supported for all snapshot types)", err)
	} else {
		t.Log("CloudSnapshots.Refresh succeeded")
	}

	// Test FindVMs action
	if err := client.CloudSnapshots.FindVMs(ctx, snapshotID); err != nil {
		t.Logf("CloudSnapshots.FindVMs: %v (may take time to complete)", err)
	} else {
		t.Log("CloudSnapshots.FindVMs initiated")
	}

	// Test FindTenants action
	if err := client.CloudSnapshots.FindTenants(ctx, snapshotID); err != nil {
		t.Logf("CloudSnapshots.FindTenants: %v (may take time to complete)", err)
	} else {
		t.Log("CloudSnapshots.FindTenants initiated")
	}

	// Note: Clone test would create another snapshot, skipping to avoid excessive snapshots
	// t.Run("Clone", func(t *testing.T) {
	// 	err := client.CloudSnapshots.Clone(ctx, snapshotID, &vergeos.CloudSnapshotCloneOptions{
	// 		Name: "sdk-wave7-test-clone",
	// 	})
	// 	if err != nil {
	// 		t.Errorf("CloudSnapshots.Clone failed: %v", err)
	// 	}
	// })

	t.Log("CloudSnapshot CRUD test completed successfully")
}

// setupTestClientWave7 creates a client from environment variables
func setupTestClientWave7(t *testing.T) *vergeos.Client {
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		t.Skip("Skipping integration test: VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD must be set")
	}

	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true),
		vergeos.WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

// prettyPrintWave7 logs a struct as formatted JSON for field verification
func prettyPrintWave7(t *testing.T, label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}
