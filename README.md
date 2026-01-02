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
        fmt.Printf("Node: %s (Cluster: %s)\n", node.Name, node.Cluster)
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
vm, err := client.VMs.Update(ctx, vmID, &vergeos.VMUpdateRequest{
    Name: "renamed-vm",
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
network, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
    Name:        "my-network",
    Network:     "10.0.0.0/24",
    DHCPEnabled: true,
})

// Update a network
network, err := client.Networks.Update(ctx, networkID, &vergeos.NetworkUpdateRequest{
    DHCPStart: "10.0.0.100",
    DHCPStop:  "10.0.0.200",
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

// Get cluster statistics
stats, err := client.Clusters.GetStats(ctx, clusterID)
```

### Storage Tiers

```go
// List storage tiers
tiers, err := client.StorageTiers.List(ctx)

// Get cluster tier details
clusterTiers, err := client.ClusterTiers.List(ctx)
```

### Drives

```go
// List machine drives
drives, err := client.Drives.List(ctx)

// List with filter
drives, err := client.Drives.List(ctx,
    vergeos.WithFilter("status eq 'online'"),
)
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

// Get a specific setting
cloudName, err := client.Settings.Get(ctx, "cloud_name")
```

### System

```go
// Get system version and update info
info, err := client.System.GetUpdateDashboard(ctx)
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

- [Basic Usage](./examples/basic) - Simple client setup and API calls
- [List Resources](./examples/list-resources) - Listing nodes, clusters, and storage
- [Metrics Collection](./examples/metrics) - Collecting system metrics

## API Endpoints Reference

The SDK wraps the VergeOS API v4 (`/api/v4/`). Key endpoints include:

| Resource | Endpoint | Operations |
|----------|----------|------------|
| VMs | `/api/v4/vms` | CRUD, power actions |
| VM Actions | `/api/v4/vm_actions` | Power state changes |
| Networks | `/api/v4/vnets` | CRUD, network actions |
| Nodes | `/api/v4/nodes` | Read |
| Clusters | `/api/v4/clusters` | Read |
| Storage Tiers | `/api/v4/storage_tiers` | Read |
| Cluster Tiers | `/api/v4/cluster_tiers` | Read |
| Drives | `/api/v4/machine_drives` | Read |
| NICs | `/api/v4/machine_nics` | CRUD |
| Users | `/api/v4/users` | CRUD |
| Groups | `/api/v4/groups` | Read |
| Members | `/api/v4/members` | CRUD |
| Settings | `/api/v4/settings` | Read |
| Media Sources | `/api/v4/files` | Read |
| CloudInit Files | `/api/v4/cloudinit_files` | CRUD |
| Resource Groups | `/api/v4/resource_groups` | Read |

## Related Projects

- [terraform-provider-vergeio](https://github.com/verge-io/terraform-provider-vergeio) - Terraform provider for VergeOS infrastructure as code
- [vergeos-exporter](https://github.com/verge-io/vergeos-exporter) - Prometheus exporter for VergeOS metrics
- [ansible-collection-vergeos](https://github.com/verge-io/ansible-collection-vergeos) - Ansible collection for VergeOS automation

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
