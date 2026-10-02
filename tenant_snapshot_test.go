package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTenantSnapshotService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantSnapshot{
				{Key: 1, Tenant: 10, Name: "snap1"},
				{Key: 2, Tenant: 10, Name: "snap2"},
			})
		},
	}))

	snapshots, err := client.TenantSnapshots.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].Name != "snap1" {
		t.Errorf("expected name 'snap1', got %q", snapshots[0].Name)
	}
}

func TestTenantSnapshotService_ListByTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "tenant eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantSnapshot{{Key: 1, Tenant: 10, Name: "snap1"}})
		},
	}))

	snapshots, err := client.TenantSnapshots.ListByTenant(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
}

func TestTenantSnapshotService_ListExpiring(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			// 7 days = 604800 seconds
			expected := "expires ne 0 and expires lt {$add({$now},604800)}"
			if filter != expected {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantSnapshot{{Key: 1, Expires: 1700000000}})
		},
	}))

	snapshots, err := client.TenantSnapshots.ListExpiring(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListExpiring failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
}

func TestTenantSnapshotService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{Key: 1, Tenant: 10, Name: "snap1", Description: "test snapshot"})
		},
	}))

	snapshot, err := client.TenantSnapshots.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if snapshot.Name != "snap1" {
		t.Errorf("expected name 'snap1', got %q", snapshot.Name)
	}
	if snapshot.Description != "test snapshot" {
		t.Errorf("expected description 'test snapshot', got %q", snapshot.Description)
	}
}

func TestTenantSnapshotService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TenantSnapshots.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantSnapshotService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantSnapshot{{Key: 1, Tenant: 10, Name: "snap1"}})
		},
		"GET /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{Key: 1, Tenant: 10, Name: "snap1"})
		},
	}))

	snapshot, err := client.TenantSnapshots.GetByName(context.Background(), 10, "snap1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if snapshot.Name != "snap1" {
		t.Errorf("expected name 'snap1', got %q", snapshot.Name)
	}
}

func TestTenantSnapshotService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantSnapshot{})
		},
	}))

	_, err := client.TenantSnapshots.GetByName(context.Background(), 10, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantSnapshotService_Update(t *testing.T) {
	newDesc := "updated description"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Description == nil || *req.Description != newDesc {
				t.Errorf("expected description %q, got %v", newDesc, req.Description)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{Key: 1, Name: "snap1", Description: newDesc})
		},
	}))

	snapshot, err := client.TenantSnapshots.Update(context.Background(), 1, &TenantSnapshotUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if snapshot.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, snapshot.Description)
	}
}

func TestTenantSnapshotService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantSnapshots.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantSnapshotService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TenantSnapshots.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTenantSnapshotService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tenant_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.TenantSnapshots.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTenantSnapshotService_Refresh(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_snapshot_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "refresh" {
				t.Errorf("expected action 'refresh', got %v", body["action"])
			}
			if int(body["tenant"].(float64)) != 10 {
				t.Errorf("expected tenant 10, got %v", body["tenant"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.TenantSnapshots.Refresh(context.Background(), 10)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

func TestTenantSnapshotService_SetNeverExpires(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Expires == nil || *req.Expires != 0 {
				t.Errorf("expected expires 0, got %v", req.Expires)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{Key: 1, Expires: 0})
		},
	}))

	snapshot, err := client.TenantSnapshots.SetNeverExpires(context.Background(), 1)
	if err != nil {
		t.Fatalf("SetNeverExpires failed: %v", err)
	}
	if snapshot.Expires != 0 {
		t.Errorf("expected expires 0, got %d", snapshot.Expires)
	}
}

func TestTenantSnapshotService_SetExpires(t *testing.T) {
	expiry := int64(1700000000)
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req TenantSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Expires == nil || *req.Expires != expiry {
				t.Errorf("expected expires %d, got %v", expiry, req.Expires)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tenant_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{Key: 1, Expires: expiry})
		},
	}))

	snapshot, err := client.TenantSnapshots.SetExpires(context.Background(), 1, expiry)
	if err != nil {
		t.Fatalf("SetExpires failed: %v", err)
	}
	if snapshot.Expires != expiry {
		t.Errorf("expected expires %d, got %d", expiry, snapshot.Expires)
	}
}

func TestTenantSnapshotService_Create(t *testing.T) {
	minSnaps := 2
	expires := int64(1700000000)
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var req TenantSnapshotCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if req.Tenant != 10 {
				t.Errorf("expected tenant 10, got %d", req.Tenant)
			}
			if req.Name != "pre-upgrade" {
				t.Errorf("expected name 'pre-upgrade', got %q", req.Name)
			}
			if req.Description != "before upgrade" {
				t.Errorf("expected description 'before upgrade', got %q", req.Description)
			}
			if req.Profile != "daily" {
				t.Errorf("expected profile 'daily', got %q", req.Profile)
			}
			if req.Period != "nightly" {
				t.Errorf("expected period 'nightly', got %q", req.Period)
			}
			if req.MinSnapshots == nil || *req.MinSnapshots != minSnaps {
				t.Errorf("expected min_snapshots %d, got %v", minSnaps, req.MinSnapshots)
			}
			if req.Expires == nil || *req.Expires != expires {
				t.Errorf("expected expires %d, got %v", expires, req.Expires)
			}
			if req.Type != TenantSnapshotTypeFull {
				t.Errorf("expected type %q, got %q", TenantSnapshotTypeFull, req.Type)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/tenant_snapshots/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{
				Key:          5,
				Tenant:       10,
				Name:         "pre-upgrade",
				Description:  "before upgrade",
				Profile:      "daily",
				Period:       "nightly",
				MinSnapshots: minSnaps,
				Expires:      expires,
				Type:         TenantSnapshotTypeFull,
			})
		},
	}))

	snap, err := client.TenantSnapshots.Create(context.Background(), &TenantSnapshotCreateRequest{
		Tenant:       10,
		Name:         "pre-upgrade",
		Description:  "before upgrade",
		Profile:      "daily",
		Period:       "nightly",
		MinSnapshots: &minSnaps,
		Expires:      &expires,
		Type:         TenantSnapshotTypeFull,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if snap.Key.Int() != 5 {
		t.Errorf("expected key 5, got %d", snap.Key.Int())
	}
	if snap.Name != "pre-upgrade" {
		t.Errorf("expected name 'pre-upgrade', got %q", snap.Name)
	}
	if snap.Type != TenantSnapshotTypeFull {
		t.Errorf("expected type %q, got %q", TenantSnapshotTypeFull, snap.Type)
	}
}

func TestTenantSnapshotService_Create_WithoutName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tenant_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var req TenantSnapshotCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if req.Tenant != 10 {
				t.Errorf("expected tenant 10, got %d", req.Tenant)
			}
			if req.Name != "" {
				t.Errorf("expected empty name, got %q", req.Name)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(6)})
		},
		"GET /api/v4/tenant_snapshots/6": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantSnapshot{Key: 6, Tenant: 10, Name: "auto-assigned", Type: TenantSnapshotTypeFull})
		},
	}))

	snap, err := client.TenantSnapshots.Create(context.Background(), &TenantSnapshotCreateRequest{
		Tenant: 10,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if snap.Name != "auto-assigned" {
		t.Errorf("expected server-assigned name, got %q", snap.Name)
	}
}

func TestTenantSnapshotService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantSnapshots.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTenantSnapshotService_Create_MissingTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TenantSnapshots.Create(context.Background(), &TenantSnapshotCreateRequest{
		Name: "snap1",
	})
	if err == nil {
		t.Fatal("expected error for missing tenant")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}
