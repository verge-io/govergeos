---
title: Networking
description: Manage virtual networks, firewall rules, rule aliases, IP addresses, DNS, and host overrides
tags: [network, vnet, firewall, rule, alias, address, dhcp, dns, dns-view, dns-zone, dns-record, host-override, diagnostics, statistics, ping, traceroute]
categories: [Networking]
---

# Networking

Manage virtual networks, firewall rules, DNS, and IP addressing.

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

// Network Diagnostics - run diagnostic queries on a running network
// Run a ping test
result, err := client.Networks.Ping(ctx, networkID, "8.8.8.8", 4) // 4 pings
fmt.Println(result.Result)

// Run a traceroute
result, err := client.Networks.Traceroute(ctx, networkID, "google.com")
fmt.Println(result.Result)

// Run a DNS lookup
result, err := client.Networks.DNSLookup(ctx, networkID, "example.com")
fmt.Println(result.Result)

// Run a custom diagnostic query
query, err := client.Networks.RunQueryWait(ctx, &vergeos.NetworkQueryRequest{
    VNet:  networkID,
    Query: vergeos.NetworkQueryFirewall,  // Show firewall rules
})
fmt.Printf("Query status: %s\nResult:\n%s\n", query.Status, query.Result)

// Available query types:
// NetworkQueryPing, NetworkQueryTraceroute, NetworkQueryDNS
// NetworkQueryTCPDump, NetworkQueryFirewall, NetworkQueryARP
// NetworkQueryWhatsMyIP, NetworkQueryNmap, NetworkQueryTCPConnect
// NetworkQueryWireGuard, NetworkQueryIPSec, and more

// Network Statistics - get monitoring stats (requires monitor_gateway=true)
stats, err := client.Networks.GetStatistics(ctx, networkID)
for _, s := range stats {
    fmt.Printf("Quality: %d%%, Latency: %dms, Dropped: %d\n",
        s.Quality, s.LatencyUSAvg/1000, s.Dropped)
}

// Get just the latest statistics
latest, err := client.Networks.GetLatestStatistics(ctx, networkID)
if latest != nil {
    fmt.Printf("Current quality: %d%%\n", latest.Quality)
}
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
