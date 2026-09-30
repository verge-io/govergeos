package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// Keys used throughout these tests differ on purpose. VM $key 45 and machine
// key 26 are the shape of a VM that has been around long enough for the two
// counters to diverge. Posting 26 to vm_actions, or filtering machine eq 45,
// targets a different object.

func TestMachineKeyForVM_ResolvesDistinctKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got != "$key,machine" {
				t.Errorf("fields = %q, want $key,machine", got)
			}
			jsonResponse(w, 200, VM{Key: FlexInt(45), Machine: 26})
		},
	}))

	machine, err := client.machineKeyForVM(context.Background(), 45)
	if err != nil {
		t.Fatalf("machineKeyForVM: %v", err)
	}
	if machine != 26 {
		t.Fatalf("machine = %d, want 26", machine)
	}
}

func TestMachineKeyForVM_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.machineKeyForVM(context.Background(), 45)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMachineKeyForVM_MissingMachine(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 0),
	}))

	_, err := client.machineKeyForVM(context.Background(), 45)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMKeyForMachine_SkipsOtherRows(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "machine eq 63" {
				t.Errorf("filter = %q, want machine eq 63", got)
			}
			if got := r.URL.Query().Get("fields"); got != "$key,machine,is_snapshot" {
				t.Errorf("fields = %q, want $key,machine,is_snapshot", got)
			}
			jsonResponse(w, 200, []VM{
				{Key: FlexInt(63), Machine: 26, IsSnapshot: false},
				{Key: FlexInt(45), Machine: 63, IsSnapshot: false},
				{Key: FlexInt(77), Machine: 99, IsSnapshot: true},
				{Key: FlexInt(90), Machine: 63, IsSnapshot: true},
			})
		},
	}))

	got, err := client.vmKeyForMachine(context.Background(), 63, true)
	if err != nil {
		t.Fatalf("vmKeyForMachine snapshot: %v", err)
	}
	if got != 90 {
		t.Fatalf("snapshot VM key = %d, want 90", got)
	}

	// A second client so the live-VM lookup is a separate request assertion.
	live := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{
				{Key: FlexInt(88), Machine: 26, IsSnapshot: true},
				{Key: FlexInt(26), Machine: 99, IsSnapshot: false},
				{Key: FlexInt(45), Machine: 26, IsSnapshot: false},
			})
		},
	}))
	got, err = live.vmKeyForMachine(context.Background(), 26, false)
	if err != nil {
		t.Fatalf("vmKeyForMachine live: %v", err)
	}
	if got != 45 {
		t.Fatalf("live VM key = %d, want 45", got)
	}
}

func TestVMSnapshotService_ListByVM_ResolvesMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "machine eq 26" {
				t.Errorf("filter = %q, want machine eq 26", got)
			}
			jsonResponse(w, 200, []VMSnapshot{{Key: FlexInt(1), Machine: FlexInt(26), Name: "snap"}})
		},
	}))

	snaps, err := client.VMSnapshots.ListByVM(context.Background(), 45)
	if err != nil {
		t.Fatalf("ListByVM: %v", err)
	}
	if len(snaps) != 1 || int(snaps[0].Machine) != 26 {
		t.Fatalf("snapshots = %#v", snaps)
	}
}

func TestVMSnapshotService_GetByName_ResolvesMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"GET /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			want := "(name eq 'pre-upgrade') and machine eq 26"
			if got := r.URL.Query().Get("filter"); got != want {
				t.Errorf("filter = %q, want %q", got, want)
			}
			jsonResponse(w, 200, []VMSnapshot{{Key: FlexInt(3), Machine: FlexInt(26), Name: "pre-upgrade"}})
		},
	}))

	snap, err := client.VMSnapshots.GetByName(context.Background(), 45, "pre-upgrade")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if snap.Key.Int() != 3 {
		t.Fatalf("key = %d, want 3", snap.Key.Int())
	}
}

func TestVMSnapshotService_Create_ResolvesMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var body vmSnapshotCreateBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Machine != 26 {
				t.Errorf("machine = %d, want 26", body.Machine)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(8)})
		},
		"GET /api/v4/machine_snapshots/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(8), Machine: FlexInt(26), Name: "snap"})
		},
	}))

	snap, err := client.VMSnapshots.Create(context.Background(), &VMSnapshotCreateRequest{
		VM:   45,
		Name: "snap",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if snap.Key.Int() != 8 {
		t.Fatalf("key = %d, want 8", snap.Key.Int())
	}
}

func TestVMSnapshotService_Restore_MissingSnapMachine(t *testing.T) {
	posted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), Machine: FlexInt(26)})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			posted = true
			w.WriteHeader(200)
		},
	}))

	err := client.VMSnapshots.Restore(context.Background(), 1, nil)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if posted {
		t.Fatal("vm_actions was posted without a snap_machine")
	}
}

func TestVMSnapshotService_Restore_SnapshotVMNotFound(t *testing.T) {
	posted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_snapshots/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMSnapshot{Key: FlexInt(1), Machine: FlexInt(26), SnapMachine: FlexInt(63)})
		},
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			// A live VM and a snapshot of a different machine must not be used.
			jsonResponse(w, 200, []VM{
				{Key: FlexInt(63), Machine: 26, IsSnapshot: false},
				{Key: FlexInt(77), Machine: 99, IsSnapshot: true},
			})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			posted = true
			w.WriteHeader(200)
		},
	}))

	err := client.VMSnapshots.Restore(context.Background(), 1, nil)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
	if posted {
		t.Fatal("vm_actions was posted without a snapshot VM")
	}
}

func TestVMDriveService_ListAndCreate_ResolveMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "machine eq 26" {
				t.Errorf("list filter = %q, want machine eq 26", got)
			}
			jsonResponse(w, 200, []VMDrive{})
		},
		"POST /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			var req VMDriveCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.Machine != 26 {
				t.Errorf("machine = %d, want 26", req.Machine)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(7)})
		},
		"GET /api/v4/machine_drives/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{Key: FlexInt(7), Machine: 26, Name: "disk0"})
		},
	}))

	if _, err := client.VMDrives.List(context.Background(), 45); err != nil {
		t.Fatalf("List: %v", err)
	}
	drive, err := client.VMDrives.Create(context.Background(), 45, &VMDriveCreateRequest{Name: "disk0"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if drive.Machine != 26 {
		t.Fatalf("drive machine = %d, want 26", drive.Machine)
	}
}

func TestVMDriveService_GetByName_ResolvesMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"GET /api/v4/machine_drives": func(w http.ResponseWriter, r *http.Request) {
			want := "machine eq 26 and name eq 'disk0'"
			if got := r.URL.Query().Get("filter"); got != want {
				t.Errorf("filter = %q, want %q", got, want)
			}
			jsonResponse(w, 200, []VMDrive{{Key: FlexInt(7), Machine: 26, Name: "disk0"}})
		},
		"GET /api/v4/machine_drives/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{Key: FlexInt(7), Machine: 26, Name: "disk0"})
		},
	}))

	drive, err := client.VMDrives.GetByName(context.Background(), 45, "disk0")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if drive.Key.Int() != 7 {
		t.Fatalf("id = %d, want 7", drive.Key.Int())
	}
}

func TestVMDriveService_Delete_HotUnplug_UsesVMKey(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			state := "online"
			if getCalls > 1 {
				state = "offline"
			}
			jsonResponse(w, 200, VMDrive{Key: FlexInt(1), Machine: 26, PowerState: state})
		},
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "machine eq 26" {
				t.Errorf("filter = %q, want machine eq 26", got)
			}
			jsonResponse(w, 200, []VM{
				{Key: FlexInt(26), Machine: 99, IsSnapshot: false},
				{Key: FlexInt(88), Machine: 26, IsSnapshot: true},
				{Key: FlexInt(45), Machine: 26, IsSnapshot: false},
			})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body["action"] != "hotplugdrive" {
				t.Errorf("action = %v, want hotplugdrive", body["action"])
			}
			if int(body["vm"].(float64)) != 45 {
				t.Errorf("vm = %v, want VM key 45 (not machine 26)", body["vm"])
			}
			w.WriteHeader(200)
		},
		"DELETE /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	if err := client.VMDrives.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestVMDriveService_Delete_HotUnplug_VMNotFound(t *testing.T) {
	deleted := false
	posted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDrive{Key: FlexInt(1), Machine: 26, PowerState: "online"})
		},
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{
				{Key: FlexInt(88), Machine: 26, IsSnapshot: true},
			})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			posted = true
			w.WriteHeader(200)
		},
		"DELETE /api/v4/machine_drives/1": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(200)
		},
	}))

	err := client.VMDrives.Delete(context.Background(), 1)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
	if posted || deleted {
		t.Fatalf("posted=%v deleted=%v; unplug and delete must not run", posted, deleted)
	}
}

func TestVMNICService_CreateAndDelete_ResolveKeys(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"POST /api/v4/machine_nics": func(w http.ResponseWriter, r *http.Request) {
			var req VMNICCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.Machine != 26 {
				t.Errorf("machine = %d, want 26", req.Machine)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(4)})
		},
		"GET /api/v4/machine_nics/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMNIC{Key: FlexInt(4), Machine: 26, Name: "nic0"})
		},
		"GET /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			state := "up"
			if getCalls > 1 {
				state = "down"
			}
			jsonResponse(w, 200, VMNIC{Key: FlexInt(1), Machine: 26, PowerState: state})
		},
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{
				{Key: FlexInt(26), Machine: 99, IsSnapshot: false},
				{Key: FlexInt(45), Machine: 26, IsSnapshot: false},
			})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body["action"] != "hotplugnic" {
				t.Errorf("action = %v, want hotplugnic", body["action"])
			}
			if int(body["vm"].(float64)) != 45 {
				t.Errorf("vm = %v, want VM key 45 (not machine 26)", body["vm"])
			}
			w.WriteHeader(200)
		},
		"DELETE /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	nic, err := client.VMNICs.Create(context.Background(), 45, &VMNICCreateRequest{Name: "nic0"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if nic.Machine != 26 {
		t.Fatalf("nic machine = %d, want 26", nic.Machine)
	}
	if err := client.VMNICs.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestVMDeviceService_Create_ResolvesMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/45": stubVM(45, 26),
		"POST /api/v4/machine_devices": func(w http.ResponseWriter, r *http.Request) {
			var req VMDeviceCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.Machine != 26 {
				t.Errorf("machine = %d, want 26", req.Machine)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/machine_devices/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{Key: FlexInt(5), Machine: 26, Name: "tpm0", Type: DeviceTypeTPM})
		},
		"GET /api/v4/machine_device_settings_tpm": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TPMDeviceSettings{})
		},
	}))

	device, err := client.VMDevices.Create(context.Background(), 45, &VMDeviceCreateRequest{
		Name: "tpm0",
		Type: DeviceTypeTPM,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if device.Machine != 26 {
		t.Fatalf("device machine = %d, want 26", device.Machine)
	}
}
