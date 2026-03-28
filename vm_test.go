package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestVMService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{
				{ID: 1, Name: "vm-alpha", CPUCores: 2, RAM: 4096},
				{ID: 2, Name: "vm-beta", CPUCores: 4, RAM: 8192},
			})
		},
	}))

	vms, err := client.VMs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(vms))
	}
	if vms[0].Name != "vm-alpha" {
		t.Errorf("expected name 'vm-alpha', got %q", vms[0].Name)
	}
}

func TestVMService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{})
		},
	}))

	vms, err := client.VMs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(vms) != 0 {
		t.Fatalf("expected 0 VMs, got %d", len(vms))
	}
}

func TestVMService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter query param")
			}
			jsonResponse(w, 200, []VM{{ID: 1, Name: "vm-alpha"}})
		},
	}))

	vms, err := client.VMs.List(context.Background(), WithFilter("name eq 'vm-alpha'"))
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(vms) != 1 {
		t.Fatalf("expected 1 VM, got %d", len(vms))
	}
}

func TestVMService_List_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	_, err := client.VMs.List(context.Background())
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestVMService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: "vm-alpha", CPUCores: 2, RAM: 4096})
		},
	}))

	vm, err := client.VMs.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if vm.Name != "vm-alpha" {
		t.Errorf("expected name 'vm-alpha', got %q", vm.Name)
	}
	if vm.CPUCores != 2 {
		t.Errorf("expected 2 cores, got %d", vm.CPUCores)
	}
}

func TestVMService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMs.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestVMService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			var req VMCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "new-vm" {
				t.Errorf("expected name 'new-vm', got %q", req.Name)
			}
			if req.CPUCores != 4 {
				t.Errorf("expected 4 cores, got %d", req.CPUCores)
			}
			if req.RAM != 8192 {
				t.Errorf("expected 8192 RAM, got %d", req.RAM)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(42)})
		},
		"GET /api/v4/vms/42": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 42, Name: "new-vm", CPUCores: 4, RAM: 8192})
		},
	}))

	vm, err := client.VMs.Create(context.Background(), &VMCreateRequest{
		Name:     "new-vm",
		CPUCores: 4,
		RAM:      8192,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if vm.ID.Int() != 42 {
		t.Errorf("expected ID 42, got %d", vm.ID.Int())
	}
	if vm.Name != "new-vm" {
		t.Errorf("expected name 'new-vm', got %q", vm.Name)
	}
}

func TestVMService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMs.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMs.Create(context.Background(), &VMCreateRequest{
		CPUCores: 2,
		RAM:      4096,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Create_ZeroCPUCores(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMs.Create(context.Background(), &VMCreateRequest{
		Name: "bad-vm",
		RAM:  4096,
	})
	if err == nil {
		t.Fatal("expected error for zero cpu_cores")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Create_ZeroRAM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMs.Create(context.Background(), &VMCreateRequest{
		Name:     "bad-vm",
		CPUCores: 2,
	})
	if err == nil {
		t.Fatal("expected error for zero ram")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Create_DefaultsEnabled(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			var req VMCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected Enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: float64(1)})
		},
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", CPUCores: 1, RAM: 1024, Enabled: true})
		},
	}))

	_, err := client.VMs.Create(context.Background(), &VMCreateRequest{
		Name:     "vm",
		CPUCores: 1,
		RAM:      1024,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestVMService_Update(t *testing.T) {
	newName := "renamed-vm"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: newName, CPUCores: 2, RAM: 4096})
		},
	}))

	vm, err := client.VMs.Update(context.Background(), 1, &VMUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if vm.Name != newName {
		t.Errorf("expected name %q, got %q", newName, vm.Name)
	}
}

func TestVMService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMs.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Update_NotFound(t *testing.T) {
	newName := "ghost"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMs.Update(context.Background(), 999, &VMUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestVMService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVMService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VMs.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// PowerOn
// ---------------------------------------------------------------------------

func TestVMService_PowerOn_AlreadyRunning(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", PowerState: true})
		},
	}))

	err := client.VMs.PowerOn(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOn (already running) failed: %v", err)
	}
}

func TestVMService_PowerOn(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			// First call: stopped. Subsequent calls: running.
			running := getCalls > 1
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", PowerState: running})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "poweron" {
				t.Errorf("expected action 'poweron', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 1 {
				t.Errorf("expected vm 1, got %v", body["vm"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOn(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOn failed: %v", err)
	}
}

func TestVMService_PowerOn_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VMs.PowerOn(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// PowerOff
// ---------------------------------------------------------------------------

func TestVMService_PowerOff_AlreadyStopped(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", PowerState: false})
		},
	}))

	err := client.VMs.PowerOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOff (already stopped) failed: %v", err)
	}
}

func TestVMService_PowerOff(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			// First call: running. Subsequent calls: stopped.
			running := getCalls <= 1
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", PowerState: running})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "kill" {
				t.Errorf("expected action 'kill', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOff failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Reset
// ---------------------------------------------------------------------------

func TestVMService_Reset(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "reset" {
				t.Errorf("expected action 'reset', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 1 {
				t.Errorf("expected vm 1, got %v", body["vm"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Reset(context.Background(), 1)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestVMService_Reset_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.VMs.Reset(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// GuestReboot
// ---------------------------------------------------------------------------

func TestVMService_GuestReboot(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "guestreboot" {
				t.Errorf("expected action 'guestreboot', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 5 {
				t.Errorf("expected vm 5, got %v", body["vm"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.GuestReboot(context.Background(), 5)
	if err != nil {
		t.Fatalf("GuestReboot failed: %v", err)
	}
}

func TestVMService_GuestReboot_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.VMs.GuestReboot(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// GuestShutdown
// ---------------------------------------------------------------------------

func TestVMService_GuestShutdown(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "poweroff" {
				t.Errorf("expected action 'poweroff', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 3 {
				t.Errorf("expected vm 3, got %v", body["vm"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.GuestShutdown(context.Background(), 3)
	if err != nil {
		t.Fatalf("GuestShutdown failed: %v", err)
	}
}

func TestVMService_GuestShutdown_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.VMs.GuestShutdown(context.Background(), 3)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// Clone
// ---------------------------------------------------------------------------

func TestVMService_Clone(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "clone" {
				t.Errorf("expected action 'clone', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 1 {
				t.Errorf("expected vm 1, got %v", body["vm"])
			}
			params := body["params"].(map[string]any)
			if params["name"] != "cloned-vm" {
				t.Errorf("expected clone name 'cloned-vm', got %v", params["name"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Clone(context.Background(), 1, &VMCloneOptions{Name: "cloned-vm"})
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
}

func TestVMService_Clone_NilOpts(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "clone" {
				t.Errorf("expected action 'clone', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Clone(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("Clone with nil opts failed: %v", err)
	}
}

func TestVMService_Clone_PreserveMACs(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			params := body["params"].(map[string]any)
			if params["preserve_macs"] != true {
				t.Errorf("expected preserve_macs true, got %v", params["preserve_macs"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Clone(context.Background(), 1, &VMCloneOptions{PreserveMACs: true})
	if err != nil {
		t.Fatalf("Clone with PreserveMACs failed: %v", err)
	}
}

func TestVMService_Clone_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.VMs.Clone(context.Background(), 1, &VMCloneOptions{Name: "fail"})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// Snapshot
// ---------------------------------------------------------------------------

func TestVMService_Snapshot(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "quiesce_snapshot" {
				t.Errorf("expected action 'quiesce_snapshot', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 2 {
				t.Errorf("expected vm 2, got %v", body["vm"])
			}
			params := body["params"].(map[string]any)
			if int(params["retention"].(float64)) != 3600 {
				t.Errorf("expected retention 3600, got %v", params["retention"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Snapshot(context.Background(), 2, &VMSnapshotOptions{Retention: 3600})
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
}

func TestVMService_Snapshot_NilOpts(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "quiesce_snapshot" {
				t.Errorf("expected action 'quiesce_snapshot', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Snapshot(context.Background(), 2, nil)
	if err != nil {
		t.Fatalf("Snapshot with nil opts failed: %v", err)
	}
}

func TestVMService_Snapshot_WithQuiesce(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			params := body["params"].(map[string]any)
			if params["quiesce"] != true {
				t.Errorf("expected quiesce true, got %v", params["quiesce"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Snapshot(context.Background(), 2, &VMSnapshotOptions{Quiesce: true})
	if err != nil {
		t.Fatalf("Snapshot with quiesce failed: %v", err)
	}
}

func TestVMService_Snapshot_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.VMs.Snapshot(context.Background(), 2, nil)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// Migrate
// ---------------------------------------------------------------------------

func TestVMService_Migrate(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "migrate" {
				t.Errorf("expected action 'migrate', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 1 {
				t.Errorf("expected vm 1, got %v", body["vm"])
			}
			params := body["params"].(map[string]any)
			if int(params["node"].(float64)) != 3 {
				t.Errorf("expected node 3, got %v", params["node"])
			}
			if params["method"] != "live" {
				t.Errorf("expected method 'live', got %v", params["method"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.Migrate(context.Background(), 1, &VMMigrateOptions{TargetNode: 3})
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
}

func TestVMService_Migrate_NonLive(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			params := body["params"].(map[string]any)
			if params["method"] != "auto" {
				t.Errorf("expected method 'auto', got %v", params["method"])
			}
			w.WriteHeader(200)
		},
	}))

	live := false
	err := client.VMs.Migrate(context.Background(), 1, &VMMigrateOptions{
		TargetNode: 3,
		Live:       &live,
	})
	if err != nil {
		t.Fatalf("Migrate non-live failed: %v", err)
	}
}

func TestVMService_Migrate_NilOpts(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.VMs.Migrate(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil opts")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Migrate_ZeroTargetNode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.VMs.Migrate(context.Background(), 1, &VMMigrateOptions{TargetNode: 0})
	if err == nil {
		t.Fatal("expected error for zero target_node")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMService_Migrate_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
	}))

	err := client.VMs.Migrate(context.Background(), 1, &VMMigrateOptions{TargetNode: 3})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

// ---------------------------------------------------------------------------
// GetConsoleURL
// ---------------------------------------------------------------------------

func TestVMService_GetConsoleURL(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", PowerState: true})
		},
	}))

	url, err := client.VMs.GetConsoleURL(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetConsoleURL failed: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty console URL")
	}
}

func TestVMService_GetConsoleURL_VMStopped(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{ID: 1, Name: "vm", PowerState: false})
		},
	}))

	_, err := client.VMs.GetConsoleURL(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for stopped VM")
	}
}

func TestVMService_GetConsoleURL_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMs.GetConsoleURL(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
