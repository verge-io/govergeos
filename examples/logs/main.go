// Example: System Logs
//
// This example demonstrates how to:
// - Query system logs by level (audit, warning, error)
// - Filter logs by object type (VM, tenant, network)
// - Search logs by text pattern
// - Get recent errors for monitoring
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/logs/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/verge-io/govergeos"
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
	fmt.Println("=== Recent Logs ===")
	showRecentLogs(ctx, client)

	fmt.Println("\n=== Audit Logs ===")
	showAuditLogs(ctx, client)

	fmt.Println("\n=== Errors and Warnings ===")
	showErrorsAndWarnings(ctx, client)

	fmt.Println("\n=== Logs by Object Type ===")
	showLogsByObjectType(ctx, client)

	fmt.Println("\n=== Search Logs ===")
	searchLogs(ctx, client)
}

func showRecentLogs(ctx context.Context, client *vergeos.Client) {
	// Get most recent 20 logs
	logs, err := client.Logs.GetRecent(ctx, 20)
	if err != nil {
		log.Printf("Failed to get recent logs: %v", err)
		return
	}

	fmt.Printf("Last %d log entries:\n", len(logs))
	for _, l := range logs {
		ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
		fmt.Printf("  [%s] %s - %s: %s\n", ts, l.Level, l.ObjectType, truncate(l.Text, 60))
	}

	// Get logs from last 24 hours
	fmt.Println("\n--- Logs in the last 24 hours ---")
	since := time.Now().Add(-24 * time.Hour).UnixMicro()
	recentLogs, err := client.Logs.ListSince(ctx, since, vergeos.WithLimit(100))
	if err != nil {
		log.Printf("Failed to get logs since: %v", err)
		return
	}
	fmt.Printf("Found %d log entries in the last 24 hours\n", len(recentLogs))

	// Count by level
	byLevel := make(map[string]int)
	for _, l := range recentLogs {
		byLevel[l.Level]++
	}
	fmt.Println("By level:")
	for level, count := range byLevel {
		fmt.Printf("  %s: %d\n", level, count)
	}
}

func showAuditLogs(ctx context.Context, client *vergeos.Client) {
	// Audit logs track user actions (logins, resource changes, etc.)
	audits, err := client.Logs.ListAudit(ctx, vergeos.WithLimit(10))
	if err != nil {
		log.Printf("Failed to list audit logs: %v", err)
		return
	}

	fmt.Printf("Recent %d audit entries:\n", len(audits))
	for _, l := range audits {
		ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
		user := l.User
		if user == "" {
			user = "(system)"
		}
		fmt.Printf("  [%s] %s - %s\n", ts, user, truncate(l.Text, 70))
	}

	// Filter by specific user
	fmt.Println("\n--- Logs by user 'admin' ---")
	adminLogs, err := client.Logs.ListByUser(ctx, "admin", vergeos.WithLimit(5))
	if err != nil {
		log.Printf("Failed to list admin logs: %v", err)
		return
	}
	fmt.Printf("Found %d log entries by 'admin'\n", len(adminLogs))
	for _, l := range adminLogs {
		ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
		fmt.Printf("  [%s] %s: %s\n", ts, l.Level, truncate(l.Text, 60))
	}
}

func showErrorsAndWarnings(ctx context.Context, client *vergeos.Client) {
	// Get recent errors (includes error and critical levels)
	errors, err := client.Logs.GetRecentErrors(ctx, 10)
	if err != nil {
		log.Printf("Failed to get recent errors: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("No errors found - system is healthy!")
	} else {
		fmt.Printf("Recent %d error/critical entries:\n", len(errors))
		for _, l := range errors {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s (%s): %s\n", ts, l.Level, l.ObjectType, truncate(l.Text, 50))
		}
	}

	// Get warnings
	fmt.Println("\n--- Recent Warnings ---")
	warnings, err := client.Logs.ListWarnings(ctx, vergeos.WithLimit(10))
	if err != nil {
		log.Printf("Failed to list warnings: %v", err)
		return
	}

	if len(warnings) == 0 {
		fmt.Println("No warnings found")
	} else {
		fmt.Printf("Recent %d warning entries:\n", len(warnings))
		for _, l := range warnings {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s: %s\n", ts, l.ObjectType, truncate(l.Text, 60))
		}
	}
}

func showLogsByObjectType(ctx context.Context, client *vergeos.Client) {
	// VM logs
	vmLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeVM, vergeos.WithLimit(5))
	if err != nil {
		log.Printf("Failed to list VM logs: %v", err)
	} else {
		fmt.Printf("VM logs (%d entries):\n", len(vmLogs))
		for _, l := range vmLogs {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s - %s: %s\n", ts, l.Level, l.ObjectName, truncate(l.Text, 50))
		}
	}

	// Tenant logs
	fmt.Println()
	tenantLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeTenant, vergeos.WithLimit(5))
	if err != nil {
		log.Printf("Failed to list tenant logs: %v", err)
	} else {
		fmt.Printf("Tenant logs (%d entries):\n", len(tenantLogs))
		for _, l := range tenantLogs {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s - %s: %s\n", ts, l.Level, l.ObjectName, truncate(l.Text, 50))
		}
	}

	// User logs (authentication, permission changes)
	fmt.Println()
	userLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeUser, vergeos.WithLimit(5))
	if err != nil {
		log.Printf("Failed to list user logs: %v", err)
	} else {
		fmt.Printf("User logs (%d entries):\n", len(userLogs))
		for _, l := range userLogs {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s: %s\n", ts, l.Level, truncate(l.Text, 60))
		}
	}

	// System logs
	fmt.Println()
	sysLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeSystem, vergeos.WithLimit(5))
	if err != nil {
		log.Printf("Failed to list system logs: %v", err)
	} else {
		fmt.Printf("System logs (%d entries):\n", len(sysLogs))
		for _, l := range sysLogs {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s: %s\n", ts, l.Level, truncate(l.Text, 60))
		}
	}
}

func searchLogs(ctx context.Context, client *vergeos.Client) {
	// Search for login-related logs
	loginLogs, err := client.Logs.Search(ctx, "login", vergeos.WithLimit(10))
	if err != nil {
		log.Printf("Failed to search for 'login': %v", err)
	} else {
		fmt.Printf("Logs containing 'login' (%d entries):\n", len(loginLogs))
		for _, l := range loginLogs {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s: %s\n", ts, l.Level, truncate(l.Text, 60))
		}
	}

	// Search for power-related logs
	fmt.Println()
	powerLogs, err := client.Logs.Search(ctx, "power", vergeos.WithLimit(10))
	if err != nil {
		log.Printf("Failed to search for 'power': %v", err)
	} else {
		fmt.Printf("Logs containing 'power' (%d entries):\n", len(powerLogs))
		for _, l := range powerLogs {
			ts := time.UnixMicro(l.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Printf("  [%s] %s - %s: %s\n", ts, l.Level, l.ObjectType, truncate(l.Text, 50))
		}
	}
}

// truncate truncates a string to the specified length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
