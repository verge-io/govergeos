package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequiredMajorStays26(t *testing.T) {
	if RequiredMajorVersion != 26 {
		t.Fatalf("RequiredMajorVersion = %d, want 26", RequiredMajorVersion)
	}
}

func TestVersion262Boundaries(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{version: "", want: false},
		{version: "26.0.0", want: false},
		{version: "26.0.2.2", want: false},
		{version: "26.1.8", want: false},
		{version: "26.1.999", want: false},
		{version: "26.2.0-rc1", want: false},
		{version: "26.2.0-beta1", want: false},
		{version: "26.2", want: true},
		{version: "26.2.0", want: true},
		{version: "v26.2.0", want: true},
		{version: "26.2.2.1", want: true},
		{version: "27.0.0", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			client := &Client{serverVersion: tt.version}
			if got := client.atLeast262(); got != tt.want {
				t.Fatalf("atLeast262(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestVersion262Projections(t *testing.T) {
	type proj struct {
		name   string
		fields func(*Client) string
		added  []string
		// kept is present on the older server as well (the 26.1.5 drive field).
		kept []string
	}
	projections := []proj{
		{
			name:   "vm list",
			fields: func(c *Client) string { return (&VMService{client: c}).listFields() },
			added:  []string{"bios_serial", "vgamem_mb"},
		},
		{
			name:   "vm get",
			fields: func(c *Client) string { return (&VMService{client: c}).getFields() },
			added:  []string{"bios_serial", "vgamem_mb"},
		},
		{
			name:   "drive list",
			fields: func(c *Client) string { return (&VMDriveService{client: c}).listFields() },
			added:  []string{"uuid", "is_primary", "share", "ovmf_code"},
			kept:   []string{driveMS2023KEKField},
		},
		{
			name:   "drive get",
			fields: func(c *Client) string { return (&VMDriveService{client: c}).getFields() },
			added:  []string{"uuid", "is_primary", "share", "ovmf_code"},
			kept:   []string{driveMS2023KEKField},
		},
		{
			name:   "file",
			fields: func(c *Client) string { return (&FileService{client: c}).listFields() },
			added:  []string{"uuid"},
		},
		{
			name:   "machine status list",
			fields: func(c *Client) string { return (&MachineStatusService{client: c}).listFields() },
			added:  []string{"agent_ip_list", "time_sync_status"},
		},
		{
			name:   "machine status get",
			fields: func(c *Client) string { return (&MachineStatusService{client: c}).getFields() },
			added:  []string{"agent_ip_list", "time_sync_status"},
		},
		{
			name:   "network list",
			fields: func(c *Client) string { return (&NetworkService{client: c}).listFields() },
			added:  strings.Split(network262Fields, ","),
		},
		{
			name:   "network get",
			fields: func(c *Client) string { return (&NetworkService{client: c}).getFields() },
			added:  strings.Split(network262Fields, ","),
		},
		{
			name:   "vm import",
			fields: func(c *Client) string { return (&VMImportService{client: c}).fields() },
			added:  strings.Split(vmImport262Fields, ","),
		},
		{
			name:   "update settings",
			fields: func(c *Client) string { return (&UpdateSettingsService{client: c}).getFields() },
			added:  []string{"restart_pending_packages"},
		},
		{
			name:   "update package",
			fields: func(c *Client) string { return (&UpdatePackageService{client: c}).listFields() },
			added:  []string{"dependencies"},
		},
		{
			name:   "bond interface",
			fields: func(c *Client) string { return (&VNetBondInterfaceService{client: c}).listFields() },
			added:  []string{"primary"},
		},
		{
			name:   "vnet bgp",
			fields: func(c *Client) string { return (&VNetBGPService{client: c}).listFields() },
			added:  []string{"advanced", "advanced_config", "console_rows", "console_cols", "daemon_bgpd", "daemon_ospfd", "daemon_eigrpd"},
		},
	}

	// 26.1.8 is the boundary below 26.2.0. A 26.2 prerelease stays on the old set.
	oldVersions := []string{"", "26.0.0", "26.1.8", "26.2.0-rc1"}
	newVersions := []string{"26.2.0", "26.2.2.1", "27.0.0"}

	for _, p := range projections {
		t.Run(p.name, func(t *testing.T) {
			for _, version := range oldVersions {
				fields := p.fields(&Client{serverVersion: version})
				for _, name := range p.added {
					if fieldSet(fields)[name] {
						t.Errorf("%s on %q includes %s: %s", p.name, version, name, fields)
					}
				}
			}
			for _, version := range newVersions {
				fields := p.fields(&Client{serverVersion: version})
				got := fieldSet(fields)
				for _, name := range p.added {
					if !got[name] {
						t.Errorf("%s on %q missing %s: %s", p.name, version, name, fields)
					}
				}
			}
		})
	}

	// The 26.1.5 KEK column stays on 26.1.8 and is still there on 26.2.
	drive := projections[2]
	oldDrive := drive.fields(&Client{serverVersion: "26.1.8"})
	if !fieldSet(oldDrive)[driveMS2023KEKField] {
		t.Fatalf("26.1.8 drive fields missing %s: %s", driveMS2023KEKField, oldDrive)
	}
	preKEK := drive.fields(&Client{serverVersion: "26.1.4"})
	if fieldSet(preKEK)[driveMS2023KEKField] || fieldSet(preKEK)["uuid"] {
		t.Fatalf("26.1.4 drive fields = %s", preKEK)
	}
	newDrive := drive.fields(&Client{serverVersion: "26.2.0"})
	if !fieldSet(newDrive)[driveMS2023KEKField] || !fieldSet(newDrive)["uuid"] {
		t.Fatalf("26.2.0 drive fields = %s", newDrive)
	}

	// Saved config and daemon flags stay on vnet_bgp, off the vnet projection.
	network := (&NetworkService{client: &Client{serverVersion: "26.2.2.1"}}).listFields()
	got := fieldSet(network)
	if got["bgp_config"] || got["frr_daemons"] || got["advanced_config"] || got["daemon_bgpd"] {
		t.Fatalf("default network projection includes vnet_bgp columns: %s", network)
	}
	if !got["console_status"] || !got["bgp_console_rows"] || !got["bgp_advanced"] {
		t.Fatalf("network projection = %s", network)
	}
}

func fieldSet(fields string) map[string]bool {
	set := map[string]bool{}
	for _, part := range strings.Split(fields, ",") {
		set[strings.TrimSpace(part)] = true
	}
	return set
}

func TestVersion262JSONNames(t *testing.T) {
	serial := "VM-SERIAL"
	mem := VGAMem32
	body, err := json.Marshal(VMCreateRequest{Name: "vm", CPUCores: 1, RAM: 512, BIOSSerial: serial, VGAMemMB: &mem})
	if err != nil {
		t.Fatal(err)
	}
	assertJSONKeys(t, body, map[string]any{"bios_serial": "VM-SERIAL", "vgamem_mb": "32"})

	var vm VM
	if err := json.Unmarshal([]byte(`{"bios_serial":"abc","vgamem_mb":"default"}`), &vm); err != nil {
		t.Fatal(err)
	}
	if vm.BIOSSerial != "abc" || vm.VGAMemMB != VGAMemDefault {
		t.Fatalf("vm = %+v", vm)
	}
	if err := json.Unmarshal([]byte(`{"vgamem_mb":16}`), &vm); err == nil {
		t.Fatal("vgamem_mb number was accepted")
	}

	share := true
	firmware := 9
	driveBody, err := json.Marshal(VMDriveCreateRequest{Name: "disk0", Share: &share, OVMFCode: &firmware})
	if err != nil {
		t.Fatal(err)
	}
	assertJSONKeys(t, driveBody, map[string]any{"share": true, "ovmf_code": float64(9)})

	updateBody, err := json.Marshal(VMDriveUpdateRequest{Name: strPtr("disk0")})
	if err != nil {
		t.Fatal(err)
	}
	var updateKeys map[string]any
	if err := json.Unmarshal(updateBody, &updateKeys); err != nil {
		t.Fatal(err)
	}
	if _, ok := updateKeys["share"]; ok {
		t.Fatalf("drive update includes share: %s", updateBody)
	}
	if _, ok := updateKeys["ovmf_code"]; ok {
		t.Fatalf("drive update includes ovmf_code: %s", updateBody)
	}

	var drive VMDrive
	if err := json.Unmarshal([]byte(`{"uuid":"abc","is_primary":true,"share":true,"ovmf_code":"fd-1"}`), &drive); err != nil {
		t.Fatal(err)
	}
	if drive.UUID != "abc" || !drive.IsPrimary || !drive.Share || drive.OVMFCode.String() != "fd-1" {
		t.Fatalf("drive = %+v", drive)
	}
	if err := json.Unmarshal([]byte(`{"ovmf_code":4}`), &drive); err != nil {
		t.Fatal(err)
	}
	if drive.OVMFCode.String() != "4" {
		t.Fatalf("ovmf_code = %q", drive.OVMFCode)
	}

	var file File
	if err := json.Unmarshal([]byte(`{"type":"fd","uuid":"file-uuid"}`), &file); err != nil {
		t.Fatal(err)
	}
	if file.Type != FileTypeFD || file.UUID != "file-uuid" {
		t.Fatalf("file = %+v", file)
	}

	var status MachineStatus
	if err := json.Unmarshal([]byte(`{"agent_ip_list":"10.0.0.5","time_sync_status":{"mode":"ntp","synced":true,"service_active":true,"status":"synced","detail":"ok","port_state":"","last_check":9}}`), &status); err != nil {
		t.Fatal(err)
	}
	if status.AgentIPList != "10.0.0.5" || !status.TimeSyncStatus.Present || status.TimeSyncStatus.Mode != TimeModeNTP || !status.TimeSyncStatus.Synced || status.TimeSyncStatus.Status != "synced" || status.TimeSyncStatus.LastCheck != 9 {
		t.Fatalf("status = %+v", status)
	}
	if err := json.Unmarshal([]byte(`{"agent_ip_list":"","time_sync_status":[]}`), &status); err != nil {
		t.Fatal(err)
	}
	if status.AgentIPList != "" || status.TimeSyncStatus.Present || status.TimeSyncStatus.Mode != "" {
		t.Fatalf("empty sync = %+v", status.TimeSyncStatus)
	}

	broadcast := true
	primary := 2
	mode := VLANBondModeBalanceSLB
	netBody, err := json.Marshal(NetworkCreateRequest{
		Name: "ext", DHCPBroadcast: &broadcast, VLANBondMode: mode, VLANBondPrimary: &primary,
		BGPAdvanced: boolPtr(true), BGPConsoleRows: intPtr(40), BGPConsoleCols: intPtr(120),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertJSONKeys(t, netBody, map[string]any{
		"dhcp_broadcast":    true,
		"vlan_bond_mode":    "balance-slb",
		"vlan_bond_primary": float64(2),
		"bgp_advanced":      true,
		"bgp_console_rows":  float64(40),
		"bgp_console_cols":  float64(120),
	})

	var network Network
	if err := json.Unmarshal([]byte(`{"dhcp_broadcast":true,"vlan_bond_primary":"nic0","console_status":{"$key":4},"vlan_bond_miimon":100}`), &network); err != nil {
		t.Fatal(err)
	}
	if !network.DHCPBroadcast || network.VLANBondPrimary.String() != "nic0" || network.ConsoleStatus == nil || network.ConsoleStatus.Int() != 4 || network.VLANBondMIIMON != 100 {
		t.Fatalf("network = %+v", network)
	}
	if err := json.Unmarshal([]byte(`{"console_status":null}`), &network); err != nil {
		t.Fatal(err)
	}
	if network.ConsoleStatus != nil {
		t.Fatalf("null console_status = %#v", network.ConsoleStatus)
	}

	var settings UpdateSettings
	if err := json.Unmarshal([]byte(`{"restart_pending_packages":["vergeos","vergeos-ui"]}`), &settings); err != nil {
		t.Fatal(err)
	}
	if len(settings.RestartPendingPackages) != 2 || settings.RestartPendingPackages[0] != "vergeos" {
		t.Fatalf("packages = %#v", settings.RestartPendingPackages)
	}

	var pkg UpdatePackage
	if err := json.Unmarshal([]byte(`{"$key":"ai","dependencies":"cuda, driver"}`), &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Key.String() != "ai" || len(pkg.Dependencies) != 2 || pkg.Dependencies[1] != "driver" {
		t.Fatalf("package = %+v", pkg)
	}

	access := false
	userBody, err := json.Marshal(UserCreateRequest{Name: "ada", PhysicalAccess: &access})
	if err != nil {
		t.Fatal(err)
	}
	assertJSONKeys(t, userBody, map[string]any{"physical_access": false})
	omitted, err := json.Marshal(UserCreateRequest{Name: "ada"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(omitted), "physical_access") {
		t.Fatalf("omitted physical_access was sent: %s", omitted)
	}
	var user User
	if err := json.Unmarshal([]byte(`{"name":"ada"}`), &user); err != nil {
		t.Fatal(err)
	}
	if user.PhysicalAccess {
		t.Fatal("physical_access defaulted to true")
	}
}

func TestVMClonePreserveBIOSSerialByVersion(t *testing.T) {
	var got map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method %s", r.Method)
			}
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
		},
	}))

	client.serverVersion = "26.1.8"
	if err := client.VMs.Clone(context.Background(), 4, &VMCloneOptions{PreserveBIOSSerial: true}); err != nil {
		t.Fatal(err)
	}
	params := got["params"].(map[string]any)
	if _, ok := params["preserve_bios_serial"]; ok {
		t.Fatalf("26.1 clone sent preserve_bios_serial: %#v", params)
	}
	if got["action"] != "clone" || got["vm"].(float64) != 4 {
		t.Fatalf("clone body %#v", got)
	}

	client.serverVersion = "26.2.0"
	if err := client.VMs.Clone(context.Background(), 4, &VMCloneOptions{Name: "copy", PreserveBIOSSerial: true}); err != nil {
		t.Fatal(err)
	}
	params = got["params"].(map[string]any)
	if params["preserve_bios_serial"] != true || params["name"] != "copy" {
		t.Fatalf("26.2 clone params %#v", params)
	}

	client.serverVersion = "26.2.0"
	handlerErr := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "clone failed"})
		},
	}))
	handlerErr.serverVersion = "26.2.0"
	if err := handlerErr.VMs.Clone(context.Background(), 4, &VMCloneOptions{PreserveBIOSSerial: true}); err == nil {
		t.Fatal("expected clone error")
	}
}

func TestVMImport262Options(t *testing.T) {
	preserve := true
	agent := true
	tags := false
	var posted map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{})
		},
		"POST /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatal(err)
			}
			jsonResponse(w, 200, map[string]any{"$key": "abc"})
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []map[string]any{{
				"$key": "abc", "id": "abc", "name": "imported",
				"preserve_bios_serial": true, "guest_agent_preinstalled": true, "replicate_tags": false,
			}})
		},
	}))
	client.serverVersion = "26.2.0"

	got, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{
		Name: "imported", File: intPtr(3),
		PreserveBIOSSerial: &preserve, GuestAgentPreinstalled: &agent, ReplicateTags: &tags,
	})
	if err != nil {
		t.Fatal(err)
	}
	if posted["preserve_bios_serial"] != true || posted["guest_agent_preinstalled"] != true || posted["replicate_tags"] != false {
		t.Fatalf("posted %#v", posted)
	}
	if !got.PreserveBIOSSerial || !got.GuestAgentPreinstalled || got.ReplicateTags {
		t.Fatalf("import = %+v", got)
	}
}

func TestUpdateUninstallAction(t *testing.T) {
	var posted map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("fields"), "restart_pending_packages") {
				t.Errorf("fields %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, UpdateSettings{Key: 1, Source: 5})
		},
		"POST /api/v4/update_actions": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v4/update_actions" {
				t.Errorf("path %s", r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusCreated)
		},
	}))
	client.serverVersion = "26.2.0"
	if err := client.UpdateSettings.Uninstall(context.Background(), "private-ai"); err != nil {
		t.Fatal(err)
	}
	if posted["action"] != UpdateActionUninstall || posted["installed_package"] != "private-ai" || posted["source"].(float64) != 5 {
		t.Fatalf("body %#v", posted)
	}
	if _, ok := posted["force"]; ok {
		t.Fatalf("uninstall sent force: %#v", posted)
	}

	if err := client.UpdateSettings.Uninstall(context.Background(), ""); err == nil || !IsValidationError(err) {
		t.Fatalf("empty package: %v", err)
	}

	failing := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateSettings{Key: 1, Source: 5})
		},
		"POST /api/v4/update_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 422, map[string]string{"err": "package is required by another package"})
		},
	}))
	failing.serverVersion = "26.2.0"
	err := failing.UpdateSettings.Uninstall(context.Background(), "private-ai")
	if err == nil || !strings.Contains(err.Error(), "uninstall package") {
		t.Fatalf("error = %v", err)
	}
}

func TestUpdatePackageAndSourceStatus(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_packages": func(w http.ResponseWriter, r *http.Request) {
			if fieldSet(r.URL.Query().Get("fields"))["dependencies"] {
				t.Errorf("26.1 fields include dependencies: %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []UpdatePackage{{Key: "vergeos", Name: "vergeos"}})
		},
		"GET /api/v4/update_source_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []UpdateSourceStatus{{Key: 1, Source: 5, Status: UpdateSourceStatusUninstalling}})
		},
	}))
	client.serverVersion = "26.1.8"
	packages, err := client.UpdatePackages.List(context.Background())
	if err != nil || packages[0].Name != "vergeos" {
		t.Fatalf("packages %#v err %v", packages, err)
	}
	status, err := client.UpdateSourceStatus.GetBySource(context.Background(), 5)
	if err != nil || status.Status != "uninstalling" {
		t.Fatalf("status %#v err %v", status, err)
	}

	newer := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_packages/private-ai": func(w http.ResponseWriter, r *http.Request) {
			if !fieldSet(r.URL.Query().Get("fields"))["dependencies"] {
				t.Errorf("fields %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, UpdatePackage{Key: "private-ai", Dependencies: StringList{"cuda"}})
		},
	}))
	newer.serverVersion = "26.2.0"
	pkg, err := newer.UpdatePackages.Get(context.Background(), "private-ai")
	if err != nil || len(pkg.Dependencies) != 1 || pkg.Dependencies[0] != "cuda" {
		t.Fatalf("pkg %#v err %v", pkg, err)
	}
}

func TestTimeSettingsAndDashboard(t *testing.T) {
	old := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/time_settings/self": func(w http.ResponseWriter, r *http.Request) {
			t.Error("time settings requested on 26.1")
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/time_settings/1": func(w http.ResponseWriter, r *http.Request) {
			t.Error("time settings used the numeric key")
			w.WriteHeader(http.StatusOK)
		},
	}))
	old.serverVersion = "26.1.8"
	if _, err := old.TimeSettings.Get(context.Background()); err == nil || !strings.Contains(err.Error(), "26.2") {
		t.Fatalf("old server error = %v", err)
	}

	var put map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/time_settings/self": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, "ntp_maxclock") || !strings.Contains(fields, "interface_vnet") || !strings.Contains(fields, "ptp_use_hw_timestamps") || !strings.Contains(fields, "ptp_neighbor_prop_delay_thresh") {
				t.Errorf("fields %s", fields)
			}
			if strings.Contains(fields, "ntp_max_clocks") || strings.Contains(fields, "ptp_interface") {
				t.Errorf("fields still use the old names: %s", fields)
			}
			jsonResponse(w, 200, TimeSettings{Key: "self", Mode: TimeModeNTP, NTPServers: "ntp1.example.com", NTPMaxClocks: 11})
		},
		"PUT /api/v4/time_settings/self": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&put); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/time_dashboard": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{
				"settings": map[string]any{"$key": "self", "mode": TimeModePTP, "ntp_servers": "ntp1.example.com"},
				"nodes": []map[string]any{{
					"$key": 1, "name": "node1",
					"machine": map[string]any{
						"status": map[string]any{
							"time_sync_status": map[string]any{"mode": TimeModePTP, "synced": true, "status": "synced"},
						},
					},
				}},
				"alarms": []map[string]any{{"$key": 8, "sub_owner_type": AlarmSubOwnerTypeWire}},
			})
		},
	}))
	client.serverVersion = "26.2.2.1"

	settings, err := client.TimeSettings.Get(context.Background())
	if err != nil || settings.Key.String() != "self" || settings.NTPServers != "ntp1.example.com" || settings.Mode != TimeModeNTP || settings.NTPMaxClocks != 11 {
		t.Fatalf("settings %#v err %v", settings, err)
	}
	mode := TimeModePTP
	domain := 24
	updated, err := client.TimeSettings.Update(context.Background(), &TimeSettingsUpdateRequest{
		Mode: &mode, PTPDomain: &domain, PTPTransport: strPtr(PTPTransportL2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if put["mode"] != TimeModePTP || put["ptp_domain"].(float64) != 24 || put["ptp_transport"] != PTPTransportL2 {
		t.Fatalf("put %#v", put)
	}
	if updated.Mode != TimeModeNTP {
		t.Fatalf("updated mode %s", updated.Mode)
	}

	dash, err := client.TimeDashboard.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dash.Settings == nil || dash.Settings.Mode != TimeModePTP || len(dash.Nodes) != 1 || dash.Nodes[0].Name != "node1" || dash.Nodes[0].Machine == nil || !dash.Nodes[0].Machine.Status.TimeSyncStatus.Present || len(dash.Alarms) != 1 {
		t.Fatalf("dashboard %+v", dash)
	}

	arrayClient := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/time_dashboard": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TimeNodeSync{{Key: 2, Name: "node2"}})
		},
	}))
	arrayClient.serverVersion = "26.2.0"
	arrayDash, err := arrayClient.TimeDashboard.Get(context.Background())
	if err != nil || len(arrayDash.Nodes) != 1 || arrayDash.Nodes[0].Name != "node2" {
		t.Fatalf("array dashboard %#v err %v", arrayDash, err)
	}

	failing := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/time_settings/self": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 400, map[string]string{"err": "ptp interface is required"})
		},
	}))
	failing.serverVersion = "26.2.0"
	servers := "ntp1.example.com"
	if _, err := failing.TimeSettings.Update(context.Background(), &TimeSettingsUpdateRequest{NTPServers: &servers}); err == nil {
		t.Fatal("expected update error")
	}
}

func TestAdvancedRouterAndBondPrimary(t *testing.T) {
	old := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnets/3": func(w http.ResponseWriter, r *http.Request) {
			t.Error("advanced router requested on 26.1")
			w.WriteHeader(http.StatusOK)
		},
	}))
	old.serverVersion = "26.1.8"
	if _, err := old.Networks.GetAdvancedRouter(context.Background(), 3); err == nil {
		t.Fatal("expected version error")
	}

	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
			fields := fieldSet(r.URL.Query().Get("fields"))
			if !fields["advanced_config"] || !fields["daemon_bgpd"] || !fields["daemon_ospfd"] {
				t.Errorf("fields %s", r.URL.Query().Get("fields"))
			}
			if r.URL.Query().Get("filter") != "vnet eq 3" {
				t.Errorf("filter %s", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []VNetBGP{{Key: 7, VNet: 3}})
		},
		"GET /api/v4/vnet_bgp/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBGP{
				Key: 7, VNet: 3, Advanced: true, AdvancedConfig: "router bgp 65000",
				ConsoleRows: 40, ConsoleCols: 120,
				FRRDaemons: FRRDaemons{BGP: true, OSPF: true},
			})
		},
		"GET /api/v4/vnet_bond_interfaces": func(w http.ResponseWriter, r *http.Request) {
			if !fieldSet(r.URL.Query().Get("fields"))["primary"] {
				t.Errorf("fields %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []VNetBondInterface{{Key: 8, VNet: 3, Interface: 2, Primary: true}})
		},
		"PUT /api/v4/vnet_bond_interfaces/8": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["primary"] != true {
				t.Fatalf("body %#v", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vnet_bond_interfaces/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBondInterface{Key: 8, Primary: true})
		},
	}))
	client.serverVersion = "26.2.0"

	router, err := client.Networks.GetAdvancedRouter(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if !router.Enabled || router.Config != "router bgp 65000" || !router.Daemons.BGP || !router.Daemons.OSPF || router.ConsoleRows != 40 || int(router.VNet) != 3 {
		t.Fatalf("router %+v", router)
	}

	rows, err := client.VNetBondInterfaces.ListByNetwork(context.Background(), 3)
	if err != nil || len(rows) != 1 || !rows[0].Primary {
		t.Fatalf("bonds %#v err %v", rows, err)
	}
	updated, err := client.VNetBondInterfaces.Update(context.Background(), 8, &VNetBondInterfaceUpdateRequest{Primary: boolPtr(true)})
	if err != nil || !updated.Primary {
		t.Fatalf("updated %#v err %v", updated, err)
	}

	missing := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetBGP{})
		},
	}))
	missing.serverVersion = "26.2.0"
	if _, err := missing.Networks.GetAdvancedRouter(context.Background(), 9); !IsNotFoundError(err) {
		t.Fatalf("missing router error %v", err)
	}
}

func TestTimeSettingsNullPTPInterface(t *testing.T) {
	var settings TimeSettings
	if err := json.Unmarshal([]byte(`{"$key":"self","mode":"ntp","interface_vnet":null,"ntp_maxclock":11,"ntp_servers":"0.pool.ntp.org"}`), &settings); err != nil {
		t.Fatal(err)
	}
	if settings.Key.String() != "self" || settings.Mode != TimeModeNTP || settings.PTPInterface != nil || settings.NTPMaxClocks != 11 || settings.NTPServers != "0.pool.ntp.org" {
		t.Fatalf("settings = %+v", settings)
	}
}

func TestFRRDaemonsObject(t *testing.T) {
	var daemons FRRDaemons
	if err := json.Unmarshal([]byte(`{"daemon_bgpd":true,"daemon_eigrpd":true}`), &daemons); err != nil {
		t.Fatal(err)
	}
	if !daemons.BGP || !daemons.EIGRP || daemons.OSPF {
		t.Fatalf("daemons %+v", daemons)
	}
}

func TestSiteSyncIdleAndQueryConstants(t *testing.T) {
	if SiteSyncIncomingStatusIdle != "idle" || SiteSyncOutgoingStatusIdle != "idle" {
		t.Fatalf("idle = %q %q", SiteSyncIncomingStatusIdle, SiteSyncOutgoingStatusIdle)
	}
	if SiteSyncIncomingStatusOffline != "offline" || SiteSyncOutgoingStatusOffline != "offline" {
		t.Fatal("offline status was removed")
	}
	if NetworkQueryWakeOnLAN != "wakeonlan" {
		t.Fatalf("wakeonlan = %q", NetworkQueryWakeOnLAN)
	}
	if AlarmSubOwnerTypeWire != "vnet_wires" {
		t.Fatalf("wire = %q", AlarmSubOwnerTypeWire)
	}
	if UpdateSourceStatusUninstalling != "uninstalling" {
		t.Fatalf("uninstalling = %q", UpdateSourceStatusUninstalling)
	}
}

func TestNoFindInodeQuery(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(body), `"findinode"`) {
			t.Errorf("%s contains a findinode query value", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertJSONKeys(t *testing.T, body []byte, want map[string]any) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %#v, want %#v (body %s)", key, got[key], value, body)
		}
	}
}
