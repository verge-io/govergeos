//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
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

// TestAlarmsSnooze snoozes and unsnoozes one alarm.
//
// Alarms are raised by the platform and cannot be created through the API,
// so this test runs only when VERGEOS_TEST_ALARM_ID names the alarm. The
// previous snooze timestamp is restored with t.Cleanup, registered
// immediately after the snooze.
func TestAlarmsSnooze(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	alarmIDStr := os.Getenv("VERGEOS_TEST_ALARM_ID")
	if alarmIDStr == "" {
		t.Skip("Skipping snooze test: VERGEOS_TEST_ALARM_ID not set")
	}

	alarmID, err := strconv.Atoi(alarmIDStr)
	if err != nil {
		t.Fatalf("Invalid VERGEOS_TEST_ALARM_ID: %v", err)
	}

	alarm, err := client.Alarms.Get(ctx, alarmID)
	if err != nil {
		t.Fatalf("Alarms.Get(%d) failed: %v", alarmID, err)
	}
	originalSnooze := alarm.Snooze
	t.Logf("Modifying alarm key=%d alarm_id=%q status=%q owner=%q level=%q snooze=%d",
		alarmID, alarm.AlarmID, alarm.Status, alarm.Owner, alarm.Level, originalSnooze)

	snoozeUntil := time.Now().Add(time.Hour).Unix()
	if err := client.Alarms.Snooze(ctx, alarmID, snoozeUntil); err != nil {
		t.Fatalf("Failed to snooze alarm key=%d alarm_id=%q status=%q: %v",
			alarmID, alarm.AlarmID, alarm.Status, err)
	}
	t.Logf("Snoozed alarm key=%d alarm_id=%q status=%q until %s",
		alarmID, alarm.AlarmID, alarm.Status, time.Unix(snoozeUntil, 0).UTC().Format(time.RFC3339))

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := client.Alarms.Snooze(cleanupCtx, alarmID, originalSnooze)
		if err != nil && originalSnooze <= time.Now().Unix() {
			t.Logf("Restoring snooze=%d on alarm key=%d alarm_id=%q status=%q failed: %v; unsnoozing",
				originalSnooze, alarmID, alarm.AlarmID, alarm.Status, err)
			err = client.Alarms.Unsnooze(cleanupCtx, alarmID)
		}
		if err != nil {
			t.Errorf("Failed to restore alarm key=%d alarm_id=%q status=%q owner=%q to snooze=%d: %v",
				alarmID, alarm.AlarmID, alarm.Status, alarm.Owner, originalSnooze, err)
			return
		}
		t.Logf("Restored alarm key=%d alarm_id=%q status=%q owner=%q snooze=%d",
			alarmID, alarm.AlarmID, alarm.Status, alarm.Owner, originalSnooze)
	})

	snoozed, err := client.Alarms.Get(ctx, alarmID)
	if err != nil {
		t.Fatalf("Failed to get snoozed alarm key=%d alarm_id=%q: %v", alarmID, alarm.AlarmID, err)
	}
	if snoozed.Snooze == 0 {
		t.Error("Expected alarm.Snooze to be non-zero after snoozing")
	} else {
		t.Logf("Verified: alarm key=%d alarm_id=%q snooze=%d", alarmID, alarm.AlarmID, snoozed.Snooze)
	}

	if err := client.Alarms.Unsnooze(ctx, alarmID); err != nil {
		t.Fatalf("Failed to unsnooze alarm key=%d alarm_id=%q status=%q: %v",
			alarmID, alarm.AlarmID, alarm.Status, err)
	}
	t.Logf("Unsnoozed alarm key=%d alarm_id=%q status=%q", alarmID, alarm.AlarmID, alarm.Status)

	unsnoozed, err := client.Alarms.Get(ctx, alarmID)
	if err != nil {
		t.Fatalf("Failed to get unsnoozed alarm key=%d alarm_id=%q: %v", alarmID, alarm.AlarmID, err)
	}
	if unsnoozed.Snooze != 0 {
		t.Errorf("Expected alarm.Snooze to be 0 after unsnoozing, got %d", unsnoozed.Snooze)
	} else {
		t.Logf("Verified: alarm key=%d alarm_id=%q is unsnoozed", alarmID, alarm.AlarmID)
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

// TestTasksEnableDisable creates a disabled task on a throwaway VM, enables
// it, and disables it again. The task has no schedule or event trigger, and
// the VM is not powered on. Both objects are deleted with t.Cleanup,
// registered as soon as each one exists. Cleanup disables the task before
// deleting it.
func TestTasksEnableDisable(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	stamp := time.Now().Format("20060102-150405")
	vmName := "sdk-test-task-vm-" + stamp
	vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
		Name:        vmName,
		Description: "goVergeOS integration test VM for task enable/disable - safe to delete",
		CPUCores:    1,
		RAM:         512,
		OSFamily:    "linux",
	})
	if err != nil {
		t.Fatalf("VMs.Create(%q) failed: %v", vmName, err)
	}
	vmID := vm.ID.Int()
	if vm.Name != "" {
		vmName = vm.Name
	}
	if vmID == 0 {
		t.Fatalf("VMs.Create returned VM %q with an empty key", vmName)
	}
	t.Logf("Created VM key=%d name=%q", vmID, vmName)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := client.VMs.Delete(cleanupCtx, vmID); err != nil && !vergeos.IsNotFoundError(err) {
			t.Errorf("Failed to delete VM key=%d name=%q: %v", vmID, vmName, err)
			return
		}
		t.Logf("Deleted VM key=%d name=%q", vmID, vmName)
	})

	taskName := "sdk-test-task-" + stamp
	enabled := false
	owner := fmt.Sprintf("vms/%d", vmID)
	task, err := client.Tasks.Create(ctx, &vergeos.TaskCreateRequest{
		Owner:       owner,
		Action:      "snapshot",
		Name:        taskName,
		Description: "goVergeOS integration test task - safe to delete",
		Enabled:     &enabled,
	})
	if err != nil {
		t.Fatalf("Tasks.Create(%q) on VM key=%d name=%q failed: %v", taskName, vmID, vmName, err)
	}
	taskID := int(task.Key)
	if task.Name != "" {
		taskName = task.Name
	}
	if taskID == 0 {
		t.Fatalf("Tasks.Create returned task %q with an empty key", taskName)
	}
	t.Logf("Created task key=%d name=%q owner=%q action=%q enabled=%v",
		taskID, taskName, owner, task.Action, task.Enabled)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := client.Tasks.Disable(cleanupCtx, taskID); err != nil && !vergeos.IsNotFoundError(err) {
			t.Logf("Cleanup disable task key=%d name=%q: %v", taskID, taskName, err)
		}
		if err := client.Tasks.Delete(cleanupCtx, taskID); err != nil && !vergeos.IsNotFoundError(err) {
			t.Errorf("Failed to delete task key=%d name=%q owner=%q: %v", taskID, taskName, owner, err)
			return
		}
		t.Logf("Deleted task key=%d name=%q", taskID, taskName)
	})

	t.Logf("Enabling task key=%d name=%q", taskID, taskName)
	if err := client.Tasks.Enable(ctx, taskID); err != nil {
		t.Fatalf("Failed to enable task key=%d name=%q: %v", taskID, taskName, err)
	}

	enabledTask, err := client.Tasks.Get(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to get enabled task key=%d name=%q: %v", taskID, taskName, err)
	}
	if !enabledTask.Enabled {
		t.Error("Expected task.Enabled to be true after enabling")
	} else {
		t.Logf("Verified: task key=%d name=%q is enabled", taskID, taskName)
	}

	t.Logf("Disabling task key=%d name=%q", taskID, taskName)
	if err := client.Tasks.Disable(ctx, taskID); err != nil {
		t.Fatalf("Failed to disable task key=%d name=%q: %v", taskID, taskName, err)
	}

	disabledTask, err := client.Tasks.Get(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to get disabled task key=%d name=%q: %v", taskID, taskName, err)
	}
	if disabledTask.Enabled {
		t.Error("Expected task.Enabled to be false after disabling")
	} else {
		t.Logf("Verified: task key=%d name=%q is disabled", taskID, taskName)
	}
}
