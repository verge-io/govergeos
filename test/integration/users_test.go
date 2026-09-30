//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestUserEnableDisable creates a disposable user and toggles enabled.
// Enabling and disabling is a PUT of the enabled field, not a user action.
func TestUserEnableDisable(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	testName := "sdk-test-user-" + time.Now().Format("20060102150405")
	created, err := client.Users.Create(ctx, &vergeos.UserCreateRequest{
		Name:        testName,
		DisplayName: "SDK test user",
		Password:    "Sdk-Test-Password-1",
		Enabled:     ptr(true),
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID := int(created.Key)
	t.Logf("Created user %d (%s)", userID, testName)

	defer func() {
		if err := client.Users.Delete(ctx, userID); err != nil {
			t.Errorf("Failed to delete user %d: %v", userID, err)
		}
	}()

	if err := client.Users.Disable(ctx, userID); err != nil {
		t.Fatalf("Users.Disable failed: %v", err)
	}
	disabled, err := client.Users.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Users.Get after disable failed: %v", err)
	}
	if disabled.Enabled {
		t.Fatal("expected user to be disabled")
	}

	if err := client.Users.Enable(ctx, userID); err != nil {
		t.Fatalf("Users.Enable failed: %v", err)
	}
	enabled, err := client.Users.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Users.Get after enable failed: %v", err)
	}
	if !enabled.Enabled {
		t.Fatal("expected user to be enabled")
	}
}
