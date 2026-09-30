package vergeos

// BillingRecord is one system-wide usage sample (billing).
//
// Records are written by VergeOS. RAM fields are megabytes. Physical RAM,
// VRAM, and tier storage are bytes. Generate posts a new sample.
type BillingRecord struct {
	// Key is the billing row key.
	Key FlexInt `json:"$key,omitempty"`
	// Created is when the record was written (Unix epoch).
	Created int64 `json:"created,omitempty"`
	// From is the start of the billing period (Unix epoch).
	From int64 `json:"from,omitempty"`
	// To is the end of the billing period (Unix epoch).
	To int64 `json:"to,omitempty"`
	// Sent is when the record was reported (Unix epoch).
	Sent int64 `json:"sent,omitempty"`
	// Description is the record description.
	Description string `json:"description,omitempty"`

	// UsedCores is the number of CPU cores in use.
	UsedCores int `json:"used_cores,omitempty"`
	// TotalCores is the number of CPU cores installed.
	TotalCores int `json:"total_cores,omitempty"`
	// OnlineCores is the number of CPU cores online.
	OnlineCores int `json:"online_cores,omitempty"`
	// PhysTotalCPU is the physical CPU percentage.
	PhysTotalCPU int `json:"phys_total_cpu,omitempty"`

	// TotalNodes is the number of nodes.
	TotalNodes int `json:"total_nodes,omitempty"`
	// OnlineNodes is the number of nodes online.
	OnlineNodes int `json:"online_nodes,omitempty"`
	// RunningMachines is the number of running VMs.
	RunningMachines int `json:"running_machines,omitempty"`

	// UsedRAM is RAM in use, in megabytes.
	UsedRAM int64 `json:"used_ram,omitempty"`
	// TotalRAM is installed RAM, in megabytes.
	TotalRAM int64 `json:"total_ram,omitempty"`
	// OnlineRAM is RAM online, in megabytes.
	OnlineRAM int64 `json:"online_ram,omitempty"`
	// PhysRAMUsed is physical RAM in use, in bytes.
	PhysRAMUsed int64 `json:"phys_ram_used,omitempty"`
	// PhysVRAMUsed is physical VRAM in use, in bytes.
	PhysVRAMUsed int64 `json:"phys_vram_used,omitempty"`

	// Tier0Used is tier 0 storage in use, in bytes.
	Tier0Used int64 `json:"tier_0_used,omitempty"`
	// Tier0Total is tier 0 capacity, in bytes.
	Tier0Total int64 `json:"tier_0_total,omitempty"`
	// Tier1Used is tier 1 storage in use, in bytes.
	Tier1Used int64 `json:"tier_1_used,omitempty"`
	// Tier1Total is tier 1 capacity, in bytes.
	Tier1Total int64 `json:"tier_1_total,omitempty"`
	// Tier2Used is tier 2 storage in use, in bytes.
	Tier2Used int64 `json:"tier_2_used,omitempty"`
	// Tier2Total is tier 2 capacity, in bytes.
	Tier2Total int64 `json:"tier_2_total,omitempty"`
	// Tier3Used is tier 3 storage in use, in bytes.
	Tier3Used int64 `json:"tier_3_used,omitempty"`
	// Tier3Total is tier 3 capacity, in bytes.
	Tier3Total int64 `json:"tier_3_total,omitempty"`
	// Tier4Used is tier 4 storage in use, in bytes.
	Tier4Used int64 `json:"tier_4_used,omitempty"`
	// Tier4Total is tier 4 capacity, in bytes.
	Tier4Total int64 `json:"tier_4_total,omitempty"`
	// Tier5Used is tier 5 storage in use, in bytes.
	Tier5Used int64 `json:"tier_5_used,omitempty"`
	// Tier5Total is tier 5 capacity, in bytes.
	Tier5Total int64 `json:"tier_5_total,omitempty"`

	// GPUsTotal is the number of physical GPUs.
	GPUsTotal int `json:"gpus_total,omitempty"`
	// GPUs is the number of physical GPUs in use.
	GPUs int `json:"gpus,omitempty"`
	// GPUsIdle is the number of idle physical GPUs.
	GPUsIdle int `json:"gpus_idle,omitempty"`
	// VGPUsTotal is the number of vGPUs.
	VGPUsTotal int `json:"vgpus_total,omitempty"`
	// VGPUs is the number of vGPUs in use.
	VGPUs int `json:"vgpus,omitempty"`
	// VGPUsIdle is the number of idle vGPUs.
	VGPUsIdle int `json:"vgpus_idle,omitempty"`

	// WorkloadDatapoints is the number of workload samples in the record.
	WorkloadDatapoints int `json:"workload_datapoints,omitempty"`
	// StorageDatapoints is the number of storage samples in the record.
	StorageDatapoints int `json:"storage_datapoints,omitempty"`
}

// BillingSummary is average and peak usage across billing records.
//
// Totals for GPUs come from the newest record in the range.
type BillingSummary struct {
	// RecordCount is the number of billing records in the range.
	RecordCount int
	// AvgCPUUtilization is the average of used cores over total cores, in percent.
	AvgCPUUtilization float64
	// PeakCPUCores is the highest used-core count in the range.
	PeakCPUCores int
	// AvgRAMUtilization is the average of used RAM over total RAM, in percent.
	AvgRAMUtilization float64
	// PeakRAMGB is the highest used RAM in the range, in gigabytes.
	PeakRAMGB float64
	// AvgStorageUsedGB is the average storage used across all tiers, in gigabytes.
	AvgStorageUsedGB float64
	// PeakStorageUsedGB is the highest storage used across all tiers, in gigabytes.
	PeakStorageUsedGB float64
	// TotalGPUs is the physical GPU count on the newest record.
	TotalGPUs int
	// AvgGPUsUsed is the average number of physical GPUs in use.
	AvgGPUsUsed float64
	// TotalVGPUs is the vGPU count on the newest record.
	TotalVGPUs int
	// AvgVGPUsUsed is the average number of vGPUs in use.
	AvgVGPUsUsed float64
}

const (
	billingMBPerGB    = 1024
	billingBytesPerGB = 1024 * 1024 * 1024

	billingListFields = "$key,created,from,to,sent,description," +
		"used_cores,total_cores,online_cores,total_nodes,online_nodes,running_machines," +
		"used_ram,total_ram,online_ram,phys_ram_used,phys_vram_used,phys_total_cpu," +
		"tier_0_used,tier_0_total,tier_1_used,tier_1_total,tier_2_used,tier_2_total," +
		"tier_3_used,tier_3_total,tier_4_used,tier_4_total,tier_5_used,tier_5_total," +
		"gpus_total,gpus,gpus_idle,vgpus_total,vgpus,vgpus_idle," +
		"workload_datapoints,storage_datapoints"
)

// UsedRAMGB returns used RAM in gigabytes.
func (r BillingRecord) UsedRAMGB() float64 {
	return float64(r.UsedRAM) / billingMBPerGB
}

// TotalRAMGB returns installed RAM in gigabytes.
func (r BillingRecord) TotalRAMGB() float64 {
	return float64(r.TotalRAM) / billingMBPerGB
}

// OnlineRAMGB returns online RAM in gigabytes.
func (r BillingRecord) OnlineRAMGB() float64 {
	return float64(r.OnlineRAM) / billingMBPerGB
}

// PhysRAMUsedGB returns physical RAM in use, in gigabytes.
func (r BillingRecord) PhysRAMUsedGB() float64 {
	return float64(r.PhysRAMUsed) / billingBytesPerGB
}

// PhysVRAMUsedGB returns physical VRAM in use, in gigabytes.
func (r BillingRecord) PhysVRAMUsedGB() float64 {
	return float64(r.PhysVRAMUsed) / billingBytesPerGB
}

// TierUsed returns storage in use for tier 0 through 5, in bytes.
func (r BillingRecord) TierUsed(tier int) (int64, error) {
	switch tier {
	case 0:
		return r.Tier0Used, nil
	case 1:
		return r.Tier1Used, nil
	case 2:
		return r.Tier2Used, nil
	case 3:
		return r.Tier3Used, nil
	case 4:
		return r.Tier4Used, nil
	case 5:
		return r.Tier5Used, nil
	default:
		return 0, &ValidationError{Field: "tier", Message: "tier must be 0-5"}
	}
}

// TierTotal returns capacity for tier 0 through 5, in bytes.
func (r BillingRecord) TierTotal(tier int) (int64, error) {
	switch tier {
	case 0:
		return r.Tier0Total, nil
	case 1:
		return r.Tier1Total, nil
	case 2:
		return r.Tier2Total, nil
	case 3:
		return r.Tier3Total, nil
	case 4:
		return r.Tier4Total, nil
	case 5:
		return r.Tier5Total, nil
	default:
		return 0, &ValidationError{Field: "tier", Message: "tier must be 0-5"}
	}
}

// TierUsedGB returns storage in use for tier 0 through 5, in gigabytes.
func (r BillingRecord) TierUsedGB(tier int) (float64, error) {
	used, err := r.TierUsed(tier)
	if err != nil {
		return 0, err
	}
	return float64(used) / billingBytesPerGB, nil
}

// TierTotalGB returns capacity for tier 0 through 5, in gigabytes.
func (r BillingRecord) TierTotalGB(tier int) (float64, error) {
	total, err := r.TierTotal(tier)
	if err != nil {
		return 0, err
	}
	return float64(total) / billingBytesPerGB, nil
}

// TotalStorageUsed returns storage in use across tiers 0 through 5, in bytes.
func (r BillingRecord) TotalStorageUsed() int64 {
	return r.Tier0Used + r.Tier1Used + r.Tier2Used + r.Tier3Used + r.Tier4Used + r.Tier5Used
}

// TotalStorageUsedGB returns storage in use across tiers 0 through 5, in gigabytes.
func (r BillingRecord) TotalStorageUsedGB() float64 {
	return float64(r.TotalStorageUsed()) / billingBytesPerGB
}

// TotalStorageTotal returns capacity across tiers 0 through 5, in bytes.
func (r BillingRecord) TotalStorageTotal() int64 {
	return r.Tier0Total + r.Tier1Total + r.Tier2Total + r.Tier3Total + r.Tier4Total + r.Tier5Total
}

// TotalStorageTotalGB returns capacity across tiers 0 through 5, in gigabytes.
func (r BillingRecord) TotalStorageTotalGB() float64 {
	return float64(r.TotalStorageTotal()) / billingBytesPerGB
}

// CPUUtilization returns used cores as a percentage of total cores.
func (r BillingRecord) CPUUtilization() float64 {
	if r.TotalCores == 0 {
		return 0
	}
	return float64(r.UsedCores) / float64(r.TotalCores) * 100
}

// RAMUtilization returns used RAM as a percentage of total RAM.
func (r BillingRecord) RAMUtilization() float64 {
	if r.TotalRAM == 0 {
		return 0
	}
	return float64(r.UsedRAM) / float64(r.TotalRAM) * 100
}

// GPUUtilization returns physical GPUs in use as a percentage of the total.
func (r BillingRecord) GPUUtilization() float64 {
	if r.GPUsTotal == 0 {
		return 0
	}
	return float64(r.GPUs) / float64(r.GPUsTotal) * 100
}

// VGPUUtilization returns vGPUs in use as a percentage of the total.
func (r BillingRecord) VGPUUtilization() float64 {
	if r.VGPUsTotal == 0 {
		return 0
	}
	return float64(r.VGPUs) / float64(r.VGPUsTotal) * 100
}
