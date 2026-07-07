//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/macstadium/govergeos"
)

// TestSitesList tests the Sites service.
func TestSitesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.Sites.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Sites.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Sites.Get succeeded: Name=%q, Domain=%q, City=%q, Country=%q",
				fetched.Name, fetched.Domain, fetched.City, fetched.Country)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" {
			t.Skip("No site name available")
		}
		byName, err := client.Sites.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("Sites.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Test GetBySiteID if we have a SHA1 ID
	t.Run("GetBySiteID", func(t *testing.T) {
		if first.ID == "" {
			t.Skip("No site ID available")
		}
		bySiteID, err := client.Sites.GetBySiteID(ctx, first.ID)
		if err != nil {
			t.Errorf("Sites.GetBySiteID failed: %v", err)
		} else {
			t.Logf("GetBySiteID succeeded: Key=%d", int(bySiteID.Key))
		}
	})

	// Pretty print first site for field verification
	prettyPrint(t, "Sample Site", first)
}

// TestSiteSyncsIncomingList tests the SiteSyncsIncoming service.
func TestSiteSyncsIncomingList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.SiteSyncsIncoming.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("SiteSyncsIncoming.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("SiteSyncsIncoming.Get succeeded: Name=%q, PublicIP=%q, ForceTier=%q",
				fetched.Name, fetched.PublicIP, fetched.ForceTier)
		}
	})

	// Test ListBySite
	t.Run("ListBySite", func(t *testing.T) {
		if first.Site == 0 {
			t.Skip("No Site ID available")
		}
		siteSyncs, err := client.SiteSyncsIncoming.ListBySite(ctx, int(first.Site))
		if err != nil {
			t.Errorf("SiteSyncsIncoming.ListBySite failed: %v", err)
		} else {
			t.Logf("Found %d incoming syncs for site %d", len(siteSyncs), int(first.Site))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.Site == 0 {
			t.Skip("No name or Site ID available")
		}
		byName, err := client.SiteSyncsIncoming.GetByName(ctx, int(first.Site), first.Name)
		if err != nil {
			t.Errorf("SiteSyncsIncoming.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Test GetBySyncID if we have a SHA1 ID
	t.Run("GetBySyncID", func(t *testing.T) {
		if first.SyncID == "" {
			t.Skip("No SyncID available")
		}
		bySyncID, err := client.SiteSyncsIncoming.GetBySyncID(ctx, first.SyncID)
		if err != nil {
			t.Errorf("SiteSyncsIncoming.GetBySyncID failed: %v", err)
		} else {
			t.Logf("GetBySyncID succeeded: Key=%d", int(bySyncID.Key))
		}
	})

	// Pretty print first sync for field verification
	prettyPrint(t, "Sample SiteSyncIncoming", first)
}

// TestSiteSyncsOutgoingList tests the SiteSyncsOutgoing service.
func TestSiteSyncsOutgoingList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.SiteSyncsOutgoing.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("SiteSyncsOutgoing.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("SiteSyncsOutgoing.Get succeeded: Name=%q, DestinationTier=%q, Encryption=%v, Compression=%v",
				fetched.Name, fetched.DestinationTier, fetched.Encryption, fetched.Compression)
		}
	})

	// Test ListBySite
	t.Run("ListBySite", func(t *testing.T) {
		if first.Site == 0 {
			t.Skip("No Site ID available")
		}
		siteSyncs, err := client.SiteSyncsOutgoing.ListBySite(ctx, int(first.Site))
		if err != nil {
			t.Errorf("SiteSyncsOutgoing.ListBySite failed: %v", err)
		} else {
			t.Logf("Found %d outgoing syncs for site %d", len(siteSyncs), int(first.Site))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" || first.Site == 0 {
			t.Skip("No name or Site ID available")
		}
		byName, err := client.SiteSyncsOutgoing.GetByName(ctx, int(first.Site), first.Name)
		if err != nil {
			t.Errorf("SiteSyncsOutgoing.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first sync for field verification
	prettyPrint(t, "Sample SiteSyncOutgoing", first)
}

// TestCloudSnapshotsList tests the CloudSnapshots service.
func TestCloudSnapshotsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.CloudSnapshots.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("CloudSnapshots.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("CloudSnapshots.Get succeeded: Name=%q, Description=%q, Provider=%v, RemoteSync=%v",
				fetched.Name, fetched.Description, fetched.Provider, fetched.RemoteSync)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" {
			t.Skip("No snapshot name available")
		}
		byName, err := client.CloudSnapshots.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("CloudSnapshots.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Test ListExpiring
	t.Run("ListExpiring", func(t *testing.T) {
		expiring, err := client.CloudSnapshots.ListExpiring(ctx)
		if err != nil {
			t.Errorf("CloudSnapshots.ListExpiring failed: %v", err)
		} else {
			t.Logf("Found %d expiring cloud snapshots", len(expiring))
		}
	})

	// Test ListLocal
	t.Run("ListLocal", func(t *testing.T) {
		local, err := client.CloudSnapshots.ListLocal(ctx)
		if err != nil {
			t.Errorf("CloudSnapshots.ListLocal failed: %v", err)
		} else {
			t.Logf("Found %d local cloud snapshots (non-provider)", len(local))
		}
	})

	// Test ListByProfile if we have a profile
	t.Run("ListByProfile", func(t *testing.T) {
		if first.SnapshotProfile == 0 {
			t.Skip("No SnapshotProfile ID available")
		}
		profileSnaps, err := client.CloudSnapshots.ListByProfile(ctx, int(first.SnapshotProfile))
		if err != nil {
			t.Errorf("CloudSnapshots.ListByProfile failed: %v", err)
		} else {
			t.Logf("Found %d snapshots for profile %d", len(profileSnaps), int(first.SnapshotProfile))
		}
	})

	// Pretty print first snapshot for field verification
	prettyPrint(t, "Sample CloudSnapshot", first)
}

// TestCloudSnapshotVMsList tests the CloudSnapshotVMs service.
func TestCloudSnapshotVMsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.CloudSnapshotVMs.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("CloudSnapshotVMs.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("CloudSnapshotVMs.Get succeeded: Name=%q, Description=%q, OSFamily=%q",
				fetched.Name, fetched.Description, fetched.OSFamily)
		}
	})

	// Test ListBySnapshot
	t.Run("ListBySnapshot", func(t *testing.T) {
		if first.CloudSnapshot == 0 {
			t.Skip("No CloudSnapshot ID available")
		}
		snapVMs, err := client.CloudSnapshotVMs.ListBySnapshot(ctx, int(first.CloudSnapshot))
		if err != nil {
			t.Errorf("CloudSnapshotVMs.ListBySnapshot failed: %v", err)
		} else {
			t.Logf("Found %d VMs in snapshot %d", len(snapVMs), int(first.CloudSnapshot))
		}
	})

	// Pretty print first VM for field verification
	prettyPrint(t, "Sample CloudSnapshotVM", first)
}

// TestCloudSnapshotTenantsList tests the CloudSnapshotTenants service.
func TestCloudSnapshotTenantsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.CloudSnapshotTenants.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("CloudSnapshotTenants.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("CloudSnapshotTenants.Get succeeded: Name=%q, Description=%q",
				fetched.Name, fetched.Description)
		}
	})

	// Test ListBySnapshot
	t.Run("ListBySnapshot", func(t *testing.T) {
		if first.CloudSnapshot == 0 {
			t.Skip("No CloudSnapshot ID available")
		}
		snapTenants, err := client.CloudSnapshotTenants.ListBySnapshot(ctx, int(first.CloudSnapshot))
		if err != nil {
			t.Errorf("CloudSnapshotTenants.ListBySnapshot failed: %v", err)
		} else {
			t.Logf("Found %d tenants in snapshot %d", len(snapTenants), int(first.CloudSnapshot))
		}
	})

	// Pretty print first tenant for field verification
	prettyPrint(t, "Sample CloudSnapshotTenant", first)
}

// TestCloudSnapshotsCRUD tests Create/Update/Delete operations for cloud snapshots.
// WARNING: This test creates a real cloud snapshot which captures system state.
func TestCloudSnapshotsCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing CloudSnapshot CRUD operations...")

	// Create a cloud snapshot
	retention := 60 // 1 minute retention for test
	minSnaps := 1
	snapshot, err := client.CloudSnapshots.Create(ctx, &vergeos.CloudSnapshotCreateRequest{
		Name:         "sdk-dr-test-snapshot",
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

	t.Log("CloudSnapshot CRUD test completed successfully")
}
