package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestVMDriveService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/42": stubVM(42, 42),
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "ms_2023_kek_applied") {
				t.Errorf("fields missing ms_2023_kek_applied: %s", fields)
			}
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VMDrive{
				{Key: FlexInt(1), Machine: 42, Name: "disk0", SizeBytes: 10 * bytesPerGB, MS2023KEKApplied: true},
				{Key: FlexInt(2), Machine: 42, Name: "disk1", SizeBytes: 20 * bytesPerGB},
			})
		},
	}))
	client.serverVersion = "26.1.5"

	drives, err := client.VMDrives.List(context.Background(), 42)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(drives) != 2 {
		t.Fatalf("expected 2 drives, got %d", len(drives))
	}
	if drives[0].SizeGB != 10 {
		t.Errorf("expected SizeGB 10, got %d", drives[0].SizeGB)
	}
	if !drives[0].MS2023KEKApplied {
		t.Error("expected MS2023KEKApplied to be true")
	}
	if drives[1].SizeGB != 20 {
		t.Errorf("expected SizeGB 20, got %d", drives[1].SizeGB)
	}
}

func TestVMDriveService_FieldsByServerVersion(t *testing.T) {
	tests := []struct {
		version   string
		wantField bool
	}{
		{version: "26.0.2.2", wantField: false},
		{version: "26.1.4", wantField: false},
		{version: "26.1.5", wantField: true},
		{version: "26.1.8", wantField: true},
		{version: "26.2.0", wantField: true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			service := &VMDriveService{client: &Client{serverVersion: tt.version}}
			for name, fields := range map[string]string{
				"list": service.listFields(),
				"get":  service.getFields(),
			} {
				if got := strings.Contains(fields, driveMS2023KEKField); got != tt.wantField {
					t.Errorf("%s fields contain %q = %v, want %v", name, driveMS2023KEKField, got, tt.wantField)
				}
			}
		})
	}
}

func TestVMDriveService_ApplyUniversalVars(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/machine_drives/7/apply_universal_vars": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("expected empty request body, got %#v", body)
			}
			jsonResponse(w, http.StatusOK, map[string]any{})
		},
	}))

	if err := client.VMDrives.ApplyUniversalVars(context.Background(), 7); err != nil {
		t.Fatalf("ApplyUniversalVars failed: %v", err)
	}
}

func TestVMDriveService_ApplyUniversalVars_Error(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/machine_drives/7/apply_universal_vars": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, http.StatusForbidden, map[string]string{"err": "machine must be offline"})
		},
	}))

	err := client.VMDrives.ApplyUniversalVars(context.Background(), 7)
	if err == nil {
		t.Fatal("expected ApplyUniversalVars to return an error")
	}
	if !IsPermissionError(err) {
		t.Fatalf("expected 403 PermissionError, got %T: %v", err, err)
	}
	if IsAuthError(err) {
		t.Fatalf("403 reported as authentication failure: %v", err)
	}
}

func TestVMDriveService_ListAll(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "" {
				t.Errorf("expected no filter for ListAll, got %q", filter)
			}
			jsonResponse(w, 200, []VMDrive{
				{Key: FlexInt(1), Machine: 42, Name: "disk0", SizeBytes: 10 * bytesPerGB},
				{Key: FlexInt(2), Machine: 99, Name: "disk1", SizeBytes: 20 * bytesPerGB},
			})
		},
	}))

	drives, err := client.VMDrives.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(drives) != 2 {
		t.Fatalf("expected 2 drives, got %d", len(drives))
	}
	// Verify drives span multiple VMs
	if drives[0].Machine == drives[1].Machine {
		t.Error("expected drives from different VMs")
	}
	if drives[0].SizeGB != 10 {
		t.Errorf("expected SizeGB 10, got %d", drives[0].SizeGB)
	}
}

func TestVMDriveService_ListAll_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "status eq 'online'" {
				t.Errorf("expected status filter, got %q", filter)
			}
			jsonResponse(w, 200, []VMDrive{})
		},
	}))

	_, err := client.VMDrives.ListAll(context.Background(), WithFilter("status eq 'online'"))
	if err != nil {
		t.Fatalf("ListAll with filter failed: %v", err)
	}
}

func TestVMDriveService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{
				Key:       FlexInt(1),
				Machine:   42,
				Name:      "disk0",
				SizeBytes: 50 * bytesPerGB,
				Interface: "virtio-scsi",
			})
		},
	}))

	drive, err := client.VMDrives.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if drive.SizeGB != 50 {
		t.Errorf("expected SizeGB 50, got %d", drive.SizeGB)
	}
	if drive.Interface != "virtio-scsi" {
		t.Errorf("expected interface 'virtio-scsi', got %q", drive.Interface)
	}
}

func TestVMDriveService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/42": stubVM(42, 42),
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			expected := "machine eq 42 and name eq 'disk0'"
			if filter != expected {
				t.Errorf("expected filter %q, got %q", expected, filter)
			}
			jsonResponse(w, 200, []VMDrive{
				{Key: FlexInt(7), Machine: 42, Name: "disk0", SizeBytes: 10 * bytesPerGB},
			})
		},
		"GET /api/v4/machine_drives/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{
				Key:       FlexInt(7),
				Machine:   42,
				Name:      "disk0",
				SizeBytes: 10 * bytesPerGB,
				Interface: "virtio-scsi",
			})
		},
	}))

	drive, err := client.VMDrives.GetByName(context.Background(), 42, "disk0")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if drive.Name != "disk0" {
		t.Errorf("expected name 'disk0', got %q", drive.Name)
	}
	if drive.Key.Int() != 7 {
		t.Errorf("expected ID 7, got %d", drive.Key.Int())
	}
	if drive.SizeGB != 10 {
		t.Errorf("expected SizeGB 10, got %d", drive.SizeGB)
	}
}

func TestVMDriveService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/42": stubVM(42, 42),
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMDrive{})
		},
	}))

	_, err := client.VMDrives.GetByName(context.Background(), 42, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMDriveService_GetByName_EscapesName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/42": stubVM(42, 42),
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			expected := "machine eq 42 and name eq 'o\\'malley'"
			if filter != expected {
				t.Errorf("expected filter %q, got %q", expected, filter)
			}
			jsonResponse(w, 200, []VMDrive{})
		},
	}))

	_, err := client.VMDrives.GetByName(context.Background(), 42, "o'malley")
	if err == nil {
		t.Fatal("expected NotFoundError")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMDriveService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMDrives.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMDriveService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/10": stubVM(10, 10),
		"POST /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			var req VMDriveCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Machine != 10 {
				t.Errorf("expected machine 10, got %d", req.Machine)
			}
			if req.Name != "disk0" {
				t.Errorf("expected name 'disk0', got %q", req.Name)
			}
			if req.SizeBytes != 25*bytesPerGB {
				t.Errorf("expected SizeBytes %d, got %d", 25*bytesPerGB, req.SizeBytes)
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: float64(7)})
		},
		"GET /api/v4/machine_drives/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{
				Key:       FlexInt(7),
				Machine:   10,
				Name:      "disk0",
				SizeBytes: 25 * bytesPerGB,
				Enabled:   true,
			})
		},
	}))

	drive, err := client.VMDrives.Create(context.Background(), 10, &VMDriveCreateRequest{
		Name:   "disk0",
		SizeGB: 25,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if drive.Key.Int() != 7 {
		t.Errorf("expected ID 7, got %d", drive.Key.Int())
	}
	if drive.SizeGB != 25 {
		t.Errorf("expected SizeGB 25, got %d", drive.SizeGB)
	}
}

func TestVMDriveService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDrives.Create(context.Background(), 10, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDriveService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDrives.Create(context.Background(), 10, &VMDriveCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDriveService_Update(t *testing.T) {
	newName := "disk-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMDriveUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{
				Key:       FlexInt(1),
				Name:      newName,
				SizeBytes: 10 * bytesPerGB,
			})
		},
	}))

	drive, err := client.VMDrives.Update(context.Background(), 1, &VMDriveUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if drive.Name != newName {
		t.Errorf("expected name %q, got %q", newName, drive.Name)
	}
}

func TestVMDriveService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDrives.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDriveService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_drives/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "x"
	_, err := client.VMDrives.Update(context.Background(), 999, &VMDriveUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	apiErr, ok := err.(*APIError)
	if IsNotFoundError(err) || !ok || apiErr.StatusCode != 404 {
		t.Errorf("expected API 404, not a missing resource, got %T: %v", err, err)
	}
}

func TestVMDriveService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{Key: FlexInt(1), Machine: 10, PowerState: "offline"})
		},
		"DELETE /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMDrives.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVMDriveService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete of already-gone drive returns NotFoundError
	err := client.VMDrives.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted drive")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMDriveService_Delete_HotUnplug(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": stubVMsByMachine(10, 10, false),
		"GET /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			state := "online"
			if getCalls > 1 {
				state = "offline"
			}
			jsonResponse(w, 200, VMDrive{Key: FlexInt(1), Machine: 10, PowerState: state})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "hotplugdrive" {
				t.Errorf("expected action 'hotplugdrive', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
		"DELETE /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMDrives.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete with hot-unplug failed: %v", err)
	}
}

func TestVMDriveService_HotplugDrive(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "hotplugdrive" {
				t.Errorf("expected action 'hotplugdrive', got %v", body["action"])
			}
			params := body["params"].(map[string]any)
			if params["device"] != "5" {
				t.Errorf("expected device '5', got %v", params["device"])
			}
			// Unplug should NOT be set for hotplug
			if unplug, ok := params["unplug"]; ok && unplug == true {
				t.Error("expected unplug to be false for hotplug")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_drives/5": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			jsonResponse(w, 200, VMDrive{Key: FlexInt(5), Machine: 10, PowerState: "online"})
		},
	}))

	err := client.VMDrives.HotplugDrive(context.Background(), 10, 5)
	if err != nil {
		t.Fatalf("HotplugDrive failed: %v", err)
	}
}

func TestVMDriveService_HotUnplugDrive(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "hotplugdrive" {
				t.Errorf("expected action 'hotplugdrive', got %v", body["action"])
			}
			params := body["params"].(map[string]any)
			if params["unplug"] != true {
				t.Error("expected unplug to be true for hot-unplug")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_drives/5": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			jsonResponse(w, 200, VMDrive{Key: FlexInt(5), Machine: 10, PowerState: "offline"})
		},
	}))

	err := client.VMDrives.HotUnplugDrive(context.Background(), 10, 5)
	if err != nil {
		t.Fatalf("HotUnplugDrive failed: %v", err)
	}
}
