//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestTenantsList tests the Tenants service list and get operations.
func TestTenantsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Tenants service...")

	// List all tenants
	tenants, err := client.Tenants.List(ctx)
	if err != nil {
		t.Fatalf("Tenants.List failed: %v", err)
	}

	t.Logf("Found %d tenants", len(tenants))

	if len(tenants) == 0 {
		t.Log("No tenants found - this is expected on systems without multi-tenancy")
		return
	}

	// Log first tenant to verify field mapping
	first := tenants[0]
	t.Logf("First tenant: Key=%d, Name=%q, UUID=%q, VNet=%d, Isolate=%v",
		int(first.Key), first.Name, first.UUID, int(first.VNet), first.Isolate)

	// Test Get
	t.Run("Get", func(t *testing.T) {
		if first.Key == 0 {
			t.Skip("No tenant key available")
		}
		fetched, err := client.Tenants.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Tenants.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Tenants.Get succeeded: Name=%q, Description=%q, URL=%q",
				fetched.Name, fetched.Description, fetched.URL)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" {
			t.Skip("No tenant name available")
		}
		byName, err := client.Tenants.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("Tenants.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first tenant for field verification
	prettyPrint(t, "Sample Tenant", first)
}

// TestTenantNodesList tests the TenantNodes service.
func TestTenantNodesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing TenantNodes service...")

	// List all tenant nodes
	nodes, err := client.TenantNodes.List(ctx)
	if err != nil {
		t.Fatalf("TenantNodes.List failed: %v", err)
	}

	t.Logf("Found %d tenant nodes", len(nodes))

	if len(nodes) == 0 {
		t.Log("No tenant nodes found")
		return
	}

	// Log first node to verify field mapping
	first := nodes[0]
	t.Logf("First tenant node: Key=%d, Name=%q, Tenant=%d, CPUCores=%d, RAM=%d MB",
		int(first.Key), first.Name, int(first.Tenant), first.CPUCores, first.RAM)

	// Test Get
	t.Run("Get", func(t *testing.T) {
		if first.Key == 0 {
			t.Skip("No node key available")
		}
		fetched, err := client.TenantNodes.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("TenantNodes.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("TenantNodes.Get succeeded: Name=%q, NodeID=%d, Machine=%d",
				fetched.Name, fetched.NodeID, int(fetched.Machine))
		}
	})

	// Test ListByTenant
	t.Run("ListByTenant", func(t *testing.T) {
		if first.Tenant == 0 {
			t.Skip("No tenant ID available")
		}
		tenantNodes, err := client.TenantNodes.ListByTenant(ctx, int(first.Tenant))
		if err != nil {
			t.Errorf("TenantNodes.ListByTenant failed: %v", err)
		} else {
			t.Logf("Found %d nodes in tenant %d", len(tenantNodes), int(first.Tenant))
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Tenant == 0 || first.Name == "" {
			t.Skip("No tenant ID or name available")
		}
		byName, err := client.TenantNodes.GetByName(ctx, int(first.Tenant), first.Name)
		if err != nil {
			t.Errorf("TenantNodes.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	prettyPrint(t, "Sample TenantNode", first)
}

// TestTenantStorageList tests the TenantStorage service.
func TestTenantStorageList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing TenantStorage service...")

	// List all tenant storage allocations
	storage, err := client.TenantStorage.List(ctx)
	if err != nil {
		t.Fatalf("TenantStorage.List failed: %v", err)
	}

	t.Logf("Found %d tenant storage allocations", len(storage))

	if len(storage) == 0 {
		t.Log("No tenant storage allocations found")
		return
	}

	// Log first storage allocation to verify field mapping
	first := storage[0]
	t.Logf("First tenant storage: Key=%d, Tenant=%d, Tier=%d, Provisioned=%d bytes",
		int(first.Key), int(first.Tenant), int(first.Tier), first.Provisioned)

	// Test Get
	t.Run("Get", func(t *testing.T) {
		if first.Key == 0 {
			t.Skip("No storage key available")
		}
		fetched, err := client.TenantStorage.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("TenantStorage.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("TenantStorage.Get succeeded: Provisioned=%d, Used=%d, Allocated=%d",
				fetched.Provisioned, fetched.Used, fetched.Allocated)
		}
	})

	// Test ListByTenant
	t.Run("ListByTenant", func(t *testing.T) {
		if first.Tenant == 0 {
			t.Skip("No tenant ID available")
		}
		tenantStorage, err := client.TenantStorage.ListByTenant(ctx, int(first.Tenant))
		if err != nil {
			t.Errorf("TenantStorage.ListByTenant failed: %v", err)
		} else {
			t.Logf("Found %d storage allocations for tenant %d", len(tenantStorage), int(first.Tenant))
		}
	})

	prettyPrint(t, "Sample TenantStorage", first)
}

// TestTenantSnapshotsList tests the TenantSnapshots service.
func TestTenantSnapshotsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing TenantSnapshots service...")

	// List all tenant snapshots
	snapshots, err := client.TenantSnapshots.List(ctx)
	if err != nil {
		t.Fatalf("TenantSnapshots.List failed: %v", err)
	}
	t.Logf("Found %d tenant snapshots", len(snapshots))

	if len(snapshots) > 0 {
		first := snapshots[0]
		prettyPrint(t, "First Tenant Snapshot", first)

		// Test Get by ID
		t.Run("Get", func(t *testing.T) {
			snapshot, err := client.TenantSnapshots.Get(ctx, int(first.Key))
			if err != nil {
				t.Fatalf("TenantSnapshots.Get failed: %v", err)
			}
			if int(snapshot.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", snapshot.Key, first.Key)
			}
		})

		// Test GetByName if we have a tenant reference
		if int(first.Tenant) > 0 && first.Name != "" {
			t.Run("GetByName", func(t *testing.T) {
				snapshot, err := client.TenantSnapshots.GetByName(ctx, int(first.Tenant), first.Name)
				if err != nil {
					t.Fatalf("TenantSnapshots.GetByName failed: %v", err)
				}
				if int(snapshot.Key) != int(first.Key) {
					t.Errorf("Key mismatch: got %d, want %d", snapshot.Key, first.Key)
				}
			})
		}
	}

	// Test ListExpiring (within next 30 days)
	t.Run("ListExpiring", func(t *testing.T) {
		expiring, err := client.TenantSnapshots.ListExpiring(ctx, 30)
		if err != nil {
			t.Fatalf("TenantSnapshots.ListExpiring failed: %v", err)
		}
		t.Logf("Found %d tenant snapshots expiring within 30 days", len(expiring))
	})

	// Get tenants to test ListByTenant
	tenants, err := client.Tenants.List(ctx)
	if err != nil || len(tenants) == 0 {
		t.Log("No tenants available - skipping ListByTenant test")
		return
	}

	// Test ListByTenant
	t.Run("ListByTenant", func(t *testing.T) {
		tenantID := int(tenants[0].Key)
		tenantSnapshots, err := client.TenantSnapshots.ListByTenant(ctx, tenantID)
		if err != nil {
			t.Fatalf("TenantSnapshots.ListByTenant failed: %v", err)
		}
		t.Logf("Found %d snapshots for tenant %d", len(tenantSnapshots), tenantID)
	})
}

// TestTenantLayer2List tests the TenantLayer2Networks service.
func TestTenantLayer2List(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing TenantLayer2Networks service...")

	// List all tenant layer2 network assignments
	assignments, err := client.TenantLayer2Networks.List(ctx)
	if err != nil {
		t.Fatalf("TenantLayer2Networks.List failed: %v", err)
	}
	t.Logf("Found %d tenant layer2 network assignments", len(assignments))

	if len(assignments) > 0 {
		first := assignments[0]
		prettyPrint(t, "First Tenant Layer2 Network Assignment", first)

		// Test Get by ID
		t.Run("Get", func(t *testing.T) {
			assignment, err := client.TenantLayer2Networks.Get(ctx, int(first.Key))
			if err != nil {
				t.Fatalf("TenantLayer2Networks.Get failed: %v", err)
			}
			if int(assignment.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", assignment.Key, first.Key)
			}
		})

		// Test GetByTenantAndNetwork
		if int(first.Tenant) > 0 && int(first.VNet) > 0 {
			t.Run("GetByTenantAndNetwork", func(t *testing.T) {
				assignment, err := client.TenantLayer2Networks.GetByTenantAndNetwork(ctx, int(first.Tenant), int(first.VNet))
				if err != nil {
					t.Fatalf("TenantLayer2Networks.GetByTenantAndNetwork failed: %v", err)
				}
				if int(assignment.Key) != int(first.Key) {
					t.Errorf("Key mismatch: got %d, want %d", assignment.Key, first.Key)
				}
			})
		}
	}

	// Get tenants to test ListByTenant
	tenants, err := client.Tenants.List(ctx)
	if err != nil || len(tenants) == 0 {
		t.Log("No tenants available - skipping ListByTenant test")
		return
	}

	// Test ListByTenant
	t.Run("ListByTenant", func(t *testing.T) {
		tenantID := int(tenants[0].Key)
		tenantAssignments, err := client.TenantLayer2Networks.ListByTenant(ctx, tenantID)
		if err != nil {
			t.Fatalf("TenantLayer2Networks.ListByTenant failed: %v", err)
		}
		t.Logf("Found %d layer2 network assignments for tenant %d", len(tenantAssignments), tenantID)
	})

	// Get networks to test ListByNetwork
	networks, err := client.Networks.List(ctx)
	if err != nil || len(networks) == 0 {
		t.Log("No networks available - skipping ListByNetwork test")
		return
	}

	// Test ListByNetwork
	t.Run("ListByNetwork", func(t *testing.T) {
		networkID := int(networks[0].ID)
		networkAssignments, err := client.TenantLayer2Networks.ListByNetwork(ctx, networkID)
		if err != nil {
			t.Fatalf("TenantLayer2Networks.ListByNetwork failed: %v", err)
		}
		t.Logf("Found %d tenant assignments for network %d", len(networkAssignments), networkID)
	})
}

// TestTenantStatusList tests the TenantStatus service.
func TestTenantStatusList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing TenantStatus service...")

	// List all tenant statuses
	statuses, err := client.TenantStatus.List(ctx)
	if err != nil {
		t.Fatalf("TenantStatus.List failed: %v", err)
	}

	t.Logf("Found %d tenant statuses", len(statuses))

	if len(statuses) == 0 {
		t.Log("No tenant statuses found - this is expected on systems without tenants")
		return
	}

	// Log first status to verify field mapping
	first := statuses[0]
	t.Logf("First tenant status: Key=%d, Tenant=%d, Status=%q, State=%q, Running=%v",
		int(first.Key), first.Tenant, first.Status, first.State, first.Running)

	// Test Get by tenant key
	t.Run("Get", func(t *testing.T) {
		if first.Tenant == 0 {
			t.Skip("No tenant reference available")
		}
		fetched, err := client.TenantStatus.Get(ctx, first.Tenant)
		if err != nil {
			t.Errorf("TenantStatus.Get(%d) failed: %v", first.Tenant, err)
		} else {
			t.Logf("TenantStatus.Get succeeded: Status=%q, State=%q, Running=%v, Starting=%v, Stopping=%v",
				fetched.Status, fetched.State, fetched.Running, fetched.Starting, fetched.Stopping)
		}
	})

	// Test GetByKey
	t.Run("GetByKey", func(t *testing.T) {
		if first.Key == 0 {
			t.Skip("No status key available")
		}
		fetched, err := client.TenantStatus.GetByKey(ctx, int(first.Key))
		if err != nil {
			t.Errorf("TenantStatus.GetByKey(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("TenantStatus.GetByKey succeeded: Tenant=%d, Status=%q", fetched.Tenant, fetched.Status)
		}
	})

	prettyPrint(t, "Sample TenantStatus", first)
}

// TestTenantStatsHistoryShort tests the TenantStatsHistoryShort service.
func TestTenantStatsHistoryShort(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Get tenants for testing
	tenants, err := client.Tenants.List(ctx)
	if err != nil {
		t.Fatalf("Tenants.List failed: %v", err)
	}
	if len(tenants) == 0 {
		t.Skip("No tenants found - skipping TenantStatsHistoryShort tests")
	}
	tenantID := int(tenants[0].Key)
	t.Logf("Testing with tenant %d (%s)", tenantID, tenants[0].Name)

	t.Run("List", func(t *testing.T) {
		stats, err := client.TenantStatsHistoryShort.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("TenantStatsHistoryShort.List failed: %v", err)
		}
		t.Logf("Found %d short-term history records (limited to 5)", len(stats))

		if len(stats) > 0 {
			first := stats[0]
			t.Logf("First record: Key=%d, Tenant=%d, Timestamp=%d",
				int(first.Key), first.Tenant, first.Timestamp)
			t.Logf("Resources: RAMUsed=%d, VRAMUsed=%d, TotalCPU=%d, CoreCount=%d, IPCount=%d",
				first.RAMUsed, first.VRAMUsed, first.TotalCPU, first.CoreCount, first.IPCount)
			prettyPrint(t, "Sample TenantStatsHistoryShort", first)
		}
	})

	t.Run("ListByTenant", func(t *testing.T) {
		stats, err := client.TenantStatsHistoryShort.ListByTenant(ctx, tenantID, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("TenantStatsHistoryShort.ListByTenant(%d) failed: %v", tenantID, err)
		}
		t.Logf("Found %d short-term records for tenant %d", len(stats), tenantID)
	})

	t.Run("GetLatest", func(t *testing.T) {
		latest, err := client.TenantStatsHistoryShort.GetLatest(ctx, tenantID)
		if err != nil {
			if vergeos.IsNotFoundError(err) {
				t.Logf("No recent stats for tenant %d (this is normal for new/idle tenants)", tenantID)
			} else {
				t.Fatalf("TenantStatsHistoryShort.GetLatest(%d) failed: %v", tenantID, err)
			}
		} else {
			t.Logf("Latest stats for tenant %d: Timestamp=%d, RAMUsed=%d, RAMAllocated=%d, RAMPct=%d%%",
				tenantID, latest.Timestamp, latest.RAMUsed, latest.RAMAllocated, latest.RAMPct)
		}
	})

	t.Run("Get", func(t *testing.T) {
		stats, err := client.TenantStatsHistoryShort.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(stats) == 0 {
			t.Skip("No stats history records available")
		}

		first := stats[0]
		fetched, err := client.TenantStatsHistoryShort.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("TenantStatsHistoryShort.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("TenantStatsHistoryShort.Get succeeded: Tenant=%d, Timestamp=%d", fetched.Tenant, fetched.Timestamp)
	})
}

// TestTenantsCRUD tests Create/Update/Delete operations for Tenant services.
func TestTenantsCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Tenants CRUD operations...")

	// Create a test tenant
	t.Log("Creating test tenant...")
	tenant, err := client.Tenants.Create(ctx, &vergeos.TenantCreateRequest{
		Name:        "sdk-test-tenant",
		Description: "goVergeOS integration test tenant - safe to delete",
		Password:    "TestPassword123!",
	})
	if err != nil {
		t.Fatalf("Tenants.Create failed: %v", err)
	}
	tenantID := int(tenant.Key)
	t.Logf("Created tenant: [%d] %s (UUID: %s)", tenantID, tenant.Name, tenant.UUID)

	// Cleanup: delete test tenant when done
	defer func() {
		t.Log("Cleaning up: deleting test tenant...")
		if err := client.Tenants.Delete(ctx, tenantID); err != nil {
			t.Logf("Warning: failed to delete test tenant: %v", err)
		} else {
			t.Log("Test tenant deleted successfully")
		}
	}()

	// Read
	tenant, err = client.Tenants.Get(ctx, tenantID)
	if err != nil {
		t.Fatalf("Tenants.Get failed: %v", err)
	}
	t.Logf("Read tenant: [%d] %s (VNet: %d)", tenantID, tenant.Name, int(tenant.VNet))

	// Update
	newDesc := "Updated goVergeOS test tenant description"
	tenant, err = client.Tenants.Update(ctx, tenantID, &vergeos.TenantUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Tenants.Update failed: %v", err)
	}
	if tenant.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, tenant.Description)
	} else {
		t.Logf("Updated tenant description: %q", tenant.Description)
	}

	// Test GetByName
	byName, err := client.Tenants.GetByName(ctx, "sdk-test-tenant")
	if err != nil {
		t.Errorf("Tenants.GetByName failed: %v", err)
	} else {
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	}

	// Test isolation operations
	t.Run("IsolateOn", func(t *testing.T) {
		err = client.Tenants.IsolateOn(ctx, tenantID)
		if err != nil {
			t.Errorf("Tenants.IsolateOn failed: %v", err)
		} else {
			t.Log("IsolateOn succeeded")
		}
	})

	t.Run("IsolateOff", func(t *testing.T) {
		err = client.Tenants.IsolateOff(ctx, tenantID)
		if err != nil {
			t.Errorf("Tenants.IsolateOff failed: %v", err)
		} else {
			t.Log("IsolateOff succeeded")
		}
	})

	t.Log("Tenant CRUD test completed successfully")
}
