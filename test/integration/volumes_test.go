//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/macstadium/govergeos"
)

// TestVolumesList tests the Volumes service list and get operations.
func TestVolumesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Volumes service...")

	// List all volumes
	volumes, err := client.Volumes.List(ctx)
	if err != nil {
		t.Fatalf("Volumes.List failed: %v", err)
	}

	t.Logf("Found %d volumes", len(volumes))

	if len(volumes) == 0 {
		t.Log("No volumes found - skipping Get tests (requires NAS service)")
		return
	}

	// Log first volume to verify field mapping
	first := volumes[0]
	t.Logf("First volume: Key=%q, Name=%q, Enabled=%v, FSType=%q, Service=%d",
		first.Key, first.Name, first.Enabled, first.FSType, int(first.Service))

	// Verify Key is a non-empty string (volumes use SHA1 hash string keys)
	if first.Key == "" {
		t.Error("Volume.Key is empty - expected SHA1 hash string")
	}

	// Test Get by string key
	t.Run("Get", func(t *testing.T) {
		if first.Key == "" {
			t.Skip("No volume key available")
		}
		fetched, err := client.Volumes.Get(ctx, first.Key)
		if err != nil {
			t.Errorf("Volumes.Get(%q) failed: %v", first.Key, err)
		} else {
			t.Logf("Volumes.Get succeeded: Key=%q, Name=%q, MaxSize=%d", fetched.Key, fetched.Name, fetched.MaxSize)
		}
	})

	// Test ListByService if we have a service ID
	t.Run("ListByService", func(t *testing.T) {
		if first.Service == 0 {
			t.Skip("No service ID available")
		}
		serviceVolumes, err := client.Volumes.ListByService(ctx, int(first.Service))
		if err != nil {
			t.Errorf("Volumes.ListByService failed: %v", err)
		} else {
			t.Logf("Found %d volumes in service %d", len(serviceVolumes), int(first.Service))
		}
	})

	// Test GetByName within service
	t.Run("GetByName", func(t *testing.T) {
		if first.Service == 0 || first.Name == "" {
			t.Skip("No service ID or name available")
		}
		byName, err := client.Volumes.GetByName(ctx, int(first.Service), first.Name)
		if err != nil {
			t.Errorf("Volumes.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%q", byName.Key)
		}
	})

	// Pretty print first volume for field verification
	prettyPrint(t, "Sample Volume", first)
}

// TestVolumesCRUD tests Create/Update/Delete operations for Volumes.
// Requires a NAS service to be available.
func TestVolumesCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Find a NAS service to create volumes in
	services, err := client.NASServices.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list NAS services: %v", err)
	}
	if len(services) == 0 {
		t.Skip("No NAS service available for volume CRUD test")
	}

	serviceID := int(services[0].Key)
	t.Logf("Using NAS service %d for CRUD test", serviceID)

	// Create
	maxSize := int64(1073741824) // 1GB
	volume, err := client.Volumes.Create(ctx, &vergeos.VolumeCreateRequest{
		Service:     serviceID,
		Name:        "sdk-test-volume",
		Description: "goVergeOS integration test volume - safe to delete",
		MaxSize:     &maxSize,
	})
	if err != nil {
		t.Fatalf("Volumes.Create failed: %v", err)
	}
	volumeKey := volume.Key
	t.Logf("Created volume: Key=%q, Name=%q", volumeKey, volume.Name)

	// Cleanup
	defer func() {
		t.Log("Cleaning up: deleting test volume...")
		if err := client.Volumes.Delete(ctx, volumeKey); err != nil {
			t.Logf("Warning: failed to delete test volume: %v", err)
		} else {
			t.Log("Test volume deleted successfully")
		}
	}()

	// Read
	volume, err = client.Volumes.Get(ctx, volumeKey)
	if err != nil {
		t.Fatalf("Volumes.Get failed: %v", err)
	}
	t.Logf("Read volume: Key=%q, Name=%q, MaxSize=%d", volumeKey, volume.Name, volume.MaxSize)

	// Update
	newDesc := "Updated goVergeOS test volume"
	volume, err = client.Volumes.Update(ctx, volumeKey, &vergeos.VolumeUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Volumes.Update failed: %v", err)
	}
	if volume.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, volume.Description)
	} else {
		t.Logf("Updated volume description: %q", volume.Description)
	}

	// Test Disable/Enable
	t.Run("Disable", func(t *testing.T) {
		err := client.Volumes.Disable(ctx, volumeKey)
		if err != nil {
			t.Errorf("Volumes.Disable failed: %v", err)
			return
		}
		vol, _ := client.Volumes.Get(ctx, volumeKey)
		if vol != nil && vol.Enabled {
			t.Error("Volume should be disabled")
		} else {
			t.Log("Volume disabled successfully")
		}
	})

	t.Run("Enable", func(t *testing.T) {
		err := client.Volumes.Enable(ctx, volumeKey)
		if err != nil {
			t.Errorf("Volumes.Enable failed: %v", err)
			return
		}
		vol, _ := client.Volumes.Get(ctx, volumeKey)
		if vol != nil && !vol.Enabled {
			t.Error("Volume should be enabled")
		} else {
			t.Log("Volume enabled successfully")
		}
	})

	t.Log("Volume CRUD test completed")
}
