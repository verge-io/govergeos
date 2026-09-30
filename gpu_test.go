package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVGPUProfileService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nvidia_vgpu_profiles": func(w http.ResponseWriter, r *http.Request) {
			if fields := r.URL.Query().Get("fields"); fields != vgpuProfileListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []VGPUProfile{{
				Key: 1, Name: "nvidia-256", TypeID: 45, DeviceHex: "10de:1eb8",
				Framebuffer: "256M", ProfileType: VGPUProfileTypeCompute, VirtualFunction: false,
			}})
		},
	}))

	profiles, err := client.VGPUProfiles.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "nvidia-256" {
		t.Errorf("expected name nvidia-256, got %q", profiles[0].Name)
	}
	if profiles[0].ProfileTypeDisplay() != "AI/Machine Learning/Training (vCS or vWS)" {
		t.Errorf("unexpected profile type display %q", profiles[0].ProfileTypeDisplay())
	}
}

func TestVGPUProfileService_ListByProfileType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nvidia_vgpu_profiles": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "profile_type eq 'C'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VGPUProfile{{Key: 1, Name: "nvidia-256", ProfileType: "C"}})
		},
	}))

	profiles, err := client.VGPUProfiles.ListByProfileType(context.Background(), VGPUProfileTypeCompute)
	if err != nil {
		t.Fatalf("ListByProfileType failed: %v", err)
	}
	if len(profiles) != 1 || profiles[0].ProfileType != "C" {
		t.Fatalf("unexpected profiles: %+v", profiles)
	}
}

func TestVGPUProfileService_ListByProfileType_Invalid(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.VGPUProfiles.ListByProfileType(context.Background(), "Z")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestVGPUProfileService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nvidia_vgpu_profiles": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "name eq 'nvidia-256'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VGPUProfile{{Key: 1, Name: "nvidia-256", TypeID: 45}})
		},
	}))

	profile, err := client.VGPUProfiles.GetByName(context.Background(), "nvidia-256")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if profile.TypeID != 45 {
		t.Errorf("expected type id 45, got %d", profile.TypeID)
	}
}

func TestVGPUProfileService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/nvidia_vgpu_profiles/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VGPUProfiles.Get(context.Background(), 999)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestNodeGPU_ModeHelpers(t *testing.T) {
	gpu := NodeGPU{Mode: GPUModeNVIDIAvGPU}
	if !gpu.IsVGPU() || gpu.IsPassthrough() || gpu.IsDisabled() {
		t.Fatalf("vGPU mode flags: vgpu=%v passthrough=%v disabled=%v", gpu.IsVGPU(), gpu.IsPassthrough(), gpu.IsDisabled())
	}
	if gpu.ModeDisplay() != "NVIDIA vGPU" {
		t.Errorf("mode display %q", gpu.ModeDisplay())
	}

	gpu.Mode = GPUModePassthrough
	if !gpu.IsPassthrough() || gpu.ModeDisplay() != "PCI Passthrough" {
		t.Errorf("passthrough: %+v display %q", gpu, gpu.ModeDisplay())
	}

	gpu.Mode = GPUModeNone
	if !gpu.IsDisabled() || gpu.ModeDisplay() != "None" {
		t.Errorf("disabled: %+v display %q", gpu, gpu.ModeDisplay())
	}
}

func TestNodeGPUService_ListByNode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpus": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node eq 2" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != nodeGPUListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []NodeGPU{{
				Key: 1, Name: "GPU_1", Node: 2, NodeName: "node2",
				Mode: GPUModeNVIDIAvGPU, NVIDIAVgpuProfile: 1,
				NVIDIAVgpuProfileDisplay: "nvidia-256", MaxInstances: 24, InstancesCount: 4,
			}})
		},
	}))

	gpus, err := client.NodeGPUs.ListByNode(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListByNode failed: %v", err)
	}
	if len(gpus) != 1 || gpus[0].Name != "GPU_1" || gpus[0].InstancesCount != 4 {
		t.Fatalf("unexpected gpus: %+v", gpus)
	}
}

func TestNodeGPUService_ListEnabled(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpus": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "mode ne 'none'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeGPU{})
		},
	}))

	gpus, err := client.NodeGPUs.ListEnabled(context.Background())
	if err != nil {
		t.Fatalf("ListEnabled failed: %v", err)
	}
	if len(gpus) != 0 {
		t.Fatalf("expected empty list, got %d", len(gpus))
	}
}

func TestNodeGPUService_GetByName_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpus": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NodeGPU{
				{Key: 1, Name: "GPU_1", Node: 2},
				{Key: 2, Name: "GPU_1", Node: 3},
			})
		},
	}))

	_, err := client.NodeGPUs.GetByName(context.Background(), "GPU_1")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("expected AmbiguousNameError, got %v", err)
	}
}

func TestNodeGPUService_GetByNodeAndName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpus": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node eq 2 and name eq 'GPU_1'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeGPU{{Key: 1, Name: "GPU_1", Node: 2, Mode: GPUModePassthrough}})
		},
	}))

	gpu, err := client.NodeGPUs.GetByNodeAndName(context.Background(), 2, "GPU_1")
	if err != nil {
		t.Fatalf("GetByNodeAndName failed: %v", err)
	}
	if !gpu.IsPassthrough() {
		t.Errorf("expected passthrough, got mode %q", gpu.Mode)
	}
}

func TestNodeGPUService_Update(t *testing.T) {
	profile := 7
	mode := GPUModeNVIDIAvGPU
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/node_gpus/1": func(w http.ResponseWriter, r *http.Request) {
			var body NodeGPUUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode body: %v", err)
			}
			if body.Mode == nil || *body.Mode != GPUModeNVIDIAvGPU {
				t.Errorf("mode = %v", body.Mode)
			}
			if body.NVIDIAVgpuProfile == nil || *body.NVIDIAVgpuProfile != 7 {
				t.Errorf("profile = %v", body.NVIDIAVgpuProfile)
			}
			jsonResponse(w, 200, map[string]any{})
		},
		"GET /api/v4/node_gpus/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NodeGPU{Key: 1, Name: "GPU_1", Mode: GPUModeNVIDIAvGPU, NVIDIAVgpuProfile: 7})
		},
	}))

	gpu, err := client.NodeGPUs.Update(context.Background(), 1, &NodeGPUUpdateRequest{
		Mode:              &mode,
		NVIDIAVgpuProfile: &profile,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if gpu.NVIDIAVgpuProfile != 7 || !gpu.IsVGPU() {
		t.Fatalf("unexpected gpu: %+v", gpu)
	}
}

func TestNodeGPUService_Update_InvalidMode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/node_gpus/1": func(w http.ResponseWriter, r *http.Request) {
			t.Error("PUT should not be sent for an invalid mode")
		},
	}))
	bad := "sriov"
	_, err := client.NodeGPUs.Update(context.Background(), 1, &NodeGPUUpdateRequest{Mode: &bad})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestNodeGPUStatsService_GetByGPU(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpu_stats": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node_gpu eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if r.URL.Query().Get("limit") != "1" {
				t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
			}
			jsonResponse(w, 200, []NodeGPUStats{{
				Key: 1, NodeGPU: 1, GPUsTotal: 1, GPUs: 1, VGPUsTotal: 24, VGPUs: 4, VGPUsIdle: 20,
			}})
		},
	}))

	stats, err := client.NodeGPUStats.GetByGPU(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByGPU failed: %v", err)
	}
	if stats.VGPUs != 4 || stats.VGPUsTotal != 24 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestNodeGPUStatsService_GetByGPU_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpu_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NodeGPUStats{})
		},
	}))
	_, err := client.NodeGPUStats.GetByGPU(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestNodeGPUStatsService_ListHistoryShortByGPU(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpu_stats_history_short": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node_gpu eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if sort := r.URL.Query().Get("sort"); sort != "-timestamp" {
				t.Errorf("unexpected sort: %s", sort)
			}
			jsonResponse(w, 200, []NodeGPUStats{{Key: 4, NodeGPU: 1, Timestamp: 1704067200, VGPUs: 4}})
		},
	}))

	rows, err := client.NodeGPUStats.ListHistoryShortByGPU(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListHistoryShortByGPU failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Timestamp != 1704067200 {
		t.Fatalf("unexpected history: %+v", rows)
	}
}

func TestNodeGPUInstanceService_ListByGPU(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_gpu_instances": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "gpu eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != nodeGPUInstanceListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []NodeGPUInstance{{
				Key: 1, GPUKey: 1, GPUName: "GPU_1", MachineName: "ai-workload-1", Mode: GPUModeNVIDIAvGPU,
			}})
		},
	}))

	instances, err := client.NodeGPUInstances.ListByGPU(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByGPU failed: %v", err)
	}
	if len(instances) != 1 || instances[0].MachineName != "ai-workload-1" {
		t.Fatalf("unexpected instances: %+v", instances)
	}
}

func TestNodeVGPUDeviceService_ListByVendor(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_nvidia_vgpu_devices": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "vendor ct 'NVIDIA'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeVGPUDevice{{
				Key: 1, Node: 2, NodeName: "node2", Slot: "0000:41:00.0", Vendor: "NVIDIA Corporation", MaxInstances: 24,
			}})
		},
	}))

	devices, err := client.NodeVGPUDevices.ListByVendor(context.Background(), "NVIDIA")
	if err != nil {
		t.Fatalf("ListByVendor failed: %v", err)
	}
	if len(devices) != 1 || devices[0].Slot != "0000:41:00.0" {
		t.Fatalf("unexpected devices: %+v", devices)
	}
}

func TestNodeHostGPUDeviceService_ListByNode(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_host_gpu_devices": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "node eq 2" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeHostGPUDevice{{
				Key: 1, Node: 2, Name: "RTX A6000", Driver: "vfio-pci", MaxInstances: 1,
			}})
		},
	}))

	devices, err := client.NodeHostGPUDevices.ListByNode(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListByNode failed: %v", err)
	}
	if len(devices) != 1 || devices[0].Driver != "vfio-pci" {
		t.Fatalf("unexpected devices: %+v", devices)
	}
}

func TestNodeHostGPUDeviceService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_host_gpu_devices/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.NodeHostGPUDevices.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestNodeVGPUProfileService_ListByPhysicalGPU(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_nvidia_vgpu_profiles": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "physical_gpu eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []NodeVGPUProfile{{
				Key: 1, PhysicalGPU: 10, Name: "nvidia-256", ProfileType: "C",
				AvailableInstances: 20, MaxInstance: 24, Framebuffer: "256M",
			}})
		},
	}))

	profiles, err := client.NodeVGPUProfiles.ListByPhysicalGPU(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByPhysicalGPU failed: %v", err)
	}
	if len(profiles) != 1 || profiles[0].AvailableInstances != 20 {
		t.Fatalf("unexpected profiles: %+v", profiles)
	}
	if profiles[0].ProfileTypeDisplay() != "AI/Machine Learning/Training (vCS or vWS)" {
		t.Errorf("unexpected display %q", profiles[0].ProfileTypeDisplay())
	}
}

func TestNodeVGPUProfileService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/node_nvidia_vgpu_profiles": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []NodeVGPUProfile{})
		},
	}))
	_, err := client.NodeVGPUProfiles.GetByName(context.Background(), "missing")
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
