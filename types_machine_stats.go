package vergeos

import "encoding/json"

// MachineStats represents per-machine statistics including CPU, RAM, and temperature metrics.
// There is one stats row per machine, linked via the machine FK.
type MachineStats struct {
	// Key is the unique identifier for this stats record.
	Key FlexInt `json:"$key,omitempty"`
	// Machine is the parent machine ID.
	Machine int `json:"machine,omitempty"`
	// TotalCPU is the total CPU usage percentage (0-100).
	TotalCPU uint8 `json:"total_cpu,omitempty"`
	// UserCPU is the user-space CPU usage percentage.
	UserCPU uint8 `json:"user_cpu,omitempty"`
	// SystemCPU is the system/kernel CPU usage percentage.
	SystemCPU uint8 `json:"system_cpu,omitempty"`
	// IOWaitCPU is the I/O wait CPU percentage.
	IOWaitCPU uint8 `json:"iowait_cpu,omitempty"`
	// RAMUsed is the RAM used in MB.
	RAMUsed uint32 `json:"ram_used,omitempty"`
	// RAMPct is the physical RAM used percentage (0-100).
	RAMPct uint8 `json:"ram_pct,omitempty"`
	// CoreUsageList is the per-core CPU usage as a JSON array of floats.
	CoreUsageList json.RawMessage `json:"core_usagelist,omitempty"`
	// CoreTemp is the average core temperature in Celsius.
	CoreTemp uint16 `json:"core_temp,omitempty"`
	// CoreTempTop is the maximum core temperature in Celsius.
	CoreTempTop uint16 `json:"core_temp_top,omitempty"`
}

// GetCoreUsages parses the CoreUsageList JSON into a slice of float64 values.
// Returns nil if CoreUsageList is empty or not set.
func (ms *MachineStats) GetCoreUsages() ([]float64, error) {
	if len(ms.CoreUsageList) == 0 {
		return nil, nil
	}
	var usages []float64
	if err := json.Unmarshal(ms.CoreUsageList, &usages); err != nil {
		return nil, err
	}
	return usages, nil
}

// Field list constants for machine stats
const (
	machineStatsListFields = "$key,machine,total_cpu,user_cpu,system_cpu,iowait_cpu,ram_used,ram_pct,core_usagelist,core_temp,core_temp_top"
)
