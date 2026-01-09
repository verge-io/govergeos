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
