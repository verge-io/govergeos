# goVergeOS

A Go client library for the VergeOS API (v4), serving as the foundation for the VergeOS Terraform Provider and Prometheus Exporter.

## Tech Stack

- **Language**: Go 1.21+
- **Module**: `github.com/verge-io/goVergeOS`
- **Dependencies**: Standard library only (no external deps)
- **Authentication**: HTTP Basic Auth
- **Testing**: `go test` with integration tests (build tag: `integration`)

## Project Structure

```
goVergeOS/
├── client.go             # Client struct, HTTP methods, service init
├── options.go            # ListOptions and functional options
├── errors.go             # Error types (APIError, NotFoundError, etc.)
├── types.go              # FlexInt custom type for ID handling
├── doc.go                # Package documentation
│
├── vm.go                 # VM service and operations
├── vm_nic.go             # VM NIC service
├── vm_drive.go           # VM Drive service
├── vm_device.go          # VM Device service
├── network.go            # Network service
├── user.go               # User service
├── member.go             # Member service
├── cloudinit.go          # CloudInit service
├── cluster.go            # Cluster service
├── node.go               # Node service
├── group.go              # Group service
├── file.go               # File service
├── resourcegroup.go      # ResourceGroup service
├── settings.go           # Settings service
├── system.go             # System service
├── schema.go             # Schema service
│
├── types_vm.go           # VM type definitions
├── types_vm_nic.go       # VM NIC types
├── types_vm_drive.go     # VM Drive types
├── types_vm_device.go    # VM Device types
├── types_network.go      # Network types
├── types_user.go         # User types
├── types_member.go       # Member types
├── types_cloudinit.go    # CloudInit types
├── types_readonly.go     # Read-only resource types
│
├── examples/
│   ├── basic/            # Basic usage
│   ├── network-management/
│   └── vm-lifecycle/
│
├── DECISIONS.md          # Architecture Decision Records
├── README.md             # User-facing documentation
└── .claude/
    ├── PRD.md            # Product requirements
    └── reference/        # Best practices docs and API refrence schemas
```

## Commands

```bash
# Build
go build ./...

# Run tests
go test ./...
go test -run TestFunctionName ./...    # Single test
go test -v ./...                        # Verbose

# Format and lint
go fmt ./...
go vet ./...

# Run examples (requires env vars)
VERGEOS_HOST=https://your-host \
VERGEOS_USERNAME=user \
VERGEOS_PASSWORD=pass \
VERGEOS_VERIFY_SSL=false \
go run ./examples/basic/
```

## Environment Variables

The SDK supports configuration via environment variables with explicit opt-in:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VERGEOS_HOST` | Yes | - | API base URL |
| `VERGEOS_USERNAME` | No* | - | Basic auth username |
| `VERGEOS_PASSWORD` | No* | - | Basic auth password |
| `VERGEOS_API_KEY` | No* | - | Bearer token auth |
| `VERGEOS_VERIFY_SSL` | No | `true` | TLS verification |
| `VERGEOS_TIMEOUT` | No | `30` | Request timeout (seconds) |

*One of (USERNAME+PASSWORD) or API_KEY required.

Use `vergeos.NewClient(vergeos.WithEnvConfig())` to opt-in to environment configuration.
Environment variables are never automatically picked up - this is intentional (see ADR-017).

## Reference Documentation

| Document | When to Read |
|----------|--------------|
| `.claude/PRD.md` | Understanding requirements, API coverage, feature scope |
| `.claude/reference/API_SCHEMA_GUIDE.md` | Adding/updating services, type mapping, field coverage |
| `.claude/reference/api-schmea/schema/v4/` | Raw API schema files for each resource |
| `DECISIONS.md` | Architecture decisions, design rationale |
| `README.md` | User-facing examples, quick start guide |
| `examples/` | Working code examples for common use cases |

## Architecture

### Flat Package Structure
All code lives in the root `vergeos` package. This provides simple imports (`vergeos.NewClient()`) and follows patterns of aws-sdk-go and google-cloud-go.

### Service-Oriented Design
The `Client` struct holds 16 service instances, each handling operations for one resource type:

```go
client.VMs.List(ctx)
client.VMs.Create(ctx, req)
client.Networks.Get(ctx, id)
client.Networks.Update(ctx, id, req)
```

**Services**: VMs, VMNICs, VMDrives, VMDevices, Networks, Users, Members, CloudInitFiles, Clusters, Nodes, Groups, Files, ResourceGroups, Settings, System, Schema

### Key Files
- `client.go` - Client struct, HTTP methods (`get`, `post`, `put`, `delete`), service initialization
- `options.go` - ListOptions and functional options (`WithFilter`, `WithFields`, `WithSort`)
- `errors.go` - Error types (`APIError`, `NotFoundError`, `AuthError`, `ValidationError`)
- `types.go` - FlexInt custom type for handling API ID inconsistencies

## Code Conventions

### Functional Options Pattern
Client configuration uses builder-style options:

```go
client, err := vergeos.NewClient(
    vergeos.WithBaseURL("https://host"),
    vergeos.WithCredentials("user", "pass"),
    vergeos.WithInsecureTLS(true),
)
```

### Pointer Types for Updates
Update request structs use pointers (`*string`, `*int`) to distinguish "don't update" (nil) from "set to zero value" (non-nil):

```go
type VMUpdateRequest struct {
    Name        *string `json:"name,omitempty"`
    Description *string `json:"description,omitempty"`
    RAM         *int    `json:"ram,omitempty"`
}
```

### FlexInt Type
The VergeOS API sometimes returns IDs as strings, sometimes as integers. The `FlexInt` type in `types.go` handles this transparently.

### API Method Patterns
- All methods take `context.Context` as first parameter
- List methods accept variadic `ListOption` arguments
- Create returns the created resource (reads back after POST)
- Update returns the updated resource (reads back after PUT)
- Delete returns only error
- Action methods (PowerOn, Clone, etc.) return only error

### Error Handling
Use typed error checking:

```go
if vergeos.IsNotFoundError(err) { ... }
if vergeos.IsAuthError(err) { ... }
if vergeos.IsValidationError(err) { ... }
```

### Naming Conventions
- `camelCase` for variables and functions
- `PascalCase` for exported types and methods
- Service files: `resourcename.go` (e.g., `vm.go`, `network.go`)
- Type files: `types_resourcename.go` (e.g., `types_vm.go`)

## Adding a New Service

### Step 0: Review the API Schema

Before writing code, check the schema for complete field coverage:

```bash
# Find the schema file
ls .claude/reference/api-schmea/schema/v4/ | grep resourcename

# Read the schema
cat .claude/reference/api-schmea/schema/v4/resourcename/table.json
```

See `.claude/reference/API_SCHEMA_GUIDE.md` for detailed type mapping and patterns.

### Step 1: Create Types

Create `types_newresource.go` with resource and request types based on schema fields:

```go
type NewResource struct {
    Key  FlexInt `json:"$key"`           // Row key (use FlexInt for IDs)
    Name string  `json:"name"`            // Required fields as values
    Status string `json:"status,omitempty"` // Optional fields with omitempty
}

type NewResourceCreateRequest struct {
    Name string `json:"name"`             // Required - value type
    Description string `json:"description,omitempty"` // Optional
}

type NewResourceUpdateRequest struct {
    Name *string `json:"name,omitempty"`  // Pointer for optional update
    Description *string `json:"description,omitempty"`
}

// Field list constant - include all fields from schema
const newResourceListFields = "$key,name,status,description,created,..."
```

### Step 2: Create Service

Create `newresource.go` with service struct and methods:

```go
type NewResourceService struct {
    client *Client
}

func (s *NewResourceService) List(ctx context.Context, opts ...ListOption) ([]NewResource, error) {
    // Implementation
}
```

### Step 3: Register Service

Add service to Client in `client.go`:

```go
// In Client struct
NewResources *NewResourceService

// In NewClient()
c.NewResources = &NewResourceService{client: c}
```

### Step 4: Verify Coverage

Compare your types against the schema to ensure complete coverage:
- All schema fields mapped to struct fields
- Correct Go types (see API_SCHEMA_GUIDE.md for mapping)
- JSON tags match schema field names exactly
- Field list constant includes all relevant fields

## Testing Strategy

Tests follow Go conventions with integration tests against a live VergeOS API.

### Test Organization

Integration tests are organized by service category in `test/integration/`:

```
test/integration/
├── helpers_test.go          # Shared utilities (setupTestClient, prettyPrint, ptr)
├── api_keys_test.go         # User API Keys
├── certificates_test.go     # SSL/TLS Certificates
├── clusters_test.go         # Clusters, Network Diagnostics
├── dr_test.go               # Sites, Site Syncs, Cloud Snapshots
├── files_test.go            # File upload/download
├── groups_test.go           # Groups
├── logs_test.go             # System Logs
├── monitoring_test.go       # Alarms, Alarm Types, Tasks
├── nas_test.go              # NAS Services, Users, Syncs, Snapshots, CIFS/NFS Shares
├── networking_test.go       # VNet Addresses, DNS, Hosts
├── permissions_test.go      # Permissions
├── rules_test.go            # VNet Rules, Rule Aliases
├── snapshot_profiles_test.go # Snapshot Profiles, Periods
├── tags_test.go             # Tags, Tag Categories
├── tenants_test.go          # Tenants, Tenant Nodes/Storage/Snapshots/Layer2
├── version_test.go          # Version check
├── vm_snapshots_test.go     # VM Snapshots, VM Migration
├── volumes_test.go          # Volumes
├── vpn_test.go              # WireGuard, IPSec
├── vsan_test.go             # Storage Tiers, Cluster Tiers, Drive Phys, Stats
└── webhooks_test.go         # Webhook URLs, Webhooks
```

### Running Integration Tests

```bash
# Set credentials
export VERGEOS_HOST=https://your-vergeos-host
export VERGEOS_USERNAME=admin
export VERGEOS_PASSWORD=your-password

# Run all integration tests
go test -tags=integration -v ./test/integration/

# Run by category
go test -tags=integration -v ./test/integration/ -run "TestVMSnapshots"
go test -tags=integration -v ./test/integration/ -run "TestNAS"
go test -tags=integration -v ./test/integration/ -run "TestClusters"
go test -tags=integration -v ./test/integration/ -run "TestVSAN"
go test -tags=integration -v ./test/integration/ -run "TestDR"
go test -tags=integration -v ./test/integration/ -run "TestVPN"

# Run CRUD lifecycle tests only
go test -tags=integration -v ./test/integration/ -run "CRUD"

# Run multiple categories
go test -tags=integration -v ./test/integration/ -run "TestNAS|TestVolume"
```

### Test Conventions

- All tests require `//go:build integration` build tag
- Tests use `setupTestClient(t)` from helpers_test.go
- Context timeout: 2 minutes per test function
- CRUD tests clean up created resources in defer blocks
- Destructive tests (Create/Delete) often require opt-in via environment variables

## Reserved Networks
IMPORTANT: Never use "Core" or "DMZ" networks for workloads, services, VMs, or NAS. These networks are reserved for the VergeOS operating system. Always create a new network (e.g., "Internal") for test workloads and examples.
