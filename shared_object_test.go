package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestSharedObject_SnapshotKey(t *testing.T) {
	key, ok := (SharedObject{Snapshot: "machine_snapshots/14"}).SnapshotKey()
	if !ok || key != 14 {
		t.Fatalf("snapshot key %d ok %v", key, ok)
	}
	if _, ok := (SharedObject{Snapshot: "14"}).SnapshotKey(); ok {
		t.Fatal("a value without a slash is not a snapshot path")
	}
	if _, ok := (SharedObject{Snapshot: "machine_snapshots/abc"}).SnapshotKey(); ok {
		t.Fatal("a non-numeric snapshot key should be rejected")
	}
	if _, ok := (SharedObject{}).SnapshotKey(); ok {
		t.Fatal("an empty snapshot should be rejected")
	}
}

func TestSharedObjectService_ListByTenant(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/shared_objects": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "recipient eq 10 and inbox eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			if fields := r.URL.Query().Get("fields"); fields != sharedObjectListFields {
				t.Errorf("unexpected fields: %s", fields)
			}
			jsonResponse(w, 200, []SharedObject{{
				Key: 1, Name: "Ubuntu Template", Recipient: 10, Type: "vm", Inbox: true,
			}})
		},
	}))

	objects, err := client.SharedObjects.ListByTenant(context.Background(), 10, WithFilter("inbox eq true"))
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(objects) != 1 || objects[0].Name != "Ubuntu Template" || !objects[0].Inbox {
		t.Fatalf("unexpected objects: %+v", objects)
	}
}

func TestSharedObjectService_ListInbox(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/shared_objects": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("filter") != "inbox eq true" {
				t.Errorf("unexpected filter: %s", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []SharedObject{})
		},
	}))
	objects, err := client.SharedObjects.ListInbox(context.Background())
	if err != nil {
		t.Fatalf("ListInbox failed: %v", err)
	}
	if len(objects) != 0 {
		t.Fatalf("expected empty list, got %d", len(objects))
	}
}

func TestSharedObjectService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/shared_objects": func(w http.ResponseWriter, r *http.Request) {
			want := "recipient eq 10 and name eq 'O\\'Brien'"
			if r.URL.Query().Get("filter") != want {
				t.Errorf("filter %q", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []SharedObject{{Key: 2, Name: "O'Brien", Recipient: 10}})
		},
	}))
	object, err := client.SharedObjects.GetByName(context.Background(), 10, "O'Brien")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if object.Key.Int() != 2 {
		t.Fatalf("unexpected object: %+v", object)
	}
}

func TestSharedObjectService_GetByName_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/shared_objects": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []SharedObject{
				{Key: 1, Name: "Template", Recipient: 10},
				{Key: 2, Name: "Template", Recipient: 10},
			})
		},
	}))
	_, err := client.SharedObjects.GetByName(context.Background(), 10, "Template")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("expected AmbiguousNameError, got %v", err)
	}
}

func TestSharedObjectService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/shared_objects/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.SharedObjects.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestSharedObjectService_Create(t *testing.T) {
	var snapshotDeleted bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/100": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 100, Machine: 200, Name: "template-vm"})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode snapshot: %v", err)
			}
			if body["machine"] != float64(200) || body["expires_type"] != "never" || body["created_manually"] != true {
				t.Errorf("snapshot body %#v", body)
			}
			name, _ := body["name"].(string)
			if !strings.HasPrefix(name, "share-template-vm-") {
				t.Errorf("snapshot name %q", name)
			}
			suffix := strings.TrimPrefix(name, "share-template-vm-")
			n, err := strconv.Atoi(suffix)
			if err != nil || n < 10000 || n > 99999 {
				t.Errorf("snapshot suffix %q", suffix)
			}
			jsonResponse(w, 201, map[string]any{"$key": 50})
		},
		"POST /api/v4/shared_objects": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode share: %v", err)
			}
			if body["recipient"] != float64(10) || body["type"] != "vm" || body["name"] != "template-vm" {
				t.Errorf("share body %#v", body)
			}
			if body["snapshot"] != "machine_snapshots/50" {
				t.Errorf("snapshot path %#v", body["snapshot"])
			}
			if body["description"] != "Shared template" {
				t.Errorf("description %#v", body["description"])
			}
			jsonResponse(w, 201, map[string]any{"$key": 1})
		},
		"GET /api/v4/shared_objects/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SharedObject{
				Key: 1, Name: "template-vm", Recipient: 10, Type: "vm", Snapshot: "machine_snapshots/50",
			})
		},
		"DELETE /api/v4/machine_snapshots/50": func(w http.ResponseWriter, r *http.Request) {
			snapshotDeleted = true
			w.WriteHeader(http.StatusNoContent)
		},
	}))

	object, err := client.SharedObjects.Create(context.Background(), &SharedObjectCreateRequest{
		Tenant:      10,
		VM:          100,
		Description: "Shared template",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if object.Name != "template-vm" || object.Snapshot != "machine_snapshots/50" {
		t.Fatalf("unexpected object: %+v", object)
	}
	key, ok := object.SnapshotKey()
	if !ok || key != 50 {
		t.Fatalf("snapshot key %d ok %v", key, ok)
	}
	if snapshotDeleted {
		t.Fatal("snapshot was deleted after a successful share")
	}
}

func TestSharedObjectService_Create_CleansUpSnapshot(t *testing.T) {
	var snapshotDeleted bool
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/100": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 100, Machine: 200, Name: "template-vm"})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 201, map[string]any{"$key": 50})
		},
		"POST /api/v4/shared_objects": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 500, map[string]string{"err": "failed"})
		},
		"DELETE /api/v4/machine_snapshots/50": func(w http.ResponseWriter, r *http.Request) {
			snapshotDeleted = true
			w.WriteHeader(http.StatusNoContent)
		},
	}))

	_, err := client.SharedObjects.Create(context.Background(), &SharedObjectCreateRequest{
		Tenant:       10,
		VM:           100,
		Name:         "template-vm",
		SnapshotName: "share-explicit",
	})
	if err == nil {
		t.Fatal("expected create error")
	}
	if !snapshotDeleted {
		t.Fatal("snapshot was left behind after the share failed")
	}
}

func TestSharedObjectService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/100": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VM{Key: 100, Name: "template-vm"})
		},
		"POST /api/v4/machine_snapshots": func(w http.ResponseWriter, r *http.Request) {
			t.Error("snapshot was posted for an invalid request")
			w.WriteHeader(http.StatusCreated)
		},
	}))
	ctx := context.Background()
	if _, err := client.SharedObjects.Create(ctx, nil); !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}
	if _, err := client.SharedObjects.Create(ctx, &SharedObjectCreateRequest{VM: 100}); !IsValidationError(err) {
		t.Fatalf("missing tenant: %v", err)
	}
	if _, err := client.SharedObjects.Create(ctx, &SharedObjectCreateRequest{Tenant: 10}); !IsValidationError(err) {
		t.Fatalf("missing vm: %v", err)
	}
	if _, err := client.SharedObjects.Create(ctx, &SharedObjectCreateRequest{Tenant: 10, VM: 100}); !IsValidationError(err) {
		t.Fatalf("missing machine: %v", err)
	}
}

func TestSharedObjectService_ImportAndRefresh(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/shared_object_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["shared_object"] != float64(7) {
				t.Errorf("shared_object %#v", body["shared_object"])
			}
			action, _ := body["action"].(string)
			if action != "import" && action != "refresh" {
				t.Errorf("action %#v", body["action"])
			}
			w.WriteHeader(http.StatusCreated)
		},
	}))
	ctx := context.Background()
	if err := client.SharedObjects.Import(ctx, 7); err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if err := client.SharedObjects.Refresh(ctx, 7); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	if err := client.SharedObjects.Import(ctx, 0); !IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSharedObjectService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/shared_objects/7": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		"DELETE /api/v4/shared_objects/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	if err := client.SharedObjects.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if err := client.SharedObjects.Delete(context.Background(), 9); !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
