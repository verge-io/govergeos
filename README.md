# goVergeOS

> **Pre-release**: This library is under active development. APIs may change before v1.0.0.

A Go client library for managing VergeOS infrastructure programmatically. goVergeOS provides complete API coverage for virtual machines, networking, storage, multi-tenancy, and disaster recovery operations.

Built for infrastructure administrators and Go developers who want to automate VergeOS management. This library serves as the foundation for the [Terraform Provider](https://github.com/verge-io/terraform-provider-vergeio), [Prometheus Exporter](https://github.com/verge-io/vergeos-exporter), and other VergeOS tooling.

## Key Features

- **Complete VM Management** - Create, configure, power, clone, and snapshot virtual machines with full drive and NIC control
- **Advanced Networking** - Virtual networks, firewall rules, DHCP, DNS views/zones/records, WireGuard and IPSec VPNs
- **NAS & Storage** - Volume management, CIFS/NFS shares, async volume browsing, snapshot profiles
- **Multi-Tenancy** - Tenant provisioning, resource allocation, node management for MSPs and enterprises
- **Disaster Recovery** - Cloud snapshots, remote site synchronization, backup scheduling
- **Type-Safe Go API** - Interfaces for mocking, context support, thread-safe concurrent operations

## Requirements

- Go 1.21 or later
- VergeOS 26.0 or later (for full feature support)
- Standard library only - no external dependencies

## Installation

This repository is currently private. Configure Go to bypass the module proxy:

```bash
# Set GOPRIVATE (add to your shell profile for persistence)
export GOPRIVATE=github.com/verge-io/govergeos

# Install the library
go get github.com/verge-io/govergeos
```

You'll need GitHub access configured (SSH key or `gh auth login`).

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
| `VERGEOS_HOST` | Yes | - | Base URL (e.g., `https://vergeos.example.com`) |
| `VERGEOS_USERNAME` | No* | - | Username for basic auth |
| `VERGEOS_PASSWORD` | No* | - | Password for basic auth |
| `VERGEOS_API_KEY` | No* | - | API key for bearer auth |
| `VERGEOS_VERIFY_SSL` | No | `true` | Verify TLS certificates (`true`/`false`) |
| `VERGEOS_TIMEOUT` | No | `30` | Request timeout in seconds |

*One of (USERNAME+PASSWORD) or API_KEY is required.

```bash
export VERGEOS_HOST=https://vergeos.example.com
export VERGEOS_USERNAME=admin
export VERGEOS_PASSWORD=secret
export VERGEOS_VERIFY_SSL=false  # For self-signed certs
```

### Virtual Machine Operations

```go
ctx := context.Background()

// List all VMs
vms, err := client.VMs.List(ctx)

// Create a VM
vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
    Name: "web-server", CPUCores: 4, RAM: 8192, Cluster: clusterID,
})

// Power operations
err = client.VMs.PowerOn(ctx, vmID)
err = client.VMs.PowerOff(ctx, vmID)

// Create a snapshot
err = client.VMs.Snapshot(ctx, vmID, "pre-upgrade")
```

### Network Management

```go
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
        fmt.Printf("VM: %s, Status: %s\n", vm.Name, vm.Status)
    }(id)
}
wg.Wait()
```

### Multiple Clients

```go
// Manage multiple environments
prodClient, _ := vergeos.NewClient(vergeos.WithBaseURL("https://prod.example.com"), ...)
devClient, _ := vergeos.NewClient(vergeos.WithBaseURL("https://dev.example.com"), ...)
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

### Networking

| Service | Description |
|---------|-------------|
| `Networks` | Virtual network CRUD, power control, diagnostics, and statistics |
| `VNetRules` | Firewall rule management |
| `VNetRuleAliases` | IP/port alias groups for rules |
| `VNetAddresses` | IP address management (static, DHCP, aliases) |
| `VNetDNSViews` | DNS view configuration |
| `VNetDNSZones` | DNS zone management |
| `VNetDNSRecords` | DNS record management (A, AAAA, CNAME, MX, TXT) |
| `VNetHosts` | DHCP reservations and host overrides |

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
| `Tenants` | Tenant CRUD, power operations, cloning, isolation |
| `TenantNodes` | Tenant virtual node management |
| `TenantStorage` | Tenant storage allocation |
| `TenantSnapshots` | Tenant snapshot management |
| `TenantLayer2Networks` | Layer 2 network assignments to tenants |

### Users & Groups

| Service | Description |
|---------|-------------|
| `Users` | User account management |
| `Groups` | Group management |
| `Members` | Group membership management |
| `UserAPIKeys` | API key management |
| `Permissions` | Resource-level access control (Grant/Revoke) |

### System Administration

| Service | Description |
|---------|-------------|
| `Clusters` | Cluster CRUD operations, status monitoring, and configuration |
| `Nodes` | Node information |
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
| [vm-snapshots](./examples/vm-snapshots/) | VM snapshots, tags, and migration |
| [network-management](./examples/network-management/) | Create and manage virtual networks |
| [tenants](./examples/tenants/) | Multi-tenant management for MSPs |
| [volumes](./examples/volumes/) | NAS volume management and shares |
| [firewall-rules](./examples/firewall-rules/) | Network firewall rules and aliases |
| [vpn](./examples/vpn/) | WireGuard and IPSec VPN |
| [certificates](./examples/certificates/) | SSL/TLS certificate management |
| [tags](./examples/tags/) | Tag management and assignments |
| [users](./examples/users/) | User, group, and membership management |
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
| `WithCredentials(user, pass)` | Username and password authentication | - |
| `WithAPIKey(token)` | API key authentication | - |
| `WithInsecureTLS(bool)` | Skip TLS certificate verification | `false` |
| `WithTimeout(duration)` | HTTP request timeout | `30s` |
| `WithHTTPClient(client)` | Custom `*http.Client` | Default client |

### Authentication

The library supports HTTP Basic Authentication and API key authentication:

- **Basic Auth**: User must have list and read permissions; MFA must be disabled
- **API Keys**: Created via `UserAPIKeys` service; token shown only on creation

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
        // Authentication failure
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
| `logs_test.go` | System Logs |
| `monitoring_test.go` | Alarms, Tasks |
| `nas_test.go` | NAS Services, Users, Syncs, Snapshots, Shares |
| `networking_test.go` | VNet Addresses, DNS, Hosts |
| `permissions_test.go` | Permissions |
| `rules_test.go` | VNet Rules, Aliases |
| `snapshot_profiles_test.go` | Snapshot Profiles |
| `tags_test.go` | Tags, Tag Categories |
| `tenants_test.go` | Tenants, Nodes, Storage, Snapshots, Layer2 |
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

- [VergeOS Documentation](https://docs.verge.io/) - Official VergeOS documentation
- [VergeOS API Reference](https://docs.verge.io/knowledge-base/category/api/) - REST API documentation
- [GitHub Issues](https://github.com/verge-io/govergeos/issues) - Bug reports and feature requests
- [VergeOS Support](https://www.verge.io/support) - Commercial support options

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
