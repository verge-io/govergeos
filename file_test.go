package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFileService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []File{
				{ID: 1, Name: "ubuntu.iso", Type: "iso"},
				{ID: 2, Name: "data.img", Type: "img"},
			})
		},
	}))

	files, err := client.Files.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Name != "ubuntu.iso" {
		t.Errorf("expected name 'ubuntu.iso', got %q", files[0].Name)
	}
}

func TestFileService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter parameter")
			}
			jsonResponse(w, 200, []File{{ID: 1, Name: "ubuntu.iso", Type: "iso"}})
		},
	}))

	files, err := client.Files.List(context.Background(), WithFilter("type eq 'iso'"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestFileService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []File{})
		},
	}))

	files, err := client.Files.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestFileService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{ID: 1, Name: "ubuntu.iso", Type: "iso", Filesize: 1048576})
		},
	}))

	file, err := client.Files.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if file.Name != "ubuntu.iso" {
		t.Errorf("expected name 'ubuntu.iso', got %q", file.Name)
	}
	if file.Filesize != 1048576 {
		t.Errorf("expected filesize 1048576, got %d", file.Filesize)
	}
}

func TestFileService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Files.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestFileService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "name eq 'ubuntu.iso'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []File{{ID: 1, Name: "ubuntu.iso", Type: "iso"}})
		},
	}))

	file, err := client.Files.GetByName(context.Background(), "ubuntu.iso")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if file.Name != "ubuntu.iso" {
		t.Errorf("expected name 'ubuntu.iso', got %q", file.Name)
	}
}

func TestFileService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []File{})
		},
	}))

	_, err := client.Files.GetByName(context.Background(), "nonexistent.iso")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestFileService_ListISOs(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if !strings.Contains(filter, "type eq 'iso'") {
				t.Errorf("expected ISO filter, got %q", filter)
			}
			jsonResponse(w, 200, []File{
				{ID: 1, Name: "ubuntu.iso", Type: "iso"},
				{ID: 2, Name: "windows.iso", Type: "iso"},
			})
		},
	}))

	files, err := client.Files.ListISOs(context.Background())
	if err != nil {
		t.Fatalf("ListISOs failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 ISOs, got %d", len(files))
	}
}

func TestFileService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			var req FileCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "test.iso" {
				t.Errorf("expected name 'test.iso', got %q", req.Name)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(10)})
		},
		"GET /api/v4/files/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{ID: 10, Name: "test.iso", Type: "iso"})
		},
	}))

	file, err := client.Files.Create(context.Background(), &FileCreateRequest{
		Name:           "test.iso",
		AllocatedBytes: "1048576",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if file.Name != "test.iso" {
		t.Errorf("expected name 'test.iso', got %q", file.Name)
	}
	if int(file.ID) != 10 {
		t.Errorf("expected ID 10, got %d", int(file.ID))
	}
}

func TestFileService_Create_WithURL(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			var req FileCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.URL != "https://example.com/file.iso" {
				t.Errorf("expected URL, got %q", req.URL)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(11)})
		},
		"GET /api/v4/files/11": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{ID: 11, Name: "file.iso", URL: "https://example.com/file.iso"})
		},
	}))

	file, err := client.Files.Create(context.Background(), &FileCreateRequest{
		Name: "file.iso",
		URL:  "https://example.com/file.iso",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if file.URL != "https://example.com/file.iso" {
		t.Errorf("expected URL in response, got %q", file.URL)
	}
}

func TestFileService_Update(t *testing.T) {
	newName := "renamed.iso"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			var req FileUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{ID: 1, Name: newName, Type: "iso"})
		},
	}))

	file, err := client.Files.Update(context.Background(), 1, &FileUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if file.Name != newName {
		t.Errorf("expected name %q, got %q", newName, file.Name)
	}
}

func TestFileService_Update_NotFound(t *testing.T) {
	newName := "renamed.iso"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Files.Update(context.Background(), 999, &FileUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestFileService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Files.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestFileService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Files.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestFileService_Download(t *testing.T) {
	fileContent := "fake ISO file content here"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			// Download request has download=1 param; Get request does not
			if r.URL.Query().Get("download") == "1" {
				w.Header().Set("Content-Type", "application/octet-stream")
				w.WriteHeader(200)
				w.Write([]byte(fileContent))
				return
			}
			// Regular Get for file metadata
			jsonResponse(w, 200, File{ID: 1, Name: "test.iso", Type: "iso", Filesize: int64(len(fileContent))})
		},
	}))

	reader, file, err := client.Files.Download(context.Background(), 1)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	defer reader.Close()

	if file.Name != "test.iso" {
		t.Errorf("expected file name 'test.iso', got %q", file.Name)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read download body: %v", err)
	}
	if string(body) != fileContent {
		t.Errorf("expected body %q, got %q", fileContent, string(body))
	}
}

func TestFileService_Download_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, _, err := client.Files.Download(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestFileService_Upload(t *testing.T) {
	content := "upload test data 1234567890"
	size := int64(len(content))

	var receivedData []byte
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			receivedData = append(receivedData, body...)
			w.WriteHeader(200)
		},
		"GET /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{ID: 1, Name: "uploaded.bin", Filesize: size})
		},
	}))

	file, err := client.Files.Upload(context.Background(), 1, strings.NewReader(content), size)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if file.Name != "uploaded.bin" {
		t.Errorf("expected name 'uploaded.bin', got %q", file.Name)
	}
	if string(receivedData) != content {
		t.Errorf("expected uploaded data %q, got %q", content, string(receivedData))
	}
}

func TestFileService_UploadWithChunkSize(t *testing.T) {
	content := "abcdefghijklmnop" // 16 bytes
	size := int64(len(content))
	chunkSize := 8 // 8 bytes per chunk => 2 chunks

	var chunkCount int
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			chunkCount++
			w.WriteHeader(200)
		},
		"GET /api/v4/files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{ID: 1, Name: "chunked.bin", Filesize: size})
		},
	}))

	file, err := client.Files.UploadWithChunkSize(context.Background(), 1, strings.NewReader(content), size, chunkSize)
	if err != nil {
		t.Fatalf("UploadWithChunkSize failed: %v", err)
	}
	if file.Name != "chunked.bin" {
		t.Errorf("expected name 'chunked.bin', got %q", file.Name)
	}
	if chunkCount != 2 {
		t.Errorf("expected 2 chunks, got %d", chunkCount)
	}
}
