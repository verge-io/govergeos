# Changelog

## Unreleased

Behavior change for the next minor release.

### Changed

- `WithEnvConfig` treats a `VERGEOS_HOST` with no scheme as `https://`, and rejects any other scheme with an error that names the variable. `VERGEOS_INSECURE=true` skips TLS verification, the same as `VERGEOS_VERIFY_SSL=false`; setting both to conflicting values is an error. When an API key and username/password are both configured, the API key is used. That applies to `WithEnvConfig` and to `WithAPIKey` together with `WithCredentials`.
- `VMService.PowerOff` sends the `poweroff` action and waits for the VM to stop. It previously sent `kill`, which is a hard power-off. Callers that need the hard stop, including the docker-machine driver's force stop, should call `Kill`.

### Added

- `VMService.Kill` sends `kill` and waits for the VM to stop.
- `VMService.PowerOffWithOptions` accepts a timeout, a poll interval, and `ForceAfterTimeout`, which sends `kill` when the guest does not stop in time. A cancelled context does not escalate to `kill`.
- `WithPowerWait` sets the default timeout and poll interval for `PowerOn`, `PowerOff`, and `Kill`. The default remains 150 seconds, polled every 5 seconds. A context deadline can end the wait sooner. A longer timeout extends it.
