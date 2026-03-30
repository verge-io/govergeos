package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMachineStatusService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineStatus{
				{Key: 1, Machine: 10, Running: true, Status: "running", State: "online"},
				{Key: 2, Machine: 20, Running: false, Status: "stopped", State: "offline"},
			})
		},
	}))

	statuses, err := client.MachineStatus.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if !statuses[0].Running {
		t.Error("expected first status to be running")
	}
	if statuses[1].Status != "stopped" {
		t.Errorf("expected status 'stopped', got %q", statuses[1].Status)
	}
}

func TestMachineStatusService_List_Empty(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineStatus{})
		},
	}))

	statuses, err := client.MachineStatus.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestMachineStatusService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_status": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "machine eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []MachineStatus{
				{Key: 1, Machine: 10, Running: true, Status: "running", State: "online", RunningCores: 4, RunningRAM: 8192},
			})
		},
	}))

	status, err := client.MachineStatus.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if status.Machine != 10 {
		t.Errorf("expected machine 10, got %d", status.Machine)
	}
	if status.RunningCores != 4 {
		t.Errorf("expected 4 cores, got %d", status.RunningCores)
	}
}

func TestMachineStatusService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []MachineStatus{})
		},
	}))

	_, err := client.MachineStatus.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMachineStatusService_GetByKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_status/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, MachineStatus{
				Key: 1, Machine: 10, Running: true, Status: "running", NodeName: "node1",
			})
		},
	}))

	status, err := client.MachineStatus.GetByKey(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByKey failed: %v", err)
	}
	if status.NodeName != "node1" {
		t.Errorf("expected node_name 'node1', got %q", status.NodeName)
	}
}

func TestGuestInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty array", `[]`, false},
		{"empty array with spaces", `[ ]`, false},
		{"valid object", `{"hostname":"test","last_refresh":1234}`, false},
		{"empty object", `{}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g GuestInfo
			err := json.Unmarshal([]byte(tt.input), &g)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGuestInfo_UnmarshalJSON_InMachineStatus(t *testing.T) {
	// Simulate the full MachineStatus response where agent_guest_info is []
	input := `{"$key":1,"machine":10,"running":true,"agent_guest_info":[]}`
	var ms MachineStatus
	if err := json.Unmarshal([]byte(input), &ms); err != nil {
		t.Fatalf("unexpected error unmarshaling MachineStatus with empty guest info: %v", err)
	}
	if ms.AgentGuestInfo == nil {
		t.Fatal("expected AgentGuestInfo to be non-nil (allocated but zero-value)")
	}
	if ms.AgentGuestInfo.Hostname != "" {
		t.Errorf("expected empty hostname, got %q", ms.AgentGuestInfo.Hostname)
	}
}

func TestMachineStatusService_GetByKey_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/machine_status/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.MachineStatus.GetByKey(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
