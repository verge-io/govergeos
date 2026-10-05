# Changelog

## Unreleased

## v0.4.0 - 2026-10-05

Breaking change: the `$key` field is named `Key` on every type. Types that exposed it as `ID` are renamed below.

### Breaking

- Removed read fields that VergeOS 26 does not return. They were always the zero value.
  - `VMSnapshot.ExpiresType` and `VMSnapshot.Quiesce`. `expires_type` remains a write argument on create and update (`never` stores `Expires` as 0). `Quiesce` remains on create. The snapshot row does not return either field.
  - `SnapshotProfilePeriod.SkipMissed`, including create and update. `snapshot_profile_periods` has no `skip_missed` column, and a write was discarded.
  - `UserAPIKey.ExpiresType` on the key. Create and update still send `expires_type`. Read `Expires`; 0 means the key does not expire.
  - `Network.PowerState` and the read field `Network.BGPASN`. The live state is `Network.Running` (`machine#status#running`). `PowerOn` and `PowerOff` wait on `Running`. Create can still send `bgp_asn`.
  - `NASService.Enabled`, `NASService.Created`, and `NASService.Modified`.
  - `Group.Type`.
- The Go field for `$key` is `Key`. These types previously named that field `ID`. JSON is unchanged (`$key`).
  - `FlexInt`, renamed from `ID`: `VM`, `Network`, `VMNIC`, `VMDrive`, `VMDevice`, `Group`, `File`, `Member`, `CloudInitFile`. Use `vm.Key` where code used `vm.ID`, and the same for the other types in this list.
  - `int`, renamed from `ID`: `USBDeviceSettings`, `TPMDeviceSettings`, `VGPUDeviceSettings`.
  - `string`, renamed from `ID`: `ResourceGroup` (UUID).
- A separate `id` column is still `ID` (`User.ID`, volume SHA1 `ID`, and the same pattern on other string-key tables). `Node.ID` is that `id` column and is not renamed. Method parameters that take a key as an `int` are unchanged.

### Added

- `TenantRecipes` and `TenantRecipeInstances` mirror VM recipes for catalog tenant templates (`tenant_recipes`, `tenant_recipe_instances`). List and get recipes (including by catalog and by name), read questions from `recipe_questions`, and `Deploy` a tenant from a recipe with the same answer validation as VM recipe deploy.

- `TenantSnapshots.Create` and `TenantSnapshotCreateRequest` take a tenant snapshot from the parent (`POST /tenant_snapshots`). Fields: tenant (required), name (optional; platform may assign), profile, period, min_snapshots, description, expires, and type (`full` / `partial_include` / `partial_exclude`, default `full`). `TenantSnapshot.Type` is returned on list and get.

- `TenantCreateRequest` and `TenantUpdateRequest` accept `ui_address` (address row ID) and `ui_fqdn`, so callers can set or move the tenant UI address. VergeOS accepts both on the `tenants` table; when `ui_address` is omitted on create, the platform still defaults to the first assigned external IP.

- `VNetProxies` and `VNetProxyTenants` publish tenant UIs through one proxy on a parent network (`vnet_proxy`, `vnet_proxy_tenants`). A network has one proxy. `Create` stores `listen_address` as `0.0.0.0` and `default_self` as true when those fields are omitted, and it refuses a second proxy on that network. `GetOrCreate` returns the existing row. Tenant FQDNs on that proxy are managed by `VNetProxyTenants`.

- `Tenants.Connect` and `Tenants.ConnectByName` return a client for a running tenant's own UI. The address comes from the tenant `ui_address` row. The new client keeps the parent's TLS, timeout, retry, rate limit, recorded server version, user agent, and power wait settings. It does not keep the parent's username, password, or API key. `WithCredentials` or `WithAPIKey` is required. A snapshot, a tenant that is not running, and a tenant with no UI address are refused before the new client is created.

- Billing, NAS antivirus, and tenant shares matching pyVergeOS: `Billing` (`billing`, plus `POST /billing_actions` action `generate`), `NASServiceAntivirus` (`vm_service_antivirus`), and `SharedObjects` (`shared_objects`). Sharing a VM posts a machine snapshot with `expires_type` `never` and `created_manually` true, then the share. Import and refresh post `shared_object_actions`.
- Hardware inventory services matching pyVergeOS: `VGPUProfiles` (`nvidia_vgpu_profiles`), `NodeGPUs` (`node_gpus`, including mode update), `NodeGPUStats` (current stats plus short and long history), `NodeGPUInstances`, `NodeVGPUDevices`, `NodeHostGPUDevices`, `NodeVGPUProfiles`, `NodeMemory` (`node_memory`, with `IsHealthy`), and `NodeLLDPNeighbors` (`node_lldp_neighbors`).
- `UpdateSettings.Check`, `Download`, and `Install` post to `update_actions` for the source configured in settings. Check sends action `refresh`. Download and install use those names. None of them reboot nodes. `UpdateAll` posts action `all` with `force` and starts the platform rolling reboot. `Get` also returns `applying_updates`.
- Task engine services: `TaskSchedules` (`task_schedules`), `TaskScheduleTriggers` (`task_schedule_triggers`), `TaskEvents` (`task_events`), and `TaskScripts` (`task_scripts`). Schedule create sends the same defaults as pyVergeOS (enabled, every day, hourly, iteration 1, start of day through 86400, day of month `start_date`). `GetSchedule`, trigger, and script `Run` call the row action (`PUT /{table}/{key}?action=...`). A script created without settings sends `{"questions": []}`.
- Dynamic routing services for BGP, OSPF, and EIGRP: `VNetBGP`, `VNetBGPRouters`, `VNetBGPRouterCommands`, `VNetBGPInterfaces`, `VNetBGPInterfaceCommands`, `VNetBGPRouteMaps`, `VNetBGPRouteMapCommands`, `VNetBGPIPCommands`, `VNetOSPFCommands`, `VNetEIGRPRouters`, and `VNetEIGRPRouterCommands`. `VNetBGP.GetOrCreate` returns the per-network `vnet_bgp` row the other tables use. Create, update, and delete return `RoutingRestartStatus`. `Pending` is the network's `need_restart` flag. `WithRestartNetwork` restarts the network in the same call.

### Fixed

- Create and update no longer report a missing related row as `NotFoundError` for the row being written. VergeOS returns HTTP 404 for that lookup (a network `interface_vnet`, a VM `snapshot_profile`, and the same pattern on other writes). The error stays an `APIError` with the platform message, and `IsNotFoundError` is false, so a live object is not treated as deleted. `Get` and `Delete` of an id that is not there are still `NotFoundError`.
- `CloudInitFiles.GetContents` reads the file body from `GET /cloudinit_files/{id}?download=1`. VergeOS omits `contents` from list and get responses, including `fields=all`, so `Contents` stayed empty on every read, including the file returned by `Create` and `CreateForVM`. Assign the string to `Contents` when the file object should carry the body. This is the request pyVergeOS `get_content()` uses.
- Documentation and comments described a bare disk image (qcow2, vmdk, vhd, raw, img) as a `VMImports` source, the same call as an OVA. VergeOS 26.1 rejects those types on `vm_imports` with `Unknown Import file type`. The catalog disk-image path is `VMDrives.Create` with `Media` set to `import`.
- `Files.Upload` and `Files.UploadWithChunkSize` replay a chunk after a connection reset, a close before any response, a timeout before any response, or HTTP 429, 502, or 503. The chunk body is a `bytes.Reader`, so the retry sends the same bytes again at the same `filepos`. This is the same retry policy as other PUT calls. `WithRetry` with `MaxAttempts` of 1 turns it off.

## v0.3.1 - 2026-09-30

### Changed

- `NewClient` records HTTP options and builds the client once, after every option has run. `WithHTTPClient` supplies the base. Timeout, TLS, and rate limit settings are applied to a copy, in either order, and the `*http.Client` you pass in is not modified. Skipping certificate verification requires an `*http.Transport`; any other transport type makes `NewClient` return an error instead of dropping the setting.
- `WithEnvConfig` treats a `VERGEOS_HOST` with no scheme as `https://`, and rejects any other scheme with an error that names the variable. `VERGEOS_INSECURE=true` skips TLS verification, the same as `VERGEOS_VERIFY_SSL=false`; setting both to conflicting values is an error. When an API key and username/password are both configured, the API key is used. That applies to `WithEnvConfig` and to `WithAPIKey` together with `WithCredentials`.
- `VMService.PowerOff` sends the `poweroff` action and waits for the VM to stop. It previously sent `kill`, which is a hard power-off. Callers that need the hard stop, including the docker-machine driver's force stop, should call `Kill`.
- `NewClient` accepts VergeOS major 26 and every later major. It previously required an exact match on 26, so the next year-based release would fail client creation. Callers can raise the floor or skip the rejection with the existing options; the credential check still runs.
- `NewClient` follows the version check with one `clusters` read (`limit=1`) so a wrong password fails at construction. `/version.json` does not require authentication, so a bad credential previously surfaced only on the first real API call. The version request does not send credentials, and a failed login is not retried.
- HTTP 403 is `PermissionError` and 409 is `ConflictError`. 401 remains `AuthError`, so callers can tell a missing permission from a bad password.
- GET, PUT, and DELETE retry connection resets, timeouts before a response, and HTTP 429, 502, and 503. POST and 401 stay single-attempt.

### Added

- `Catalogs`, `VMRecipes`, and `VMRecipeInstances` deploy a VM from a recipe. List catalogs and recipes (including by name), read questions from `recipe_questions`, and `Deploy` with answer validation before POST. `Preview` is the dry run and returns a `*VMRecipePreview`; a 2xx from the platform means a VM was created (`RecipePreviewPersistedError`).
- `VMImports` creates a VM import from a media-catalog file or an http(s) URL (OVA or OVF), and can also import from a NAS volume path or a shared object. A bare disk image (qcow2, vmdk, vhd, raw, img) is rejected by `vm_imports`. Attach it with `VMDrives.Create` and `Media` set to `import`. `Wait` returns as soon as the import reports an error, an abort, or a failed drive, including `status_info` and the error log lines. `Create` does not post another import when a VM with that name already exists. `DeleteByName` removes every finished row with the name. When several rows share the name and one is still importing, it deletes none. `VMImportLogs` reads `vm_import_logs`.
- `VMExports` exports a VM to a NAS volume. `Run` ensures the volume's export configuration and starts the export. `Wait` returns as soon as the export reports an error, or when the newest statistics row recorded errors.
- `VMService.GetByName` looks up a VM by name. The `vms` table also stores snapshots, so the lookup filters `is_snapshot eq false`. Snapshot lookup stays on `VMSnapshots.GetByName`.
- `NetworkService.GetByName` looks up a network by name. Both lookups use the shared exact-name and ambiguous-name checks.
- `TenantNetworkBlocks` manages tenant CIDR blocks on `vnet_cidrs` (`List`, `ListByTenant`, `Get`, `GetByTenantAndCIDR`, `Create`, `Delete`).
- `TenantExternalIPs` gives a tenant a virtual IP on `vnet_addresses`. The tenant is the owner (`tenants/{id}`), separate from the generic `VNetAddresses` service.
- `Create` and `Delete` on both services return `ParentFirewallStatus`. `Pending` is the parent network's `need_fw_apply` flag. `WithApplyParentFirewall` applies that network's rules in the same call.
- `AuthSources` creates, reads, updates, and deletes external authentication sources. `Update` reads the stored settings and merges the caller's keys before sending, because the API replaces the whole settings object. `client_secret` is write-only: it is sent on create and update, removed from returned settings, and redacted when a value is printed.
- `OIDCApplications` creates, reads, updates, and deletes OIDC applications where VergeOS is the identity provider. `Create` returns the generated client secret as a `WriteOnlySecret`. Printing that value redacts it. The secret is not stored on `OIDCApplication`.
- `VMService.Kill` sends `kill` and waits for the VM to stop.
- `VMService.PowerOffWithOptions` accepts a timeout, a poll interval, and `ForceAfterTimeout`, which sends `kill` when the guest does not stop in time. A cancelled context does not escalate to `kill`.
- `WithPowerWait` sets the default timeout and poll interval for `PowerOn`, `PowerOff`, and `Kill`. The default remains 150 seconds, polled every 5 seconds. A context deadline can end the wait sooner. A longer timeout extends it.

### Fixed

- `NetworkService.GetLatestStatistics` requests a single row (`sort=-timestamp`, `limit=1`) from `/vnet_monitor_stats_history_short`. It previously called `GetStatistics`, which downloads up to 100 history rows and returns only the first. `GetStatistics` is unchanged.
- `GetByName` returns `AmbiguousNameError` with the matching keys when more than one row shares the name, instead of the first row and a nil error. `Tags.GetByName` takes the category, the same way snapshot-profile periods take a profile.
- Filter name literals use VergeOS backslash escapes (`\\`, `'`, `{`). Apostrophe doubling was SQL quoting and could match the wrong object or return HTTP 422. Every `GetByName` also compares the returned name with the requested name and treats a mismatch as not found.
- `CloudInitFiles.Create` requires a VM owner (`vms/<id>`). VergeOS rejects creates that omit it, and the field cannot be changed later. `CreateForVM` and `ListByVM` set that reference so callers do not pass the machine key.
- Public methods that identify a VM take the VM `$key` and resolve the machine key internally. Snapshot restore, drive and NIC delete, and VM-scoped lookups previously mixed the two keys; on a system in use those numbers differ and the call can hit another VM.
- `VMs.Snapshot` creates the row through `VMSnapshots.Create` (`machine_snapshots`), converts retention into an expires timestamp, and sends the quiesce action only when `Quiesce` is set. It previously posted `quiesce_snapshot` and returned success without inserting a snapshot row.
- Guest reboot, user enable/disable, webhook send, node maintenance, clear-pstore, volume sync, and outgoing throttle use the action names and endpoints VergeOS accepts. Those methods previously posted names or tables the API rejects.

## v0.3.0 - 2026-07-09

### Added

- `VMService.GetGuestAgentInfo` and `GuestInfo.IPsForMAC` for guest agent IP discovery.
- Network machine status and router NIC linkage on `Network`, with power and DNS actions on the live vnet machine.
- `VMDriveService.GetByName` looks up a drive by name.

### Fixed

- Cloud snapshot create sends `expires` / `expires_type` instead of an ignored retention field.
- Network `PowerOn` / `PowerOff` use `vnet_actions` and wait on machine status. `ApplyDNS` posts refresh with `target=dnsonly`.

## v0.2.0 - 2026-03-30

### Added

- `FlexFK`, ClusterTier online counts, and Node VM aggregates.
- `TenantStatus` and `TenantStatsHistoryShort` services.
- Unit tests for all 77 services.

### Fixed

- Confirmed issues from the codebase audit before the v0.2.0 cut.
