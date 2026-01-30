//go:build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestFiles tests the Files service against a live VergeOS API.
func TestFiles(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("List", func(t *testing.T) {
		files, err := client.Files.List(ctx)
		if err != nil {
			t.Fatalf("Files.List failed: %v", err)
		}
		t.Logf("Found %d file(s)", len(files))

		for i, f := range files {
			if i >= 5 {
				t.Logf("  ... and %d more", len(files)-5)
				break
			}
			sizeMB := f.Filesize / (1024 * 1024)
			t.Logf("  - %s (ID: %d, Type: %s, Size: %d MB)", f.Name, f.ID, f.Type, sizeMB)
		}
	})

	t.Run("ListISOs", func(t *testing.T) {
		isos, err := client.Files.ListISOs(ctx)
		if err != nil {
			t.Fatalf("Files.ListISOs failed: %v", err)
		}
		t.Logf("Found %d ISO file(s)", len(isos))

		for i, iso := range isos {
			if i >= 5 {
				t.Logf("  ... and %d more", len(isos)-5)
				break
			}
			t.Logf("  - %s (ID: %d)", iso.Name, iso.ID)
		}
	})

	t.Run("Get", func(t *testing.T) {
		files, err := client.Files.List(ctx, vergeos.WithLimit(1))
		if err != nil {
			t.Fatalf("Files.List failed: %v", err)
		}

		if len(files) == 0 {
			t.Skip("No files found to test Get()")
		}

		file, err := client.Files.Get(ctx, files[0].ID.Int())
		if err != nil {
			t.Fatalf("Files.Get failed: %v", err)
		}

		t.Logf("File: %s", file.Name)
		t.Logf("  ID: %d", file.ID)
		t.Logf("  Type: %s", file.Type)
		t.Logf("  Filesize: %d bytes", file.Filesize)
		t.Logf("  AllocatedBytes: %d bytes", file.AllocatedBytes)
		t.Logf("  UsedBytes: %d bytes", file.UsedBytes)
		t.Logf("  PreferredTier: %s", file.PreferredTier)
		t.Logf("  Creator: %s", file.Creator)
	})

	t.Run("GetByName", func(t *testing.T) {
		files, err := client.Files.List(ctx, vergeos.WithLimit(1))
		if err != nil {
			t.Fatalf("Files.List failed: %v", err)
		}

		if len(files) == 0 {
			t.Skip("No files found to test GetByName()")
		}

		file, err := client.Files.GetByName(ctx, files[0].Name)
		if err != nil {
			t.Fatalf("Files.GetByName failed: %v", err)
		}
		t.Logf("Found file by name: %s (ID: %d)", file.Name, file.ID)
	})
}

// TestFilesUploadDownloadDelete tests the full file lifecycle.
func TestFilesUploadDownloadDelete(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create test content
	testContent := []byte("This is a test file created by goVergeOS integration tests.\n")
	testContent = append(testContent, []byte("Generated at: "+time.Now().Format(time.RFC3339)+"\n")...)
	testFileName := fmt.Sprintf("goVergeOS-test-%d.txt", time.Now().UnixNano())

	t.Logf("Creating test file: %s (%d bytes)", testFileName, len(testContent))

	// Step 1: Create file entry
	t.Log("Step 1: Creating file entry...")
	createReq := &vergeos.FileCreateRequest{
		Name:           testFileName,
		Description:    "Integration test file - safe to delete",
		AllocatedBytes: fmt.Sprintf("%d", len(testContent)),
	}

	createdFile, err := client.Files.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Files.Create failed: %v", err)
	}
	t.Logf("  Created file entry with ID: %d", createdFile.ID)

	defer func() {
		t.Log("Cleanup: Deleting test file...")
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := client.Files.Delete(cleanupCtx, createdFile.ID.Int()); err != nil {
			t.Logf("  Warning: Failed to delete test file: %v", err)
		} else {
			t.Log("  Cleanup successful")
		}
	}()

	// Step 2: Upload content
	t.Log("Step 2: Uploading file content...")
	reader := bytes.NewReader(testContent)
	uploadedFile, err := client.Files.Upload(ctx, createdFile.ID.Int(), reader, int64(len(testContent)))
	if err != nil {
		t.Fatalf("Files.Upload failed: %v", err)
	}
	t.Logf("  Upload complete. UsedBytes: %d", uploadedFile.UsedBytes)

	// Step 3: Download and verify
	t.Log("Step 3: Downloading file...")
	downloadReader, downloadedFile, err := client.Files.Download(ctx, createdFile.ID.Int())
	if err != nil {
		t.Fatalf("Files.Download failed: %v", err)
	}
	defer downloadReader.Close()

	downloadedContent, err := io.ReadAll(downloadReader)
	if err != nil {
		t.Fatalf("Failed to read downloaded content: %v", err)
	}

	t.Logf("  Downloaded %d bytes from file: %s", len(downloadedContent), downloadedFile.Name)

	// Verify content matches
	if !bytes.Equal(testContent, downloadedContent) {
		t.Errorf("Downloaded content does not match uploaded content")
		t.Logf("  Expected: %s", string(testContent))
		t.Logf("  Got: %s", string(downloadedContent))
	} else {
		t.Log("  Content verification: PASSED")
	}

	t.Log("Step 4: Delete will be performed in cleanup...")
}

// TestFilesUploadFromFile tests uploading from a local file path.
func TestFilesUploadFromFile(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create a temporary local file
	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "test-upload.txt")
	testContent := []byte("Test content for UploadFromFile test.\n")
	testContent = append(testContent, []byte("Timestamp: "+time.Now().Format(time.RFC3339)+"\n")...)

	if err := os.WriteFile(localPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	testFileName := fmt.Sprintf("goVergeOS-upload-test-%d.txt", time.Now().UnixNano())
	t.Logf("Uploading local file %s as %s", localPath, testFileName)

	// Upload the file
	uploadedFile, err := client.Files.UploadFromFile(ctx, localPath, &vergeos.FileCreateRequest{
		Name:        testFileName,
		Description: "Integration test upload - safe to delete",
	})
	if err != nil {
		t.Fatalf("Files.UploadFromFile failed: %v", err)
	}
	t.Logf("  Upload successful. File ID: %d, Size: %d bytes", uploadedFile.ID, uploadedFile.Filesize)

	defer func() {
		t.Log("Cleanup: Deleting uploaded file...")
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if err := client.Files.Delete(cleanupCtx, uploadedFile.ID.Int()); err != nil {
			t.Logf("  Warning: Failed to delete test file: %v", err)
		} else {
			t.Log("  Cleanup successful")
		}
	}()

	// Verify by downloading
	downloadReader, _, err := client.Files.Download(ctx, uploadedFile.ID.Int())
	if err != nil {
		t.Fatalf("Failed to download uploaded file: %v", err)
	}
	defer downloadReader.Close()

	downloadedContent, err := io.ReadAll(downloadReader)
	if err != nil {
		t.Fatalf("Failed to read downloaded content: %v", err)
	}

	if !bytes.Equal(testContent, downloadedContent) {
		t.Errorf("Downloaded content does not match original")
	} else {
		t.Log("Content verification: PASSED")
	}
}

// TestFilesDownloadToFile tests downloading to a local file path.
func TestFilesDownloadToFile(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// First, list files to find one to download
	files, err := client.Files.List(ctx, vergeos.WithLimit(1))
	if err != nil {
		t.Fatalf("Files.List failed: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No files found to test DownloadToFile()")
	}

	// Only test with small files (< 10MB)
	if files[0].Filesize > 10*1024*1024 {
		t.Skipf("Skipping download test - file too large (%d bytes)", files[0].Filesize)
	}

	tmpDir := t.TempDir()

	t.Logf("Downloading file %s (ID: %d, Size: %d bytes) to %s",
		files[0].Name, files[0].ID, files[0].Filesize, tmpDir)

	downloadedPath, err := client.Files.DownloadToFile(ctx, files[0].ID.Int(), tmpDir)
	if err != nil {
		t.Fatalf("Files.DownloadToFile failed: %v", err)
	}

	t.Logf("Downloaded to: %s", downloadedPath)

	// Verify the file exists and has content
	info, err := os.Stat(downloadedPath)
	if err != nil {
		t.Fatalf("Downloaded file does not exist: %v", err)
	}

	t.Logf("Local file size: %d bytes", info.Size())

	// For files we know the size of, verify it matches
	if files[0].Filesize > 0 && info.Size() != files[0].Filesize {
		t.Errorf("Downloaded file size mismatch: expected %d, got %d", files[0].Filesize, info.Size())
	}
}
