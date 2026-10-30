package vergeos

// Time synchronization modes.
const (
	// TimeModeNTP is the default. Node 1 follows the NTP servers.
	// The other nodes follow node 1.
	TimeModeNTP = "ntp"
	// TimeModePTP uses Precision Time Protocol instead of NTP.
	TimeModePTP = "ptp"
)

// PTP transport values.
const (
	PTPTransportUDPv4 = "udpv4"
	PTPTransportUDPv6 = "udpv6"
	PTPTransportL2    = "l2"
)

// PTP delay mechanism values.
const (
	PTPDelayE2E  = "e2e"
	PTPDelayP2P  = "p2p"
	PTPDelayAuto = "auto"
)

// TimeSettings is the NTP/PTP configuration singleton (time_settings, key 1).
// VergeOS 26.2. The system time zone stays in System settings.
type TimeSettings struct {
	// Key is the row key. The settings row is 1.
	Key FlexInt `json:"$key,omitempty"`
	// Mode is ntp or ptp.
	Mode string `json:"mode,omitempty"`
	// NTPServers is the space-separated server list node 1 follows.
	NTPServers string `json:"ntp_servers,omitempty"`
	// NTPMaxClocks is the maximum number of NTP sources.
	NTPMaxClocks int `json:"ntp_max_clocks,omitempty"`
	// NTPMinClocks is the minimum number of NTP sources NTP tries to use.
	NTPMinClocks int `json:"ntp_min_clocks,omitempty"`
	// NTPMinSane is the minimum number of sources that must agree.
	NTPMinSane int `json:"ntp_min_sane,omitempty"`
	// PTPInterface is the physical network that carries PTP traffic.
	// Null means no network is selected.
	PTPInterface *FlexInt `json:"ptp_interface,omitempty"`
	// PTPTransport is udpv4, udpv6, or l2.
	PTPTransport string `json:"ptp_transport,omitempty"`
	// PTPDomain is the PTP domain, 0 through 127.
	PTPDomain int `json:"ptp_domain,omitempty"`
	// PTPDelayMechanism is e2e, p2p, or auto.
	PTPDelayMechanism string `json:"ptp_delay_mechanism,omitempty"`
	// PTPHardwareTimestamp uses the NIC hardware clock when true.
	PTPHardwareTimestamp bool `json:"ptp_hardware_timestamp,omitempty"`
	// PTPPriority1 is BMCA priority 1. A lower value is a higher priority.
	PTPPriority1 int `json:"ptp_priority1,omitempty"`
	// PTPPriority2 is BMCA priority 2.
	PTPPriority2 int `json:"ptp_priority2,omitempty"`
	// PTPLogSyncInterval is the sync rate as log2 seconds.
	PTPLogSyncInterval int `json:"ptp_log_sync_interval,omitempty"`
	// PTPLogAnnounceInterval is the announce rate as log2 seconds.
	PTPLogAnnounceInterval int `json:"ptp_log_announce_interval,omitempty"`
	// PTPLogMinDelayReqInterval is the minimum delay-request rate as log2 seconds.
	PTPLogMinDelayReqInterval int `json:"ptp_log_min_delay_req_interval,omitempty"`
	// PTPLogMinPDelayReqInterval is the minimum peer-delay rate as log2 seconds.
	PTPLogMinPDelayReqInterval int `json:"ptp_log_min_pdelay_req_interval,omitempty"`
	// PTPUnicast uses unicast PTP messages when true.
	PTPUnicast bool `json:"ptp_unicast,omitempty"`
	// PTPDSCPEvent is the DSCP value for PTP event messages (port 319).
	PTPDSCPEvent int `json:"ptp_dscp_event,omitempty"`
	// PTPDSCPGeneral is the DSCP value for PTP general messages (port 320).
	PTPDSCPGeneral int `json:"ptp_dscp_general,omitempty"`
	// PTPSyncReceiptTimeout is how many sync intervals pass before a timeout. 0 disables it.
	PTPSyncReceiptTimeout int `json:"ptp_sync_receipt_timeout,omitempty"`
	// PTPAnnounceReceiptTimeout is how many announce intervals pass before a timeout.
	PTPAnnounceReceiptTimeout int `json:"ptp_announce_receipt_timeout,omitempty"`
	// PTPTwoStep sends the timestamp in a Follow_Up message when true.
	PTPTwoStep bool `json:"ptp_two_step,omitempty"`
	// PTPMasterOnly makes the port act as master and skip BMCA.
	PTPMasterOnly bool `json:"ptp_master_only,omitempty"`
	// PTPInhibitAnnounce stops announce messages.
	PTPInhibitAnnounce bool `json:"ptp_inhibit_announce,omitempty"`
	// PTPInhibitDelayReq stops delay-request messages.
	PTPInhibitDelayReq bool `json:"ptp_inhibit_delay_req,omitempty"`
	// PTPFollowUpInfo adds the follow-up information TLV.
	PTPFollowUpInfo bool `json:"ptp_follow_up_info,omitempty"`
	// PTPTransportSpecific is the PTP transportSpecific field.
	PTPTransportSpecific int `json:"ptp_transport_specific,omitempty"`
	// PTPNeighborPropDelay is the maximum neighbor propagation delay in nanoseconds.
	PTPNeighborPropDelay int `json:"ptp_neighbor_prop_delay,omitempty"`
}

// TimeSettingsUpdateRequest updates the time settings singleton.
// Nil fields are left unchanged.
type TimeSettingsUpdateRequest struct {
	Mode                       *string `json:"mode,omitempty"`
	NTPServers                 *string `json:"ntp_servers,omitempty"`
	NTPMaxClocks               *int    `json:"ntp_max_clocks,omitempty"`
	NTPMinClocks               *int    `json:"ntp_min_clocks,omitempty"`
	NTPMinSane                 *int    `json:"ntp_min_sane,omitempty"`
	PTPInterface               *int    `json:"ptp_interface,omitempty"`
	PTPTransport               *string `json:"ptp_transport,omitempty"`
	PTPDomain                  *int    `json:"ptp_domain,omitempty"`
	PTPDelayMechanism          *string `json:"ptp_delay_mechanism,omitempty"`
	PTPHardwareTimestamp       *bool   `json:"ptp_hardware_timestamp,omitempty"`
	PTPPriority1               *int    `json:"ptp_priority1,omitempty"`
	PTPPriority2               *int    `json:"ptp_priority2,omitempty"`
	PTPLogSyncInterval         *int    `json:"ptp_log_sync_interval,omitempty"`
	PTPLogAnnounceInterval     *int    `json:"ptp_log_announce_interval,omitempty"`
	PTPLogMinDelayReqInterval  *int    `json:"ptp_log_min_delay_req_interval,omitempty"`
	PTPLogMinPDelayReqInterval *int    `json:"ptp_log_min_pdelay_req_interval,omitempty"`
	PTPUnicast                 *bool   `json:"ptp_unicast,omitempty"`
	PTPDSCPEvent               *int    `json:"ptp_dscp_event,omitempty"`
	PTPDSCPGeneral             *int    `json:"ptp_dscp_general,omitempty"`
	PTPSyncReceiptTimeout      *int    `json:"ptp_sync_receipt_timeout,omitempty"`
	PTPAnnounceReceiptTimeout  *int    `json:"ptp_announce_receipt_timeout,omitempty"`
	PTPTwoStep                 *bool   `json:"ptp_two_step,omitempty"`
	PTPMasterOnly              *bool   `json:"ptp_master_only,omitempty"`
	PTPInhibitAnnounce         *bool   `json:"ptp_inhibit_announce,omitempty"`
	PTPInhibitDelayReq         *bool   `json:"ptp_inhibit_delay_req,omitempty"`
	PTPFollowUpInfo            *bool   `json:"ptp_follow_up_info,omitempty"`
	PTPTransportSpecific       *int    `json:"ptp_transport_specific,omitempty"`
	PTPNeighborPropDelay       *int    `json:"ptp_neighbor_prop_delay,omitempty"`
}

func (r *TimeSettingsUpdateRequest) empty() bool {
	if r == nil {
		return true
	}
	return r.Mode == nil && r.NTPServers == nil && r.NTPMaxClocks == nil &&
		r.NTPMinClocks == nil && r.NTPMinSane == nil && r.PTPInterface == nil &&
		r.PTPTransport == nil && r.PTPDomain == nil && r.PTPDelayMechanism == nil &&
		r.PTPHardwareTimestamp == nil && r.PTPPriority1 == nil && r.PTPPriority2 == nil &&
		r.PTPLogSyncInterval == nil && r.PTPLogAnnounceInterval == nil &&
		r.PTPLogMinDelayReqInterval == nil && r.PTPLogMinPDelayReqInterval == nil &&
		r.PTPUnicast == nil && r.PTPDSCPEvent == nil && r.PTPDSCPGeneral == nil &&
		r.PTPSyncReceiptTimeout == nil && r.PTPAnnounceReceiptTimeout == nil &&
		r.PTPTwoStep == nil && r.PTPMasterOnly == nil && r.PTPInhibitAnnounce == nil &&
		r.PTPInhibitDelayReq == nil && r.PTPFollowUpInfo == nil &&
		r.PTPTransportSpecific == nil && r.PTPNeighborPropDelay == nil
}

// timeSettingsFields is the projection for the singleton.
// The two-step, master, inhibit, follow-up, transport-specific, and
// neighbor-delay columns stay off this list. Update can still write them.
const timeSettingsFields = "$key,mode,ntp_servers,ntp_max_clocks,ntp_min_clocks,ntp_min_sane," +
	"ptp_interface,ptp_transport,ptp_domain,ptp_delay_mechanism,ptp_hardware_timestamp," +
	"ptp_priority1,ptp_priority2,ptp_log_sync_interval,ptp_log_announce_interval," +
	"ptp_log_min_delay_req_interval,ptp_log_min_pdelay_req_interval,ptp_unicast," +
	"ptp_dscp_event,ptp_dscp_general,ptp_sync_receipt_timeout,ptp_announce_receipt_timeout"

// TimeNodeSync is one physical node's time sync state on the time dashboard.
type TimeNodeSync struct {
	// Node is the node key.
	Node FlexInt `json:"node,omitempty"`
	// Name is the node name.
	Name string `json:"name,omitempty"`
	// Status is the sync state the dashboard reports.
	Status string `json:"status,omitempty"`
	// Source is the time source, a public server for node 1 and node 1 for the rest.
	Source string `json:"source,omitempty"`
	// LastCheck is when the status was last checked, as VergeOS reported it.
	LastCheck int64 `json:"last_check,omitempty"`
}

// TimeDashboard is the read-only time view. It combines the settings,
// per-node sync status, and time alarms. VergeOS 26.2.
type TimeDashboard struct {
	TimeSettings
	// Settings is set when the view nests the configuration.
	Settings *TimeSettings `json:"time_settings,omitempty"`
	// Nodes is the sync status of each physical node.
	Nodes []TimeNodeSync `json:"nodes,omitempty"`
	// Alarms are the time-related alarms included in the view.
	Alarms []Alarm `json:"alarms,omitempty"`
}

const timeDashboardFields = timeSettingsFields + ",nodes,alarms"
