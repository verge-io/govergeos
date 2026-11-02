package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// The tenant node $key, its machine key, and an unrelated VM $key are three
// different numbers. A lookup that uses the wrong one shows up in the filter
// or the posted machine.

const (
	testTenantNodeKey = 11
	testTenantMachine = 80
	testUnrelatedVM   = 45
)

func stubTenantNodeMachine(t *testing.T, key, machine int) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("fields"); got != "$key,machine" {
			t.Errorf("tenant node fields = %q, want $key,machine", got)
		}
		jsonResponse(w, 200, TenantNode{Key: FlexInt(key), Machine: FlexInt(machine)})
	}
}

func rejectVMLookup(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("tenant device call looked up a VM: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}
}

func TestTenantNodeDeviceService_List_UsesMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes/11": stubTenantNodeMachine(t, testTenantNodeKey, testTenantMachine),
		"GET /api/v4/vms/":            rejectVMLookup(t),
		"GET /api/v4/machine_devices": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != "machine eq 80" {
				t.Errorf("filter = %q, want machine eq 80", got)
			}
			if got := r.URL.Query().Get("fields"); got != deviceListFields {
				t.Errorf("fields = %q", got)
			}
			jsonResponse(w, 200, []VMDevice{
				{Key: FlexInt(7), Machine: testTenantMachine, Name: "pci0", Type: DeviceTypePCI, MachineType: DeviceMachineTypeTenant},
			})
		},
	}))

	devices, err := client.TenantNodeDevices.List(context.Background(), testTenantNodeKey)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("got %d devices, want 1", len(devices))
	}
	if devices[0].Machine != testTenantMachine {
		t.Errorf("device machine = %d, want %d", devices[0].Machine, testTenantMachine)
	}
	if devices[0].Key.Int() == testTenantNodeKey || devices[0].Key.Int() == testUnrelatedVM {
		t.Errorf("device key %d collided with the tenant node or the unrelated VM", devices[0].Key.Int())
	}
}

func TestTenantNodeDeviceService_List_NodeNotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes/11": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
		"GET /api/v4/machine_devices": func(w http.ResponseWriter, r *http.Request) {
			t.Error("listed devices without a tenant node")
		},
	}))

	_, err := client.TenantNodeDevices.List(context.Background(), testTenantNodeKey)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestTenantNodeDeviceService_List_NoMachineKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes/11": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantNode{Key: FlexInt(testTenantNodeKey)})
		},
	}))

	_, err := client.TenantNodeDevices.List(context.Background(), testTenantNodeKey)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestMachineKeyForTenantNode_InvalidID(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.machineKeyForTenantNode(context.Background(), 0)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:           FlexInt(7),
				Machine:       testTenantMachine,
				MachineType:   DeviceMachineTypeTenant,
				Name:          "pci0",
				Type:          DeviceTypePCI,
				ResourceGroup: DeviceResourceGroup("24122fab-e8fa-c4a5-dcf5-b52bf1efa474"),
			})
		},
	}))

	device, err := client.TenantNodeDevices.Get(context.Background(), 7)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if device.MachineType != DeviceMachineTypeTenant {
		t.Errorf("machine_type = %q", device.MachineType)
	}
	if device.ResourceGroup.String() != "24122fab-e8fa-c4a5-dcf5-b52bf1efa474" {
		t.Errorf("resource group = %q", device.ResourceGroup)
	}
}

func TestTenantNodeDeviceService_Get_NotTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:         FlexInt(7),
				Machine:     22,
				MachineType: "vm",
				Name:        "pci0",
				Type:        DeviceTypePCI,
			})
		},
	}))

	_, err := client.TenantNodeDevices.Get(context.Background(), 7)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TenantNodeDevices.Get(context.Background(), 999)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Create_UsesMachineKey(t *testing.T) {
	group := DeviceResourceGroup("24122fab-e8fa-c4a5-dcf5-b52bf1efa474")
	count := 2
	optional := true
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_nodes/11": stubTenantNodeMachine(t, testTenantNodeKey, testTenantMachine),
		"GET /api/v4/vms/":            rejectVMLookup(t),
		"POST /api/v4/machine_devices": func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			var req VMDeviceCreateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.Machine != testTenantMachine {
				t.Errorf("machine = %d, want %d (node key %d, unrelated VM %d)", req.Machine, testTenantMachine, testTenantNodeKey, testUnrelatedVM)
			}
			if req.Type != DeviceTypePCI || req.Name != "pci0" {
				t.Errorf("type/name = %s/%s", req.Type, req.Name)
			}
			if req.ResourceGroup != group {
				t.Errorf("resource group = %q", req.ResourceGroup)
			}
			if req.Count == nil || *req.Count != 2 {
				t.Errorf("count = %v", req.Count)
			}
			if req.Optional == nil || !*req.Optional {
				t.Error("optional was not set")
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("enabled should default to true")
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				t.Fatalf("raw decode: %v", err)
			}
			if string(raw["resource_group"]) != `"24122fab-e8fa-c4a5-dcf5-b52bf1efa474"` {
				t.Errorf("resource_group JSON = %s", raw["resource_group"])
			}
			if string(raw["machine"]) != "80" {
				t.Errorf("machine JSON = %s", raw["machine"])
			}
			jsonResponse(w, 200, apiResponse{Key: float64(7)})
		},
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:         FlexInt(7),
				Machine:     testTenantMachine,
				MachineType: DeviceMachineTypeTenant,
				Name:        "pci0",
				Type:        DeviceTypePCI,
			})
		},
	}))

	device, err := client.TenantNodeDevices.Create(context.Background(), testTenantNodeKey, &VMDeviceCreateRequest{
		Name:          "pci0",
		Type:          DeviceTypePCI,
		ResourceGroup: group,
		Count:         &count,
		Optional:      &optional,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if device.Key.Int() != 7 || device.Machine != testTenantMachine {
		t.Fatalf("device = %+v", device)
	}
}

func TestTenantNodeDeviceService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.TenantNodeDevices.Create(context.Background(), testTenantNodeKey, nil)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.TenantNodeDevices.Create(context.Background(), testTenantNodeKey, &VMDeviceCreateRequest{
		Type: DeviceTypePCI,
	})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Create_MissingType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.TenantNodeDevices.Create(context.Background(), testTenantNodeKey, &VMDeviceCreateRequest{
		Name: "pci0",
	})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Update(t *testing.T) {
	newName := "pci-renamed"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:         FlexInt(7),
				Machine:     testTenantMachine,
				MachineType: DeviceMachineTypeTenant,
				Name:        newName,
				Type:        DeviceTypePCI,
			})
		},
		"PUT /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			var req VMDeviceUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("name = %v", req.Name)
			}
			w.WriteHeader(200)
		},
	}))

	device, err := client.TenantNodeDevices.Update(context.Background(), 7, &VMDeviceUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if device.Name != newName {
		t.Errorf("name = %q", device.Name)
	}
}

func TestTenantNodeDeviceService_Update_NotTenant(t *testing.T) {
	put := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:         FlexInt(7),
				MachineType: "vm",
				Type:        DeviceTypePCI,
				Name:        "pci0",
			})
		},
		"PUT /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			put = true
			w.WriteHeader(200)
		},
	}))

	name := "nope"
	_, err := client.TenantNodeDevices.Update(context.Background(), 7, &VMDeviceUpdateRequest{Name: &name})
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if put {
		t.Fatal("updated a device that is not on a tenant node")
	}
}

func TestTenantNodeDeviceService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.TenantNodeDevices.Update(context.Background(), 7, nil)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantNodeDeviceService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:         FlexInt(7),
				MachineType: DeviceMachineTypeTenant,
				Type:        DeviceTypePCI,
			})
		},
		"DELETE /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	if err := client.TenantNodeDevices.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestTenantNodeDeviceService_Delete_NotTenant(t *testing.T) {
	deleted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				Key:         FlexInt(7),
				MachineType: "node",
				Type:        DeviceTypePCI,
			})
		},
		"DELETE /api/v4/machine_devices/7": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(200)
		},
	}))

	err := client.TenantNodeDevices.Delete(context.Background(), 7)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if deleted {
		t.Fatal("deleted a device that is not on a tenant node")
	}
}

func TestDeviceResourceGroup_JSON(t *testing.T) {
	uuid := DeviceResourceGroup("24122fab-e8fa-c4a5-dcf5-b52bf1efa474")
	encoded, err := json.Marshal(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `"24122fab-e8fa-c4a5-dcf5-b52bf1efa474"` {
		t.Fatalf("uuid JSON = %s", encoded)
	}

	number := DeviceResourceGroup("5")
	encoded, err = json.Marshal(number)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "5" {
		t.Fatalf("decimal JSON = %s, want 5", encoded)
	}

	var got DeviceResourceGroup
	if err := json.Unmarshal([]byte(`"24122fab-e8fa-c4a5-dcf5-b52bf1efa474"`), &got); err != nil {
		t.Fatal(err)
	}
	if got.String() != uuid.String() {
		t.Fatalf("decoded uuid %q", got)
	}
	if err := json.Unmarshal([]byte(`5`), &got); err != nil {
		t.Fatal(err)
	}
	if got.String() != "5" {
		t.Fatalf("decoded number %q", got)
	}

	var device VMDevice
	payload := []byte(`{"$key":7,"resource_group":"24122fab-e8fa-c4a5-dcf5-b52bf1efa474","type":"node_pci_devices","name":"pci0"}`)
	if err := json.Unmarshal(payload, &device); err != nil {
		t.Fatal(err)
	}
	if device.ResourceGroup.String() != uuid.String() {
		t.Fatalf("device resource group %q", device.ResourceGroup)
	}
}
