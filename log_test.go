package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestLogService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			sort := r.URL.Query().Get("sort")
			if sort != "-timestamp" {
				t.Errorf("expected default sort '-timestamp', got %q", sort)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "audit", Text: "user logged in", User: "admin", Timestamp: 1700000002},
				{Key: 2, Level: "error", Text: "disk failure", ObjectType: "cluster", Timestamp: 1700000001},
			})
		},
	}))

	logs, err := client.Logs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
	if logs[0].Level != "audit" {
		t.Errorf("expected level 'audit', got %q", logs[0].Level)
	}
}

func TestLogService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Log{})
		},
	}))

	logs, err := client.Logs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("expected 0 logs, got %d", len(logs))
	}
}

func TestLogService_ListByLevel(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "level eq 'warning'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "warning", Text: "disk usage high"},
			})
		},
	}))

	logs, err := client.Logs.ListByLevel(context.Background(), "warning")
	if err != nil {
		t.Fatalf("ListByLevel failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}

func TestLogService_ListByObjectType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "object_type eq 'vm'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "audit", ObjectType: "vm", ObjectName: "web-server-01"},
			})
		},
	}))

	logs, err := client.Logs.ListByObjectType(context.Background(), "vm")
	if err != nil {
		t.Fatalf("ListByObjectType failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].ObjectName != "web-server-01" {
		t.Errorf("expected ObjectName 'web-server-01', got %q", logs[0].ObjectName)
	}
}

func TestLogService_ListErrors(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "(level eq 'error') or (level eq 'critical')" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "error", Text: "disk failure"},
				{Key: 2, Level: "critical", Text: "node unreachable"},
			})
		},
	}))

	logs, err := client.Logs.ListErrors(context.Background())
	if err != nil {
		t.Fatalf("ListErrors failed: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
}

func TestLogService_ListAudit(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "level eq 'audit'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "audit", Text: "user created VM"},
			})
		},
	}))

	logs, err := client.Logs.ListAudit(context.Background())
	if err != nil {
		t.Fatalf("ListAudit failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}

func TestLogService_ListWarnings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "level eq 'warning'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "warning", Text: "high memory usage"},
			})
		},
	}))

	logs, err := client.Logs.ListWarnings(context.Background())
	if err != nil {
		t.Fatalf("ListWarnings failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}

func TestLogService_ListByUser(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "user eq 'admin'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "audit", User: "admin", Text: "changed settings"},
			})
		},
	}))

	logs, err := client.Logs.ListByUser(context.Background(), "admin")
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].User != "admin" {
		t.Errorf("expected user 'admin', got %q", logs[0].User)
	}
}

func TestLogService_ListSince(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "timestamp ge 1700000000" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Timestamp: 1700000001, Text: "recent event"},
			})
		},
	}))

	logs, err := client.Logs.ListSince(context.Background(), 1700000000)
	if err != nil {
		t.Fatalf("ListSince failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}

func TestLogService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs/42": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Log{
				Key: 42, Level: "error", Text: "disk failure detected", ObjectType: "cluster", ObjectName: "storage-01",
			})
		},
	}))

	log, err := client.Logs.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if log.Text != "disk failure detected" {
		t.Errorf("expected text 'disk failure detected', got %q", log.Text)
	}
	if log.ObjectType != "cluster" {
		t.Errorf("expected object_type 'cluster', got %q", log.ObjectType)
	}
}

func TestLogService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Logs.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestLogService_GetRecent(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			limit := r.URL.Query().Get("limit")
			if limit != "5" {
				t.Errorf("expected limit '5', got %q", limit)
			}
			jsonResponse(w, 200, []Log{
				{Key: 10, Text: "event 1"},
				{Key: 9, Text: "event 2"},
				{Key: 8, Text: "event 3"},
				{Key: 7, Text: "event 4"},
				{Key: 6, Text: "event 5"},
			})
		},
	}))

	logs, err := client.Logs.GetRecent(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	if len(logs) != 5 {
		t.Fatalf("expected 5 logs, got %d", len(logs))
	}
}

func TestLogService_GetRecentErrors(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "(level eq 'error') or (level eq 'critical')" {
				t.Errorf("unexpected filter: %s", filter)
			}
			limit := r.URL.Query().Get("limit")
			if limit != "3" {
				t.Errorf("expected limit '3', got %q", limit)
			}
			jsonResponse(w, 200, []Log{
				{Key: 10, Level: "error", Text: "error 1"},
				{Key: 9, Level: "critical", Text: "critical 1"},
				{Key: 8, Level: "error", Text: "error 2"},
			})
		},
	}))

	logs, err := client.Logs.GetRecentErrors(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetRecentErrors failed: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}
}

func TestLogService_Search(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "text ct 'disk failure'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Log{
				{Key: 1, Level: "error", Text: "disk failure detected on node2"},
				{Key: 2, Level: "warning", Text: "potential disk failure imminent"},
			})
		},
	}))

	logs, err := client.Logs.Search(context.Background(), "disk failure")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
}
