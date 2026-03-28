package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// --- CIFS Share Tests ---

func TestVolumeCIFSShareService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_cifs_shares": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeCIFSShare{
				{Key: "cifs1", ID: "cifs1", Name: "share1", Enabled: true},
				{Key: "cifs2", ID: "cifs2", Name: "share2", Enabled: false},
			})
		},
	}))

	shares, err := client.VolumeCIFSShares.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(shares) != 2 {
		t.Fatalf("expected 2 shares, got %d", len(shares))
	}
	if shares[0].Name != "share1" {
		t.Errorf("expected name 'share1', got %q", shares[0].Name)
	}
}

func TestVolumeCIFSShareService_ListByVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_cifs_shares": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "volume eq 'volhash'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeCIFSShare{{Key: "cifs1", ID: "cifs1", Name: "share1"}})
		},
	}))

	shares, err := client.VolumeCIFSShares.ListByVolume(context.Background(), "volhash")
	if err != nil {
		t.Fatalf("ListByVolume failed: %v", err)
	}
	if len(shares) != 1 {
		t.Fatalf("expected 1 share, got %d", len(shares))
	}
}

func TestVolumeCIFSShareService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_cifs_shares/cifs1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeCIFSShare{Key: "cifs1", ID: "cifs1", Name: "share1", Browseable: true})
		},
	}))

	share, err := client.VolumeCIFSShares.Get(context.Background(), "cifs1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if share.Name != "share1" {
		t.Errorf("expected name 'share1', got %q", share.Name)
	}
	if !share.Browseable {
		t.Error("expected browseable to be true")
	}
}

func TestVolumeCIFSShareService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_cifs_shares/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeCIFSShares.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_cifs_shares": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeCIFSShare{{Key: "cifs1", ID: "cifs1", Name: "share1"}})
		},
		"GET /api/v4/volume_cifs_shares/cifs1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeCIFSShare{Key: "cifs1", ID: "cifs1", Name: "share1"})
		},
	}))

	share, err := client.VolumeCIFSShares.GetByName(context.Background(), "volhash", "share1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if share.Name != "share1" {
		t.Errorf("expected name 'share1', got %q", share.Name)
	}
}

func TestVolumeCIFSShareService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_cifs_shares": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeCIFSShare{})
		},
	}))

	_, err := client.VolumeCIFSShares.GetByName(context.Background(), "volhash", "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_cifs_shares": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeCIFSShareCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "myshare" {
				t.Errorf("expected name 'myshare', got %q", req.Name)
			}
			if req.Volume != "volhash" {
				t.Errorf("expected volume 'volhash', got %q", req.Volume)
			}
			jsonResponse(w, 200, apiResponse{Key: "newcifs"})
		},
		"GET /api/v4/volume_cifs_shares/newcifs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeCIFSShare{Key: "newcifs", ID: "newcifs", Name: "myshare", Enabled: true})
		},
	}))

	share, err := client.VolumeCIFSShares.Create(context.Background(), &VolumeCIFSShareCreateRequest{
		Name:   "myshare",
		Volume: "volhash",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if share.Name != "myshare" {
		t.Errorf("expected name 'myshare', got %q", share.Name)
	}
}

func TestVolumeCIFSShareService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeCIFSShares.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeCIFSShares.Create(context.Background(), &VolumeCIFSShareCreateRequest{Volume: "volhash"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_Create_MissingVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeCIFSShares.Create(context.Background(), &VolumeCIFSShareCreateRequest{Name: "share1"})
	if err == nil {
		t.Fatal("expected error for missing volume")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_Update(t *testing.T) {
	newComment := "updated comment"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_cifs_shares/cifs1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_cifs_shares/cifs1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeCIFSShare{Key: "cifs1", ID: "cifs1", Name: "share1", Comment: newComment})
		},
	}))

	share, err := client.VolumeCIFSShares.Update(context.Background(), "cifs1", &VolumeCIFSShareUpdateRequest{Comment: &newComment})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if share.Comment != newComment {
		t.Errorf("expected comment %q, got %q", newComment, share.Comment)
	}
}

func TestVolumeCIFSShareService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeCIFSShares.Update(context.Background(), "cifs1", nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_Update_NotFound(t *testing.T) {
	name := "test"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_cifs_shares/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeCIFSShares.Update(context.Background(), "missing", &VolumeCIFSShareUpdateRequest{Name: &name})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeCIFSShareService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_cifs_shares/cifs1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VolumeCIFSShares.Delete(context.Background(), "cifs1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVolumeCIFSShareService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_cifs_shares/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VolumeCIFSShares.Delete(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- NFS Share Tests ---

func TestVolumeNFSShareService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_nfs_shares": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeNFSShare{
				{Key: "nfs1", ID: "nfs1", Name: "export1", Enabled: true},
				{Key: "nfs2", ID: "nfs2", Name: "export2", Enabled: false},
			})
		},
	}))

	shares, err := client.VolumeNFSShares.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(shares) != 2 {
		t.Fatalf("expected 2 shares, got %d", len(shares))
	}
	if shares[0].Name != "export1" {
		t.Errorf("expected name 'export1', got %q", shares[0].Name)
	}
}

func TestVolumeNFSShareService_ListByVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_nfs_shares": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "volume eq 'volhash'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VolumeNFSShare{{Key: "nfs1", ID: "nfs1", Name: "export1"}})
		},
	}))

	shares, err := client.VolumeNFSShares.ListByVolume(context.Background(), "volhash")
	if err != nil {
		t.Fatalf("ListByVolume failed: %v", err)
	}
	if len(shares) != 1 {
		t.Fatalf("expected 1 share, got %d", len(shares))
	}
}

func TestVolumeNFSShareService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_nfs_shares/nfs1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeNFSShare{Key: "nfs1", ID: "nfs1", Name: "export1", Squash: "root_squash"})
		},
	}))

	share, err := client.VolumeNFSShares.Get(context.Background(), "nfs1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if share.Name != "export1" {
		t.Errorf("expected name 'export1', got %q", share.Name)
	}
	if share.Squash != "root_squash" {
		t.Errorf("expected squash 'root_squash', got %q", share.Squash)
	}
}

func TestVolumeNFSShareService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_nfs_shares/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeNFSShares.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_nfs_shares": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeNFSShare{{Key: "nfs1", ID: "nfs1", Name: "export1"}})
		},
		"GET /api/v4/volume_nfs_shares/nfs1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeNFSShare{Key: "nfs1", ID: "nfs1", Name: "export1"})
		},
	}))

	share, err := client.VolumeNFSShares.GetByName(context.Background(), "volhash", "export1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if share.Name != "export1" {
		t.Errorf("expected name 'export1', got %q", share.Name)
	}
}

func TestVolumeNFSShareService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_nfs_shares": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VolumeNFSShare{})
		},
	}))

	_, err := client.VolumeNFSShares.GetByName(context.Background(), "volhash", "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/volume_nfs_shares": func(w http.ResponseWriter, r *http.Request) {
			var req VolumeNFSShareCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "export1" {
				t.Errorf("expected name 'export1', got %q", req.Name)
			}
			if req.Volume != "volhash" {
				t.Errorf("expected volume 'volhash', got %q", req.Volume)
			}
			jsonResponse(w, 200, apiResponse{Key: "newnfs"})
		},
		"GET /api/v4/volume_nfs_shares/newnfs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeNFSShare{Key: "newnfs", ID: "newnfs", Name: "export1", Enabled: true})
		},
	}))

	share, err := client.VolumeNFSShares.Create(context.Background(), &VolumeNFSShareCreateRequest{
		Name:   "export1",
		Volume: "volhash",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if share.Name != "export1" {
		t.Errorf("expected name 'export1', got %q", share.Name)
	}
}

func TestVolumeNFSShareService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeNFSShares.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeNFSShares.Create(context.Background(), &VolumeNFSShareCreateRequest{Volume: "volhash"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_Create_MissingVolume(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeNFSShares.Create(context.Background(), &VolumeNFSShareCreateRequest{Name: "export1"})
	if err == nil {
		t.Fatal("expected error for missing volume")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_Update(t *testing.T) {
	squash := "all_squash"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_nfs_shares/nfs1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/volume_nfs_shares/nfs1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VolumeNFSShare{Key: "nfs1", ID: "nfs1", Name: "export1", Squash: squash})
		},
	}))

	share, err := client.VolumeNFSShares.Update(context.Background(), "nfs1", &VolumeNFSShareUpdateRequest{Squash: &squash})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if share.Squash != squash {
		t.Errorf("expected squash %q, got %q", squash, share.Squash)
	}
}

func TestVolumeNFSShareService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VolumeNFSShares.Update(context.Background(), "nfs1", nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_Update_NotFound(t *testing.T) {
	name := "test"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/volume_nfs_shares/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VolumeNFSShares.Update(context.Background(), "missing", &VolumeNFSShareUpdateRequest{Name: &name})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVolumeNFSShareService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_nfs_shares/nfs1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VolumeNFSShares.Delete(context.Background(), "nfs1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVolumeNFSShareService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/volume_nfs_shares/missing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VolumeNFSShares.Delete(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
