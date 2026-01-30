//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestPermissions tests the Permissions service against a live VergeOS API.
func TestPermissions(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("List", func(t *testing.T) {
		permissions, err := client.Permissions.List(ctx)
		if err != nil {
			t.Fatalf("Permissions.List failed: %v", err)
		}
		t.Logf("Found %d permissions", len(permissions))

		if len(permissions) > 0 {
			prettyPrint(t, "First Permission", permissions[0])
		}
	})

	t.Run("Get", func(t *testing.T) {
		permissions, err := client.Permissions.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(permissions) == 0 {
			t.Skip("No permissions available")
		}

		first := permissions[0]
		perm, err := client.Permissions.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("Permissions.Get failed: %v", err)
		}
		if int(perm.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", perm.Key, first.Key)
		}
	})

	t.Run("ListByTable_VMs", func(t *testing.T) {
		vmPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableVMs)
		if err != nil {
			t.Fatalf("Permissions.ListByTable(vms) failed: %v", err)
		}
		t.Logf("Found %d VM permissions", len(vmPerms))
	})

	t.Run("ListByTable_Networks", func(t *testing.T) {
		netPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableNetworks)
		if err != nil {
			t.Fatalf("Permissions.ListByTable(vnets) failed: %v", err)
		}
		t.Logf("Found %d network permissions", len(netPerms))
	})

	t.Run("ListByIdentity", func(t *testing.T) {
		permissions, err := client.Permissions.List(ctx, vergeos.WithLimit(10))
		if err != nil || len(permissions) == 0 {
			t.Skip("No permissions available")
		}

		// Find a permission with an identity
		var identityID int
		for _, p := range permissions {
			if int(p.Identity) > 0 {
				identityID = int(p.Identity)
				break
			}
		}
		if identityID == 0 {
			t.Skip("No permissions with identity found")
		}

		identityPerms, err := client.Permissions.ListByIdentity(ctx, identityID)
		if err != nil {
			t.Fatalf("Permissions.ListByIdentity failed: %v", err)
		}
		t.Logf("Found %d permissions for identity %d", len(identityPerms), identityID)
	})
}
