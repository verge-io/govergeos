//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave13Sprint3 tests the Sprint 3 services:
// - TenantSnapshots
// - TenantLayer2Networks
// - SiteSyncProfilePeriods
// - Logs
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave13
func TestWave13Sprint3(t *testing.T) {
	client := setupTestClientWave13(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("TenantSnapshots", func(t *testing.T) {
		testTenantSnapshots(t, ctx, client)
	})

	t.Run("TenantLayer2Networks", func(t *testing.T) {
		testTenantLayer2Networks(t, ctx, client)
	})

	t.Run("SiteSyncProfilePeriods", func(t *testing.T) {
		testSiteSyncProfilePeriods(t, ctx, client)
	})

	t.Run("Logs", func(t *testing.T) {
		testLogs(t, ctx, client)
	})
}

func testTenantSnapshots(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing TenantSnapshots service...")

	// List all tenant snapshots
	snapshots, err := client.TenantSnapshots.List(ctx)
	if err != nil {
		t.Fatalf("TenantSnapshots.List failed: %v", err)
	}
	t.Logf("Found %d tenant snapshots", len(snapshots))

	if len(snapshots) > 0 {
		prettyPrintWave13(t, "First Tenant Snapshot", snapshots[0])

		// Test Get by ID
		first := snapshots[0]
		t.Run("GetByID", func(t *testing.T) {
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

func testTenantLayer2Networks(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing TenantLayer2Networks service...")

	// List all tenant layer2 network assignments
	assignments, err := client.TenantLayer2Networks.List(ctx)
	if err != nil {
		t.Fatalf("TenantLayer2Networks.List failed: %v", err)
	}
	t.Logf("Found %d tenant layer2 network assignments", len(assignments))

	if len(assignments) > 0 {
		prettyPrintWave13(t, "First Tenant Layer2 Network Assignment", assignments[0])

		// Test Get by ID
		first := assignments[0]
		t.Run("GetByID", func(t *testing.T) {
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

func testSiteSyncProfilePeriods(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing SiteSyncProfilePeriods service...")

	// List all site sync profile periods
	periods, err := client.SiteSyncProfilePeriods.List(ctx)
	if err != nil {
		t.Fatalf("SiteSyncProfilePeriods.List failed: %v", err)
	}
	t.Logf("Found %d site sync profile periods", len(periods))

	if len(periods) > 0 {
		prettyPrintWave13(t, "First Site Sync Profile Period", periods[0])

		// Test Get by ID
		first := periods[0]
		t.Run("GetByID", func(t *testing.T) {
			period, err := client.SiteSyncProfilePeriods.Get(ctx, int(first.Key))
			if err != nil {
				t.Fatalf("SiteSyncProfilePeriods.Get failed: %v", err)
			}
			if int(period.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", period.Key, first.Key)
			}
		})
	}

	// Get outgoing syncs to test ListByOutgoingSync
	outgoingSyncs, err := client.SiteSyncsOutgoing.List(ctx)
	if err != nil || len(outgoingSyncs) == 0 {
		t.Log("No outgoing syncs available - skipping ListByOutgoingSync test")
		return
	}

	// Test ListByOutgoingSync
	t.Run("ListByOutgoingSync", func(t *testing.T) {
		syncID := int(outgoingSyncs[0].Key)
		syncPeriods, err := client.SiteSyncProfilePeriods.ListByOutgoingSync(ctx, syncID)
		if err != nil {
			t.Fatalf("SiteSyncProfilePeriods.ListByOutgoingSync failed: %v", err)
		}
		t.Logf("Found %d profile periods for outgoing sync %d", len(syncPeriods), syncID)
	})
}

func testLogs(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Logs service...")

	// List recent logs
	logs, err := client.Logs.List(ctx, vergeos.WithLimit(50))
	if err != nil {
		t.Fatalf("Logs.List failed: %v", err)
	}
	t.Logf("Found %d logs (limited to 50)", len(logs))

	if len(logs) > 0 {
		prettyPrintWave13(t, "First Log Entry", logs[0])

		// Test Get by ID
		first := logs[0]
		t.Run("GetByID", func(t *testing.T) {
			log, err := client.Logs.Get(ctx, int(first.Key))
			if err != nil {
				t.Fatalf("Logs.Get failed: %v", err)
			}
			if int(log.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", log.Key, first.Key)
			}
		})
	}

	// Test GetRecent
	t.Run("GetRecent", func(t *testing.T) {
		recent, err := client.Logs.GetRecent(ctx, 10)
		if err != nil {
			t.Fatalf("Logs.GetRecent failed: %v", err)
		}
		t.Logf("Got %d recent logs", len(recent))
	})

	// Test ListAudit
	t.Run("ListAudit", func(t *testing.T) {
		audits, err := client.Logs.ListAudit(ctx, vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Logs.ListAudit failed: %v", err)
		}
		t.Logf("Found %d audit logs (limited to 20)", len(audits))
	})

	// Test ListWarnings
	t.Run("ListWarnings", func(t *testing.T) {
		warnings, err := client.Logs.ListWarnings(ctx, vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Logs.ListWarnings failed: %v", err)
		}
		t.Logf("Found %d warning logs (limited to 20)", len(warnings))
	})

	// Test ListErrors
	t.Run("ListErrors", func(t *testing.T) {
		errors, err := client.Logs.ListErrors(ctx, vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Logs.ListErrors failed: %v", err)
		}
		t.Logf("Found %d error/critical logs (limited to 20)", len(errors))
	})

	// Test GetRecentErrors
	t.Run("GetRecentErrors", func(t *testing.T) {
		recentErrors, err := client.Logs.GetRecentErrors(ctx, 5)
		if err != nil {
			t.Fatalf("Logs.GetRecentErrors failed: %v", err)
		}
		t.Logf("Got %d recent errors", len(recentErrors))
	})

	// Test ListByObjectType
	t.Run("ListByObjectType_VM", func(t *testing.T) {
		vmLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeVM, vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Logs.ListByObjectType(vm) failed: %v", err)
		}
		t.Logf("Found %d VM logs (limited to 20)", len(vmLogs))
	})

	t.Run("ListByObjectType_System", func(t *testing.T) {
		sysLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeSystem, vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Logs.ListByObjectType(system) failed: %v", err)
		}
		t.Logf("Found %d system logs (limited to 20)", len(sysLogs))
	})

	// Test ListSince (last 24 hours)
	t.Run("ListSince", func(t *testing.T) {
		since := time.Now().Add(-24 * time.Hour).UnixMicro()
		recentLogs, err := client.Logs.ListSince(ctx, since, vergeos.WithLimit(50))
		if err != nil {
			t.Fatalf("Logs.ListSince failed: %v", err)
		}
		t.Logf("Found %d logs in the last 24 hours (limited to 50)", len(recentLogs))
	})

	// Test Search (search for common terms)
	t.Run("Search", func(t *testing.T) {
		searchResults, err := client.Logs.Search(ctx, "login", vergeos.WithLimit(20))
		if err != nil {
			t.Fatalf("Logs.Search failed: %v", err)
		}
		t.Logf("Found %d logs matching 'login' (limited to 20)", len(searchResults))
	})
}

// setupTestClientWave13 creates a client from environment variables
func setupTestClientWave13(t *testing.T) *vergeos.Client {
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

// prettyPrintWave13 logs a struct as formatted JSON for field verification
func prettyPrintWave13(t *testing.T, label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}
