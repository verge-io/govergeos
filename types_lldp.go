package vergeos

// NodeLLDPNeighbor is one LLDP neighbor discovered on a node NIC (node_lldp_neighbors).
//
// Rows are filled by LLDP discovery and are read-only. Chassis, Port, VLAN,
// and Other are the JSON objects the discovery payload stores.
type NodeLLDPNeighbor struct {
	// Key is the neighbor row key.
	Key FlexInt `json:"$key,omitempty"`
	// Node is the parent node key.
	Node int `json:"node,omitempty"`
	// NIC is the NIC key where the neighbor was discovered.
	NIC int `json:"nic,omitempty"`
	// RemoteID is the remote device identifier (rid).
	RemoteID string `json:"rid,omitempty"`
	// Via is the discovery method.
	Via string `json:"via,omitempty"`
	// Age is how old the LLDP data is.
	Age string `json:"age,omitempty"`
	// Chassis is the LLDP chassis object.
	Chassis JSONObject `json:"chassis,omitempty"`
	// Port is the LLDP port object.
	Port JSONObject `json:"port,omitempty"`
	// VLAN is the LLDP VLAN object.
	VLAN JSONObject `json:"vlan,omitempty"`
	// Other is the remaining LLDP payload.
	Other JSONObject `json:"other,omitempty"`
}

// ChassisName returns the chassis name, or the chassis id when the name is empty.
func (n NodeLLDPNeighbor) ChassisName() string {
	return firstJSONString(n.Chassis, "name", "ChassisID")
}

// PortID returns the remote port id.
func (n NodeLLDPNeighbor) PortID() string {
	return firstJSONString(n.Port, "PortID", "id")
}

func firstJSONString(obj JSONObject, keys ...string) string {
	if obj == nil {
		return ""
	}
	for _, key := range keys {
		v, ok := obj[key]
		if !ok || v == nil {
			continue
		}
		s, ok := v.(string)
		if ok && s != "" {
			return s
		}
	}
	return ""
}

const nodeLLDPNeighborListFields = "$key,node,nic,rid,via,age,chassis,port,vlan,other"
