//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestBillingNASAntivirusSharedObjects reads billing records, NAS service
// antivirus settings, and shared objects. It does not generate a report,
// change antivirus settings, or create a share.
func TestBillingNASAntivirusSharedObjects(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("Billing", func(t *testing.T) {
		records, err := client.Billing.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("Billing.List failed: %v", err)
		}
		t.Logf("Found %d billing records", len(records))
		if len(records) == 0 {
			return
		}
		got, err := client.Billing.Get(ctx, records[0].Key.Int())
		if err != nil {
			t.Fatalf("Billing.Get(%d) failed: %v", records[0].Key.Int(), err)
		}
		if got.Key != records[0].Key {
			t.Errorf("Get key %v, list key %v", got.Key, records[0].Key)
		}
		latest, err := client.Billing.GetLatest(ctx)
		if err != nil {
			t.Fatalf("Billing.GetLatest failed: %v", err)
		}
		t.Logf("Latest billing record %d created %d cores %d/%d", latest.Key.Int(), latest.Created, latest.UsedCores, latest.TotalCores)
	})

	t.Run("NASServiceAntivirus", func(t *testing.T) {
		rows, err := client.NASServiceAntivirus.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("NASServiceAntivirus.List failed: %v", err)
		}
		t.Logf("Found %d NAS service antivirus rows", len(rows))
		if len(rows) == 0 {
			return
		}
		got, err := client.NASServiceAntivirus.Get(ctx, rows[0].Key.Int())
		if err != nil {
			t.Fatalf("NASServiceAntivirus.Get(%d) failed: %v", rows[0].Key.Int(), err)
		}
		if got.Service != rows[0].Service {
			t.Errorf("Get service %v, list service %v", got.Service, rows[0].Service)
		}
	})

	t.Run("SharedObjects", func(t *testing.T) {
		objects, err := client.SharedObjects.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("SharedObjects.List failed: %v", err)
		}
		t.Logf("Found %d shared objects", len(objects))
		if len(objects) == 0 {
			return
		}
		got, err := client.SharedObjects.Get(ctx, objects[0].Key.Int())
		if err != nil {
			t.Fatalf("SharedObjects.Get(%d) failed: %v", objects[0].Key.Int(), err)
		}
		if got.Name != objects[0].Name {
			t.Errorf("Get name %q, list name %q", got.Name, objects[0].Name)
		}
	})
}
