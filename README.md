# VergeOS Go SDK

A Go client library for interacting with the VergeOS API.

## Overview

The VergeOS Go SDK provides a convenient way to interact with VergeOS infrastructure from Go applications. It handles authentication, request building, and response parsing, allowing you to focus on building your application.

This SDK serves as the foundation for other VergeOS tooling including the [Terraform Provider](https://github.com/verge-io/terraform-provider-vergeio) and [Prometheus Exporter](https://github.com/verge-io/vergeos-exporter).

## Features

- Simple, idiomatic Go API
- Full CRUD operations for VergeOS resources
- HTTP Basic Authentication
- Configurable HTTP client with TLS options
- Support for API v4 endpoints
- Field selection, filtering, sorting, and pagination
- Strongly typed resource models
- Thread-safe for concurrent use

## Installation

```bash
go get github.com/verge-io/vergeos-go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    vergeos "github.com/verge-io/vergeos-go-sdk"
)

func main() {
    // Create a new client
    client, err := vergeos.NewClient(
        vergeos.WithBaseURL("https://your-vergeos-host"),
        vergeos.WithCredentials("username", "password"),
        vergeos.WithInsecureTLS(true), // For self-signed certificates
    )
    if err != nil {
        log.Fatal(err)
    }

    // List all nodes
    nodes, err := client.Nodes.List(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for _, node := range nodes {
        fmt.Printf("Node: %s (Cluster: %d)\n", node.Name, node.Cluster)
    }
}
```

## Configuration

### Client Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithBaseURL(url)` | VergeOS API base URL | Required |
| `WithCredentials(user, pass)` | API credentials | Required |
| `WithInsecureTLS(bool)` | Skip TLS certificate verification | `false` |
| `WithTimeout(duration)` | HTTP request timeout | `30s` |
| `WithHTTPClient(client)` | Custom `*http.Client` | Default client |

### Authentication

The SDK uses HTTP Basic Authentication. You can use either a "Normal" user or an "API" user account.

**Requirements:**
- User must have list and read permissions on the cloud
- MFA must be disabled for the user account

## API Resources

The SDK provides access to the following VergeOS resources:

### Virtual Machines

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
    Cluster:  clusterID,
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
```

### Networks

```go
// List all networks
networks, err := client.Networks.List(ctx)

// Create a network
dhcpEnabled := true
network, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
    Name:        "my-network",
    Network:     "10.0.0.0/24",
    DHCPEnabled: &dhcpEnabled,
})

// Update a network
dhcpStart := "10.0.0.100"
dhcpStop := "10.0.0.200"
network, err := client.Networks.Update(ctx, networkID, &vergeos.NetworkUpdateRequest{
    DHCPStart: &dhcpStart,
    DHCPStop:  &dhcpStop,
})
```

### Nodes

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

### Clusters

```go
// List all clusters
clusters, err := client.Clusters.List(ctx)

// Get cluster details
cluster, err := client.Clusters.Get(ctx, clusterID)

// Get cluster status
status, err := client.Clusters.GetStatus(ctx, clusterID)
```

### VM Drives

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
    SizeGB: ptr(100), // Increase size
})
```

### VM NICs

```go
// List NICs for a VM
nics, err := client.VMNICs.List(ctx, vmMachineID)

// Create a NIC
nic, err := client.VMNICs.Create(ctx, vmMachineID, &vergeos.VMNICCreateRequest{
    Name: "eth0",
    VNET: networkID,
})
```

### Users

```go
// List users
users, err := client.Users.List(ctx)

// Create a user
user, err := client.Users.Create(ctx, &vergeos.UserCreateRequest{
    Name:     "newuser",
    Password: "securepassword",
})
```

### Groups

```go
// List groups
groups, err := client.Groups.List(ctx)

// Manage group members
err = client.Members.Add(ctx, groupID, userID)
```

### Settings

```go
// Get system settings
settings, err := client.Settings.List(ctx)

// Get a specific setting by key
setting, err := client.Settings.GetByKey(ctx, "cloud_name")
fmt.Printf("Cloud: %s\n", setting.Value)

// Convenience method for cloud name
cloudName, err := client.Settings.GetCloudName(ctx)
```

### System

```go
// Get system version info (uses /version.json endpoint)
info, err := client.System.GetInfo(ctx)
fmt.Printf("API: %s, Version: %s\n", info.Name, info.Version)

// Get just the version string
version, err := client.System.GetVersion(ctx)
```

### Tags

Tags allow you to categorize and organize resources. Requires VergeOS v26+.

```go
// List all tags
tags, err := client.Tags.List(ctx)

// Get a tag by ID
tag, err := client.Tags.Get(ctx, tagID)

// Get a tag by name
tag, err := client.Tags.GetByName(ctx, "production")

// List tags in a specific category
tags, err := client.Tags.ListByCategory(ctx, categoryID)
```

### Tag Members

Tag members represent tag assignments to resources (VMs, networks, etc.).

```go
// List all tag assignments
members, err := client.TagMembers.List(ctx)

// List tags assigned to a specific VM
members, err := client.TagMembers.ListByMember(ctx, "vms/123")

// List all resources with a specific tag
members, err := client.TagMembers.ListByTag(ctx, tagID)

// Assign a tag to a VM
member, err := client.TagMembers.Assign(ctx, tagID, "vms/123")

// Remove a tag from a VM
err = client.TagMembers.Unassign(ctx, tagID, "vms/123")

// Or use Create/Delete directly
member, err := client.TagMembers.Create(ctx, &vergeos.TagMemberCreateRequest{
    Tag:    tagID,
    Member: "vms/123",  // format: "object_type/object_id"
})
err = client.TagMembers.Delete(ctx, memberID)
```

### Additional Resources

The SDK also supports:
- **Media Sources** - ISO images and media files
- **CloudInit Files** - Cloud-init configuration management
- **Resource Groups** - Logical grouping of resources
- **NICs** - Virtual machine network interfaces
- **Devices** - VM devices (USB, TPM, vGPU)

## Query Options

### Field Selection

Request specific fields to optimize response size:

```go
nodes, err := client.Nodes.List(ctx,
    vergeos.WithFields("most"),      // Common fields (default)
    vergeos.WithFields("dashboard"), // Dashboard-specific fields
    vergeos.WithFields("all"),       // All available fields
)
```

### Filtering

Filter results using VergeOS query syntax:

```go
// Filter by field value
nodes, err := client.Nodes.List(ctx,
    vergeos.WithFilter("physical eq true"),
)

// Filter by name
vms, err := client.VMs.List(ctx,
    vergeos.WithFilter("name eq 'my-vm'"),
)
```

### Sorting

Sort results by field:

```go
vms, err := client.VMs.List(ctx,
    vergeos.WithSort("name"),
    vergeos.WithSort("-created"),  // Descending order
)
```

### Pagination

Paginate through large result sets:

```go
// Get first 50 results
vms, err := client.VMs.List(ctx,
    vergeos.WithLimit(50),
    vergeos.WithOffset(0),
)

// Get next page
vms, err := client.VMs.List(ctx,
    vergeos.WithLimit(50),
    vergeos.WithOffset(50),
)
```

## Error Handling

The SDK returns typed errors for common scenarios:

```go
node, err := client.Nodes.Get(ctx, nodeID)
if err != nil {
    if vergeos.IsNotFoundError(err) {
        // Handle not found
    }
    if vergeos.IsAuthError(err) {
        // Handle authentication failure
    }
    // Handle other errors
    log.Fatal(err)
}
```

## Thread Safety

The SDK client is safe for concurrent use. You can share a single client instance across multiple goroutines.

## Examples

See the [examples](./examples) directory for complete working examples:

- [Basic Usage](./examples/basic) - Simple client setup, list resources, and get system info
- [VM Lifecycle](./examples/vm-lifecycle) - Create, configure, power, and delete VMs
- [Network Management](./examples/network-management) - Create and manage virtual networks
- [Tags](./examples/tags) - List tags and manage tag assignments on resources
- [Users](./examples/users) - User management, groups, and memberships
- [Cloud-Init](./examples/cloudinit) - Cloud-init file management for VM provisioning
- [Media Sources](./examples/media-sources) - List available ISOs and boot media

## API Endpoints Reference

The SDK wraps the VergeOS API v4 (`/api/v4/`). Key endpoints include:

| Resource | Endpoint | Operations |
|----------|----------|------------|
| VMs | `/api/v4/vms` | CRUD |
| VM Actions | `/api/v4/vm_actions` | Power on/off, hotplug |
| Networks | `/api/v4/vnets` | CRUD |
| Network Actions | `/api/v4/vnet_actions` | Power on/off |
| Network Addresses | `/api/v4/vnet_addresses` | IP assignment |
| Nodes | `/api/v4/nodes` | Read |
| Clusters | `/api/v4/clusters` | Read |
| Drives | `/api/v4/machine_drives` | CRUD |
| NICs | `/api/v4/machine_nics` | CRUD |
| Devices | `/api/v4/machine_devices` | CRUD |
| Users | `/api/v4/users` | CRUD |
| Groups | `/api/v4/groups` | Read |
| Members | `/api/v4/members` | CRUD |
| Settings | `/api/v4/settings` | Read |
| Media Sources | `/api/v4/files` | Read |
| CloudInit Files | `/api/v4/cloudinit_files` | CRUD |
| Resource Groups | `/api/v4/resource_groups` | Read |
| Tags | `/api/v4/tags` | Read (v26+) |
| Tag Members | `/api/v4/tag_members` | CRUD (v26+) |
| System Info | `/version.json` | Read (outside API v4) |

## Related Projects

- [terraform-provider-vergeio](https://github.com/verge-io/terraform-provider-vergeio) - Terraform provider for VergeOS infrastructure as code
- [vergeos-exporter](https://github.com/verge-io/vergeos-exporter) - Prometheus exporter for VergeOS metrics
- [ansible-collection-vergeos](https://github.com/verge-io/ansible-collection-vergeos) - Ansible collection for VergeOS automation

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
