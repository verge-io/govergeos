# goVergeOS

[![Go Reference](https://pkg.go.dev/badge/github.com/verge-io/govergeos.svg)](https://pkg.go.dev/github.com/verge-io/govergeos)
[![Go 1.21+](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Pre-release**: This library is under active development. APIs may change before v1.0.0. The freeze criteria and compatibility policy are in [docs/COMPATIBILITY.md](./docs/COMPATIBILITY.md). v1.0.0 is not tagged yet.

A Go client library for managing VergeOS infrastructure programmatically. goVergeOS provides complete API coverage for virtual machines, networking, storage, multi-tenancy, and disaster recovery operations.

Built for infrastructure administrators and Go developers who want to automate VergeOS management. This library serves as the foundation for the [Terraform Provider](https://github.com/verge-io/terraform-provider-vergeio), [Prometheus Exporter](https://github.com/verge-io/vergeos-exporter), and other VergeOS tooling.

## Key Features

- **Complete VM Management** - Create, configure, power, clone, and snapshot virtual machines with full drive and NIC control
- **Advanced Networking** - Virtual networks, firewall rules, DHCP, DNS views/zones/records, BGP/OSPF/EIGRP, WireGuard and IPSec VPNs
- **NAS & Storage** - Volume management, CIFS/NFS shares, async volume browsing, snapshot profiles
- **Multi-Tenancy** - Tenant provisioning, resource allocation, node management for MSPs and enterprises
- **Disaster Recovery** - Cloud snapshots, remote site synchronization, backup scheduling
- **Type-Safe Go API** - Interfaces for mocking, context support, thread-safe concurrent operations

## Requirements

- Go 1.21 or later
- VergeOS 26.0 or later (for full feature support)
- Standard library only - no external dependencies

## Installation

```bash
go get github.com/verge-io/govergeos
```

## Quick Start

### Connect to VergeOS

```go
import vergeos "github.com/verge-io/govergeos"

// Basic authentication
client, err := vergeos.NewClient(
    vergeos.WithBaseURL("https://your-vergeos-host"),
    vergeos.WithCredentials("username", "password"),
    vergeos.WithInsecureTLS(true), // For self-signed certificates
)

// API key authentication
client, err := vergeos.NewClient(
    vergeos.WithBaseURL("https://your-vergeos-host"),
    vergeos.WithAPIKey("your-api-key-token"),
)
```

`NewClient` accepts VergeOS 26 and every later major, and rejects anything older. `WithMinimumVersion` changes that floor. `WithSkipVersionCheck` still reads the server version (so per-feature gates keep working) and still checks credentials, but does not reject the major. Supported versions, deprecation, and the path to v1.0.0 are in [docs/COMPATIBILITY.md](./docs/COMPATIBILITY.md).

### Environment Configuration

Configure the client from environment variables using `WithEnvConfig()`:

```go
// Simple: all config from environment
client, err := vergeos.NewClient(vergeos.WithEnvConfig())

// With explicit override
client, err := vergeos.NewClient(
    vergeos.WithEnvConfig(),
    vergeos.WithTimeout(60*time.Second),  // Override just timeout
)
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VERGEOS_HOST` | Yes | - | Host or base URL. A value with no scheme is treated as `https://`. `http` and `https` are accepted |
| `VERGEOS_USERNAME` | No* | - | Username for basic auth |
| `VERGEOS_PASSWORD` | No* | - | Password for basic auth |
| `VERGEOS_API_KEY` | No* | - | API key for bearer auth. Used when set, even if username and password are also set |
| `VERGEOS_VERIFY_SSL` | No | `true` | Verify TLS certificates (`true`/`false`). `false` and `0` skip verification |
| `VERGEOS_INSECURE` | No | `false` | Skip TLS verification when `true` (also `yes`, `on`, or `1`). Same as `VERGEOS_VERIFY_SSL=false`. An error if the two disagree |
| `VERGEOS_TIMEOUT` | No | `30` | Request timeout in seconds |

*One of (USERNAME+PASSWORD) or API_KEY is required. When both are set, the API key is used.

```bash
export VERGEOS_HOST=vergeos.example.com   # https:// is assumed
export VERGEOS_USERNAME=admin
export VERGEOS_PASSWORD=secret
export VERGEOS_INSECURE=true              # same as VERGEOS_VERIFY_SSL=false
```

### Virtual Machine Operations

```go
ctx := context.Background()

// List all VMs
vms, err := client.VMs.List(ctx)

// Look up a VM by name. Snapshots are excluded.
vm, err := client.VMs.GetByName(ctx, "web-server")

// Create a VM
vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
    Name: "web-server", CPUCores: 4, RAM: 8192, Cluster: &clusterID,
})

// Power operations
// PowerOn waits until the VM is running.
err = client.VMs.PowerOn(ctx, vmID)
// PowerOff asks the guest to shut down and waits until the VM stops.
err = client.VMs.PowerOff(ctx, vmID)
// Kill is an immediate power-off. It waits until the VM stops.
err = client.VMs.Kill(ctx, vmID)
// Wait longer than the default 150s, then kill if the guest is still up.
err = client.VMs.PowerOffWithOptions(ctx, vmID, &vergeos.VMPowerOffOptions{
    Timeout:           10 * time.Minute,
    ForceAfterTimeout: true,
})

// Read a catalog and the questions a recipe will ask.
catalogs, err := client.Catalogs.List(ctx)
recipe, err := client.VMRecipes.GetByName(ctx, "Debian 12 (Bookworm)")
questions, err := client.VMRecipes.Questions(ctx, recipe.Key)

// Answers are checked against the recipe's questions before either call.
// Bool values the SDK does not recognize are refused. Disk sizes are bytes:
// 20 GiB is the number below, and 50 would be refused as fifty bytes.
answers := vergeos.RecipeAnswers{
    "HOSTNAME":           "web-01",
    "YB_DRIVE_OS_SIZE":   int64(20 * 1024 * 1024 * 1024),
    "SELECT_CREATE_UEFI": true,
}

// Preview walks the deploy and creates nothing. A successful HTTP status
// from the platform is an error here: that means a VM was created, and
// Preview will not return it.
preview, err := client.VMRecipeInstances.Preview(ctx, &vergeos.VMRecipeDeployRequest{
    Recipe: recipe.Key, Name: "web-01", Answers: answers,
})

// Deploy is the call that creates the VM.
instance, err := client.VMRecipeInstances.Deploy(ctx, &vergeos.VMRecipeDeployRequest{
    Recipe: recipe.Key, Name: "web-01", Answers: answers,
})

// Take a snapshot. Retention is a lifetime in seconds, sent as expires.
// Quiesce requires a running guest agent; leave it false to snapshot
// a VM that does not have one.
snapshot, err := client.VMs.Snapshot(ctx, vmID, &vergeos.VMSnapshotOptions{
    Name:      "pre-upgrade",
    Retention: 86400, // 24 hours
})
```

### Network Management

```go
// Look up a network by name
network, err := client.Networks.GetByName(ctx, "External")

// Create a network with DHCP
network, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
    Name: "internal", Network: "10.0.0.0/24", DHCPEnabled: ptr(true),
})

// Add a firewall rule
rule, err := client.VNetRules.Create(ctx, &vergeos.VNetRuleCreateRequest{
    VNet: networkID, Name: "allow-ssh", Protocol: ptr("tcp"),
    Direction: ptr("incoming"), DestinationPorts: ptr("22"), Action: ptr("accept"),
})

// Apply rules to the network
err = client.Networks.ApplyRules(ctx, networkID)

// Dynamic routing takes effect when the network restarts.
cfg, _, err := client.VNetBGP.GetOrCreate(ctx, networkID)
router, status, err := client.VNetBGPRouters.Create(ctx, &vergeos.VNetBGPRouterCreateRequest{
    BGP: int(cfg.Key), ASN: 65000,
}, vergeos.WithRestartNetwork())
fmt.Println(status.Pending, status.Restarted, router.ASN)
```

### Bulk Operations with Goroutines

```go
// Concurrent VM queries
var wg sync.WaitGroup
for _, id := range vmIDs {
    wg.Add(1)
    go func(vmID int) {
        defer wg.Done()
        vm, _ := client.VMs.Get(ctx, vmID)
        fmt.Printf("VM: %s, Running: %v\n", vm.Name, vm.PowerState)
    }(id)
}
wg.Wait()
```

### Multiple Clients

```go
// Manage multiple environments
prodClient, _ := vergeos.NewClient(
    vergeos.WithBaseURL("https://prod.example.com"),
    vergeos.WithAPIKey("prod-api-key"),
)
devClient, _ := vergeos.NewClient(
    vergeos.WithBaseURL("https://dev.example.com"),
    vergeos.WithAPIKey("dev-api-key"),
)
```

## Service Reference

### Virtual Machines

| Service | Description |
|---------|-------------|
| `VMs` | VM CRUD, power control, clone, snapshot, migrate, console |
| `VMSnapshots` | VM snapshot CRUD, restore, expiration management |
| `VMDrives` | VM disk management (attach, resize, detach) |
| `VMNICs` | VM network interface management |
| `VMDevices` | VM device management (USB, TPM, vGPU) |
| `Catalogs` | Recipe catalogs (read) |
| `VMRecipes` | VM recipes and their questions (read) |
| `VMRecipeInstances` | Deploy a VM from a recipe, or preview that deploy |
| `VMImports` | Import an OVA or disk image from a file or URL, with status, logs, and delete |
| `VMImportLogs` | Log lines for VM imports |
| `VMExports` | Export a VM to a NAS volume |

### Networking

| Service | Description |
|---------|-------------|
| `Networks` | Virtual network CRUD, power control, diagnostics, and statistics |
| `VNetRules` | Firewall rule management |
| `VNetRuleAliases` | IP/port alias groups for rules |
| `VNetAddresses` | IP address management (static, DHCP, aliases) |
| `VNetProxies` | Proxy that publishes tenant UIs through a parent network (`vnet_proxy`) |
| `VNetProxyTenants` | Tenant FQDN mappings for a network proxy (`vnet_proxy_tenants`) |
| `VNetDNSViews` | DNS view configuration |
| `VNetDNSZones` | DNS zone management |
| `VNetDNSRecords` | DNS record management (A, AAAA, CNAME, MX, TXT) |
| `VNetHosts` | DHCP reservations and host overrides |

### Dynamic Routing

BGP, OSPF, and EIGRP rows belong to one `vnet_bgp` record per network (`VNetBGP.GetOrCreate`). Create, update, and delete return `RoutingRestartStatus`. `Pending` is the network's `need_restart` flag. Pass `WithRestartNetwork` to restart the network in that call (`Networks.Reset`). Routing changes take effect on the restart.

| Service | Description |
|---------|-------------|
| `VNetBGP` | Per-network routing record (`vnet_bgp`), including get-or-create |
| `VNetBGPRouters` | BGP routers (`vnet_bgp_routers`) |
| `VNetBGPRouterCommands` | BGP router commands (`vnet_bgp_router_commands`) |
| `VNetBGPInterfaces` | BGP interfaces (`vnet_bgp_interfaces`) |
| `VNetBGPInterfaceCommands` | BGP interface commands (`vnet_bgp_interface_commands`) |
| `VNetBGPRouteMaps` | BGP route map entries (`vnet_bgp_routemaps`) |
| `VNetBGPRouteMapCommands` | BGP route map commands (`vnet_bgp_routemap_commands`) |
| `VNetBGPIPCommands` | BGP IP commands such as prefix-lists (`vnet_bgp_ip`) |
| `VNetOSPFCommands` | OSPF commands (`vnet_ospf_commands`) |
| `VNetEIGRPRouters` | EIGRP routers (`vnet_eigrp_routers`) |
| `VNetEIGRPRouterCommands` | EIGRP router commands (`vnet_eigrp_router_commands`) |

### VPN

| Service | Description |
|---------|-------------|
| `VNetWireGuards` | WireGuard interface management |
| `VNetWireGuardPeers` | WireGuard peer configuration |
| `VNetWireGuardPeerStatus` | WireGuard peer connection status (read-only) |
| `VNetIPSecs` | IPSec VPN configuration |
| `VNetIPSecPhase1s` | IPSec Phase 1 (IKE SA) settings |
| `VNetIPSecPhase2s` | IPSec Phase 2 (IPsec SA) settings |
| `VNetIPSecConnections` | Active IPSec connections (read-only) |

### NAS & Storage

| Service | Description |
|---------|-------------|
| `NASServices` | NAS service VM management and configuration |
| `NASServiceUsers` | NAS service user accounts (uses SHA1 string IDs) |
| `Volumes` | NAS volume management (uses SHA1 string IDs) |
| `VolumeSnapshots` | NAS volume snapshot management |
| `VolumeSyncs` | Volume replication/sync jobs (uses SHA1 string IDs) |
| `VolumeCIFSShares` | CIFS/SMB share management |
| `VolumeNFSShares` | NFS share management |
| `VolumeBrowser` | Async volume file browsing |

### Tenants

| Service | Description |
|---------|-------------|
| `Tenants` | Tenant CRUD, power operations, cloning, isolation, and a client for a running tenant's UI |
| `TenantNodes` | Tenant virtual node management |
| `TenantStorage` | Tenant storage allocation |
| `TenantSnapshots` | Tenant snapshot management |
| `TenantLayer2Networks` | Layer 2 network assignments to tenants |
| `TenantNetworkBlocks` | CIDR blocks assigned to a tenant (`vnet_cidrs`) |
| `TenantExternalIPs` | Virtual IPs given to a tenant (`vnet_addresses`) |

### Users & Groups

| Service | Description |
|---------|-------------|
| `Users` | User account management |
| `Groups` | Group management |
| `Members` | Group membership management |
| `UserAPIKeys` | API key management |
| `AuthSources` | External identity providers (SSO). Update merges settings |
| `OIDCApplications` | OIDC applications where VergeOS is the identity provider |
| `Permissions` | Resource-level access control (Grant/Revoke) |

### System Administration

| Service | Description |
|---------|-------------|
| `Clusters` | Cluster CRUD operations, status monitoring, and configuration |
| `Nodes` | Node information |
| `VGPUProfiles` | NVIDIA vGPU profile catalog (`nvidia_vgpu_profiles`, read-only) |
| `NodeGPUs` | Node GPU configuration (`node_gpus`), including passthrough and vGPU mode |
| `NodeGPUStats` | Current GPU stats and short/long history (read-only) |
| `NodeGPUInstances` | GPU instances assigned to VMs (read-only) |
| `NodeVGPUDevices` | Detected NVIDIA vGPU devices (read-only) |
| `NodeHostGPUDevices` | Detected GPUs available for passthrough (read-only) |
| `NodeVGPUProfiles` | vGPU profiles available on a physical GPU (read-only) |
| `NodeMemory` | DIMM inventory and health (`node_memory`, read-only) |
| `NodeLLDPNeighbors` | LLDP neighbors on node NICs (read-only) |
| `UpdateSettings` | Update settings, and check, download, install, and rolling apply |
| `UpdateBranches` | Update branches (read-only) |
| `UpdateSourcePackages` | Packages offered by an update source (read-only) |
| `Settings` | System settings (read-only) |
| `Schema` | API schema introspection |
| `System` | System version and info |
| `Logs` | System logs (audit, errors, warnings) |

### Monitoring & Tasks

| Service | Description |
|---------|-------------|
| `Alarms` | Alarm management (snooze, resolve, delete) |
| `AlarmTypes` | Alarm type reference data (read-only, string keys) |
| `Tasks` | Task monitoring, execution, and scheduling |
| `TaskSchedules` | Reusable schedules (`task_schedules`), including upcoming run times |
| `TaskScheduleTriggers` | Links a task to a schedule (`task_schedule_triggers`) |
| `TaskEvents` | Event triggers (`task_events`), including a manual trigger |
| `TaskScripts` | GCS scripts (`task_scripts`) and run |

### VSAN & Storage Monitoring

| Service | Description |
|---------|-------------|
| `StorageTiers` | System-wide storage tier capacity, usage, deduplication (read-only) |
| `ClusterTiers` | Cluster-specific tier status, redundancy, encryption (read-only) |
| `MachineDrivePhys` | Physical drive metrics: temperature, wear, SMART, VSAN status (read-only) |
| `ClusterStatsHistory` | Historical cluster stats: RAM, CPU, nodes, machines (read-only) |

### Backup & DR

| Service | Description |
|---------|-------------|
| `SnapshotProfiles` | Snapshot schedule profiles |
| `SnapshotProfilePeriods` | Snapshot schedule periods |
| `CloudSnapshots` | Cloud-level snapshot management |
| `CloudSnapshotVMs` | VMs in cloud snapshots (read-only) |
| `CloudSnapshotTenants` | Tenants in cloud snapshots (read-only) |
| `Sites` | Remote site connections |
| `SiteSyncsIncoming` | Incoming sync configurations |
| `SiteSyncsOutgoing` | Outgoing sync configurations |
| `SiteSyncProfilePeriods` | Site sync schedule periods |

### Automation & Files

| Service | Description |
|---------|-------------|
| `CloudInitFiles` | Cloud-init file management |
| `Files` | File CRUD, upload, and download (ISOs, images, etc.) |
| `WebhookURLs` | Webhook endpoint configuration |
| `Webhooks` | Webhook delivery log (read-only) |
| `Certificates` | SSL/TLS certificate management |

### Tags & Organization

| Service | Description |
|---------|-------------|
| `Tags` | Tag CRUD operations (v26+) |
| `TagCategories` | Tag category CRUD (define taggable resource types) |
| `TagMembers` | Tag assignment management |
| `ResourceGroups` | Resource group listing (read-only) |

## Examples

| Example | Description |
|---------|-------------|
| [basic](./examples/basic/) | Client setup, list resources, system info |
| [apikey-auth](./examples/apikey-auth/) | API key authentication |
| [vm-lifecycle](./examples/vm-lifecycle/) | VM create, configure, power, delete |
| [vm-recipes](./examples/vm-recipes/) | List catalogs, VM recipes, and recipe questions |
| [vm-import-export](./examples/vm-import-export/) | List VM imports, their logs, and VM exports |
| [vm-snapshots](./examples/vm-snapshots/) | VM snapshots, tags, and migration |
| [network-management](./examples/network-management/) | Create and manage virtual networks |
| [tenants](./examples/tenants/) | Multi-tenant management for MSPs |
| [volumes](./examples/volumes/) | NAS volume management and shares |
| [nas-services](./examples/nas-services/) | NAS services, users, syncs, and snapshots |
| [firewall-rules](./examples/firewall-rules/) | Network firewall rules and aliases |
| [vpn](./examples/vpn/) | WireGuard and IPSec VPN |
| [certificates](./examples/certificates/) | SSL/TLS certificate management |
| [tags](./examples/tags/) | Tag management and assignments |
| [users](./examples/users/) | User, group, and membership management |
| [permissions](./examples/permissions/) | Resource-level access control |
| [cloudinit](./examples/cloudinit/) | Cloud-init file management |
| [files](./examples/files/) | List available files (ISOs, images) |
| [snapshot-profiles](./examples/snapshot-profiles/) | Snapshot scheduling |
| [monitoring](./examples/monitoring/) | Alarms and tasks |
| [logs](./examples/logs/) | System logs and audit trails |
| [networking](./examples/networking/) | DNS, IP addresses, host overrides |
| [dr-sites](./examples/dr-sites/) | Remote sites and sync configuration |
| [cloud-snapshots](./examples/cloud-snapshots/) | Cloud snapshot management |
| [webhooks](./examples/webhooks/) | Webhook configuration |
| [vsan-monitoring](./examples/vsan-monitoring/) | VSAN storage metrics for Prometheus exporters |

## Configuration

### Client Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithEnvConfig()` | Configure from environment variables | - |
| `WithBaseURL(url)` | VergeOS API base URL | Required |
| `WithCredentials(user, pass)` | Username and password authentication. Ignored when an API key is also set | - |
| `WithAPIKey(token)` | API key authentication. Takes precedence over username and password | - |
| `WithInsecureTLS(bool)` | Skip TLS certificate verification | `false` |
| `WithTimeout(duration)` | HTTP request timeout, including retries | `30s` |
| `WithPowerWait(timeout, interval)` | Default VM power-wait budget and poll interval (`PowerOn`, `PowerOff`, `Kill`) | `150s`, every `5s` |
| `WithHTTPClient(client)` | Base `*http.Client`. Timeout, TLS, and rate limit are applied to a copy; the value you pass is not modified | Default client |
| `WithRetry(policy)` | Retry GET, PUT, and DELETE when the connection fails before a response, and on HTTP 429, 502, and 503 | 3 attempts, 100ms backoff doubling to 2s with jitter |
| `WithRateLimit(interval)` | Minimum time between request starts, including retries | off |

`WithHTTPClient` supplies the base client. `WithTimeout`, `WithInsecureTLS`, `WithEnvConfig`, and `WithRateLimit` are applied to a copy after every option has run, so putting `WithHTTPClient` before or after them does not drop those settings. A later `WithTimeout` or `WithInsecureTLS` still replaces an earlier one. Skipping certificate verification clones an `*http.Transport`. Any other transport type cannot take that change, and `NewClient` returns an error.

`WithRetry(RetryPolicy{MaxAttempts: 1})` turns retries off. A zero `MaxAttempts` keeps the default of 3. POST is not retried: creates and actions are not idempotent, and a POST that fails before a response is returned to the caller. HTTP 401 is never retried, because another attempt with the same rejected password counts toward account lockout.

`WithRateLimit` is off unless you set it. VergeOS drops connections when a session exceeds its webserver API rate limit. Pass an interval such as `50*time.Millisecond` when one client issues bursts faster than that limit.

`WithPowerWait` changes how long `PowerOn`, `PowerOff`, and `Kill` poll for a VM power state. The default is 150 seconds at a 5-second interval. `PowerOffWithOptions` sets the timeout and poll interval for one shutdown. `ForceAfterTimeout` sends `kill` when the guest is still running at that timeout. A context deadline ends the wait sooner, and a longer timeout extends it.

### Authentication

The library supports HTTP Basic Authentication and API key authentication:

- **Basic Auth**: User must have list and read permissions; MFA must be disabled
- **API Keys**: Created via `UserAPIKeys` service; token shown only on creation. When an API key and username/password are both configured, the API key is used
- **Auth sources**: `AuthSources.Update` reads the stored settings and merges the change before sending. The API would otherwise replace the whole settings object and drop keys that were not sent, including `client_secret`. That secret is write-only and is omitted from returned auth sources
- **OIDC applications**: `OIDCApplications.Create` returns the generated client secret as a `WriteOnlySecret`. Printing it shows `[redacted]`. `Value` reads it. Later reads do not include the secret

## Query Options

### Field Selection

```go
nodes, err := client.Nodes.List(ctx,
    vergeos.WithFields("most"),      // Common fields (default)
    vergeos.WithFields("dashboard"), // Dashboard-specific fields
    vergeos.WithFields("all"),       // All available fields
)
```

### Filtering

```go
nodes, err := client.Nodes.List(ctx, vergeos.WithFilter("physical eq true"))
vms, err := client.VMs.List(ctx, vergeos.WithFilter("name eq 'my-vm'"))
```

### Sorting

```go
vms, err := client.VMs.List(ctx, vergeos.WithSort("name"))
vms, err := client.VMs.List(ctx, vergeos.WithSort("-created"))  // Descending
```

### Pagination

```go
vms, err := client.VMs.List(ctx, vergeos.WithLimit(50), vergeos.WithOffset(0))
```

## Error Handling

The library returns typed errors for common scenarios:

```go
vm, err := client.VMs.Get(ctx, vmID)
if err != nil {
    if vergeos.IsNotFoundError(err) {
        // Resource doesn't exist
    }
    if vergeos.IsAuthError(err) {
        // Authentication failed (HTTP 401)
    }
    if vergeos.IsPermissionError(err) {
        // Authenticated, but not allowed (HTTP 403)
    }
    if vergeos.IsConflictError(err) {
        // Conflicts with existing state, such as a name already taken (HTTP 409)
    }
    if vergeos.IsAmbiguousNameError(err) {
        // GetByName matched more than one object. Use List and choose a key.
    }
    if vergeos.IsValidationError(err) {
        // Invalid request parameters
    }
    log.Fatal(err)
}
```

## Thread Safety

The client is safe for concurrent use. Share a single client instance across multiple goroutines.

## Development

```bash
# Build
go build ./...

# Run tests
go test ./...
go test -v ./...                    # Verbose
go test -run TestName ./...         # Single test

# Code quality
go fmt ./...
go vet ./...
```

## Integration Tests

Integration tests run against a live VergeOS instance. Tests are organized by category for selective execution.

### Setup

```bash
export VERGEOS_HOST=https://your-vergeos-host
export VERGEOS_USERNAME=admin
export VERGEOS_PASSWORD=your-password
```

### Run All Tests

```bash
go test -tags=integration -v ./test/integration/
```

### Run Tests by Category

```bash
# Virtual machines and snapshots
go test -tags=integration -v ./test/integration/ -run "TestVMSnapshots"

# Networking (addresses, DNS, hosts)
go test -tags=integration -v ./test/integration/ -run "TestVNet"

# NAS and storage
go test -tags=integration -v ./test/integration/ -run "TestNAS|TestVolume"

# Tenants
go test -tags=integration -v ./test/integration/ -run "TestTenants"

# Monitoring (alarms, tasks)
go test -tags=integration -v ./test/integration/ -run "TestAlarms|TestTasks"

# VSAN storage metrics
go test -tags=integration -v ./test/integration/ -run "TestVSAN|TestStorage|TestClusterTiers"

# DR and backup
go test -tags=integration -v ./test/integration/ -run "TestSites|TestCloudSnapshots"

# VPN (WireGuard, IPSec)
go test -tags=integration -v ./test/integration/ -run "TestWireGuard|TestIPSec"

# Files and certificates
go test -tags=integration -v ./test/integration/ -run "TestFiles|TestCertificates"

# Tags and permissions
go test -tags=integration -v ./test/integration/ -run "TestTags|TestPermissions"

# CRUD lifecycle tests only
go test -tags=integration -v ./test/integration/ -run "CRUD"
```

### Test Files

| File | Services Tested |
|------|-----------------|
| `api_keys_test.go` | User API Keys |
| `certificates_test.go` | Certificates |
| `clusters_test.go` | Clusters, Network Diagnostics |
| `dr_test.go` | Sites, Site Syncs, Cloud Snapshots |
| `files_test.go` | Files (upload/download) |
| `hardware_inventory_test.go` | GPUs, vGPU profiles, node memory, LLDP neighbors |
| `logs_test.go` | System Logs |
| `monitoring_test.go` | Alarms, Tasks |
| `nas_test.go` | NAS Services, Users, Syncs, Snapshots, Shares |
| `networking_test.go` | VNet Addresses, DNS, Hosts |
| `permissions_test.go` | Permissions |
| `rules_test.go` | VNet Rules, Aliases |
| `snapshot_profiles_test.go` | Snapshot Profiles |
| `tags_test.go` | Tags, Tag Categories |
| `tenants_test.go` | Tenants, Nodes, Storage, Snapshots, Layer2 |
| `vm_imports_test.go` | VM Imports, Import Logs, VM Exports |
| `vm_recipes_test.go` | Catalogs, VM Recipes, Recipe Questions |
| `vm_snapshots_test.go` | VM Snapshots |
| `volumes_test.go` | Volumes |
| `vpn_test.go` | WireGuard, IPSec |
| `vsan_test.go` | Storage Tiers, Cluster Tiers, Drive Metrics |
| `webhooks_test.go` | Webhooks |

## API Reference

For detailed API examples for all services, see [docs/REFERENCE.md](./docs/REFERENCE.md).

## Related Projects

- [terraform-provider-vergeio](https://github.com/verge-io/terraform-provider-vergeio) - Terraform provider for VergeOS
- [vergeos-exporter](https://github.com/verge-io/vergeos-exporter) - Prometheus exporter for VergeOS metrics
- [ansible-collection-vergeos](https://github.com/verge-io/ansible-collection-vergeos) - Ansible collection for VergeOS

## Resources

- [Compatibility and the 1.0 freeze](./docs/COMPATIBILITY.md) - Supported VergeOS versions, deprecation, and the v1.0.0 criteria
- [VergeOS Documentation](https://docs.verge.io/) - Official VergeOS documentation
- [VergeOS API Reference](https://docs.verge.io/knowledge-base/category/api/) - REST API documentation
- [GitHub Issues](https://github.com/verge-io/govergeos/issues) - Bug reports and feature requests
- [VergeOS Support](https://www.verge.io/support) - Commercial support options

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. By contributing you agree to the [Contributor License Agreement](CLA.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
