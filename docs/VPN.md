---
title: VPN
description: Manage WireGuard and IPSec VPN tunnels, peers, and connections
tags: [vpn, wireguard, ipsec, peer, tunnel, site-to-site, remote-access, phase1, phase2, ike]
categories: [VPN]
---

# VPN

Manage WireGuard and IPSec VPN tunnels and peer connections.

## WireGuard VPNs

```go
// List all WireGuard interfaces
wgs, err := client.VNetWireGuards.List(ctx)

// List WireGuard interfaces for a specific network
wgs, err := client.VNetWireGuards.ListByNetwork(ctx, networkID)

// Get a WireGuard interface by name
wg, err := client.VNetWireGuards.GetByName(ctx, networkID, "wg0")

// Create a WireGuard interface
wg, err := client.VNetWireGuards.Create(ctx, &vergeos.VNetWireGuardCreateRequest{
    VNet:       networkID,
    Name:       "wg0",
    IP:         "192.168.255.1/24",
    ListenPort: ptr(51820),
})

// Update a WireGuard interface
wg, err := client.VNetWireGuards.Update(ctx, wgID, &vergeos.VNetWireGuardUpdateRequest{
    Description: ptr("Main VPN tunnel"),
})

// Delete a WireGuard interface
err = client.VNetWireGuards.Delete(ctx, wgID)
```

---

## WireGuard Peers

Manage WireGuard peer connections (clients or site-to-site tunnels).

```go
// List all peers for a WireGuard interface
peers, err := client.VNetWireGuardPeers.ListByWireGuard(ctx, wgID)

// Get a peer by name
peer, err := client.VNetWireGuardPeers.GetByName(ctx, wgID, "remote-office")

// Create a peer (site-to-site)
peer, err := client.VNetWireGuardPeers.Create(ctx, &vergeos.VNetWireGuardPeerCreateRequest{
    WireGuard:         wgID,
    Name:              "remote-office",
    PeerIP:            "192.168.255.2",
    PublicKey:         "base64-encoded-public-key",
    AllowedIPs:        "10.0.0.0/24,172.16.0.0/16",
    Endpoint:          "vpn.remote-office.com",
    Port:              ptr(51820),
    ConfigureFirewall: ptr(vergeos.WireGuardPeerFirewallSiteToSite),
    Keepalive:         ptr(25),
})

// Create a peer (remote user/roaming client)
peer, err := client.VNetWireGuardPeers.Create(ctx, &vergeos.VNetWireGuardPeerCreateRequest{
    WireGuard:         wgID,
    Name:              "laptop-user",
    PeerIP:            "192.168.255.10",
    PublicKey:         "base64-encoded-public-key",
    AllowedIPs:        "192.168.255.10/32",
    ConfigureFirewall: ptr(vergeos.WireGuardPeerFirewallRemoteUser),
    AutogeneratePeer:  ptr(true),  // Enable config file download
})

// Get peer configuration file (for clients with autogenerate_peer=true)
config, err := client.VNetWireGuardPeers.GetConfig(ctx, peerID)
fmt.Println(config)  // WireGuard .conf file content

// Update a peer
peer, err := client.VNetWireGuardPeers.Update(ctx, peerID, &vergeos.VNetWireGuardPeerUpdateRequest{
    AllowedIPs: ptr("10.0.0.0/24,172.16.0.0/16,192.168.0.0/24"),
})

// Delete a peer
err = client.VNetWireGuardPeers.Delete(ctx, peerID)

// Get peer connection status
status, err := client.VNetWireGuardPeerStatus.GetByPeer(ctx, peerID)
fmt.Printf("Last handshake: %d, TX: %d bytes, RX: %d bytes\n",
    status.LastHandshake, status.TXBytes, status.RXBytes)
```

---

## IPSec VPNs

Manage IPSec VPN tunnels with Phase 1 (IKE) and Phase 2 (IPsec SA) configurations.

```go
// Get or create IPSec configuration for a network
ipsec, err := client.VNetIPSecs.GetByNetwork(ctx, networkID)
if vergeos.IsNotFoundError(err) {
    ipsec, err = client.VNetIPSecs.Create(ctx, &vergeos.VNetIPSecCreateRequest{
        VNet: networkID,
    })
}

// Create a Phase 1 (IKE SA) configuration
phase1, err := client.VNetIPSecPhase1s.Create(ctx, &vergeos.VNetIPSecPhase1CreateRequest{
    IPSec:         int(ipsec.Key),
    Name:          "branch-office",
    RemoteGateway: "203.0.113.1",
    PSK:           ptr("super-secret-preshared-key"),
    KeyExchange:   ptr(vergeos.IPSecKeyExchangeIKEv2),
    Auto:          ptr(vergeos.IPSecAutoStart),
})

// Create a Phase 2 (IPsec SA) configuration
phase2, err := client.VNetIPSecPhase2s.Create(ctx, &vergeos.VNetIPSecPhase2CreateRequest{
    Phase1: int(phase1.Key),
    Name:   "lan-to-lan",
    Local:  "10.0.0.0/24",
    Remote: "192.168.0.0/24",
})

// List active connections
conns, err := client.VNetIPSecConnections.ListByNetwork(ctx, networkID)
for _, conn := range conns {
    fmt.Printf("Connection: %s, Local: %s, Remote: %s\n",
        conn.Connection, conn.LocalNetwork, conn.RemoteNetwork)
}

// Update Phase 1 settings
phase1, err = client.VNetIPSecPhase1s.Update(ctx, phase1ID, &vergeos.VNetIPSecPhase1UpdateRequest{
    DPDAction: ptr(vergeos.IPSecDPDRestart),
    DPDDelay:  ptr(30),
})

// Delete Phase 2, Phase 1 (cascade deletes Phase 2s)
err = client.VNetIPSecPhase2s.Delete(ctx, phase2ID)
err = client.VNetIPSecPhase1s.Delete(ctx, phase1ID)
```
