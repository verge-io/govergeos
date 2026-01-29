//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestWave9WebhookURLsList tests listing webhook URL configurations.
func TestWave9WebhookURLsList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	webhooks, err := client.WebhookURLs.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list webhook URLs: %v", err)
	}

	t.Logf("Found %d webhook URL(s)", len(webhooks))
	for i, wh := range webhooks {
		if i >= 3 {
			t.Logf("  ... and %d more", len(webhooks)-3)
			break
		}
		t.Logf("  - %s (ID: %d, URL: %s, Auth: %s)",
			wh.Name, wh.Key, wh.URL, wh.AuthorizationType)
	}
}

// TestWave9WebhooksList tests listing webhook message logs.
func TestWave9WebhooksList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	webhooks, err := client.Webhooks.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}

	t.Logf("Found %d webhook message(s)", len(webhooks))
	for i, wh := range webhooks {
		if i >= 3 {
			t.Logf("  ... and %d more", len(webhooks)-3)
			break
		}
		t.Logf("  - ID: %d, Status: %s, WebhookURL: %d",
			wh.Key, wh.Status, wh.WebhookURL)
	}
}

// TestWave9WebhooksByStatus tests listing webhooks by status.
func TestWave9WebhooksByStatus(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test listing sent webhooks
	sent, err := client.Webhooks.ListByStatus(ctx, vergeos.WebhookStatusSent)
	if err != nil {
		t.Fatalf("Failed to list sent webhooks: %v", err)
	}
	t.Logf("Found %d sent webhook(s)", len(sent))

	// Test listing failed webhooks
	failed, err := client.Webhooks.ListFailed(ctx)
	if err != nil {
		t.Fatalf("Failed to list failed webhooks: %v", err)
	}
	t.Logf("Found %d failed webhook(s)", len(failed))

	// Test listing pending webhooks
	pending, err := client.Webhooks.ListPending(ctx)
	if err != nil {
		t.Fatalf("Failed to list pending webhooks: %v", err)
	}
	t.Logf("Found %d pending webhook(s)", len(pending))
}

// TestWave9UserAPIKeysList tests listing user API keys.
func TestWave9UserAPIKeysList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys, err := client.UserAPIKeys.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list user API keys: %v", err)
	}

	t.Logf("Found %d user API key(s)", len(keys))
	for i, key := range keys {
		if i >= 3 {
			t.Logf("  ... and %d more", len(keys)-3)
			break
		}
		expires := "never"
		if key.Expires > 0 {
			expires = time.Unix(key.Expires, 0).Format("2006-01-02")
		}
		t.Logf("  - %s (ID: %d, User: %s, Expires: %s)",
			key.Name, key.Key, key.UserName, expires)
	}
}

// TestWave9UserAPIKeysByUser tests listing API keys for a specific user.
func TestWave9UserAPIKeysByUser(t *testing.T) {
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

// TestWave9WebhookURLCRUD tests full CRUD operations on webhook URLs.
// This test creates a webhook URL, updates it, and deletes it.
func TestWave9WebhookURLCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	testName := "sdk-test-webhook-" + time.Now().Format("20060102150405")
	testURL := "https://httpbin.org/post"

	// Create
	t.Logf("Creating webhook URL: %s", testName)
	created, err := client.WebhookURLs.Create(ctx, &vergeos.WebhookURLCreateRequest{
		Name:              testName,
		URL:               testURL,
		AuthorizationType: vergeos.WebhookAuthNone,
		Timeout:           ptr(10),
		Retries:           ptr(2),
	})
	if err != nil {
		t.Fatalf("Failed to create webhook URL: %v", err)
	}
	webhookID := int(created.Key)
	t.Logf("Created webhook URL with ID: %d", webhookID)

	// Cleanup
	defer func() {
		t.Logf("Cleaning up webhook URL: %d", webhookID)
		if err := client.WebhookURLs.Delete(ctx, webhookID); err != nil {
			t.Errorf("Failed to delete webhook URL: %v", err)
		}
	}()

	// Verify creation
	if created.Name != testName {
		t.Errorf("Expected name %q, got %q", testName, created.Name)
	}
	if created.URL != testURL {
		t.Errorf("Expected URL %q, got %q", testURL, created.URL)
	}

	// Get
	t.Logf("Getting webhook URL by ID: %d", webhookID)
	got, err := client.WebhookURLs.Get(ctx, webhookID)
	if err != nil {
		t.Fatalf("Failed to get webhook URL: %v", err)
	}
	if got.Name != testName {
		t.Errorf("Expected name %q, got %q", testName, got.Name)
	}

	// GetByName
	t.Logf("Getting webhook URL by name: %s", testName)
	byName, err := client.WebhookURLs.GetByName(ctx, testName)
	if err != nil {
		t.Fatalf("Failed to get webhook URL by name: %v", err)
	}
	if int(byName.Key) != webhookID {
		t.Errorf("Expected ID %d, got %d", webhookID, byName.Key)
	}

	// Update
	newTimeout := 15
	t.Logf("Updating webhook URL timeout to: %d", newTimeout)
	updated, err := client.WebhookURLs.Update(ctx, webhookID, &vergeos.WebhookURLUpdateRequest{
		Timeout: &newTimeout,
	})
	if err != nil {
		t.Fatalf("Failed to update webhook URL: %v", err)
	}
	if updated.Timeout != newTimeout {
		t.Errorf("Expected timeout %d, got %d", newTimeout, updated.Timeout)
	}

	t.Log("Webhook URL CRUD test completed successfully")
}
