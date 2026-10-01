---
title: Virtual Machines
description: Manage VM lifecycle, snapshots, drives, NICs, cloning, and migration
tags: [vm, virtual-machine, snapshot, drive, nic, clone, migration, power-management, console, restore, cloud-init]
categories: [Virtual Machines]
---

# Virtual Machines

Manage virtual machine lifecycle, configuration, and hardware.

## VM Lifecycle

```go
// List all VMs
vms, err := client.VMs.List(ctx)

// Get a specific VM
vm, err := client.VMs.Get(ctx, vmID)

// Create a new VM
vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
    Name:     "my-vm",
    CPUCores: 4,
    RAM:      8192,
    Cluster:  &clusterID,
})

// Update a VM
newName := "renamed-vm"
vm, err := client.VMs.Update(ctx, vmID, &vergeos.VMUpdateRequest{
    Name: &newName,
})

// Delete a VM
err = client.VMs.Delete(ctx, vmID)

// Power operations
// PowerOn waits until the VM is running.
err = client.VMs.PowerOn(ctx, vmID)
// PowerOff asks the guest to shut down (ACPI poweroff) and waits until it stops.
err = client.VMs.PowerOff(ctx, vmID)
// GuestShutdown sends the same poweroff request and returns immediately.
err = client.VMs.GuestShutdown(ctx, vmID)
// Kill is an immediate power-off (the kill action). It waits until the VM stops.
err = client.VMs.Kill(ctx, vmID)
// Give a slow guest more than the default 150 seconds, then kill if it is still running.
err = client.VMs.PowerOffWithOptions(ctx, vmID, &vergeos.VMPowerOffOptions{
    Timeout:           10 * time.Minute,
    PollInterval:      5 * time.Second,
    ForceAfterTimeout: true,
})

// Clone a VM
err = client.VMs.Clone(ctx, vmID, &vergeos.VMCloneOptions{
    Name:         "my-vm-clone",
    PreserveMACs: false,
})

// Take a snapshot. Retention is a lifetime in seconds, sent as expires.
// Quiesce requires a running guest agent; leave it false to snapshot
// a VM that does not have one.
snapshot, err := client.VMs.Snapshot(ctx, vmID, &vergeos.VMSnapshotOptions{
    Name:      "pre-upgrade",
    Retention: 86400, // 24 hours
})

// Migrate a VM to another node
err = client.VMs.Migrate(ctx, vmID, &vergeos.VMMigrateOptions{
    TargetNode: targetNodeID,
    Live:       ptr(true), // Live migration
})

// Get console URL
consoleURL, err := client.VMs.GetConsoleURL(ctx, vmID)
```

---

## VM Import and Export

Import an OVA or OVF from a file already in the media catalog, from an http(s) URL, from a NAS volume path, or from a shared object. The URL is downloaded into the catalog and then imported. `Wait` returns as soon as the import reports an error, an abort, or a failed drive. It does not sit through the rest of the timeout.

`vm_imports` accepts an OVA or OVF. VergeOS 26.1 rejects a bare disk image (qcow2, vmdk, vhd, raw, img) with status `error` and `Unknown Import file type`. The import row has no field that selects a disk-image mode. Attach that file to an existing VM with `VMDrives.Create` and `Media` set to `import`.

VergeOS keeps every import row and does not require those names to be unique. `Create` does not post again when a VM with that name already exists (`AlreadyExisted`). `GetByName` returns `AmbiguousNameError` when two rows share a name. `DeleteByName` removes every finished row with the name, and leaves them all in place when one of several rows is still importing.

```go
fileID := 41
imp, err := client.VMImports.Create(ctx, &vergeos.VMImportCreateRequest{
    Name: "imported-vm",
    File: &fileID,
})
if !imp.AlreadyExisted {
    imp, err = client.VMImports.Wait(ctx, imp.Key, nil)
}
logs, err := client.VMImports.Logs(ctx, imp.Key)

imp, err = client.VMImports.Create(ctx, &vergeos.VMImportCreateRequest{
    Name: "from-ova",
    URL:  "https://example.com/images/guest.ova",
})

err = client.VMImports.DeleteByName(ctx, "imported-vm")
```

Export writes a VM onto a NAS volume. `Run` creates the volume's export configuration when it does not exist, then starts the export. `Wait` returns as soon as the export reports an error.

```go
volumeKey := "c3437883534918dcf2abbb3e9b9622b865226c68"
exp, err := client.VMExports.Run(ctx, &vergeos.VMExportRunRequest{
    Volume: volumeKey,
    VMs:    vmIDs,
    Name:   "nightly",
})
exp, err = client.VMExports.Wait(ctx, exp.Key.Int(), nil)
stats, err := client.VMExports.Stats(ctx, exp.Key.Int())
```

---

## VM Snapshots

Manage point-in-time snapshots of virtual machines.

```go
// List all VM snapshots
snapshots, err := client.VMSnapshots.List(ctx)

// List snapshots for a specific VM
snapshots, err := client.VMSnapshots.ListByVM(ctx, vmID)

// List snapshots expiring within 7 days
expiring, err := client.VMSnapshots.ListExpiring(ctx, 7)

// Get a snapshot by ID
snapshot, err := client.VMSnapshots.Get(ctx, snapshotID)

// Get a snapshot by name within a VM
snapshot, err := client.VMSnapshots.GetByName(ctx, vmID, "pre-upgrade")

// Create a snapshot. VM is the VM $key (VM.Key). Create resolves the machine key.
snapshot, err := client.VMSnapshots.Create(ctx, &vergeos.VMSnapshotCreateRequest{
    VM:          vmID,
    Name:        "pre-upgrade",
    Description: "Snapshot before upgrade",
    ExpiresType: "date", // write-only: "never" stores Expires as 0. Not returned.
})

// Update a snapshot
newDesc := "Updated description"
snapshot, err := client.VMSnapshots.Update(ctx, snapshotID, &vergeos.VMSnapshotUpdateRequest{
    Description: &newDesc,
})

// Set snapshot to never expire
snapshot, err := client.VMSnapshots.SetNeverExpires(ctx, snapshotID)

// Set snapshot expiration (Unix timestamp)
expires := time.Now().Add(7 * 24 * time.Hour).Unix()
snapshot, err := client.VMSnapshots.SetExpires(ctx, snapshotID, expires)

// Restore a VM from snapshot
err = client.VMSnapshots.Restore(ctx, snapshotID, &vergeos.VMSnapshotRestoreOptions{
    PowerOn: true, // Power on VM after restore
})

// Delete a snapshot
err = client.VMSnapshots.Delete(ctx, snapshotID)
```

---

## VM Drives

Drive, NIC, and device methods take the VM `$key` (`VM.Key`). Rows in
`machine_drives`, `machine_nics`, and `machine_devices` are stored against
the machine key (`VM.Machine`), which these methods resolve internally.
`vm_actions` (hotplug, restore) uses the VM `$key`. The two numbers differ
on a system that has been in use.

```go
// List drives for a VM
drives, err := client.VMDrives.List(ctx, vmID)

// Create a drive
drive, err := client.VMDrives.Create(ctx, vmID, &vergeos.VMDriveCreateRequest{
    Name:      "disk0",
    Interface: "virtio",
    Media:     "disk",
    SizeGB:    50,
})

// Attach a media-catalog disk image (qcow2, vmdk, vhd, raw, img).
// Create waits until the drive leaves status "importing".
// VMImports rejects these file types.
fileID := 41
drive, err = client.VMDrives.Create(ctx, vmID, &vergeos.VMDriveCreateRequest{
    Name:  "imported-disk",
    Media: "import",
    File:  fileID,
})

// Update a drive (resize)
drive, err := client.VMDrives.Update(ctx, driveID, &vergeos.VMDriveUpdateRequest{
    SizeGB: ptr(int64(100)), // Increase size
})
```

---

## Cloud-Init Files

Every cloud-init file belongs to one VM. `owner` is required on create, cannot be changed later, and is a reference that uses the VM $key (`vms/<VM.Key>`). It is not the machine key (`VM.Machine`).

```go
file, err := client.CloudInitFiles.CreateForVM(ctx, vm.Key.Int(), &vergeos.CloudInitFileCreateRequest{
    Name:     "/user-data",
    Contents: "#cloud-config\nhostname: web-01\n",
})

// List and Get leave Contents empty. VergeOS returns the body only from
// GET /cloudinit_files/{id}?download=1.
contents, err := client.CloudInitFiles.GetContents(ctx, file.Key.Int())

files, err := client.CloudInitFiles.ListByVM(ctx, vm.Key.Int())
```

`Create` takes the same request with `Owner` set to that reference.

---

## VM NICs

```go
// List NICs for a VM
nics, err := client.VMNICs.List(ctx, vmID)

// Create a NIC
nic, err := client.VMNICs.Create(ctx, vmID, &vergeos.VMNICCreateRequest{
    Name: "eth0",
    VNET: networkID,
})
```
