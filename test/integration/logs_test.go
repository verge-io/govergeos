//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestLogsList tests the Logs service list and query operations.
func TestLogsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Logs service...")

	// List recent logs
	logs, err := client.Logs.List(ctx, vergeos.WithLimit(50))
	if err != nil {
		t.Fatalf("Logs.List failed: %v", err)
	}
	t.Logf("Found %d logs (limited to 50)", len(logs))

	if len(logs) > 0 {
		first := logs[0]
		prettyPrint(t, "First Log Entry", first)

		// Test Get by ID
		t.Run("Get", func(t *testing.T) {
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
