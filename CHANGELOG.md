# Changelog

## Unreleased

Breaking change: the `$key` field is named `Key` on every type. Types that exposed it as `ID` are renamed below.

Behavior change: `VMService.PowerOff` is a graceful shutdown. See Changed.

### Breaking

- The Go field for `$key` is `Key`. These types previously named that field `ID`. JSON is unchanged (`$key`).
  - `FlexInt`, renamed from `ID`: `VM`, `Network`, `VMNIC`, `VMDrive`, `VMDevice`, `Group`, `File`, `Member`, `CloudInitFile`. Use `vm.Key` where code used `vm.ID`, and the same for the other types in this list.
  - `int`, renamed from `ID`: `USBDeviceSettings`, `TPMDeviceSettings`, `VGPUDeviceSettings`.
  - `string`, renamed from `ID`: `ResourceGroup` (UUID).
- A separate `id` column is still `ID` (`User.ID`, volume SHA1 `ID`, and the same pattern on other string-key tables). `Node.ID` is that `id` column and is not renamed. Method parameters that take a key as an `int` are unchanged.

### Changed

- `NewClient` records HTTP options and builds the client once, after every option has run. `WithHTTPClient` supplies the base. Timeout, TLS, and rate limit settings are applied to a copy, in either order, and the `*http.Client` you pass in is not modified. Skipping certificate verification requires an `*http.Transport`; any other transport type makes `NewClient` return an error instead of dropping the setting.
- `WithEnvConfig` treats a `VERGEOS_HOST` with no scheme as `https://`, and rejects any other scheme with an error that names the variable. `VERGEOS_INSECURE=true` skips TLS verification, the same as `VERGEOS_VERIFY_SSL=false`; setting both to conflicting values is an error. When an API key and username/password are both configured, the API key is used. That applies to `WithEnvConfig` and to `WithAPIKey` together with `WithCredentials`.
- `VMService.PowerOff` sends the `poweroff` action and waits for the VM to stop. It previously sent `kill`, which is a hard power-off. Callers that need the hard stop, including the docker-machine driver's force stop, should call `Kill`.

### Added

- Billing, NAS antivirus, and tenant shares matching pyVergeOS: `Billing` (`billing`, plus `POST /billing_actions` action `generate`), `NASServiceAntivirus` (`vm_service_antivirus`), and `SharedObjects` (`shared_objects`). Sharing a VM posts a machine snapshot with `expires_type` `never` and `created_manually` true, then the share. Import and refresh post `shared_object_actions`.
- Hardware inventory services matching pyVergeOS: `VGPUProfiles` (`nvidia_vgpu_profiles`), `NodeGPUs` (`node_gpus`, including mode update), `NodeGPUStats` (current stats plus short and long history), `NodeGPUInstances`, `NodeVGPUDevices`, `NodeHostGPUDevices`, `NodeVGPUProfiles`, `NodeMemory` (`node_memory`, with `IsHealthy`), and `NodeLLDPNeighbors` (`node_lldp_neighbors`).
- `UpdateSettings.Check`, `Download`, and `Install` post to `update_actions` for the source configured in settings. Check sends action `refresh`. Download and install use those names. None of them reboot nodes. `UpdateAll` posts action `all` with `force` and starts the platform rolling reboot. `Get` also returns `applying_updates`.
- Task engine services: `TaskSchedules` (`task_schedules`), `TaskScheduleTriggers` (`task_schedule_triggers`), `TaskEvents` (`task_events`), and `TaskScripts` (`task_scripts`). Schedule create sends the same defaults as pyVergeOS (enabled, every day, hourly, iteration 1, start of day through 86400, day of month `start_date`). `GetSchedule`, trigger, and script `Run` call the row action (`PUT /{table}/{key}?action=...`). A script created without settings sends `{"questions": []}`.
- `VMImports` creates a VM import from a media-catalog file or an http(s) URL (OVA, OVF, or a disk image such as qcow2, vmdk, vhd, or raw), and can also import from a NAS volume path or a shared object. `Wait` returns as soon as the import reports an error, an abort, or a failed drive, including `status_info` and the error log lines. `Create` does not post another import when a VM with that name already exists. `DeleteByName` removes every finished row with the name. When several rows share the name and one is still importing, it deletes none. `VMImportLogs` reads `vm_import_logs`.
- `VMExports` exports a VM to a NAS volume. `Run` ensures the volume's export configuration and starts the export. `Wait` returns as soon as the export reports an error, or when the newest statistics row recorded errors.
- `VMService.GetByName` looks up a VM by name. The `vms` table also stores snapshots, so the lookup filters `is_snapshot eq false`. Snapshot lookup stays on `VMSnapshots.GetByName`.
- `NetworkService.GetByName` looks up a network by name. Both lookups use the shared exact-name and ambiguous-name checks.
- `TenantNetworkBlocks` manages tenant CIDR blocks on `vnet_cidrs` (`List`, `ListByTenant`, `Get`, `GetByTenantAndCIDR`, `Create`, `Delete`).
- `TenantExternalIPs` gives a tenant a virtual IP on `vnet_addresses`. The tenant is the owner (`tenants/{id}`), separate from the generic `VNetAddresses` service.
- `Create` and `Delete` on both services return `ParentFirewallStatus`. `Pending` is the parent network's `need_fw_apply` flag. `WithApplyParentFirewall` applies that network's rules in the same call.
- Dynamic routing services for BGP, OSPF, and EIGRP: `VNetBGP`, `VNetBGPRouters`, `VNetBGPRouterCommands`, `VNetBGPInterfaces`, `VNetBGPInterfaceCommands`, `VNetBGPRouteMaps`, `VNetBGPRouteMapCommands`, `VNetBGPIPCommands`, `VNetOSPFCommands`, `VNetEIGRPRouters`, and `VNetEIGRPRouterCommands`. `VNetBGP.GetOrCreate` returns the per-network `vnet_bgp` row the other tables use. Create, update, and delete return `RoutingRestartStatus`. `Pending` is the network's `need_restart` flag. `WithRestartNetwork` restarts the network in the same call.
- `AuthSources` creates, reads, updates, and deletes external authentication sources. `Update` reads the stored settings and merges the caller's keys before sending, because the API replaces the whole settings object. `client_secret` is write-only: it is sent on create and update, removed from returned settings, and redacted when a value is printed.
- `OIDCApplications` creates, reads, updates, and deletes OIDC applications where VergeOS is the identity provider. `Create` returns the generated client secret as a `WriteOnlySecret`. Printing that value redacts it. The secret is not stored on `OIDCApplication`.
- `VMService.Kill` sends `kill` and waits for the VM to stop.
- `VMService.PowerOffWithOptions` accepts a timeout, a poll interval, and `ForceAfterTimeout`, which sends `kill` when the guest does not stop in time. A cancelled context does not escalate to `kill`.
- `WithPowerWait` sets the default timeout and poll interval for `PowerOn`, `PowerOff`, and `Kill`. The default remains 150 seconds, polled every 5 seconds. A context deadline can end the wait sooner. A longer timeout extends it.

### Fixed

- `NetworkService.GetLatestStatistics` requests a single row (`sort=-timestamp`, `limit=1`) from `/vnet_monitor_stats_history_short`. It previously called `GetStatistics`, which downloads up to 100 history rows and returns only the first. `GetStatistics` is unchanged.
