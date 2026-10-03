# Contributing to goVergeOS

Thanks for contributing. By submitting a pull request you agree to the [Contributor License Agreement](CLA.md).

## Branch model

`main` and `dev` are the long-lived branches. Keep them aligned: they should contain the same commits.

- `main` is the release branch. Version tags are created from commits that are on `main`.
- `dev` tracks the same history. It is not a holding area for commits that are missing from `main`, and `main` should not move ahead of `dev`.

Pull requests have historically been opened against both branches. Open the pull request against `main`. After it merges, merge that same result into `dev` (a fast-forward when the branches were already aligned) so the two branches match again. A change that exists on only one of them is not finished.

If the branches have drifted, reconcile them before stacking more work on the one that is behind.

## Pull requests

1. Create a branch from `main`.
2. Run the checks in the next section locally.
3. Open the pull request against `main`.
4. After it merges, update `dev` so it points at the same commit as `main`.

The `test` job from [`.github/workflows/test.yml`](.github/workflows/test.yml) is the check to require on `main` and `dev`. It runs on every pull request and every push.

## Checks

These are the same checks CI runs:

```bash
# Formatting. CI fails if this prints any files.
gofmt -l .

go vet ./...

# Pinned to the version used in CI. Needs a current Go toolchain.
go install honnef.co/go/tools/cmd/staticcheck@2026.2.1
staticcheck ./...

go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out

go build ./...
```

`go test ./...` does not run the integration suite. Those files are built with the `integration` tag.

Build each example:

```bash
while IFS= read -r dir; do
  go build -o /dev/null "$dir"
done < <(go list -f '{{if eq .Name "main"}}{{.Dir}}{{end}}' ./examples/...)
```

The library targets Go 1.21 and later, as declared in `go.mod`. CI tests that version and the current stable Go release.

## Integration tests

Integration tests talk to a live VergeOS 26.0 or later system. They are not part of the required pull request check.

```bash
export VERGEOS_HOST=https://your-vergeos-host
export VERGEOS_USERNAME=admin
export VERGEOS_PASSWORD=your-password
# or: export VERGEOS_API_KEY=your-api-key
export VERGEOS_INSECURE=true   # self-signed certificates

go run ./test/integration/sweep
go test -tags=integration -count=1 -timeout 90m ./test/integration/
go run ./test/integration/sweep -fail-if-left
```

The sweep deletes objects the suite creates: names that start with `sdk-` or `goVergeOS-test-`, plus children of those objects (a DNS view on an `sdk-` network, a cloud-init file on an `sdk-` VM, and so on). It does not delete other objects. `-fail-if-left` exits non-zero when the sweep had to delete something or when a test object is still there. Run it before the suite without the flag, and after the suite with the flag.

GitHub Actions runs the same steps from [`.github/workflows/integration.yml`](.github/workflows/integration.yml). Start it with **Run workflow**. It reads these repository secrets and does not store credentials in the repo:

| Secret | Required | Purpose |
|--------|----------|---------|
| `VERGEOS_HOST` | yes | Host or base URL |
| `VERGEOS_API_KEY` | one of these | Bearer token. Used when set, even if username and password are also set |
| `VERGEOS_USERNAME` | one of these | Basic auth, together with `VERGEOS_PASSWORD` |
| `VERGEOS_PASSWORD` | one of these | Basic auth, together with `VERGEOS_USERNAME` |
| `VERGEOS_VERIFY_SSL` | no | `true` or `false` |
| `VERGEOS_INSECURE` | no | `true` skips TLS verification |
| `VERGEOS_TIMEOUT` | no | Request timeout in seconds. The workflow uses 120 when this is unset |
| `VERGEOS_TEST_VM_ID` | no | VM `$key` for `TestVMKeyAndMachineKeyDiffer`. That test skips when unset |

A nightly run is already scheduled and stays skipped until a dedicated lab exists. Set the repository variable `VERGEOS_INTEGRATION_LAB` to `true` to let the schedule run. On-demand runs do not need that variable.

## Releases

Tag and publish from the Release workflow. The workflow creates an annotated tag and the GitHub release when someone runs it. Existing tags such as v0.3.1 are left in place; the workflow refuses a version that is already tagged.

Supported VergeOS versions, how long a deprecated method stays, and the criteria for tagging v1.0.0 are in [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md).

1. On the release commit, set `defaultUserAgent` in `client.go` (and the test that locks it) to `govergeos/X.Y.Z`, matching the tag without the leading `v`.
2. Move the `CHANGELOG.md` notes for the release out of `Unreleased` and under a heading for that version.
3. Merge the commit to `main`, then align `dev` with `main`.
4. Run **Release** (`.github/workflows/release.yml`) from the Actions tab. Pass the tag (`vX.Y.Z`) and the ref (`main`). The workflow checks that the commit is on `main`, that the user agent matches the tag, and that formatting, `go vet`, `staticcheck`, race tests, and the example builds pass. It then creates an annotated tag and the GitHub release.
5. Do not move an existing tag. The workflow refuses a version that is already tagged.

The workflow publishes the GitHub release with notes generated from pull requests. It does not copy `CHANGELOG.md` into that body. The changelog in the repository is still the record of what the tag contains. Edit the GitHub release afterward if the generated notes are not the wording you want. The v0.3.1 release body ("26.1 API sync") was written that way. The workflow leaves an existing GitHub release in place and does not replace its notes.

Pushing a `vX.Y.Z` tag yourself still publishes the GitHub release, after the same checks. Prefer the workflow dispatch so the tag is created only when those checks pass.
