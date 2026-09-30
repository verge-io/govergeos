package vergeos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestCloudInitService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudInitFile{
				{ID: 1, Name: "user-data", Contents: "#cloud-config"},
				{ID: 2, Name: "meta-data", Contents: "instance-id: test"},
			})
		},
	}))

	files, err := client.CloudInitFiles.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Name != "user-data" {
		t.Errorf("expected name 'user-data', got %q", files[0].Name)
	}
}

func TestCloudInitService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter parameter")
			}
			jsonResponse(w, 200, []CloudInitFile{{ID: 1, Name: "user-data"}})
		},
	}))

	files, err := client.CloudInitFiles.List(context.Background(), WithFilter("name eq 'user-data'"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestCloudInitService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudInitFile{ID: 1, Name: "user-data", Contents: "#cloud-config"})
		},
	}))

	file, err := client.CloudInitFiles.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if file.Name != "user-data" {
		t.Errorf("expected name 'user-data', got %q", file.Name)
	}
	if file.Contents != "#cloud-config" {
		t.Errorf("expected contents '#cloud-config', got %q", file.Contents)
	}
}

func TestCloudInitService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.CloudInitFiles.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudInitService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudInitFile{{ID: 1, Name: "user-data"}})
		},
	}))

	file, err := client.CloudInitFiles.GetByName(context.Background(), "user-data")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if file.Name != "user-data" {
		t.Errorf("expected name 'user-data', got %q", file.Name)
	}
}

func TestCloudInitService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []CloudInitFile{})
		},
	}))

	_, err := client.CloudInitFiles.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudInitService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			var req CloudInitFileCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "user-data" {
				t.Errorf("expected name 'user-data', got %q", req.Name)
			}
			if req.Owner != "vms/12" {
				t.Errorf("expected owner 'vms/12', got %q", req.Owner)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/cloudinit_files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudInitFile{ID: 1, Name: "user-data", Owner: "vms/12", Contents: "#cloud-config"})
		},
	}))

	file, err := client.CloudInitFiles.Create(context.Background(), &CloudInitFileCreateRequest{
		Owner:    "vms/12",
		Name:     "user-data",
		Contents: "#cloud-config",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if file.Name != "user-data" {
		t.Errorf("expected name 'user-data', got %q", file.Name)
	}
	if file.Owner != "vms/12" {
		t.Errorf("expected owner 'vms/12', got %q", file.Owner)
	}
}

func TestCloudInitService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudInitService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.Create(context.Background(), &CloudInitFileCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "name" {
		t.Fatalf("expected name validation error, got %v", err)
	}
}

func TestCloudInitService_Create_MissingOwner(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.Create(context.Background(), &CloudInitFileCreateRequest{
		Name: "/user-data",
	})
	if err == nil {
		t.Fatal("expected error for missing owner")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "owner" {
		t.Fatalf("expected owner validation error, got %v", err)
	}
}

func TestCloudInitService_CreateForVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			var req CloudInitFileCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Owner != "vms/12" {
				t.Errorf("expected owner 'vms/12', got %q", req.Owner)
			}
			if req.Name != "/user-data" {
				t.Errorf("expected name '/user-data', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]any{"$key": 7})
		},
		"GET /api/v4/cloudinit_files/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudInitFile{ID: 7, Name: "/user-data", Owner: "vms/12"})
		},
	}))

	// A caller-supplied owner must not be posted. 99 stands in for a machine
	// key that is not this VM's $key.
	req := &CloudInitFileCreateRequest{
		Owner:    "vms/99",
		Name:     "/user-data",
		Contents: "#cloud-config\n",
	}
	file, err := client.CloudInitFiles.CreateForVM(context.Background(), 12, req)
	if err != nil {
		t.Fatalf("CreateForVM failed: %v", err)
	}
	if file.Owner != "vms/12" {
		t.Errorf("expected owner 'vms/12', got %q", file.Owner)
	}
	if req.Owner != "vms/99" {
		t.Errorf("CreateForVM mutated request owner to %q", req.Owner)
	}
}

func TestCloudInitService_CreateForVM_InvalidVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.CreateForVM(context.Background(), 0, &CloudInitFileCreateRequest{
		Name: "/user-data",
	})
	if err == nil {
		t.Fatal("expected error for missing VM ID")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "vm" {
		t.Fatalf("expected vm validation error, got %v", err)
	}
}

func TestCloudInitService_CreateForVM_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.CreateForVM(context.Background(), 12, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudInitService_ListByVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "owner eq 'vms/12'" {
				t.Errorf("expected owner filter, got %q", filter)
			}
			jsonResponse(w, 200, []CloudInitFile{{ID: 7, Name: "/user-data", Owner: "vms/12"}})
		},
	}))

	files, err := client.CloudInitFiles.ListByVM(context.Background(), 12)
	if err != nil {
		t.Fatalf("ListByVM failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Owner != "vms/12" {
		t.Errorf("expected owner 'vms/12', got %q", files[0].Owner)
	}
}

func TestCloudInitService_ListByVM_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/cloudinit_files": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			expected := "owner eq 'vms/12' and name eq '/user-data'"
			if filter != expected {
				t.Errorf("expected filter %q, got %q", expected, filter)
			}
			jsonResponse(w, 200, []CloudInitFile{})
		},
	}))

	_, err := client.CloudInitFiles.ListByVM(context.Background(), 12, WithFilter("name eq '/user-data'"))
	if err != nil {
		t.Fatalf("ListByVM with filter failed: %v", err)
	}
}

func TestCloudInitService_ListByVM_InvalidVM(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.ListByVM(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for missing VM ID")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "vm" {
		t.Fatalf("expected vm validation error, got %v", err)
	}
}

func TestCloudInitService_Update(t *testing.T) {
	newName := "meta-data"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/cloudinit_files/1": func(w http.ResponseWriter, r *http.Request) {
			var req CloudInitFileUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/cloudinit_files/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, CloudInitFile{ID: 1, Name: newName})
		},
	}))

	file, err := client.CloudInitFiles.Update(context.Background(), 1, &CloudInitFileUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if file.Name != newName {
		t.Errorf("expected name %q, got %q", newName, file.Name)
	}
}

func TestCloudInitService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.CloudInitFiles.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCloudInitService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/cloudinit_files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "updated"
	_, err := client.CloudInitFiles.Update(context.Background(), 999, &CloudInitFileUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCloudInitService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/cloudinit_files/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.CloudInitFiles.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestCloudInitService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/cloudinit_files/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete returns nil for not found (already deleted)
	err := client.CloudInitFiles.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted cloud-init")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
