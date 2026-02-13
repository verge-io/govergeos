---
title: Files
description: Upload, download, and manage ISO images, disk images, and media files
tags: [file, iso, upload, download, media, qcow2, vmdk, vhd, ova, image, chunk]
categories: [Files]
---

# Files

Upload, download, and manage ISO images, disk images, and other media files stored in VergeOS.

```go
// List all files
files, err := client.Files.List(ctx)

// List ISO files only
isos, err := client.Files.ListISOs(ctx)

// Get a specific file
file, err := client.Files.Get(ctx, fileID)

// Get a file by name
file, err := client.Files.GetByName(ctx, "ubuntu-24.04.iso")

// Create a file entry (for upload or URL import)
file, err := client.Files.Create(ctx, &vergeos.FileCreateRequest{
    Name:           "my-image.iso",
    Description:    "Ubuntu Server ISO",
    AllocatedBytes: "1073741824", // 1GB - must match actual file size
    PreferredTier:  "1",
})

// Upload a local file (creates entry and uploads content)
file, err := client.Files.UploadFromFile(ctx, "/path/to/local.iso", &vergeos.FileCreateRequest{
    Name:        "uploaded.iso",
    Description: "Uploaded via SDK",
})

// Upload to existing file entry with custom chunk size
file, err := client.Files.UploadWithChunkSize(ctx, fileID, reader, fileSize, 524288) // 512KB chunks

// Download a file (returns io.ReadCloser)
reader, file, err := client.Files.Download(ctx, fileID)
if err != nil {
    return err
}
defer reader.Close()
// Use io.Copy to write to destination
io.Copy(destination, reader)

// Download to a local file path
localPath, err := client.Files.DownloadToFile(ctx, fileID, "/local/download/dir/")

// Update file metadata
newDesc := "Updated description"
file, err := client.Files.Update(ctx, fileID, &vergeos.FileUpdateRequest{
    Description: &newDesc,
})

// Delete a file
err = client.Files.Delete(ctx, fileID)
```

## Upload Notes

- Files are uploaded in 256KB chunks by default (matching verge-cli and PSVergeOS)
- The `AllocatedBytes` in `FileCreateRequest` must match the actual file size
- Use `UploadFromFile` for the simplest upload experience - it handles entry creation and chunked upload
- Use `Upload` or `UploadWithChunkSize` when you need fine-grained control over the upload process

## Supported File Types

| Type | Description |
|------|-------------|
| `iso` | ISO images for VM boot |
| `img` | Raw disk images |
| `qcow2` | QEMU/KVM disk format |
| `vmdk` | VMware disk format |
| `vhd`/`vhdx` | Hyper-V disk formats |
| `ova`/`ovf` | VMware/VirtualBox exports |
| `raw` | Raw binary disk images |
