// Example: Webhooks - Event Notification Management
//
// This example demonstrates how to:
// - List and manage webhook endpoint configurations
// - View webhook delivery logs and status
// - Query pending and failed webhook deliveries
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/webhooks/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/verge-io/goVergeOS"
)

func main() {
	// Create client from environment variables
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(os.Getenv("VERGEOS_HOST")),
		vergeos.WithCredentials(os.Getenv("VERGEOS_USERNAME"), os.Getenv("VERGEOS_PASSWORD")),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Run examples
	fmt.Println("=== Webhook Endpoints ===")
	showWebhookURLs(ctx, client)

	fmt.Println("\n=== Webhook Delivery Log ===")
	showWebhookDeliveries(ctx, client)
}

func showWebhookURLs(ctx context.Context, client *vergeos.Client) {
	// List all webhook URL configurations
	webhooks, err := client.WebhookURLs.List(ctx)
	if err != nil {
		log.Printf("Failed to list webhook URLs: %v", err)
		return
	}

	fmt.Printf("Found %d webhook endpoints configured\n", len(webhooks))

	if len(webhooks) == 0 {
		fmt.Println("No webhook endpoints configured")
		fmt.Println("\nTo receive event notifications, create a webhook endpoint:")
		fmt.Println("  - Slack: Use incoming webhook URL")
		fmt.Println("  - PagerDuty: Use events API endpoint")
		fmt.Println("  - Custom: Any HTTP/HTTPS endpoint that accepts JSON")
		return
	}

	// Group by authorization type
	byAuthType := make(map[string]int)
	for _, w := range webhooks {
		authType := w.AuthorizationType
		if authType == "" {
			authType = "none"
		}
		byAuthType[authType]++
	}
	fmt.Println("\nBy authorization type:")
	for authType, count := range byAuthType {
		fmt.Printf("  %s: %d\n", authType, count)
	}

	// Show webhook details
	fmt.Println("\nWebhook endpoints:")
	for _, webhook := range webhooks {
		fmt.Printf("\n  Webhook: %s (ID: %d)\n", webhook.Name, int(webhook.Key))
		fmt.Printf("    URL: %s\n", webhook.URL)
		fmt.Printf("    Type: %s\n", webhook.Type)

		// Auth info (don't show actual credentials)
		authType := webhook.AuthorizationType
		if authType == "" {
			authType = "none"
		}
		fmt.Printf("    Authorization: %s\n", authType)

		// Connection settings
		fmt.Printf("    Timeout: %d seconds\n", webhook.Timeout)
		fmt.Printf("    Retries: %d\n", webhook.Retries)
		fmt.Printf("    Allow Insecure: %v\n", webhook.AllowInsecure)

		// Custom headers (if any)
		if webhook.Headers != "" {
			fmt.Printf("    Custom Headers: (configured)\n")
		}
	}

	// Get a specific webhook by name (if we have any)
	if len(webhooks) > 0 {
		webhook, err := client.WebhookURLs.GetByName(ctx, webhooks[0].Name)
		if err != nil {
			log.Printf("Failed to get webhook by name: %v", err)
		} else {
			fmt.Printf("\nRetrieved webhook by name: %s (Key: %d)\n", webhook.Name, int(webhook.Key))
		}
	}

	// Example: Create a webhook endpoint (commented out to avoid side effects)
	// webhook, err := client.WebhookURLs.Create(ctx, &vergeos.WebhookURLCreateRequest{
	// 	Name:              "slack-alerts",
	// 	URL:               "https://hooks.slack.com/services/xxx/yyy/zzz",
	// 	AuthorizationType: "none",
	// 	Timeout:           ptr(10),
	// 	Retries:           ptr(3),
	// })
	// if err != nil {
	// 	log.Printf("Failed to create webhook: %v", err)
	// } else {
	// 	fmt.Printf("Created webhook: %s (ID: %d)\n", webhook.Name, int(webhook.Key))
	// }

	// Example: Send a test message (commented out to avoid side effects)
	// if len(webhooks) > 0 {
	// 	webhookID := int(webhooks[0].Key)
	// 	testMessage := `{"text": "Test message from goVergeOS"}`
	// 	if err := client.WebhookURLs.Send(ctx, webhookID, testMessage); err != nil {
	// 		log.Printf("Failed to send test message: %v", err)
	// 	} else {
	// 		fmt.Printf("Sent test message to webhook %d\n", webhookID)
	// 	}
	// }
}

func showWebhookDeliveries(ctx context.Context, client *vergeos.Client) {
	// List all webhook deliveries (the message log)
	deliveries, err := client.Webhooks.List(ctx)
	if err != nil {
		log.Printf("Failed to list webhook deliveries: %v", err)
		return
	}

	fmt.Printf("Found %d webhook deliveries in log\n", len(deliveries))

	if len(deliveries) == 0 {
		fmt.Println("No webhook deliveries recorded")
		return
	}

	// Group by status
	byStatus := make(map[string]int)
	for _, d := range deliveries {
		byStatus[d.Status]++
	}
	fmt.Println("\nBy status:")
	for status, count := range byStatus {
		fmt.Printf("  %s: %d\n", status, count)
	}

	// Check for pending deliveries
	pending, err := client.Webhooks.ListPending(ctx)
	if err != nil {
		log.Printf("Failed to list pending deliveries: %v", err)
	} else {
		fmt.Printf("\nPending deliveries: %d\n", len(pending))
		if len(pending) > 0 {
			fmt.Println("Pending messages waiting to be sent:")
			for _, p := range pending {
				fmt.Printf("  - ID: %d, Webhook: %d, Created: %s\n",
					int(p.Key), int(p.WebhookURL),
					time.Unix(p.Created, 0).Format(time.RFC3339))
			}
		}
	}

	// Check for failed deliveries
	failed, err := client.Webhooks.ListFailed(ctx)
	if err != nil {
		log.Printf("Failed to list failed deliveries: %v", err)
	} else {
		fmt.Printf("\nFailed deliveries: %d\n", len(failed))
		if len(failed) > 0 {
			fmt.Println("Failed messages (may need attention):")
			limit := 5
			if len(failed) < limit {
				limit = len(failed)
			}
			for i := 0; i < limit; i++ {
				f := failed[i]
				fmt.Printf("  - ID: %d, Webhook: %d\n",
					int(f.Key), int(f.WebhookURL))
				if f.StatusInfo != "" {
					fmt.Printf("      Error: %s\n", f.StatusInfo)
				}
				if f.LastAttempt > 0 {
					fmt.Printf("      Last Attempt: %s\n", time.Unix(f.LastAttempt, 0).Format(time.RFC3339))
				}
			}
			if len(failed) > limit {
				fmt.Printf("  ... and %d more failed deliveries\n", len(failed)-limit)
			}
		}
	}

	// Show recent deliveries
	fmt.Println("\nRecent webhook deliveries:")
	limit := 10
	if len(deliveries) < limit {
		limit = len(deliveries)
	}
	for i := 0; i < limit; i++ {
		d := deliveries[i]
		fmt.Printf("  - [%s] ID: %d, Webhook: %d\n",
			d.Status, int(d.Key), int(d.WebhookURL))
		if d.StatusInfo != "" && d.Status != "sent" {
			fmt.Printf("      Info: %s\n", d.StatusInfo)
		}
		if d.LastAttempt > 0 {
			fmt.Printf("      Last Attempt: %s\n", time.Unix(d.LastAttempt, 0).Format(time.RFC3339))
		}
		fmt.Printf("      Created: %s\n", time.Unix(d.Created, 0).Format(time.RFC3339))
	}

	// Show a specific delivery's details
	if len(deliveries) > 0 {
		delivery, err := client.Webhooks.Get(ctx, int(deliveries[0].Key))
		if err != nil {
			log.Printf("Failed to get delivery details: %v", err)
		} else {
			fmt.Printf("\nDelivery details (ID: %d):\n", int(delivery.Key))
			fmt.Printf("  Webhook URL ID: %d\n", int(delivery.WebhookURL))
			fmt.Printf("  Status: %s\n", delivery.Status)
			if delivery.StatusInfo != "" {
				fmt.Printf("  Status Info: %s\n", delivery.StatusInfo)
			}
			if delivery.Message != "" {
				// Truncate long messages
				msg := delivery.Message
				if len(msg) > 200 {
					msg = msg[:200] + "..."
				}
				fmt.Printf("  Message: %s\n", msg)
			}
		}
	}
}
