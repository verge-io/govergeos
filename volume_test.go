package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVolumeService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Volume{
				{Key: "abc123", ID: "abc123", Name: "vol1", Enabled: true},
				{Key: "def456", ID: "def456", Name: "vol2", Enabled: false},
			})
		},
	}))

	volumes, err := client.Volumes.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(volumes))
	}
	if volumes[0].Name != "vol1" {
		t.Errorf("expected name 'vol1', got %q", volumes[0].Name)
	}
}

func TestVolumeService_ListByService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "service eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Volume{{Key: "abc123", ID: "abc123", Name: "vol1"}})
		},
	}))

	volumes, err := client.Volumes.ListByService(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByService failed: %v", err)
	}
	if len(volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(volumes))
	}
}

func TestVolumeService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volumes/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Volume{Key: "abc123", ID: "abc123", Name: "vol1", FSType: "ext4"})
		},
	}))

	vol, err := client.Volumes.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if vol.Name != "vol1" {
		t.Errorf("expected name 'vol1', got %q", vol.Name)
	}
	if vol.FSType != "ext4" {
		t.Errorf("expected fs_type 'ext4', got %q", vol.FSType)
	}
}

func TestVolumeService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volumes/nonexistent": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Volumes.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Volume{{Key: "abc123", ID: "abc123", Name: "vol1"}})
		},
		"GET /api/v4/volumes/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Volume{Key: "abc123", ID: "abc123", Name: "vol1", FSType: "ext4"})
		},
	}))

	vol, err := client.Volumes.GetByName(context.Background(), 5, "vol1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if vol.Name != "vol1" {
		t.Errorf("expected name 'vol1', got %q", vol.Name)
	}
}

func TestVolumeService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Volume{})
		},
	}))

	_, err := client.Volumes.GetByName(context.Background(), 5, "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "newvol" {
				t.Errorf("expected name 'newvol', got %q", req.Name)
			}
			if req.Service != 5 {
				t.Errorf("expected service 5, got %d", req.Service)
			}
			jsonResponse(w, 200, apiResponse{Key: "sha1hash"})
		},
		"GET /api/v4/volumes/sha1hash": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Volume{Key: "sha1hash", ID: "sha1hash", Name: "newvol", Enabled: true})
		},
	}))

	vol, err := client.Volumes.Create(context.Background(), &VolumeCreateRequest{
		Name:    "newvol",
		Service: 5,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if vol.Name != "newvol" {
		t.Errorf("expected name 'newvol', got %q", vol.Name)
	}
	if vol.ID != "sha1hash" {
		t.Errorf("expected ID 'sha1hash', got %q", vol.ID)
	}
}

func TestVolumeService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Volumes.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Volumes.Create(context.Background(), &VolumeCreateRequest{Service: 5})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeService_Create_MissingService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Volumes.Create(context.Background(), &VolumeCreateRequest{Name: "vol1"})
	if err == nil {
		t.Fatal("expected error for missing service")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeService_Update(t *testing.T) {
	newDesc := "updated description"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volumes/abc123": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Description == nil || *req.Description != newDesc {
				t.Errorf("expected description %q, got %v", newDesc, req.Description)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/volumes/abc123": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Volume{Key: "abc123", ID: "abc123", Name: "vol1", Description: newDesc})
		},
	}))

	vol, err := client.Volumes.Update(context.Background(), "abc123", &VolumeUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if vol.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, vol.Description)
	}
}

func TestVolumeService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Volumes.Update(context.Background(), "abc123", nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeService_Update_NotFound(t *testing.T) {
	newDesc := "test"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volumes/nonexistent": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Volumes.Update(context.Background(), "nonexistent", &VolumeUpdateRequest{Description: &newDesc})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volumes/abc123": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Volumes.Delete(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVolumeService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volumes/nonexistent": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Volumes.Delete(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_actions": func(w http.ResponseWriter, r *http.Request) {
			var body volumeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Volume != "abc123" {
				t.Errorf("expected volume 'abc123', got %q", body.Volume)
			}
			if body.Action != "enable" {
				t.Errorf("expected action 'enable', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Volumes.Enable(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestVolumeService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_actions": func(w http.ResponseWriter, r *http.Request) {
			var body volumeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Volume != "abc123" {
				t.Errorf("expected volume 'abc123', got %q", body.Volume)
			}
			if body.Action != "disable" {
				t.Errorf("expected action 'disable', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Volumes.Disable(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

func TestVolumeService_Reset(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_actions": func(w http.ResponseWriter, r *http.Request) {
			var body volumeAction
			json.NewDecoder(r.Body).Decode(&body)
			if body.Volume != "abc123" {
				t.Errorf("expected volume 'abc123', got %q", body.Volume)
			}
			if body.Action != "reset" {
				t.Errorf("expected action 'reset', got %q", body.Action)
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Volumes.Reset(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}
