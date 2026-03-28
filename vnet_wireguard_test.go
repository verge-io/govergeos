package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// VNetWireGuardService tests
// ---------------------------------------------------------------------------

func TestVNetWireGuardService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguards": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuard{
				{Key: 1, Name: "wg0", IP: "10.0.0.1/24"},
				{Key: 2, Name: "wg1", IP: "10.0.1.1/24"},
			})
		},
	}))

	wgs, err := client.VNetWireGuards.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(wgs) != 2 {
		t.Fatalf("expected 2 wireguards, got %d", len(wgs))
	}
	if wgs[0].Name != "wg0" {
		t.Errorf("expected name 'wg0', got %q", wgs[0].Name)
	}
}

func TestVNetWireGuardService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguards": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetWireGuard{{Key: 1, VNet: 10, Name: "wg0"}})
		},
	}))

	wgs, err := client.VNetWireGuards.ListByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(wgs) != 1 {
		t.Fatalf("expected 1 wireguard, got %d", len(wgs))
	}
}

func TestVNetWireGuardService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguards/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuard{Key: 1, Name: "wg0", IP: "10.0.0.1/24", ListenPort: 51820})
		},
	}))

	wg, err := client.VNetWireGuards.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if wg.Name != "wg0" {
		t.Errorf("expected name 'wg0', got %q", wg.Name)
	}
	if wg.ListenPort != 51820 {
		t.Errorf("expected port 51820, got %d", wg.ListenPort)
	}
}

func TestVNetWireGuardService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguards/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetWireGuards.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguards": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuard{{Key: 1, VNet: 10, Name: "wg0"}})
		},
		"GET /api/v4/vnet_wireguards/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuard{Key: 1, VNet: 10, Name: "wg0", IP: "10.0.0.1/24"})
		},
	}))

	wg, err := client.VNetWireGuards.GetByName(context.Background(), 10, "wg0")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if wg.Name != "wg0" {
		t.Errorf("expected name 'wg0', got %q", wg.Name)
	}
}

func TestVNetWireGuardService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguards": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuard{})
		},
	}))

	_, err := client.VNetWireGuards.GetByName(context.Background(), 10, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_wireguards": func(w http.ResponseWriter, r *http.Request) {
			var req VNetWireGuardCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "wg0" {
				t.Errorf("expected name 'wg0', got %q", req.Name)
			}
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/vnet_wireguards/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuard{Key: 1, VNet: 10, Name: "wg0", IP: "10.0.0.1/24"})
		},
	}))

	wg, err := client.VNetWireGuards.Create(context.Background(), &VNetWireGuardCreateRequest{
		VNet: 10,
		Name: "wg0",
		IP:   "10.0.0.1/24",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if wg.Name != "wg0" {
		t.Errorf("expected name 'wg0', got %q", wg.Name)
	}
}

func TestVNetWireGuardService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuards.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuards.Create(context.Background(), &VNetWireGuardCreateRequest{
		Name: "wg0",
		IP:   "10.0.0.1/24",
	})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuards.Create(context.Background(), &VNetWireGuardCreateRequest{
		VNet: 10,
		IP:   "10.0.0.1/24",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_Create_MissingIP(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuards.Create(context.Background(), &VNetWireGuardCreateRequest{
		VNet: 10,
		Name: "wg0",
	})
	if err == nil {
		t.Fatal("expected error for missing ip")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_Update(t *testing.T) {
	newName := "wg0-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_wireguards/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetWireGuardUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_wireguards/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuard{Key: 1, Name: newName})
		},
	}))

	wg, err := client.VNetWireGuards.Update(context.Background(), 1, &VNetWireGuardUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if wg.Name != newName {
		t.Errorf("expected name %q, got %q", newName, wg.Name)
	}
}

func TestVNetWireGuardService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuards.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_wireguards/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetWireGuards.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetWireGuardService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_wireguards/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetWireGuards.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// VNetWireGuardPeerService tests
// ---------------------------------------------------------------------------

func TestVNetWireGuardPeerService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuardPeer{
				{Key: 1, Name: "peer-a", PublicKey: "abc123"},
				{Key: 2, Name: "peer-b", PublicKey: "def456"},
			})
		},
	}))

	peers, err := client.VNetWireGuardPeers.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(peers))
	}
}

func TestVNetWireGuardPeerService_ListByWireGuard(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "wireguard eq 5" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetWireGuardPeer{{Key: 1, WireGuard: 5, Name: "peer-a"}})
		},
	}))

	peers, err := client.VNetWireGuardPeers.ListByWireGuard(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByWireGuard failed: %v", err)
	}
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
}

func TestVNetWireGuardPeerService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuardPeer{Key: 1, Name: "peer-a", PeerIP: "10.0.0.2", PublicKey: "abc123"})
		},
	}))

	peer, err := client.VNetWireGuardPeers.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if peer.Name != "peer-a" {
		t.Errorf("expected name 'peer-a', got %q", peer.Name)
	}
}

func TestVNetWireGuardPeerService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetWireGuardPeers.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuardPeer{{Key: 1, WireGuard: 5, Name: "peer-a"}})
		},
		"GET /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuardPeer{Key: 1, WireGuard: 5, Name: "peer-a", PeerIP: "10.0.0.2"})
		},
	}))

	peer, err := client.VNetWireGuardPeers.GetByName(context.Background(), 5, "peer-a")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if peer.Name != "peer-a" {
		t.Errorf("expected name 'peer-a', got %q", peer.Name)
	}
}

func TestVNetWireGuardPeerService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuardPeer{})
		},
	}))

	_, err := client.VNetWireGuardPeers.GetByName(context.Background(), 5, "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_wireguard_peers": func(w http.ResponseWriter, r *http.Request) {
			var req VNetWireGuardPeerCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "peer-a" {
				t.Errorf("expected name 'peer-a', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuardPeer{Key: 1, WireGuard: 5, Name: "peer-a", PeerIP: "10.0.0.2", PublicKey: "abc123", AllowedIPs: "10.0.0.0/24"})
		},
	}))

	peer, err := client.VNetWireGuardPeers.Create(context.Background(), &VNetWireGuardPeerCreateRequest{
		WireGuard:  5,
		Name:       "peer-a",
		PeerIP:     "10.0.0.2",
		PublicKey:  "abc123",
		AllowedIPs: "10.0.0.0/24",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if peer.Name != "peer-a" {
		t.Errorf("expected name 'peer-a', got %q", peer.Name)
	}
}

func TestVNetWireGuardPeerService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Create_MissingWireGuard(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Create(context.Background(), &VNetWireGuardPeerCreateRequest{
		Name:       "peer-a",
		PeerIP:     "10.0.0.2",
		PublicKey:  "abc123",
		AllowedIPs: "10.0.0.0/24",
	})
	if err == nil {
		t.Fatal("expected error for missing wireguard")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Create(context.Background(), &VNetWireGuardPeerCreateRequest{
		WireGuard:  5,
		PeerIP:     "10.0.0.2",
		PublicKey:  "abc123",
		AllowedIPs: "10.0.0.0/24",
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Create_MissingPeerIP(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Create(context.Background(), &VNetWireGuardPeerCreateRequest{
		WireGuard:  5,
		Name:       "peer-a",
		PublicKey:  "abc123",
		AllowedIPs: "10.0.0.0/24",
	})
	if err == nil {
		t.Fatal("expected error for missing peer_ip")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Create_MissingPublicKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Create(context.Background(), &VNetWireGuardPeerCreateRequest{
		WireGuard:  5,
		Name:       "peer-a",
		PeerIP:     "10.0.0.2",
		AllowedIPs: "10.0.0.0/24",
	})
	if err == nil {
		t.Fatal("expected error for missing public_key")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Create_MissingAllowedIPs(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Create(context.Background(), &VNetWireGuardPeerCreateRequest{
		WireGuard: 5,
		Name:      "peer-a",
		PeerIP:    "10.0.0.2",
		PublicKey: "abc123",
	})
	if err == nil {
		t.Fatal("expected error for missing allowed_ips")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Update(t *testing.T) {
	newName := "peer-updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuardPeer{Key: 1, Name: newName})
		},
	}))

	peer, err := client.VNetWireGuardPeers.Update(context.Background(), 1, &VNetWireGuardPeerUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if peer.Name != newName {
		t.Errorf("expected name %q, got %q", newName, peer.Name)
	}
}

func TestVNetWireGuardPeerService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetWireGuardPeers.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetWireGuardPeers.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetWireGuardPeerService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_wireguard_peers/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetWireGuardPeers.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerService_GetConfig(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]string{"wg_config": "[Interface]\nPrivateKey = abc\nAddress = 10.0.0.2/24"})
		},
	}))

	config, err := client.VNetWireGuardPeers.GetConfig(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if config == "" {
		t.Error("expected non-empty config")
	}
}

func TestVNetWireGuardPeerService_GetConfig_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peers/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetWireGuardPeers.GetConfig(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// VNetWireGuardPeerStatusService tests
// ---------------------------------------------------------------------------

func TestVNetWireGuardPeerStatusService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peer_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuardPeerStatus{
				{Key: 1, Peer: 10, TXBytes: 1024, RXBytes: 2048},
				{Key: 2, Peer: 11, TXBytes: 512, RXBytes: 4096},
			})
		},
	}))

	statuses, err := client.VNetWireGuardPeerStatus.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if statuses[0].TXBytes != 1024 {
		t.Errorf("expected tx_bytes 1024, got %d", statuses[0].TXBytes)
	}
}

func TestVNetWireGuardPeerStatusService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peer_status/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetWireGuardPeerStatus{Key: 1, Peer: 10, TXBytes: 1024, RXBytes: 2048})
		},
	}))

	status, err := client.VNetWireGuardPeerStatus.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if status.RXBytes != 2048 {
		t.Errorf("expected rx_bytes 2048, got %d", status.RXBytes)
	}
}

func TestVNetWireGuardPeerStatusService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peer_status/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetWireGuardPeerStatus.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetWireGuardPeerStatusService_GetByPeer(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peer_status": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "peer eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetWireGuardPeerStatus{{Key: 1, Peer: 10, TXBytes: 1024}})
		},
	}))

	status, err := client.VNetWireGuardPeerStatus.GetByPeer(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByPeer failed: %v", err)
	}
	if int(status.Peer) != 10 {
		t.Errorf("expected peer 10, got %d", status.Peer)
	}
}

func TestVNetWireGuardPeerStatusService_GetByPeer_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_wireguard_peer_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetWireGuardPeerStatus{})
		},
	}))

	_, err := client.VNetWireGuardPeerStatus.GetByPeer(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
