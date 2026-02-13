---
title: Storage
description: Manage NAS services, users, volumes, snapshots, syncs, CIFS/NFS shares, and volume browsing
tags: [nas, volume, storage, cifs, nfs, share, sync, snapshot, browse, replication, backup]
categories: [Storage]
---

# Storage

Manage NAS services, volumes, shares, and data replication.

## NAS Services

NAS services are specialized VMs that provide NAS functionality including volumes, shares, and sync operations.

```go
// List all NAS services
services, err := client.NASServices.List(ctx)

// Get a NAS service by ID
service, err := client.NASServices.Get(ctx, serviceID)

// Get a NAS service by VM ID
service, err := client.NASServices.GetByVM(ctx, vmID)

// Get a NAS service by name
service, err := client.NASServices.GetByName(ctx, "nas-service-1")

// Create a NAS service (typically created automatically with NAS VM)
service, err := client.NASServices.Create(ctx, &vergeos.NASServiceCreateRequest{
    VM:         vmID,
    MaxImports: ptr(8),    // Concurrent imports (1-200)
    MaxSyncs:   ptr(4),    // Concurrent syncs (0-200, 0=disabled)
})

// Update NAS service settings
service, err := client.NASServices.Update(ctx, serviceID, &vergeos.NASServiceUpdateRequest{
    MaxImports:         ptr(16),
    MaxSyncs:           ptr(8),
    DisableSwap:        ptr(true),
    ReadAheadKBDefault: ptr("1024"), // 0, 64, 128, 256, 512, 1024, 2048, 4096 KB
})

// Delete a NAS service
err = client.NASServices.Delete(ctx, serviceID)
```

---

## NAS Service Users

Manage user accounts for NAS services to access CIFS shares. Note: Uses SHA1 hash strings as IDs.

```go
// List all NAS service users
users, err := client.NASServiceUsers.List(ctx)

// List users for a specific NAS service
users, err := client.NASServiceUsers.ListByService(ctx, serviceID)

// Get a user by ID (SHA1 hash string)
user, err := client.NASServiceUsers.Get(ctx, "abc123...")

// Get a user by name within a service
user, err := client.NASServiceUsers.GetByName(ctx, serviceID, "jsmith")

// Create a NAS service user
user, err := client.NASServiceUsers.Create(ctx, &vergeos.NASServiceUserCreateRequest{
    Service:     serviceID,
    Name:        "jsmith",
    Password:    "secure-password",
    DisplayName: "John Smith",
    Description: "Marketing department",
    HomeShare:   ptr(shareID),      // Optional CIFS share for home directory
    HomeDrive:   ptr("H"),          // Windows drive letter (A-Z)
})

// Update a user
user, err := client.NASServiceUsers.Update(ctx, userID, &vergeos.NASServiceUserUpdateRequest{
    Password:    ptr("new-password"),
    DisplayName: ptr("John Q. Smith"),
})

// Enable/disable a user
err = client.NASServiceUsers.Enable(ctx, userID)
err = client.NASServiceUsers.Disable(ctx, userID)

// Delete a user
err = client.NASServiceUsers.Delete(ctx, userID)
```

---

## Volumes

Manage NAS volumes for storage. Note: Volumes use SHA1 hash strings as IDs.

```go
// List all volumes
volumes, err := client.Volumes.List(ctx)

// List volumes for a specific NAS service
volumes, err := client.Volumes.ListByService(ctx, serviceID)

// Get a volume by ID (SHA1 hash string)
volume, err := client.Volumes.Get(ctx, "0d25c256a0c561c0b5bb9087f04fcb49f16a8048")

// Get a volume by name within a service
volume, err := client.Volumes.GetByName(ctx, serviceID, "data-vol")

// Create a volume
volume, err := client.Volumes.Create(ctx, &vergeos.VolumeCreateRequest{
    Name:    "data-vol",
    Service: serviceID,
    MaxSize: ptr(int64(100 * 1024 * 1024 * 1024)), // 100GB
})

// Update a volume
volume, err := client.Volumes.Update(ctx, volumeID, &vergeos.VolumeUpdateRequest{
    Description: ptr("Production data volume"),
})

// Delete a volume
err = client.Volumes.Delete(ctx, volumeID)

// Enable/disable/reset a volume
err = client.Volumes.Enable(ctx, volumeID)
err = client.Volumes.Disable(ctx, volumeID)
err = client.Volumes.Reset(ctx, volumeID)
```

---

## Volume Snapshots

Manage point-in-time snapshots of NAS volumes.

```go
// List all volume snapshots
snapshots, err := client.VolumeSnapshots.List(ctx)

// List snapshots for a specific volume
snapshots, err := client.VolumeSnapshots.ListByVolume(ctx, volumeID)

// List snapshots expiring within 7 days
expiring, err := client.VolumeSnapshots.ListExpiring(ctx, 7)

// List manually created snapshots (not scheduled)
manual, err := client.VolumeSnapshots.ListManual(ctx)

// Get a snapshot by ID
snapshot, err := client.VolumeSnapshots.Get(ctx, snapshotID)

// Get a snapshot by name within a volume
snapshot, err := client.VolumeSnapshots.GetByName(ctx, volumeID, "pre-migration")

// Create a volume snapshot
snapshot, err := client.VolumeSnapshots.Create(ctx, &vergeos.VolumeSnapshotCreateRequest{
    Volume:      volumeID,
    Name:        "pre-migration",
    Description: "Snapshot before data migration",
    ExpiresType: ptr("date"),    // "date" or "never"
    Expires:     ptr(time.Now().Add(7 * 24 * time.Hour).Unix()),
    Quiesce:     ptr(true),      // Freeze I/O during snapshot
})

// Update a snapshot
snapshot, err := client.VolumeSnapshots.Update(ctx, snapshotID, &vergeos.VolumeSnapshotUpdateRequest{
    Description: ptr("Updated description"),
})

// Set snapshot to never expire
snapshot, err := client.VolumeSnapshots.SetNeverExpires(ctx, snapshotID)

// Set snapshot expiration (Unix timestamp)
expires := time.Now().Add(30 * 24 * time.Hour).Unix()
snapshot, err := client.VolumeSnapshots.SetExpires(ctx, snapshotID, expires)

// Enable/disable a snapshot
err = client.VolumeSnapshots.Enable(ctx, snapshotID)
err = client.VolumeSnapshots.Disable(ctx, snapshotID)

// Delete a snapshot
err = client.VolumeSnapshots.Delete(ctx, snapshotID)
```

---

## Volume Syncs

Manage volume replication/sync jobs between volumes. Note: Uses SHA1 hash strings as IDs.

```go
// List all volume syncs
syncs, err := client.VolumeSyncs.List(ctx)

// List syncs for a specific NAS service
syncs, err := client.VolumeSyncs.ListByService(ctx, serviceID)

// List enabled syncs
syncs, err := client.VolumeSyncs.ListEnabled(ctx)

// Get a sync by ID (SHA1 hash string)
sync, err := client.VolumeSyncs.Get(ctx, "abc123...")

// Get a sync by name within a service
sync, err := client.VolumeSyncs.GetByName(ctx, serviceID, "nightly-backup")

// Create a volume sync job
sync, err := client.VolumeSyncs.Create(ctx, &vergeos.VolumeSyncCreateRequest{
    Service:           serviceID,
    Name:              "nightly-backup",
    Description:       "Nightly backup to secondary volume",
    SourceVolume:      sourceVolumeID,
    SourcePath:        ptr("/data"),
    DestinationVolume: destVolumeID,
    DestinationPath:   ptr("/backup"),

    // Sync behavior
    PreserveACLs:        ptr(true),
    PreservePermissions: ptr(true),
    PreserveOwner:       ptr(true),
    CopySymlinks:        ptr(true),

    // Delete behavior: never, delete, delete-before, delete-during, delete-delay, delete-after
    DestinationDelete: ptr("delete-after"),

    // Performance tuning
    Workers:   ptr(8),      // 1-128 workers
    ErrorsMax: ptr(int64(1000)),

    // Scheduling (use snapshot profile for automated runs)
    StartTimeProfile: ptr(profileID),
})

// Update a sync job
sync, err := client.VolumeSyncs.Update(ctx, syncID, &vergeos.VolumeSyncUpdateRequest{
    Workers:     ptr(16),
    Description: ptr("Updated description"),
})

// Enable/disable a sync job
err = client.VolumeSyncs.Enable(ctx, syncID)
err = client.VolumeSyncs.Disable(ctx, syncID)

// Start a sync job immediately
err = client.VolumeSyncs.Start(ctx, syncID)

// Stop a running sync job
err = client.VolumeSyncs.Stop(ctx, syncID)

// Delete a sync job
err = client.VolumeSyncs.Delete(ctx, syncID)
```

---

## Volume Shares

Manage CIFS (SMB) and NFS shares on volumes.

```go
// List all CIFS shares
shares, err := client.VolumeCIFSShares.List(ctx)

// Create a CIFS share
share, err := client.VolumeCIFSShares.Create(ctx, &vergeos.VolumeCIFSShareCreateRequest{
    Volume:     volumeID,
    Name:       "shared-data",
    Path:       "/data",
    Browseable: ptr(true),
    ReadOnly:   ptr(false),
})

// Create an NFS share
share, err := client.VolumeNFSShares.Create(ctx, &vergeos.VolumeNFSShareCreateRequest{
    Volume:     volumeID,
    Name:       "nfs-export",
    Path:       "/export",
    AllowedIPs: ptr("10.0.0.0/8"),
    Squash:     ptr(vergeos.NFSSquashRoot),
})

// Delete a share
err = client.VolumeCIFSShares.Delete(ctx, shareID)
```

---

## Volume Browser

Browse files within volumes asynchronously.

```go
// Simple browse (handles async polling internally)
entries, err := client.VolumeBrowser.Browse(ctx, volumeID, "/", 100)
for _, entry := range entries {
    fmt.Printf("%s %s (%d bytes)\n", entry.Type, entry.Name, entry.Size)
}

// Advanced browse with options
entries, err := client.VolumeBrowser.BrowseWithOptions(ctx, volumeID, "/data", 50,
    ptr(0),       // offset
    ".txt,.log",  // file extensions filter
)
```
