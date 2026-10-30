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
		t.Logf("mode=%s ntp_servers=%q ptp_transport=%s ptp_domain=%d hardware_timestamp=%v",
			settings.Mode, settings.NTPServers, settings.PTPTransport, settings.PTPDomain, settings.PTPHardwareTimestamp)
		prettyPrint(t, "TimeSettings", settings)
	})

	t.Run("TimeDashboard", func(t *testing.T) {
		dash, err := client.TimeDashboard.Get(ctx)
		if err != nil {
			t.Fatalf("TimeDashboard.Get failed: %v", err)
		}
		t.Logf("dashboard mode=%s nodes=%d alarms=%d", dash.Mode, len(dash.Nodes), len(dash.Alarms))
		for _, node := range dash.Nodes {
			t.Logf("  node %s status=%s source=%s", node.Name, node.Status, node.Source)
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
		vms, err := client.VMs.List(ctx, vergeos.WithLimit(1))
		if err != nil {
			t.Fatalf("VMs.List failed: %v", err)
		}
		if len(vms) == 0 {
			t.Log("no VMs")
			return
		}
		vm, err := client.VMs.Get(ctx, int(vms[0].Key))
		if err != nil {
			t.Fatalf("VMs.Get failed: %v", err)
		}
		t.Logf("vm %s bios_serial=%q vgamem_mb=%d machine_type=%s", vm.Name, vm.BIOSSerial, vm.VGAMemMB, vm.MachineType)
		if vm.Machine != 0 {
			status, err := client.MachineStatus.Get(ctx, vm.Machine)
			if err != nil {
				t.Fatalf("MachineStatus.Get failed: %v", err)
			}
			t.Logf("agent_ip_list=%v time_sync_status=%s", []string(status.AgentIPList), status.TimeSyncStatus)
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
			t.Logf("network %s dhcp_broadcast=%v vlan_bond_mode=%s bgp_advanced=%v console=%s",
				network.Name, network.DHCPBroadcast, network.VLANBondMode, network.BGPAdvanced, network.BGPConsoleStatus)
		}
		if len(networks) == 0 {
			return
		}
		id := int(networks[0].Key)
		members, err := client.VNetBondInterfaces.ListByNetwork(ctx, id)
		if err != nil {
			t.Fatalf("VNetBondInterfaces.ListByNetwork failed: %v", err)
		}
		t.Logf("bond members on %s: %d", networks[0].Name, len(members))
		router, err := client.Networks.GetAdvancedRouter(ctx, id)
		if err != nil {
			t.Fatalf("Networks.GetAdvancedRouter failed: %v", err)
		}
		t.Logf("advanced router enabled=%v config_len=%d bgpd=%v status=%s",
			router.Enabled, len(router.Config), router.Daemons.BGP, router.ConsoleStatus)
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
