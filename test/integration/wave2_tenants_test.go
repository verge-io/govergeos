//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave2Tenants tests the Wave 2 multi-tenancy services (Tenants, TenantNodes, TenantStorage)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave2
func TestWave2Tenants(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("Tenants", func(t *testing.T) {
		testTenants(t, ctx, client)
	})

	t.Run("TenantNodes", func(t *testing.T) {
		testTenantNodes(t, ctx, client)
	})

	t.Run("TenantStorage", func(t *testing.T) {
		testTenantStorage(t, ctx, client)
	})
}

func testTenants(t *testing.T, ctx context.Context, client *vergeos.Client) {
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
	if first.Key > 0 {
		fetched, err := client.Tenants.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Tenants.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Tenants.Get succeeded: Name=%q, Description=%q, URL=%q",
				fetched.Name, fetched.Description, fetched.URL)
		}
	}

	// Test GetByName
	if first.Name != "" {
		byName, err := client.Tenants.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("Tenants.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first tenant for field verification
	prettyPrint(t, "Sample Tenant", first)
}

func testTenantNodes(t *testing.T, ctx context.Context, client *vergeos.Client) {
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
	if first.Key > 0 {
		fetched, err := client.TenantNodes.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("TenantNodes.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("TenantNodes.Get succeeded: Name=%q, NodeID=%d, Machine=%d",
				fetched.Name, fetched.NodeID, int(fetched.Machine))
		}
	}

	// Test ListByTenant
	if first.Tenant > 0 {
		tenantNodes, err := client.TenantNodes.ListByTenant(ctx, int(first.Tenant))
		if err != nil {
			t.Errorf("TenantNodes.ListByTenant failed: %v", err)
		} else {
			t.Logf("Found %d nodes in tenant %d", len(tenantNodes), int(first.Tenant))
		}

		// Test GetByName
		if first.Name != "" {
			byName, err := client.TenantNodes.GetByName(ctx, int(first.Tenant), first.Name)
			if err != nil {
				t.Errorf("TenantNodes.GetByName failed: %v", err)
			} else {
				t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
			}
		}
	}

	// Pretty print first node for field verification
	prettyPrint(t, "Sample TenantNode", first)
}

func testTenantStorage(t *testing.T, ctx context.Context, client *vergeos.Client) {
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
	if first.Key > 0 {
		fetched, err := client.TenantStorage.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("TenantStorage.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("TenantStorage.Get succeeded: Provisioned=%d, Used=%d, Allocated=%d",
				fetched.Provisioned, fetched.Used, fetched.Allocated)
		}
	}

	// Test ListByTenant
	if first.Tenant > 0 {
		tenantStorage, err := client.TenantStorage.ListByTenant(ctx, int(first.Tenant))
		if err != nil {
			t.Errorf("TenantStorage.ListByTenant failed: %v", err)
		} else {
			t.Logf("Found %d storage allocations for tenant %d", len(tenantStorage), int(first.Tenant))
		}
	}

	// Pretty print first storage allocation for field verification
	prettyPrint(t, "Sample TenantStorage", first)
}

// TestWave2TenantsCRUD tests Create/Update/Delete operations for Wave 2 tenant services.
// WARNING: This test creates actual tenants which consume resources.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave2TenantsCRUD
func TestWave2TenantsCRUD(t *testing.T) {
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

	// Note: Not testing power operations or tenant nodes/storage creation
	// as those require more setup and have longer execution times.
	// The read tests above verify those services work with existing tenants.

	t.Log("Tenant CRUD test completed successfully")
}

// TestWave2TenantPowerOperations tests power operations on tenants.
// This test is separate because power operations take time.
// It requires an existing tenant that is safe to power cycle.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave2TenantPowerOperations
func TestWave2TenantPowerOperations(t *testing.T) {
	t.Skip("Skipping power operations test - enable manually for testing specific tenants")

	client := setupTestClient(t)
	ctx := context.Background()

	// Find a non-production tenant to test with
	tenants, err := client.Tenants.List(ctx, vergeos.WithFilter("name ct 'test'"))
	if err != nil {
		t.Fatalf("Failed to list tenants: %v", err)
	}

	if len(tenants) == 0 {
		t.Skip("No test tenants found for power operation testing")
	}

	tenant := tenants[0]
	tenantID := int(tenant.Key)
	t.Logf("Testing power operations on tenant %d (%s)", tenantID, tenant.Name)

	// Test IsolateOn/IsolateOff
	t.Log("Testing IsolateOn...")
	err = client.Tenants.IsolateOn(ctx, tenantID)
	if err != nil {
		t.Errorf("Tenants.IsolateOn failed: %v", err)
	} else {
		t.Log("IsolateOn succeeded")
	}

	t.Log("Testing IsolateOff...")
	err = client.Tenants.IsolateOff(ctx, tenantID)
	if err != nil {
		t.Errorf("Tenants.IsolateOff failed: %v", err)
	} else {
		t.Log("IsolateOff succeeded")
	}
}
