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

// TestWave11NASExtended tests the extended NAS services added in Wave 2 of the convergence plan:
// - NASServices (vm_services)
// - NASServiceUsers (vm_service_users)
// - VolumeSyncs (volume_syncs)
// - VolumeSnapshots (volume_snapshots)
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave11
func TestWave11NASExtended(t *testing.T) {
	client := setupTestClientWave11(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("NASServices", func(t *testing.T) {
		testNASServices(t, ctx, client)
	})

	t.Run("NASServiceUsers", func(t *testing.T) {
		testNASServiceUsers(t, ctx, client)
	})

	t.Run("VolumeSyncs", func(t *testing.T) {
		testVolumeSyncs(t, ctx, client)
	})

	t.Run("VolumeSnapshots", func(t *testing.T) {
		testVolumeSnapshots(t, ctx, client)
	})

	t.Run("Permissions", func(t *testing.T) {
		testPermissions(t, ctx, client)
	})
}

func testNASServices(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing NASServices service...")

	// List all NAS services
	services, err := client.NASServices.List(ctx)
	if err != nil {
		t.Fatalf("NASServices.List failed: %v", err)
	}
	t.Logf("Found %d NAS services", len(services))

	if len(services) == 0 {
		t.Log("No NAS services found - skipping Get tests")
		return
	}

	// Get the first service
	first := services[0]
	prettyPrintWave11(t, "First NAS Service", first)

	// Test Get by ID
	t.Run("GetByID", func(t *testing.T) {
		service, err := client.NASServices.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("NASServices.Get failed: %v", err)
		}
		if int(service.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", service.Key, first.Key)
		}
	})

	// Test GetByVM if we have a VM reference
	if int(first.VM) > 0 {
		t.Run("GetByVM", func(t *testing.T) {
			service, err := client.NASServices.GetByVM(ctx, int(first.VM))
			if err != nil {
				t.Fatalf("NASServices.GetByVM failed: %v", err)
			}
			if int(service.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", service.Key, first.Key)
			}
		})
	}

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		service, err := client.NASServices.GetByName(ctx, first.Name)
		if err != nil {
			t.Fatalf("NASServices.GetByName failed: %v", err)
		}
		if int(service.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", service.Key, first.Key)
		}
	})
}

func testNASServiceUsers(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing NASServiceUsers service...")

	// List all NAS service users
	users, err := client.NASServiceUsers.List(ctx)
	if err != nil {
		t.Fatalf("NASServiceUsers.List failed: %v", err)
	}
	t.Logf("Found %d NAS service users", len(users))

	if len(users) > 0 {
		prettyPrintWave11(t, "First NAS Service User", users[0])
	}

	// Get a NAS service to test user operations
	services, err := client.NASServices.List(ctx)
	if err != nil || len(services) == 0 {
		t.Log("No NAS services available - skipping user CRUD tests")
		return
	}

	// Test ListByService
	serviceID := int(services[0].Key)
	t.Run("ListByService", func(t *testing.T) {
		serviceUsers, err := client.NASServiceUsers.ListByService(ctx, serviceID)
		if err != nil {
			t.Fatalf("NASServiceUsers.ListByService failed: %v", err)
		}
		t.Logf("Found %d users for service %d", len(serviceUsers), serviceID)
	})
}

func testVolumeSyncs(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VolumeSyncs service...")

	// List all volume syncs
	syncs, err := client.VolumeSyncs.List(ctx)
	if err != nil {
		t.Fatalf("VolumeSyncs.List failed: %v", err)
	}
	t.Logf("Found %d volume syncs", len(syncs))

	if len(syncs) > 0 {
		prettyPrintWave11(t, "First Volume Sync", syncs[0])

		// Test Get by ID
		first := syncs[0]
		t.Run("GetByID", func(t *testing.T) {
			sync, err := client.VolumeSyncs.Get(ctx, first.ID)
			if err != nil {
				t.Fatalf("VolumeSyncs.Get failed: %v", err)
			}
			if sync.ID != first.ID {
				t.Errorf("ID mismatch: got %s, want %s", sync.ID, first.ID)
			}
		})
	}

	// Test ListEnabled
	t.Run("ListEnabled", func(t *testing.T) {
		enabled, err := client.VolumeSyncs.ListEnabled(ctx)
		if err != nil {
			t.Fatalf("VolumeSyncs.ListEnabled failed: %v", err)
		}
		t.Logf("Found %d enabled volume syncs", len(enabled))
	})

	// Get a NAS service to test sync operations by service
	services, err := client.NASServices.List(ctx)
	if err != nil || len(services) == 0 {
		t.Log("No NAS services available - skipping ListByService test")
		return
	}

	t.Run("ListByService", func(t *testing.T) {
		serviceID := int(services[0].Key)
		serviceSyncs, err := client.VolumeSyncs.ListByService(ctx, serviceID)
		if err != nil {
			t.Fatalf("VolumeSyncs.ListByService failed: %v", err)
		}
		t.Logf("Found %d syncs for service %d", len(serviceSyncs), serviceID)
	})
}

func testVolumeSnapshots(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VolumeSnapshots service...")

	// List all volume snapshots
	snapshots, err := client.VolumeSnapshots.List(ctx)
	if err != nil {
		t.Fatalf("VolumeSnapshots.List failed: %v", err)
	}
	t.Logf("Found %d volume snapshots", len(snapshots))

	if len(snapshots) > 0 {
		prettyPrintWave11(t, "First Volume Snapshot", snapshots[0])

		// Test Get by ID
		first := snapshots[0]
		t.Run("GetByID", func(t *testing.T) {
			snapshot, err := client.VolumeSnapshots.Get(ctx, int(first.Key))
			if err != nil {
				t.Fatalf("VolumeSnapshots.Get failed: %v", err)
			}
			if int(snapshot.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", snapshot.Key, first.Key)
			}
		})
	}

	// Test ListExpiring (within next 30 days)
	t.Run("ListExpiring", func(t *testing.T) {
		expiring, err := client.VolumeSnapshots.ListExpiring(ctx, 30)
		if err != nil {
			t.Fatalf("VolumeSnapshots.ListExpiring failed: %v", err)
		}
		t.Logf("Found %d volume snapshots expiring within 30 days", len(expiring))
	})

	// Test ListManual
	t.Run("ListManual", func(t *testing.T) {
		manual, err := client.VolumeSnapshots.ListManual(ctx)
		if err != nil {
			t.Fatalf("VolumeSnapshots.ListManual failed: %v", err)
		}
		t.Logf("Found %d manually created volume snapshots", len(manual))
	})

	// Get volumes to test ListByVolume
	volumes, err := client.Volumes.List(ctx)
	if err != nil || len(volumes) == 0 {
		t.Log("No volumes available - skipping ListByVolume test")
		return
	}

	// Find a volume with snapshots
	for _, vol := range volumes {
		volID := int(vol.Service) // Use volume service ID to filter
		t.Run("ListByVolume", func(t *testing.T) {
			volSnapshots, err := client.VolumeSnapshots.ListByVolume(ctx, volID)
			if err != nil {
				t.Fatalf("VolumeSnapshots.ListByVolume failed: %v", err)
			}
			t.Logf("Found %d snapshots for volume service %d", len(volSnapshots), volID)
		})
		break // Only test one volume
	}
}

func testPermissions(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Permissions service...")

	// List all permissions
	permissions, err := client.Permissions.List(ctx)
	if err != nil {
		t.Fatalf("Permissions.List failed: %v", err)
	}
	t.Logf("Found %d permissions", len(permissions))

	if len(permissions) > 0 {
		prettyPrintWave11(t, "First Permission", permissions[0])

		// Test Get by ID
		first := permissions[0]
		t.Run("GetByID", func(t *testing.T) {
			perm, err := client.Permissions.Get(ctx, int(first.Key))
			if err != nil {
				t.Fatalf("Permissions.Get failed: %v", err)
			}
			if int(perm.Key) != int(first.Key) {
				t.Errorf("Key mismatch: got %d, want %d", perm.Key, first.Key)
			}
		})
	}

	// Test ListByTable for VMs
	t.Run("ListByTable_VMs", func(t *testing.T) {
		vmPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableVMs)
		if err != nil {
			t.Fatalf("Permissions.ListByTable(vms) failed: %v", err)
		}
		t.Logf("Found %d VM permissions", len(vmPerms))
	})

	// Test ListByTable for networks
	t.Run("ListByTable_Networks", func(t *testing.T) {
		netPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableNetworks)
		if err != nil {
			t.Fatalf("Permissions.ListByTable(vnets) failed: %v", err)
		}
		t.Logf("Found %d network permissions", len(netPerms))
	})

	// Test ListByIdentity if we have permissions with identities
	if len(permissions) > 0 && int(permissions[0].Identity) > 0 {
		t.Run("ListByIdentity", func(t *testing.T) {
			identityID := int(permissions[0].Identity)
			identityPerms, err := client.Permissions.ListByIdentity(ctx, identityID)
			if err != nil {
				t.Fatalf("Permissions.ListByIdentity failed: %v", err)
			}
			t.Logf("Found %d permissions for identity %d", len(identityPerms), identityID)
		})
	}
}

// setupTestClientWave11 creates a client from environment variables
func setupTestClientWave11(t *testing.T) *vergeos.Client {
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

// prettyPrintWave11 logs a struct as formatted JSON for field verification
func prettyPrintWave11(t *testing.T, label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}
