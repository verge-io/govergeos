package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTaskEventService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_events": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got != taskEventListFields {
				t.Errorf("fields = %q", got)
			}
			jsonResponse(w, 200, []map[string]any{{
				"$key": 1, "owner": 123, "table": "alarms", "event": "lowered",
				"task": 10, "task_display": "notify",
				"table_event_filters": map[string]any{"level": "summary"},
			}})
		},
	}))

	rows, err := client.TaskEvents.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Owner.String() != "123" || rows[0].Event != "lowered" {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].TableEventFilters["level"] != "summary" {
		t.Fatalf("filters = %#v", rows[0].TableEventFilters)
	}
}

func TestTaskEventService_ListFilters(t *testing.T) {
	want := map[string]string{
		"task":  "task eq 10",
		"owner": "owner eq 8",
		"table": "table eq 'vms'",
		"event": `event eq 'power\'on'`,
	}
	seen := map[string]bool{}
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_events": func(w http.ResponseWriter, r *http.Request) {
			seen[r.URL.Query().Get("filter")] = true
			jsonResponse(w, 200, []TaskEvent{})
		},
	}))
	ctx := context.Background()
	if _, err := client.TaskEvents.ListByTask(ctx, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TaskEvents.ListByOwner(ctx, 8); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TaskEvents.ListByTable(ctx, "vms"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TaskEvents.ListByEvent(ctx, `power'on`); err != nil {
		t.Fatal(err)
	}
	for _, filter := range want {
		if !seen[filter] {
			t.Errorf("missing filter %q in %#v", filter, seen)
		}
	}
}

func TestTaskEventService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_events": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if _, ok := body["owner"]; ok {
				t.Fatalf("owner must not be sent: %#v", body)
			}
			if body["task"] != float64(10) || body["table"] != "alarms" || body["event"] != "lowered" {
				t.Fatalf("body = %#v", body)
			}
			filters, _ := body["table_event_filters"].(map[string]any)
			if filters["level"] != "summary" || body["event_name"] != "Alarm lowered" {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 4})
		},
		"GET /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{
				"$key": 4, "event": "lowered", "owner": "alarms/1", "task": 10,
			})
		},
	}))

	row, err := client.TaskEvents.Create(context.Background(), &TaskEventCreateRequest{
		Task:              10,
		Table:             "alarms",
		Event:             "lowered",
		EventName:         "Alarm lowered",
		TableEventFilters: map[string]any{"level": "summary"},
		Context:           map[string]any{"notify": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if row.Owner.String() != "alarms/1" || row.Key != 4 {
		t.Fatalf("row = %+v", row)
	}
}

func TestTaskEventService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	ctx := context.Background()
	if _, err := client.TaskEvents.Create(ctx, nil); !IsValidationError(err) {
		t.Fatalf("nil: %v", err)
	}
	if _, err := client.TaskEvents.Create(ctx, &TaskEventCreateRequest{Table: "vms", Event: "login"}); !IsValidationError(err) {
		t.Fatalf("task: %v", err)
	}
	if _, err := client.TaskEvents.Create(ctx, &TaskEventCreateRequest{Task: 1, Event: "login"}); !IsValidationError(err) {
		t.Fatalf("table: %v", err)
	}
	if _, err := client.TaskEvents.Create(ctx, &TaskEventCreateRequest{Task: 1, Table: "vms"}); !IsValidationError(err) {
		t.Fatalf("event: %v", err)
	}
}

func TestTaskEventService_Update(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			filters, _ := body["table_event_filters"].(map[string]any)
			if len(body) != 1 || filters["level"] != "critical" {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{})
		},
		"GET /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{
				"$key": 4, "table_event_filters": `{"level":"critical"}`,
			})
		},
	}))

	row, err := client.TaskEvents.Update(context.Background(), 4, &TaskEventUpdateRequest{
		TableEventFilters: map[string]any{"level": "critical"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if row.TableEventFilters["level"] != "critical" {
		t.Fatalf("filters = %#v", row.TableEventFilters)
	}
}

func TestTaskEventService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/task_events/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	if err := client.TaskEvents.Delete(context.Background(), 9); !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskEventService_Trigger(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("action") != "trigger" {
				t.Errorf("action = %q", r.URL.Query().Get("action"))
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			ctx, _ := body["context"].(map[string]any)
			if ctx["custom"] != "data" {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"triggered": true})
		},
	}))

	got, err := client.TaskEvents.Trigger(context.Background(), 4, map[string]any{"custom": "data"})
	if err != nil {
		t.Fatal(err)
	}
	if got["triggered"] != true {
		t.Fatalf("got = %#v", got)
	}
}

func TestTaskEventService_Trigger_NoContext(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if _, ok := body["context"]; ok || len(body) != 0 {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, []any{})
		},
	}))

	got, err := client.TaskEvents.Trigger(context.Background(), 4, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("non-object action result = %#v", got)
	}
}

func TestTaskEventService_Update_EmptyFilters(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			filters, ok := body["table_event_filters"].(map[string]any)
			if !ok || len(filters) != 0 {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{})
		},
		"GET /api/v4/task_events/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskEvent{Key: 4})
		},
	}))

	if _, err := client.TaskEvents.Update(context.Background(), 4, &TaskEventUpdateRequest{
		TableEventFilters: map[string]any{},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestJSONObject_NonObject(t *testing.T) {
	var row TaskEvent
	if err := json.Unmarshal([]byte(`{"$key":1,"context":"not-json","table_event_filters":null}`), &row); err != nil {
		t.Fatal(err)
	}
	if row.Context != nil || row.TableEventFilters != nil {
		t.Fatalf("row = %+v", row)
	}
}
