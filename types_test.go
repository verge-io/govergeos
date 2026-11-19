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

func TestFlexString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{`"users/3"`, "users/3"},
		{`123`, "123"},
		{`null`, ""},
		{`""`, ""},
	}
	for _, tt := range tests {
		var got FlexString
		if err := json.Unmarshal([]byte(tt.in), &got); err != nil {
			t.Fatalf("unmarshal %s: %v", tt.in, err)
		}
		if got.String() != tt.want {
			t.Errorf("unmarshal %s: got %q, want %q", tt.in, got, tt.want)
		}
	}

	var bad FlexString
	if err := json.Unmarshal([]byte(`{"$key":1}`), &bad); err == nil {
		t.Fatal("expected error for object")
	}
}

func TestFlexString_MarshalJSON(t *testing.T) {
	data, err := json.Marshal(FlexString("vms/39"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"vms/39"` {
		t.Fatalf("got %s", data)
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

func TestIdentityKey_JSON(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    IdentityKey
		wantErr bool
	}{
		{name: "number", in: `16`, want: 16},
		{name: "string", in: `"16"`, want: 16},
		{name: "empty string", in: `""`, want: 0},
		{name: "object", in: `{"$key":16}`, wantErr: true},
		{name: "text", in: `"abc"`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got IdentityKey
			err := json.Unmarshal([]byte(tt.in), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}

	data, err := json.Marshal(IdentityKey(16))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "16" {
		t.Fatalf("marshal = %s, want 16", data)
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
