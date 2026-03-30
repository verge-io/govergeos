package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVMDeviceService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VMDevice{
				{ID: FlexInt(1), Machine: 42, Name: "tpm0", Type: DeviceTypeTPM},
				{ID: FlexInt(2), Machine: 42, Name: "usb0", Type: DeviceTypeUSB},
			})
		},
		// Settings endpoints return empty arrays (no settings configured)
		"GET /api/v4/machine_device_settings_tpm": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TPMDeviceSettings{})
		},
		"GET /api/v4/machine_device_settings_usb": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []USBDeviceSettings{})
		},
	}))

	devices, err := client.VMDevices.List(context.Background(), 42)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].Type != DeviceTypeTPM {
		t.Errorf("expected type %q, got %q", DeviceTypeTPM, devices[0].Type)
	}
}

func TestVMDeviceService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				ID:      FlexInt(1),
				Machine: 42,
				Name:    "tpm0",
				Type:    DeviceTypeTPM,
				Enabled: true,
			})
		},
		"GET /api/v4/machine_device_settings_tpm": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TPMDeviceSettings{
				{ID: 10, MachineDevice: 1, Model: "tpm-crb", Version: "2.0"},
			})
		},
	}))

	device, err := client.VMDevices.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if device.Name != "tpm0" {
		t.Errorf("expected name 'tpm0', got %q", device.Name)
	}
	if device.TPMSettings == nil {
		t.Fatal("expected TPMSettings to be loaded")
	}
	if device.TPMSettings.Version != "2.0" {
		t.Errorf("expected TPM version '2.0', got %q", device.TPMSettings.Version)
	}
}

func TestVMDeviceService_Get_USBSettings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				ID:   FlexInt(2),
				Name: "usb0",
				Type: DeviceTypeUSB,
			})
		},
		"GET /api/v4/machine_device_settings_usb": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []USBDeviceSettings{
				{ID: 20, MachineDevice: 2, GuestReset: true},
			})
		},
	}))

	device, err := client.VMDevices.Get(context.Background(), 2)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if device.USBSettings == nil {
		t.Fatal("expected USBSettings to be loaded")
	}
	if !device.USBSettings.GuestReset {
		t.Error("expected GuestReset to be true")
	}
}

func TestVMDeviceService_Get_VGPUSettings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				ID:   FlexInt(3),
				Name: "vgpu0",
				Type: DeviceTypeVGPU,
			})
		},
		"GET /api/v4/machine_device_settings_nvidia_vgpu": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VGPUDeviceSettings{
				{ID: 30, MachineDevice: 3, ProfileType: "nvidia-123"},
			})
		},
	}))

	device, err := client.VMDevices.Get(context.Background(), 3)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if device.VGPUSettings == nil {
		t.Fatal("expected VGPUSettings to be loaded")
	}
	if device.VGPUSettings.ProfileType != "nvidia-123" {
		t.Errorf("expected ProfileType 'nvidia-123', got %q", device.VGPUSettings.ProfileType)
	}
}

func TestVMDeviceService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_devices/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMDevices.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMDeviceService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/machine_devices": func(w http.ResponseWriter, r *http.Request) {
			var req VMDeviceCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Machine != 10 {
				t.Errorf("expected machine 10, got %d", req.Machine)
			}
			if req.Name != "tpm0" {
				t.Errorf("expected name 'tpm0', got %q", req.Name)
			}
			if req.Type != DeviceTypeTPM {
				t.Errorf("expected type %q, got %q", DeviceTypeTPM, req.Type)
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/machine_devices/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				ID:      FlexInt(5),
				Machine: 10,
				Name:    "tpm0",
				Type:    DeviceTypeTPM,
				Enabled: true,
			})
		},
		"GET /api/v4/machine_device_settings_tpm": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TPMDeviceSettings{})
		},
	}))

	device, err := client.VMDevices.Create(context.Background(), 10, &VMDeviceCreateRequest{
		Name: "tpm0",
		Type: DeviceTypeTPM,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if device.ID.Int() != 5 {
		t.Errorf("expected ID 5, got %d", device.ID.Int())
	}
}

func TestVMDeviceService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDevices.Create(context.Background(), 10, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDeviceService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDevices.Create(context.Background(), 10, &VMDeviceCreateRequest{
		Type: DeviceTypeTPM,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDeviceService_Create_MissingType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDevices.Create(context.Background(), 10, &VMDeviceCreateRequest{
		Name: "tpm0",
	})
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDeviceService_Update(t *testing.T) {
	newName := "device-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_devices/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMDeviceUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_devices/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMDevice{
				ID:   FlexInt(1),
				Name: newName,
				Type: DeviceTypeTPM,
			})
		},
		"GET /api/v4/machine_device_settings_tpm": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TPMDeviceSettings{})
		},
	}))

	device, err := client.VMDevices.Update(context.Background(), 1, &VMDeviceUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if device.Name != newName {
		t.Errorf("expected name %q, got %q", newName, device.Name)
	}
}

func TestVMDeviceService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMDevices.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMDeviceService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_devices/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "x"
	_, err := client.VMDevices.Update(context.Background(), 999, &VMDeviceUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMDeviceService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/machine_devices/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMDevices.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVMDeviceService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/machine_devices/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete of already-gone device should succeed (idempotent)
	err := client.VMDevices.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted device")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
