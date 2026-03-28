package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVolumeBrowserService_CreateJob(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_browser": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeBrowserRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Volume == "" {
				t.Error("expected volume in request")
			}
			if req.Query != VolumeBrowserQueryGetDir {
				t.Errorf("expected query 'get-dir', got %q", req.Query)
			}
			// Return the job ID as a string key
			jsonResponse(w, 200, map[string]interface{}{
				"$key": "abc123sha1hash",
			})
		},
	}))

	job, err := client.VolumeBrowser.CreateJob(context.Background(), &VolumeBrowserRequest{
		Volume: "vol-sha1",
		Query:  VolumeBrowserQueryGetDir,
		Params: &VolumeBrowserParams{
			Dir:   "",
			Limit: 100,
		},
	})
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if job.Status != VolumeBrowserStatusRunning {
		t.Errorf("expected status 'running', got %q", job.Status)
	}
	if job.Volume != "vol-sha1" {
		t.Errorf("expected volume 'vol-sha1', got %q", job.Volume)
	}
}

func TestVolumeBrowserService_CreateJob_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeBrowser.CreateJob(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeBrowserService_CreateJob_EmptyVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeBrowser.CreateJob(context.Background(), &VolumeBrowserRequest{
		Query:  VolumeBrowserQueryGetDir,
		Params: &VolumeBrowserParams{},
	})
	if err == nil {
		t.Fatal("expected error for empty volume")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeBrowserService_CreateJob_EmptyQuery(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeBrowser.CreateJob(context.Background(), &VolumeBrowserRequest{
		Volume: "vol-sha1",
		Params: &VolumeBrowserParams{},
	})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeBrowserService_CreateJob_NilParams(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeBrowser.CreateJob(context.Background(), &VolumeBrowserRequest{
		Volume: "vol-sha1",
		Query:  VolumeBrowserQueryGetDir,
	})
	if err == nil {
		t.Fatal("expected error for nil params")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeBrowserService_GetJob(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/abc123": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != volumeBrowserGetFields {
				t.Errorf("expected get fields, got %q", fields)
			}
			jsonResponse(w, 200, VolumeBrowserJob{
				Key:    "abc123",
				ID:     "abc123",
				Volume: "vol-sha1",
				Status: VolumeBrowserStatusComplete,
				Result: json.RawMessage(`[{"name":"file1.txt","size":1024,"date":1700000000,"type":"file"}]`),
			})
		},
	}))

	job, err := client.VolumeBrowser.GetJob(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.Status != VolumeBrowserStatusComplete {
		t.Errorf("expected status 'complete', got %q", job.Status)
	}
	if job.Volume != "vol-sha1" {
		t.Errorf("expected volume 'vol-sha1', got %q", job.Volume)
	}
}

func TestVolumeBrowserService_GetJob_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/nonexistent": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeBrowser.GetJob(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeBrowserService_WaitForResult_Complete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/job1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeBrowserJob{
				Key:    "job1",
				ID:     "job1",
				Status: VolumeBrowserStatusComplete,
				Result: json.RawMessage(`[{"name":"docs","size":0,"date":1700000000,"type":"directory"},{"name":"readme.txt","size":512,"date":1700000001,"type":"file"}]`),
			})
		},
	}))

	entries, err := client.VolumeBrowser.WaitForResult(context.Background(), "job1", 5*1e9)
	if err != nil {
		t.Fatalf("WaitForResult failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "docs" {
		t.Errorf("expected name 'docs', got %q", entries[0].Name)
	}
	if entries[0].Type != "directory" {
		t.Errorf("expected type 'directory', got %q", entries[0].Type)
	}
	if entries[1].Size != 512 {
		t.Errorf("expected size 512, got %d", entries[1].Size)
	}
}

func TestVolumeBrowserService_WaitForResult_EmptyDir(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/job2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeBrowserJob{
				Key:    "job2",
				ID:     "job2",
				Status: VolumeBrowserStatusComplete,
				Result: json.RawMessage(`null`),
			})
		},
	}))

	entries, err := client.VolumeBrowser.WaitForResult(context.Background(), "job2", 5*1e9)
	if err != nil {
		t.Fatalf("WaitForResult failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for empty dir, got %d", len(entries))
	}
}

func TestVolumeBrowserService_WaitForResult_Error(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/joberr": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeBrowserJob{
				Key:    "joberr",
				ID:     "joberr",
				Status: VolumeBrowserStatusError,
				Result: json.RawMessage(`"NAS service not running"`),
			})
		},
	}))

	_, err := client.VolumeBrowser.WaitForResult(context.Background(), "joberr", 5*1e9)
	if err == nil {
		t.Fatal("expected error for error status")
	}
}

func TestVolumeBrowserService_WaitForResult_ContextCanceled(t *testing.T) {
	callCount := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/jobslow": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			jsonResponse(w, 200, VolumeBrowserJob{
				Key:    "jobslow",
				ID:     "jobslow",
				Status: VolumeBrowserStatusRunning,
			})
		},
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.VolumeBrowser.WaitForResult(ctx, "jobslow", 30*1e9)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestVolumeBrowserService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeBrowserJob{
				{Key: "j1", ID: "j1", Status: VolumeBrowserStatusComplete},
				{Key: "j2", ID: "j2", Status: VolumeBrowserStatusRunning},
			})
		},
	}))

	jobs, err := client.VolumeBrowser.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestVolumeBrowserService_ListByVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "volume eq 'vol-abc'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeBrowserJob{{Key: "j1", Volume: "vol-abc"}})
		},
	}))

	jobs, err := client.VolumeBrowser.ListByVolume(context.Background(), "vol-abc")
	if err != nil {
		t.Fatalf("ListByVolume failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

func TestVolumeBrowserService_WaitForResult_UnexpectedStatus(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_browser/jobweird": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeBrowserJob{
				Key:    "jobweird",
				ID:     "jobweird",
				Status: "unknown_status",
			})
		},
	}))

	_, err := client.VolumeBrowser.WaitForResult(context.Background(), "jobweird", 5*1e9)
	if err == nil {
		t.Fatal("expected error for unexpected status")
	}
}
