// Example: Monitoring - Alarms and Tasks
//
// This example demonstrates how to:
// - List and filter system alarms
// - Query alarm type definitions
// - Snooze and resolve alarms
// - List and manage scheduled tasks
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/monitoring/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/macstadium/govergeos"
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
	fmt.Println("=== Alarm Types ===")
	showAlarmTypes(ctx, client)

	fmt.Println("\n=== Alarms ===")
	showAlarms(ctx, client)

	fmt.Println("\n=== Tasks ===")
	showTasks(ctx, client)
}

func showAlarmTypes(ctx context.Context, client *vergeos.Client) {
	// List all alarm types
	alarmTypes, err := client.AlarmTypes.List(ctx)
	if err != nil {
		log.Printf("Failed to list alarm types: %v", err)
		return
	}

	fmt.Printf("Found %d alarm types\n", len(alarmTypes))

	// Group by level
	byLevel := make(map[string]int)
	for _, at := range alarmTypes {
		byLevel[at.Level]++
	}
	fmt.Println("\nAlarm types by default level:")
	for level, count := range byLevel {
		fmt.Printf("  %s: %d\n", level, count)
	}

	// Show a few example alarm types
	fmt.Println("\nSample alarm types:")
	limit := 5
	if len(alarmTypes) < limit {
		limit = len(alarmTypes)
	}
	for i := 0; i < limit; i++ {
		at := alarmTypes[i]
		fmt.Printf("  - %s (%s): %s\n", at.Key, at.Level, at.Name)
	}

	// Get a specific alarm type by key (if we have any)
	if len(alarmTypes) > 0 {
		key := alarmTypes[0].Key
		alarmType, err := client.AlarmTypes.Get(ctx, key)
		if err != nil {
			log.Printf("Failed to get alarm type %q: %v", key, err)
		} else {
			fmt.Printf("\nAlarm type details for %q:\n", key)
			fmt.Printf("  Name: %s\n", alarmType.Name)
			fmt.Printf("  Level: %s\n", alarmType.Level)
			fmt.Printf("  Description: %s\n", alarmType.Description)
			fmt.Printf("  Default Snooze: %d seconds\n", alarmType.DefaultSnoozeSeconds)
			fmt.Printf("  Max Snooze: %d seconds\n", alarmType.MaxSnoozeSeconds)
		}
	}
}

func showAlarms(ctx context.Context, client *vergeos.Client) {
	// List all alarms
	alarms, err := client.Alarms.List(ctx)
	if err != nil {
		log.Printf("Failed to list alarms: %v", err)
		return
	}

	fmt.Printf("Found %d total alarms\n", len(alarms))

	if len(alarms) == 0 {
		fmt.Println("No alarms - system is healthy!")
		return
	}

	// Group by level
	byLevel := make(map[string]int)
	for _, a := range alarms {
		byLevel[a.Level]++
	}
	fmt.Println("\nAlarms by level:")
	for level, count := range byLevel {
		fmt.Printf("  %s: %d\n", level, count)
	}

	// List active (non-snoozed) alarms
	activeAlarms, err := client.Alarms.ListActive(ctx)
	if err != nil {
		log.Printf("Failed to list active alarms: %v", err)
	} else {
		fmt.Printf("\nActive (non-snoozed) alarms: %d\n", len(activeAlarms))
	}

	// List critical alarms
	criticalAlarms, err := client.Alarms.ListByLevel(ctx, vergeos.AlarmLevelCritical)
	if err != nil {
		log.Printf("Failed to list critical alarms: %v", err)
	} else {
		fmt.Printf("Critical alarms: %d\n", len(criticalAlarms))
		for _, a := range criticalAlarms {
			fmt.Printf("  - [%s] %s: %s\n", a.AlarmID, a.Owner, a.Status)
		}
	}

	// List error alarms
	errorAlarms, err := client.Alarms.ListByLevel(ctx, vergeos.AlarmLevelError)
	if err != nil {
		log.Printf("Failed to list error alarms: %v", err)
	} else {
		fmt.Printf("Error alarms: %d\n", len(errorAlarms))
		for _, a := range errorAlarms {
			fmt.Printf("  - [%s] %s: %s\n", a.AlarmID, a.Owner, a.Status)
		}
	}

	// Show details of first alarm
	if len(alarms) > 0 {
		first := alarms[0]
		fmt.Printf("\nFirst alarm details:\n")
		fmt.Printf("  ID: %d\n", int(first.Key))
		fmt.Printf("  Alarm ID: %s\n", first.AlarmID)
		fmt.Printf("  Owner: %s (%s)\n", first.Owner, first.OwnerType)
		fmt.Printf("  Level: %s\n", first.Level)
		fmt.Printf("  Status: %s\n", first.Status)
		fmt.Printf("  Resolvable: %v\n", first.Resolvable)
		fmt.Printf("  Created: %s\n", time.Unix(first.Created, 0).Format(time.RFC3339))
		if first.Snooze > 0 {
			fmt.Printf("  Snoozed until: %s\n", time.Unix(first.Snooze, 0).Format(time.RFC3339))
		}
	}

	// Example: Snooze an alarm (commented out to avoid modifying data)
	// if len(alarms) > 0 {
	// 	alarmID := int(alarms[0].Key)
	// 	snoozeUntil := time.Now().Add(1 * time.Hour).Unix()
	// 	if err := client.Alarms.Snooze(ctx, alarmID, snoozeUntil); err != nil {
	// 		log.Printf("Failed to snooze alarm: %v", err)
	// 	} else {
	// 		fmt.Printf("Snoozed alarm %d until %s\n", alarmID, time.Unix(snoozeUntil, 0))
	// 	}
	// }
}

func showTasks(ctx context.Context, client *vergeos.Client) {
	// List all tasks
	tasks, err := client.Tasks.List(ctx)
	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
		return
	}

	fmt.Printf("Found %d tasks\n", len(tasks))

	if len(tasks) == 0 {
		fmt.Println("No scheduled tasks configured")
		return
	}

	// Group by status
	byStatus := make(map[string]int)
	enabledCount := 0
	for _, t := range tasks {
		byStatus[t.Status]++
		if t.Enabled {
			enabledCount++
		}
	}
	fmt.Printf("\nEnabled tasks: %d\n", enabledCount)
	fmt.Println("Tasks by status:")
	for status, count := range byStatus {
		fmt.Printf("  %s: %d\n", status, count)
	}

	// List running tasks
	runningTasks, err := client.Tasks.ListRunning(ctx)
	if err != nil {
		log.Printf("Failed to list running tasks: %v", err)
	} else {
		fmt.Printf("\nCurrently running: %d\n", len(runningTasks))
		for _, t := range runningTasks {
			fmt.Printf("  - %s (owner: %s, action: %s)\n", t.Name, t.Owner, t.Action)
		}
	}

	// Show sample tasks
	fmt.Println("\nSample tasks:")
	limit := 10
	if len(tasks) < limit {
		limit = len(tasks)
	}
	for i := 0; i < limit; i++ {
		t := tasks[i]
		status := "disabled"
		if t.Enabled {
			status = "enabled"
		}
		fmt.Printf("  - %s [%s] (owner: %s, action: %s)\n", t.Name, status, t.Owner, t.Action)
	}

	// Show details of first task
	if len(tasks) > 0 {
		first := tasks[0]
		fmt.Printf("\nFirst task details:\n")
		fmt.Printf("  Key: %d\n", int(first.Key))
		fmt.Printf("  ID (SHA1): %s\n", first.ID)
		fmt.Printf("  Name: %s\n", first.Name)
		fmt.Printf("  Owner: %s\n", first.Owner)
		fmt.Printf("  Table: %s\n", first.Table)
		fmt.Printf("  Action: %s\n", first.Action)
		fmt.Printf("  Enabled: %v\n", first.Enabled)
		fmt.Printf("  Status: %s\n", first.Status)
		fmt.Printf("  Last Run: %s\n", first.LastRun)
		fmt.Printf("  Delete After Run: %v\n", first.DeleteAfterRun)

		// Get by SHA1 ID
		if first.ID != "" {
			byID, err := client.Tasks.GetByID(ctx, first.ID)
			if err != nil {
				log.Printf("Failed to get task by ID: %v", err)
			} else {
				fmt.Printf("\nRetrieved task by SHA1 ID: %s (Key: %d)\n", byID.ID, int(byID.Key))
			}
		}
	}

	// Example: Execute a task (commented out to avoid side effects)
	// if len(tasks) > 0 && tasks[0].Enabled {
	// 	taskID := int(tasks[0].Key)
	// 	if err := client.Tasks.Execute(ctx, taskID, nil); err != nil {
	// 		log.Printf("Failed to execute task: %v", err)
	// 	} else {
	// 		fmt.Printf("Executed task %d\n", taskID)
	// 	}
	// }
}
