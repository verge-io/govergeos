package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// CloudSnapshotService
// ---------------------------------------------------------------------------

func TestCloudSnapshotService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudSnapshot{
				{Key: 1, Name: "snap-daily-20260101", Status: "normal"},
				{Key: 2, Name: "snap-daily-20260102", Status: "normal"},
			})
		},
	}))

	snaps, err := client.CloudSnapshots.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].Name != "snap-daily-20260101" {
		t.Errorf("expected name 'snap-daily-20260101', got %q", snaps[0].Name)
	}
}

func TestCloudSnapshotService_ListExpiring(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for expiring snapshots")
			}
			jsonResponse(w, 200, []CloudSnapshot{{Key: 1, Expires: 1735776000}})
		},
	}))

	snaps, err := client.CloudSnapshots.ListExpiring(context.Background())
	if err != nil {
		t.Fatalf("ListExpiring failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestCloudSnapshotService_ListLocal(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for local snapshots")
			}
			jsonResponse(w, 200, []CloudSnapshot{{Key: 1, Provider: false}})
		},
	}))

	snaps, err := client.CloudSnapshots.ListLocal(context.Background())
	if err != nil {
		t.Fatalf("ListLocal failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestCloudSnapshotService_ListByProfile(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for profile")
			}
			jsonResponse(w, 200, []CloudSnapshot{{Key: 1, SnapshotProfile: 5}})
		},
	}))

	snaps, err := client.CloudSnapshots.ListByProfile(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByProfile failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestCloudSnapshotService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudSnapshot{Key: 1, Name: "snap-daily-20260101", Status: "normal"})
		},
	}))

	snap, err := client.CloudSnapshots.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if snap.Name != "snap-daily-20260101" {
		t.Errorf("expected name 'snap-daily-20260101', got %q", snap.Name)
	}
}

func TestCloudSnapshotService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.CloudSnapshots.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudSnapshot{{Key: 1, Name: "snap-daily-20260101"}})
		},
		"GET /api/v4/cloud_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudSnapshot{Key: 1, Name: "snap-daily-20260101", Status: "normal"})
		},
	}))

	snap, err := client.CloudSnapshots.GetByName(context.Background(), "snap-daily-20260101")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if snap.Name != "snap-daily-20260101" {
		t.Errorf("expected name 'snap-daily-20260101', got %q", snap.Name)
	}
}

func TestCloudSnapshotService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudSnapshot{})
		},
	}))

	_, err := client.CloudSnapshots.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloud_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] != "manual-snap" {
				t.Errorf("expected name 'manual-snap', got %v", body["name"])
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/cloud_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudSnapshot{Key: 1, Name: "manual-snap", Status: "normal"})
		},
	}))

	snap, err := client.CloudSnapshots.Create(context.Background(), &CloudSnapshotCreateRequest{
		Name: "manual-snap",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if snap.Name != "manual-snap" {
		t.Errorf("expected name 'manual-snap', got %q", snap.Name)
	}
}

func TestCloudSnapshotService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudSnapshots.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudSnapshots.Create(context.Background(), &CloudSnapshotCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Update(t *testing.T) {
	newDesc := "updated description"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/cloud_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req CloudSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Description == nil || *req.Description != newDesc {
				t.Errorf("expected description %q, got %v", newDesc, req.Description)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/cloud_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudSnapshot{Key: 1, Name: "snap", Description: newDesc})
		},
	}))

	snap, err := client.CloudSnapshots.Update(context.Background(), 1, &CloudSnapshotUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if snap.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, snap.Description)
	}
}

func TestCloudSnapshotService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudSnapshots.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/cloud_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newDesc := "test"
	_, err := client.CloudSnapshots.Update(context.Background(), 999, &CloudSnapshotUpdateRequest{Description: &newDesc})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/cloud_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.CloudSnapshots.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestCloudSnapshotService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/cloud_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.CloudSnapshots.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Refresh(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloud_snapshot_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "refresh" {
				t.Errorf("expected action 'refresh', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.CloudSnapshots.Refresh(context.Background(), 1)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

func TestCloudSnapshotService_Clone(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloud_snapshot_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "clone" {
				t.Errorf("expected action 'clone', got %v", body["action"])
			}
			params, ok := body["params"].(map[string]any)
			if !ok {
				t.Fatal("expected params in body")
			}
			if params["name"] != "cloned-snap" {
				t.Errorf("expected name 'cloned-snap', got %v", params["name"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.CloudSnapshots.Clone(context.Background(), 1, &CloudSnapshotCloneOptions{Name: "cloned-snap"})
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
}

func TestCloudSnapshotService_Clone_NilOptions(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.CloudSnapshots.Clone(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil options")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_Clone_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.CloudSnapshots.Clone(context.Background(), 1, &CloudSnapshotCloneOptions{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudSnapshotService_RequestFromProvider(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloud_snapshot_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "request" {
				t.Errorf("expected action 'request', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.CloudSnapshots.RequestFromProvider(context.Background(), 1)
	if err != nil {
		t.Fatalf("RequestFromProvider failed: %v", err)
	}
}

func TestCloudSnapshotService_FindTenants(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloud_snapshot_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "find_tenants" {
				t.Errorf("expected action 'find_tenants', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.CloudSnapshots.FindTenants(context.Background(), 1)
	if err != nil {
		t.Fatalf("FindTenants failed: %v", err)
	}
}

func TestCloudSnapshotService_FindVMs(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloud_snapshot_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "find_vms" {
				t.Errorf("expected action 'find_vms', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.CloudSnapshots.FindVMs(context.Background(), 1)
	if err != nil {
		t.Fatalf("FindVMs failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CloudSnapshotVMService
// ---------------------------------------------------------------------------

func TestCloudSnapshotVMService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudSnapshotVM{
				{Key: 1, Name: "vm-web-01", CloudSnapshot: 10, CPUCores: 4, RAM: 8192},
				{Key: 2, Name: "vm-db-01", CloudSnapshot: 10, CPUCores: 8, RAM: 16384},
			})
		},
	}))

	vms, err := client.CloudSnapshotVMs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(vms))
	}
	if vms[0].Name != "vm-web-01" {
		t.Errorf("expected name 'vm-web-01', got %q", vms[0].Name)
	}
}

func TestCloudSnapshotVMService_ListBySnapshot(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_vms": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for snapshot")
			}
			jsonResponse(w, 200, []CloudSnapshotVM{{Key: 1, CloudSnapshot: 10, Name: "vm-web-01"}})
		},
	}))

	vms, err := client.CloudSnapshotVMs.ListBySnapshot(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListBySnapshot failed: %v", err)
	}
	if len(vms) != 1 {
		t.Fatalf("expected 1 VM, got %d", len(vms))
	}
}

func TestCloudSnapshotVMService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudSnapshotVM{Key: 1, Name: "vm-web-01", CPUCores: 4, RAM: 8192})
		},
	}))

	vm, err := client.CloudSnapshotVMs.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if vm.Name != "vm-web-01" {
		t.Errorf("expected name 'vm-web-01', got %q", vm.Name)
	}
	if vm.CPUCores != 4 {
		t.Errorf("expected 4 cpu cores, got %d", vm.CPUCores)
	}
}

func TestCloudSnapshotVMService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.CloudSnapshotVMs.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// CloudSnapshotTenantService
// ---------------------------------------------------------------------------

func TestCloudSnapshotTenantService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_tenants": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudSnapshotTenant{
				{Key: 1, Name: "tenant-prod", CloudSnapshot: 10},
				{Key: 2, Name: "tenant-dev", CloudSnapshot: 10},
			})
		},
	}))

	tenants, err := client.CloudSnapshotTenants.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tenants) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(tenants))
	}
	if tenants[0].Name != "tenant-prod" {
		t.Errorf("expected name 'tenant-prod', got %q", tenants[0].Name)
	}
}

func TestCloudSnapshotTenantService_ListBySnapshot(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_tenants": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for snapshot")
			}
			jsonResponse(w, 200, []CloudSnapshotTenant{{Key: 1, CloudSnapshot: 10, Name: "tenant-prod"}})
		},
	}))

	tenants, err := client.CloudSnapshotTenants.ListBySnapshot(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListBySnapshot failed: %v", err)
	}
	if len(tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d", len(tenants))
	}
}

func TestCloudSnapshotTenantService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_tenants/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudSnapshotTenant{Key: 1, Name: "tenant-prod", CloudSnapshot: 10})
		},
	}))

	tenant, err := client.CloudSnapshotTenants.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if tenant.Name != "tenant-prod" {
		t.Errorf("expected name 'tenant-prod', got %q", tenant.Name)
	}
}

func TestCloudSnapshotTenantService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloud_snapshot_tenants/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.CloudSnapshotTenants.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
