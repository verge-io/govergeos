---
title: System
description: Query nodes, clusters, settings, and system version information
tags: [node, cluster, settings, system, version, status, infrastructure]
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
