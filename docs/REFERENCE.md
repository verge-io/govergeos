# goVergeOS API Reference

This document contains detailed API examples for all goVergeOS services. For a quick start guide and overview, see the main [README](../README.md).

## Table of Contents

- [Virtual Machines](#virtual-machines)
- [VM Drives](#vm-drives)
- [VM NICs](#vm-nics)
- [Networks](#networks)
- [Firewall Rules](#firewall-rules)
- [Rule Aliases](#rule-aliases)
- [Network Addresses](#network-addresses)
- [DNS Views](#dns-views)
- [DNS Zones](#dns-zones)
- [DNS Records](#dns-records)
- [Host Overrides](#host-overrides)
- [WireGuard VPNs](#wireguard-vpns)
- [WireGuard Peers](#wireguard-peers)
- [IPSec VPNs](#ipsec-vpns)
- [Certificates](#certificates)
- [Volumes](#volumes)
- [Volume Shares](#volume-shares)
- [Volume Browser](#volume-browser)
- [Tenants](#tenants)
- [Tenant Nodes](#tenant-nodes)
- [Tenant Storage](#tenant-storage)
- [Users](#users)
- [Groups](#groups)
- [User API Keys](#user-api-keys)
- [Nodes](#nodes)
- [Clusters](#clusters)
- [Settings](#settings)
- [System](#system)
- [Tags](#tags)
- [Tag Members](#tag-members)
- [Snapshot Profiles](#snapshot-profiles)
- [Snapshot Profile Periods](#snapshot-profile-periods)
- [Alarms](#alarms)
- [Alarm Types](#alarm-types)
- [Tasks](#tasks)
- [Sites & DR](#sites--dr)
- [Cloud Snapshots](#cloud-snapshots)
- [Webhooks](#webhooks)

---

## Virtual Machines

```go
// List all VMs
vms, err := client.VMs.List(ctx)

// Get a specific VM
vm, err := client.VMs.Get(ctx, vmID)

// Create a new VM
vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
    Name:     "my-vm",
    CPUCores: 4,
    RAM:      8192,
    Cluster:  clusterID,
})

// Update a VM
newName := "renamed-vm"
vm, err := client.VMs.Update(ctx, vmID, &vergeos.VMUpdateRequest{
    Name: &newName,
})

// Delete a VM
err = client.VMs.Delete(ctx, vmID)

// Power operations
err = client.VMs.PowerOn(ctx, vmID)
err = client.VMs.PowerOff(ctx, vmID)
```

---

## VM Drives

```go
// List drives for a VM
drives, err := client.VMDrives.List(ctx, vmMachineID)

// Create a drive
drive, err := client.VMDrives.Create(ctx, vmMachineID, &vergeos.VMDriveCreateRequest{
    Name:      "disk0",
    Interface: "virtio",
    Media:     "disk",
    SizeGB:    50,
})

// Update a drive (resize)
drive, err := client.VMDrives.Update(ctx, driveID, &vergeos.VMDriveUpdateRequest{
    SizeGB: ptr(100), // Increase size
})
```

---

## VM NICs

```go
// List NICs for a VM
nics, err := client.VMNICs.List(ctx, vmMachineID)

// Create a NIC
nic, err := client.VMNICs.Create(ctx, vmMachineID, &vergeos.VMNICCreateRequest{
    Name: "eth0",
    VNET: networkID,
})
```

---

## Networks

```go
// List all networks
networks, err := client.Networks.List(ctx)

// Create a network
dhcpEnabled := true
network, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
    Name:        "my-network",
    Network:     "10.0.0.0/24",
    DHCPEnabled: &dhcpEnabled,
})

// Update a network
dhcpStart := "10.0.0.100"
dhcpStop := "10.0.0.200"
network, err := client.Networks.Update(ctx, networkID, &vergeos.NetworkUpdateRequest{
    DHCPStart: &dhcpStart,
    DHCPStop:  &dhcpStop,
})
```

---

## Firewall Rules

Manage network firewall rules for traffic control.

```go
// List all firewall rules
rules, err := client.VNetRules.List(ctx)

// List rules for a specific network
rules, err := client.VNetRules.ListByNetwork(ctx, networkID)

// Get a rule by ID
rule, err := client.VNetRules.Get(ctx, ruleID)

// Get a rule by name within a network
rule, err := client.VNetRules.GetByName(ctx, networkID, "allow-ssh")

// Create a firewall rule
rule, err := client.VNetRules.Create(ctx, &vergeos.VNetRuleCreateRequest{
    VNet:             networkID,
    Name:             "allow-ssh",
    Protocol:         ptr("tcp"),
    Direction:        ptr("incoming"),
    DestinationPorts: ptr("22"),
    Action:           ptr("accept"),
})

// Update a rule
rule, err := client.VNetRules.Update(ctx, ruleID, &vergeos.VNetRuleUpdateRequest{
    Enabled: ptr(false),
})

// Delete a rule
err = client.VNetRules.Delete(ctx, ruleID)

// Enable/disable with automatic rule application
err = client.VNetRules.Enable(ctx, ruleID, true)   // apply=true
err = client.VNetRules.Disable(ctx, ruleID, true)
```

---

## Rule Aliases

Rule aliases define reusable address lists for firewall rules.

```go
// List all aliases
aliases, err := client.VNetRuleAliases.List(ctx)

// Get an alias by name
alias, err := client.VNetRuleAliases.GetByName(ctx, "trusted-networks")

// Create an alias
alias, err := client.VNetRuleAliases.Create(ctx, &vergeos.VNetRuleAliasCreateRequest{
    Name:  "trusted-networks",
    Value: "192.168.1.0/24,10.0.0.0/8",
})

// Update an alias
alias, err := client.VNetRuleAliases.Update(ctx, aliasID, &vergeos.VNetRuleAliasUpdateRequest{
    Value: ptr("192.168.1.0/24,10.0.0.0/8,172.16.0.0/12"),
})

// Delete an alias
err = client.VNetRuleAliases.Delete(ctx, aliasID)
```

---

## Network Addresses

Manage IP addresses within a network (static, DHCP, IP aliases, proxy ARP, virtual IPs).

```go
// List all addresses in a network
addresses, err := client.VNetAddresses.ListByNetwork(ctx, networkID)

// List static addresses only
addresses, err := client.VNetAddresses.ListByType(ctx, networkID, vergeos.AddressTypeStatic)

// Get an address by IP
address, err := client.VNetAddresses.GetByIP(ctx, networkID, "192.168.1.100")

// Create a static address
address, err := client.VNetAddresses.Create(ctx, &vergeos.VNetAddressCreateRequest{
    VNet:     networkID,
    Type:     vergeos.AddressTypeStatic,
    IP:       "192.168.1.100",
    Hostname: "web-server",
})

// Update an address
address, err := client.VNetAddresses.Update(ctx, addressID, &vergeos.VNetAddressUpdateRequest{
    Hostname: ptr("web-server-01"),
})

// Delete an address
err = client.VNetAddresses.Delete(ctx, addressID)
```

---

## DNS Views

DNS views allow different DNS responses based on client IP addresses.

```go
// List all DNS views in a network
views, err := client.VNetDNSViews.ListByNetwork(ctx, networkID)

// Create a DNS view with recursion enabled
view, err := client.VNetDNSViews.Create(ctx, &vergeos.VNetDNSViewCreateRequest{
    VNet:         networkID,
    Name:         "internal",
    Recursion:    ptr(true),
    MatchClients: ptr("10.0.0.0/8;172.16.0.0/12;"),
})

// Update a view
view, err := client.VNetDNSViews.Update(ctx, viewID, &vergeos.VNetDNSViewUpdateRequest{
    MatchClients: ptr("10.0.0.0/8;172.16.0.0/12;192.168.0.0/16;"),
})

// Delete a view
err = client.VNetDNSViews.Delete(ctx, viewID)
```

---

## DNS Zones

Manage DNS zones within DNS views.

```go
// List all zones in a view
zones, err := client.VNetDNSZones.ListByView(ctx, viewID)

// Create a primary DNS zone
zone, err := client.VNetDNSZones.Create(ctx, &vergeos.VNetDNSZoneCreateRequest{
    View:       viewID,
    Domain:     "example.com",
    Type:       ptr(vergeos.DNSZoneTypeMaster),
    Nameserver: ptr("ns1.example.com"),
    Email:      ptr("admin@example.com"),
})

// Get a zone by domain
zone, err := client.VNetDNSZones.GetByDomain(ctx, viewID, "example.com")

// Update a zone
zone, err := client.VNetDNSZones.Update(ctx, zoneID, &vergeos.VNetDNSZoneUpdateRequest{
    DefaultTTL: ptr("2h"),
})

// Delete a zone
err = client.VNetDNSZones.Delete(ctx, zoneID)
```

---

## DNS Records

Manage DNS records within zones.

```go
// List all records in a zone
records, err := client.VNetDNSRecords.ListByZone(ctx, zoneID)

// List only A records
records, err := client.VNetDNSRecords.ListByType(ctx, zoneID, vergeos.DNSRecordTypeA)

// Create an A record
record, err := client.VNetDNSRecords.Create(ctx, &vergeos.VNetDNSRecordCreateRequest{
    Zone:  zoneID,
    Host:  "www",
    Type:  vergeos.DNSRecordTypeA,
    Value: "192.168.1.100",
    TTL:   ptr("1h"),
})

// Create an MX record
record, err := client.VNetDNSRecords.Create(ctx, &vergeos.VNetDNSRecordCreateRequest{
    Zone:         zoneID,
    Host:         "@",
    Type:         vergeos.DNSRecordTypeMX,
    Value:        "mail.example.com",
    MXPreference: ptr(10),
})

// Update a record
record, err := client.VNetDNSRecords.Update(ctx, recordID, &vergeos.VNetDNSRecordUpdateRequest{
    Value: ptr("192.168.1.101"),
})

// Delete a record
err = client.VNetDNSRecords.Delete(ctx, recordID)
```

---

## Host Overrides

Static hostname-to-IP mappings for DNS and DHCP.

```go
// List all host overrides in a network
hosts, err := client.VNetHosts.ListByNetwork(ctx, networkID)

// Get a host by hostname
host, err := client.VNetHosts.GetByHost(ctx, networkID, "printer.local")

// Create a host override
host, err := client.VNetHosts.Create(ctx, &vergeos.VNetHostCreateRequest{
    VNet: networkID,
    Host: "printer.local",
    IP:   "192.168.1.50",
})

// Update a host override
host, err := client.VNetHosts.Update(ctx, hostID, &vergeos.VNetHostUpdateRequest{
    IP: ptr("192.168.1.51"),
})

// Delete a host override
err = client.VNetHosts.Delete(ctx, hostID)
```

---

## WireGuard VPNs

Manage WireGuard VPN interfaces and peer connections.

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

---

## Certificates

Manage SSL/TLS certificates including Let's Encrypt, manual, and self-signed.

```go
// List all certificates
certs, err := client.Certificates.List(ctx)

// List valid certificates
validCerts, err := client.Certificates.ListValid(ctx)

// Get a certificate by domain
cert, err := client.Certificates.GetByDomain(ctx, "example.com")

// Create a Let's Encrypt certificate
cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
    DomainList: "example.com,www.example.com",
    Type:       vergeos.CertificateTypeLetsEncrypt,
    Contact:    ptr(adminUserID),
    AgreeTOS:   ptr(true),
})

// Create a self-signed certificate
cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
    DomainName: "internal.local",
    Type:       vergeos.CertificateTypeSelfSigned,
    KeyType:    vergeos.CertificateKeyTypeECDSA,
})

// Upload a manual certificate
cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
    DomainName: "secure.example.com",
    Type:       vergeos.CertificateTypeManual,
    Public:     publicKeyPEM,
    Private:    privateKeyPEM,
    Chain:      chainPEM,
})

// Get certificate with keys (for export)
certWithKeys, err := client.Certificates.GetWithKeys(ctx, certID)
fmt.Println(certWithKeys.Public)   // PEM public key
fmt.Println(certWithKeys.Private)  // PEM private key

// Renew a Let's Encrypt certificate
cert, err = client.Certificates.Renew(ctx, certID)

// Delete a certificate
err = client.Certificates.Delete(ctx, certID)
```

---

## Volumes

Manage NAS volumes for storage. Note: Volumes use SHA1 hash strings as IDs.

```go
// List all volumes
volumes, err := client.Volumes.List(ctx)

// List volumes for a specific NAS service
volumes, err := client.Volumes.ListByService(ctx, serviceID)

// Get a volume by ID (SHA1 hash string)
volume, err := client.Volumes.Get(ctx, "0d25c256a0c561c0b5bb9087f04fcb49f16a8048")

// Get a volume by name within a service
volume, err := client.Volumes.GetByName(ctx, serviceID, "data-vol")

// Create a volume
volume, err := client.Volumes.Create(ctx, &vergeos.VolumeCreateRequest{
    Name:    "data-vol",
    Service: serviceID,
    MaxSize: ptr(int64(100 * 1024 * 1024 * 1024)), // 100GB
})

// Update a volume
volume, err := client.Volumes.Update(ctx, volumeID, &vergeos.VolumeUpdateRequest{
    Description: ptr("Production data volume"),
})

// Delete a volume
err = client.Volumes.Delete(ctx, volumeID)

// Enable/disable/reset a volume
err = client.Volumes.Enable(ctx, volumeID)
err = client.Volumes.Disable(ctx, volumeID)
err = client.Volumes.Reset(ctx, volumeID)
```

---

## Volume Shares

Manage CIFS (SMB) and NFS shares on volumes.

```go
// List all CIFS shares
shares, err := client.VolumeCIFSShares.List(ctx)

// Create a CIFS share
share, err := client.VolumeCIFSShares.Create(ctx, &vergeos.VolumeCIFSShareCreateRequest{
    Volume:     volumeID,
    Name:       "shared-data",
    Path:       "/data",
    Browseable: ptr(true),
    ReadOnly:   ptr(false),
})

// Create an NFS share
share, err := client.VolumeNFSShares.Create(ctx, &vergeos.VolumeNFSShareCreateRequest{
    Volume:     volumeID,
    Name:       "nfs-export",
    Path:       "/export",
    AllowedIPs: ptr("10.0.0.0/8"),
    Squash:     ptr(vergeos.NFSSquashRoot),
})

// Delete a share
err = client.VolumeCIFSShares.Delete(ctx, shareID)
```

---

## Volume Browser

Browse files within volumes asynchronously.

```go
// Simple browse (handles async polling internally)
entries, err := client.VolumeBrowser.Browse(ctx, volumeID, "/", 100)
for _, entry := range entries {
    fmt.Printf("%s %s (%d bytes)\n", entry.Type, entry.Name, entry.Size)
}

// Advanced browse with options
entries, err := client.VolumeBrowser.BrowseWithOptions(ctx, volumeID, "/data", 50,
    ptr(0),       // offset
    ".txt,.log",  // file extensions filter
)
```

---

## Tenants

Manage multi-tenant virtual data centers.

```go
// List all tenants
tenants, err := client.Tenants.List(ctx)

// Get a tenant by ID
tenant, err := client.Tenants.Get(ctx, tenantID)

// Get a tenant by name
tenant, err := client.Tenants.GetByName(ctx, "customer-a")

// Create a tenant
tenant, err := client.Tenants.Create(ctx, &vergeos.TenantCreateRequest{
    Name:        "customer-a",
    Password:    "admin-password",
    Description: ptr("Customer A production environment"),
})

// Update a tenant
tenant, err := client.Tenants.Update(ctx, tenantID, &vergeos.TenantUpdateRequest{
    Description: ptr("Updated description"),
})

// Delete a tenant
err = client.Tenants.Delete(ctx, tenantID)

// Power operations
err = client.Tenants.PowerOn(ctx, tenantID)
err = client.Tenants.PowerOff(ctx, tenantID)
err = client.Tenants.Reset(ctx, tenantID)

// Clone a tenant
err = client.Tenants.Clone(ctx, tenantID, &vergeos.TenantCloneOptions{
    Name:      "customer-a-clone",
    NoStorage: false,
    NoNodes:   false,
})

// Network isolation
err = client.Tenants.IsolateOn(ctx, tenantID)
err = client.Tenants.IsolateOff(ctx, tenantID)
```

---

## Tenant Nodes

Manage virtual nodes within tenants.

```go
// List nodes for a tenant
nodes, err := client.TenantNodes.ListByTenant(ctx, tenantID)

// Create a tenant node
node, err := client.TenantNodes.Create(ctx, &vergeos.TenantNodeCreateRequest{
    Tenant:   tenantID,
    Name:     "node1",
    CPUCores: 4,
    RAM:      16384, // 16GB in MB
})

// Power operations
err = client.TenantNodes.PowerOn(ctx, nodeID)
err = client.TenantNodes.PowerOff(ctx, nodeID)
err = client.TenantNodes.Reset(ctx, nodeID)

// Migrate a node to another host
err = client.TenantNodes.Migrate(ctx, nodeID, targetHostNodeID)
```

---

## Tenant Storage

Manage storage allocations for tenants.

```go
// List storage allocations for a tenant
storage, err := client.TenantStorage.ListByTenant(ctx, tenantID)

// Create a storage allocation
storage, err := client.TenantStorage.Create(ctx, &vergeos.TenantStorageCreateRequest{
    Tenant:      tenantID,
    Tier:        tierID,
    Provisioned: 100 * 1024 * 1024 * 1024, // 100GB in bytes
})

// Update storage allocation
storage, err := client.TenantStorage.Update(ctx, storageID, &vergeos.TenantStorageUpdateRequest{
    Provisioned: ptr(int64(200 * 1024 * 1024 * 1024)), // 200GB
})
```

---

## Users

```go
// List users
users, err := client.Users.List(ctx)

// Create a user
user, err := client.Users.Create(ctx, &vergeos.UserCreateRequest{
    Name:     "newuser",
    Password: "securepassword",
})
```

---

## Groups

```go
// List groups
groups, err := client.Groups.List(ctx)

// Get a group by ID
group, err := client.Groups.Get(ctx, groupID)

// Get a group by name
group, err := client.Groups.GetByName(ctx, "developers")

// Create a group
group, err := client.Groups.Create(ctx, &vergeos.GroupCreateRequest{
    Name:        "developers",
    Description: "Development team",
})

// Update a group
group, err := client.Groups.Update(ctx, groupID, &vergeos.GroupUpdateRequest{
    Description: ptr("Updated description"),
})

// Delete a group
err = client.Groups.Delete(ctx, groupID)

// Manage group members
err = client.Members.Add(ctx, groupID, userID)
```

---

## User API Keys

Manage API keys for programmatic access.

```go
// List all API keys
keys, err := client.UserAPIKeys.List(ctx)

// List API keys for a specific user
keys, err := client.UserAPIKeys.ListByUser(ctx, userID)

// Create an API key (token only returned on creation!)
key, token, err := client.UserAPIKeys.Create(ctx, &vergeos.UserAPIKeyCreateRequest{
    User:        userID,
    Name:        "automation-key",
    Description: "Key for CI/CD pipeline",
    ExpiresType: vergeos.APIKeyExpiresDate,
    Expires:     ptr(time.Now().AddDate(1, 0, 0).Unix()), // 1 year
})
fmt.Printf("Save this token (shown only once): %s\n", token)

// Create a non-expiring key with IP restrictions
key, token, err := client.UserAPIKeys.Create(ctx, &vergeos.UserAPIKeyCreateRequest{
    User:        userID,
    Name:        "restricted-key",
    ExpiresType: vergeos.APIKeyExpiresNever,
    IPAllowList: "10.0.0.0/8,192.168.1.0/24",
})

// Update an API key
key, err := client.UserAPIKeys.Update(ctx, keyID, &vergeos.UserAPIKeyUpdateRequest{
    Description: ptr("Updated description"),
    IPDenyList:  ptr("192.168.1.100"),
})

// Delete an API key
err = client.UserAPIKeys.Delete(ctx, keyID)
```

---

## Nodes

```go
// List all nodes
nodes, err := client.Nodes.List(ctx)

// Get a specific node
node, err := client.Nodes.Get(ctx, nodeID)

// List with options
nodes, err := client.Nodes.List(ctx,
    vergeos.WithFilter("physical eq true"),
    vergeos.WithFields("dashboard"),
)
```

---

## Clusters

```go
// List all clusters
clusters, err := client.Clusters.List(ctx)

// Get cluster details
cluster, err := client.Clusters.Get(ctx, clusterID)

// Get cluster status
status, err := client.Clusters.GetStatus(ctx, clusterID)
```

---

## Settings

```go
// Get system settings
settings, err := client.Settings.List(ctx)

// Get a specific setting by key
setting, err := client.Settings.GetByKey(ctx, "cloud_name")
fmt.Printf("Cloud: %s\n", setting.Value)

// Convenience method for cloud name
cloudName, err := client.Settings.GetCloudName(ctx)
```

---

## System

```go
// Get system version info (uses /version.json endpoint)
info, err := client.System.GetInfo(ctx)
fmt.Printf("API: %s, Version: %s\n", info.Name, info.Version)

// Get just the version string
version, err := client.System.GetVersion(ctx)
```

---

## Tags

Tags allow you to categorize and organize resources. Requires VergeOS v26+.

```go
// List all tags
tags, err := client.Tags.List(ctx)

// Get a tag by ID
tag, err := client.Tags.Get(ctx, tagID)

// Get a tag by name
tag, err := client.Tags.GetByName(ctx, "production")

// List tags in a specific category
tags, err := client.Tags.ListByCategory(ctx, categoryID)
```

---

## Tag Members

Tag members represent tag assignments to resources (VMs, networks, etc.).

```go
// List all tag assignments
members, err := client.TagMembers.List(ctx)

// List tags assigned to a specific VM
members, err := client.TagMembers.ListByMember(ctx, "vms/123")

// List all resources with a specific tag
members, err := client.TagMembers.ListByTag(ctx, tagID)

// Assign a tag to a VM
member, err := client.TagMembers.Assign(ctx, tagID, "vms/123")

// Remove a tag from a VM
err = client.TagMembers.Unassign(ctx, tagID, "vms/123")

// Or use Create/Delete directly
member, err := client.TagMembers.Create(ctx, &vergeos.TagMemberCreateRequest{
    Tag:    tagID,
    Member: "vms/123",  // format: "object_type/object_id"
})
err = client.TagMembers.Delete(ctx, memberID)
```

---

## Snapshot Profiles

Manage automated snapshot schedules for VMs, volumes, and system snapshots.

```go
// List all snapshot profiles
profiles, err := client.SnapshotProfiles.List(ctx)

// Get a profile by name
profile, err := client.SnapshotProfiles.GetByName(ctx, "daily-backups")

// Create a snapshot profile
profile, err := client.SnapshotProfiles.Create(ctx, &vergeos.SnapshotProfileCreateRequest{
    Name:        "daily-backups",
    Description: "Daily backup schedule",
})

// Update a profile
profile, err := client.SnapshotProfiles.Update(ctx, profileID, &vergeos.SnapshotProfileUpdateRequest{
    Description: ptr("Updated daily backup schedule"),
})

// Delete a profile
err = client.SnapshotProfiles.Delete(ctx, profileID)
```

---

## Snapshot Profile Periods

Define scheduling periods within snapshot profiles.

```go
// List periods for a profile
periods, err := client.SnapshotProfilePeriods.ListByProfile(ctx, profileID)

// Create a daily period (snapshots at 2:00 AM, retain for 7 days)
period, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
    Profile:   profileID,
    Name:      "daily",
    Frequency: vergeos.FrequencyDaily,
    Hour:      ptr(2),
    Minute:    ptr(0),
    Retention: 7 * 24 * 60 * 60, // 7 days in seconds
})

// Create a weekly period (Sundays at midnight, retain for 4 weeks)
period, err := client.SnapshotProfilePeriods.Create(ctx, &vergeos.SnapshotProfilePeriodCreateRequest{
    Profile:   profileID,
    Name:      "weekly",
    Frequency: vergeos.FrequencyWeekly,
    DayOfWeek: ptr(vergeos.DayOfWeekSunday),
    Hour:      ptr(0),
    Minute:    ptr(0),
    Retention: 28 * 24 * 60 * 60, // 4 weeks in seconds
})

// Update a period
period, err := client.SnapshotProfilePeriods.Update(ctx, periodID, &vergeos.SnapshotProfilePeriodUpdateRequest{
    Retention: ptr(14 * 24 * 60 * 60), // Extend to 14 days
    Quiesce:   ptr(true),              // Enable quiescing
})

// Delete a period
err = client.SnapshotProfilePeriods.Delete(ctx, periodID)
```

---

## Alarms

Monitor system health with alarms for VMs, nodes, networks, and other resources.

```go
// List all alarms
alarms, err := client.Alarms.List(ctx)

// List active (non-snoozed) alarms
alarms, err := client.Alarms.ListActive(ctx)

// List alarms for a specific resource
alarms, err := client.Alarms.ListByOwner(ctx, "vms/123")

// List alarms by severity level
alarms, err := client.Alarms.ListByLevel(ctx, vergeos.AlarmLevelCritical)

// Get an alarm by ID
alarm, err := client.Alarms.Get(ctx, alarmID)

// Snooze an alarm until a specific timestamp
err = client.Alarms.Snooze(ctx, alarmID, time.Now().Add(24*time.Hour).Unix())

// Unsnooze an alarm
err = client.Alarms.Unsnooze(ctx, alarmID)

// Resolve a resolvable alarm
err = client.Alarms.Resolve(ctx, alarmID)

// Delete an alarm
err = client.Alarms.Delete(ctx, alarmID)
```

---

## Alarm Types

Query alarm type definitions (read-only reference data).

```go
// List all alarm types
alarmTypes, err := client.AlarmTypes.List(ctx)

// Get an alarm type by key (string, not integer)
alarmType, err := client.AlarmTypes.Get(ctx, "vm_cpu_high")

// List alarm types by default severity level
alarmTypes, err := client.AlarmTypes.ListByLevel(ctx, vergeos.AlarmLevelWarning)
```

---

## Tasks

Manage scheduled and automated tasks.

```go
// List all tasks
tasks, err := client.Tasks.List(ctx)

// List running tasks
tasks, err := client.Tasks.ListRunning(ctx)

// List tasks for a specific resource
tasks, err := client.Tasks.ListByOwner(ctx, "vms/123")

// Get a task by ID
task, err := client.Tasks.Get(ctx, taskID)

// Get a task by its 40-character SHA1 ID
task, err := client.Tasks.GetByID(ctx, "abc123...")

// Create a task
task, err := client.Tasks.Create(ctx, &vergeos.TaskCreateRequest{
    Owner:  "vms/123",
    Action: "snapshot",
    Name:   "Daily Snapshot",
})

// Update a task
task, err := client.Tasks.Update(ctx, taskID, &vergeos.TaskUpdateRequest{
    Name: ptr("Updated Task Name"),
})

// Execute a task immediately
err = client.Tasks.Execute(ctx, taskID, nil)

// Enable/disable a task
err = client.Tasks.Enable(ctx, taskID)
err = client.Tasks.Disable(ctx, taskID)

// Delete a task
err = client.Tasks.Delete(ctx, taskID)
```

---

## Sites & DR

Manage remote site connections for disaster recovery and data replication.

```go
// List all sites
sites, err := client.Sites.List(ctx)

// Get a site by name
site, err := client.Sites.GetByName(ctx, "dr-site")

// Create a site connection
site, err := client.Sites.Create(ctx, &vergeos.SiteCreateRequest{
    URL:      "https://dr-site.example.com",
    Username: "sync-user",
    Password: "sync-password",
})

// Refresh site connection
err = client.Sites.Refresh(ctx, siteID)

// Run updates from remote site
err = client.Sites.RunUpdates(ctx, siteID)
```

---

## Cloud Snapshots

Manage cloud-level snapshots for backup and recovery.

```go
// List all cloud snapshots
snapshots, err := client.CloudSnapshots.List(ctx)

// Create a cloud snapshot
snapshot, err := client.CloudSnapshots.Create(ctx, &vergeos.CloudSnapshotCreateRequest{
    Name:        "pre-upgrade-snapshot",
    Description: ptr("Snapshot before system upgrade"),
})

// List VMs in a snapshot
vms, err := client.CloudSnapshotVMs.ListBySnapshot(ctx, snapshotID)

// List tenants in a snapshot
tenants, err := client.CloudSnapshotTenants.ListBySnapshot(ctx, snapshotID)

// Delete a snapshot
err = client.CloudSnapshots.Delete(ctx, snapshotID)
```

---

## Webhooks

Configure webhook endpoints for event notifications.

```go
// List all webhook URLs
webhooks, err := client.WebhookURLs.List(ctx)

// Create a webhook endpoint
webhook, err := client.WebhookURLs.Create(ctx, &vergeos.WebhookURLCreateRequest{
    Name:              "slack-alerts",
    URL:               "https://hooks.slack.com/services/xxx",
    AuthorizationType: vergeos.WebhookAuthNone,
    Timeout:           ptr(10),
    Retries:           ptr(3),
})

// Create a webhook with bearer token auth
webhook, err := client.WebhookURLs.Create(ctx, &vergeos.WebhookURLCreateRequest{
    Name:               "api-endpoint",
    URL:                "https://api.example.com/webhook",
    AuthorizationType:  vergeos.WebhookAuthBearer,
    AuthorizationValue: "your-bearer-token",
})

// Send a test message
err = client.WebhookURLs.Send(ctx, webhookID, `{"text": "Test message"}`)

// View webhook delivery log
messages, err := client.Webhooks.List(ctx)

// List failed deliveries
failed, err := client.Webhooks.ListFailed(ctx)

// Delete a webhook
err = client.WebhookURLs.Delete(ctx, webhookID)
```

---

## Additional Resources

The library also supports:

- **Files** (`client.Files`) - ISO images and media files
- **CloudInit Files** (`client.CloudInitFiles`) - Cloud-init configuration management
- **Resource Groups** (`client.ResourceGroups`) - Logical grouping of resources
- **VM Devices** (`client.VMDevices`) - VM devices (USB, TPM, vGPU)
