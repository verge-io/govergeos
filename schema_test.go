package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestSchemaService_GetTableSchema(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/$table": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TableSchema{
				Fields: map[string]TableField{
					"name":         {Type: "string", Required: true},
					"machine_type": {Type: "string", List: map[string]string{"pc": "PC", "q35": "Q35"}},
				},
			})
		},
	}))

	schema, err := client.Schema.GetTableSchema(context.Background(), "vms")
	if err != nil {
		t.Fatalf("GetTableSchema failed: %v", err)
	}
	if len(schema.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(schema.Fields))
	}
	if schema.Fields["name"].Type != "string" {
		t.Errorf("expected name type 'string', got %q", schema.Fields["name"].Type)
	}
	if !schema.Fields["name"].Required {
		t.Error("expected name to be required")
	}
}

func TestSchemaService_GetValidValues(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/$table": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TableSchema{
				Fields: map[string]TableField{
					"machine_type": {Type: "string", List: map[string]string{"pc": "Standard PC", "q35": "Q35 Chipset"}},
				},
			})
		},
	}))

	values, err := client.Schema.GetValidValues(context.Background(), "vms", "machine_type")
	if err != nil {
		t.Fatalf("GetValidValues failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values["pc"] != "Standard PC" {
		t.Errorf("expected 'Standard PC', got %q", values["pc"])
	}
}

func TestSchemaService_GetValidValues_FieldNotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/$table": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TableSchema{
				Fields: map[string]TableField{
					"name": {Type: "string"},
				},
			})
		},
	}))

	_, err := client.Schema.GetValidValues(context.Background(), "vms", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent field")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestSchemaService_GetValidValues_NoList(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/$table": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TableSchema{
				Fields: map[string]TableField{
					"name": {Type: "string"},
				},
			})
		},
	}))

	values, err := client.Schema.GetValidValues(context.Background(), "vms", "name")
	if err != nil {
		t.Fatalf("GetValidValues failed: %v", err)
	}
	if len(values) != 0 {
		t.Errorf("expected empty map, got %d values", len(values))
	}
}

func TestSchemaService_GetVMMachineTypes(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/$table": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TableSchema{
				Fields: map[string]TableField{
					"machine_type": {Type: "string", List: map[string]string{"pc": "PC", "q35": "Q35"}},
				},
			})
		},
	}))

	types, err := client.Schema.GetVMMachineTypes(context.Background())
	if err != nil {
		t.Fatalf("GetVMMachineTypes failed: %v", err)
	}
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}
}

func TestSchemaService_GetVMOSFamilies(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms/$table": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TableSchema{
				Fields: map[string]TableField{
					"os_family": {Type: "string", List: map[string]string{"linux": "Linux", "windows": "Windows"}},
				},
			})
		},
	}))

	families, err := client.Schema.GetVMOSFamilies(context.Background())
	if err != nil {
		t.Fatalf("GetVMOSFamilies failed: %v", err)
	}
	if len(families) != 2 {
		t.Fatalf("expected 2 families, got %d", len(families))
	}
	if families["linux"] != "Linux" {
		t.Errorf("expected 'Linux', got %q", families["linux"])
	}
}
