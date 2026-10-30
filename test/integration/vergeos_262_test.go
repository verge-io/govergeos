//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestVergeOS262Fields reads the VergeOS 26.2 additions against the lab.
// Writes that would change the system (uninstall, time settings, bond primary)
// stay out of this test.
func TestVergeOS262Fields(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("TimeSettings", func(t *testing.T) {
		settings, err := client.TimeSettings.Get(ctx)
		if err != nil {
			t.Fatalf("TimeSettings.Get failed: %v", err)
		}
		if settings.Key.String() != "self" {
			t.Fatalf("time settings key = %q, want self", settings.Key)
		}
		if settings.Mode == "" {
			t.Fatal("time settings mode is empty")
		}
		if settings.Mode == vergeos.TimeModeNTP && settings.NTPServers == "" {
			t.Fatal("ntp mode has no ntp_servers")
		}
		if settings.NTPMaxClocks == 0 || settings.NTPMinClocks == 0 || settings.NTPMinSane == 0 {
			t.Fatalf("ntp clock limits read back empty: max=%d min=%d sane=%d",
				settings.NTPMaxClocks, settings.NTPMinClocks, settings.NTPMinSane)
		}
		t.Logf("mode=%s ntp_servers=%q maxclock=%d minclock=%d minsane=%d hw_timestamps=%v",
			settings.Mode, settings.NTPServers, settings.NTPMaxClocks, settings.NTPMinClocks, settings.NTPMinSane, settings.PTPHardwareTimestamp)
		prettyPrint(t, "TimeSettings", settings)
	})

	t.Run("TimeDashboard", func(t *testing.T) {
		dash, err := client.TimeDashboard.Get(ctx)
		if err != nil {
			t.Fatalf("TimeDashboard.Get failed: %v", err)
		}
		if dash.Settings == nil || dash.Settings.Mode == "" {
			t.Fatalf("dashboard settings mode is empty: %+v", dash.Settings)
		}
		if len(dash.Nodes) == 0 {
			t.Fatal("time dashboard returned no nodes")
		}
		synced := 0
		for _, node := range dash.Nodes {
			status := vergeos.TimeSyncStatus{}
			if node.Machine != nil {
				status = node.Machine.Status.TimeSyncStatus
			}
			t.Logf("  node %s present=%v mode=%s status=%s synced=%v", node.Name, status.Present, status.Mode, status.Status, status.Synced)
			if status.Present && (status.Mode != "" || status.Status != "") {
				synced++
			}
		}
		if synced == 0 {
			t.Fatal("no dashboard node reported time_sync_status")
		}
	})

	t.Run("UpdatePackages", func(t *testing.T) {
		packages, err := client.UpdatePackages.List(ctx)
		if err != nil {
			t.Fatalf("UpdatePackages.List failed: %v", err)
		}
		t.Logf("Found %d update packages", len(packages))
		if len(packages) == 0 {
			return
		}
		first := packages[0]
		t.Logf("  %s %s dependencies=%v", first.Name, first.Version, []string(first.Dependencies))
		fetched, err := client.UpdatePackages.Get(ctx, first.Name)
		if err != nil && first.Key.String() != "" {
			fetched, err = client.UpdatePackages.Get(ctx, first.Key.String())
		}
		if err != nil {
			t.Fatalf("UpdatePackages.Get failed: %v", err)
		}
		t.Logf("Get %s dependencies=%v", fetched.Name, []string(fetched.Dependencies))
	})

	t.Run("UpdateSourceStatus", func(t *testing.T) {
		rows, err := client.UpdateSourceStatus.List(ctx)
		if err != nil {
			t.Fatalf("UpdateSourceStatus.List failed: %v", err)
		}
		for _, row := range rows {
			t.Logf("source %d (%s) status=%s", row.Source, row.SourceName, row.Status)
		}
	})

	t.Run("UpdateSettingsRestartPending", func(t *testing.T) {
		settings, err := client.UpdateSettings.Get(ctx)
		if err != nil {
			t.Fatalf("UpdateSettings.Get failed: %v", err)
		}
		t.Logf("restart_pending_packages=%v", []string(settings.RestartPendingPackages))
	})

	t.Run("VMFields", func(t *testing.T) {
		vms, err := client.VMs.List(ctx, vergeos.WithFilter("is_snapshot eq false"), vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("VMs.List failed: %v", err)
		}
		if len(vms) == 0 {
			t.Fatal("no VMs to check 26.2 fields")
		}
		vm, err := client.VMs.Get(ctx, int(vms[0].Key))
		if err != nil {
			t.Fatalf("VMs.Get failed: %v", err)
		}
		if vm.BIOSSerial == "" {
			t.Fatalf("vm %s bios_serial is empty", vm.Name)
		}
		if !validVGAMem(vm.VGAMemMB) {
			t.Fatalf("vm %s vgamem_mb = %q", vm.Name, vm.VGAMemMB)
		}
		t.Logf("vm %s bios_serial=%q vgamem_mb=%s machine_type=%s", vm.Name, vm.BIOSSerial, vm.VGAMemMB, vm.MachineType)
		if vm.Machine != 0 {
			status, err := client.MachineStatus.Get(ctx, vm.Machine)
			if err != nil {
				t.Fatalf("MachineStatus.Get failed: %v", err)
			}
			if status.TimeSyncStatus.Present && status.TimeSyncStatus.Mode == "" && status.TimeSyncStatus.Status == "" {
				t.Fatalf("time_sync_status object has no mode or status: %+v", status.TimeSyncStatus)
			}
			t.Logf("agent_ip_list=%q time_sync present=%v mode=%s status=%s",
				status.AgentIPList, status.TimeSyncStatus.Present, status.TimeSyncStatus.Mode, status.TimeSyncStatus.Status)
		}
		drives, err := client.VMDrives.List(ctx, int(vm.Key))
		if err != nil {
			t.Fatalf("VMDrives.List failed: %v", err)
		}
		for _, drive := range drives {
			t.Logf("  drive %s uuid=%s primary=%v share=%v ovmf_code=%s", drive.Name, drive.UUID, drive.IsPrimary, drive.Share, drive.OVMFCode)
		}
	})

	t.Run("Files", func(t *testing.T) {
		files, err := client.Files.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("Files.List failed: %v", err)
		}
		for _, file := range files {
			t.Logf("file %s type=%s uuid=%s", file.Name, file.Type, file.UUID)
		}
	})

	t.Run("Networks", func(t *testing.T) {
		networks, err := client.Networks.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("Networks.List failed: %v", err)
		}
		for _, network := range networks {
			console := 0
			if network.ConsoleStatus != nil {
				console = network.ConsoleStatus.Int()
			}
			t.Logf("network %s dhcp_broadcast=%v vlan_bond_mode=%s bgp_advanced=%v console=%d",
				network.Name, network.DHCPBroadcast, network.VLANBondMode, network.BGPAdvanced, console)
		}
		if len(networks) == 0 {
			t.Fatal("no networks")
		}
		id := int(networks[0].Key)
		members, err := client.VNetBondInterfaces.ListByNetwork(ctx, id)
		if err != nil {
			t.Fatalf("VNetBondInterfaces.ListByNetwork failed: %v", err)
		}
		t.Logf("bond members on %s: %d", networks[0].Name, len(members))
		rows, err := client.VNetBGP.List(ctx, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("VNetBGP.List failed: %v", err)
		}
		if len(rows) == 0 {
			t.Log("no vnet_bgp rows")
			return
		}
		row, err := client.VNetBGP.Get(ctx, int(rows[0].Key))
		if err != nil {
			t.Fatalf("VNetBGP.Get failed: %v", err)
		}
		if row.Key == 0 {
			t.Fatal("vnet_bgp key read back empty")
		}
		t.Logf("vnet_bgp %d advanced=%v config_len=%d bgpd=%v rows=%d cols=%d",
			int(row.Key), row.Advanced, len(row.AdvancedConfig), row.BGP, row.ConsoleRows, row.ConsoleCols)
	})

	t.Run("Schema", func(t *testing.T) {
		machines, err := client.Schema.GetVMMachineTypes(ctx)
		if err != nil {
			t.Fatalf("GetVMMachineTypes failed: %v", err)
		}
		cpus, err := client.Schema.GetVMCPUTypes(ctx)
		if err != nil {
			t.Fatalf("GetVMCPUTypes failed: %v", err)
		}
		t.Logf("machine types=%d cpu types=%d", len(machines), len(cpus))
		if _, ok := machines["pc-q35-11.0"]; !ok {
			t.Logf("pc-q35-11.0 is not in the machine type list (%d entries)", len(machines))
		}
	})
}

func validVGAMem(value string) bool {
	switch value {
	case vergeos.VGAMemDefault, vergeos.VGAMem8, vergeos.VGAMem16, vergeos.VGAMem32, vergeos.VGAMem64, vergeos.VGAMem128, vergeos.VGAMem256:
		return true
	default:
		return false
	}
}
