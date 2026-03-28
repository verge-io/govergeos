package vergeos

import (
	"context"
	"net/http"
	"testing"
)

// UpdateSettings tests

func TestUpdateSettingsService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/update_settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, UpdateSettings{
				Key:        1,
				Source:     1,
				Branch:     2,
				BranchName: "stable-4.13",
				AutoUpdate: true,
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
