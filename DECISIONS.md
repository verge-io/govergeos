# Architecture Decision Records

This document captures key design decisions made during the development of goVergeOS.

---

## ADR-001: Flat Package Structure

**Date:** 2025-01-02

**Status:** Accepted

**Context:** Go SDKs can be organized with nested packages (e.g., `vergeos/vm`, `vergeos/network`) or as a flat single package.

**Decision:** Use a flat package structure with all code in the root `vergeos` package.

**Rationale:**
- Simpler import paths for users (`vergeos.NewClient()` vs `vergeos.New()` + `vm.Service`)
- Follows patterns of popular Go SDKs (aws-sdk-go, google-cloud-go client libraries)
- Reduces package coupling issues
- Easier to maintain type relationships across services

**Consequences:**
- All types share the same namespace (prefix types if conflicts arise)
- Single `go.mod` file simplifies dependency management

---

## ADR-002: FlexInt Type for ID Handling

**Date:** 2025-01-02

**Status:** Accepted

**Context:** The VergeOS API inconsistently returns IDs as either integers or strings depending on the endpoint and context.

**Decision:** Create a custom `FlexInt` type that implements `json.Unmarshaler` to handle both string and integer JSON values.

**Rationale:**
- Provides seamless handling of API inconsistencies
- Users always work with `int` values regardless of API response format
- Encapsulates parsing logic in one place

**Consequences:**
- Slight overhead for JSON unmarshaling
- Type must be used for all ID fields that may have inconsistent API responses

---

## ADR-003: Functional Options Pattern for Client Configuration

**Date:** 2025-01-02

**Status:** Accepted

**Context:** The SDK client requires multiple configuration options (base URL, credentials, TLS settings, timeouts).

**Decision:** Use the functional options pattern (`WithBaseURL()`, `WithCredentials()`, etc.).

**Rationale:**
- Clean, readable API for users
- Easy to add new options without breaking changes
- Optional parameters have sensible defaults
- Self-documenting code

**Consequences:**
- Slightly more verbose than a config struct for simple cases
- Each option requires its own function

---

## ADR-004: Pointer Types for Optional Update Fields

**Date:** 2025-01-02

**Status:** Accepted

**Context:** Update requests need to distinguish between "not provided" and "set to zero/empty value".

**Decision:** Use pointer types (`*string`, `*int`, `*bool`) for all fields in Update request structs.

**Rationale:**
- `nil` pointer = field not included in request (don't update)
- Non-nil pointer = update field to this value (including zero values)
- Matches common Go SDK patterns (AWS, GCP)

**Consequences:**
- Slightly more verbose for users (need to use `&value` or helper functions)
- Clear semantics for partial updates

---

## ADR-005: Service-Oriented Architecture

**Date:** 2025-01-02

**Status:** Accepted

**Context:** Need to organize API operations in a way that's intuitive and maintainable.

**Decision:** Group operations by resource type into service structs (VMService, NetworkService, etc.) accessed via the main Client.

**Rationale:**
- Intuitive API: `client.VMs.List()`, `client.Networks.Create()`
- Easy to find related operations
- Services can maintain their own state if needed
- Follows patterns of GitHub, Stripe, and other popular SDKs

**Consequences:**
- Client struct holds references to all services
- Adding new resources requires adding new service

---

## ADR-006: Explicit Field Selection

**Date:** 2025-01-02

**Status:** Accepted

**Context:** VergeOS API supports field selection to optimize response payloads.

**Decision:** Define default field lists (`vmListFields`, `vmGetFields`) and use them automatically, while allowing user override via `WithFields()` option.

**Rationale:**
- Optimizes API calls by default
- Users don't need to know field names for common operations
- Power users can customize when needed

**Consequences:**
- Must maintain field lists as API evolves
- Different field sets for List vs Get operations

---

## ADR-007: Drive Interface Default Changed to virtio-scsi

**Date:** 2025-01-02

**Status:** Accepted

**Context:** The VergeOS schema indicates `virtio-scsi` is now the recommended default disk interface, replacing legacy `virtio`.

**Decision:** Document `virtio-scsi` as the default/recommended interface in SDK type comments.

**Rationale:**
- Aligns with VergeOS schema defaults
- Better performance and feature support
- Legacy `virtio` still available for compatibility

**Consequences:**
- Documentation updated to reflect new default
- Existing code using `virtio` continues to work

---

## ADR-008: Action Methods Return Error Only

**Date:** 2025-01-02

**Status:** Accepted

**Context:** VM actions (Clone, Snapshot, Reset, etc.) are asynchronous operations that may not return immediate results.

**Decision:** Action methods return only `error`, not the resulting resource.

**Rationale:**
- Actions are fire-and-forget at the API level
- Clone/Snapshot create new resources with different IDs
- Users can query for results separately if needed
- Simpler, more honest API

**Consequences:**
- Users must query separately to get cloned VM or snapshot details
- Consistent behavior across all action methods
- Exception: `VMs.Snapshot` returns the created `*VMSnapshot`. Creating a VM snapshot is a synchronous `machine_snapshots` insert, not a fire-and-forget `vm_actions` call.

---

## ADR-009: Context Support for All Operations

**Date:** 2025-01-02

**Status:** Accepted

**Context:** Go best practices recommend using `context.Context` for cancellation and timeouts.

**Decision:** All SDK methods that make API calls accept `context.Context` as their first parameter.

**Rationale:**
- Enables request cancellation
- Supports timeouts
- Follows Go idioms and standard library patterns
- Required for production-grade applications

**Consequences:**
- All method signatures include context parameter
- Users must pass context (can use `context.Background()` for simple cases)

---

## ADR-010: MIT License

**Date:** 2025-01-02

**Status:** Accepted

**Context:** Need to choose an open source license for the SDK.

**Decision:** Use MIT License.

**Rationale:**
- Permissive license encourages adoption
- Compatible with most other licenses
- Simple and well-understood
- Standard for SDK/library projects

**Consequences:**
- Users can use SDK in proprietary projects
- No copyleft obligations

---

## ADR-011: "goVergeOS" Naming Convention

**Date:** 2026-01-21 (Updated: 2026-01-29)

**Status:** Accepted

**Context:** Verge is standardizing library naming across all language-specific packages:
- **PSVergeOS** - PowerShell module
- **pyvergeos** - Python package (lowercase module path)
- **goVergeOS** - Go library (this project)

The original name "VergeOS Go SDK" used the term "SDK" which implies more maturity and comprehensiveness than appropriate for what is essentially a Go API client library.

This project serves as the foundation for the Terraform Provider, Prometheus Exporter, and the upcoming Cluster API (CAPI) provider.

**Evolution of Naming:**

1. **Original**: `vergeos-go-sdk` - SDK terminology was too formal
2. **v0.1.0-alpha**: `goVergeOS` - Brand consistency with mixed case
3. **v0.1.0**: `govergeos` - Go-idiomatic lowercase module path

The mixed-case `goVergeOS` import path caused friction:
- Go module proxy encodes uppercase as `go!verge!o!s` in URLs
- Users naturally type lowercase and get errors
- No major Go libraries use mixed-case module paths (the `github.com/Sirupsen/logrus` → `github.com/sirupsen/logrus` rename is a cautionary tale)

**Decision:** Use lowercase `govergeos` for all code paths (module path, repository name, User-Agent) while keeping "goVergeOS" as the marketing/branding name in documentation and titles.

**Rationale:**
- Go conventions strongly prefer lowercase import paths
- All major Go SDKs use lowercase: `aws-sdk-go`, `go-github`, `stripe-go`
- The package name `vergeos` is what developers actually type in code
- Brand name in docs ("goVergeOS") vs. import path (`govergeos`) is a common pattern
- Matches `pyvergeos` Python package naming at the module level

**Consequences:**
- Repository: `github.com/verge-io/govergeos`
- Module path: `github.com/verge-io/govergeos`
- Import: `import vergeos "github.com/verge-io/govergeos"`
- User-Agent: `govergeos/1.0`
- Documentation titles: "goVergeOS" (branding preserved)
- Package name: `vergeos` (unchanged, fully idiomatic)
- Users can now `go get github.com/verge-io/govergeos` without case sensitivity issues

---

## ADR-012: Service Interfaces for Mock Testing

**Date:** 2026-01-21

**Status:** Accepted

**Context:** The SDK currently implements all services as concrete structs with no corresponding interfaces. This makes it difficult for consumers (Terraform Provider, Prometheus Exporter, CAPI Provider) to:
- Write unit tests with mocked SDK calls
- Swap implementations for different behaviors
- Use dependency injection patterns

Currently, testing SDK consumers requires either:
- Spinning up a real VergeOS instance (integration testing only)
- HTTP-level mocking with `httptest`, which is brittle and couples tests to implementation details

**Decision:** Add interface definitions for all services to enable mock testing and dependency injection.

**Proposed Implementation:**
1. Define an interface for each service (e.g., `VMServiceInterface`, `NetworkServiceInterface`)
2. Update Client struct to use interface types instead of concrete pointers
3. Concrete service structs continue to implement the interfaces
4. Optionally provide a `mock` subpackage with test doubles or `mockgen` compatibility

**Rationale:**
- Enables consumers to write fast, isolated unit tests
- Follows Go best practices for testable library design
- Matches patterns in other infrastructure SDKs (AWS, GCP, Azure)
- Supports dependency injection for advanced use cases
- No breaking changes for existing consumers (interfaces are additive)

**Consequences:**
- Additional code to maintain (interface definitions mirror method signatures)
- Must keep interfaces in sync when adding/modifying service methods
- Enables significantly better testing experience for all SDK consumers
- Positions SDK as enterprise-ready with professional testing support

---

## ADR-013: API Schema Source and Type Mapping

**Date:** 2026-01-23

**Status:** Accepted

**Context:** The SDK initially used JSON schema files copied from `/usr/lib/appserver/schema/v4/` on a VergeOS system. Testing against live API revealed that these schema files describe *logical types* but not *runtime JSON serialization*. Key discrepancies discovered:

1. **`parent_system` fields**: Schema shows `$type: "row"` (integer) but API returns `"self"` (string)
2. **Nullable row fields**: Schema doesn't indicate when foreign key fields can be `null`
3. **Polymorphic IDs**: `$key` fields sometimes serialize as strings instead of integers

The internal API team clarified that filesystem schema files are pre-translation definitions and are **not authoritative**.

**Decision:**

1. **Schema Source**: Use runtime-extracted schema via `root-yb-api /v4 -f 'name,schema'` command as the authoritative source. Store in `.claude/reference/API-Schema/`.

2. **Type Mapping Rules**:
   - `parent_system` type fields → `string` (returns special values like `"self"`)
   - Optional `row` fields → `*int` or `FlexInt` (may be `null`)
   - Required `row` fields → `int` or `FlexInt`
   - All `$key` fields → `FlexInt` (handles string/int polymorphism)
   - `rows` type → omit from struct (one-to-many, not directly serialized)

3. **Verification Requirement**: Always test new field mappings against live API before committing.

**Rationale:**
- Runtime schema reflects actual API behavior after server-side translation
- Explicit type mapping rules prevent repeated discovery of the same issues
- `FlexInt` already exists in SDK (ADR-002) and handles ID polymorphism
- Documented exceptions prevent future developers from "fixing" correct behavior

**Consequences:**
- Schema in `.claude/reference/API-Schema/` is the single source of truth
- Deprecated schema folders (`api-schema-old-local/`, `dz/`) can be removed
- Type mapping exceptions documented in `ENDPOINTS.md` for quick reference
- Must verify against live API when schema `$type` doesn't match expected Go type

**References:**
- GitHub Issue: https://github.com/verge-io/goVergeOS/issues/2
- Related: ADR-002 (FlexInt Type for ID Handling)
- Documentation: `.claude/reference/API-Schema/ENDPOINTS.md`

---

## ADR-014: Volume String Keys

**Date:** 2026-01-23

**Status:** Accepted

**Context:** When implementing the Volumes service, testing against the live API revealed that volumes use SHA1 hash strings as their primary key (`$key`), not integers like other resources. The schema defines `"keyfield": "id"` where `id` is a 40-character SHA1 hash.

API Response example:
```json
[{"$key":"0d25c256a0c561c0b5bb9087f04fcb49f16a8048","id":"0d25c256a0c561c0b5bb9087f04fcb49f16a8048","name":"system-logs",...}]
```

**Decision:** The Volume type uses `string` for its Key field instead of `FlexInt`. All Volume service methods accept string IDs.

```go
type Volume struct {
    Key string `json:"$key,omitempty"`  // SHA1 hash, not FlexInt
    ID  string `json:"id,omitempty"`    // Same as Key for volumes
    // ...
}

func (s *VolumeService) Get(ctx context.Context, id string) (*Volume, error)
func (s *VolumeService) Update(ctx context.Context, id string, req *VolumeUpdateRequest) (*Volume, error)
func (s *VolumeService) Delete(ctx context.Context, id string) error
```

**Rationale:**
- Volumes are the first (and possibly only) resource type with string keys
- Using `string` directly is cleaner than extending `FlexInt` to handle this case
- The schema clearly indicates `id` is a 40-character SHA1 hash string
- Action endpoints (`volume_actions`) accept the string ID as the `volume` field

**Consequences:**
- Volume service has different method signatures than other services
- Must document this exception in interface comments
- Future resources with non-integer keys should follow this pattern
- Added `getStringKey()` helper function for extracting string keys from API responses

**Related:**
- ADR-002 (FlexInt Type for ID Handling) - contrast with integer key handling
- ADR-013 (API Schema Source) - verifying types against live API

---

## ADR-015: VNet Rule Enable/Disable Uses PUT Instead of Action Endpoint

**Date:** 2026-01-26

**Status:** Accepted

**Context:** The VNet Rules service initially implemented `Enable()` and `Disable()` methods that attempted to POST to a `/vnet_rule_actions` endpoint, following the pattern used by VMs (`/vm_actions`) and Networks (`/vnet_actions`). However, integration testing revealed this endpoint does not exist.

Analysis of the API schema showed:
1. The `vnet_rules.json` schema defines `enable` and `disable` actions within the table schema (lines 358-412)
2. No `vnet_rule_actions.json` endpoint file exists in the schema (45 other `*_actions.json` files exist)
3. The main `v4.json` schema has no `vnet_rule_actions` entry
4. The `enabled` field is a standard writable boolean field on the rule

This is an exception to the typical VergeOS API pattern where resources have dedicated `*_actions` endpoints for state-changing operations.

**Decision:** Implement `Enable()` and `Disable()` methods using PUT to update the `enabled` field directly, then optionally call `Networks.ApplyRules()` to apply the changes.

```go
func (s *VNetRuleService) Enable(ctx context.Context, id int, apply bool) error {
    return s.setEnabled(ctx, id, true, apply)
}

func (s *VNetRuleService) setEnabled(ctx context.Context, id int, enabled bool, apply bool) error {
    rule, err := s.Update(ctx, id, &VNetRuleUpdateRequest{Enabled: &enabled})
    if err != nil {
        return err
    }
    if apply && rule.VNet > 0 {
        return s.client.Networks.ApplyRules(ctx, int(rule.VNet))
    }
    return nil
}
```

**Rationale:**
- The `/vnet_rule_actions` endpoint does not exist in the VergeOS API
- Updating the `enabled` field via PUT is the standard RESTful approach
- The `Networks.ApplyRules()` method already exists and correctly calls `/vnet_actions` with action "apply"
- This approach composes existing, working functionality rather than inventing new endpoints
- Removed the `forceApply` parameter since the underlying `ApplyRules()` doesn't support it

**Consequences:**
- Breaking change: `Enable()` and `Disable()` signatures changed from `(id, apply, forceApply)` to `(id, apply)`
- Method now requires two API calls when `apply=true` (PUT + POST to vnet_actions)
- Follows RESTful conventions more closely than action-based approach
- Documents an API exception that future maintainers should be aware of

**Related:**
- ADR-013 (API Schema Source) - importance of verifying against live API
- `.claude/reference/API-Schema/ENDPOINTS.md` - documents this exception

---

## ADR-016: Mandatory Version and Credential Check in NewClient

**Date:** 2026-01-29

**Revised:** 2026-09-30

**Status:** Accepted

**Context:** The VergeOS API uses `/api/v4/` for all endpoints regardless of the actual software version. This means a server running VergeOS 4.x and one running 26.x both expose the same `/api/v4/` path, making it impossible to detect version mismatches from API paths alone.

The SDK targets VergeOS 26 and later. It relies on features, fields, and behaviors introduced in 26. Using the SDK against an older VergeOS installation may result in:
- Missing fields in API responses
- Endpoints that don't exist
- Different behavior for existing endpoints
- Confusing partial failures that are hard to diagnose

VergeOS major versions are year-based. Requiring an exact match on major 26 would make `NewClient` fail for every consumer the day 27 ships, even when the API those consumers use did not change. Individual features that appeared after 26.0 gate themselves with `isVersionAtLeast` against the version recorded at startup (for example, the drive field added in 26.1.5).

`/version.json` answers without credentials. A version check alone lets `NewClient()` succeed with a wrong password, and the `AuthError` (`Login required`) appears only on the first real API call. Terraform then reports a successful `configure` and fails halfway through a plan. A long-running exporter starts cleanly and fails every scrape.

**Decision:** Perform a blocking version check in `NewClient()` that fetches `/version.json` and accepts VergeOS 26 and every later major. A major older than 26 fails client creation with an `UnsupportedVersionError`. The comparison is a minimum, not an equality check.

Two options adjust that floor:

- `WithMinimumVersion(major)` replaces the default floor of 26. A caller can raise it when their application needs a newer major, or lower it to keep talking to an older server.
- `WithSkipVersionCheck()` still reads and stores the server version, so feature gates keep working, but does not reject the major. The credential check still runs. This is the escape hatch for a deployment that must keep running if a later SDK release narrows the accepted range.

Credentials are validated too. After the version check, `NewClient()` makes one cheap authenticated request: `GET /api/v4/clusters?limit=1&fields=$key`. Clusters is a small table present on every system. If that request fails authentication, `NewClient()` returns the `AuthError`. The credential check is skipped when the version check already failed.

A 401 from the credential check is never retried. VergeOS locks an account after a configurable number of failed logins (five on the system this was measured against). Retrying a rejected password locks the account for every client that shares it, including its API keys. The version request does not send an `Authorization` header, so a rejected password is presented only on that one authenticated read. If that GET is reset or times out before a response, it is retried like any other idempotent request (ADR-020). Those retries are not failed logins. The default cap is 3 attempts, which stays under the lockout threshold measured above.

```go
client, err := vergeos.NewClient(
    vergeos.WithBaseURL("https://host"),
    vergeos.WithCredentials("user", "pass"),
)
if err != nil {
    // Unsupported version:
    // "unsupported server version 4.2.0: this SDK requires VergeOS 26.0 or later"
    // Wrong password:
    // "vergeos: authentication failed: Login required"
    log.Fatal(err)
}

// Keep running if a future release rejects this server's major.
client, err = vergeos.NewClient(
    vergeos.WithBaseURL("https://host"),
    vergeos.WithCredentials("user", "pass"),
    vergeos.WithSkipVersionCheck(),
)
```

**Alternatives Considered:**

1. **Lazy check on first API call** - Rejected because it delays error discovery; users might write significant code before hitting the check
2. **Optional check via `WithVersionCheck()` option** - Rejected because it leads to confusing partial failures when users forget to enable it
3. **No check at all** - Rejected because the SDK may silently malfunction on incompatible versions
4. **Warning instead of error** - Rejected because warnings are easily ignored and don't prevent the underlying compatibility issues
5. **Retry the credential check** - Rejected because each failed login counts toward account lockout
6. **Authenticate by calling `/version.json`** - Rejected because that endpoint does not require credentials
7. **Exact major match (26.x only)** - Rejected because a year-based major would break every consumer on the day the next major ships, including when the API did not change
8. **Cap the newest accepted major** - Rejected because it recreates the same outage for the next major. New behavior is gated per feature with `isVersionAtLeast`

**Rationale:**
- **Fail fast on known-old servers** - Users get immediate, clear feedback if the server is older than 26 or the credentials are rejected
- **Later majors keep working** - A new major does not require a govergeos release, or a release of every consumer, before `NewClient` succeeds
- **Feature gates stay precise** - Code that depends on a specific minor still checks `serverVersion` with `isVersionAtLeast`
- **Escape hatch** - `WithSkipVersionCheck` and `WithMinimumVersion` let a caller proceed when the default floor does not match their server
- **Simple implementation** - Two checks at startup, no caching or repeated checks needed
- **Clear error message** - Users know exactly what's wrong and what version is required
- **Choke point** - `NewClient()` is the natural place to validate prerequisites before any API calls
- **Lockout safety** - A rejected password is presented once. The default retry cap cannot turn one rejection into a lockout

**Consequences:**
- `NewClient()` requests `/version.json` without credentials, then one authenticated `clusters` row, before returning
- Client creation fails if the server major is older than 26, unless `WithMinimumVersion` or `WithSkipVersionCheck` says otherwise
- Client creation succeeds on VergeOS 27 and later without an SDK release
- Client creation fails with an `AuthError` when the credentials are rejected
- Users connecting to older VergeOS installations must use an older SDK version, lower the floor with `WithMinimumVersion`, or skip the check
- New `UnsupportedVersionError` type and `IsUnsupportedVersionError()` helper added
- Adds `getAbsolute()` internal method for fetching non-API paths
- A principal that cannot read `clusters` cannot construct a client, because that read is the credential check

---

## ADR-017: Explicit Environment Configuration

**Date:** 2026-01-30

**Status:** Accepted

**Context:** The pyVergeOS SDK provides `VergeClient.from_env()` for convenient client creation from environment variables. The Go SDK currently requires manual `os.Getenv()` calls in every example and integration test. We need to add equivalent functionality while avoiding security and debugging pitfalls.

Three approaches were considered:

1. **`NewClientFromEnv()` function** - Separate constructor that reads environment variables
2. **Automatic env var fallback in `NewClient()`** - Check env vars when options not provided
3. **`WithEnvConfig()` option** - Explicit opt-in via functional option

**Decision:** Use `WithEnvConfig()` as an explicit opt-in option that can be composed with other options.

```go
// Simple: all config from environment
client, err := vergeos.NewClient(vergeos.WithEnvConfig())

// Advanced: env vars with explicit override
client, err := vergeos.NewClient(
    vergeos.WithEnvConfig(),
    vergeos.WithTimeout(60*time.Second),  // Override timeout from env
)
```

**Alternatives Rejected:**

1. **Automatic env var fallback** - Rejected because:
   - Implicit behavior makes debugging difficult ("why is it connecting to X?")
   - Security risk from accidental credential pickup
   - Test pollution when real credentials leak into test environment

2. **Separate `NewClientFromEnv()` function** - Rejected because:
   - Two entry points increases API surface
   - Cannot easily combine env vars with explicit overrides
   - Less composable than functional options pattern (ADR-003)

**Rationale:**
- **Explicit opt-in** - Environment variables are never automatically picked up, matching pyVergeOS behavior
- **Composable** - Works with existing functional options pattern; explicit options override env vars
- **Single entry point** - `NewClient()` remains the only constructor
- **Debuggable** - When `WithEnvConfig()` is in the code, the intent is clear
- **Secure by default** - No accidental credential pickup from environment

**Environment Variables Supported:**
| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VERGEOS_HOST` | Yes | - | Host or base URL. No scheme means `https://` |
| `VERGEOS_USERNAME` | No* | - | Basic auth username |
| `VERGEOS_PASSWORD` | No* | - | Basic auth password |
| `VERGEOS_API_KEY` | No* | - | Bearer token auth. Wins when username and password are also set |
| `VERGEOS_VERIFY_SSL` | No | `true` | TLS certificate verification |
| `VERGEOS_INSECURE` | No | `false` | `true` skips TLS verification, same as `VERGEOS_VERIFY_SSL=false`. An error if they disagree |
| `VERGEOS_TIMEOUT` | No | `30` | Request timeout (seconds) |

*One of (USERNAME+PASSWORD) or API_KEY required. When both are set, the API key is used, matching `WithAPIKey` together with `WithCredentials`.

**Amendment:** `VERGEOS_HOST` may be a bare host or IP. A value with no scheme is read as `https://`; a scheme other than `http` or `https` is rejected and the error names `VERGEOS_HOST`. `VERGEOS_INSECURE` is accepted so an Ansible env file can drive the client. `true` (also `yes`, `on`, or `1`) skips certificate verification. If `VERGEOS_INSECURE` and `VERGEOS_VERIFY_SSL` are both set and disagree, `WithEnvConfig` returns an error.

**Amendment:** HTTP settings are collected while options run and applied once, to a copy of the base client. `WithHTTPClient` only supplies that base. Timeout, TLS, and rate limit still apply when it is passed before or after `WithEnvConfig`, `WithTimeout`, or `WithInsecureTLS`. A later timeout or TLS option replaces an earlier one. The caller's `*http.Client` and its transport are not modified. Skipping verification clones an `*http.Transport`; any other transport type is an error from `NewClient`.

**Consequences:**
- New `WithEnvConfig()` option function added to `client.go`
- Environment variables use `VERGEOS_` prefix (matching existing convention)
- Default SSL verification is `true` (secure by default, matching pyVergeOS)
- Integration tests can use `WithEnvConfig()` to reduce boilerplate
- Examples demonstrate both explicit options and env-based configuration

**Related:**
- ADR-003 (Functional Options Pattern) - `WithEnvConfig()` follows this pattern
- Plan: `.claude/plans/clientenvsslconfig.md`

---

## ADR-018: Integration Test Rate Limiting

**Date:** 2026-01-30

**Status:** Superseded by ADR-020

The test-only transport described below has been removed. Idempotent requests are retried in the client, and spacing is an opt-in `WithRateLimit` option. The history is kept because it records how the server fails under load.

**Context:** Running the full integration test suite (~400 API calls) against a VergeOS server caused intermittent `connection reset by peer` and `EOF` errors. Investigation revealed:

1. The VergeOS API does not return rate limit headers (`X-RateLimit-*`, `Retry-After`)
2. The API does not return HTTP 429 responses when overloaded
3. Instead, the server drops TCP connections when overwhelmed by rapid requests
4. The default server settings include "Webserver max session API rate limit: 50" which may contribute to connection drops

At ~36 requests/second, the server's connection handling became unstable, causing random test failures.

**Decision:** Add a rate-limited HTTP transport wrapper in `test/integration/helpers_test.go` that enforces a 50ms delay between requests (~20 requests/second max). This is test-infrastructure only and does not affect the SDK itself.

```go
type rateLimitedTransport struct {
    transport http.RoundTripper
    delay     time.Duration
    mu        sync.Mutex
    lastReq   time.Time
}
```

The SDK itself does NOT include built-in rate limiting because:
- VergeOS is on-premises software where users control their own server settings
- Rate limits are protective (preventing overload), not monetized
- Different deployments have different limits based on hardware and configuration
- Users can increase server-side limits if needed

**Rationale:**
- Test reliability is critical for CI/CD and development workflows
- The rate limiting is isolated to tests, not imposed on SDK users
- 50ms delay is a reasonable trade-off: tests complete in ~25 seconds vs ~11 seconds
- Matches the philosophy of ADR-017 (explicit opt-in) - users control their own rate limiting

**Consequences:**
- Integration tests now pass consistently without connection errors
- Test suite runs in ~25 seconds (previously ~11 seconds when it worked, but often failed)
- The SDK does not include built-in rate limiting - users manage this themselves
- Future consideration: Add a `RateLimitError` type for graceful 429 handling if the API ever returns them
- Future consideration: Document that users making many rapid requests may need client-side throttling

**Implications for SDK Users:**
1. **No built-in rate limiting** - Users control their own servers and can tune limits
2. **Should add `RateLimitError` type** - For graceful handling if 429s are ever returned
3. **Document connection reset behavior** - Users making many rapid requests may need their own throttling

**Related:**
- ADR-017 (Explicit Environment Configuration) - Same philosophy of explicit opt-in
- `test/integration/helpers_test.go` - Implementation location

---

## ADR-019: VM $key and Machine Key Are Different

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Every VM has a row key (`VM.Key`, `vms.$key`) and a machine key (`VM.Machine`). Drives, NICs, devices, and snapshots are stored against the machine key. `vm_actions` (power, hotplug, restore) uses the VM $key. On a system that has been in use the two counters differ, and each number is often some other object's key.

**Decision:** Public methods that identify a VM take the VM $key and resolve the machine key internally. `vm_actions` calls keep using the VM $key. Restore resolves `snap_machine` to the snapshot VM row (`is_snapshot` true) and posts that VM $key. Drive and NIC delete resolve the live VM (`is_snapshot` false) for the device's machine key before hot-unplug.

**Rationale:**
- Callers otherwise pass `VM.Key` into a parameter documented as an ID and attach the device to a different machine, or post a machine key to `vm_actions` and act on a different VM.
- `HotplugDrive` already took the VM $key. One parameter name had both meanings.
- pyVergeOS resolves the same two keys this way.

**Consequences:**
- `VMDrives`, `VMNICs`, and `VMDevices` `List` and `Create`, `VMDrives.GetByName`, and `VMSnapshots.ListByVM` / `GetByName` take the VM $key.
- `VMSnapshotCreateRequest.VM` is the VM $key. `Create` posts the resolved machine key.
- Cloud-init file `owner` is `vms/<VM $key>`. `CloudInitFiles.CreateForVM` and `ListByVM` take the VM $key and do not use `VM.Machine`.
- A call with a machine key where a VM $key is required now looks up the wrong VM, or returns not found, instead of silently targeting another machine.

---

## ADR-020: Retry Idempotent Requests and Optional Rate Limiting

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Under load, VergeOS drops a TCP connection instead of returning a rate-limit response. A cloud snapshot update failed once in four runs with `connection reset by peer` on the PUT. The next runs passed. The integration helper hid this with a transport that slept 50ms between requests. That workaround was not in the SDK, so a Terraform apply or an exporter scrape failed on a single reset.

The standard library does not retry this. Its transport retries an idempotent request only when the connection was already reused, and it does not treat PUT or DELETE as replayable. A fresh connection that is reset while the response is being read is returned to the caller. VergeOS also does not send `Retry-After` or `X-RateLimit-*` headers. When it is overloaded it sometimes returns 429, 502, or 503, and sometimes it sends a RST or closes the connection (EOF).

Repeating a rejected login is unsafe. The platform locks an account after a small number of failed logins (five on the system this was measured against), and the lock applies to every client that shares the account, including API keys. See ADR-016.

**Decision:** Retry GET, PUT, and DELETE when the attempt fails before a response, or the response is 429, 502, or 503. "Before a response" means the connection was reset, the peer closed it (EOF or unexpected EOF), or the client timed out while the request context is still active. The default is 3 attempts. The delay starts at 100ms, doubles up to 2s, and uses equal jitter so a set of clients does not retry on the same instant. `WithRetry` replaces that policy. `RetryPolicy{MaxAttempts: 1}` disables it.

POST is not retried. Creates and actions are not idempotent. A POST that fails with no response is returned to the caller, including a reset, a timeout, and 429, 502, and 503.

HTTP 401 is never retried, for any method. A 401 from `NewClient`'s credential check is still one presentation of the password. A reset on that GET is retried, because the server did not reject the credentials. Three attempts stay under the measured lockout threshold of five.

The `Timeout` on the `http.Client` remains the budget for the whole call, including backoff and later attempts. A timeout that has already cancelled the request is not retried.

`WithRateLimit(interval)` is optional and off by default. It spaces the start of each attempt, including retries, by at least `interval`. Deployments that burst faster than the server's per-session API limit (50 by default) can pass `50 * time.Millisecond` instead of copying the old test transport. Integration tests do not set it, so they exercise the retry path.

A response body that cannot be replayed (`GetBody` is nil) is not retried. `request` sends JSON from a `*bytes.Reader`, which records `GetBody`. Chunk uploads do the same: `uploadChunk` passes the chunk bytes as a `*bytes.Reader`, so a dropped connection retries that PUT at the same `filepos`.

**Alternatives Considered:**

1. **Keep the delay only in the integration tests** - Rejected because production clients hit the same resets
2. **Retry every method, including POST** - Rejected because a create or action may have been applied before the connection dropped
3. **Retry 401** - Rejected because it locks the account (ADR-016)
4. **Rate-limit by default** - Rejected because the right interval depends on the server configuration, and a default delay would slow every caller
5. **Honor `Retry-After`** - Rejected for this change because the server does not send it

**Consequences:**

- A single connection reset on GET, PUT, or DELETE is retried up to the configured cap
- POST failures are unchanged: one attempt, then the caller's error
- 401 is one attempt
- `WithRetry` and `WithRateLimit` are client options
- The integration helper no longer sleeps between requests
- A failure while reading a body after response headers were received is not retried here; the status was already delivered to `RoundTrip`

**Related:**

- ADR-016 (credential check and account lockout)
- ADR-018 (the test transport this replaces)

---

## ADR-021: GetByName Returns One Object or an Error

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Every `GetByName` listed with a name filter and returned the first row. VergeOS allows the same tag name in two categories, so `Tags.GetByName("zzgo-dup")` returned one of them with a nil error. A later assign or delete then acted on the other tag. Tables that reject duplicate names today can stop doing so. The same class of lookup was treated as a correctness bug in the Ansible collection.

**Decision:** When a name lookup matches more than one row, return `AmbiguousNameError` with the matching keys. Do not pick one. Callers that want a specific row use `List`. `Tags.GetByName` takes the category, the same way `SnapshotProfilePeriods.GetByName` takes the profile. The multi-match check stays on every `GetByName`, including tables the platform currently keeps unique.

**Rationale:**
- Returning the first row lets a caller update or delete an object it did not name.
- A category argument makes the tag lookup name the object that was asked for. Names are unique inside a category.
- One length check turns a later platform change into an error instead of a wrong object.

**Consequences:**
- `Tags.GetByName(ctx, categoryID, name)` replaces `Tags.GetByName(ctx, name)`.
- `IsAmbiguousNameError` identifies the new error. `Keys` is the list of matching resource keys.
- A single exact match is unchanged, including the name check from the filter-escaping fix.

---

## ADR-022: VM PowerOff Is a Graceful Shutdown That Waits

**Date:** 2026-09-30

**Status:** Accepted

**Context:** `VMService.PowerOff` sent the `kill` action and then polled until the VM stopped. `kill` is a hard power-off. The graceful action, `poweroff`, was only available from `GuestShutdown`, which returns as soon as the request is accepted. Callers had to choose a shutdown that can corrupt a guest, or a shutdown they could not wait on.

`PowerOff` is the method the README and the VM docs show for an ordinary stop. The Terraform provider's v3 port is moving onto this SDK, and a hard kill on `powerstate = false` is the bug that port is meant to leave behind. The wait itself was also fixed at 30 polls of 5 seconds (150 seconds). A context deadline can shorten that wait. It cannot make it longer. A busy Windows guest can take longer than 150 seconds to shut down.

`NetworkService.Kill` and `TenantNodeService.Kill` already name the hard stop `Kill`. `docker-machine-driver-vergeos` calls `PowerOff` for its force stop and before deleting a VM, and `GuestShutdown` for a graceful stop that does not wait.

**Decision:** `PowerOff` sends `poweroff` and waits until the VM stops. `Kill` sends `kill` and waits the same way. `GuestShutdown` still sends `poweroff` and returns immediately.

`PowerOffWithOptions` takes a timeout, a poll interval, and `ForceAfterTimeout`. When the graceful wait hits its timeout and `ForceAfterTimeout` is set, the call sends `kill` and waits again. A cancelled context does not escalate to `kill`.

`WithPowerWait` sets the default timeout and poll interval used by `PowerOn`, `PowerOff`, and `Kill`. The default stays 150 seconds polled every 5 seconds. A per-call timeout longer than that default extends the wait. A context deadline can still end it sooner.

**Rationale:**
- The method people already call for "stop this VM" should be the one that asks the guest to shut down and then waits.
- The hard stop keeps a name, `Kill`, that the network and tenant-node services already use, so a force stop stays explicit.
- A fixed 150-second budget cannot cover a slow guest, and a context deadline cannot raise it.
- `ForceAfterTimeout` lets a caller escalate on purpose, instead of every shutdown being a kill.

**Consequences:**
- This behavior change shipped in v0.3.1. Code that used `PowerOff` as a hard kill, including the docker-machine driver's force stop, must call `Kill`.
- `PowerOff` can now return `TimeoutError` when the guest ignores the shutdown. `Kill`, or `PowerOffWithOptions` with `ForceAfterTimeout`, is how to stop that VM anyway.

---

## ADR-023: Auth Source Settings Are Merged, and Client Secrets Are Write-Only

**Date:** 2026-09-30

**Status:** Accepted

**Context:** VergeOS replaces an auth source's `settings` object on update. A PUT that names only `scope` deletes `client_id`, `client_secret`, and the endpoints. The next login fails, and the secret usually cannot be read back from the identity provider. pyVergeOS measured this on 26.1.8 (pyVergeOS#142). The stored document also gains a `debug` key the client did not send.

OIDC applications and auth source settings both carry a `client_secret`. The auth source secret is supplied by the caller. The OIDC application secret is generated by VergeOS. Either one is enough to authenticate as that client. List and Get responses, formatted values, and logs must not retain it.

**Decision:** `AuthSourceService.Update` reads the current settings and shallow-merges the caller's keys on top before PUT. Keys the caller omits, including `client_secret` and server-injected keys, stay. The merge is not atomic: a write that lands between the read and the PUT is lost.

`client_secret` is write-only. `AuthSourceSettings` drops it when decoding a response, and formatting prints `[redacted]` instead of the value. JSON encoding of a request still sends the secret, which is how a create or a secret rotation works. `OIDCApplication` decodes and discards `client_secret`. `OIDCApplicationService.Create` returns that generated secret once, as a `WriteOnlySecret`. `Value` reads it. Printing the `WriteOnlySecret` does not.

**Rationale:**
- Callers updating one setting, including the Terraform auth source resource, would otherwise wipe a working SSO configuration with no error.
- The secret has to be on the wire for the merge to preserve it, and it must not be on the value those callers log.
- Returning the OIDC secret only from `Create` matches API keys: the credential is available when it is issued, and later reads do not carry it.

**Consequences:**
- An update cannot delete a settings key by omitting it. Send the key with a new value to change it.
- `Get` and `List` of an auth source do not include `client_secret`. `Get` still returns the other settings.
- Code that logs an `AuthSource`, an `OIDCApplication`, or a `WriteOnlySecret` does not log the client secret. `Value` and the JSON body of a request are the two places the secret is visible.

---

## ADR-024: Recipe Answers Are Checked Before Deploy, and Preview Cannot Deploy

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Most VergeOS VMs are created by answering a catalog recipe, not by posting a raw VM. The Terraform provider and other automation need that path (issue #58). Two answer shapes are accepted by the platform and then silently ignored. A bool question given a string it does not know, such as `enabled`, is stored as false, so a UEFI request builds a BIOS VM and the deploy still succeeds. A `disksize` answer is a byte count. `50`, sent by a caller who meant 50 GB, is fifty bytes; the recipe keeps the image's own disk size and the deploy still succeeds. Zero is the recipe's own default and has to keep working.

The platform's dry run is a POST to the same collection with `simulate: true`. It answers HTTP 405, body `{"err":"Simulation complete", ...}`. A 2xx response means a VM was created. Folding that into Deploy with a boolean would repeat the Ansible module's failure mode, where a check still deployed.

**Decision:** `Catalogs` and `VMRecipes` are read services. `VMRecipes.Questions` returns each question and its type. `VMRecipeInstances.Deploy` loads those questions and refuses the POST until the answers match. Recognized bools are `true`/`false`, the words `true`, `yes`, `on`, `1`, `false`, `no`, `off`, `0`, an empty string (false), and the integers 0 and 1. Any other bool answer is a validation error. A `disksize` above zero and under 1 MB (1048576) is a validation error, and the message says the unit is bytes and shows 50 GB as 53687091200. Network names resolve to vnet keys. An ambiguous name is refused. `__new_internal__` is sent unchanged.

`Preview` takes the same request struct and is a different method with a different result type, `*VMRecipePreview`. The request struct has no dry-run flag. Preview sends `simulate: true`. HTTP 405 with the simulation report is success. Log lines that record a failed step are `RecipePreviewFailedError`. A 2xx response is `RecipePreviewPersistedError` and is not turned into an instance.

**Rationale:**
- The two silent successes are the ones that ship a VM the caller did not ask for.
- A separate method and result type is what makes a dry run impossible to use as a deploy.
- Treating a successful preview POST as failure matters because that status is how the platform reports a real create.

**Consequences:**
- Callers pass disk sizes in bytes. A gigabyte value has to be converted before Deploy.
- Table-backed choices (a storage tier, for example) are still checked by the platform. Preview is how to see that failure before Deploy.
- Tenant recipes are not part of this surface.

## ADR-025: VM Import Wait Fails Fast, and a Repeated Name Is Idempotent

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Importing an OVA or OVF, and exporting a VM to a NAS volume, are the paths automation uses when a catalog recipe is not the source (issue #59). Both already exist on the platform as `vm_imports` and `volume_vm_exports`. A bare disk image is not a `vm_imports` file type (issue #106).

Two behaviors in the Ansible collection are wrong for a client library. A failed import can stay in a status the waiter still treats as unfinished, so the caller sits through the whole timeout after the platform has already recorded the failure. Drive failures show up as `failed_drive_count` while status is still `importing`. Separately, `vm_imports` names are not unique and the platform keeps finished rows. Deleting by name has to remove every finished row, and it has to refuse the sweep when one of several rows with that name is still importing. A second create for a VM name that already exists is rejected by the platform, because the VM name is taken.

**Decision:** `VMImports` creates an import from one source: a media-catalog file, an http(s) URL, a NAS volume path, or a shared object. A URL is stored with `Files.Create` and the import then uses that file id. The file is an OVA or OVF. VergeOS rejects a bare disk image (qcow2, vmdk, vhd, raw, img) on `vm_imports` with status `error` and `Unknown Import file type`. Attach that file with `VMDrives.Create` and `Media` set to `import`. `Importing` defaults to true. `Wait` returns `VMImportFailedError` as soon as the row is `error` or `aborted`, the aborted flag is set, or `failed_drive_count` is greater than zero. Status `complete` and `warning` are success even when a stale failed-drive count is still set. The error includes `status_info` and up to ten error or critical log lines from `VMImportLogs`. The default wait is 10 minutes, polled every 5 seconds.

`Create` does not post when a non-snapshot VM with that name already exists. It returns the single matching import row, or a row with an empty key when there are zero or several, and sets `AlreadyExisted`. A 409 from the post is adopted the same way when the VM is present. `GetByName` still returns `AmbiguousNameError` when two rows share a name. `DeleteByName` is success when no row matches. One row is deleted even if it is still importing. Two or more rows are deleted only when every one has finished. If any of them is still importing, `DeleteByName` returns `VMImportInProgressError` and deletes nothing.

`VMExports` is the configuration on `volume_vm_exports`, one row per NAS volume. The volume key is the volume's SHA1, quoted in filters. The export row id is an integer. `Run` creates that configuration when the volume does not have one, updates settings that differ, and starts the export. `Wait` returns `VMExportFailedError` as soon as status is `error`. Status `building` and `cleaning` keep polling. Any other non-empty status is done, and the newest `volume_vm_export_stats` row then fails the wait when its error count is greater than zero. The default export wait is 1 hour, polled every 5 seconds. `Run` does not wait.

**Rationale:**
- The failure the platform has already recorded is the result. Waiting out the timeout hides it.
- A duplicate VM name cannot create a second VM, so a second post is not a useful retry.
- Name-only delete has to be safe when several historical rows share the name and one of them is live.

**Consequences:**
- Callers that need a specific import row use its hex key. `GetByName` will not pick one of several.
- `Delete` by key still removes a live import. `DeleteByName` will not, when other rows share the name.
- Deleting an import row leaves the VM. Deleting an export configuration leaves files already written to the volume.
- The file download started for a URL import is not polled. A download that fails is reported by `Wait` on the import.

## ADR-026: The $key Field Is Named Key

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Every VergeOS row has a `$key`. Most Go types named that field `Key`. Thirteen named it `ID`: `VM`, `Network`, `VMNIC`, `VMDrive`, `VMDevice`, `USBDeviceSettings`, `TPMDeviceSettings`, `VGPUDeviceSettings`, `Group`, `File`, `Member`, `CloudInitFile`, and `ResourceGroup`. `VM` and `Network` are the types callers use most, so code that handles a VM and a tenant wrote `vm.ID` next to `tenant.Key`. Some tables also have a separate `id` column, so `ID` meant two different things (issue #53).

**Decision:** The Go field for `$key` is `Key` on every type. Those thirteen fields are renamed from `ID` to `Key`. JSON tags stay `$key`. Fields that were `FlexInt` stay `FlexInt`. `USBDeviceSettings`, `TPMDeviceSettings`, and `VGPUDeviceSettings` stay `int`. `ResourceGroup` stays a UUID `string`, the same class as the other string-key tables (volumes and the other SHA1 keys, ADR-014), which already used `Key string` plus `ID` for the `id` column. `Node.ID` maps to the `id` column, not `$key`, and is not renamed. `StorageTier.Key` was already `Key` and stays `int`, because the key is the tier number.

**Rationale:**
- One name matches `$key` and the majority of the SDK.
- `ID` stays available for the separate `id` column.
- The library is still 0.x, so the rename can land before 1.0.

**Consequences:**
- Callers replace `vm.ID` with `vm.Key`, and the same for the other twelve types. Method parameters that take the key as an `int` are unchanged.
- JSON on the wire is unchanged.

## ADR-027: Update Actions Use update_actions

**Date:** 2026-09-30

**Status:** Accepted

**Context:** Settings, branches, and source packages can be read. Checking for updates, downloading them, and installing them are separate actions. pyVergeOS `UpdateSettingsManager` and the Ansible `update` module post those to `update_actions` with the source key from settings. The action that checks the source is `refresh`. `download` and `install` use those names. Applying an installed update is `all`, with `force`. The Ansible module never sends `apply`: that name has not been verified, and `all` is the action every upgrade uses. An action name outside the table's list is rejected.

**Decision:** `UpdateSettings.Check` posts `{source, action: "refresh"}`. `Download` posts `download`. `Install` posts `install`. `UpdateAll` posts `all` and always includes `force`. Each method reads settings first and returns `ValidationError` when `source` is 0. The methods return `error` only. Check, download, and install do not reboot nodes. `UpdateAll` starts the platform rolling reboot. `Get` also requests `applying_updates`. Branch and source-package reads are unchanged.

**Rationale:**
- The request body matches pyVergeOS `UpdateSettingsManager._action` and the Ansible update module.
- Install leaves the cluster running. The reboot is a separate call, so a download or install cannot imply one.
- `force` is explicit on `all` because it decides whether workloads that cannot migrate are rebooted.

**Consequences:**
- Callers read `ApplyingUpdates`, `Installed`, and `RebootRequired` themselves. These methods do not wait.
- A system with no update source configured cannot run the actions.
- The platform `install` action runs its own pre-install check. There is no method for `POST /update_actions/runpreinstallcheck`.

## ADR-028: Billing, NAS Antivirus, and Shared Objects

**Date:** 2026-09-30

**Status:** Accepted

**Context:** pyVergeOS still had three managers with no Go service: billing records, antivirus settings on a NAS service, and VMs shared with a tenant. The table names are not the manager names. Billing rows live on `billing`, and a new sample is `POST /billing_actions` with action `generate`. NAS service antivirus is `vm_service_antivirus`, the row `NASService.Antivirus` already points at. A share is a `shared_objects` row whose snapshot is `machine_snapshots/{key}`. pyVergeOS creates that snapshot with `expires_type` `never` and `created_manually` true, then posts the share. Import and refresh are `POST /shared_object_actions`.

**Decision:** `Billing`, `NASServiceAntivirus`, and `SharedObjects` are flat services on `Client`. `Billing.List` defaults to `sort=-created`. `ListCreated` turns a non-zero `since` or `until` into `created ge` and `created le`. `Generate` posts action `generate` and returns `error`. `GetSummary` averages the listed records. `NASServiceAntivirus` lists, gets, and updates `vm_service_antivirus`. `max_recursion` outside 0-100 is refused before the PUT. An update with no fields reads the row back. `SharedObjects.Create` resolves the VM `$key` to its machine key, posts the snapshot, then posts the share with `type` `vm`. A failed share deletes that snapshot. `Import` and `Refresh` post `import` and `refresh` and return `error`.

**Rationale:**
- The requests match pyVergeOS `BillingManager`, `NasServiceAntivirusManager`, and `SharedObjectManager`.
- `vm_service_antivirus` is the settings table. Volume antivirus is a different set of tables.
- The snapshot body is posted directly so `created_manually` is sent. `VMSnapshots.Create` does not send that field and defaults the expiry to a date.

**Consequences:**
- Callers read billing rows and ask VergeOS to write one. `GetSummary` does not add an endpoint.
- Antivirus on a NAS volume stays on the volume tables. This service changes the NAS service settings row.
- A share that is created and then fails its follow-up read leaves the snapshot in place. The snapshot is removed only when the `shared_objects` POST fails.

