package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAlarmService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Alarm{
				{Key: 1, Level: "warning", Status: "disk usage high"},
				{Key: 2, Level: "error", Status: "node offline"},
			})
		},
	}))

	alarms, err := client.Alarms.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(alarms) != 2 {
		t.Fatalf("expected 2 alarms, got %d", len(alarms))
	}
	if alarms[0].Level != "warning" {
		t.Errorf("expected level 'warning', got %q", alarms[0].Level)
	}
}

func TestAlarmService_ListActive(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for active alarms")
			}
			jsonResponse(w, 200, []Alarm{{Key: 1, Level: "error"}})
		},
	}))

	alarms, err := client.Alarms.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive failed: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("expected 1 alarm, got %d", len(alarms))
	}
}

func TestAlarmService_ListByOwner(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "owner eq 'vms/123'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Alarm{{Key: 1, Owner: "vms/123"}})
		},
	}))

	alarms, err := client.Alarms.ListByOwner(context.Background(), "vms/123")
	if err != nil {
		t.Fatalf("ListByOwner failed: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("expected 1 alarm, got %d", len(alarms))
	}
}

func TestAlarmService_ListByLevel(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Alarm{{Key: 1, Level: "critical"}})
		},
	}))

	alarms, err := client.Alarms.ListByLevel(context.Background(), "critical")
	if err != nil {
		t.Fatalf("ListByLevel failed: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("expected 1 alarm, got %d", len(alarms))
	}
}

func TestAlarmService_ListByAlarmType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Alarm{{Key: 1, AlarmType: "vm_cpu_high"}})
		},
	}))

	alarms, err := client.Alarms.ListByAlarmType(context.Background(), "vm_cpu_high")
	if err != nil {
		t.Fatalf("ListByAlarmType failed: %v", err)
	}
	if len(alarms) != 1 {
		t.Fatalf("expected 1 alarm, got %d", len(alarms))
	}
}

func TestAlarmService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Alarm{Key: 1, Level: "error", Status: "node offline"})
		},
	}))

	alarm, err := client.Alarms.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if alarm.Status != "node offline" {
		t.Errorf("expected status 'node offline', got %q", alarm.Status)
	}
}

func TestAlarmService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Alarms.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestAlarmService_Update(t *testing.T) {
	snooze := int64(9999999)
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			var req AlarmUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Snooze == nil || *req.Snooze != snooze {
				t.Errorf("expected snooze %d, got %v", snooze, req.Snooze)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Alarm{Key: 1, Snooze: snooze})
		},
	}))

	alarm, err := client.Alarms.Update(context.Background(), 1, &AlarmUpdateRequest{Snooze: &snooze})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if alarm.Snooze != snooze {
		t.Errorf("expected snooze %d, got %d", snooze, alarm.Snooze)
	}
}

func TestAlarmService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Alarms.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestAlarmService_Snooze(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Alarm{Key: 1, Snooze: 9999})
		},
	}))

	err := client.Alarms.Snooze(context.Background(), 1, 9999)
	if err != nil {
		t.Fatalf("Snooze failed: %v", err)
	}
}

func TestAlarmService_Unsnooze(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Alarm{Key: 1, Snooze: 0})
		},
	}))

	err := client.Alarms.Unsnooze(context.Background(), 1)
	if err != nil {
		t.Fatalf("Unsnooze failed: %v", err)
	}
}

func TestAlarmService_Resolve(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/alarm_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "resolve" {
				t.Errorf("expected action 'resolve', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.Alarms.Resolve(context.Background(), 1)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
}

func TestAlarmService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/alarms/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Alarms.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestAlarmService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/alarms/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Alarms.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// AlarmType tests

func TestAlarmTypeService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarm_types": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []AlarmType{
				{Key: "vm_cpu_high", Name: "VM CPU High", Level: "warning"},
				{Key: "node_offline", Name: "Node Offline", Level: "critical"},
			})
		},
	}))

	types, err := client.AlarmTypes.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(types) != 2 {
		t.Fatalf("expected 2 alarm types, got %d", len(types))
	}
}

func TestAlarmTypeService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarm_types/vm_cpu_high": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, AlarmType{Key: "vm_cpu_high", Name: "VM CPU High"})
		},
	}))

	at, err := client.AlarmTypes.Get(context.Background(), "vm_cpu_high")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if at.Name != "VM CPU High" {
		t.Errorf("expected name 'VM CPU High', got %q", at.Name)
	}
}

func TestAlarmTypeService_ListByLevel(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/alarm_types": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []AlarmType{{Key: "node_offline", Level: "critical"}})
		},
	}))

	types, err := client.AlarmTypes.ListByLevel(context.Background(), "critical")
	if err != nil {
		t.Fatalf("ListByLevel failed: %v", err)
	}
	if len(types) != 1 {
		t.Fatalf("expected 1 alarm type, got %d", len(types))
	}
}
