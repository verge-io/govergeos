package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// VNetIPSecService tests
// ---------------------------------------------------------------------------

func TestVNetIPSecService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsecs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSec{
				{Key: 1, VNet: 10, Enabled: true},
				{Key: 2, VNet: 20, Enabled: false},
			})
		},
	}))

	ipsecs, err := client.VNetIPSecs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(ipsecs) != 2 {
		t.Fatalf("expected 2 ipsecs, got %d", len(ipsecs))
	}
	if !ipsecs[0].Enabled {
		t.Error("expected first ipsec to be enabled")
	}
}

func TestVNetIPSecService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsecs/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSec{Key: 1, VNet: 10, Enabled: true, Mode: "normal"})
		},
	}))

	ipsec, err := client.VNetIPSecs.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ipsec.Mode != "normal" {
		t.Errorf("expected mode 'normal', got %q", ipsec.Mode)
	}
}

func TestVNetIPSecService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsecs/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetIPSecs.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetIPSecService_GetByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsecs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetIPSec{{Key: 1, VNet: 10}})
		},
		"GET /api/v4/vnet_ipsecs/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSec{Key: 1, VNet: 10, Enabled: true})
		},
	}))

	ipsec, err := client.VNetIPSecs.GetByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByNetwork failed: %v", err)
	}
	if int(ipsec.VNet) != 10 {
		t.Errorf("expected vnet 10, got %d", ipsec.VNet)
	}
}

func TestVNetIPSecService_GetByNetwork_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsecs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSec{})
		},
	}))

	_, err := client.VNetIPSecs.GetByNetwork(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetIPSecService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_ipsecs": func(w http.ResponseWriter, r *http.Request) {
			var req VNetIPSecCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/vnet_ipsecs/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSec{Key: 1, VNet: 10, Enabled: true, Mode: "normal"})
		},
	}))

	ipsec, err := client.VNetIPSecs.Create(context.Background(), &VNetIPSecCreateRequest{VNet: 10})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(ipsec.VNet) != 10 {
		t.Errorf("expected vnet 10, got %d", ipsec.VNet)
	}
}

func TestVNetIPSecService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecs.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecs.Create(context.Background(), &VNetIPSecCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecService_Update(t *testing.T) {
	enabled := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_ipsecs/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_ipsecs/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSec{Key: 1, VNet: 10, Enabled: false})
		},
	}))

	ipsec, err := client.VNetIPSecs.Update(context.Background(), 1, &VNetIPSecUpdateRequest{Enabled: &enabled})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if ipsec.Enabled {
		t.Error("expected ipsec to be disabled")
	}
}

func TestVNetIPSecService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecs.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_ipsecs/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetIPSecs.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetIPSecService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_ipsecs/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetIPSecs.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// VNetIPSecPhase1Service tests
// ---------------------------------------------------------------------------

func TestVNetIPSecPhase1Service_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase1s": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecPhase1{
				{Key: 1, Name: "phase1-a", RemoteGateway: "1.2.3.4"},
				{Key: 2, Name: "phase1-b", RemoteGateway: "5.6.7.8"},
			})
		},
	}))

	phase1s, err := client.VNetIPSecPhase1s.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(phase1s) != 2 {
		t.Fatalf("expected 2 phase1s, got %d", len(phase1s))
	}
}

func TestVNetIPSecPhase1Service_ListByIPSec(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase1s": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "ipsec eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetIPSecPhase1{{Key: 1, IPSec: 5, Name: "phase1-a"}})
		},
	}))

	phase1s, err := client.VNetIPSecPhase1s.ListByIPSec(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByIPSec failed: %v", err)
	}
	if len(phase1s) != 1 {
		t.Fatalf("expected 1 phase1, got %d", len(phase1s))
	}
}

func TestVNetIPSecPhase1Service_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase1s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase1{Key: 1, Name: "phase1-a", RemoteGateway: "1.2.3.4", Auth: "psk"})
		},
	}))

	p1, err := client.VNetIPSecPhase1s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p1.RemoteGateway != "1.2.3.4" {
		t.Errorf("expected remote_gateway '1.2.3.4', got %q", p1.RemoteGateway)
	}
}

func TestVNetIPSecPhase1Service_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase1s/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetIPSecPhase1s.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase1s": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecPhase1{{Key: 1, IPSec: 5, Name: "phase1-a"}})
		},
		"GET /api/v4/vnet_ipsec_phase1s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase1{Key: 1, IPSec: 5, Name: "phase1-a", RemoteGateway: "1.2.3.4"})
		},
	}))

	p1, err := client.VNetIPSecPhase1s.GetByName(context.Background(), 5, "phase1-a")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if p1.Name != "phase1-a" {
		t.Errorf("expected name 'phase1-a', got %q", p1.Name)
	}
}

func TestVNetIPSecPhase1Service_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase1s": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecPhase1{})
		},
	}))

	_, err := client.VNetIPSecPhase1s.GetByName(context.Background(), 5, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_ipsec_phase1s": func(w http.ResponseWriter, r *http.Request) {
			var req VNetIPSecPhase1CreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "phase1-a" {
				t.Errorf("expected name 'phase1-a', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/vnet_ipsec_phase1s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase1{Key: 1, IPSec: 5, Name: "phase1-a", RemoteGateway: "1.2.3.4"})
		},
	}))

	p1, err := client.VNetIPSecPhase1s.Create(context.Background(), &VNetIPSecPhase1CreateRequest{
		IPSec:         5,
		Name:          "phase1-a",
		RemoteGateway: "1.2.3.4",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if p1.Name != "phase1-a" {
		t.Errorf("expected name 'phase1-a', got %q", p1.Name)
	}
}

func TestVNetIPSecPhase1Service_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase1s.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_Create_MissingIPSec(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase1s.Create(context.Background(), &VNetIPSecPhase1CreateRequest{
		Name:          "phase1-a",
		RemoteGateway: "1.2.3.4",
	})
	if err == nil {
		t.Fatal("expected error for missing ipsec")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase1s.Create(context.Background(), &VNetIPSecPhase1CreateRequest{
		IPSec:         5,
		RemoteGateway: "1.2.3.4",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_Create_MissingRemoteGateway(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase1s.Create(context.Background(), &VNetIPSecPhase1CreateRequest{
		IPSec: 5,
		Name:  "phase1-a",
	})
	if err == nil {
		t.Fatal("expected error for missing remote_gateway")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_Update(t *testing.T) {
	newName := "phase1-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_ipsec_phase1s/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_ipsec_phase1s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase1{Key: 1, Name: newName})
		},
	}))

	p1, err := client.VNetIPSecPhase1s.Update(context.Background(), 1, &VNetIPSecPhase1UpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if p1.Name != newName {
		t.Errorf("expected name %q, got %q", newName, p1.Name)
	}
}

func TestVNetIPSecPhase1Service_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase1s.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase1Service_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_ipsec_phase1s/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetIPSecPhase1s.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetIPSecPhase1Service_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_ipsec_phase1s/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetIPSecPhase1s.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// VNetIPSecPhase2Service tests
// ---------------------------------------------------------------------------

func TestVNetIPSecPhase2Service_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase2s": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecPhase2{
				{Key: 1, Name: "phase2-a", Local: "10.0.0.0/24"},
				{Key: 2, Name: "phase2-b", Local: "10.0.1.0/24"},
			})
		},
	}))

	phase2s, err := client.VNetIPSecPhase2s.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(phase2s) != 2 {
		t.Fatalf("expected 2 phase2s, got %d", len(phase2s))
	}
}

func TestVNetIPSecPhase2Service_ListByPhase1(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase2s": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "phase1 eq 3" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetIPSecPhase2{{Key: 1, Phase1: 3, Name: "phase2-a"}})
		},
	}))

	phase2s, err := client.VNetIPSecPhase2s.ListByPhase1(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListByPhase1 failed: %v", err)
	}
	if len(phase2s) != 1 {
		t.Fatalf("expected 1 phase2, got %d", len(phase2s))
	}
}

func TestVNetIPSecPhase2Service_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase2s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase2{Key: 1, Name: "phase2-a", Local: "10.0.0.0/24", Remote: "10.0.1.0/24", Mode: "tunnel"})
		},
	}))

	p2, err := client.VNetIPSecPhase2s.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p2.Mode != "tunnel" {
		t.Errorf("expected mode 'tunnel', got %q", p2.Mode)
	}
}

func TestVNetIPSecPhase2Service_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase2s/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetIPSecPhase2s.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase2s": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecPhase2{{Key: 1, Phase1: 3, Name: "phase2-a"}})
		},
		"GET /api/v4/vnet_ipsec_phase2s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase2{Key: 1, Phase1: 3, Name: "phase2-a", Local: "10.0.0.0/24"})
		},
	}))

	p2, err := client.VNetIPSecPhase2s.GetByName(context.Background(), 3, "phase2-a")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if p2.Name != "phase2-a" {
		t.Errorf("expected name 'phase2-a', got %q", p2.Name)
	}
}

func TestVNetIPSecPhase2Service_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_phase2s": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecPhase2{})
		},
	}))

	_, err := client.VNetIPSecPhase2s.GetByName(context.Background(), 3, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_ipsec_phase2s": func(w http.ResponseWriter, r *http.Request) {
			var req VNetIPSecPhase2CreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "phase2-a" {
				t.Errorf("expected name 'phase2-a', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/vnet_ipsec_phase2s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase2{Key: 1, Phase1: 3, Name: "phase2-a", Local: "10.0.0.0/24"})
		},
	}))

	p2, err := client.VNetIPSecPhase2s.Create(context.Background(), &VNetIPSecPhase2CreateRequest{
		Phase1: 3,
		Name:   "phase2-a",
		Local:  "10.0.0.0/24",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if p2.Name != "phase2-a" {
		t.Errorf("expected name 'phase2-a', got %q", p2.Name)
	}
}

func TestVNetIPSecPhase2Service_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase2s.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_Create_MissingPhase1(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase2s.Create(context.Background(), &VNetIPSecPhase2CreateRequest{
		Name:  "phase2-a",
		Local: "10.0.0.0/24",
	})
	if err == nil {
		t.Fatal("expected error for missing phase1")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase2s.Create(context.Background(), &VNetIPSecPhase2CreateRequest{
		Phase1: 3,
		Local:  "10.0.0.0/24",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_Create_MissingLocal(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase2s.Create(context.Background(), &VNetIPSecPhase2CreateRequest{
		Phase1: 3,
		Name:   "phase2-a",
	})
	if err == nil {
		t.Fatal("expected error for missing local")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_Update(t *testing.T) {
	newName := "phase2-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_ipsec_phase2s/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_ipsec_phase2s/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecPhase2{Key: 1, Name: newName})
		},
	}))

	p2, err := client.VNetIPSecPhase2s.Update(context.Background(), 1, &VNetIPSecPhase2UpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if p2.Name != newName {
		t.Errorf("expected name %q, got %q", newName, p2.Name)
	}
}

func TestVNetIPSecPhase2Service_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetIPSecPhase2s.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetIPSecPhase2Service_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_ipsec_phase2s/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetIPSecPhase2s.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetIPSecPhase2Service_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_ipsec_phase2s/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetIPSecPhase2s.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// VNetIPSecConnectionService tests (read-only)
// ---------------------------------------------------------------------------

func TestVNetIPSecConnectionService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_connections": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetIPSecConnection{
				{Key: 1, VNet: 10, Local: "10.0.0.1", Remote: "10.0.1.1"},
				{Key: 2, VNet: 10, Local: "10.0.0.1", Remote: "10.0.2.1"},
			})
		},
	}))

	conns, err := client.VNetIPSecConnections.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(conns) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(conns))
	}
}

func TestVNetIPSecConnectionService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_connections": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetIPSecConnection{{Key: 1, VNet: 10}})
		},
	}))

	conns, err := client.VNetIPSecConnections.ListByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
}

func TestVNetIPSecConnectionService_ListByPhase1(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_connections": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "phase1 eq 3" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetIPSecConnection{{Key: 1, Phase1: 3}})
		},
	}))

	conns, err := client.VNetIPSecConnections.ListByPhase1(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListByPhase1 failed: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
}

func TestVNetIPSecConnectionService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_connections/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetIPSecConnection{Key: 1, VNet: 10, Local: "10.0.0.1", Remote: "10.0.1.1", Protocol: "esp"})
		},
	}))

	conn, err := client.VNetIPSecConnections.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if conn.Protocol != "esp" {
		t.Errorf("expected protocol 'esp', got %q", conn.Protocol)
	}
}

func TestVNetIPSecConnectionService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_ipsec_connections/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetIPSecConnections.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
