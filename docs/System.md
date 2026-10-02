---
title: System
description: Query nodes, clusters, settings, system version, and machine metrics
tags: [node, cluster, settings, system, version, status, infrastructure, machine-status, machine-stats, machine-nic, machine-drive-stats]
categories: [System]
---

# System

Query nodes, clusters, settings, and system information.

## Nodes

```go
// List all nodes
nodes, err := client.Nodes.List(ctx)

// Get a specific node
node, err := client.Nodes.Get(ctx, nodeID)

// List with options
nodes, err := client.Nodes.List(ctx,
    vergeos.WithFilter("physical eq true"),
    vergeos.WithFields("dashboard"),
)
```

---

## Hardware inventory

GPUs, DIMMs, and LLDP neighbors are separate read services. `NodeGPUs.Update` is the call that changes a GPU mode.

```go
// NVIDIA vGPU profiles the driver published
profiles, err := client.VGPUProfiles.List(ctx)
compute, err := client.VGPUProfiles.ListByProfileType(ctx, vergeos.VGPUProfileTypeCompute)

// GPUs configured on a node, then switch one to PCI passthrough
gpus, err := client.NodeGPUs.ListByNode(ctx, nodeID)
mode := vergeos.GPUModePassthrough
gpu, err := client.NodeGPUs.Update(ctx, gpuID, &vergeos.NodeGPUUpdateRequest{
    Mode: &mode,
})
fmt.Println(gpu.ModeDisplay())

// DIMMs that are not online
dimms, err := client.NodeMemory.ListByNode(ctx, nodeID)
for _, dimm := range dimms {
    if !dimm.IsHealthy() {
        fmt.Println(dimm.Locator, dimm.Status)
    }
}

// Switch port discovered for each NIC on the node
neighbors, err := client.NodeLLDPNeighbors.ListByNode(ctx, nodeID)
for _, neighbor := range neighbors {
    fmt.Println(neighbor.NIC, neighbor.ChassisName(), neighbor.PortID())
}
```

---

## Machine metrics

Read-only machine tables. `MachineStatus` and `MachineStats` key off the machine ID (`VM.Machine` for a VM). `MachineNICs` and `MachineDriveStats` expose per-NIC and per-drive counters.

```go
// Runtime status for a machine (power, node, guest agent)
status, err := client.MachineStatus.Get(ctx, machineID)
statuses, err := client.MachineStatus.List(ctx)

// CPU, RAM, and temperature metrics
stats, err := client.MachineStats.GetByMachine(ctx, machineID)
stats, err = client.MachineStats.Get(ctx, statsRowID)

// Per-NIC traffic and link status (physical nodes)
nics, err := client.MachineNICs.ListByMachine(ctx, machineID)
nic, err := client.MachineNICs.Get(ctx, nicID)

// Per-drive I/O stats
driveStats, err := client.MachineDriveStats.ListPhysical(ctx)
driveStat, err := client.MachineDriveStats.GetByDrive(ctx, driveID)
```

---

## Clusters

Manage VergeOS clusters for compute and storage resources.

```go
// List all clusters
clusters, err := client.Clusters.List(ctx)

// Get cluster details
cluster, err := client.Clusters.Get(ctx, clusterID)

// Get a cluster by name
cluster, err := client.Clusters.GetByName(ctx, "vergeos")

// Get cluster status (nodes, RAM, cores, running machines)
status, err := client.Clusters.GetStatus(ctx, clusterID)
fmt.Printf("Status: %s, Nodes: %d/%d, Used RAM: %dMB/%dMB\n",
    status.Status, status.OnlineNodes, status.TotalNodes,
    status.UsedRAM, status.OnlineRAM)

// Create a cluster
cluster, err := client.Clusters.Create(ctx, &vergeos.ClusterCreateRequest{
    Name:        "new-cluster",
    Description: "Production compute cluster",
    Compute:     ptr(true),    // Compute cluster
    KVMNested:   ptr(false),   // Nested virtualization
    DefaultCPU:  ptr("qemu64"),
    RAMPerUnit:  ptr(4096),
    MaxRAMPerVM: ptr(65536),
})

// Update a cluster
newDesc := "Updated cluster description"
cluster, err := client.Clusters.Update(ctx, clusterID, &vergeos.ClusterUpdateRequest{
    Description:   &newDesc,
    MaxCoresPerVM: ptr(32),
    TargetRAMPct:  ptr(float64(85)),  // Target 85% RAM utilization
})

// Delete a cluster (requires no nodes/machines referencing it)
err = client.Clusters.Delete(ctx, clusterID)
```

---

## Billing

Usage records live on `billing`. `List` returns the newest record first. `Generate` posts action `generate` to `billing_actions`. RAM on the record is megabytes. Tier storage is bytes.

```go
records, err := client.Billing.ListCreated(ctx, since, 0)

latest, err := client.Billing.GetLatest(ctx)
fmt.Printf("CPU %.1f%%  RAM %.1fGB\n", latest.CPUUtilization(), latest.UsedRAMGB())

err = client.Billing.Generate(ctx)

summary, err := client.Billing.GetSummary(ctx, since, 0)
fmt.Printf("%d records, peak %d cores\n", summary.RecordCount, summary.PeakCPUCores)
```

---

## Settings

```go
// Get system settings
settings, err := client.Settings.List(ctx)

// Get a specific setting by key
setting, err := client.Settings.GetByKey(ctx, "cloud_name")
fmt.Printf("Cloud: %s\n", setting.Value)

// Convenience method for cloud name
cloudName, err := client.Settings.GetCloudName(ctx)
```

---

## System Info

```go
// Get system version info (uses /version.json endpoint)
info, err := client.System.GetInfo(ctx)
fmt.Printf("API: %s, Version: %s\n", info.Name, info.Version)

// Get just the version string
version, err := client.System.GetVersion(ctx)
```
