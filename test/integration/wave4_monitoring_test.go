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

// TestWave4Monitoring tests the Wave 4 monitoring services (Alarms, AlarmTypes, Tasks)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave4
func TestWave4Monitoring(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("AlarmTypes", func(t *testing.T) {
		testAlarmTypes(t, ctx, client)
	})

	t.Run("Alarms", func(t *testing.T) {
		testAlarms(t, ctx, client)
	})

	t.Run("Tasks", func(t *testing.T) {
		testTasks(t, ctx, client)
	})
}

func testAlarmTypes(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing AlarmTypes service...")

	// List all alarm types
	alarmTypes, err := client.AlarmTypes.List(ctx)
	if err != nil {
		t.Fatalf("AlarmTypes.List failed: %v", err)
	}

	t.Logf("Found %d alarm types", len(alarmTypes))

	if len(alarmTypes) == 0 {
		t.Log("No alarm types found - this is unusual, system should have default types")
		return
	}

	// Log first alarm type to verify field mapping
	first := alarmTypes[0]
	t.Logf("First alarm type: Key=%q, Name=%q, Level=%q, DefaultSnoozeSeconds=%d",
		first.Key, first.Name, first.Level, first.DefaultSnoozeSeconds)

	// Verify Key is a non-empty string (alarm types use string keys)
	if first.Key == "" {
		t.Error("AlarmType.Key is empty - expected string key like 'vm_cpu_high'")
	}

	// Test Get by string key
	if first.Key != "" {
		fetched, err := client.AlarmTypes.Get(ctx, first.Key)
		if err != nil {
			t.Errorf("AlarmTypes.Get(%q) failed: %v", first.Key, err)
		} else {
			t.Logf("AlarmTypes.Get succeeded: Key=%q, Description=%q", fetched.Key, fetched.Description)
		}
	}

	// Test ListByLevel
	warningTypes, err := client.AlarmTypes.ListByLevel(ctx, vergeos.AlarmLevelWarning)
	if err != nil {
		t.Errorf("AlarmTypes.ListByLevel failed: %v", err)
	} else {
		t.Logf("Found %d alarm types with level 'warning'", len(warningTypes))
	}

	// Pretty print first alarm type for field verification
	prettyPrint(t, "Sample AlarmType", first)
}

func testAlarms(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Alarms service...")

	// List all alarms
	alarms, err := client.Alarms.List(ctx)
	if err != nil {
		t.Fatalf("Alarms.List failed: %v", err)
	}

	t.Logf("Found %d alarms", len(alarms))

	if len(alarms) == 0 {
		t.Log("No alarms found - system is healthy or no alerts configured")
		// Still test ListActive to ensure it works
		activeAlarms, err := client.Alarms.ListActive(ctx)
		if err != nil {
			t.Errorf("Alarms.ListActive failed: %v", err)
		} else {
			t.Logf("ListActive returned %d active alarms", len(activeAlarms))
		}
		return
	}

	// Log first alarm to verify field mapping
	first := alarms[0]
	t.Logf("First alarm: Key=%d, Owner=%q, OwnerType=%q, Level=%q, Status=%q",
		int(first.Key), first.Owner, first.OwnerType, first.Level, first.Status)

	// Verify Owner is a string path (e.g., "vms/123")
	if first.Owner == "" {
		t.Log("Warning: Alarm.Owner is empty")
	} else {
		t.Logf("Alarm.Owner format verified: %q", first.Owner)
	}

	// Test Get by ID
	fetched, err := client.Alarms.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("Alarms.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("Alarms.Get succeeded: AlarmID=%q, Resolvable=%v", fetched.AlarmID, fetched.Resolvable)
	}

	// Test ListActive
	activeAlarms, err := client.Alarms.ListActive(ctx)
	if err != nil {
		t.Errorf("Alarms.ListActive failed: %v", err)
	} else {
		t.Logf("Found %d active (non-snoozed) alarms", len(activeAlarms))
	}

	// Test ListByLevel
	errorAlarms, err := client.Alarms.ListByLevel(ctx, vergeos.AlarmLevelError)
	if err != nil {
		t.Errorf("Alarms.ListByLevel failed: %v", err)
	} else {
		t.Logf("Found %d alarms with level 'error'", len(errorAlarms))
	}

	// Test ListByOwner if we have an owner
	if first.Owner != "" {
		ownerAlarms, err := client.Alarms.ListByOwner(ctx, first.Owner)
		if err != nil {
			t.Errorf("Alarms.ListByOwner failed: %v", err)
		} else {
			t.Logf("Found %d alarms for owner %q", len(ownerAlarms), first.Owner)
		}
	}

	// Pretty print first alarm for field verification
	prettyPrint(t, "Sample Alarm", first)
}

func testTasks(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Tasks service...")

	// List all tasks
	tasks, err := client.Tasks.List(ctx)
	if err != nil {
		t.Fatalf("Tasks.List failed: %v", err)
	}

	t.Logf("Found %d tasks", len(tasks))

	if len(tasks) == 0 {
		t.Log("No tasks found - no scheduled tasks configured")
		// Still test ListRunning to ensure it works
		runningTasks, err := client.Tasks.ListRunning(ctx)
		if err != nil {
			t.Errorf("Tasks.ListRunning failed: %v", err)
		} else {
			t.Logf("ListRunning returned %d running tasks", len(runningTasks))
		}
		return
	}

	// Log first task to verify field mapping
	first := tasks[0]
	t.Logf("First task: Key=%d, ID=%q, Owner=%q, Action=%q, Name=%q, Status=%q",
		int(first.Key), first.ID, first.Owner, first.Action, first.Name, first.Status)

	// Verify ID is a 40-character SHA1 hash
	if len(first.ID) != 40 {
		t.Logf("Warning: Task.ID length is %d, expected 40 (SHA1 hash)", len(first.ID))
	}

	// Verify Owner is a string path
	if first.Owner == "" {
		t.Log("Warning: Task.Owner is empty")
	}

	// Test Get by Key
	fetched, err := client.Tasks.Get(ctx, int(first.Key))
	if err != nil {
		t.Errorf("Tasks.Get(%d) failed: %v", int(first.Key), err)
	} else {
		t.Logf("Tasks.Get succeeded: Name=%q, Enabled=%v, Status=%q",
			fetched.Name, fetched.Enabled, fetched.Status)
	}

	// Test GetByID (SHA1 hash)
	if first.ID != "" {
		byID, err := client.Tasks.GetByID(ctx, first.ID)
		if err != nil {
			t.Errorf("Tasks.GetByID(%q) failed: %v", first.ID, err)
		} else {
			t.Logf("Tasks.GetByID succeeded: Key=%d", int(byID.Key))
		}
	}

	// Test ListRunning
	runningTasks, err := client.Tasks.ListRunning(ctx)
	if err != nil {
		t.Errorf("Tasks.ListRunning failed: %v", err)
	} else {
		t.Logf("Found %d running tasks", len(runningTasks))
	}

	// Test ListEnabled
	enabledTasks, err := client.Tasks.ListEnabled(ctx)
	if err != nil {
		t.Errorf("Tasks.ListEnabled failed: %v", err)
	} else {
		t.Logf("Found %d enabled tasks", len(enabledTasks))
	}

	// Test ListByOwner if we have an owner
	if first.Owner != "" {
		ownerTasks, err := client.Tasks.ListByOwner(ctx, first.Owner)
		if err != nil {
			t.Errorf("Tasks.ListByOwner failed: %v", err)
		} else {
			t.Logf("Found %d tasks for owner %q", len(ownerTasks), first.Owner)
		}
	}

	// Pretty print first task for field verification
	prettyPrint(t, "Sample Task", first)
}

// setupTestClient creates a client from environment variables
func setupTestClient(t *testing.T) *vergeos.Client {
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

// prettyPrint logs a struct as formatted JSON for field verification
func prettyPrint(t *testing.T, label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}

// TestAlarmSnoozeWorkflow tests snoozing and unsnoozing an alarm
// This is a separate test because it modifies data
func TestAlarmSnoozeWorkflow(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Find a non-critical alarm to test with
	alarms, err := client.Alarms.List(ctx, vergeos.WithFilter("level ne 'critical'"))
	if err != nil {
		t.Fatalf("Failed to list alarms: %v", err)
	}

	if len(alarms) == 0 {
		t.Skip("No non-critical alarms available for snooze testing")
	}

	alarm := alarms[0]
	alarmID := int(alarm.Key)
	t.Logf("Testing snooze workflow on alarm %d (Level: %s, Status: %s)",
		alarmID, alarm.Level, alarm.Status)

	// Snooze the alarm for 1 hour
	snoozeUntil := time.Now().Add(1 * time.Hour).Unix()
	err = client.Alarms.Snooze(ctx, alarmID, snoozeUntil)
	if err != nil {
		t.Fatalf("Failed to snooze alarm: %v", err)
	}
	t.Logf("Snoozed alarm until %v", time.Unix(snoozeUntil, 0))

	// Verify the alarm is snoozed
	snoozed, err := client.Alarms.Get(ctx, alarmID)
	if err != nil {
		t.Fatalf("Failed to get snoozed alarm: %v", err)
	}
	if snoozed.Snooze == 0 {
		t.Error("Expected alarm.Snooze to be non-zero after snoozing")
	} else {
		t.Logf("Verified: Snooze timestamp is %d", snoozed.Snooze)
	}

	// Unsnooze the alarm
	err = client.Alarms.Unsnooze(ctx, alarmID)
	if err != nil {
		t.Fatalf("Failed to unsnooze alarm: %v", err)
	}
	t.Log("Unsnoozed alarm")

	// Verify the alarm is unsnoozed
	unsnoozed, err := client.Alarms.Get(ctx, alarmID)
	if err != nil {
		t.Fatalf("Failed to get unsnoozed alarm: %v", err)
	}
	if unsnoozed.Snooze != 0 {
		t.Errorf("Expected alarm.Snooze to be 0 after unsnoozing, got %d", unsnoozed.Snooze)
	} else {
		t.Log("Verified: Alarm is unsnoozed")
	}
}

// TestTaskEnableDisable tests enabling and disabling a task
// This is a separate test because it modifies data
func TestTaskEnableDisable(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Find an enabled task to test with
	tasks, err := client.Tasks.List(ctx, vergeos.WithFilter("enabled eq true"))
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) == 0 {
		t.Skip("No enabled tasks available for enable/disable testing")
	}

	task := tasks[0]
	taskID := int(task.Key)
	t.Logf("Testing enable/disable workflow on task %d (Name: %s)", taskID, task.Name)

	// Disable the task
	err = client.Tasks.Disable(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to disable task: %v", err)
	}
	t.Log("Disabled task")

	// Verify the task is disabled
	disabled, err := client.Tasks.Get(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to get disabled task: %v", err)
	}
	if disabled.Enabled {
		t.Error("Expected task.Enabled to be false after disabling")
	} else {
		t.Log("Verified: Task is disabled")
	}

	// Re-enable the task
	err = client.Tasks.Enable(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to enable task: %v", err)
	}
	t.Log("Re-enabled task")

	// Verify the task is enabled
	enabled, err := client.Tasks.Get(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to get enabled task: %v", err)
	}
	if !enabled.Enabled {
		t.Error("Expected task.Enabled to be true after enabling")
	} else {
		t.Log("Verified: Task is enabled")
	}
}
