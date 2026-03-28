package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// SiteService
// ---------------------------------------------------------------------------

func TestSiteService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/sites": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Site{
				{Key: 1, Name: "site-alpha", Status: "idle"},
				{Key: 2, Name: "site-beta", Status: "syncing"},
			})
		},
	}))

	sites, err := client.Sites.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sites) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(sites))
	}
	if sites[0].Name != "site-alpha" {
		t.Errorf("expected name 'site-alpha', got %q", sites[0].Name)
	}
}

func TestSiteService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Site{Key: 1, Name: "site-alpha", URL: "https://remote.example.com"})
		},
	}))

	site, err := client.Sites.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if site.Name != "site-alpha" {
		t.Errorf("expected name 'site-alpha', got %q", site.Name)
	}
	if site.URL != "https://remote.example.com" {
		t.Errorf("expected URL 'https://remote.example.com', got %q", site.URL)
	}
}

func TestSiteService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/sites/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Sites.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/sites": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Site{{Key: 1, Name: "site-alpha"}})
		},
		"GET /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Site{Key: 1, Name: "site-alpha", URL: "https://remote.example.com"})
		},
	}))

	site, err := client.Sites.GetByName(context.Background(), "site-alpha")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if site.Name != "site-alpha" {
		t.Errorf("expected name 'site-alpha', got %q", site.Name)
	}
}

func TestSiteService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/sites": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Site{})
		},
	}))

	_, err := client.Sites.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteService_GetBySiteID(t *testing.T) {
	siteID := "abcdef1234567890abcdef1234567890abcdef12"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/sites": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Site{{Key: 1, ID: siteID}})
		},
		"GET /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Site{Key: 1, ID: siteID, Name: "site-alpha"})
		},
	}))

	site, err := client.Sites.GetBySiteID(context.Background(), siteID)
	if err != nil {
		t.Fatalf("GetBySiteID failed: %v", err)
	}
	if site.ID != siteID {
		t.Errorf("expected ID %q, got %q", siteID, site.ID)
	}
}

func TestSiteService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/sites": func(w http.ResponseWriter, r *http.Request) {
			var req SiteCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.URL == "" {
				t.Error("expected url in request body")
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Site{Key: 1, Name: "new-site", URL: "https://remote.example.com"})
		},
	}))

	site, err := client.Sites.Create(context.Background(), &SiteCreateRequest{
		Name: "new-site",
		URL:  "https://remote.example.com",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if site.Name != "new-site" {
		t.Errorf("expected name 'new-site', got %q", site.Name)
	}
}

func TestSiteService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Sites.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteService_Create_MissingURL(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Sites.Create(context.Background(), &SiteCreateRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing url")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteService_Update(t *testing.T) {
	newName := "updated-site"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			var req SiteUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Site{Key: 1, Name: newName})
		},
	}))

	site, err := client.Sites.Update(context.Background(), 1, &SiteUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if site.Name != newName {
		t.Errorf("expected name %q, got %q", newName, site.Name)
	}
}

func TestSiteService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Sites.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/sites/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Sites.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSiteService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/sites/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Sites.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteService_Refresh(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "refresh" {
				t.Errorf("expected action 'refresh', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Sites.Refresh(context.Background(), 1)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

func TestSiteService_RefreshSettings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "refresh_settings" {
				t.Errorf("expected action 'refresh_settings', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Sites.RefreshSettings(context.Background(), 1)
	if err != nil {
		t.Fatalf("RefreshSettings failed: %v", err)
	}
}

func TestSiteService_Reauthenticate(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "reauthenticate" {
				t.Errorf("expected action 'reauthenticate', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Sites.Reauthenticate(context.Background(), 1)
	if err != nil {
		t.Fatalf("Reauthenticate failed: %v", err)
	}
}

func TestSiteService_RunUpdates(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "run_updates" {
				t.Errorf("expected action 'run_updates', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Sites.RunUpdates(context.Background(), 1)
	if err != nil {
		t.Fatalf("RunUpdates failed: %v", err)
	}
}

func TestSiteService_ClearSyncedLogs(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "clear_synced_logs" {
				t.Errorf("expected action 'clear_synced_logs', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Sites.ClearSyncedLogs(context.Background(), 1)
	if err != nil {
		t.Fatalf("ClearSyncedLogs failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SiteSyncIncomingService
// ---------------------------------------------------------------------------

func TestSiteSyncIncomingService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncIncoming{
				{Key: 1, Name: "inc-sync-1", Site: 10},
				{Key: 2, Name: "inc-sync-2", Site: 10},
			})
		},
	}))

	syncs, err := client.SiteSyncsIncoming.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(syncs) != 2 {
		t.Fatalf("expected 2 syncs, got %d", len(syncs))
	}
}

func TestSiteSyncIncomingService_ListBySite(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for site")
			}
			jsonResponse(w, 200, []SiteSyncIncoming{{Key: 1, Site: 10}})
		},
	}))

	syncs, err := client.SiteSyncsIncoming.ListBySite(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListBySite failed: %v", err)
	}
	if len(syncs) != 1 {
		t.Fatalf("expected 1 sync, got %d", len(syncs))
	}
}

func TestSiteSyncIncomingService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncIncoming{Key: 1, Name: "inc-sync-1", Site: 10})
		},
	}))

	sync, err := client.SiteSyncsIncoming.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sync.Name != "inc-sync-1" {
		t.Errorf("expected name 'inc-sync-1', got %q", sync.Name)
	}
}

func TestSiteSyncIncomingService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.SiteSyncsIncoming.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncIncoming{{Key: 1, Name: "inc-sync-1", Site: 10}})
		},
		"GET /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncIncoming{Key: 1, Name: "inc-sync-1", Site: 10})
		},
	}))

	sync, err := client.SiteSyncsIncoming.GetByName(context.Background(), 10, "inc-sync-1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if sync.Name != "inc-sync-1" {
		t.Errorf("expected name 'inc-sync-1', got %q", sync.Name)
	}
}

func TestSiteSyncIncomingService_GetBySyncID(t *testing.T) {
	syncID := "abcdef1234567890abcdef1234567890abcdef12"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncIncoming{{Key: 1, SyncID: syncID}})
		},
		"GET /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncIncoming{Key: 1, SyncID: syncID, Name: "inc-sync-1"})
		},
	}))

	sync, err := client.SiteSyncsIncoming.GetBySyncID(context.Background(), syncID)
	if err != nil {
		t.Fatalf("GetBySyncID failed: %v", err)
	}
	if sync.SyncID != syncID {
		t.Errorf("expected sync_id %q, got %q", syncID, sync.SyncID)
	}
}

func TestSiteSyncIncomingService_GetBySyncID_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_incoming": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncIncoming{})
		},
	}))

	_, err := client.SiteSyncsIncoming.GetBySyncID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_incoming": func(w http.ResponseWriter, r *http.Request) {
			var req SiteSyncIncomingCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Site != 10 {
				t.Errorf("expected site 10, got %d", req.Site)
			}
			if req.Name != "new-inc-sync" {
				t.Errorf("expected name 'new-inc-sync', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncIncoming{Key: 1, Name: "new-inc-sync", Site: 10})
		},
	}))

	sync, err := client.SiteSyncsIncoming.Create(context.Background(), &SiteSyncIncomingCreateRequest{
		Site: 10,
		Name: "new-inc-sync",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sync.Name != "new-inc-sync" {
		t.Errorf("expected name 'new-inc-sync', got %q", sync.Name)
	}
}

func TestSiteSyncIncomingService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsIncoming.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_Create_MissingSite(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsIncoming.Create(context.Background(), &SiteSyncIncomingCreateRequest{
		Name: "test",
	})
	if err == nil {
		t.Fatal("expected error for missing site")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsIncoming.Create(context.Background(), &SiteSyncIncomingCreateRequest{
		Site: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_Update(t *testing.T) {
	newName := "updated-inc-sync"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncIncoming{Key: 1, Name: newName})
		},
	}))

	sync, err := client.SiteSyncsIncoming.Update(context.Background(), 1, &SiteSyncIncomingUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if sync.Name != newName {
		t.Errorf("expected name %q, got %q", newName, sync.Name)
	}
}

func TestSiteSyncIncomingService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsIncoming.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/site_syncs_incoming/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsIncoming.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSiteSyncIncomingService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/site_syncs_incoming/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.SiteSyncsIncoming.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncIncomingService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_incoming_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "enable" {
				t.Errorf("expected action 'enable', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsIncoming.Enable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestSiteSyncIncomingService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_incoming_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "disable" {
				t.Errorf("expected action 'disable', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsIncoming.Disable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SiteSyncOutgoingService
// ---------------------------------------------------------------------------

func TestSiteSyncOutgoingService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncOutgoing{
				{Key: 1, Name: "out-sync-1", Site: 10},
				{Key: 2, Name: "out-sync-2", Site: 10},
			})
		},
	}))

	syncs, err := client.SiteSyncsOutgoing.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(syncs) != 2 {
		t.Fatalf("expected 2 syncs, got %d", len(syncs))
	}
}

func TestSiteSyncOutgoingService_ListBySite(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for site")
			}
			jsonResponse(w, 200, []SiteSyncOutgoing{{Key: 1, Site: 10}})
		},
	}))

	syncs, err := client.SiteSyncsOutgoing.ListBySite(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListBySite failed: %v", err)
	}
	if len(syncs) != 1 {
		t.Fatalf("expected 1 sync, got %d", len(syncs))
	}
}

func TestSiteSyncOutgoingService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncOutgoing{Key: 1, Name: "out-sync-1", Site: 10})
		},
	}))

	sync, err := client.SiteSyncsOutgoing.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sync.Name != "out-sync-1" {
		t.Errorf("expected name 'out-sync-1', got %q", sync.Name)
	}
}

func TestSiteSyncOutgoingService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.SiteSyncsOutgoing.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncOutgoing{{Key: 1, Name: "out-sync-1", Site: 10}})
		},
		"GET /api/v4/site_syncs_outgoing/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncOutgoing{Key: 1, Name: "out-sync-1", Site: 10})
		},
	}))

	sync, err := client.SiteSyncsOutgoing.GetByName(context.Background(), 10, "out-sync-1")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if sync.Name != "out-sync-1" {
		t.Errorf("expected name 'out-sync-1', got %q", sync.Name)
	}
}

func TestSiteSyncOutgoingService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncOutgoing{})
		},
	}))

	_, err := client.SiteSyncsOutgoing.GetByName(context.Background(), 10, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing": func(w http.ResponseWriter, r *http.Request) {
			var req SiteSyncOutgoingCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Site != 10 {
				t.Errorf("expected site 10, got %d", req.Site)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/site_syncs_outgoing/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncOutgoing{Key: 1, Name: "new-out-sync", Site: 10})
		},
	}))

	sync, err := client.SiteSyncsOutgoing.Create(context.Background(), &SiteSyncOutgoingCreateRequest{
		Site: 10,
		Name: "new-out-sync",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sync.Name != "new-out-sync" {
		t.Errorf("expected name 'new-out-sync', got %q", sync.Name)
	}
}

func TestSiteSyncOutgoingService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsOutgoing.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_Create_MissingSite(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsOutgoing.Create(context.Background(), &SiteSyncOutgoingCreateRequest{
		Name: "test",
	})
	if err == nil {
		t.Fatal("expected error for missing site")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsOutgoing.Create(context.Background(), &SiteSyncOutgoingCreateRequest{
		Site: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_Update(t *testing.T) {
	newName := "updated-out-sync"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/site_syncs_outgoing/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/site_syncs_outgoing/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncOutgoing{Key: 1, Name: newName})
		},
	}))

	sync, err := client.SiteSyncsOutgoing.Update(context.Background(), 1, &SiteSyncOutgoingUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if sync.Name != newName {
		t.Errorf("expected name %q, got %q", newName, sync.Name)
	}
}

func TestSiteSyncOutgoingService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncsOutgoing.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/site_syncs_outgoing/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsOutgoing.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSiteSyncOutgoingService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/site_syncs_outgoing/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.SiteSyncsOutgoing.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncOutgoingService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "enable" {
				t.Errorf("expected action 'enable', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsOutgoing.Enable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestSiteSyncOutgoingService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "disable" {
				t.Errorf("expected action 'disable', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsOutgoing.Disable(context.Background(), 1)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

func TestSiteSyncOutgoingService_Throttle(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "throttle" {
				t.Errorf("expected action 'throttle', got %v", body["action"])
			}
			params, ok := body["params"].(map[string]any)
			if !ok {
				t.Fatal("expected params in body")
			}
			if params["throttle"] != float64(1024) {
				t.Errorf("expected throttle 1024, got %v", params["throttle"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsOutgoing.Throttle(context.Background(), 1, 1024)
	if err != nil {
		t.Fatalf("Throttle failed: %v", err)
	}
}

func TestSiteSyncOutgoingService_DisableThrottle(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "throttle_disable" {
				t.Errorf("expected action 'throttle_disable', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsOutgoing.DisableThrottle(context.Background(), 1)
	if err != nil {
		t.Fatalf("DisableThrottle failed: %v", err)
	}
}

func TestSiteSyncOutgoingService_RefreshSnapshots(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "refresh" {
				t.Errorf("expected action 'refresh', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncsOutgoing.RefreshSnapshots(context.Background(), 1)
	if err != nil {
		t.Fatalf("RefreshSnapshots failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SiteSyncProfilePeriodService
// ---------------------------------------------------------------------------

func TestSiteSyncProfilePeriodService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SiteSyncProfilePeriod{
				{Key: 1, SiteSyncsOutgoing: 5, Retention: 86400},
				{Key: 2, SiteSyncsOutgoing: 5, Retention: 604800},
			})
		},
	}))

	periods, err := client.SiteSyncProfilePeriods.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 periods, got %d", len(periods))
	}
}

func TestSiteSyncProfilePeriodService_ListByOutgoingSync(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for outgoing sync")
			}
			jsonResponse(w, 200, []SiteSyncProfilePeriod{{Key: 1, SiteSyncsOutgoing: 5}})
		},
	}))

	periods, err := client.SiteSyncProfilePeriods.ListByOutgoingSync(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByOutgoingSync failed: %v", err)
	}
	if len(periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(periods))
	}
}

func TestSiteSyncProfilePeriodService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncProfilePeriod{Key: 1, Retention: 86400})
		},
	}))

	period, err := client.SiteSyncProfilePeriods.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if period.Retention != 86400 {
		t.Errorf("expected retention 86400, got %d", period.Retention)
	}
}

func TestSiteSyncProfilePeriodService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/site_syncs_outgoing_profile_periods/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.SiteSyncProfilePeriods.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSiteSyncProfilePeriodService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/site_syncs_outgoing_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			var req SiteSyncProfilePeriodCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.SiteSyncsOutgoing != 5 {
				t.Errorf("expected outgoing sync 5, got %d", req.SiteSyncsOutgoing)
			}
			if req.ProfilePeriod != 3 {
				t.Errorf("expected profile period 3, got %d", req.ProfilePeriod)
			}
			if req.Retention != 86400 {
				t.Errorf("expected retention 86400, got %d", req.Retention)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/site_syncs_outgoing_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncProfilePeriod{Key: 1, SiteSyncsOutgoing: 5, ProfilePeriod: 3, Retention: 86400})
		},
	}))

	period, err := client.SiteSyncProfilePeriods.Create(context.Background(), &SiteSyncProfilePeriodCreateRequest{
		SiteSyncsOutgoing: 5,
		ProfilePeriod:     3,
		Retention:         86400,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if period.Retention != 86400 {
		t.Errorf("expected retention 86400, got %d", period.Retention)
	}
}

func TestSiteSyncProfilePeriodService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncProfilePeriods.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncProfilePeriodService_Create_MissingOutgoingSync(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncProfilePeriods.Create(context.Background(), &SiteSyncProfilePeriodCreateRequest{
		ProfilePeriod: 3,
		Retention:     86400,
	})
	if err == nil {
		t.Fatal("expected error for missing outgoing sync")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncProfilePeriodService_Create_MissingProfilePeriod(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncProfilePeriods.Create(context.Background(), &SiteSyncProfilePeriodCreateRequest{
		SiteSyncsOutgoing: 5,
		Retention:         86400,
	})
	if err == nil {
		t.Fatal("expected error for missing profile period")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncProfilePeriodService_Create_MissingRetention(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncProfilePeriods.Create(context.Background(), &SiteSyncProfilePeriodCreateRequest{
		SiteSyncsOutgoing: 5,
		ProfilePeriod:     3,
	})
	if err == nil {
		t.Fatal("expected error for missing retention")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncProfilePeriodService_Update(t *testing.T) {
	newRetention := 172800
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/site_syncs_outgoing_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/site_syncs_outgoing_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SiteSyncProfilePeriod{Key: 1, Retention: newRetention})
		},
	}))

	period, err := client.SiteSyncProfilePeriods.Update(context.Background(), 1, &SiteSyncProfilePeriodUpdateRequest{Retention: &newRetention})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if period.Retention != newRetention {
		t.Errorf("expected retention %d, got %d", newRetention, period.Retention)
	}
}

func TestSiteSyncProfilePeriodService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SiteSyncProfilePeriods.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSiteSyncProfilePeriodService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/site_syncs_outgoing_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.SiteSyncProfilePeriods.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSiteSyncProfilePeriodService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/site_syncs_outgoing_profile_periods/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.SiteSyncProfilePeriods.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
