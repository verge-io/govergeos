//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestWebhookURLsList tests listing webhook URL configurations.
func TestWebhookURLsList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Testing WebhookURLs service...")

	webhooks, err := client.WebhookURLs.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list webhook URLs: %v", err)
	}

	t.Logf("Found %d webhook URL(s)", len(webhooks))

	if len(webhooks) == 0 {
		t.Log("No webhook URLs found - this is normal if webhooks are not configured")
		return
	}

	// Log first webhook URL to verify field mapping
	first := webhooks[0]
	t.Logf("First webhook URL: Key=%d, Name=%q, URL=%q, AuthType=%q",
		int(first.Key), first.Name, first.URL, first.AuthorizationType)

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.WebhookURLs.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("WebhookURLs.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("WebhookURLs.Get succeeded: Name=%q, Timeout=%d, Retries=%d",
				fetched.Name, fetched.Timeout, fetched.Retries)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" {
			t.Skip("No webhook URL name available")
		}
		byName, err := client.WebhookURLs.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("WebhookURLs.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	prettyPrint(t, "Sample WebhookURL", first)
}

// TestWebhooksList tests listing webhook message logs.
func TestWebhooksList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Testing Webhooks service...")

	webhooks, err := client.Webhooks.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}

	t.Logf("Found %d webhook message(s)", len(webhooks))

	if len(webhooks) == 0 {
		t.Log("No webhook messages found - this is normal if no webhooks have been triggered")
		return
	}

	// Log first webhook to verify field mapping
	first := webhooks[0]
	t.Logf("First webhook: Key=%d, Status=%q, WebhookURL=%d",
		int(first.Key), first.Status, int(first.WebhookURL))

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.Webhooks.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Webhooks.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Webhooks.Get succeeded: Status=%q", fetched.Status)
		}
	})

	prettyPrint(t, "Sample Webhook", first)
}

// TestWebhooksByStatus tests listing webhooks by status.
func TestWebhooksByStatus(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test listing sent webhooks
	t.Run("ListByStatus_Sent", func(t *testing.T) {
		sent, err := client.Webhooks.ListByStatus(ctx, vergeos.WebhookStatusSent)
		if err != nil {
			t.Fatalf("Failed to list sent webhooks: %v", err)
		}
		t.Logf("Found %d sent webhook(s)", len(sent))
	})

	// Test listing failed webhooks
	t.Run("ListFailed", func(t *testing.T) {
		failed, err := client.Webhooks.ListFailed(ctx)
		if err != nil {
			t.Fatalf("Failed to list failed webhooks: %v", err)
		}
		t.Logf("Found %d failed webhook(s)", len(failed))
	})

	// Test listing pending webhooks
	t.Run("ListPending", func(t *testing.T) {
		pending, err := client.Webhooks.ListPending(ctx)
		if err != nil {
			t.Fatalf("Failed to list pending webhooks: %v", err)
		}
		t.Logf("Found %d pending webhook(s)", len(pending))
	})
}

// TestWebhookURLsCRUD tests full CRUD operations on webhook URLs.
func TestWebhookURLsCRUD(t *testing.T) {
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

	// Read
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
