package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVolumeSyncService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeSync{
				{Key: "sync1", ID: "sync1", Name: "daily-backup", Enabled: true},
				{Key: "sync2", ID: "sync2", Name: "hourly-mirror", Enabled: false},
			})
		},
	}))

	syncs, err := client.VolumeSyncs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(syncs) != 2 {
		t.Fatalf("expected 2 syncs, got %d", len(syncs))
	}
	if syncs[0].Name != "daily-backup" {
		t.Errorf("expected name 'daily-backup', got %q", syncs[0].Name)
	}
}

func TestVolumeSyncService_ListByService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "service eq 3" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeSync{{Key: "sync1", ID: "sync1", Name: "daily-backup"}})
		},
	}))

	syncs, err := client.VolumeSyncs.ListByService(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListByService failed: %v", err)
	}
	if len(syncs) != 1 {
		t.Fatalf("expected 1 sync, got %d", len(syncs))
	}
}

func TestVolumeSyncService_ListEnabled(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "enabled eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeSync{{Key: "sync1", ID: "sync1", Name: "daily-backup", Enabled: true}})
		},
	}))

	syncs, err := client.VolumeSyncs.ListEnabled(context.Background())
	if err != nil {
		t.Fatalf("ListEnabled failed: %v", err)
	}
	if len(syncs) != 1 {
		t.Fatalf("expected 1 sync, got %d", len(syncs))
	}
}

func TestVolumeSyncService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSync{Key: "sync1", ID: "sync1", Name: "daily-backup", SyncMethod: "rsync"})
		},
	}))

	sync, err := client.VolumeSyncs.Get(context.Background(), "sync1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sync.Name != "daily-backup" {
		t.Errorf("expected name 'daily-backup', got %q", sync.Name)
	}
	if sync.SyncMethod != "rsync" {
		t.Errorf("expected sync_method 'rsync', got %q", sync.SyncMethod)
	}
}

func TestVolumeSyncService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeSyncs.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeSync{{Key: "sync1", ID: "sync1", Name: "daily-backup"}})
		},
		"GET /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSync{Key: "sync1", ID: "sync1", Name: "daily-backup"})
		},
	}))

	sync, err := client.VolumeSyncs.GetByName(context.Background(), 3, "daily-backup")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if sync.Name != "daily-backup" {
		t.Errorf("expected name 'daily-backup', got %q", sync.Name)
	}
}

func TestVolumeSyncService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_syncs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeSync{})
		},
	}))

	_, err := client.VolumeSyncs.GetByName(context.Background(), 3, "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_syncs": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSyncCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "daily-backup" {
				t.Errorf("expected name 'daily-backup', got %q", req.Name)
			}
			if req.Service != 3 {
				t.Errorf("expected service 3, got %d", req.Service)
			}
			if req.SourceVolume != 10 {
				t.Errorf("expected source_volume 10, got %d", req.SourceVolume)
			}
			if req.DestinationVolume != 20 {
				t.Errorf("expected destination_volume 20, got %d", req.DestinationVolume)
			}
			jsonResponse(w, 200, apiResponse{Key: "newsync"})
		},
		"GET /api/v4/volume_syncs/newsync": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSync{Key: "newsync", ID: "newsync", Name: "daily-backup", Enabled: true})
		},
	}))

	sync, err := client.VolumeSyncs.Create(context.Background(), &VolumeSyncCreateRequest{
		Service:           3,
		Name:              "daily-backup",
		SourceVolume:      10,
		DestinationVolume: 20,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sync.Name != "daily-backup" {
		t.Errorf("expected name 'daily-backup', got %q", sync.Name)
	}
}

func TestVolumeSyncService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSyncs.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Create_MissingService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSyncs.Create(context.Background(), &VolumeSyncCreateRequest{
		Name:              "sync1",
		SourceVolume:      10,
		DestinationVolume: 20,
	})
	if err == nil {
		t.Fatal("expected error for missing service")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSyncs.Create(context.Background(), &VolumeSyncCreateRequest{
		Service:           3,
		SourceVolume:      10,
		DestinationVolume: 20,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Create_MissingSourceVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSyncs.Create(context.Background(), &VolumeSyncCreateRequest{
		Service:           3,
		Name:              "sync1",
		DestinationVolume: 20,
	})
	if err == nil {
		t.Fatal("expected error for missing source_volume")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Create_MissingDestinationVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSyncs.Create(context.Background(), &VolumeSyncCreateRequest{
		Service:      3,
		Name:         "sync1",
		SourceVolume: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing destination_volume")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Update(t *testing.T) {
	newDesc := "updated sync"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSync{Key: "sync1", ID: "sync1", Name: "daily-backup", Description: newDesc})
		},
	}))

	sync, err := client.VolumeSyncs.Update(context.Background(), "sync1", &VolumeSyncUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if sync.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, sync.Description)
	}
}

func TestVolumeSyncService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSyncs.Update(context.Background(), "sync1", nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Update_NotFound(t *testing.T) {
	name := "test"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_syncs/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeSyncs.Update(context.Background(), "missing", &VolumeSyncUpdateRequest{Name: &name})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VolumeSyncs.Delete(context.Background(), "sync1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVolumeSyncService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_syncs/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VolumeSyncs.Delete(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSyncService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSyncUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled=true in update request")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSync{Key: "sync1", ID: "sync1", Name: "daily-backup", Enabled: true})
		},
	}))

	err := client.VolumeSyncs.Enable(context.Background(), "sync1")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestVolumeSyncService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSyncUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled=false in update request")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_syncs/sync1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSync{Key: "sync1", ID: "sync1", Name: "daily-backup", Enabled: false})
		},
	}))

	err := client.VolumeSyncs.Disable(context.Background(), "sync1")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

func TestVolumeSyncService_Start(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_sync_actions": func(w http.ResponseWriter, r *http.Request) {
			var body volumeSyncAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Sync != "sync1" {
				t.Errorf("expected sync 'sync1', got %q", body.Sync)
			}
			if body.Action != "start" {
				t.Errorf("expected action 'start', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VolumeSyncs.Start(context.Background(), "sync1")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

func TestVolumeSyncService_Stop(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_sync_actions": func(w http.ResponseWriter, r *http.Request) {
			var body volumeSyncAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Sync != "sync1" {
				t.Errorf("expected sync 'sync1', got %q", body.Sync)
			}
			if body.Action != "stop" {
				t.Errorf("expected action 'stop', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VolumeSyncs.Stop(context.Background(), "sync1")
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}
