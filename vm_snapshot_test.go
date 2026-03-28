package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVMSnapshotService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMSnapshot{
				{Key: FlexInt(1), Name: "snap1", Machine: FlexInt(10)},
				{Key: FlexInt(2), Name: "snap2", Machine: FlexInt(20)},
			})
		},
	}))

	snaps, err := client.VMSnapshots.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[0].Name != "snap1" {
		t.Errorf("expected name 'snap1', got %q", snaps[0].Name)
	}
}

func TestVMSnapshotService_ListByVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VMSnapshot{
				{Key: FlexInt(1), Name: "snap1", Machine: FlexInt(42)},
			})
		},
	}))

	snaps, err := client.VMSnapshots.ListByVM(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListByVM failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestVMSnapshotService_ListByVM_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			// Should combine user filter with machine filter
			expected := "(name eq 'daily') and machine eq 42"
			if filter != expected {
				t.Errorf("expected filter %q, got %q", expected, filter)
			}
			jsonResponse(w, 200, []VMSnapshot{})
		},
	}))

	_, err := client.VMSnapshots.ListByVM(context.Background(), 42, WithFilter("name eq 'daily'"))
	if err != nil {
		t.Fatalf("ListByVM with filter failed: %v", err)
	}
}

func TestVMSnapshotService_ListExpiring(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			// 7 days = 7*86400 = 604800
			expected := "expires lt {$add({$now},604800)} and expires gt 0"
			if filter != expected {
				t.Errorf("expected filter %q, got %q", expected, filter)
			}
			jsonResponse(w, 200, []VMSnapshot{
				{Key: FlexInt(1), Name: "expiring-snap", Expires: 1234567890},
			})
		},
	}))

	snaps, err := client.VMSnapshots.ListExpiring(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListExpiring failed: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestVMSnapshotService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{
				Key:         FlexInt(1),
				Machine:     FlexInt(42),
				Name:        "snap1",
				ExpiresType: "never",
			})
		},
	}))

	snap, err := client.VMSnapshots.Get(context.Background(), 1)
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

func TestVMSnapshotService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMSnapshots.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			expected := "(name eq 'daily-backup') and machine eq 42"
			if filter != expected {
				t.Errorf("expected filter %q, got %q", expected, filter)
			}
			jsonResponse(w, 200, []VMSnapshot{
				{Key: FlexInt(1), Machine: FlexInt(42), Name: "daily-backup"},
			})
		},
	}))

	snap, err := client.VMSnapshots.GetByName(context.Background(), 42, "daily-backup")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if snap.Name != "daily-backup" {
		t.Errorf("expected name 'daily-backup', got %q", snap.Name)
	}
}

func TestVMSnapshotService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMSnapshot{})
		},
	}))

	_, err := client.VMSnapshots.GetByName(context.Background(), 42, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var req VMSnapshotCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Machine != 42 {
				t.Errorf("expected machine 42, got %d", req.Machine)
			}
			if req.Name != "test-snap" {
				t.Errorf("expected name 'test-snap', got %q", req.Name)
			}
			if req.ExpiresType != "date" {
				t.Errorf("expected default expires_type 'date', got %q", req.ExpiresType)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/machine_snapshots/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{
				Key:         FlexInt(5),
				Machine:     FlexInt(42),
				Name:        "test-snap",
				ExpiresType: "date",
			})
		},
	}))

	snap, err := client.VMSnapshots.Create(context.Background(), &VMSnapshotCreateRequest{
		Machine: 42,
		Name:    "test-snap",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if snap.Key.Int() != 5 {
		t.Errorf("expected key 5, got %d", snap.Key.Int())
	}
}

func TestVMSnapshotService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMSnapshots.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Create_MissingMachine(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMSnapshots.Create(context.Background(), &VMSnapshotCreateRequest{
		Name: "snap1",
	})
	if err == nil {
		t.Fatal("expected error for missing machine")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMSnapshots.Create(context.Background(), &VMSnapshotCreateRequest{
		Machine: 42,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Update(t *testing.T) {
	newName := "snap-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), Name: newName})
		},
	}))

	snap, err := client.VMSnapshots.Update(context.Background(), 1, &VMSnapshotUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if snap.Name != newName {
		t.Errorf("expected name %q, got %q", newName, snap.Name)
	}
}

func TestVMSnapshotService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMSnapshots.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "x"
	_, err := client.VMSnapshots.Update(context.Background(), 999, &VMSnapshotUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMSnapshots.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVMSnapshotService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/machine_snapshots/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VMSnapshots.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMSnapshotService_Restore(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), Machine: FlexInt(42), Name: "snap1"})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "restore" {
				t.Errorf("expected action 'restore', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 42 {
				t.Errorf("expected vm 42, got %v", body["vm"])
			}
			params := body["params"].(map[string]interface{})
			if int(params["snapshot"].(float64)) != 1 {
				t.Errorf("expected snapshot param 1, got %v", params["snapshot"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMSnapshots.Restore(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
}

func TestVMSnapshotService_Restore_WithPowerOn(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), Machine: FlexInt(42)})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			params := body["params"].(map[string]interface{})
			if params["poweron"] != true {
				t.Errorf("expected poweron true, got %v", params["poweron"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMSnapshots.Restore(context.Background(), 1, &VMSnapshotRestoreOptions{PowerOn: true})
	if err != nil {
		t.Fatalf("Restore with PowerOn failed: %v", err)
	}
}

func TestVMSnapshotService_SetNeverExpires(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.ExpiresType == nil || *req.ExpiresType != "never" {
				t.Errorf("expected expires_type 'never', got %v", req.ExpiresType)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), ExpiresType: "never"})
		},
	}))

	snap, err := client.VMSnapshots.SetNeverExpires(context.Background(), 1)
	if err != nil {
		t.Fatalf("SetNeverExpires failed: %v", err)
	}
	if snap.ExpiresType != "never" {
		t.Errorf("expected expires_type 'never', got %q", snap.ExpiresType)
	}
}

func TestVMSnapshotService_SetExpires(t *testing.T) {
	var expireTS int64 = 1700000000
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMSnapshotUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.ExpiresType == nil || *req.ExpiresType != "date" {
				t.Errorf("expected expires_type 'date', got %v", req.ExpiresType)
			}
			if req.Expires == nil || *req.Expires != expireTS {
				t.Errorf("expected expires %d, got %v", expireTS, req.Expires)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), ExpiresType: "date", Expires: expireTS})
		},
	}))

	snap, err := client.VMSnapshots.SetExpires(context.Background(), 1, expireTS)
	if err != nil {
		t.Fatalf("SetExpires failed: %v", err)
	}
	if snap.Expires != expireTS {
		t.Errorf("expected expires %d, got %d", expireTS, snap.Expires)
	}
}
