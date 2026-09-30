package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTaskScheduleTriggerService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedule_triggers": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != taskScheduleTriggerListFields {
				t.Errorf("fields = %q", fields)
			}
			jsonResponse(w, 200, []TaskScheduleTrigger{{
				Key: 1, Task: 10, Schedule: 20, TaskDisplay: "backup", ScheduleDisplay: "nightly",
				ScheduleEnabled: true, ScheduleRepeatEvery: "Day(s)",
			}})
		},
	}))

	rows, err := client.TaskScheduleTriggers.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Task != 10 || rows[0].ScheduleDisplay != "nightly" || !rows[0].ScheduleEnabled {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestTaskScheduleTriggerService_ListByTaskAndSchedule(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedule_triggers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TaskScheduleTrigger{})
			filter := r.URL.Query().Get("filter")
			switch filter {
			case "task eq 10", "schedule eq 20":
			default:
				t.Errorf("filter = %q", filter)
			}
		},
	}))
	if _, err := client.TaskScheduleTriggers.ListByTask(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TaskScheduleTriggers.ListBySchedule(context.Background(), 20); err != nil {
		t.Fatal(err)
	}
}

func TestTaskScheduleTriggerService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedule_triggers/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.TaskScheduleTriggers.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScheduleTriggerService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_schedule_triggers": func(w http.ResponseWriter, r *http.Request) {
			var body TaskScheduleTriggerCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Task != 10 || body.Schedule != 20 {
				t.Fatalf("body = %+v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 5})
		},
		"GET /api/v4/task_schedule_triggers/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskScheduleTrigger{Key: 5, Task: 10, Schedule: 20})
		},
	}))

	row, err := client.TaskScheduleTriggers.Create(context.Background(), &TaskScheduleTriggerCreateRequest{
		Task: 10, Schedule: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if row.Key != 5 {
		t.Fatalf("key = %d", row.Key)
	}
}

func TestTaskScheduleTriggerService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	if _, err := client.TaskScheduleTriggers.Create(context.Background(), nil); !IsValidationError(err) {
		t.Fatalf("nil: %v", err)
	}
	if _, err := client.TaskScheduleTriggers.Create(context.Background(), &TaskScheduleTriggerCreateRequest{Schedule: 1}); !IsValidationError(err) {
		t.Fatalf("task: %v", err)
	}
	if _, err := client.TaskScheduleTriggers.Create(context.Background(), &TaskScheduleTriggerCreateRequest{Task: 1}); !IsValidationError(err) {
		t.Fatalf("schedule: %v", err)
	}
}

func TestTaskScheduleTriggerService_Delete(t *testing.T) {
	deleted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/task_schedule_triggers/5": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(204)
		},
	}))
	if err := client.TaskScheduleTriggers.Delete(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("delete was not called")
	}
}

func TestTaskScheduleTriggerService_Trigger(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedule_triggers/5": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("action") != "trigger" {
				t.Errorf("action = %q", r.URL.Query().Get("action"))
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if len(body) != 0 {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"triggered": true})
		},
	}))

	got, err := client.TaskScheduleTriggers.Trigger(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got["triggered"] != true {
		t.Fatalf("got = %#v", got)
	}
}

func TestTaskScheduleTriggerService_Trigger_EmptyBody(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedule_triggers/5": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))
	got, err := client.TaskScheduleTriggers.Trigger(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got = %#v", got)
	}
}
