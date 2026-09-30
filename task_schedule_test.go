package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTaskScheduleService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != taskScheduleListFields {
				t.Errorf("fields = %q", fields)
			}
			jsonResponse(w, 200, []TaskSchedule{
				{Key: 1, Name: "nightly", RepeatEvery: TaskScheduleRepeatDay},
			})
		},
	}))

	rows, err := client.TaskSchedules.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "nightly" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestTaskScheduleService_ListEnabled(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "enabled eq true" {
				t.Errorf("filter = %q", got)
			}
			jsonResponse(w, 200, []TaskSchedule{{Key: 1, Enabled: true}})
		},
	}))

	rows, err := client.TaskSchedules.ListEnabled(context.Background())
	if err != nil {
		t.Fatalf("ListEnabled: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows", len(rows))
	}
}

func TestTaskScheduleService_ListDisabled_CustomFields(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "enabled eq false" {
				t.Errorf("filter = %q", got)
			}
			if got := r.URL.Query().Get("fields"); got != "all" {
				t.Errorf("fields = %q", got)
			}
			jsonResponse(w, 200, []TaskSchedule{})
		},
	}))

	if _, err := client.TaskSchedules.ListDisabled(context.Background(), WithFields("all")); err != nil {
		t.Fatal(err)
	}
}

func TestTaskScheduleService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskSchedule{
				Key: 4, Name: "friday", RepeatEvery: TaskScheduleRepeatWeek, Friday: true, Creator: "users/2",
			})
		},
	}))

	row, err := client.TaskSchedules.Get(context.Background(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "friday" || row.Creator.String() != "users/2" || !row.Friday {
		t.Fatalf("row = %+v", row)
	}
}

func TestTaskScheduleService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TaskSchedules.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScheduleService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "name eq 'friday'" {
				t.Errorf("filter = %q", got)
			}
			jsonResponse(w, 200, []TaskSchedule{{Key: 4, Name: "friday"}})
		},
		"GET /api/v4/task_schedules/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskSchedule{Key: 4, Name: "friday"})
		},
	}))

	row, err := client.TaskSchedules.GetByName(context.Background(), "friday")
	if err != nil {
		t.Fatal(err)
	}
	if row.Key != 4 {
		t.Fatalf("key = %d", row.Key)
	}
}

func TestTaskScheduleService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TaskSchedule{})
		},
	}))

	_, err := client.TaskSchedules.GetByName(context.Background(), "missing")
	if !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScheduleService_GetByName_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TaskSchedule{{Key: 1, Name: "daily"}, {Key: 2, Name: "daily"}})
		},
	}))

	_, err := client.TaskSchedules.GetByName(context.Background(), "daily")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScheduleService_Create_Defaults(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "hourly" || body["repeat_every"] != "hour" {
				t.Fatalf("body = %#v", body)
			}
			if body["enabled"] != true || body["repeat_iteration"] != float64(1) {
				t.Fatalf("body = %#v", body)
			}
			if body["start_time_of_day"] != float64(0) || body["end_time_of_day"] != float64(86400) {
				t.Fatalf("body = %#v", body)
			}
			if body["day_of_month"] != "start_date" || body["monday"] != true || body["sunday"] != true {
				t.Fatalf("body = %#v", body)
			}
			if _, ok := body["description"]; ok {
				t.Fatalf("description should be omitted: %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 7})
		},
		"GET /api/v4/task_schedules/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskSchedule{Key: 7, Name: "hourly", RepeatEvery: "hour"})
		},
	}))

	row, err := client.TaskSchedules.Create(context.Background(), &TaskScheduleCreateRequest{Name: "hourly"})
	if err != nil {
		t.Fatal(err)
	}
	if row.Key != 7 {
		t.Fatalf("key = %d", row.Key)
	}
}

func TestTaskScheduleService_Create_AllOptions(t *testing.T) {
	off := false
	on := true
	iteration := 2
	start := 32400
	end := 64800
	taskID := 100
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_schedules": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "weekly" || body["description"] != "reports" || body["repeat_every"] != "week" {
				t.Fatalf("body = %#v", body)
			}
			if body["saturday"] != false || body["sunday"] != false || body["friday"] != true {
				t.Fatalf("days = %#v", body)
			}
			if body["task"] != float64(100) || body["start_date"] != "2024-01-01" || body["day_of_month"] != "first" {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskSchedule{Key: 1, Name: "weekly"})
		},
	}))

	_, err := client.TaskSchedules.Create(context.Background(), &TaskScheduleCreateRequest{
		Name:            "weekly",
		Description:     "reports",
		Enabled:         &on,
		RepeatEvery:     TaskScheduleRepeatWeek,
		RepeatIteration: &iteration,
		StartDate:       "2024-01-01",
		EndDate:         "2024-12-31 23:59:59",
		StartTimeOfDay:  &start,
		EndTimeOfDay:    &end,
		DayOfMonth:      TaskScheduleDayOfMonthFirst,
		Monday:          &on,
		Friday:          &on,
		Saturday:        &off,
		Sunday:          &off,
		Task:            &taskID,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskScheduleService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	if _, err := client.TaskSchedules.Create(context.Background(), nil); !IsValidationError(err) {
		t.Fatalf("nil: %v", err)
	}
	if _, err := client.TaskSchedules.Create(context.Background(), &TaskScheduleCreateRequest{}); !IsValidationError(err) {
		t.Fatalf("name: %v", err)
	}
	bad := 90000
	if _, err := client.TaskSchedules.Create(context.Background(), &TaskScheduleCreateRequest{
		Name: "x", StartTimeOfDay: &bad,
	}); !IsValidationError(err) {
		t.Fatalf("time: %v", err)
	}
}

func TestTaskScheduleService_Update(t *testing.T) {
	name := "renamed"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if len(body) != 1 || body["name"] != "renamed" {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{})
		},
		"GET /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskSchedule{Key: 1, Name: "renamed"})
		},
	}))

	row, err := client.TaskSchedules.Update(context.Background(), 1, &TaskScheduleUpdateRequest{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "renamed" {
		t.Fatalf("name = %q", row.Name)
	}
}

func TestTaskScheduleService_Update_NotFound(t *testing.T) {
	name := "renamed"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedules/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TaskSchedules.Update(context.Background(), 9, &TaskScheduleUpdateRequest{Name: &name})
	if !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScheduleService_EnableDisable(t *testing.T) {
	var enabled []bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Enabled *bool `json:"enabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Enabled == nil {
				t.Fatal("enabled missing")
			}
			enabled = append(enabled, *body.Enabled)
			jsonResponse(w, 200, map[string]any{})
		},
		"GET /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskSchedule{Key: 1})
		},
	}))

	if _, err := client.TaskSchedules.Disable(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TaskSchedules.Enable(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if len(enabled) != 2 || enabled[0] || !enabled[1] {
		t.Fatalf("enabled = %v", enabled)
	}
}

func TestTaskScheduleService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(204)
		},
	}))
	if err := client.TaskSchedules.Delete(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
}

func TestTaskScheduleService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/task_schedules/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	if err := client.TaskSchedules.Delete(context.Background(), 9); !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScheduleService_GetSchedule(t *testing.T) {
	start := int64(1704067200)
	end := int64(1704153600)
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedules/1": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("action"); got != "get_schedule" {
				t.Errorf("action = %q", got)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["max"] != float64(10) || body["start_time"] != float64(start) || body["end_time"] != float64(end) {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{
				"times": []map[string]any{{"time": start}, {"time": end}},
			})
		},
	}))

	times, err := client.TaskSchedules.GetSchedule(context.Background(), 1, &TaskScheduleQuery{
		MaxResults: intPtr(10),
		StartTime:  &start,
		EndTime:    &end,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(times) != 2 || times[0]["time"] != float64(start) {
		t.Fatalf("times = %#v", times)
	}
}

func TestTaskScheduleService_GetSchedule_BoundsAndShapes(t *testing.T) {
	tests := []struct {
		name    string
		query   *TaskScheduleQuery
		respond any
		wantMax float64
		wantLen int
	}{
		{name: "default", query: nil, respond: map[string]any{}, wantMax: 100, wantLen: 0},
		{name: "clamp low", query: &TaskScheduleQuery{MaxResults: intPtr(0)}, respond: []any{}, wantMax: 1, wantLen: 0},
		{name: "clamp high", query: &TaskScheduleQuery{MaxResults: intPtr(9000)}, respond: map[string]any{"result": "ok"}, wantMax: 1440, wantLen: 0},
		{name: "empty times uses schedule", query: nil, respond: map[string]any{
			"times":    []any{},
			"schedule": []any{map[string]any{"time": 5}},
		}, wantMax: 100, wantLen: 1},
		{name: "non-list times", query: nil, respond: map[string]any{"times": "soon"}, wantMax: 100, wantLen: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"PUT /api/v4/task_schedules/3": func(w http.ResponseWriter, r *http.Request) {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if body["max"] != tt.wantMax {
						t.Fatalf("max = %v", body["max"])
					}
					jsonResponse(w, 200, tt.respond)
				},
			}))
			times, err := client.TaskSchedules.GetSchedule(context.Background(), 3, tt.query)
			if err != nil {
				t.Fatal(err)
			}
			if len(times) != tt.wantLen {
				t.Fatalf("len = %d, times = %#v", len(times), times)
			}
		})
	}
}

func TestTaskScheduleService_GetSchedule_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_schedules/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.TaskSchedules.GetSchedule(context.Background(), 9, nil)
	if !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestScheduleTimes_Scalar(t *testing.T) {
	times, err := scheduleTimes(json.RawMessage(`true`))
	if err != nil {
		t.Fatal(err)
	}
	if len(times) != 0 {
		t.Fatalf("times = %#v", times)
	}
}

func TestTaskSchedule_NullTask(t *testing.T) {
	var row TaskSchedule
	if err := json.Unmarshal([]byte(`{"$key":1,"name":"open","task":null,"creator":9}`), &row); err != nil {
		t.Fatal(err)
	}
	if row.Task != 0 || row.Creator.String() != "9" {
		t.Fatalf("row = %+v", row)
	}
}
