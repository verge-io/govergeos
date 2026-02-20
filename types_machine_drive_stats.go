package vergeos

// MachineDriveStats represents per-drive I/O statistics.
// There is one stats row per drive, linked via the parent_drive FK.
type MachineDriveStats struct {
	// Key is the unique identifier for this stats record.
	Key FlexInt `json:"$key,omitempty"`
	// ParentDrive is the parent machine drive ID.
	ParentDrive int `json:"parent_drive,omitempty"`
	// Reads is the total read operations.
	Reads uint64 `json:"reads,omitempty"`
	// Writes is the total write operations.
	Writes uint64 `json:"writes,omitempty"`
	// ReadBytes is the total bytes read.
	ReadBytes uint64 `json:"read_bytes,omitempty"`
	// WriteBytes is the total bytes written.
	WriteBytes uint64 `json:"write_bytes,omitempty"`
	// Rops is the read operations per second.
	Rops uint64 `json:"rops,omitempty"`
	// Wops is the write operations per second.
	Wops uint64 `json:"wops,omitempty"`
	// Rbps is the read bytes per second.
	Rbps uint64 `json:"rbps,omitempty"`
	// Wbps is the write bytes per second.
	Wbps uint64 `json:"wbps,omitempty"`
	// ServiceTime is the average I/O service time in milliseconds.
	ServiceTime float64 `json:"service_time,omitempty"`
	// Util is the I/O utilization percentage.
	Util float64 `json:"util,omitempty"`
	// Physical indicates whether these are physical drive stats.
	Physical bool `json:"physical,omitempty"`
}

// Field list constants for machine drive stats
const (
	machineDriveStatsListFields = "$key,parent_drive,reads,writes,read_bytes,write_bytes,rops,wops,rbps,wbps,service_time,util,physical"
)
