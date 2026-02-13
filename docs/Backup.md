---
title: Backup & DR
description: Manage snapshot profiles, disaster recovery sites, site syncs, and cloud snapshots
tags: [backup, disaster-recovery, dr, snapshot-profile, site, site-sync, cloud-snapshot, replication, schedule, retention]
categories: [Backup & DR]
---

# Backup & DR

Manage snapshot schedules, disaster recovery sites, and cloud snapshots.

## Snapshot Profiles

Manage automated snapshot schedules for VMs, volumes, and system snapshots.

```go
// List all snapshot profiles
profiles, err := client.SnapshotProfiles.List(ctx)

// Get a profile by name
profile, err := client.SnapshotProfiles.GetByName(ctx, "daily-backups")

// Create a snapshot profile
profile, err := client.SnapshotProfiles.Create(ctx, &vergeos.SnapshotProfileCreateRequest{
    Name:        "daily-backups",
    Description: "Daily backup schedule",
})

// Update a profile
profile, err := client.SnapshotProfiles.Update(ctx, profileID, &vergeos.SnapshotProfileUpdateRequest{
    Description: ptr("Updated daily backup schedule"),
})

// Delete a profile
err = client.SnapshotProfiles.Delete(ctx, profileID)
```

---

## Snapshot Profile Periods

Define scheduling periods within snapshot profiles.

```go
// List periods for a profile
periods, err := client.SnapshotProfilePeriods.ListByProfile(ctx, profileID)

// Create a daily period (snapshots at 2:00 AM, retain for 7 days)
period, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
    Profile:   profileID,
    Name:      "daily",
    Frequency: vergeos.FrequencyDaily,
    Hour:      ptr(2),
    Minute:    ptr(0),
    Retention: 7 * 24 * 60 * 60, // 7 days in seconds
})

// Create a weekly period (Sundays at midnight, retain for 4 weeks)
period, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
    Profile:   profileID,
    Name:      "weekly",
    Frequency: vergeos.FrequencyWeekly,
    DayOfWeek: ptr(vergeos.DayOfWeekSunday),
    Hour:      ptr(0),
    Minute:    ptr(0),
    Retention: 28 * 24 * 60 * 60, // 4 weeks in seconds
})

// Update a period
period, err := client.SnapshotProfilePeriods.Update(ctx, periodID, &vergeos.SnapshotProfilePeriodUpdateRequest{
    Retention: ptr(14 * 24 * 60 * 60), // Extend to 14 days
    Quiesce:   ptr(true),              // Enable quiescing
})

// Delete a period
err = client.SnapshotProfilePeriods.Delete(ctx, periodID)
```

---

## Sites & DR

Manage remote site connections for disaster recovery and data replication.

```go
// List all sites
sites, err := client.Sites.List(ctx)

// Get a site by name
site, err := client.Sites.GetByName(ctx, "dr-site")

// Create a site connection
site, err := client.Sites.Create(ctx, &vergeos.SiteCreateRequest{
    URL:          "https://dr-site.example.com",
    AuthUser:     "sync-user",
    AuthPassword: "sync-password",
})

// Refresh site connection
err = client.Sites.Refresh(ctx, siteID)

// Run updates from remote site
err = client.Sites.RunUpdates(ctx, siteID)
```

---

## Site Sync Profile Periods

Configure schedule periods for outgoing site syncs based on snapshot profile periods.

```go
// List all site sync profile periods
periods, err := client.SiteSyncProfilePeriods.List(ctx)

// List periods for a specific outgoing sync
periods, err := client.SiteSyncProfilePeriods.ListByOutgoingSync(ctx, outgoingSyncID)

// Get a period by ID
period, err := client.SiteSyncProfilePeriods.Get(ctx, periodID)

// Create a site sync profile period
period, err := client.SiteSyncProfilePeriods.Create(ctx, &vergeos.SiteSyncProfilePeriodCreateRequest{
    SiteSyncsOutgoing: outgoingSyncID,
    ProfilePeriod:     profilePeriodID, // FK to snapshot_profile_periods
    Retention:         7 * 24 * 60 * 60, // 7 days in seconds
    Priority:          ptr(10),          // Lower = higher priority
    DoNotExpire:       ptr(true),        // Don't expire source until sent
    DestinationPrefix: ptr("backup-"),   // Prefix on destination snapshots
})

// Update a period
period, err := client.SiteSyncProfilePeriods.Update(ctx, periodID, &vergeos.SiteSyncProfilePeriodUpdateRequest{
    Retention: ptr(14 * 24 * 60 * 60), // Extend to 14 days
    Priority:  ptr(5),
})

// Delete a period
err = client.SiteSyncProfilePeriods.Delete(ctx, periodID)
```

---

## Cloud Snapshots

Manage cloud-level snapshots for backup and recovery.

```go
// List all cloud snapshots
snapshots, err := client.CloudSnapshots.List(ctx)

// Create a cloud snapshot
snapshot, err := client.CloudSnapshots.Create(ctx, &vergeos.CloudSnapshotCreateRequest{
    Name:        "pre-upgrade-snapshot",
    Description: "Snapshot before system upgrade",
})

// List VMs in a snapshot
vms, err := client.CloudSnapshotVMs.ListBySnapshot(ctx, snapshotID)

// List tenants in a snapshot
tenants, err := client.CloudSnapshotTenants.ListBySnapshot(ctx, snapshotID)

// Delete a snapshot
err = client.CloudSnapshots.Delete(ctx, snapshotID)
```
