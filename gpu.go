package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// VGPUProfileService reads NVIDIA vGPU profiles (nvidia_vgpu_profiles).
//
// Profiles are produced by the NVIDIA driver. They are the catalog a VM
// vGPU device selects from.
type VGPUProfileService struct {
	client *Client
}

// List returns vGPU profiles, with optional filtering and pagination.
func (s *VGPUProfileService) List(ctx context.Context, opts ...ListOption) ([]VGPUProfile, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vgpuProfileListFields
	}

	var profiles []VGPUProfile
	if err := s.client.get(ctx, "/nvidia_vgpu_profiles", options.toQueryParams(), &profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

// ListByProfileType returns profiles of one NVIDIA type code (A, B, C, or Q).
func (s *VGPUProfileService) ListByProfileType(ctx context.Context, profileType string, opts ...ListOption) ([]VGPUProfile, error) {
	if !validVGPUProfileType(profileType) {
		return nil, &ValidationError{Field: "profile_type", Message: "profile_type must be A, B, C, or Q"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("profile_type eq '%s'", escapeFilterValue(profileType))))
	return s.List(ctx, opts...)
}

// Get returns one vGPU profile by key.
func (s *VGPUProfileService) Get(ctx context.Context, id int) (*VGPUProfile, error) {
	params := url.Values{}
	params.Set("fields", vgpuProfileListFields)

	var profile VGPUProfile
	endpoint := fmt.Sprintf("/nvidia_vgpu_profiles/%d", id)
	if err := s.client.get(ctx, endpoint, params, &profile); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VGPUProfile", ID: id}
		}
		return nil, err
	}
	return &profile, nil
}

// GetByName returns the vGPU profile with this name.
func (s *VGPUProfileService) GetByName(ctx context.Context, name string) (*VGPUProfile, error) {
	profiles, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, &NotFoundError{Resource: "VGPUProfile", ID: name}
	}
	if err := requireUniqueName("VGPUProfile", name, profiles, func(p VGPUProfile) any { return p.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("VGPUProfile", name, profiles[0].Name, name); err != nil {
		return nil, err
	}
	return &profiles[0], nil
}

// NodeGPUService reads and updates node GPU configuration (node_gpus).
type NodeGPUService struct {
	client *Client
}

// List returns node GPUs, with optional filtering and pagination.
func (s *NodeGPUService) List(ctx context.Context, opts ...ListOption) ([]NodeGPU, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeGPUListFields
	}

	var gpus []NodeGPU
	if err := s.client.get(ctx, "/node_gpus", options.toQueryParams(), &gpus); err != nil {
		return nil, err
	}
	return gpus, nil
}

// ListByNode returns GPUs configured on one node.
func (s *NodeGPUService) ListByNode(ctx context.Context, nodeID int, opts ...ListOption) ([]NodeGPU, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node eq %d", nodeID)))
	return s.List(ctx, opts...)
}

// ListByMode returns GPUs in one mode.
// mode is GPUModeNone, GPUModePassthrough, or GPUModeNVIDIAvGPU.
func (s *NodeGPUService) ListByMode(ctx context.Context, mode string, opts ...ListOption) ([]NodeGPU, error) {
	if !validGPUMode(mode) {
		return nil, &ValidationError{Field: "mode", Message: "mode must be none, gpu, or nvidia_vgpu"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("mode eq '%s'", escapeFilterValue(mode))))
	return s.List(ctx, opts...)
}

// ListEnabled returns GPUs that have a mode other than GPUModeNone.
func (s *NodeGPUService) ListEnabled(ctx context.Context, opts ...ListOption) ([]NodeGPU, error) {
	opts = append(opts, WithFilter("mode ne 'none'"))
	return s.List(ctx, opts...)
}

// Get returns one node GPU by key.
func (s *NodeGPUService) Get(ctx context.Context, id int) (*NodeGPU, error) {
	params := url.Values{}
	params.Set("fields", nodeGPUListFields)

	var gpu NodeGPU
	endpoint := fmt.Sprintf("/node_gpus/%d", id)
	if err := s.client.get(ctx, endpoint, params, &gpu); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeGPU", ID: id}
		}
		return nil, err
	}
	return &gpu, nil
}

// GetByName returns the node GPU with this name.
// The same name can exist on more than one node. That match is AmbiguousNameError.
// Use GetByNodeAndName when the node is known.
func (s *NodeGPUService) GetByName(ctx context.Context, name string) (*NodeGPU, error) {
	return s.oneByName(ctx, name, fmt.Sprintf("name eq '%s'", escapeFilterValue(name)))
}

// GetByNodeAndName returns the GPU with this name on one node.
func (s *NodeGPUService) GetByNodeAndName(ctx context.Context, nodeID int, name string) (*NodeGPU, error) {
	filter := fmt.Sprintf("node eq %d and name eq '%s'", nodeID, escapeFilterValue(name))
	return s.oneByName(ctx, name, filter)
}

func (s *NodeGPUService) oneByName(ctx context.Context, name, filter string) (*NodeGPU, error) {
	gpus, err := s.List(ctx, WithFilter(filter))
	if err != nil {
		return nil, err
	}
	if len(gpus) == 0 {
		return nil, &NotFoundError{Resource: "NodeGPU", ID: name}
	}
	if err := requireUniqueName("NodeGPU", name, gpus, func(g NodeGPU) any { return g.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("NodeGPU", name, gpus[0].Name, name); err != nil {
		return nil, err
	}
	return &gpus[0], nil
}

// Update changes a node GPU and returns the updated row.
//
// Mode must be GPUModeNone, GPUModePassthrough, or GPUModeNVIDIAvGPU.
// Set NVIDIAVgpuProfile when switching to GPUModeNVIDIAvGPU.
func (s *NodeGPUService) Update(ctx context.Context, id int, req *NodeGPUUpdateRequest) (*NodeGPU, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if req.Mode != nil && !validGPUMode(*req.Mode) {
		return nil, &ValidationError{Field: "mode", Message: "mode must be none, gpu, or nvidia_vgpu"}
	}

	endpoint := fmt.Sprintf("/node_gpus/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeGPU", ID: id}
		}
		return nil, err
	}
	return s.Get(ctx, id)
}

// NodeGPUStatsService reads current and historical GPU utilization (node_gpu_stats).
type NodeGPUStatsService struct {
	client *Client
}

// List returns current GPU stats rows, with optional filtering and pagination.
func (s *NodeGPUStatsService) List(ctx context.Context, opts ...ListOption) ([]NodeGPUStats, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeGPUStatsListFields
	}

	var stats []NodeGPUStats
	if err := s.client.get(ctx, "/node_gpu_stats", options.toQueryParams(), &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// Get returns one current stats row by key.
func (s *NodeGPUStatsService) Get(ctx context.Context, id int) (*NodeGPUStats, error) {
	return s.get(ctx, "/node_gpu_stats", id)
}

// GetByGPU returns the current stats row for a node GPU.
func (s *NodeGPUStatsService) GetByGPU(ctx context.Context, gpuID int) (*NodeGPUStats, error) {
	stats, err := s.List(ctx, WithFilter(fmt.Sprintf("node_gpu eq %d", gpuID)), WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(stats) == 0 {
		return nil, &NotFoundError{Resource: "NodeGPUStats", ID: fmt.Sprintf("node_gpu=%d", gpuID)}
	}
	return &stats[0], nil
}

// ListHistoryShort returns short-term history, newest first unless a sort is set.
func (s *NodeGPUStatsService) ListHistoryShort(ctx context.Context, opts ...ListOption) ([]NodeGPUStats, error) {
	return s.listHistory(ctx, "/node_gpu_stats_history_short", opts)
}

// ListHistoryShortByGPU returns short-term history for one GPU, newest first unless a sort is set.
func (s *NodeGPUStatsService) ListHistoryShortByGPU(ctx context.Context, gpuID int, opts ...ListOption) ([]NodeGPUStats, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node_gpu eq %d", gpuID)))
	return s.ListHistoryShort(ctx, opts...)
}

// ListHistoryLong returns long-term history, newest first unless a sort is set.
func (s *NodeGPUStatsService) ListHistoryLong(ctx context.Context, opts ...ListOption) ([]NodeGPUStats, error) {
	return s.listHistory(ctx, "/node_gpu_stats_history_long", opts)
}

// ListHistoryLongByGPU returns long-term history for one GPU, newest first unless a sort is set.
func (s *NodeGPUStatsService) ListHistoryLongByGPU(ctx context.Context, gpuID int, opts ...ListOption) ([]NodeGPUStats, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node_gpu eq %d", gpuID)))
	return s.ListHistoryLong(ctx, opts...)
}

// GetHistoryShort returns one short-term history row by key.
func (s *NodeGPUStatsService) GetHistoryShort(ctx context.Context, id int) (*NodeGPUStats, error) {
	return s.get(ctx, "/node_gpu_stats_history_short", id)
}

// GetHistoryLong returns one long-term history row by key.
func (s *NodeGPUStatsService) GetHistoryLong(ctx context.Context, id int) (*NodeGPUStats, error) {
	return s.get(ctx, "/node_gpu_stats_history_long", id)
}

func (s *NodeGPUStatsService) listHistory(ctx context.Context, endpoint string, opts []ListOption) ([]NodeGPUStats, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeGPUStatsListFields
	}
	if options.Sort == "" {
		options.Sort = "-timestamp"
	}

	var stats []NodeGPUStats
	if err := s.client.get(ctx, endpoint, options.toQueryParams(), &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *NodeGPUStatsService) get(ctx context.Context, path string, id int) (*NodeGPUStats, error) {
	params := url.Values{}
	params.Set("fields", nodeGPUStatsListFields)

	var stats NodeGPUStats
	endpoint := fmt.Sprintf("%s/%d", path, id)
	if err := s.client.get(ctx, endpoint, params, &stats); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeGPUStats", ID: id}
		}
		return nil, err
	}
	return &stats, nil
}

// NodeGPUInstanceService reads GPU instances assigned to VMs (node_gpu_instances).
type NodeGPUInstanceService struct {
	client *Client
}

// List returns GPU instances, with optional filtering and pagination.
func (s *NodeGPUInstanceService) List(ctx context.Context, opts ...ListOption) ([]NodeGPUInstance, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeGPUInstanceListFields
	}

	var instances []NodeGPUInstance
	if err := s.client.get(ctx, "/node_gpu_instances", options.toQueryParams(), &instances); err != nil {
		return nil, err
	}
	return instances, nil
}

// ListByGPU returns instances assigned from one node GPU.
func (s *NodeGPUInstanceService) ListByGPU(ctx context.Context, gpuID int, opts ...ListOption) ([]NodeGPUInstance, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("gpu eq %d", gpuID)))
	return s.List(ctx, opts...)
}

// Get returns one GPU instance by key.
func (s *NodeGPUInstanceService) Get(ctx context.Context, id int) (*NodeGPUInstance, error) {
	params := url.Values{}
	params.Set("fields", nodeGPUInstanceListFields)

	var instance NodeGPUInstance
	endpoint := fmt.Sprintf("/node_gpu_instances/%d", id)
	if err := s.client.get(ctx, endpoint, params, &instance); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeGPUInstance", ID: id}
		}
		return nil, err
	}
	return &instance, nil
}

// NodeVGPUDeviceService reads detected NVIDIA vGPU devices (node_nvidia_vgpu_devices).
type NodeVGPUDeviceService struct {
	client *Client
}

// List returns vGPU-capable devices, with optional filtering and pagination.
func (s *NodeVGPUDeviceService) List(ctx context.Context, opts ...ListOption) ([]NodeVGPUDevice, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeVGPUDeviceListFields
	}

	var devices []NodeVGPUDevice
	if err := s.client.get(ctx, "/node_nvidia_vgpu_devices", options.toQueryParams(), &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

// ListByNode returns vGPU-capable devices on one node.
func (s *NodeVGPUDeviceService) ListByNode(ctx context.Context, nodeID int, opts ...ListOption) ([]NodeVGPUDevice, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node eq %d", nodeID)))
	return s.List(ctx, opts...)
}

// ListByVendor returns devices whose vendor contains vendor.
func (s *NodeVGPUDeviceService) ListByVendor(ctx context.Context, vendor string, opts ...ListOption) ([]NodeVGPUDevice, error) {
	if vendor == "" {
		return nil, &ValidationError{Field: "vendor", Message: "vendor is required"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("vendor ct '%s'", escapeFilterValue(vendor))))
	return s.List(ctx, opts...)
}

// Get returns one vGPU-capable device by key.
func (s *NodeVGPUDeviceService) Get(ctx context.Context, id int) (*NodeVGPUDevice, error) {
	params := url.Values{}
	params.Set("fields", nodeVGPUDeviceListFields)

	var device NodeVGPUDevice
	endpoint := fmt.Sprintf("/node_nvidia_vgpu_devices/%d", id)
	if err := s.client.get(ctx, endpoint, params, &device); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeVGPUDevice", ID: id}
		}
		return nil, err
	}
	return &device, nil
}

// NodeHostGPUDeviceService reads detected passthrough GPUs (node_host_gpu_devices).
type NodeHostGPUDeviceService struct {
	client *Client
}

// List returns host GPU devices, with optional filtering and pagination.
func (s *NodeHostGPUDeviceService) List(ctx context.Context, opts ...ListOption) ([]NodeHostGPUDevice, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeHostGPUDeviceListFields
	}

	var devices []NodeHostGPUDevice
	if err := s.client.get(ctx, "/node_host_gpu_devices", options.toQueryParams(), &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

// ListByNode returns host GPU devices on one node.
func (s *NodeHostGPUDeviceService) ListByNode(ctx context.Context, nodeID int, opts ...ListOption) ([]NodeHostGPUDevice, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("node eq %d", nodeID)))
	return s.List(ctx, opts...)
}

// ListByVendor returns devices whose vendor contains vendor.
func (s *NodeHostGPUDeviceService) ListByVendor(ctx context.Context, vendor string, opts ...ListOption) ([]NodeHostGPUDevice, error) {
	if vendor == "" {
		return nil, &ValidationError{Field: "vendor", Message: "vendor is required"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("vendor ct '%s'", escapeFilterValue(vendor))))
	return s.List(ctx, opts...)
}

// Get returns one host GPU device by key.
func (s *NodeHostGPUDeviceService) Get(ctx context.Context, id int) (*NodeHostGPUDevice, error) {
	params := url.Values{}
	params.Set("fields", nodeHostGPUDeviceListFields)

	var device NodeHostGPUDevice
	endpoint := fmt.Sprintf("/node_host_gpu_devices/%d", id)
	if err := s.client.get(ctx, endpoint, params, &device); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeHostGPUDevice", ID: id}
		}
		return nil, err
	}
	return &device, nil
}

// NodeVGPUProfileService reads vGPU profiles available on a physical GPU
// (node_nvidia_vgpu_profiles).
type NodeVGPUProfileService struct {
	client *Client
}

// List returns per-GPU vGPU profiles, with optional filtering and pagination.
func (s *NodeVGPUProfileService) List(ctx context.Context, opts ...ListOption) ([]NodeVGPUProfile, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = nodeVGPUProfileListFields
	}

	var profiles []NodeVGPUProfile
	if err := s.client.get(ctx, "/node_nvidia_vgpu_profiles", options.toQueryParams(), &profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

// ListByPhysicalGPU returns profiles for one physical GPU.
func (s *NodeVGPUProfileService) ListByPhysicalGPU(ctx context.Context, pciDeviceID int, opts ...ListOption) ([]NodeVGPUProfile, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("physical_gpu eq %d", pciDeviceID)))
	return s.List(ctx, opts...)
}

// ListByProfileType returns profiles of one NVIDIA type code (A, B, C, or Q).
func (s *NodeVGPUProfileService) ListByProfileType(ctx context.Context, profileType string, opts ...ListOption) ([]NodeVGPUProfile, error) {
	if !validVGPUProfileType(profileType) {
		return nil, &ValidationError{Field: "profile_type", Message: "profile_type must be A, B, C, or Q"}
	}
	opts = append(opts, WithFilter(fmt.Sprintf("profile_type eq '%s'", escapeFilterValue(profileType))))
	return s.List(ctx, opts...)
}

// Get returns one per-GPU vGPU profile by key.
func (s *NodeVGPUProfileService) Get(ctx context.Context, id int) (*NodeVGPUProfile, error) {
	params := url.Values{}
	params.Set("fields", nodeVGPUProfileListFields)

	var profile NodeVGPUProfile
	endpoint := fmt.Sprintf("/node_nvidia_vgpu_profiles/%d", id)
	if err := s.client.get(ctx, endpoint, params, &profile); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "NodeVGPUProfile", ID: id}
		}
		return nil, err
	}
	return &profile, nil
}

// GetByName returns the per-GPU vGPU profile with this name.
// The same name can exist on more than one physical GPU. That match is AmbiguousNameError.
func (s *NodeVGPUProfileService) GetByName(ctx context.Context, name string) (*NodeVGPUProfile, error) {
	return s.oneByName(ctx, name, fmt.Sprintf("name eq '%s'", escapeFilterValue(name)))
}

// GetByPhysicalGPUAndName returns the profile with this name on one physical GPU.
func (s *NodeVGPUProfileService) GetByPhysicalGPUAndName(ctx context.Context, pciDeviceID int, name string) (*NodeVGPUProfile, error) {
	filter := fmt.Sprintf("physical_gpu eq %d and name eq '%s'", pciDeviceID, escapeFilterValue(name))
	return s.oneByName(ctx, name, filter)
}

func (s *NodeVGPUProfileService) oneByName(ctx context.Context, name, filter string) (*NodeVGPUProfile, error) {
	profiles, err := s.List(ctx, WithFilter(filter))
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, &NotFoundError{Resource: "NodeVGPUProfile", ID: name}
	}
	if err := requireUniqueName("NodeVGPUProfile", name, profiles, func(p NodeVGPUProfile) any { return p.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("NodeVGPUProfile", name, profiles[0].Name, name); err != nil {
		return nil, err
	}
	return &profiles[0], nil
}
