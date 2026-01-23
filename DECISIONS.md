# Architecture Decision Records

This document captures key design decisions made during the development of the VergeOS Go SDK.

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

## ADR-011: "VergeOS Go SDK" Naming Convention

**Date:** 2026-01-21

**Status:** Accepted

**Context:** The project needed a clear identity that reflects its scope and maturity. In the Go ecosystem:
- "Library" typically implies a low-level wrapper around a single API
- "SDK" (Software Development Kit) implies a comprehensive package including client, high-level helpers, authentication logic, and potentially additional tooling

This project serves as the foundation for the Terraform Provider, Prometheus Exporter, and the upcoming Cluster API (CAPI) provider.

**Decision:** Refer to this project as the "VergeOS Go SDK" rather than "Verge OS Go library" or similar alternatives.

**Rationale:**
- Signals operational readiness and enterprise worthiness to external consumers
- Represents a significant milestone in VergeOS platform maturity
- Consolidates API client logic into a single source of truth, drastically simplifying maintenance of downstream consumers (Terraform Provider, Prometheus Exporter)
- Provides a solid foundation for the upcoming CAPI provider
- Accurately reflects scope: this is a platform for developers to build on VergeOS, not just an endpoint wrapper
- Aligns with industry conventions (AWS SDK, Google Cloud SDK, Azure SDK)

**Consequences:**
- Consistent naming across documentation, package references, and marketing materials
- Sets expectations for comprehensive functionality beyond basic API wrapping
- Establishes branding pattern for potential future SDKs (Python, JavaScript)

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
- GitHub Issue: https://github.com/verge-io/vergeos-go-sdk/issues/2
- Related: ADR-002 (FlexInt Type for ID Handling)
- Documentation: `.claude/reference/API-Schema/ENDPOINTS.md`
