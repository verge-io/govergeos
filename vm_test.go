package vergeos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestVMService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{
				{Key: 1, Name: "vm-alpha", CPUCores: 2, RAM: 4096},
				{Key: 2, Name: "vm-beta", CPUCores: 4, RAM: 8192},
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
			jsonResponse(w, 200, []VM{{Key: 1, Name: "vm-alpha"}})
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm-alpha", CPUCores: 2, RAM: 4096})
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

func TestVMService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "is_snapshot eq false and name eq 'vm-alpha'" {
				t.Errorf("filter = %q", got)
			}
			jsonResponse(w, 200, []VM{{Key: 1, Name: "vm-alpha"}})
		},
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm-alpha", Description: "full"})
		},
	}))

	vm, err := client.VMs.GetByName(context.Background(), "vm-alpha")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if vm.Name != "vm-alpha" || vm.Description != "full" {
		t.Fatalf("got %+v", vm)
	}
}

func TestVMService_GetByName_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{
				{Key: 1, Name: "vm-alpha"},
				{Key: 9, Name: "vm-alpha"},
			})
		},
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			t.Error("GetByName fetched the first match")
		},
	}))

	_, err := client.VMs.GetByName(context.Background(), "vm-alpha")
	amb, ok := err.(*AmbiguousNameError)
	if !ok {
		t.Fatalf("expected *AmbiguousNameError, got %T: %v", err, err)
	}
	if amb.Resource != "VM" || amb.Name != "vm-alpha" || len(amb.Keys) != 2 || amb.Keys[0] != FlexInt(1) || amb.Keys[1] != FlexInt(9) {
		t.Fatalf("unexpected details: %+v", amb)
	}
}

func TestVMService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{})
		},
	}))

	_, err := client.VMs.GetByName(context.Background(), "missing")
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
			jsonResponse(w, 200, VM{Key: 42, Name: "new-vm", CPUCores: 4, RAM: 8192})
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
	if vm.Key.Int() != 42 {
		t.Errorf("expected ID 42, got %d", vm.Key.Int())
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", CPUCores: 1, RAM: 1024, Enabled: true})
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
			jsonResponse(w, 200, VM{Key: 1, Name: newName, CPUCores: 2, RAM: 4096})
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
	apiErr, ok := err.(*APIError)
	if IsNotFoundError(err) || !ok || apiErr.StatusCode != 404 {
		t.Errorf("expected API 404, not a missing resource, got %T: %v", err, err)
	}
}

func TestVMService_Update_MissingSnapshotProfile(t *testing.T) {
	const platform = "Error from put in trigger 'put:machines/{machine}?snapshot_profile={snapshot_profile}' table 'vms' - No such file or directory\n"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vms/44": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": platform})
		},
		"GET /api/v4/vms/44": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 44, Name: "zzrc3-nf-vm"})
		},
	}))

	profile := 99999
	_, err := client.VMs.Update(context.Background(), 44, &VMUpdateRequest{SnapshotProfile: &profile})
	if err == nil {
		t.Fatal("expected error")
	}
	if IsNotFoundError(err) {
		t.Fatalf("reference 404 reported the VM as missing: %v", err)
	}
	if !strings.Contains(err.Error(), "snapshot_profile") {
		t.Fatalf("error = %q, want platform message", err.Error())
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 404 || apiErr.Message != platform {
		t.Fatalf("got %T %#v", err, err)
	}

	got, gerr := client.VMs.Get(context.Background(), 44)
	if gerr != nil {
		t.Fatalf("VM still exists, Get failed: %v", gerr)
	}
	if got.Name != "zzrc3-nf-vm" {
		t.Fatalf("vm = %+v", got)
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: running})
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: false})
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: running})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "poweroff" {
				t.Errorf("expected action 'poweroff', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOff(context.Background(), 1)
	if err != nil {
		t.Fatalf("PowerOff failed: %v", err)
	}
}

func TestVMService_PowerOffWithOptions_NilOpts(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: false})
		},
	}))

	if err := client.VMs.PowerOffWithOptions(context.Background(), 1, nil); err != nil {
		t.Fatalf("PowerOffWithOptions(nil) failed: %v", err)
	}
}

func TestVMService_PowerOff_UsesClientPowerWait(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))
	client.powerWaitTimeout = 40 * time.Millisecond
	client.powerWaitInterval = 5 * time.Millisecond

	start := time.Now()
	err := client.VMs.PowerOff(context.Background(), 1)
	elapsed := time.Since(start)
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError, got %v", err)
	}
	if elapsed > time.Second {
		t.Fatalf("PowerOff waited %s; client power-wait timeout was not applied", elapsed)
	}
}

func TestVMService_PowerOffWithOptions_ExtendsClientTimeout(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))
	client.powerWaitTimeout = 40 * time.Millisecond
	client.powerWaitInterval = 5 * time.Millisecond

	start := time.Now()
	err := client.VMs.PowerOff(context.Background(), 1)
	shortWait := time.Since(start)
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError from client default, got %v", err)
	}

	start = time.Now()
	err = client.VMs.PowerOffWithOptions(context.Background(), 1, &VMPowerOffOptions{
		Timeout:      200 * time.Millisecond,
		PollInterval: 5 * time.Millisecond,
	})
	longWait := time.Since(start)
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError from call timeout, got %v", err)
	}
	if longWait < shortWait*3 {
		t.Fatalf("call timeout %s did not extend client timeout %s", longWait, shortWait)
	}
	if longWait > 2*time.Second {
		t.Fatalf("call timeout waited %s; the fixed 150s budget is still in effect", longWait)
	}
}

func TestVMService_PowerOffWithOptions_PollInterval(t *testing.T) {
	var reads []time.Time
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			reads = append(reads, time.Now())
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))
	// A 5ms client interval would space reads tightly. The call asks for 40ms.
	client.powerWaitInterval = 5 * time.Millisecond

	start := time.Now()
	err := client.VMs.PowerOffWithOptions(context.Background(), 1, &VMPowerOffOptions{
		Timeout:      180 * time.Millisecond,
		PollInterval: 50 * time.Millisecond,
	})
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("poll interval was not applied; wait took %s", time.Since(start))
	}
	if len(reads) < 3 {
		t.Fatalf("expected several polls, got %d", len(reads))
	}
	gaps := make([]time.Duration, 0, len(reads)-1)
	for i := 1; i < len(reads); i++ {
		gaps = append(gaps, reads[i].Sub(reads[i-1]))
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i] < gaps[j] })
	median := gaps[len(gaps)/2]
	// The client interval is 5ms. A median near 50ms shows the call interval won.
	if median < 30*time.Millisecond {
		t.Fatalf("median poll gap = %s, want at least 30ms", median)
	}
}

func TestVMService_PowerOffWithOptions_TimeoutDoesNotKill(t *testing.T) {
	var actions []string
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			actions = append(actions, body["action"].(string))
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOffWithOptions(context.Background(), 1, &VMPowerOffOptions{
		Timeout:      25 * time.Millisecond,
		PollInterval: 5 * time.Millisecond,
	})
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError, got %v", err)
	}
	var timeoutErr *TimeoutError
	if !errors.As(err, &timeoutErr) || timeoutErr.Action != "become stopped" {
		t.Fatalf("timeout action = %v, want become stopped", err)
	}
	if len(actions) != 1 || actions[0] != "poweroff" {
		t.Fatalf("actions = %v, want [poweroff]", actions)
	}
}

func TestVMService_PowerOffWithOptions_ForceAfterTimeout(t *testing.T) {
	var actions []string
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			killed := false
			for _, action := range actions {
				if action == "kill" {
					killed = true
				}
			}
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: !killed})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			actions = append(actions, body["action"].(string))
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOffWithOptions(context.Background(), 1, &VMPowerOffOptions{
		Timeout:           30 * time.Millisecond,
		PollInterval:      5 * time.Millisecond,
		ForceAfterTimeout: true,
	})
	if err != nil {
		t.Fatalf("PowerOffWithOptions force failed: %v", err)
	}
	if len(actions) != 2 || actions[0] != "poweroff" || actions[1] != "kill" {
		t.Fatalf("actions = %v, want [poweroff kill]", actions)
	}
}

func TestVMService_PowerOffWithOptions_ForceAfterTimeout_KillAlsoTimesOut(t *testing.T) {
	var actions []string
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			actions = append(actions, body["action"].(string))
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOffWithOptions(context.Background(), 1, &VMPowerOffOptions{
		Timeout:           25 * time.Millisecond,
		PollInterval:      5 * time.Millisecond,
		ForceAfterTimeout: true,
	})
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError, got %v", err)
	}
	var timeoutErr *TimeoutError
	if !errors.As(err, &timeoutErr) || timeoutErr.Action != "become stopped after kill" {
		t.Fatalf("timeout action = %v, want become stopped after kill", err)
	}
	if len(actions) != 2 || actions[0] != "poweroff" || actions[1] != "kill" {
		t.Fatalf("actions = %v, want [poweroff kill]", actions)
	}
}

func TestVMService_PowerOffWithOptions_CancelledContextDoesNotKill(t *testing.T) {
	var actions []string
	getCalls := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			// The state check is the first read. Cancel once the wait starts.
			if getCalls >= 2 {
				cancel()
			}
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			actions = append(actions, body["action"].(string))
			w.WriteHeader(200)
		},
	}))

	err := client.VMs.PowerOffWithOptions(ctx, 1, &VMPowerOffOptions{
		Timeout:           5 * time.Second,
		PollInterval:      5 * time.Millisecond,
		ForceAfterTimeout: true,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if len(actions) != 1 || actions[0] != "poweroff" {
		t.Fatalf("actions = %v, want [poweroff]", actions)
	}
}

func TestVMService_PowerOffWithOptions_NegativeTimeout(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	err := client.VMs.PowerOffWithOptions(context.Background(), 1, &VMPowerOffOptions{
		Timeout: -time.Second,
	})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestVMService_PowerOn_UsesClientPowerWait(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: false})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "poweron" {
				t.Errorf("expected action 'poweron', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))
	client.powerWaitTimeout = 40 * time.Millisecond
	client.powerWaitInterval = 5 * time.Millisecond

	start := time.Now()
	err := client.VMs.PowerOn(context.Background(), 1)
	if !IsTimeoutError(err) {
		t.Fatalf("expected TimeoutError, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("PowerOn waited %s; client power-wait timeout was not applied", time.Since(start))
	}
}

// ---------------------------------------------------------------------------
// Kill
// ---------------------------------------------------------------------------

func TestVMService_Kill_AlreadyStopped(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: false})
		},
	}))

	if err := client.VMs.Kill(context.Background(), 1); err != nil {
		t.Fatalf("Kill (already stopped) failed: %v", err)
	}
}

func TestVMService_Kill(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/4": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			running := getCalls <= 1
			jsonResponse(w, 200, VM{Key: 4, Name: "vm", PowerState: running})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "kill" {
				t.Errorf("expected action 'kill', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 4 {
				t.Errorf("expected vm 4, got %v", body["vm"])
			}
			w.WriteHeader(200)
		},
	}))

	if err := client.VMs.Kill(context.Background(), 4); err != nil {
		t.Fatalf("Kill failed: %v", err)
	}
	if getCalls < 2 {
		t.Fatalf("Kill returned before reading the stopped state: %d reads", getCalls)
	}
}

func TestVMService_Kill_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VMs.Kill(context.Background(), 999)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestWithPowerWait(t *testing.T) {
	c := &Client{}
	if err := WithPowerWait(2*time.Minute, time.Second)(c); err != nil {
		t.Fatalf("WithPowerWait: %v", err)
	}
	timeout, interval := c.vmPowerWait()
	if timeout != 2*time.Minute || interval != time.Second {
		t.Fatalf("power wait = %s / %s, want 2m0s / 1s", timeout, interval)
	}

	// A zero duration keeps the stored value, which is the default when unset.
	fresh := &Client{}
	if err := WithPowerWait(0, 0)(fresh); err != nil {
		t.Fatalf("WithPowerWait zero: %v", err)
	}
	timeout, interval = fresh.vmPowerWait()
	if timeout != defaultPowerWaitTimeout || interval != defaultPowerWaitInterval {
		t.Fatalf("default power wait = %s / %s, want %s / %s", timeout, interval, defaultPowerWaitTimeout, defaultPowerWaitInterval)
	}

	if err := WithPowerWait(-time.Second, time.Second)(c); err == nil {
		t.Fatal("expected error for a negative timeout")
	}
	if err := WithPowerWait(time.Second, -time.Second)(c); err == nil {
		t.Fatal("expected error for a negative poll interval")
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
			if params, ok := body["params"].(map[string]any); ok {
				if _, set := params["graceful"]; set {
					t.Errorf("hard reset must not set graceful, got %v", params)
				}
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
			if body["action"] != "reset" {
				t.Errorf("expected action 'reset', got %v", body["action"])
			}
			if int(body["vm"].(float64)) != 5 {
				t.Errorf("expected vm 5, got %v", body["vm"])
			}
			params, ok := body["params"].(map[string]any)
			if !ok {
				t.Fatal("expected params in body")
			}
			if params["graceful"] != true {
				t.Errorf("expected params.graceful true, got %v", params["graceful"])
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
	const retention = 3600
	var created map[string]any
	var actionCalled bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 2, Name: "vm-beta", Machine: 99})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&created)
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/machine_snapshots/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(5), Machine: FlexInt(99), Name: "pre-upgrade"})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			actionCalled = true
			w.WriteHeader(200)
		},
	}))

	before := time.Now().Unix()
	snap, err := client.VMs.Snapshot(context.Background(), 2, &VMSnapshotOptions{
		Name:      "pre-upgrade",
		Retention: retention,
	})
	after := time.Now().Unix()
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	if snap == nil || snap.Key.Int() != 5 {
		t.Fatalf("expected snapshot key 5, got %#v", snap)
	}
	if snap.Name != "pre-upgrade" {
		t.Errorf("expected name %q, got %q", "pre-upgrade", snap.Name)
	}
	if actionCalled {
		t.Error("quiesce_snapshot action must not be sent when Quiesce is false")
	}
	if _, has := created["retention"]; has {
		t.Error("machine_snapshots body must not include retention")
	}
	if created["machine"] != float64(99) {
		t.Errorf("expected machine key 99, got %v", created["machine"])
	}
	if created["name"] != "pre-upgrade" {
		t.Errorf("expected name pre-upgrade, got %v", created["name"])
	}
	if created["expires_type"] != "date" {
		t.Errorf("expected expires_type date, got %v", created["expires_type"])
	}
	if _, has := created["quiesce"]; has {
		t.Error("quiesce must be omitted when Quiesce is false")
	}
	expires, ok := created["expires"].(float64)
	if !ok {
		t.Fatalf("expected numeric expires, got %T", created["expires"])
	}
	minExpires := float64(before + retention)
	maxExpires := float64(after + retention)
	if expires < minExpires || expires > maxExpires {
		t.Errorf("expires %v outside [%v, %v]", expires, minExpires, maxExpires)
	}
}

func TestVMService_Snapshot_NilOpts(t *testing.T) {
	var created map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 2, Machine: 99})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&created)
			jsonResponse(w, 200, apiResponse{Key: float64(6)})
		},
		"GET /api/v4/machine_snapshots/6": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(6), Machine: FlexInt(99), Name: "generated"})
		},
	}))

	before := time.Now().Unix()
	snap, err := client.VMs.Snapshot(context.Background(), 2, nil)
	after := time.Now().Unix()
	if err != nil {
		t.Fatalf("Snapshot with nil opts failed: %v", err)
	}
	if snap == nil || snap.Key.Int() != 6 {
		t.Fatalf("expected snapshot key 6, got %#v", snap)
	}
	name, _ := created["name"].(string)
	if len(name) < len("snapshot-") || name[:len("snapshot-")] != "snapshot-" {
		t.Errorf("expected generated snapshot- name, got %q", name)
	}
	expires, ok := created["expires"].(float64)
	if !ok {
		t.Fatalf("expected numeric expires, got %T", created["expires"])
	}
	minExpires := float64(before + vmSnapshotDefaultRetention)
	maxExpires := float64(after + vmSnapshotDefaultRetention)
	if expires < minExpires || expires > maxExpires {
		t.Errorf("default expires %v outside [%v, %v]", expires, minExpires, maxExpires)
	}
}

func TestVMService_Snapshot_WithQuiesce(t *testing.T) {
	var created map[string]any
	var action map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 2, Machine: 99})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&created)
			jsonResponse(w, 200, apiResponse{Key: float64(7)})
		},
		"GET /api/v4/machine_snapshots/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(7), Machine: FlexInt(99), Name: "q"})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&action)
			w.WriteHeader(200)
		},
	}))

	snap, err := client.VMs.Snapshot(context.Background(), 2, &VMSnapshotOptions{Name: "q", Quiesce: true})
	if err != nil {
		t.Fatalf("Snapshot with quiesce failed: %v", err)
	}
	if snap == nil || snap.Key.Int() != 7 {
		t.Fatalf("expected snapshot key 7, got %#v", snap)
	}
	if created["quiesce"] != true {
		t.Errorf("expected quiesce true on machine_snapshots, got %v", created["quiesce"])
	}
	if action == nil {
		t.Fatal("expected quiesce_snapshot action when Quiesce is true")
	}
	if action["action"] != vmActionSnapshot {
		t.Errorf("expected action %q, got %v", vmActionSnapshot, action["action"])
	}
	if action["vm"] != float64(2) {
		t.Errorf("expected vm 2, got %v", action["vm"])
	}
	params, _ := action["params"].(map[string]any)
	if params["quiesce"] != true {
		t.Errorf("expected quiesce true on action, got %v", params["quiesce"])
	}
}

func TestVMService_Snapshot_QuiesceActionError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 2, Machine: 99})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, apiResponse{Key: float64(8)})
		},
		"GET /api/v4/machine_snapshots/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(8), Machine: FlexInt(99), Name: "q"})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "guest agent unavailable"})
		},
	}))

	snap, err := client.VMs.Snapshot(context.Background(), 2, &VMSnapshotOptions{Quiesce: true})
	if err == nil {
		t.Fatal("expected error when quiesce action fails")
	}
	if snap == nil || snap.Key.Int() != 8 {
		t.Fatalf("expected created snapshot alongside quiesce error, got %#v", snap)
	}
}

func TestVMService_Snapshot_ServerError(t *testing.T) {
	var actionCalled bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 2, Machine: 99})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			actionCalled = true
			w.WriteHeader(200)
		},
	}))

	snap, err := client.VMs.Snapshot(context.Background(), 2, nil)
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if snap != nil {
		t.Errorf("expected nil snapshot on create failure, got %#v", snap)
	}
	if actionCalled {
		t.Error("quiesce action must not be sent when snapshot creation fails")
	}
}

func TestVMService_Snapshot_VMNotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMs.Snapshot(context.Background(), 999, nil)
	if err == nil {
		t.Fatal("expected error for missing VM")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMService_Snapshot_MissingMachine(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 2, Name: "vm-beta"})
		},
	}))

	_, err := client.VMs.Snapshot(context.Background(), 2, nil)
	if err == nil {
		t.Fatal("expected error when VM has no machine key")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: true})
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
			jsonResponse(w, 200, VM{Key: 1, Name: "vm", PowerState: false})
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

// ---------------------------------------------------------------------------
// GetGuestAgentInfo
// ---------------------------------------------------------------------------

func TestVMService_GetGuestAgentInfo(t *testing.T) {
	const body = `{
		"machine": {
			"status": {
				"agent_guest_info": {
					"network": [
						{
							"name": "lo",
							"ip-addresses": [
								{"ip-address-type": "ipv4", "ip-address": "127.0.0.1"}
							]
						},
						{
							"name": "ens1",
							"ip-addresses": [
								{"ip-address-type": "ipv4", "ip-address": "10.0.0.42"},
								{"ip-address-type": "ipv6", "ip-address": "fe80::1"}
							]
						}
					]
				}
			}
		}
	}`

	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got != "dashboard" {
				t.Errorf("expected fields=dashboard, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(body))
		},
	}))

	info, err := client.VMs.GetGuestAgentInfo(context.Background(), 45)
	if err != nil {
		t.Fatalf("GetGuestAgentInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil info")
	}
	if len(info.Network) != 2 {
		t.Fatalf("expected 2 network interfaces, got %d", len(info.Network))
	}
	if info.Network[1].Name != "ens1" {
		t.Errorf("expected interface name 'ens1', got %q", info.Network[1].Name)
	}
	if len(info.Network[1].IPAddresses) != 2 {
		t.Fatalf("expected 2 IPs on ens1, got %d", len(info.Network[1].IPAddresses))
	}
	if info.Network[1].IPAddresses[0].Address != "10.0.0.42" {
		t.Errorf("expected IP '10.0.0.42', got %q", info.Network[1].IPAddresses[0].Address)
	}
	if info.Network[1].IPAddresses[0].Type != "ipv4" {
		t.Errorf("expected type 'ipv4', got %q", info.Network[1].IPAddresses[0].Type)
	}
}

func TestVMService_GetGuestAgentInfo_NetworkAsObject(t *testing.T) {
	// Some guest agent versions report network as an object keyed by interface
	// name instead of an array.
	const body = `{
		"machine": {
			"status": {
				"agent_guest_info": {
					"network": {
						"ens1": {
							"name": "ens1",
							"ip-addresses": [
								{"ip-address-type": "ipv4", "ip-address": "10.0.0.42"}
							]
						}
					}
				}
			}
		}
	}`

	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(body))
		},
	}))

	info, err := client.VMs.GetGuestAgentInfo(context.Background(), 45)
	if err != nil {
		t.Fatalf("GetGuestAgentInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil info")
	}
	if len(info.Network) != 1 {
		t.Fatalf("expected 1 network interface, got %d", len(info.Network))
	}
	if info.Network[0].Name != "ens1" {
		t.Errorf("expected interface name 'ens1', got %q", info.Network[0].Name)
	}
	if info.Network[0].IPAddresses[0].Address != "10.0.0.42" {
		t.Errorf("expected IP '10.0.0.42', got %q", info.Network[0].IPAddresses[0].Address)
	}
}

func TestVMService_GetGuestAgentInfo_NotReported(t *testing.T) {
	// Two "not reported" wire shapes: agent_guest_info absent entirely, and
	// the API's agent_guest_info: [] placeholder. Both must yield nil.
	for name, body := range map[string]string{
		"absent":     `{"machine": {"status": {"status": "running"}}}`,
		"emptyArray": `{"machine": {"status": {"status": "running", "agent_guest_info": []}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"GET /api/v4/vms/45": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(200)
					w.Write([]byte(body))
				},
			}))

			info, err := client.VMs.GetGuestAgentInfo(context.Background(), 45)
			if err != nil {
				t.Fatalf("GetGuestAgentInfo failed: %v", err)
			}
			if info != nil {
				t.Errorf("expected nil info when guest agent has not reported, got %+v", info)
			}
		})
	}
}

func TestGuestInfo_IPsForMAC(t *testing.T) {
	info := &GuestInfo{
		Network: []GuestNetworkInterface{
			{Name: "lo", HardwareAddress: "00:00:00:00:00:00", IPAddresses: []GuestIPAddress{{Type: "ipv4", Address: "127.0.0.1"}}},
			{Name: "docker0", HardwareAddress: "86:8f:54:f5:b9:00", IPAddresses: []GuestIPAddress{{Type: "ipv4", Address: "172.17.0.1"}}},
			{Name: "Ethernet 2", HardwareAddress: "F0:DB:30:BD:95:0E", IPAddresses: []GuestIPAddress{
				{Type: "ipv6", Address: "fe80::1%14"},
				{Type: "ipv4", Address: "192.168.10.176"},
			}},
		},
	}

	ips := info.IPsForMAC("f0:db:30:bd:95:0e")
	if len(ips) != 2 {
		t.Fatalf("expected 2 IPs (case-insensitive match), got %d", len(ips))
	}
	if ips[1].Address != "192.168.10.176" {
		t.Errorf("expected '192.168.10.176', got %q", ips[1].Address)
	}

	if got := info.IPsForMAC("aa:bb:cc:dd:ee:ff"); got != nil {
		t.Errorf("expected nil for unknown MAC, got %v", got)
	}
	var nilInfo *GuestInfo
	if got := nilInfo.IPsForMAC("f0:db:30:bd:95:0e"); got != nil {
		t.Errorf("expected nil for nil GuestInfo, got %v", got)
	}
}

func TestVMService_GetGuestAgentInfo_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMs.GetGuestAgentInfo(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
