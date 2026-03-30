package vergeos

import (
	"encoding/json"
	"testing"
)

func TestFlexFK_UnmarshalJSON_Int(t *testing.T) {
	var f FlexFK
	if err := json.Unmarshal([]byte(`42`), &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Int() != 42 {
		t.Errorf("expected 42, got %d", f.Int())
	}
}

func TestFlexFK_UnmarshalJSON_String(t *testing.T) {
	var f FlexFK
	if err := json.Unmarshal([]byte(`"99"`), &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Int() != 99 {
		t.Errorf("expected 99, got %d", f.Int())
	}
}

func TestFlexFK_UnmarshalJSON_EmptyString(t *testing.T) {
	var f FlexFK
	if err := json.Unmarshal([]byte(`""`), &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Int() != 0 {
		t.Errorf("expected 0, got %d", f.Int())
	}
}

func TestFlexFK_UnmarshalJSON_ObjectWithKey(t *testing.T) {
	var f FlexFK
	if err := json.Unmarshal([]byte(`{"$key": 5, "name": "Prod"}`), &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Int() != 5 {
		t.Errorf("expected 5, got %d", f.Int())
	}
}

func TestFlexFK_UnmarshalJSON_ObjectWithID(t *testing.T) {
	var f FlexFK
	if err := json.Unmarshal([]byte(`{"id": 7, "name": "node1"}`), &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Int() != 7 {
		t.Errorf("expected 7, got %d", f.Int())
	}
}

func TestFlexFK_UnmarshalJSON_Null(t *testing.T) {
	var f FlexFK
	if err := json.Unmarshal([]byte(`null`), &f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Int() != 0 {
		t.Errorf("expected 0, got %d", f.Int())
	}
}

func TestFlexFK_MarshalJSON(t *testing.T) {
	f := FlexFK(42)
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "42" {
		t.Errorf("expected \"42\", got %s", string(data))
	}
}

func TestFlexFK_InStruct(t *testing.T) {
	type row struct {
		Key     FlexInt `json:"$key"`
		Cluster FlexFK  `json:"cluster"`
	}

	// Scalar FK
	var r1 row
	if err := json.Unmarshal([]byte(`{"$key": 1, "cluster": 3}`), &r1); err != nil {
		t.Fatalf("scalar: %v", err)
	}
	if r1.Cluster.Int() != 3 {
		t.Errorf("scalar: expected 3, got %d", r1.Cluster.Int())
	}

	// Expanded FK
	var r2 row
	if err := json.Unmarshal([]byte(`{"$key": 1, "cluster": {"$key": 3, "name": "Prod"}}`), &r2); err != nil {
		t.Fatalf("expanded: %v", err)
	}
	if r2.Cluster.Int() != 3 {
		t.Errorf("expanded: expected 3, got %d", r2.Cluster.Int())
	}
}
