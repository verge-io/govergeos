package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// --- VNetDNSViewService tests ---

func TestVNetDNSViewService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_views": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSView{
				{Key: 1, VNet: 10, Name: "default"},
				{Key: 2, VNet: 10, Name: "internal"},
			})
		},
	}))

	views, err := client.VNetDNSViews.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	if views[0].Name != "default" {
		t.Errorf("expected name 'default', got %q", views[0].Name)
	}
}

func TestVNetDNSViewService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_views": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetDNSView{{Key: 1, VNet: 10, Name: "default"}})
		},
	}))

	views, err := client.VNetDNSViews.ListByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
}

func TestVNetDNSViewService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_views/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSView{Key: 1, VNet: 10, Name: "default", Recursion: true})
		},
	}))

	view, err := client.VNetDNSViews.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if view.Name != "default" {
		t.Errorf("expected name 'default', got %q", view.Name)
	}
	if !view.Recursion {
		t.Error("expected recursion to be true")
	}
}

func TestVNetDNSViewService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_views/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetDNSViews.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetDNSViewService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_views": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSView{{Key: 1, VNet: 10, Name: "default"}})
		},
		"GET /api/v4/vnet_dns_views/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSView{Key: 1, VNet: 10, Name: "default", Recursion: true})
		},
	}))

	view, err := client.VNetDNSViews.GetByName(context.Background(), 10, "default")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if !view.Recursion {
		t.Error("expected recursion to be true")
	}
}

func TestVNetDNSViewService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_views": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSView{})
		},
	}))

	_, err := client.VNetDNSViews.GetByName(context.Background(), 10, "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetDNSViewService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_dns_views": func(w http.ResponseWriter, r *http.Request) {
			var req VNetDNSViewCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "internal" {
				t.Errorf("expected name 'internal', got %q", req.Name)
			}
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			jsonResponse(w, 200, map[string]any{"$key": 3, "status": "ok"})
		},
		"GET /api/v4/vnet_dns_views/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSView{Key: 3, VNet: 10, Name: "internal"})
		},
	}))

	view, err := client.VNetDNSViews.Create(context.Background(), &VNetDNSViewCreateRequest{
		VNet: 10,
		Name: "internal",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(view.Key) != 3 {
		t.Errorf("expected key 3, got %d", int(view.Key))
	}
}

func TestVNetDNSViewService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSViews.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSViewService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSViews.Create(context.Background(), &VNetDNSViewCreateRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSViewService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSViews.Create(context.Background(), &VNetDNSViewCreateRequest{VNet: 10})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSViewService_Update(t *testing.T) {
	newName := "external"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_dns_views/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetDNSViewUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_dns_views/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSView{Key: 1, Name: newName})
		},
	}))

	view, err := client.VNetDNSViews.Update(context.Background(), 1, &VNetDNSViewUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if view.Name != newName {
		t.Errorf("expected name %q, got %q", newName, view.Name)
	}
}

func TestVNetDNSViewService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSViews.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSViewService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_dns_views/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetDNSViews.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetDNSViewService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_dns_views/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetDNSViews.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- VNetDNSZoneService tests ---

func TestVNetDNSZoneService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zones": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSZone{
				{Key: 1, View: 1, Domain: "example.com", Type: "master"},
				{Key: 2, View: 1, Domain: "internal.local", Type: "master"},
			})
		},
	}))

	zones, err := client.VNetDNSZones.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("expected 2 zones, got %d", len(zones))
	}
	if zones[0].Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", zones[0].Domain)
	}
}

func TestVNetDNSZoneService_ListByView(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zones": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "view eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetDNSZone{{Key: 1, View: 1, Domain: "example.com"}})
		},
	}))

	zones, err := client.VNetDNSZones.ListByView(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByView failed: %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
}

func TestVNetDNSZoneService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zones/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSZone{Key: 1, View: 1, Domain: "example.com", Type: "master", Nameserver: "ns1.example.com"})
		},
	}))

	zone, err := client.VNetDNSZones.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if zone.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", zone.Domain)
	}
	if zone.Nameserver != "ns1.example.com" {
		t.Errorf("expected nameserver 'ns1.example.com', got %q", zone.Nameserver)
	}
}

func TestVNetDNSZoneService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zones/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetDNSZones.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetDNSZoneService_GetByDomain(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zones": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSZone{{Key: 1, View: 1, Domain: "example.com"}})
		},
		"GET /api/v4/vnet_dns_zones/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSZone{Key: 1, View: 1, Domain: "example.com", Type: "master"})
		},
	}))

	zone, err := client.VNetDNSZones.GetByDomain(context.Background(), 1, "example.com")
	if err != nil {
		t.Fatalf("GetByDomain failed: %v", err)
	}
	if zone.Type != "master" {
		t.Errorf("expected type 'master', got %q", zone.Type)
	}
}

func TestVNetDNSZoneService_GetByDomain_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zones": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSZone{})
		},
	}))

	_, err := client.VNetDNSZones.GetByDomain(context.Background(), 1, "nope.com")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetDNSZoneService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_dns_zones": func(w http.ResponseWriter, r *http.Request) {
			var req VNetDNSZoneCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Domain != "test.local" {
				t.Errorf("expected domain 'test.local', got %q", req.Domain)
			}
			if req.View != 1 {
				t.Errorf("expected view 1, got %d", req.View)
			}
			jsonResponse(w, 200, map[string]any{"$key": 5, "status": "ok"})
		},
		"GET /api/v4/vnet_dns_zones/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSZone{Key: 5, View: 1, Domain: "test.local", Type: "master"})
		},
	}))

	zone, err := client.VNetDNSZones.Create(context.Background(), &VNetDNSZoneCreateRequest{
		View:   1,
		Domain: "test.local",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(zone.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(zone.Key))
	}
}

func TestVNetDNSZoneService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSZones.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSZoneService_Create_MissingView(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSZones.Create(context.Background(), &VNetDNSZoneCreateRequest{Domain: "test.local"})
	if err == nil {
		t.Fatal("expected error for missing view")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSZoneService_Create_MissingDomain(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSZones.Create(context.Background(), &VNetDNSZoneCreateRequest{View: 1})
	if err == nil {
		t.Fatal("expected error for missing domain")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSZoneService_Update(t *testing.T) {
	newDomain := "updated.local"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_dns_zones/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetDNSZoneUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Domain == nil || *req.Domain != newDomain {
				t.Errorf("expected domain %q, got %v", newDomain, req.Domain)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_dns_zones/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSZone{Key: 1, Domain: newDomain})
		},
	}))

	zone, err := client.VNetDNSZones.Update(context.Background(), 1, &VNetDNSZoneUpdateRequest{Domain: &newDomain})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if zone.Domain != newDomain {
		t.Errorf("expected domain %q, got %q", newDomain, zone.Domain)
	}
}

func TestVNetDNSZoneService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSZones.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSZoneService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_dns_zones/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetDNSZones.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetDNSZoneService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_dns_zones/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetDNSZones.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- VNetDNSRecordService tests ---

func TestVNetDNSRecordService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSRecord{
				{Key: 1, Zone: 1, Host: "www", Type: "A", Value: "192.168.1.10"},
				{Key: 2, Zone: 1, Host: "mail", Type: "A", Value: "192.168.1.20"},
			})
		},
	}))

	records, err := client.VNetDNSRecords.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].Host != "www" {
		t.Errorf("expected host 'www', got %q", records[0].Host)
	}
}

func TestVNetDNSRecordService_ListByZone(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "zone eq 1" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetDNSRecord{{Key: 1, Zone: 1, Host: "www"}})
		},
	}))

	records, err := client.VNetDNSRecords.ListByZone(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByZone failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestVNetDNSRecordService_ListByType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "zone eq 1 and type eq 'MX'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetDNSRecord{{Key: 3, Zone: 1, Type: "MX", Value: "mail.example.com"}})
		},
	}))

	records, err := client.VNetDNSRecords.ListByType(context.Background(), 1, "MX")
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestVNetDNSRecordService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSRecord{Key: 1, Zone: 1, Host: "www", Type: "A", Value: "192.168.1.10"})
		},
	}))

	record, err := client.VNetDNSRecords.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if record.Type != "A" {
		t.Errorf("expected type 'A', got %q", record.Type)
	}
	if record.Value != "192.168.1.10" {
		t.Errorf("expected value '192.168.1.10', got %q", record.Value)
	}
}

func TestVNetDNSRecordService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetDNSRecords.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_GetByHostAndType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSRecord{{Key: 1, Zone: 1, Host: "www", Type: "A"}})
		},
		"GET /api/v4/vnet_dns_zone_records/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSRecord{Key: 1, Zone: 1, Host: "www", Type: "A", Value: "192.168.1.10"})
		},
	}))

	record, err := client.VNetDNSRecords.GetByHostAndType(context.Background(), 1, "www", "A")
	if err != nil {
		t.Fatalf("GetByHostAndType failed: %v", err)
	}
	if record.Value != "192.168.1.10" {
		t.Errorf("expected value '192.168.1.10', got %q", record.Value)
	}
}

func TestVNetDNSRecordService_GetByHostAndType_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_dns_zone_records": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetDNSRecord{})
		},
	}))

	_, err := client.VNetDNSRecords.GetByHostAndType(context.Background(), 1, "nope", "A")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_dns_zone_records": func(w http.ResponseWriter, r *http.Request) {
			var req VNetDNSRecordCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Zone != 1 {
				t.Errorf("expected zone 1, got %d", req.Zone)
			}
			if req.Type != "A" {
				t.Errorf("expected type 'A', got %q", req.Type)
			}
			if req.Value != "192.168.1.30" {
				t.Errorf("expected value '192.168.1.30', got %q", req.Value)
			}
			jsonResponse(w, 200, map[string]any{"$key": 5, "status": "ok"})
		},
		"GET /api/v4/vnet_dns_zone_records/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSRecord{Key: 5, Zone: 1, Host: "api", Type: "A", Value: "192.168.1.30"})
		},
	}))

	record, err := client.VNetDNSRecords.Create(context.Background(), &VNetDNSRecordCreateRequest{
		Zone:  1,
		Host:  "api",
		Type:  "A",
		Value: "192.168.1.30",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(record.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(record.Key))
	}
}

func TestVNetDNSRecordService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSRecords.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_Create_MissingZone(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSRecords.Create(context.Background(), &VNetDNSRecordCreateRequest{Type: "A", Value: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected error for missing zone")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_Create_MissingType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSRecords.Create(context.Background(), &VNetDNSRecordCreateRequest{Zone: 1, Value: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_Create_MissingValue(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSRecords.Create(context.Background(), &VNetDNSRecordCreateRequest{Zone: 1, Type: "A"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_Update(t *testing.T) {
	newValue := "192.168.1.99"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_dns_zone_records/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetDNSRecordUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Value == nil || *req.Value != newValue {
				t.Errorf("expected value %q, got %v", newValue, req.Value)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_dns_zone_records/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetDNSRecord{Key: 1, Value: newValue})
		},
	}))

	record, err := client.VNetDNSRecords.Update(context.Background(), 1, &VNetDNSRecordUpdateRequest{Value: &newValue})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if record.Value != newValue {
		t.Errorf("expected value %q, got %q", newValue, record.Value)
	}
}

func TestVNetDNSRecordService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetDNSRecords.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetDNSRecordService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_dns_zone_records/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetDNSRecords.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetDNSRecordService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_dns_zone_records/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetDNSRecords.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
