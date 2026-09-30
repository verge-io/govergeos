// Command sweep deletes integration-test objects left on a VergeOS system.
//
// It removes objects whose names start with "sdk-" or "goVergeOS-test-",
// and children of those objects (for example a DNS view named "test-view" on
// an sdk- network, or a cloud-init file on an sdk- VM). Other objects are
// left alone.
//
//	export VERGEOS_HOST=https://lab.example.com
//	export VERGEOS_API_KEY=...
//	go run ./test/integration/sweep
//	go run ./test/integration/sweep -fail-if-left
//
// -fail-if-left exits non-zero when this run deleted anything or when a test
// object is still present. The integration workflow uses that on the sweep
// after the suite so a leaked object fails the job.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

func main() {
	failIfLeft := flag.Bool("fail-if-left", false, "exit 1 when this run deletes a test object or any remain")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %s\n", strings.Join(flag.Args(), " "))
		os.Exit(2)
	}

	client, err := vergeos.NewClient(
		vergeos.WithEnvConfig(),
		vergeos.WithTimeout(2*time.Minute),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	live := &sweeper{client: client, ctx: ctx}
	live.run()
	if len(live.failures) > 0 {
		fmt.Fprintf(os.Stderr, "sweep failed with %d error(s)\n", len(live.failures))
		os.Exit(1)
	}

	left := &sweeper{client: client, ctx: ctx, dry: true}
	left.run()
	if len(left.failures) > 0 || left.matched > 0 {
		fmt.Fprintf(os.Stderr, "%d test object(s) still present\n", left.matched)
		os.Exit(1)
	}
	if *failIfLeft && live.deleted > 0 {
		fmt.Fprintf(os.Stderr, "removed %d leftover test object(s)\n", live.deleted)
		os.Exit(1)
	}
	fmt.Printf("removed %d test object(s)\n", live.deleted)
}

// isTestObjectName reports whether name belongs to an object the integration
// suite creates. Matching is case-sensitive and prefix-based so names such as
// "test-view" on a production network are not selected on their own.
func isTestObjectName(name string) bool {
	return strings.HasPrefix(name, "sdk-") || strings.HasPrefix(name, "goVergeOS-test-")
}

type sweeper struct {
	client   *vergeos.Client
	ctx      context.Context
	dry      bool
	matched  int
	deleted  int
	failures []string
}

func (s *sweeper) run() {
	// Top-level test objects first, then VMs, tenants, and networks.
	// Children of an sdk- VM or network are removed with that parent even
	// when the child's own name does not use a test prefix.
	s.sweepTasks(nil)
	s.sweepCloudSnapshots()
	s.sweepShares()
	s.sweepVolumes()
	s.sweepTags()
	s.sweepWebhooks()
	s.sweepCertificates()
	s.sweepFiles()
	s.sweepUsers()
	s.sweepGroups()
	s.sweepVMs()
	s.sweepOrphanDrives()
	s.sweepOrphanNICs()
	s.sweepOrphanSnapshots()
	s.sweepClusters()
	s.sweepTenants()
	s.sweepNetworks()
	s.sweepRuleAliases()
	s.sweepSnapshotProfiles()
}

func (s *sweeper) listErr(kind string, err error) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf("list %s: %v", kind, err)
	fmt.Fprintln(os.Stderr, msg)
	s.failures = append(s.failures, msg)
}

func (s *sweeper) remove(kind, name string, fn func() error) {
	label := kind + " " + name
	s.matched++
	if s.dry {
		fmt.Println("would delete", label)
		return
	}
	fmt.Println("delete", label)
	if err := fn(); err != nil {
		if vergeos.IsNotFoundError(err) {
			return
		}
		msg := fmt.Sprintf("delete %s: %v", label, err)
		fmt.Fprintln(os.Stderr, msg)
		s.failures = append(s.failures, msg)
		return
	}
	s.deleted++
}

func (s *sweeper) sweepTasks(owner *string) {
	var (
		tasks []vergeos.Task
		err   error
	)
	if owner == nil {
		tasks, err = s.client.Tasks.List(s.ctx)
	} else {
		tasks, err = s.client.Tasks.ListByOwner(s.ctx, *owner)
	}
	if err != nil {
		s.listErr("tasks", err)
		return
	}
	for _, task := range tasks {
		if owner == nil && !isTestObjectName(task.Name) {
			continue
		}
		task := task
		if !s.dry {
			if err := s.client.Tasks.Disable(s.ctx, int(task.Key)); err != nil && !vergeos.IsNotFoundError(err) {
				fmt.Fprintf(os.Stderr, "disable task %s: %v\n", task.Name, err)
			}
		}
		s.remove("task", task.Name, func() error {
			return s.client.Tasks.Delete(s.ctx, int(task.Key))
		})
	}
}

func (s *sweeper) sweepCloudSnapshots() {
	items, err := s.client.CloudSnapshots.List(s.ctx)
	if err != nil {
		s.listErr("cloud snapshots", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("cloud snapshot", item.Name, func() error {
			return s.client.CloudSnapshots.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepShares() {
	cifs, err := s.client.VolumeCIFSShares.List(s.ctx)
	if err != nil {
		s.listErr("CIFS shares", err)
	} else {
		for _, item := range cifs {
			if !isTestObjectName(item.Name) {
				continue
			}
			item := item
			s.remove("CIFS share", item.Name, func() error {
				return s.client.VolumeCIFSShares.Delete(s.ctx, shareID(item.Key, item.ID))
			})
		}
	}

	nfs, err := s.client.VolumeNFSShares.List(s.ctx)
	if err != nil {
		s.listErr("NFS shares", err)
		return
	}
	for _, item := range nfs {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("NFS share", item.Name, func() error {
			return s.client.VolumeNFSShares.Delete(s.ctx, shareID(item.Key, item.ID))
		})
	}
}

func shareID(key, id string) string {
	if key != "" {
		return key
	}
	return id
}

func (s *sweeper) sweepVolumes() {
	items, err := s.client.Volumes.List(s.ctx)
	if err != nil {
		s.listErr("volumes", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("volume", item.Name, func() error {
			return s.client.Volumes.Delete(s.ctx, shareID(item.Key, item.ID))
		})
	}
}

func (s *sweeper) sweepTags() {
	categories, err := s.client.TagCategories.List(s.ctx)
	if err != nil {
		s.listErr("tag categories", err)
		return
	}
	testCategories := map[int]string{}
	for _, category := range categories {
		if isTestObjectName(category.Name) {
			testCategories[int(category.Key)] = category.Name
		}
	}

	tags, err := s.client.Tags.List(s.ctx)
	if err != nil {
		s.listErr("tags", err)
	} else {
		for _, tag := range tags {
			if !isTestObjectName(tag.Name) && testCategories[int(tag.Category)] == "" {
				continue
			}
			tag := tag
			s.remove("tag", tag.Name, func() error {
				return s.client.Tags.Delete(s.ctx, int(tag.Key))
			})
		}
	}

	for _, category := range categories {
		if !isTestObjectName(category.Name) {
			continue
		}
		category := category
		s.remove("tag category", category.Name, func() error {
			return s.client.TagCategories.Delete(s.ctx, int(category.Key))
		})
	}
}

func (s *sweeper) sweepWebhooks() {
	items, err := s.client.WebhookURLs.List(s.ctx)
	if err != nil {
		s.listErr("webhook URLs", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("webhook URL", item.Name, func() error {
			return s.client.WebhookURLs.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepCertificates() {
	items, err := s.client.Certificates.List(s.ctx)
	if err != nil {
		s.listErr("certificates", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Domain) && !certificateListsTestDomain(item.DomainList) {
			continue
		}
		item := item
		label := item.Domain
		if label == "" {
			label = item.DomainList
		}
		s.remove("certificate", label, func() error {
			return s.client.Certificates.Delete(s.ctx, int(item.Key))
		})
	}
}

func certificateListsTestDomain(list string) bool {
	for _, part := range strings.Split(list, ",") {
		if isTestObjectName(strings.TrimSpace(part)) {
			return true
		}
	}
	return false
}

func (s *sweeper) sweepFiles() {
	items, err := s.client.Files.List(s.ctx)
	if err != nil {
		s.listErr("files", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("file", item.Name, func() error {
			return s.client.Files.Delete(s.ctx, int(item.ID))
		})
	}
}

func (s *sweeper) sweepUsers() {
	items, err := s.client.Users.List(s.ctx)
	if err != nil {
		s.listErr("users", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("user", item.Name, func() error {
			return s.client.Users.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepGroups() {
	items, err := s.client.Groups.List(s.ctx)
	if err != nil {
		s.listErr("groups", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("group", item.Name, func() error {
			return s.client.Groups.Delete(s.ctx, int(item.ID))
		})
	}
}

func (s *sweeper) sweepClusters() {
	items, err := s.client.Clusters.List(s.ctx)
	if err != nil {
		s.listErr("clusters", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("cluster", item.Name, func() error {
			return s.client.Clusters.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepTenants() {
	items, err := s.client.Tenants.List(s.ctx)
	if err != nil {
		s.listErr("tenants", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("tenant", item.Name, func() error {
			return s.client.Tenants.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepRuleAliases() {
	items, err := s.client.VNetRuleAliases.List(s.ctx)
	if err != nil {
		s.listErr("rule aliases", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("rule alias", item.Name, func() error {
			return s.client.VNetRuleAliases.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepSnapshotProfiles() {
	profiles, err := s.client.SnapshotProfiles.List(s.ctx)
	if err != nil {
		s.listErr("snapshot profiles", err)
		return
	}
	for _, profile := range profiles {
		if !isTestObjectName(profile.Name) {
			continue
		}
		profile := profile
		if !s.dry {
			s.deleteProfilePeriods(int(profile.Key), profile.Name)
		}
		s.remove("snapshot profile", profile.Name, func() error {
			return s.client.SnapshotProfiles.Delete(s.ctx, int(profile.Key))
		})
	}

	periods, err := s.client.SnapshotProfilePeriods.List(s.ctx)
	if err != nil {
		s.listErr("snapshot profile periods", err)
		return
	}
	for _, period := range periods {
		if !isTestObjectName(period.Name) {
			continue
		}
		period := period
		s.remove("snapshot profile period", period.Name, func() error {
			return s.client.SnapshotProfilePeriods.Delete(s.ctx, int(period.Key))
		})
	}
}

func (s *sweeper) deleteProfilePeriods(profileID int, profileName string) {
	periods, err := s.client.SnapshotProfilePeriods.ListByProfile(s.ctx, profileID)
	if err != nil {
		s.listErr("snapshot profile periods for "+profileName, err)
		return
	}
	for _, period := range periods {
		period := period
		s.remove("snapshot profile period", profileName+"/"+period.Name, func() error {
			return s.client.SnapshotProfilePeriods.Delete(s.ctx, int(period.Key))
		})
	}
}

func (s *sweeper) sweepVMs() {
	items, err := s.client.VMs.List(s.ctx)
	if err != nil {
		s.listErr("VMs", err)
		return
	}
	for _, vm := range items {
		if !isTestObjectName(vm.Name) {
			continue
		}
		vm := vm
		if !s.dry {
			s.deleteVMChildren(vm)
			if err := s.client.VMs.Kill(s.ctx, int(vm.ID)); err != nil && !vergeos.IsNotFoundError(err) {
				fmt.Fprintf(os.Stderr, "kill VM %s: %v\n", vm.Name, err)
			}
		}
		s.remove("VM", vm.Name, func() error {
			return s.client.VMs.Delete(s.ctx, int(vm.ID))
		})
	}
}

func (s *sweeper) deleteVMChildren(vm vergeos.VM) {
	owner := fmt.Sprintf("vms/%d", int(vm.ID))
	s.sweepTasks(&owner)

	files, err := s.client.CloudInitFiles.ListByVM(s.ctx, int(vm.ID))
	if err != nil {
		s.listErr("cloud-init files for "+vm.Name, err)
	} else {
		for _, file := range files {
			file := file
			s.remove("cloud-init file", vm.Name+file.Name, func() error {
				return s.client.CloudInitFiles.Delete(s.ctx, int(file.ID))
			})
		}
	}

	drives, err := s.client.VMDrives.List(s.ctx, int(vm.ID))
	if err != nil {
		s.listErr("drives for "+vm.Name, err)
	} else {
		for _, drive := range drives {
			drive := drive
			s.remove("VM drive", vm.Name+"/"+drive.Name, func() error {
				return s.client.VMDrives.Delete(s.ctx, int(drive.ID))
			})
		}
	}

	nics, err := s.client.VMNICs.List(s.ctx, int(vm.ID))
	if err != nil {
		s.listErr("NICs for "+vm.Name, err)
	} else {
		for _, nic := range nics {
			nic := nic
			s.remove("VM NIC", vm.Name+"/"+nic.Name, func() error {
				return s.client.VMNICs.Delete(s.ctx, int(nic.ID))
			})
		}
	}

	snapshots, err := s.client.VMSnapshots.ListByVM(s.ctx, int(vm.ID))
	if err != nil {
		s.listErr("snapshots for "+vm.Name, err)
		return
	}
	for _, snapshot := range snapshots {
		snapshot := snapshot
		s.remove("VM snapshot", vm.Name+"/"+snapshot.Name, func() error {
			return s.client.VMSnapshots.Delete(s.ctx, int(snapshot.Key))
		})
	}
}

func (s *sweeper) sweepOrphanDrives() {
	items, err := s.client.VMDrives.ListAll(s.ctx)
	if err != nil {
		s.listErr("VM drives", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("VM drive", item.Name, func() error {
			return s.client.VMDrives.Delete(s.ctx, int(item.ID))
		})
	}
}

func (s *sweeper) sweepOrphanNICs() {
	items, err := s.client.MachineNICs.List(s.ctx)
	if err != nil {
		s.listErr("machine NICs", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("VM NIC", item.Name, func() error {
			return s.client.VMNICs.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepOrphanSnapshots() {
	items, err := s.client.VMSnapshots.List(s.ctx)
	if err != nil {
		s.listErr("VM snapshots", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		s.remove("VM snapshot", item.Name, func() error {
			return s.client.VMSnapshots.Delete(s.ctx, int(item.Key))
		})
	}
}

func (s *sweeper) sweepNetworks() {
	items, err := s.client.Networks.List(s.ctx)
	if err != nil {
		s.listErr("networks", err)
		return
	}
	for _, item := range items {
		if !isTestObjectName(item.Name) {
			continue
		}
		item := item
		if !s.dry {
			s.deleteNetworkChildren(int(item.ID), item.Name)
			if err := s.client.Networks.PowerOff(s.ctx, int(item.ID)); err != nil && !vergeos.IsNotFoundError(err) {
				fmt.Fprintf(os.Stderr, "power off network %s: %v\n", item.Name, err)
			}
		}
		s.remove("network", item.Name, func() error {
			return s.client.Networks.Delete(s.ctx, int(item.ID))
		})
	}
}

func (s *sweeper) deleteNetworkChildren(networkID int, networkName string) {
	s.deleteWireGuard(networkID, networkName)

	rules, err := s.client.VNetRules.ListByNetwork(s.ctx, networkID)
	if err != nil {
		s.listErr("rules for "+networkName, err)
	} else {
		for _, rule := range rules {
			rule := rule
			s.remove("network rule", networkName+"/"+rule.Name, func() error {
				return s.client.VNetRules.Delete(s.ctx, int(rule.Key))
			})
		}
	}

	hosts, err := s.client.VNetHosts.ListByNetwork(s.ctx, networkID)
	if err != nil {
		s.listErr("hosts for "+networkName, err)
	} else {
		for _, host := range hosts {
			host := host
			s.remove("network host", networkName+"/"+host.Host, func() error {
				return s.client.VNetHosts.Delete(s.ctx, int(host.Key))
			})
		}
	}

	addresses, err := s.client.VNetAddresses.ListByNetwork(s.ctx, networkID)
	if err != nil {
		s.listErr("addresses for "+networkName, err)
	} else {
		for _, addr := range addresses {
			addr := addr
			label := addr.IP
			if label == "" {
				label = addr.Hostname
			}
			s.remove("network address", networkName+"/"+label, func() error {
				return s.client.VNetAddresses.Delete(s.ctx, int(addr.Key))
			})
		}
	}

	views, err := s.client.VNetDNSViews.ListByNetwork(s.ctx, networkID)
	if err != nil {
		s.listErr("DNS views for "+networkName, err)
		return
	}
	for _, view := range views {
		view := view
		zones, err := s.client.VNetDNSZones.ListByView(s.ctx, int(view.Key))
		if err != nil {
			s.listErr("DNS zones for "+view.Name, err)
		} else {
			for _, zone := range zones {
				zone := zone
				records, err := s.client.VNetDNSRecords.ListByZone(s.ctx, int(zone.Key))
				if err != nil {
					s.listErr("DNS records for "+zone.Domain, err)
				} else {
					for _, record := range records {
						record := record
						s.remove("DNS record", zone.Domain+"/"+record.Host, func() error {
							return s.client.VNetDNSRecords.Delete(s.ctx, int(record.Key))
						})
					}
				}
				s.remove("DNS zone", zone.Domain, func() error {
					return s.client.VNetDNSZones.Delete(s.ctx, int(zone.Key))
				})
			}
		}
		s.remove("DNS view", networkName+"/"+view.Name, func() error {
			return s.client.VNetDNSViews.Delete(s.ctx, int(view.Key))
		})
	}
}

func (s *sweeper) deleteWireGuard(networkID int, networkName string) {
	ifaces, err := s.client.VNetWireGuards.ListByNetwork(s.ctx, networkID)
	if err != nil {
		s.listErr("WireGuard interfaces for "+networkName, err)
		return
	}
	for _, iface := range ifaces {
		iface := iface
		peers, err := s.client.VNetWireGuardPeers.ListByWireGuard(s.ctx, int(iface.Key))
		if err != nil {
			s.listErr("WireGuard peers for "+iface.Name, err)
		} else {
			for _, peer := range peers {
				peer := peer
				s.remove("WireGuard peer", iface.Name+"/"+peer.Name, func() error {
					return s.client.VNetWireGuardPeers.Delete(s.ctx, int(peer.Key))
				})
			}
		}
		s.remove("WireGuard interface", networkName+"/"+iface.Name, func() error {
			return s.client.VNetWireGuards.Delete(s.ctx, int(iface.Key))
		})
	}
}
