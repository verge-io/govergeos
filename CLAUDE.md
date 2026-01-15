# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VergeOS Go SDK - A Go client library for the VergeOS API (v4). Used as the foundation for the VergeOS Terraform Provider and Prometheus Exporter.

**Language:** Go 1.21+
**Module:** `github.com/verge-io/vergeos-go-sdk`
**Dependencies:** Standard library only (no external deps)

## Build Commands

```bash
# Build the package
go build ./...

# Run tests (when they exist)
go test ./...

# Run a single test
go test -run TestFunctionName ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Run examples (requires environment variables)
VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/basic/
```

## Architecture

### Flat Package Structure
All code lives in the root `vergeos` package - no nested packages. This provides simple imports (`vergeos.NewClient()`) and follows patterns of aws-sdk-go and google-cloud-go.

### Service-Oriented Design
The `Client` struct holds 16 service instances. Each service handles operations for one resource type:
- `client.VMs.List(ctx)`, `client.VMs.Create(ctx, req)`
- `client.Networks.Get(ctx, id)`, `client.Networks.Update(ctx, id, req)`

Services are defined in their own files (e.g., `vm.go`, `network.go`) with corresponding type definitions in `types_*.go` files.

### Key Files
- `client.go` - Client struct, HTTP methods (`get`, `post`, `put`, `delete`), service initialization
- `options.go` - ListOptions and functional options (`WithFilter`, `WithFields`, `WithSort`)
- `errors.go` - Error types (`APIError`, `NotFoundError`, `AuthError`, `ValidationError`)
- `types.go` - FlexInt custom type for handling API ID inconsistencies

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
Update request structs use pointers (`*string`, `*int`) to distinguish "don't update" (nil) from "set to zero value" (non-nil). See `types_vm.go:VMUpdateRequest` for example.

### FlexInt Type
The VergeOS API sometimes returns IDs as strings, sometimes as integers. The `FlexInt` type in `types.go` handles this inconsistency transparently.

## Adding a New Service

1. Create `newresource.go` with service struct and methods:
```go
type NewResourceService struct {
    client *Client
}

func (s *NewResourceService) List(ctx context.Context, opts ...ListOption) ([]NewResource, error) {
    // Implementation
}
```

2. Create `types_newresource.go` with resource and request types:
```go
type NewResource struct {
    Key  FlexInt `json:"$key"`
    Name string  `json:"name"`
}

type NewResourceCreateRequest struct {
    Name string `json:"name"`
}

type NewResourceUpdateRequest struct {
    Name *string `json:"name,omitempty"`  // Pointer for optional update
}
```

3. Add service to Client in `client.go`:
```go
// In Client struct
NewResources *NewResourceService

// In NewClient()
c.NewResources = &NewResourceService{client: c}
```

4. Define default field lists at top of service file:
```go
const (
    newResourceListFields = "most"
    newResourceGetFields  = "$key,name,status,created"
)
```

## API Patterns

- All methods take `context.Context` as first parameter
- List methods accept variadic `ListOption` arguments
- Create returns the created resource (reads back after POST)
- Update returns the updated resource (reads back after PUT)
- Delete returns only error
- Action methods (PowerOn, Clone, etc.) return only error

## Error Handling

Use typed error checking:
```go
if vergeos.IsNotFoundError(err) { ... }
if vergeos.IsAuthError(err) { ... }
if vergeos.IsValidationError(err) { ... }
```

## Related Documentation

- `DECISIONS.md` - Architecture Decision Records explaining design choices
- `README.md` - User-facing documentation with API examples
- `examples/` - Working code examples for common use cases
