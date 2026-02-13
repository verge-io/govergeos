---
title: Tenants
description: Manage multi-tenant virtual data centers, nodes, storage, snapshots, and Layer 2 networks
tags: [tenant, multi-tenant, vdc, tenant-node, tenant-storage, tenant-snapshot, layer2, isolation, clone]
categories: [Tenants]
---

# Tenants

Manage multi-tenant virtual data centers.

## Tenant Lifecycle

```go
// List all tenants
tenants, err := client.Tenants.List(ctx)

// Get a tenant by ID
tenant, err := client.Tenants.Get(ctx, tenantID)

// Get a tenant by name
tenant, err := client.Tenants.GetByName(ctx, "customer-a")

// Create a tenant
tenant, err := client.Tenants.Create(ctx, &vergeos.TenantCreateRequest{
    Name:        "customer-a",
    Password:    "admin-password",
    Description: "Customer A production environment",
})

// Update a tenant
tenant, err := client.Tenants.Update(ctx, tenantID, &vergeos.TenantUpdateRequest{
    Description: ptr("Updated description"),
})

// Delete a tenant
err = client.Tenants.Delete(ctx, tenantID)

// Power operations
err = client.Tenants.PowerOn(ctx, tenantID)
err = client.Tenants.PowerOff(ctx, tenantID)
err = client.Tenants.Reset(ctx, tenantID)

// Clone a tenant
err = client.Tenants.Clone(ctx, tenantID, &vergeos.TenantCloneOptions{
    Name:      "customer-a-clone",
    NoStorage: false,
    NoNodes:   false,
})

// Network isolation
err = client.Tenants.IsolateOn(ctx, tenantID)
err = client.Tenants.IsolateOff(ctx, tenantID)
```

---

## Tenant Nodes

Manage virtual nodes within tenants.

```go
// List nodes for a tenant
nodes, err := client.TenantNodes.ListByTenant(ctx, tenantID)

// Create a tenant node
node, err := client.TenantNodes.Create(ctx, &vergeos.TenantNodeCreateRequest{
    Tenant:   tenantID,
    Name:     "node1",
    CPUCores: 4,
    RAM:      16384, // 16GB in MB
})

// Power operations
err = client.TenantNodes.PowerOn(ctx, nodeID)
err = client.TenantNodes.PowerOff(ctx, nodeID)
err = client.TenantNodes.Reset(ctx, nodeID)

// Migrate a node to another host
err = client.TenantNodes.Migrate(ctx, nodeID, targetHostNodeID)
```

---

## Tenant Storage

Manage storage allocations for tenants.

```go
// List storage allocations for a tenant
storage, err := client.TenantStorage.ListByTenant(ctx, tenantID)

// Create a storage allocation
storage, err := client.TenantStorage.Create(ctx, &vergeos.TenantStorageCreateRequest{
    Tenant:      tenantID,
    Tier:        tierID,
    Provisioned: 100 * 1024 * 1024 * 1024, // 100GB in bytes
})

// Update storage allocation
storage, err := client.TenantStorage.Update(ctx, storageID, &vergeos.TenantStorageUpdateRequest{
    Provisioned: ptr(int64(200 * 1024 * 1024 * 1024)), // 200GB
})
```

---

## Tenant Snapshots

Manage point-in-time snapshots of tenants.

```go
// List all tenant snapshots
snapshots, err := client.TenantSnapshots.List(ctx)

// List snapshots for a specific tenant
snapshots, err := client.TenantSnapshots.ListByTenant(ctx, tenantID)

// List snapshots expiring within 7 days
expiring, err := client.TenantSnapshots.ListExpiring(ctx, 7)

// Get a snapshot by ID
snapshot, err := client.TenantSnapshots.Get(ctx, snapshotID)

// Get a snapshot by name within a tenant
snapshot, err := client.TenantSnapshots.GetByName(ctx, tenantID, "pre-upgrade")

// Update a snapshot
newDesc := "Updated description"
snapshot, err := client.TenantSnapshots.Update(ctx, snapshotID, &vergeos.TenantSnapshotUpdateRequest{
    Description: &newDesc,
})

// Set snapshot to never expire
snapshot, err := client.TenantSnapshots.SetNeverExpires(ctx, snapshotID)

// Set snapshot expiration (Unix timestamp)
expires := time.Now().Add(7 * 24 * time.Hour).Unix()
snapshot, err := client.TenantSnapshots.SetExpires(ctx, snapshotID, expires)

// Refresh tenant snapshots from snapshot profile
err = client.TenantSnapshots.Refresh(ctx, tenantID)

// Delete a snapshot
err = client.TenantSnapshots.Delete(ctx, snapshotID)
```

---

## Tenant Layer2 Networks

Manage Layer 2 network assignments to tenants for direct network connectivity.

```go
// List all tenant layer2 network assignments
networks, err := client.TenantLayer2Networks.List(ctx)

// List assignments for a specific tenant
networks, err := client.TenantLayer2Networks.ListByTenant(ctx, tenantID)

// List tenants assigned to a specific network
networks, err := client.TenantLayer2Networks.ListByNetwork(ctx, networkID)

// Get an assignment by ID
assignment, err := client.TenantLayer2Networks.Get(ctx, assignmentID)

// Get assignment by tenant and network
assignment, err := client.TenantLayer2Networks.GetByTenantAndNetwork(ctx, tenantID, networkID)

// Create an assignment (assign network to tenant)
assignment, err := client.TenantLayer2Networks.Create(ctx, &vergeos.TenantLayer2NetworkCreateRequest{
    Tenant:  tenantID,
    VNet:    networkID,
    Enabled: ptr(true),
})

// Update an assignment
assignment, err := client.TenantLayer2Networks.Update(ctx, assignmentID, &vergeos.TenantLayer2NetworkUpdateRequest{
    Enabled: ptr(false),
})

// Enable/disable an assignment
assignment, err := client.TenantLayer2Networks.Enable(ctx, assignmentID)
assignment, err := client.TenantLayer2Networks.Disable(ctx, assignmentID)

// Convenience method: Assign a network to a tenant (creates if not exists)
assignment, err := client.TenantLayer2Networks.Assign(ctx, tenantID, networkID)

// Convenience method: Unassign a network from a tenant
err = client.TenantLayer2Networks.Unassign(ctx, tenantID, networkID)

// Delete an assignment
err = client.TenantLayer2Networks.Delete(ctx, assignmentID)
```
