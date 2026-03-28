package vergeos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient creates a Client pointing at an httptest.Server.
// The handler receives requests under /api/v4/* for service calls.
// No version check is performed.
func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c := &Client{
		baseURL:    server.URL,
		username:   "test",
		password:   "test",
		httpClient: server.Client(),
		userAgent:  "govergeos-test",
	}
	initServices(c)
	return c
}

// initServices initialises every service on the client.
// Mirrors the block in NewClient() so unit tests have access to all services.
func initServices(c *Client) {
	c.VMs = &VMService{client: c}
	c.VMSnapshots = &VMSnapshotService{client: c}
	c.VMNICs = &VMNICService{client: c}
	c.VMDrives = &VMDriveService{client: c}
	c.VMDevices = &VMDeviceService{client: c}
	c.Networks = &NetworkService{client: c}
	c.Users = &UserService{client: c}
	c.Members = &MemberService{client: c}
	c.CloudInitFiles = &CloudInitService{client: c}
	c.Clusters = &ClusterService{client: c}
	c.Nodes = &NodeService{client: c}
	c.Groups = &GroupService{client: c}
	c.Files = &FileService{client: c}
	c.ResourceGroups = &ResourceGroupService{client: c}
	c.Settings = &SettingsService{client: c}
	c.System = &SystemService{client: c}
	c.Schema = &SchemaService{client: c}
	c.Tags = &TagService{client: c}
	c.TagCategories = &TagCategoryService{client: c}
	c.TagMembers = &TagMemberService{client: c}
	c.Volumes = &VolumeService{client: c}
	c.VNetRules = &VNetRuleService{client: c}
	c.VNetRuleAliases = &VNetRuleAliasService{client: c}
	c.Tenants = &TenantService{client: c}
	c.TenantNodes = &TenantNodeService{client: c}
	c.TenantStorage = &TenantStorageService{client: c}
	c.TenantStatus = &TenantStatusService{client: c}
	c.TenantStatsHistoryShort = &TenantStatsHistoryShortService{client: c}
	c.TenantSnapshots = &TenantSnapshotService{client: c}
	c.TenantLayer2Networks = &TenantLayer2NetworkService{client: c}
	c.SnapshotProfiles = &SnapshotProfileService{client: c}
	c.SnapshotProfilePeriods = &SnapshotProfilePeriodService{client: c}
	c.Alarms = &AlarmService{client: c}
	c.AlarmTypes = &AlarmTypeService{client: c}
	c.Tasks = &TaskService{client: c}
	c.VNetAddresses = &VNetAddressService{client: c}
	c.VNetDNSViews = &VNetDNSViewService{client: c}
	c.VNetDNSZones = &VNetDNSZoneService{client: c}
	c.VNetDNSRecords = &VNetDNSRecordService{client: c}
	c.VNetHosts = &VNetHostService{client: c}
	c.VNetWireGuards = &VNetWireGuardService{client: c}
	c.VNetWireGuardPeers = &VNetWireGuardPeerService{client: c}
	c.VNetWireGuardPeerStatus = &VNetWireGuardPeerStatusService{client: c}
	c.Certificates = &CertificateService{client: c}
	c.VNetIPSecs = &VNetIPSecService{client: c}
	c.VNetIPSecPhase1s = &VNetIPSecPhase1Service{client: c}
	c.VNetIPSecPhase2s = &VNetIPSecPhase2Service{client: c}
	c.VNetIPSecConnections = &VNetIPSecConnectionService{client: c}
	c.Sites = &SiteService{client: c}
	c.SiteSyncsIncoming = &SiteSyncIncomingService{client: c}
	c.SiteSyncsOutgoing = &SiteSyncOutgoingService{client: c}
	c.SiteSyncProfilePeriods = &SiteSyncProfilePeriodService{client: c}
	c.CloudSnapshots = &CloudSnapshotService{client: c}
	c.CloudSnapshotVMs = &CloudSnapshotVMService{client: c}
	c.CloudSnapshotTenants = &CloudSnapshotTenantService{client: c}
	c.VolumeCIFSShares = &VolumeCIFSShareService{client: c}
	c.VolumeNFSShares = &VolumeNFSShareService{client: c}
	c.VolumeBrowser = &VolumeBrowserService{client: c}
	c.WebhookURLs = &WebhookURLService{client: c}
	c.Webhooks = &WebhookService{client: c}
	c.UserAPIKeys = &UserAPIKeyService{client: c}
	c.NASServices = &NASServiceService{client: c}
	c.NASServiceUsers = &NASServiceUserService{client: c}
	c.VolumeSyncs = &VolumeSyncService{client: c}
	c.VolumeSnapshots = &VolumeSnapshotService{client: c}
	c.Permissions = &PermissionService{client: c}
	c.Logs = &LogService{client: c}
	c.StorageTiers = &StorageTierService{client: c}
	c.ClusterTiers = &ClusterTierService{client: c}
	c.MachineDrivePhys = &MachineDrivePhysService{client: c}
	c.ClusterStatsHistory = &ClusterStatsHistoryService{client: c}
	c.MachineStatus = &MachineStatusService{client: c}
	c.MachineStats = &MachineStatsService{client: c}
	c.MachineDriveStats = &MachineDriveStatsService{client: c}
	c.MachineNICs = &MachineNICService{client: c}
	c.UpdateSettings = &UpdateSettingsService{client: c}
	c.UpdateBranches = &UpdateBranchService{client: c}
	c.UpdateSourcePackages = &UpdateSourcePackageService{client: c}
}

// jsonResponse writes a JSON response with the given status code.
func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// apiMux builds an http.ServeMux with route handlers under /api/v4.
// Routes are specified as "METHOD /path" keys.
func apiMux(routes map[string]http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		// Try exact match first
		if h, ok := routes[key]; ok {
			h(w, r)
			return
		}
		// Try wildcard: match by method + prefix
		for pattern, h := range routes {
			parts := strings.SplitN(pattern, " ", 2)
			if len(parts) == 2 && parts[0] == r.Method && strings.HasPrefix(r.URL.Path, parts[1]) {
				h(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
	return mux
}
