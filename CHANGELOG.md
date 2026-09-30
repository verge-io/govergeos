# Changelog

## Unreleased

Behavior change for the next minor release.

### Changed

- `NewClient` records HTTP options and builds the client once, after every option has run. `WithHTTPClient` supplies the base. Timeout, TLS, and rate limit settings are applied to a copy, in either order, and the `*http.Client` you pass in is not modified. Skipping certificate verification requires an `*http.Transport`; any other transport type makes `NewClient` return an error instead of dropping the setting.
- `WithEnvConfig` treats a `VERGEOS_HOST` with no scheme as `https://`, and rejects any other scheme with an error that names the variable. `VERGEOS_INSECURE=true` skips TLS verification, the same as `VERGEOS_VERIFY_SSL=false`; setting both to conflicting values is an error. When an API key and username/password are both configured, the API key is used. That applies to `WithEnvConfig` and to `WithAPIKey` together with `WithCredentials`.
- `VMService.PowerOff` sends the `poweroff` action and waits for the VM to stop. It previously sent `kill`, which is a hard power-off. Callers that need the hard stop, including the docker-machine driver's force stop, should call `Kill`.

### Added

- `TenantNetworkBlocks` manages tenant CIDR blocks on `vnet_cidrs` (`List`, `ListByTenant`, `Get`, `GetByTenantAndCIDR`, `Create`, `Delete`).
- `TenantExternalIPs` gives a tenant a virtual IP on `vnet_addresses`. The tenant is the owner (`tenants/{id}`), separate from the generic `VNetAddresses` service.
- `Create` and `Delete` on both services return `ParentFirewallStatus`. `Pending` is the parent network's `need_fw_apply` flag. `WithApplyParentFirewall` applies that network's rules in the same call.
- `VMService.Kill` sends `kill` and waits for the VM to stop.
- `VMService.PowerOffWithOptions` accepts a timeout, a poll interval, and `ForceAfterTimeout`, which sends `kill` when the guest does not stop in time. A cancelled context does not escalate to `kill`.
- `WithPowerWait` sets the default timeout and poll interval for `PowerOn`, `PowerOff`, and `Kill`. The default remains 150 seconds, polled every 5 seconds. A context deadline can end the wait sooner. A longer timeout extends it.

### Fixed

- `NetworkService.GetLatestStatistics` requests a single row (`sort=-timestamp`, `limit=1`) from `/vnet_monitor_stats_history_short`. It previously called `GetStatistics`, which downloads up to 100 history rows and returns only the first. `GetStatistics` is unchanged.
