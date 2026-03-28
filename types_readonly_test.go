package vergeos

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNode_UnmarshalWithVMStatsTotals(t *testing.T) {
	data := `{
		"id": 1,
		"name": "node1",
		"cluster": 1,
		"vm_stats_totals": {
			"running_cores": 30,
			"running_ram": 75776
		}
	}`

	var node Node
	if err := json.Unmarshal([]byte(data), &node); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if node.ID != 1 {
		t.Errorf("ID: expected 1, got %d", node.ID)
	}
	if node.Name != "node1" {
		t.Errorf("Name: expected node1, got %s", node.Name)
	}
	if node.VMStatsTotals == nil {
		t.Fatal("VMStatsTotals is nil")
	}
	if node.VMStatsTotals.RunningCores != 30 {
		t.Errorf("RunningCores: expected 30, got %d", node.VMStatsTotals.RunningCores)
	}
	if node.VMStatsTotals.RunningRAM != 75776 {
		t.Errorf("RunningRAM: expected 75776, got %d", node.VMStatsTotals.RunningRAM)
	}
}

func TestNode_UnmarshalWithoutVMStatsTotals(t *testing.T) {
	data := `{"id": 1, "name": "node1"}`

	var node Node
	if err := json.Unmarshal([]byte(data), &node); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if node.VMStatsTotals != nil {
		t.Errorf("VMStatsTotals: expected nil, got %+v", node.VMStatsTotals)
	}
}

func TestNodeListFields_ContainsVMStatsTotals(t *testing.T) {
	if !strings.Contains(nodeListFields, "vm_stats_totals") {
		t.Error("nodeListFields missing vm_stats_totals expression")
	}
	if !strings.Contains(nodeListFields, "running_machines[sum(running_cores),sum(running_ram)]") {
		t.Error("nodeListFields missing running_machines aggregate expression")
	}
}
