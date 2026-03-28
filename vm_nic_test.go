package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVMNICService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 42" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VMNIC{
				{ID: FlexInt(1), Machine: 42, Name: "nic0"},
				{ID: FlexInt(2), Machine: 42, Name: "nic1"},
			})
		},
	}))

	nics, err := client.VMNICs.List(context.Background(), 42)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(nics) != 2 {
		t.Fatalf("expected 2 NICs, got %d", len(nics))
	}
	if nics[0].Name != "nic0" {
		t.Errorf("expected name 'nic0', got %q", nics[0].Name)
	}
}

func TestVMNICService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMNIC{
				ID:      FlexInt(1),
				Machine: 42,
				Name:    "nic0",
				MAC:     "00:11:22:33:44:55",
			})
		},
	}))

	nic, err := client.VMNICs.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if nic.MAC != "00:11:22:33:44:55" {
		t.Errorf("expected MAC '00:11:22:33:44:55', got %q", nic.MAC)
	}
}

func TestVMNICService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VMNICs.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMNICService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/machine_nics": func(w http.ResponseWriter, r *http.Request) {
			var req VMNICCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Machine != 10 {
				t.Errorf("expected machine 10, got %d", req.Machine)
			}
			if req.Name != "nic0" {
				t.Errorf("expected name 'nic0', got %q", req.Name)
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/machine_nics/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMNIC{
				ID:      FlexInt(5),
				Machine: 10,
				Name:    "nic0",
				Enabled: true,
			})
		},
	}))

	nic, err := client.VMNICs.Create(context.Background(), 10, &VMNICCreateRequest{
		Name: "nic0",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if nic.ID.Int() != 5 {
		t.Errorf("expected ID 5, got %d", nic.ID.Int())
	}
}

func TestVMNICService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMNICs.Create(context.Background(), 10, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMNICService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMNICs.Create(context.Background(), 10, &VMNICCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMNICService_Update(t *testing.T) {
	newName := "nic-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			var req VMNICUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMNIC{ID: FlexInt(1), Name: newName})
		},
	}))

	nic, err := client.VMNICs.Update(context.Background(), 1, &VMNICUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if nic.Name != newName {
		t.Errorf("expected name %q, got %q", newName, nic.Name)
	}
}

func TestVMNICService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VMNICs.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVMNICService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/machine_nics/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "x"
	_, err := client.VMNICs.Update(context.Background(), 999, &VMNICUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMNICService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMNIC{ID: FlexInt(1), Machine: 10, PowerState: "down"})
		},
		"DELETE /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMNICs.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVMNICService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete of already-gone NIC returns NotFoundError
	err := client.VMNICs.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted NIC")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVMNICService_Delete_HotUnplug(t *testing.T) {
	getCalls := 0
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			getCalls++
			// First call: NIC is up. After unplug action: NIC is down.
			state := "up"
			if getCalls > 1 {
				state = "down"
			}
			jsonResponse(w, 200, VMNIC{ID: FlexInt(1), Machine: 10, PowerState: state})
		},
		"POST /api/v4/vm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "hotplugnic" {
				t.Errorf("expected action 'hotplugnic', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
		"DELETE /api/v4/machine_nics/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VMNICs.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete with hot-unplug failed: %v", err)
	}
}
