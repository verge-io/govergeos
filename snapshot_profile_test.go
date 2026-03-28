package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// SnapshotProfile tests

func TestSnapshotProfileService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profiles": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SnapshotProfile{
				{Key: 1, Name: "default"},
				{Key: 2, Name: "custom"},
			})
		},
	}))

	profiles, err := client.SnapshotProfiles.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Name != "default" {
		t.Errorf("expected name 'default', got %q", profiles[0].Name)
	}
}

func TestSnapshotProfileService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profiles/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfile{Key: 1, Name: "default", Description: "Default profile"})
		},
	}))

	profile, err := client.SnapshotProfiles.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if profile.Name != "default" {
		t.Errorf("expected name 'default', got %q", profile.Name)
	}
	if profile.Description != "Default profile" {
		t.Errorf("expected description 'Default profile', got %q", profile.Description)
	}
}

func TestSnapshotProfileService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profiles/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.SnapshotProfiles.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSnapshotProfileService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profiles": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SnapshotProfile{{Key: 1, Name: "default"}})
		},
		"GET /api/v4/snapshot_profiles/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfile{Key: 1, Name: "default", Description: "Default profile"})
		},
	}))

	profile, err := client.SnapshotProfiles.GetByName(context.Background(), "default")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if profile.Name != "default" {
		t.Errorf("expected name 'default', got %q", profile.Name)
	}
}

func TestSnapshotProfileService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profiles": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SnapshotProfile{})
		},
	}))

	_, err := client.SnapshotProfiles.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSnapshotProfileService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/snapshot_profiles": func(w http.ResponseWriter, r *http.Request) {
			var req SnapshotProfileCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "weekly" {
				t.Errorf("expected name 'weekly', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 3})
		},
		"GET /api/v4/snapshot_profiles/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfile{Key: 3, Name: "weekly"})
		},
	}))

	profile, err := client.SnapshotProfiles.Create(context.Background(), &SnapshotProfileCreateRequest{
		Name: "weekly",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if profile.Name != "weekly" {
		t.Errorf("expected name 'weekly', got %q", profile.Name)
	}
}

func TestSnapshotProfileService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfiles.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfileService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfiles.Create(context.Background(), &SnapshotProfileCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfileService_Update(t *testing.T) {
	newName := "monthly"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/snapshot_profiles/1": func(w http.ResponseWriter, r *http.Request) {
			var req SnapshotProfileUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/snapshot_profiles/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfile{Key: 1, Name: newName})
		},
	}))

	profile, err := client.SnapshotProfiles.Update(context.Background(), 1, &SnapshotProfileUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if profile.Name != newName {
		t.Errorf("expected name %q, got %q", newName, profile.Name)
	}
}

func TestSnapshotProfileService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfiles.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfileService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/snapshot_profiles/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "updated"
	_, err := client.SnapshotProfiles.Update(context.Background(), 999, &SnapshotProfileUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSnapshotProfileService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/snapshot_profiles/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.SnapshotProfiles.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSnapshotProfileService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/snapshot_profiles/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.SnapshotProfiles.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// SnapshotProfilePeriod tests

func TestSnapshotProfilePeriodService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SnapshotProfilePeriod{
				{Key: 1, Name: "hourly", Frequency: "hourly", Retention: 86400},
				{Key: 2, Name: "daily", Frequency: "daily", Retention: 604800},
			})
		},
	}))

	periods, err := client.SnapshotProfilePeriods.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 periods, got %d", len(periods))
	}
	if periods[0].Name != "hourly" {
		t.Errorf("expected name 'hourly', got %q", periods[0].Name)
	}
}

func TestSnapshotProfilePeriodService_ListByProfile(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "profile eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []SnapshotProfilePeriod{{Key: 1, Profile: 5, Name: "hourly"}})
		},
	}))

	periods, err := client.SnapshotProfilePeriods.ListByProfile(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByProfile failed: %v", err)
	}
	if len(periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(periods))
	}
}

func TestSnapshotProfilePeriodService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfilePeriod{Key: 1, Name: "hourly", Frequency: "hourly", Retention: 86400})
		},
	}))

	period, err := client.SnapshotProfilePeriods.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if period.Name != "hourly" {
		t.Errorf("expected name 'hourly', got %q", period.Name)
	}
	if period.Retention != 86400 {
		t.Errorf("expected retention 86400, got %d", period.Retention)
	}
}

func TestSnapshotProfilePeriodService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profile_periods/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.SnapshotProfilePeriods.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SnapshotProfilePeriod{{Key: 1, Profile: 5, Name: "hourly"}})
		},
		"GET /api/v4/snapshot_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfilePeriod{Key: 1, Profile: 5, Name: "hourly", Frequency: "hourly"})
		},
	}))

	period, err := client.SnapshotProfilePeriods.GetByName(context.Background(), 5, "hourly")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if period.Name != "hourly" {
		t.Errorf("expected name 'hourly', got %q", period.Name)
	}
}

func TestSnapshotProfilePeriodService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/snapshot_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SnapshotProfilePeriod{})
		},
	}))

	_, err := client.SnapshotProfilePeriods.GetByName(context.Background(), 5, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/snapshot_profile_periods": func(w http.ResponseWriter, r *http.Request) {
			var req SnapshotProfilePeriodCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "hourly" {
				t.Errorf("expected name 'hourly', got %q", req.Name)
			}
			if req.Profile != 5 {
				t.Errorf("expected profile 5, got %d", req.Profile)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 10})
		},
		"GET /api/v4/snapshot_profile_periods/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfilePeriod{Key: 10, Profile: 5, Name: "hourly", Frequency: "hourly", Retention: 86400})
		},
	}))

	period, err := client.SnapshotProfilePeriods.Create(context.Background(), &SnapshotProfilePeriodCreateRequest{
		Profile:   5,
		Name:      "hourly",
		Frequency: "hourly",
		Retention: 86400,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if period.Name != "hourly" {
		t.Errorf("expected name 'hourly', got %q", period.Name)
	}
}

func TestSnapshotProfilePeriodService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfilePeriods.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Create_MissingProfile(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfilePeriods.Create(context.Background(), &SnapshotProfilePeriodCreateRequest{
		Name:      "hourly",
		Retention: 86400,
	})
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfilePeriods.Create(context.Background(), &SnapshotProfilePeriodCreateRequest{
		Profile:   5,
		Retention: 86400,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Create_MissingRetention(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfilePeriods.Create(context.Background(), &SnapshotProfilePeriodCreateRequest{
		Profile: 5,
		Name:    "hourly",
	})
	if err == nil {
		t.Fatal("expected error for missing retention")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Update(t *testing.T) {
	newName := "daily"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/snapshot_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			var req SnapshotProfilePeriodUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/snapshot_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SnapshotProfilePeriod{Key: 1, Name: newName})
		},
	}))

	period, err := client.SnapshotProfilePeriods.Update(context.Background(), 1, &SnapshotProfilePeriodUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if period.Name != newName {
		t.Errorf("expected name %q, got %q", newName, period.Name)
	}
}

func TestSnapshotProfilePeriodService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.SnapshotProfilePeriods.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/snapshot_profile_periods/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "updated"
	_, err := client.SnapshotProfilePeriods.Update(context.Background(), 999, &SnapshotProfilePeriodUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSnapshotProfilePeriodService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/snapshot_profile_periods/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.SnapshotProfilePeriods.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestSnapshotProfilePeriodService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/snapshot_profile_periods/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.SnapshotProfilePeriods.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
