---
title: Compatibility and the 1.0 freeze
description: Supported VergeOS versions, deprecation, and the criteria for tagging govergeos 1.0
tags: [compatibility, versioning, release, govergeos]
categories: [Reference]
---

# Compatibility and the 1.0 freeze

goVergeOS is still pre-release. The module path is `github.com/verge-io/govergeos`, and the latest tag is v0.3.1. This document is the compatibility policy for that module: which VergeOS versions a release supports, how a new VergeOS major is handled, how long a deprecated method or field stays, and what has to be true before v1.0.0 freezes the API. v1.0.0 is not tagged yet.

The version check is ADR-016 in [DECISIONS.md](../DECISIONS.md).

## Supported VergeOS versions

Each release supports VergeOS 26.0 and every later major. `NewClient` reads `/version.json` and compares the server major with `RequiredMajorVersion` (26). A server older than that floor fails client creation with `UnsupportedVersionError`.

`WithMinimumVersion(major)` replaces the floor. `major` must be at least 1. An application raises it to require a newer VergeOS. A lower floor reaches an older server, which sits outside the range that release claims to support.

`WithSkipVersionCheck` still reads and stores the server version, and still checks credentials. It skips the major comparison, for a deployment that must keep running against a server the release would otherwise reject.

The floor for a release is the `RequiredMajorVersion` constant in `version.go`. A change to it is listed in [CHANGELOG.md](../CHANGELOG.md). Fields and actions that appeared in a later 26.x minor are requested only when the version recorded at initialization includes them, so a 26.0 system stays on the field set it understands.

## A new VergeOS major

VergeOS majors are year-based. `NewClient` accepts the next major (27, and majors after it) with the same default floor. Existing govergeos releases keep connecting when that major ships, and consumers stay on their current release until they need behavior the new major introduces.

The version string from `/version.json` is stored on the client during `NewClient`. A feature that needs a specific minor or patch checks that string with `isVersionAtLeast` before it requests behavior an older system does not have. The Microsoft 2023 Secure Boot KEK drive field is the current example: list and get include it only when the server is at least 26.1.5. Older systems are left on the previous field set.

After v1.0.0, a 1.x release keeps accepting every major that v1.0.0 accepted, and it keeps accepting later majors under the same rule. Narrowing the accepted range is a breaking change and waits for the v2 module path, described below.

## Deprecation

When an export is deprecated, its doc comment carries a `Deprecated:` line that names the replacement, and the deprecation is listed in [CHANGELOG.md](../CHANGELOG.md).

Before v1.0.0, a minor release may remove an export. The export stays for at least the minor after the one that deprecated it, and the removal is listed under Breaking. Remaining breaking removals are gathered into v0.4.0. v0.5.0 is a soak for fixes and compatible additions.

From v1.0.0 on, a deprecated method or field stays callable through every 1.x release. It is removed only on the next module path, `github.com/verge-io/govergeos/v2`. A 1.x tag keeps existing exported names and the observable behavior of existing methods. New services, fields, and options ship in minor releases. A fix that restores documented behavior ships in a patch release and is described in the changelog when a caller could have depended on the previous behavior.

Code that builds against v1.0.0 keeps building, with the same behavior, on later 1.x tags. That is the rule against silent breaks after the freeze.

## 1.0 freeze checklist

v1.0.0 is tagged when every row is true. The status column is `main` at the time this policy was added.

| # | Criterion | Status |
| --- | --- | --- |
| 1 | Naming settled. One Go name for `$key` on every type: `Key`. A separate `id` column stays `ID`. (#53) | Done. |
| 2 | Behavior settled. `PowerOff` is a graceful shutdown, `Kill` is the explicit hard power-off, and `GetByName` is an exact, unambiguous lookup with filter escaping. (#43, #45, #42) | Done. |
| 3 | Errors settled. Authentication, permission, not found, conflict, timeout, and ambiguous name each have their own type: `AuthError`, `PermissionError`, `NotFoundError`, `ConflictError`, `TimeoutError`, `AmbiguousNameError`. (#46, #45) | Done. |
| 4 | Client settled. Options compose regardless of order, retries follow the documented policy, and `WithEnvConfig` matches the Ansible collection's environment variables. (#51, #49, #50, #47) | Done. |
| 5 | Compatibility policy written down. Which VergeOS versions a release supports, how a new major is handled, and how long deprecated methods stay. (#48) | Done. This document is that policy. `NewClient` on `main` already accepts VergeOS 26 and later. |
| 6 | Process in place. CI on every pull request, tagged releases with a changelog, and an integration run against each supported VergeOS release before tagging. (#44, #41) | Largely done. The `test` workflow runs on every pull request. Tags are cut from `main` with the Release workflow, and the notes live in `CHANGELOG.md`. The current tag is v0.3.1. The integration workflow (`.github/workflows/integration.yml`) runs on demand. Its nightly schedule stays off until the `VERGEOS_INTEGRATION_LAB` repository variable is set, and the Release workflow does not require that run. Tagging v1.0.0 still includes an integration run against each supported VergeOS release before the tag is created. |

## Release cadence

1. **v0.4.0** is the breaking batch. The changes already listed under Unreleased in [CHANGELOG.md](../CHANGELOG.md), including the `Key` rename and the `PowerOff` / `Kill` split, ship in this tag. Any further breaking change before the freeze ships in this same release.
2. **v0.5.0** is one cycle for consumers to use that API and report problems. It is fixes and compatible additions.
3. **v1.0.0** is the freeze, tagged when the checklist above is complete, including the consumer bumps in the next section.

After v1.0.0, a breaking change waits for `github.com/verge-io/govergeos/v2`. The frozen module stays importable as `github.com/verge-io/govergeos`.

## Known Go consumers

These modules are the Go consumers of govergeos. The pins were recorded on 2026-09-29 from each module's `go.mod`.

| Module | Pin at planning time |
| --- | --- |
| [vergeos-exporter](https://github.com/verge-io/vergeos-exporter) | v0.3.0 |
| [csi-vergeos](https://github.com/verge-io/csi-vergeos) | v0.1.9 |
| [vergeos-cloud-controller-manager](https://github.com/verge-io/vergeos-cloud-controller-manager) | a v0.1.9 pseudo-version |
| [docker-machine-driver-vergeos](https://github.com/verge-io/docker-machine-driver-vergeos) | v0.1.3 |
| [terraform-provider-vergeio](https://github.com/verge-io/terraform-provider-vergeio) | v0.2.0 on its integration branch (PR 45 in that repository), which had not merged when this list was taken |

Bumping each module is part of the freeze release. The v0.4.0 batch (the `Key` rename, `PowerOff` / `Kill`, and the error types) is applied in all five when that tag is published. Those modules are on the frozen API when v1.0.0 is tagged. `csi-vergeos`, `vergeos-cloud-controller-manager`, and `docker-machine-driver-vergeos` were several minors behind v0.3.1 when these pins were recorded, so the bump is part of the release.
