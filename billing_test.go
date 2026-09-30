package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestBillingRecord_Usage(t *testing.T) {
	raw := []byte(`{
		"$key": 1,
		"created": 1704067200,
		"from": 1704063600,
		"to": 1704067200,
		"sent": 1704067201,
		"description": "January",
		"used_cores": 32,
		"total_cores": 64,
		"online_cores": 60,
		"used_ram": 65536,
		"total_ram": 131072,
		"online_ram": 131072,
		"phys_ram_used": 68719476736,
		"phys_vram_used": 17179869184,
		"tier_0_used": 107374182400,
		"tier_0_total": 1099511627776,
		"tier_1_used": 536870912000,
		"tier_1_total": 2199023255552,
		"gpus_total": 4,
		"gpus": 2,
		"gpus_idle": 2,
		"vgpus_total": 8,
		"vgpus": 4,
		"vgpus_idle": 4
	}`)
	var record BillingRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if record.Created != 1704067200 || record.From != 1704063600 || record.To != 1704067200 || record.Sent != 1704067201 {
		t.Fatalf("timestamps: %+v", record)
	}
	if record.UsedRAMGB() != 64 || record.TotalRAMGB() != 128 || record.PhysRAMUsedGB() != 64 || record.PhysVRAMUsedGB() != 16 {
		t.Fatalf("ram gb used=%v total=%v phys=%v vram=%v", record.UsedRAMGB(), record.TotalRAMGB(), record.PhysRAMUsedGB(), record.PhysVRAMUsedGB())
	}
	used, err := record.TierUsed(0)
	if err != nil || used != 107374182400 {
		t.Fatalf("tier 0 used %d err %v", used, err)
	}
	usedGB, err := record.TierUsedGB(1)
	if err != nil || usedGB != 500 {
		t.Fatalf("tier 1 used gb %v err %v", usedGB, err)
	}
	if _, err := record.TierUsed(6); !IsValidationError(err) {
		t.Fatalf("expected tier validation error, got %v", err)
	}
	if _, err := record.TierTotal(-1); !IsValidationError(err) {
		t.Fatalf("expected tier total validation error, got %v", err)
	}
	if record.TotalStorageUsedGB() != 600 {
		t.Fatalf("total storage gb %v", record.TotalStorageUsedGB())
	}
	if record.CPUUtilization() != 50 || record.RAMUtilization() != 50 || record.GPUUtilization() != 50 || record.VGPUUtilization() != 50 {
		t.Fatalf("utilization cpu=%v ram=%v gpu=%v vgpu=%v", record.CPUUtilization(), record.RAMUtilization(), record.GPUUtilization(), record.VGPUUtilization())
	}
	empty := BillingRecord{}
	if empty.CPUUtilization() != 0 || empty.RAMUtilization() != 0 || empty.GPUUtilization() != 0 || empty.VGPUUtilization() != 0 {
		t.Fatal("empty record should report zero utilization")
	}
}

func TestBillingService_ListCreated(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "created ge 10 and created le 20" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if sort := r.URL.Query().Get("sort"); sort != "-created" {
				t.Errorf("unexpected sort: %s", sort)
			}
			if fields := r.URL.Query().Get("fields"); fields != billingListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []BillingRecord{{Key: 2, Created: 20, UsedCores: 4}})
		},
	}))

	records, err := client.Billing.ListCreated(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("ListCreated failed: %v", err)
	}
	if len(records) != 1 || records[0].UsedCores != 4 {
		t.Fatalf("unexpected records: %+v", records)
	}
}

func TestBillingService_List_SortOverride(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing": func(w http.ResponseWriter, r *http.Request) {
			if sort := r.URL.Query().Get("sort"); sort != "created" {
				t.Errorf("unexpected sort: %s", sort)
			}
			jsonResponse(w, 200, []BillingRecord{})
		},
	}))

	records, err := client.Billing.List(context.Background(), WithSort("created"))
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty list, got %d", len(records))
	}
}

func TestBillingService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.Billing.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestBillingService_GetLatest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("limit") != "1" || r.URL.Query().Get("sort") != "-created" {
				t.Errorf("query %s", r.URL.RawQuery)
			}
			jsonResponse(w, 200, []BillingRecord{{Key: 8, Created: 30, UsedCores: 2, TotalCores: 4}})
		},
	}))

	record, err := client.Billing.GetLatest(context.Background())
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}
	if record.Key.Int() != 8 || record.CPUUtilization() != 50 {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestBillingService_GetLatest_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []BillingRecord{})
		},
	}))
	_, err := client.Billing.GetLatest(context.Background())
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestBillingService_Generate(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/billing_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["action"] != "generate" || len(body) != 1 {
				t.Errorf("unexpected body: %#v", body)
			}
			w.WriteHeader(http.StatusCreated)
		},
	}))
	if err := client.Billing.Generate(context.Background()); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
}

func TestBillingService_GetSummary(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("sort") != "-created" {
				t.Errorf("unexpected sort: %s", r.URL.Query().Get("sort"))
			}
			jsonResponse(w, 200, []BillingRecord{
				{UsedCores: 30, TotalCores: 100, UsedRAM: 2048, TotalRAM: 2048, GPUs: 3, GPUsTotal: 4, VGPUs: 4, VGPUsTotal: 8, Tier0Used: 3 << 30},
				{UsedCores: 10, TotalCores: 100, UsedRAM: 1024, TotalRAM: 2048, GPUs: 1, GPUsTotal: 2, VGPUs: 2, VGPUsTotal: 4, Tier0Used: 1 << 30},
			})
		},
	}))

	summary, err := client.Billing.GetSummary(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary.RecordCount != 2 || summary.AvgCPUUtilization != 20 || summary.PeakCPUCores != 30 {
		t.Fatalf("cpu summary: %+v", summary)
	}
	if summary.AvgRAMUtilization != 75 || summary.PeakRAMGB != 2 {
		t.Fatalf("ram summary: %+v", summary)
	}
	if summary.AvgStorageUsedGB != 2 || summary.PeakStorageUsedGB != 3 {
		t.Fatalf("storage summary: %+v", summary)
	}
	if summary.TotalGPUs != 4 || summary.AvgGPUsUsed != 2 || summary.TotalVGPUs != 8 || summary.AvgVGPUsUsed != 3 {
		t.Fatalf("gpu summary: %+v", summary)
	}
}

func TestBillingService_GetSummary_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/billing": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []BillingRecord{})
		},
	}))
	summary, err := client.Billing.GetSummary(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary.RecordCount != 0 || summary.PeakCPUCores != 0 || summary.TotalGPUs != 0 {
		t.Fatalf("expected zero summary, got %+v", summary)
	}
}
