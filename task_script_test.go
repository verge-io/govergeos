package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTaskScriptService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_scripts": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got != taskScriptListFields {
				t.Errorf("fields = %q", got)
			}
			jsonResponse(w, 200, []TaskScript{{
				Key: 1, Name: "cleanup", Script: "log('hi')", TaskCount: 2,
			}})
		},
	}))

	rows, err := client.TaskScripts.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].TaskCount != 2 || rows[0].Script != "log('hi')" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestTaskScriptService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_scripts": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "name eq 'cleanup'" {
				t.Errorf("filter = %q", got)
			}
			jsonResponse(w, 200, []TaskScript{{Key: 3, Name: "cleanup"}})
		},
		"GET /api/v4/task_scripts/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskScript{Key: 3, Name: "cleanup", Script: "log(1)"})
		},
	}))

	row, err := client.TaskScripts.GetByName(context.Background(), "cleanup")
	if err != nil {
		t.Fatal(err)
	}
	if row.Script != "log(1)" {
		t.Fatalf("script = %q", row.Script)
	}
}

func TestTaskScriptService_GetByName_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_scripts": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TaskScript{{Key: 1, Name: "cleanup"}, {Key: 2, Name: "cleanup"}})
		},
	}))
	_, err := client.TaskScripts.GetByName(context.Background(), "cleanup")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestTaskScriptService_Create_DefaultSettings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_scripts": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			settings, _ := body["task_settings"].(map[string]any)
			questions, _ := settings["questions"].([]any)
			if body["name"] != "cleanup" || body["script"] != "log('hi')" || questions == nil || len(questions) != 0 {
				t.Fatalf("body = %#v", body)
			}
			if _, ok := body["description"]; ok {
				t.Fatalf("description should be omitted: %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 8})
		},
		"GET /api/v4/task_scripts/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskScript{Key: 8, Name: "cleanup"})
		},
	}))

	row, err := client.TaskScripts.Create(context.Background(), &TaskScriptCreateRequest{
		Name: "cleanup", Script: "log('hi')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if row.Key != 8 {
		t.Fatalf("key = %d", row.Key)
	}
}

func TestTaskScriptService_Create_Settings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_scripts": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			settings, _ := body["task_settings"].(map[string]any)
			questions, _ := settings["questions"].([]any)
			if body["description"] != "notes" || len(questions) != 1 {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/task_scripts/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskScript{Key: 1, Name: "full"})
		},
	}))

	_, err := client.TaskScripts.Create(context.Background(), &TaskScriptCreateRequest{
		Name:        "full",
		Script:      "log('Hello World')",
		Description: "notes",
		TaskSettings: map[string]any{
			"questions": []any{map[string]any{"name": "target", "type": "string"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskScriptService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	if _, err := client.TaskScripts.Create(context.Background(), nil); !IsValidationError(err) {
		t.Fatalf("nil: %v", err)
	}
	if _, err := client.TaskScripts.Create(context.Background(), &TaskScriptCreateRequest{Script: "log(1)"}); !IsValidationError(err) {
		t.Fatalf("name: %v", err)
	}
	if _, err := client.TaskScripts.Create(context.Background(), &TaskScriptCreateRequest{Name: "x"}); !IsValidationError(err) {
		t.Fatalf("script: %v", err)
	}
}

func TestTaskScriptService_Update(t *testing.T) {
	script := "log('updated')"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_scripts/1": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if len(body) != 1 || body["script"] != script {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{})
		},
		"GET /api/v4/task_scripts/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TaskScript{Key: 1, Script: script})
		},
	}))

	row, err := client.TaskScripts.Update(context.Background(), 1, &TaskScriptUpdateRequest{Script: &script})
	if err != nil {
		t.Fatal(err)
	}
	if row.Script != script {
		t.Fatalf("script = %q", row.Script)
	}
}

func TestTaskScriptService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/task_scripts/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(204)
		},
	}))
	if err := client.TaskScripts.Delete(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
}

func TestTaskScriptService_Run(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_scripts/1": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("action") != "run" {
				t.Errorf("action = %q", r.URL.Query().Get("action"))
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["target_vm"] != float64(100) || body["cleanup"] != true {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"task": 123, "status": "running"})
		},
	}))

	got, err := client.TaskScripts.Run(context.Background(), 1, map[string]any{
		"target_vm": 100,
		"cleanup":   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["task"] != float64(123) || got["status"] != "running" {
		t.Fatalf("got = %#v", got)
	}
}

func TestTaskScriptService_Run_NoParams(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/task_scripts/1": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if len(body) != 0 {
				t.Fatalf("body = %#v", body)
			}
			jsonResponse(w, 200, map[string]any{"task": 9})
		},
	}))
	got, err := client.TaskScripts.Run(context.Background(), 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["task"] != float64(9) {
		t.Fatalf("got = %#v", got)
	}
}

func TestTaskScriptService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/task_scripts/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.TaskScripts.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("got %v", err)
	}
}
