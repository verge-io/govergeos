package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVolumeSnapshotService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeSnapshot{
				{Key: 1, Name: "snap-daily", Enabled: true},
				{Key: 2, Name: "snap-manual", Enabled: false},
			})
		},
	}))

	snaps, err := client.VolumeSnapshots.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].Name != "snap-daily" {
		t.Errorf("expected name 'snap-daily', got %q", snaps[0].Name)
	}
}

func TestVolumeSnapshotService_ListByVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "volume eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeSnapshot{{Key: 1, Name: "snap1"}})
		},
	}))

	snaps, err := client.VolumeSnapshots.ListByVolume(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListByVolume failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestVolumeSnapshotService_ListExpiring(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for expiring snapshots")
			}
			jsonResponse(w, 200, []VolumeSnapshot{{Key: 1, Name: "expiring-snap", ExpiresType: "date"}})
		},
	}))

	snaps, err := client.VolumeSnapshots.ListExpiring(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListExpiring failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestVolumeSnapshotService_ListManual(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "created_manually eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeSnapshot{{Key: 1, Name: "manual-snap", CreatedManually: true}})
		},
	}))

	snaps, err := client.VolumeSnapshots.ListManual(context.Background())
	if err != nil {
		t.Fatalf("ListManual failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestVolumeSnapshotService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1", ExpiresType: "never"})
		},
	}))

	snap, err := client.VolumeSnapshots.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if snap.Name != "snap1" {
		t.Errorf("expected name 'snap1', got %q", snap.Name)
	}
	if snap.ExpiresType != "never" {
		t.Errorf("expected expires_type 'never', got %q", snap.ExpiresType)
	}
}

func TestVolumeSnapshotService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeSnapshots.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeSnapshot{{Key: 1, Name: "snap1"}})
		},
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1"})
		},
	}))

	snap, err := client.VolumeSnapshots.GetByName(context.Background(), 42, "snap1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if snap.Name != "snap1" {
		t.Errorf("expected name 'snap1', got %q", snap.Name)
	}
}

func TestVolumeSnapshotService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeSnapshot{})
		},
	}))

	_, err := client.VolumeSnapshots.GetByName(context.Background(), 42, "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSnapshotCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "new-snap" {
				t.Errorf("expected name 'new-snap', got %q", req.Name)
			}
			if req.Volume != 42 {
				t.Errorf("expected volume 42, got %d", req.Volume)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(10)})
		},
		"GET /api/v4/volume_snapshots/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 10, Name: "new-snap", Enabled: true})
		},
	}))

	snap, err := client.VolumeSnapshots.Create(context.Background(), &VolumeSnapshotCreateRequest{
		Volume: 42,
		Name:   "new-snap",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if snap.Name != "new-snap" {
		t.Errorf("expected name 'new-snap', got %q", snap.Name)
	}
}

func TestVolumeSnapshotService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSnapshots.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Create_MissingVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSnapshots.Create(context.Background(), &VolumeSnapshotCreateRequest{Name: "snap1"})
	if err == nil {
		t.Fatal("expected error for missing volume")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSnapshots.Create(context.Background(), &VolumeSnapshotCreateRequest{Volume: 42})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Update(t *testing.T) {
	newDesc := "updated snapshot"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1", Description: newDesc})
		},
	}))

	snap, err := client.VolumeSnapshots.Update(context.Background(), 1, &VolumeSnapshotUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if snap.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, snap.Description)
	}
}

func TestVolumeSnapshotService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeSnapshots.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Update_NotFound(t *testing.T) {
	name := "test"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeSnapshots.Update(context.Background(), 999, &VolumeSnapshotUpdateRequest{Name: &name})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VolumeSnapshots.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVolumeSnapshotService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VolumeSnapshots.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeSnapshotService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled=true in update request")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1", Enabled: true})
		},
	}))

	err := client.VolumeSnapshots.Enable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestVolumeSnapshotService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled=false in update request")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1", Enabled: false})
		},
	}))

	err := client.VolumeSnapshots.Disable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

func TestVolumeSnapshotService_SetNeverExpires(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.ExpiresType == nil || *req.ExpiresType != "never" {
				t.Errorf("expected expires_type 'never', got %v", req.ExpiresType)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1", ExpiresType: "never"})
		},
	}))

	snap, err := client.VolumeSnapshots.SetNeverExpires(context.Background(), 1)
	if err != nil {
		t.Fatalf("SetNeverExpires failed: %v", err)
	}
	if snap.ExpiresType != "never" {
		t.Errorf("expected expires_type 'never', got %q", snap.ExpiresType)
	}
}

func TestVolumeSnapshotService_SetExpires(t *testing.T) {
	expiresAt := int64(1735689600)
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.ExpiresType == nil || *req.ExpiresType != "date" {
				t.Errorf("expected expires_type 'date', got %v", req.ExpiresType)
			}
			if req.Expires == nil || *req.Expires != expiresAt {
				t.Errorf("expected expires %d, got %v", expiresAt, req.Expires)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeSnapshot{Key: 1, Name: "snap1", ExpiresType: "date", Expires: expiresAt})
		},
	}))

	snap, err := client.VolumeSnapshots.SetExpires(context.Background(), 1, expiresAt)
	if err != nil {
		t.Fatalf("SetExpires failed: %v", err)
	}
	if snap.Expires != expiresAt {
		t.Errorf("expected expires %d, got %d", expiresAt, snap.Expires)
	}
}
