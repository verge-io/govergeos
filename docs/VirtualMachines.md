---
title: Virtual Machines
description: Manage VM lifecycle, snapshots, drives, NICs, cloning, and migration
tags: [vm, virtual-machine, snapshot, drive, nic, clone, migration, power-management, console, restore]
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
err = client.VMs.PowerOn(ctx, vmID)
err = client.VMs.PowerOff(ctx, vmID)

// Clone a VM
err = client.VMs.Clone(ctx, vmID, &vergeos.VMCloneOptions{
    Name:         "my-vm-clone",
    PreserveMACs: false,
})

// Take a snapshot
err = client.VMs.Snapshot(ctx, vmID, &vergeos.VMSnapshotOptions{
    Retention: 86400, // 24 hours
    Quiesce:   true,
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

// Create a snapshot
snapshot, err := client.VMSnapshots.Create(ctx, &vergeos.VMSnapshotCreateRequest{
    Machine:     vmID,
    Name:        "pre-upgrade",
    Description: "Snapshot before upgrade",
    ExpiresType: "date", // or "never"
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

```go
// List drives for a VM
drives, err := client.VMDrives.List(ctx, vmMachineID)

// Create a drive
drive, err := client.VMDrives.Create(ctx, vmMachineID, &vergeos.VMDriveCreateRequest{
    Name:      "disk0",
    Interface: "virtio",
    Media:     "disk",
    SizeGB:    50,
})

// Update a drive (resize)
drive, err := client.VMDrives.Update(ctx, driveID, &vergeos.VMDriveUpdateRequest{
    SizeGB: ptr(int64(100)), // Increase size
})
```

---

## VM NICs

```go
// List NICs for a VM
nics, err := client.VMNICs.List(ctx, vmMachineID)

// Create a NIC
nic, err := client.VMNICs.Create(ctx, vmMachineID, &vergeos.VMNICCreateRequest{
    Name: "eth0",
    VNET: networkID,
})
```
