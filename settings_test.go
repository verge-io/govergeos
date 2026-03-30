package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestSettingsService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Setting{
				{Key: "cloud_name", Value: "mycloud"},
				{Key: "timezone", Value: "UTC"},
			})
		},
	}))

	settings, err := client.Settings.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(settings) != 2 {
		t.Fatalf("expected 2 settings, got %d", len(settings))
	}
	if settings[0].Key != "cloud_name" {
		t.Errorf("expected key 'cloud_name', got %q", settings[0].Key)
	}
}

func TestSettingsService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Setting{Key: "cloud_name", Value: "mycloud", Description: "The cloud name"})
		},
	}))

	setting, err := client.Settings.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if setting.Key != "cloud_name" {
		t.Errorf("expected key 'cloud_name', got %q", setting.Key)
	}
	if setting.Value != "mycloud" {
		t.Errorf("expected value 'mycloud', got %q", setting.Value)
	}
}

func TestSettingsService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Settings.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSettingsService_GetByKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "key eq 'cloud_name'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Setting{{Key: "cloud_name", Value: "mycloud"}})
		},
	}))

	setting, err := client.Settings.GetByKey(context.Background(), "cloud_name")
	if err != nil {
		t.Fatalf("GetByKey failed: %v", err)
	}
	if setting.Value != "mycloud" {
		t.Errorf("expected value 'mycloud', got %q", setting.Value)
	}
}

func TestSettingsService_GetByKey_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Setting{})
		},
	}))

	_, err := client.Settings.GetByKey(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSettingsService_GetValue(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Setting{{Key: "timezone", Value: "America/Denver"}})
		},
	}))

	val, err := client.Settings.GetValue(context.Background(), "timezone")
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if val != "America/Denver" {
		t.Errorf("expected 'America/Denver', got %q", val)
	}
}

func TestSettingsService_GetCloudName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/settings": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "key eq 'cloud_name'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Setting{{Key: "cloud_name", Value: "prod-cloud"}})
		},
	}))

	name, err := client.Settings.GetCloudName(context.Background())
	if err != nil {
		t.Fatalf("GetCloudName failed: %v", err)
	}
	if name != "prod-cloud" {
		t.Errorf("expected 'prod-cloud', got %q", name)
	}
}
