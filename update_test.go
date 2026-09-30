package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// UpdateSettings tests

func TestUpdateSettingsService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, "applying_updates") {
				t.Errorf("fields %q missing applying_updates", fields)
			}
			jsonResponse(w, 200, UpdateSettings{
				Key:             1,
				Source:          1,
				Branch:          2,
				BranchName:      "stable-4.13",
				AutoUpdate:      true,
				ApplyingUpdates: true,
			})
		},
	}))

	settings, err := client.UpdateSettings.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if settings.BranchName != "stable-4.13" {
		t.Errorf("expected branch name 'stable-4.13', got %q", settings.BranchName)
	}
	if !settings.AutoUpdate {
		t.Error("expected auto_update to be true")
	}
	if settings.Branch != 2 {
		t.Errorf("expected branch 2, got %d", settings.Branch)
	}
	if !settings.ApplyingUpdates {
		t.Error("expected applying_updates to be true")
	}
}

func TestUpdateSettingsService_Check(t *testing.T) {
	client := updateActionClient(t, 3, func(body map[string]any) {
		assertUpdateActionBody(t, body, 3, updateActionRefresh, nil)
	})

	if err := client.UpdateSettings.Check(context.Background()); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestUpdateSettingsService_Download(t *testing.T) {
	client := updateActionClient(t, 3, func(body map[string]any) {
		assertUpdateActionBody(t, body, 3, updateActionDownload, nil)
	})

	if err := client.UpdateSettings.Download(context.Background()); err != nil {
		t.Fatalf("Download failed: %v", err)
	}
}

func TestUpdateSettingsService_Install(t *testing.T) {
	client := updateActionClient(t, 3, func(body map[string]any) {
		assertUpdateActionBody(t, body, 3, updateActionInstall, nil)
	})

	if err := client.UpdateSettings.Install(context.Background()); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
}

func TestUpdateSettingsService_UpdateAll(t *testing.T) {
	force := false
	client := updateActionClient(t, 3, func(body map[string]any) {
		assertUpdateActionBody(t, body, 3, updateActionAll, &force)
	})

	if err := client.UpdateSettings.UpdateAll(context.Background(), false); err != nil {
		t.Fatalf("UpdateAll failed: %v", err)
	}
}

func TestUpdateSettingsService_UpdateAll_Force(t *testing.T) {
	force := true
	client := updateActionClient(t, 7, func(body map[string]any) {
		assertUpdateActionBody(t, body, 7, updateActionAll, &force)
	})

	if err := client.UpdateSettings.UpdateAll(context.Background(), true); err != nil {
		t.Fatalf("UpdateAll failed: %v", err)
	}
}

func TestUpdateSettingsService_Check_NoSource(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateSettings{Key: 1})
		},
		"POST /api/v4/update_actions": func(w http.ResponseWriter, r *http.Request) {
			t.Error("posted an update action with no source configured")
			w.WriteHeader(http.StatusCreated)
		},
	}))

	err := client.UpdateSettings.Check(context.Background())
	if err == nil {
		t.Fatal("expected error when no update source is configured")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateSettingsService_Check_GetError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "internal error"})
		},
		"POST /api/v4/update_actions": func(w http.ResponseWriter, r *http.Request) {
			t.Error("posted an update action after settings GET failed")
			w.WriteHeader(http.StatusCreated)
		},
	}))

	err := client.UpdateSettings.Check(context.Background())
	if err == nil {
		t.Fatal("expected error when settings GET fails")
	}
}

func TestUpdateSettingsService_Download_PostError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateSettings{Key: 1, Source: 3})
		},
		"POST /api/v4/update_actions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 422, map[string]string{"err": "value 'download' is not in list for field 'action'"})
		},
	}))

	err := client.UpdateSettings.Download(context.Background())
	if err == nil {
		t.Fatal("expected error when the action POST fails")
	}
	if !strings.Contains(err.Error(), "failed to download updates") {
		t.Errorf("expected download failure, got %v", err)
	}
}

// updateActionClient serves settings with the given source and checks the action body.
func updateActionClient(t *testing.T, source int, check func(map[string]any)) *Client {
	t.Helper()
	return newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateSettings{Key: 1, Source: source})
		},
		"POST /api/v4/update_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			check(body)
			w.WriteHeader(http.StatusCreated)
		},
	}))
}

func assertUpdateActionBody(t *testing.T, body map[string]any, source int, action string, force *bool) {
	t.Helper()
	if got, ok := body["source"].(float64); !ok || int(got) != source {
		t.Errorf("source: got %#v, want %d", body["source"], source)
	}
	if body["action"] != action {
		t.Errorf("action: got %#v, want %q", body["action"], action)
	}
	if force == nil {
		if _, ok := body["force"]; ok {
			t.Errorf("force was sent: %#v", body["force"])
		}
		if len(body) != 2 {
			t.Errorf("unexpected body: %#v", body)
		}
		return
	}
	got, ok := body["force"].(bool)
	if !ok || got != *force {
		t.Errorf("force: got %#v, want %v", body["force"], *force)
	}
	if len(body) != 3 {
		t.Errorf("unexpected body: %#v", body)
	}
}

// UpdateBranch tests

func TestUpdateBranchService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_branches": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []UpdateBranch{
				{Key: 1, Name: "stable-4.13", Description: "Stable release"},
				{Key: 2, Name: "beta-4.14", Description: "Beta release"},
			})
		},
	}))

	branches, err := client.UpdateBranches.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	if branches[0].Name != "stable-4.13" {
		t.Errorf("expected name 'stable-4.13', got %q", branches[0].Name)
	}
}

func TestUpdateBranchService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_branches/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateBranch{Key: 1, Name: "stable-4.13", Description: "Stable release"})
		},
	}))

	branch, err := client.UpdateBranches.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if branch.Name != "stable-4.13" {
		t.Errorf("expected name 'stable-4.13', got %q", branch.Name)
	}
	if branch.Description != "Stable release" {
		t.Errorf("expected description 'Stable release', got %q", branch.Description)
	}
}

func TestUpdateBranchService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_branches/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.UpdateBranches.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// UpdateSourcePackage tests

func TestUpdateSourcePackageService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_source_packages": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []UpdateSourcePackage{
				{Key: 1, Name: "ybos", Version: "4.13.1", Branch: 1, Source: 1, Downloaded: true},
				{Key: 2, Name: "ybos-ui", Version: "4.13.1", Branch: 1, Source: 1, Downloaded: false},
			})
		},
	}))

	packages, err := client.UpdateSourcePackages.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}
	if packages[0].Name != "ybos" {
		t.Errorf("expected name 'ybos', got %q", packages[0].Name)
	}
	if !packages[0].Downloaded {
		t.Error("expected first package to be downloaded")
	}
}

func TestUpdateSourcePackageService_ListByBranchAndSource(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_source_packages": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "branch eq 1 and source eq 2" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []UpdateSourcePackage{{Key: 1, Name: "ybos", Branch: 1, Source: 2}})
		},
	}))

	packages, err := client.UpdateSourcePackages.ListByBranchAndSource(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("ListByBranchAndSource failed: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
}

func TestUpdateSourcePackageService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_source_packages/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateSourcePackage{Key: 1, Name: "ybos", Version: "4.13.1", Downloaded: true})
		},
	}))

	pkg, err := client.UpdateSourcePackages.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if pkg.Name != "ybos" {
		t.Errorf("expected name 'ybos', got %q", pkg.Name)
	}
	if pkg.Version != "4.13.1" {
		t.Errorf("expected version '4.13.1', got %q", pkg.Version)
	}
}

func TestUpdateSourcePackageService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_source_packages/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.UpdateSourcePackages.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
