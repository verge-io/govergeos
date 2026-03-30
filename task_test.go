package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTaskService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Task{
				{Key: 1, Name: "daily-snapshot", Action: "snapshot"},
				{Key: 2, Name: "weekly-backup", Action: "backup"},
			})
		},
	}))

	tasks, err := client.Tasks.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Name != "daily-snapshot" {
		t.Errorf("expected name 'daily-snapshot', got %q", tasks[0].Name)
	}
}

func TestTaskService_ListRunning(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for running tasks")
			}
			jsonResponse(w, 200, []Task{{Key: 1, Name: "active-task", Status: "running"}})
		},
	}))

	tasks, err := client.Tasks.ListRunning(context.Background())
	if err != nil {
		t.Fatalf("ListRunning failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}

func TestTaskService_ListByOwner(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "owner eq 'vms/123'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Task{{Key: 1, Owner: "vms/123"}})
		},
	}))

	tasks, err := client.Tasks.ListByOwner(context.Background(), "vms/123")
	if err != nil {
		t.Fatalf("ListByOwner failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}

func TestTaskService_ListEnabled(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for enabled tasks")
			}
			jsonResponse(w, 200, []Task{{Key: 1, Enabled: true}})
		},
	}))

	tasks, err := client.Tasks.ListEnabled(context.Background())
	if err != nil {
		t.Fatalf("ListEnabled failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
}

func TestTaskService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 1, Name: "daily-snapshot", Action: "snapshot", Owner: "vms/123"})
		},
	}))

	task, err := client.Tasks.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if task.Name != "daily-snapshot" {
		t.Errorf("expected name 'daily-snapshot', got %q", task.Name)
	}
	if task.Owner != "vms/123" {
		t.Errorf("expected owner 'vms/123', got %q", task.Owner)
	}
}

func TestTaskService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Tasks.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTaskService_GetByID(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Task{{Key: 5, ID: "abc123def456", Name: "my-task"}})
		},
		"GET /api/v4/tasks/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 5, ID: "abc123def456", Name: "my-task", Owner: "vms/1"})
		},
	}))

	task, err := client.Tasks.GetByID(context.Background(), "abc123def456")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if task.Name != "my-task" {
		t.Errorf("expected name 'my-task', got %q", task.Name)
	}
}

func TestTaskService_GetByID_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Task{})
		},
	}))

	_, err := client.Tasks.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTaskService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Task{{Key: 3, Name: "daily-snapshot", Owner: "vms/123"}})
		},
		"GET /api/v4/tasks/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 3, Name: "daily-snapshot", Owner: "vms/123", Action: "snapshot"})
		},
	}))

	task, err := client.Tasks.GetByName(context.Background(), "vms/123", "daily-snapshot")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if task.Action != "snapshot" {
		t.Errorf("expected action 'snapshot', got %q", task.Action)
	}
}

func TestTaskService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Task{})
		},
	}))

	_, err := client.Tasks.GetByName(context.Background(), "vms/123", "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTaskService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tasks": func(w http.ResponseWriter, r *http.Request) {
			var req TaskCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "daily-snapshot" {
				t.Errorf("expected name 'daily-snapshot', got %q", req.Name)
			}
			if req.Owner != "vms/123" {
				t.Errorf("expected owner 'vms/123', got %q", req.Owner)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 1, Name: "daily-snapshot", Owner: "vms/123", Action: "snapshot"})
		},
	}))

	task, err := client.Tasks.Create(context.Background(), &TaskCreateRequest{
		Owner:  "vms/123",
		Action: "snapshot",
		Name:   "daily-snapshot",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if task.Name != "daily-snapshot" {
		t.Errorf("expected name 'daily-snapshot', got %q", task.Name)
	}
}

func TestTaskService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tasks.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTaskService_Create_MissingOwner(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tasks.Create(context.Background(), &TaskCreateRequest{
		Action: "snapshot",
		Name:   "test",
	})
	if err == nil {
		t.Fatal("expected error for missing owner")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTaskService_Create_MissingAction(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tasks.Create(context.Background(), &TaskCreateRequest{
		Owner: "vms/123",
		Name:  "test",
	})
	if err == nil {
		t.Fatal("expected error for missing action")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTaskService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tasks.Create(context.Background(), &TaskCreateRequest{
		Owner:  "vms/123",
		Action: "snapshot",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTaskService_Update(t *testing.T) {
	newName := "weekly-snapshot"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			var req TaskUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 1, Name: newName})
		},
	}))

	task, err := client.Tasks.Update(context.Background(), 1, &TaskUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if task.Name != newName {
		t.Errorf("expected name %q, got %q", newName, task.Name)
	}
}

func TestTaskService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tasks.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTaskService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tasks/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "updated"
	_, err := client.Tasks.Update(context.Background(), 999, &TaskUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTaskService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Tasks.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tasks/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Tasks.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTaskService_Execute(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "execute" {
				t.Errorf("expected action 'execute', got %v", body["action"])
			}
			// task ID should be present as a float64 from JSON
			if body["task"] != float64(1) {
				t.Errorf("expected task 1, got %v", body["task"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tasks.Execute(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestTaskService_Execute_WithParams(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/task_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			params, ok := body["params"].(map[string]any)
			if !ok {
				t.Error("expected params in request body")
			}
			if params["force"] != true {
				t.Errorf("expected force=true, got %v", params["force"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Tasks.Execute(context.Background(), 1, &TaskExecuteOptions{
		Params: map[string]any{"force": true},
	})
	if err != nil {
		t.Fatalf("Execute with params failed: %v", err)
	}
}

func TestTaskService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			var req TaskUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled=true")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 1, Enabled: true})
		},
	}))

	err := client.Tasks.Enable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestTaskService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			var req TaskUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled=false")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tasks/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Task{Key: 1, Enabled: false})
		},
	}))

	err := client.Tasks.Disable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}
