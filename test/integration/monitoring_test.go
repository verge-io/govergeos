//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestAlarmTypesList tests the AlarmTypes service.
func TestAlarmTypesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

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
	t.Run("Get", func(t *testing.T) {
		if first.Key == "" {
			t.Skip("No alarm type key available")
		}
		fetched, err := client.AlarmTypes.Get(ctx, first.Key)
		if err != nil {
			t.Errorf("AlarmTypes.Get(%q) failed: %v", first.Key, err)
		} else {
			t.Logf("AlarmTypes.Get succeeded: Key=%q, Description=%q", fetched.Key, fetched.Description)
		}
	})

	// Test ListByLevel
	t.Run("ListByLevel", func(t *testing.T) {
		warningTypes, err := client.AlarmTypes.ListByLevel(ctx, vergeos.AlarmLevelWarning)
		if err != nil {
			t.Errorf("AlarmTypes.ListByLevel failed: %v", err)
		} else {
			t.Logf("Found %d alarm types with level 'warning'", len(warningTypes))
		}
	})

	// Pretty print first alarm type for field verification
	prettyPrint(t, "Sample AlarmType", first)
}

// TestAlarmsList tests the Alarms service.
func TestAlarmsList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Alarms service...")

	// List all alarms
	alarms, err := client.Alarms.List(ctx)
	if err != nil {
		t.Fatalf("Alarms.List failed: %v", err)
	}

	t.Logf("Found %d alarms", len(alarms))

	// Test ListActive even if no alarms
	t.Run("ListActive", func(t *testing.T) {
		activeAlarms, err := client.Alarms.ListActive(ctx)
		if err != nil {
			t.Errorf("Alarms.ListActive failed: %v", err)
		} else {
			t.Logf("ListActive returned %d active alarms", len(activeAlarms))
		}
	})

	if len(alarms) == 0 {
		t.Log("No alarms found - system is healthy or no alerts configured")
		return
	}

	// Log first alarm to verify field mapping
	first := alarms[0]
	t.Logf("First alarm: Key=%d, Owner=%q, OwnerType=%q, Level=%q, Status=%q",
		int(first.Key), first.Owner, first.OwnerType, first.Level, first.Status)

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.Alarms.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Alarms.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Alarms.Get succeeded: AlarmID=%q, Resolvable=%v", fetched.AlarmID, fetched.Resolvable)
		}
	})

	// Test ListByLevel
	t.Run("ListByLevel", func(t *testing.T) {
		errorAlarms, err := client.Alarms.ListByLevel(ctx, vergeos.AlarmLevelError)
		if err != nil {
			t.Errorf("Alarms.ListByLevel failed: %v", err)
		} else {
			t.Logf("Found %d alarms with level 'error'", len(errorAlarms))
		}
	})

	// Test ListByOwner if we have an owner
	t.Run("ListByOwner", func(t *testing.T) {
		if first.Owner == "" {
			t.Skip("No owner available")
		}
		ownerAlarms, err := client.Alarms.ListByOwner(ctx, first.Owner)
		if err != nil {
			t.Errorf("Alarms.ListByOwner failed: %v", err)
		} else {
			t.Logf("Found %d alarms for owner %q", len(ownerAlarms), first.Owner)
		}
	})

	// Pretty print first alarm for field verification
	prettyPrint(t, "Sample Alarm", first)
}

// TestAlarmsSnooze tests snoozing and unsnoozing an alarm.
func TestAlarmsSnooze(t *testing.T) {
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

// TestTasksList tests the Tasks service.
func TestTasksList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Tasks service...")

	// List all tasks
	tasks, err := client.Tasks.List(ctx)
	if err != nil {
		t.Fatalf("Tasks.List failed: %v", err)
	}

	t.Logf("Found %d tasks", len(tasks))

	// Test ListRunning even if no tasks
	t.Run("ListRunning", func(t *testing.T) {
		runningTasks, err := client.Tasks.ListRunning(ctx)
		if err != nil {
			t.Errorf("Tasks.ListRunning failed: %v", err)
		} else {
			t.Logf("ListRunning returned %d running tasks", len(runningTasks))
		}
	})

	// Test ListEnabled
	t.Run("ListEnabled", func(t *testing.T) {
		enabledTasks, err := client.Tasks.ListEnabled(ctx)
		if err != nil {
			t.Errorf("Tasks.ListEnabled failed: %v", err)
		} else {
			t.Logf("Found %d enabled tasks", len(enabledTasks))
		}
	})

	if len(tasks) == 0 {
		t.Log("No tasks found - no scheduled tasks configured")
		return
	}

	// Log first task to verify field mapping
	first := tasks[0]
	t.Logf("First task: Key=%d, ID=%q, Owner=%q, Action=%q, Name=%q, Status=%q",
		int(first.Key), first.ID, first.Owner, first.Action, first.Name, first.Status)

	// Test Get by Key
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.Tasks.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Tasks.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Tasks.Get succeeded: Name=%q, Enabled=%v, Status=%q",
				fetched.Name, fetched.Enabled, fetched.Status)
		}
	})

	// Test GetByID (SHA1 hash)
	t.Run("GetByID", func(t *testing.T) {
		if first.ID == "" {
			t.Skip("No task ID available")
		}
		byID, err := client.Tasks.GetByID(ctx, first.ID)
		if err != nil {
			t.Errorf("Tasks.GetByID(%q) failed: %v", first.ID, err)
		} else {
			t.Logf("Tasks.GetByID succeeded: Key=%d", int(byID.Key))
		}
	})

	// Test ListByOwner if we have an owner
	t.Run("ListByOwner", func(t *testing.T) {
		if first.Owner == "" {
			t.Skip("No owner available")
		}
		ownerTasks, err := client.Tasks.ListByOwner(ctx, first.Owner)
		if err != nil {
			t.Errorf("Tasks.ListByOwner failed: %v", err)
		} else {
			t.Logf("Found %d tasks for owner %q", len(ownerTasks), first.Owner)
		}
	})

	// Pretty print first task for field verification
	prettyPrint(t, "Sample Task", first)
}

// TestTasksEnableDisable tests enabling and disabling a task.
func TestTasksEnableDisable(t *testing.T) {
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
