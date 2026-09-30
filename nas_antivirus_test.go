package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestNASServiceAntivirusService_ListByService(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_antivirus": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "service eq 4 and enabled eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != nasServiceAntivirusListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []NASServiceAntivirus{{
				Key: 7, Service: 4, Enabled: true, MaxRecursion: 15,
			}})
		},
	}))

	rows, err := client.NASServiceAntivirus.ListByService(context.Background(), 4, WithFilter("enabled eq true"))
	if err != nil {
		t.Fatalf("ListByService failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Key.Int() != 7 || !rows[0].Enabled {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestNASServiceAntivirusService_GetByService_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_antivirus": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("filter") != "service eq 4" || r.URL.Query().Get("limit") != "1" {
				t.Errorf("query %s", r.URL.RawQuery)
			}
			jsonResponse(w, 200, []NASServiceAntivirus{})
		},
	}))
	_, err := client.NASServiceAntivirus.GetByService(context.Background(), 4)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestNASServiceAntivirusService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_service_antivirus/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.NASServiceAntivirus.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestNASServiceAntivirusService_Update(t *testing.T) {
	enabled := false
	updates := false
	recursion := 20
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["enabled"] != false || body["database_updates_enabled"] != false {
				t.Errorf("bool fields: %#v", body)
			}
			if body["max_recursion"] != float64(20) {
				t.Errorf("max_recursion: %#v", body["max_recursion"])
			}
			if _, ok := body["database_location"]; ok {
				t.Errorf("unset field was sent: %#v", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceAntivirus{
				Key: 1, Service: 4, Enabled: false, MaxRecursion: 20, DatabaseUpdatesEnabled: false,
			})
		},
	}))

	row, err := client.NASServiceAntivirus.Update(context.Background(), 1, &NASServiceAntivirusUpdateRequest{
		Enabled:                &enabled,
		MaxRecursion:           &recursion,
		DatabaseUpdatesEnabled: &updates,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if row.MaxRecursion != 20 || row.Enabled || row.DatabaseUpdatesEnabled {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestNASServiceAntivirusService_Update_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			t.Error("empty update posted a PUT")
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceAntivirus{Key: 1, Service: 4, MaxRecursion: 15})
		},
	}))

	row, err := client.NASServiceAntivirus.Update(context.Background(), 1, &NASServiceAntivirusUpdateRequest{})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if row.MaxRecursion != 15 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestNASServiceAntivirusService_Update_RecursionRange(t *testing.T) {
	recursion := 101
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			t.Error("invalid recursion posted a PUT")
			w.WriteHeader(http.StatusOK)
		},
	}))
	_, err := client.NASServiceAntivirus.Update(context.Background(), 1, &NASServiceAntivirusUpdateRequest{
		MaxRecursion: &recursion,
	})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}

	zero := 0
	client = newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["max_recursion"] != float64(0) {
				t.Errorf("zero recursion was not sent: %#v", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vm_service_antivirus/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, NASServiceAntivirus{Key: 1, MaxRecursion: 0})
		},
	}))
	if _, err := client.NASServiceAntivirus.Update(context.Background(), 1, &NASServiceAntivirusUpdateRequest{MaxRecursion: &zero}); err != nil {
		t.Fatalf("Update 0 failed: %v", err)
	}
}

func TestNASServiceAntivirusService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	_, err := client.NASServiceAntivirus.Update(context.Background(), 1, nil)
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	_, err = client.NASServiceAntivirus.ListByService(context.Background(), 0)
	if !IsValidationError(err) {
		t.Fatalf("expected service validation error, got %v", err)
	}
}
