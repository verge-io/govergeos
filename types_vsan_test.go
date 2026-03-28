package vergeos

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClusterTier_UnmarshalWithOnlineFields(t *testing.T) {
	data := `{
		"$key": 1,
		"cluster": 1,
		"tier": 0,
		"drives_online": [
			{"state": "online"},
			{"state": "online"},
			{"state": "online"}
		],
		"nodes_online": {
			"nodes": [
				{"state": "online"},
				{"state": "online"}
			]
		}
	}`

	var ct ClusterTier
	if err := json.Unmarshal([]byte(data), &ct); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ct.Key.Int() != 1 {
		t.Errorf("Key: expected 1, got %d", ct.Key.Int())
	}
	if ct.Cluster.Int() != 1 {
		t.Errorf("Cluster: expected 1, got %d", ct.Cluster.Int())
	}
	if ct.CountOnlineDrives() != 3 {
		t.Errorf("CountOnlineDrives: expected 3, got %d", ct.CountOnlineDrives())
	}
	if ct.CountOnlineNodes() != 2 {
		t.Errorf("CountOnlineNodes: expected 2, got %d", ct.CountOnlineNodes())
	}
}

func TestClusterTier_UnmarshalWithExpandedClusterFK(t *testing.T) {
	data := `{
		"$key": 1,
		"cluster": {"$key": 3, "name": "Prod"},
		"tier": 0,
		"drives_online": [{"state": "online"}],
		"nodes_online": {"nodes": [{"state": "online"}]}
	}`

	var ct ClusterTier
	if err := json.Unmarshal([]byte(data), &ct); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ct.Cluster.Int() != 3 {
		t.Errorf("Cluster: expected 3, got %d", ct.Cluster.Int())
	}
}

func TestClusterTier_EmptyOnlineFields(t *testing.T) {
	data := `{"$key": 1, "cluster": 1, "tier": 0}`

	var ct ClusterTier
	if err := json.Unmarshal([]byte(data), &ct); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ct.CountOnlineNodes() != 0 {
		t.Errorf("CountOnlineNodes: expected 0, got %d", ct.CountOnlineNodes())
	}
	if ct.CountOnlineDrives() != 0 {
		t.Errorf("CountOnlineDrives: expected 0, got %d", ct.CountOnlineDrives())
	}
}

func TestClusterTier_MixedDriveStates(t *testing.T) {
	data := `{
		"$key": 1,
		"cluster": 1,
		"tier": 0,
		"drives_online": [
			{"state": "online"},
			{"state": "offline"},
			{"state": "online"},
			{"state": "repairing"}
		]
	}`

	var ct ClusterTier
	if err := json.Unmarshal([]byte(data), &ct); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ct.CountOnlineDrives() != 2 {
		t.Errorf("CountOnlineDrives: expected 2, got %d", ct.CountOnlineDrives())
	}
}

func TestClusterTierListFields_ContainsOnlineExpressions(t *testing.T) {
	if !strings.Contains(clusterTierListFields, "drives_online") {
		t.Error("clusterTierListFields missing drives_online expression")
	}
	if !strings.Contains(clusterTierListFields, "nodes_online") {
		t.Error("clusterTierListFields missing nodes_online expression")
	}
}
