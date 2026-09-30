---
title: Monitoring
description: Monitor system health with alarms, tasks, logs, and webhook notifications
tags: [alarm, alarm-type, task, task-schedule, task-event, task-script, log, webhook, monitoring, health, audit, notification, alert]
categories: [Monitoring]
---

# Monitoring

Monitor system health with alarms, tasks, logs, and webhooks.

## Alarms

Monitor system health with alarms for VMs, nodes, networks, and other resources.

```go
// List all alarms
alarms, err := client.Alarms.List(ctx)

// List active (non-snoozed) alarms
alarms, err := client.Alarms.ListActive(ctx)

// List alarms for a specific resource
alarms, err := client.Alarms.ListByOwner(ctx, "vms/123")

// List alarms by severity level
alarms, err := client.Alarms.ListByLevel(ctx, vergeos.AlarmLevelCritical)

// Get an alarm by ID
alarm, err := client.Alarms.Get(ctx, alarmID)

// Snooze an alarm until a specific timestamp
err = client.Alarms.Snooze(ctx, alarmID, time.Now().Add(24*time.Hour).Unix())

// Unsnooze an alarm
err = client.Alarms.Unsnooze(ctx, alarmID)

// Resolve a resolvable alarm
err = client.Alarms.Resolve(ctx, alarmID)

// Delete an alarm
err = client.Alarms.Delete(ctx, alarmID)
```

---

## Alarm Types

Query alarm type definitions (read-only reference data).

```go
// List all alarm types
alarmTypes, err := client.AlarmTypes.List(ctx)

// Get an alarm type by key (string, not integer)
alarmType, err := client.AlarmTypes.Get(ctx, "vm_cpu_high")

// List alarm types by default severity level
alarmTypes, err := client.AlarmTypes.ListByLevel(ctx, vergeos.AlarmLevelWarning)
```

---

## Tasks

Manage scheduled and automated tasks.

```go
// List all tasks
tasks, err := client.Tasks.List(ctx)

// List running tasks
tasks, err := client.Tasks.ListRunning(ctx)

// List tasks for a specific resource
tasks, err := client.Tasks.ListByOwner(ctx, "vms/123")

// Get a task by ID
task, err := client.Tasks.Get(ctx, taskID)

// Get a task by its 40-character SHA1 ID
task, err := client.Tasks.GetByID(ctx, "abc123...")

// Create a task
task, err := client.Tasks.Create(ctx, &vergeos.TaskCreateRequest{
    Owner:  "vms/123",
    Action: "snapshot",
    Name:   "Daily Snapshot",
})

// Update a task
task, err := client.Tasks.Update(ctx, taskID, &vergeos.TaskUpdateRequest{
    Name: ptr("Updated Task Name"),
})

// Execute a task immediately
err = client.Tasks.Execute(ctx, taskID, nil)

// Enable/disable a task
err = client.Tasks.Enable(ctx, taskID)
err = client.Tasks.Disable(ctx, taskID)

// Delete a task
err = client.Tasks.Delete(ctx, taskID)
```

---

## Task schedules

Reusable schedules on `task_schedules`. A schedule is created once and linked to tasks through schedule triggers. Create fills the same defaults as pyVergeOS when a field is left unset: enabled, every day of the week, `repeat_every` hour, iteration 1, start of day through 86400 seconds, and `day_of_month` `start_date`.

```go
// Weekdays at 02:00 (7200 seconds from midnight)
schedule, err := client.TaskSchedules.Create(ctx, &vergeos.TaskScheduleCreateRequest{
    Name:           "Nightly",
    RepeatEvery:    vergeos.TaskScheduleRepeatDay,
    StartTimeOfDay: ptr(7200),
    Saturday:       ptr(false),
    Sunday:         ptr(false),
})

// Upcoming runs. Nil asks for 100. A set MaxResults is clamped to 1..1440.
times, err := client.TaskSchedules.GetSchedule(ctx, schedule.Key.Int(), nil)

schedule, err = client.TaskSchedules.Disable(ctx, schedule.Key.Int())
err = client.TaskSchedules.Delete(ctx, schedule.Key.Int())
```

Repeat intervals: `TaskScheduleRepeatMinute`, `Hour`, `Day`, `Week`, `Month`, `Year`, and `Never`.

---

## Task schedule triggers

A trigger on `task_schedule_triggers` links one task to one schedule. `Trigger` runs that task immediately (`PUT ...?action=trigger`).

```go
trigger, err := client.TaskScheduleTriggers.Create(ctx, &vergeos.TaskScheduleTriggerCreateRequest{
    Task:     taskID,
    Schedule: scheduleID,
})

rows, err := client.TaskScheduleTriggers.ListBySchedule(ctx, scheduleID)
result, err := client.TaskScheduleTriggers.Trigger(ctx, trigger.Key.Int())
err = client.TaskScheduleTriggers.Delete(ctx, trigger.Key.Int())
```

---

## Task events

An event on `task_events` runs a task when something happens. The platform fills `owner` from the task, so create does not send it.

```go
event, err := client.TaskEvents.Create(ctx, &vergeos.TaskEventCreateRequest{
    Task:              taskID,
    Table:             "alarms",
    Event:             "lowered",
    TableEventFilters: map[string]any{"level": "summary"},
})

rows, err := client.TaskEvents.ListByTask(ctx, taskID)
result, err := client.TaskEvents.Trigger(ctx, event.Key.Int(), map[string]any{"custom": "data"})
err = client.TaskEvents.Delete(ctx, event.Key.Int())
```

---

## Task scripts

GCS scripts on `task_scripts`. A create that omits settings sends `{"questions": []}`. `Run` is `PUT /task_scripts/{id}?action=run`. Parameters are the action body.

```go
script, err := client.TaskScripts.Create(ctx, &vergeos.TaskScriptCreateRequest{
    Name:   "Cleanup",
    Script: "log('Cleanup started')",
})

result, err := client.TaskScripts.Run(ctx, script.Key.Int(), map[string]any{"target_vm": 100})
err = client.TaskScripts.Delete(ctx, script.Key.Int())
```

---

## Logs

Query system logs for audit, operational, and error information (read-only).

```go
// List recent logs (newest first by default)
logs, err := client.Logs.List(ctx)

// Get the most recent 50 logs
logs, err := client.Logs.GetRecent(ctx, 50)

// List logs by level
audits, err := client.Logs.ListAudit(ctx)
warnings, err := client.Logs.ListWarnings(ctx)
errors, err := client.Logs.ListErrors(ctx)  // error + critical

// Get recent errors
recentErrors, err := client.Logs.GetRecentErrors(ctx, 20)

// List logs by object type
vmLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeVM)
tenantLogs, err := client.Logs.ListByObjectType(ctx, vergeos.LogObjectTypeTenant)

// List logs by username
userLogs, err := client.Logs.ListByUser(ctx, "admin")

// List logs since a timestamp (microseconds since epoch)
since := time.Now().Add(-24 * time.Hour).UnixMicro()
logs, err := client.Logs.ListSince(ctx, since)

// Search logs by text pattern
logs, err := client.Logs.Search(ctx, "failed to connect")

// Get a specific log entry
log, err := client.Logs.Get(ctx, logID)
```

Log level constants:
- `vergeos.LogLevelAudit` - User actions and authentication
- `vergeos.LogLevelMessage` - Informational messages
- `vergeos.LogLevelWarning` - Warning conditions
- `vergeos.LogLevelError` - Error conditions
- `vergeos.LogLevelCritical` - Critical errors
- `vergeos.LogLevelSummary` - Summary information
- `vergeos.LogLevelDebug` - Debug information

Common object types:
- `vergeos.LogObjectTypeVM` - Virtual machine logs
- `vergeos.LogObjectTypeTenant` - Tenant logs
- `vergeos.LogObjectTypeVNet` - Network logs
- `vergeos.LogObjectTypeUser` - User account logs
- `vergeos.LogObjectTypeSystem` - System logs
- `vergeos.LogObjectTypeCluster` - Cluster logs
- `vergeos.LogObjectTypeSite` - DR site logs

---

## Webhooks

Configure webhook endpoints for event notifications.

```go
// List all webhook URLs
webhooks, err := client.WebhookURLs.List(ctx)

// Create a webhook endpoint
webhook, err := client.WebhookURLs.Create(ctx, &vergeos.WebhookURLCreateRequest{
    Name:              "slack-alerts",
    URL:               "https://hooks.slack.com/services/xxx",
    AuthorizationType: vergeos.WebhookAuthNone,
    Timeout:           ptr(10),
    Retries:           ptr(3),
})

// Create a webhook with bearer token auth
webhook, err := client.WebhookURLs.Create(ctx, &vergeos.WebhookURLCreateRequest{
    Name:               "api-endpoint",
    URL:                "https://api.example.com/webhook",
    AuthorizationType:  vergeos.WebhookAuthBearer,
    AuthorizationValue: "your-bearer-token",
})

// Send a test message
err = client.WebhookURLs.Send(ctx, webhookID, `{"text": "Test message"}`)

// View webhook delivery log
messages, err := client.Webhooks.List(ctx)

// List failed deliveries
failed, err := client.Webhooks.ListFailed(ctx)

// Delete a webhook
err = client.WebhookURLs.Delete(ctx, webhookID)
```
