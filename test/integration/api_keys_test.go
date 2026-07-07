//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/macstadium/govergeos"
)

// TestUserAPIKeysList tests listing user API keys.
func TestUserAPIKeysList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Testing UserAPIKeys service...")

	keys, err := client.UserAPIKeys.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list user API keys: %v", err)
	}

	t.Logf("Found %d user API key(s)", len(keys))

	if len(keys) == 0 {
		t.Log("No API keys found - this is normal if no keys have been created")
		return
	}

	// Log first key to verify field mapping
	first := keys[0]
	expires := "never"
	if first.Expires > 0 {
		expires = time.Unix(first.Expires, 0).Format("2006-01-02")
	}
	t.Logf("First API key: Key=%d, Name=%q, User=%q, Expires=%s",
		int(first.Key), first.Name, first.UserName, expires)

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.UserAPIKeys.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("UserAPIKeys.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("UserAPIKeys.Get succeeded: Name=%q, User=%d", fetched.Name, int(fetched.User))
		}
	})

	prettyPrint(t, "Sample UserAPIKey", first)
}

// TestUserAPIKeysByUser tests listing API keys for a specific user.
func TestUserAPIKeysByUser(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First get a user to query
	users, err := client.Users.List(ctx, vergeos.WithLimit(1))
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	if len(users) == 0 {
		t.Skip("No users found to test API key listing")
	}

	userID := int(users[0].Key)
	t.Logf("Testing API keys for user %s (ID: %d)", users[0].Name, userID)

	keys, err := client.UserAPIKeys.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to list API keys for user %d: %v", userID, err)
	}

	t.Logf("Found %d API key(s) for user %s", len(keys), users[0].Name)
}
