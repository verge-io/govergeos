# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

goVergeOS is a Go client library (SDK) for the VergeOS infrastructure platform. Pre-release, zero external dependencies (stdlib only), Go 1.21+, requires VergeOS 26.0+. Module path: `github.com/verge-io/govergeos`.

Foundation for the Terraform Provider, Prometheus Exporter, and other VergeOS tooling.

## Build & Test Commands

```bash
go build ./...                # Build
go test ./...                 # Unit tests
go test -v ./...              # Verbose
go test -run TestName ./...   # Single test
go vet ./...                  # Static analysis
go fmt ./...                  # Format

# Integration tests (require live VergeOS server + env vars)
go test -tags=integration -v ./test/integration/
go test -tags=integration -v ./test/integration/ -run "TestVM"
```

Integration tests use `//go:build integration` build constraint and require `VERGEOS_HOST`, `VERGEOS_USERNAME`/`VERGEOS_PASSWORD` (or `VERGEOS_API_KEY`) environment variables.

## Architecture

**Single flat package** (`vergeos`) — all code lives at the repository root. No nested packages.

### Core Pattern: Service-Oriented Design (77 services)

Each VergeOS resource has three pieces:

1. **Type definitions** in `types_{resource}.go` — request/response structs
2. **Service implementation** in `{resource}.go` — methods calling the API
3. **Interface** in `interfaces.go` — enables mocking for consumers

Services are initialized in `NewClient()` (in `client.go`) and exposed as interface-typed fields on the `Client` struct.

### Key Design Decisions

Detailed rationale in `DECISIONS.md` (ADR-001 through ADR-018). The critical ones:

- **FlexInt** (`types.go`): Custom type handling VergeOS API returning IDs as int or string. Exception: Volume service uses `string` keys (SHA1 hashes).
- **Pointer fields** in Update requests: `*int`, `*bool` etc. distinguish "not provided" (nil) from "set to zero value".
- **Functional options**: `WithBaseURL()`, `WithCredentials()`, `WithAPIKey()`, `WithEnvConfig()`, etc.
- **Context-first**: All API methods take `context.Context` as first parameter.
- **Actions return error only**: Clone, snapshot, power operations don't return the new resource.
- **Mandatory version check**: `NewClient()` validates server version during initialization.
- **WithEnvConfig() is opt-in**: Environment variables are not auto-read.

### Adding a New Service

1. Create `types_{resource}.go` with request/response structs
2. Create `{resource}.go` with service struct and methods
3. Add interface to `interfaces.go`
4. Initialize service in `NewClient()` in `client.go`
5. Add integration tests in `test/integration/` with `//go:build integration`

### Error Types

Defined in `errors.go`: `APIError`, `NotFoundError`, `AuthError`, `ValidationError`, `UnsupportedVersionError`. Check with `IsNotFoundError(err)`, `IsAuthError(err)`, etc.

### Query Options

List operations use composable options from `options.go`: `WithFilter()`, `WithSort()`, `WithFields()`, `WithLimit()`, `WithOffset()`.

### Field Selection

Services define default field sets for List vs Get operations to optimize API payloads. Users can override with `WithFields("all")`.

### Power State Polling

VM/Network power operations poll with 5-second intervals, max 30 retries.
