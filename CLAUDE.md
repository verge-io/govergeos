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

### Core Pattern: Service-Oriented Design (85 services)

Each VergeOS resource has three pieces:

1. **Type definitions** in `types_{resource}.go` — request/response structs
2. **Service implementation** in `{resource}.go` — methods calling the API
3. **Interface** in `interfaces.go` — enables mocking for consumers

Services are initialized in `NewClient()` (in `client.go`) and exposed as interface-typed fields on the `Client` struct.

### Key Design Decisions

Detailed rationale in `DECISIONS.md` (ADR-001 through ADR-024). The critical ones:

- **FlexInt** (`types.go`): Custom type handling VergeOS API returning IDs as int or string. Exception: Volume service uses `string` keys (SHA1 hashes).
- **Pointer fields** in Update requests: `*int`, `*bool` etc. distinguish "not provided" (nil) from "set to zero value".
- **Functional options**: `WithBaseURL()`, `WithCredentials()`, `WithAPIKey()`, `WithEnvConfig()`, `WithMinimumVersion()`, `WithSkipVersionCheck()`, `WithRetry()`, `WithRateLimit()`, etc.
- **Context-first**: All API methods take `context.Context` as first parameter.
- **Actions return error only**: Clone, snapshot, power operations don't return the new resource.
- **Mandatory version and credential check**: `NewClient()` accepts VergeOS 26 and later (older majors are rejected) and validates the supplied credentials during initialization. A 401 from the credential check is not retried. A connection reset on that GET is retried like any other idempotent request. `WithMinimumVersion` changes the floor. `WithSkipVersionCheck` records the server version but does not reject it. Features that need a specific minor gate on `serverVersion` with `isVersionAtLeast`.
- **Retries**: GET, PUT, and DELETE retry connection resets, timeouts before a response, and HTTP 429, 502, and 503. POST is not retried. HTTP 401 is never retried. `WithRetry` tunes or disables this. `WithRateLimit` is an optional minimum interval between request starts.
- **WithEnvConfig() is opt-in**: Environment variables are not auto-read.
- **Auth source settings are merged on update**: The API replaces the settings object. `AuthSourceService.Update` reads the stored settings and merges the caller's keys before sending, so a partial update cannot drop `client_secret`. Returned auth sources omit `client_secret`. An OIDC application's `client_secret` is write-only: `Create` returns it as a `WriteOnlySecret`, and `OIDCApplication` does not keep it.
- **Recipe deploy checks answers, and preview is a different method**: `Catalogs` and `VMRecipes` read catalogs, recipes, and each recipe's questions. `VMRecipeInstances.Deploy` checks answers against those question types before POST. Bool values it does not recognize are refused. Disk sizes are bytes; a value above zero and under 1 MB is refused. `Preview` is the dry run. It cannot return an instance. A successful HTTP status on that POST is `RecipePreviewPersistedError`.

### Adding a New Service

1. Create `types_{resource}.go` with request/response structs
2. Create `{resource}.go` with service struct and methods
3. Add interface to `interfaces.go`
4. Initialize service in `NewClient()` in `client.go` and in `initServices` in `testhelper_test.go`
5. Add integration tests in `test/integration/` with `//go:build integration`

### Error Types

Defined in `errors.go`: `APIError`, `NotFoundError`, `AmbiguousNameError`, `AuthError`, `ValidationError`, `UnsupportedVersionError`. Check with `IsNotFoundError(err)`, `IsAmbiguousNameError(err)`, `IsAuthError(err)`, etc.

### Query Options

List operations use composable options from `options.go`: `WithFilter()`, `WithSort()`, `WithFields()`, `WithLimit()`, `WithOffset()`.

### Field Selection

Services define default field sets for List vs Get operations to optimize API payloads. Users can override with `WithFields("all")`.

### Power State Polling

VM power waits (`PowerOn`, `PowerOff`, `Kill`) default to a 150-second timeout polled every 5 seconds. `WithPowerWait` changes that default. `PowerOffWithOptions` overrides it for one shutdown. Network power operations still poll with 5-second intervals, max 30 retries.
