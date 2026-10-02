---
title: Tenants
description: Manage multi-tenant virtual data centers, nodes, storage, snapshots, recipes, Layer 2 networks, shared objects, and the tenant UI proxy
tags: [tenant, multi-tenant, vdc, tenant-node, tenant-storage, tenant-snapshot, tenant-recipe, layer2, isolation, clone, vnet-proxy, shared-object]
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

// Create a tenant (UIAddress picks which external IP is the UI; omit to keep the default)
uiAddress := addressRowID
tenant, err := client.Tenants.Create(ctx, &vergeos.TenantCreateRequest{
    Name:        "customer-a",
    Password:    "admin-password",
    Description: "Customer A production environment",
    UIAddress:   &uiAddress,
    UIFqdn:      "customer-a.example.com",
})

// Update a tenant, including moving the UI address
tenant, err := client.Tenants.Update(ctx, tenantID, &vergeos.TenantUpdateRequest{
    Description: ptr("Updated description"),
    UIAddress:   &uiAddress,
    UIFqdn:      ptr("customer-a.example.com"),
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

// Create a tenant snapshot (POST /tenant_snapshots)
snapshot, err = client.TenantSnapshots.Create(ctx, &vergeos.TenantSnapshotCreateRequest{
    Tenant:      tenantID,
    Name:        "pre-upgrade",
    Description: "before upgrade",
    Type:        vergeos.TenantSnapshotTypeFull,
})

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


## Tenant Recipes

List catalog tenant recipes and deploy a tenant from one.

```go
// List tenant recipes (filter downloaded with WithFilter("downloaded eq 1"))
recipes, err := client.TenantRecipes.List(ctx)

// List recipes in a catalog
recipes, err := client.TenantRecipes.ListByCatalog(ctx, catalogID)

// Get a recipe by key or exact name
recipe, err := client.TenantRecipes.Get(ctx, recipeID)
recipe, err = client.TenantRecipes.GetByName(ctx, "30-Day Trial (POC)")

// Read the deploy form questions (answer keys and types)
questions, err := client.TenantRecipes.Questions(ctx, recipe.Key)

// Deploy creates a tenant. Answers are checked against the recipe's questions
// before the request is sent (same rules as VM recipe deploy).
instance, err := client.TenantRecipeInstances.Deploy(ctx, &vergeos.TenantRecipeDeployRequest{
    Recipe: recipe.Key,
    Name:   "customer-a",
    Answers: vergeos.RecipeAnswers{
        "YB_USER_NAME":     "admin",
        "YB_USER_PASSWORD": "secure-password",
    },
})

// List and get instances (Get includes stored answers)
instances, err := client.TenantRecipeInstances.ListByRecipe(ctx, recipe.Key)
instance, err = client.TenantRecipeInstances.Get(ctx, int(instance.Key))
instance, err = client.TenantRecipeInstances.GetByName(ctx, "customer-a")
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
err = client.TenantLayer2Networks.Enable(ctx, assignmentID)
err = client.TenantLayer2Networks.Disable(ctx, assignmentID)

// Convenience method: Assign a network to a tenant (creates if not exists)
assignment, err := client.TenantLayer2Networks.Assign(ctx, tenantID, networkID)

// Convenience method: Unassign a network from a tenant
err = client.TenantLayer2Networks.Unassign(ctx, tenantID, networkID)

// Delete an assignment
err = client.TenantLayer2Networks.Delete(ctx, assignmentID)
```

---

## Network proxy

Publish tenant UIs through one address on a parent network. Each tenant gets an FQDN. A network has one proxy.

```go
proxy, err := client.VNetProxies.GetOrCreate(ctx, &vergeos.VNetProxyCreateRequest{
    VNet: externalNetworkID,
})

mapping, err := client.VNetProxyTenants.Create(ctx, &vergeos.VNetProxyTenantCreateRequest{
    Proxy:  int(proxy.Key),
    Tenant: tenantID,
    FQDN:   "customer-a.example.com",
})

mappings, err := client.VNetProxyTenants.ListByProxy(ctx, int(proxy.Key))
```

---

## Client for a tenant

`Connect` builds a client aimed at the tenant's own UI address. The tenant must be running, and `ui_address` must already point at an address row. The new client keeps the parent's TLS and retry settings and uses the credentials passed in the options.

```go
tenantClient, err := client.Tenants.Connect(ctx, tenantID,
    vergeos.WithCredentials("admin", "tenant-password"),
)
if err != nil {
    log.Fatal(err)
}

vms, err := tenantClient.VMs.List(ctx)
```

---

## Shared objects

A parent system shares a VM with a tenant by snapshotting it and posting a `shared_objects` row. The snapshot does not expire. `Import` copies the VM into the tenant. `Delete` removes the share and leaves an already imported VM in place.

```go
shared, err := client.SharedObjects.Create(ctx, &vergeos.SharedObjectCreateRequest{
    Tenant:      tenantID,
    VM:          vmID,
    Name:        "Ubuntu Template",
    Description: "Pre-configured Ubuntu server",
})

shares, err := client.SharedObjects.ListByTenant(ctx, tenantID)
inbox, err := client.SharedObjects.ListInbox(ctx)

err = client.SharedObjects.Import(ctx, int(shared.Key))
err = client.SharedObjects.Refresh(ctx, int(shared.Key))
err = client.SharedObjects.Delete(ctx, int(shared.Key))
```
