// Example: Networking - IP Addresses, DNS, and Host Overrides
//
// This example demonstrates how to:
// - List and manage network IP addresses
// - Query DNS views, zones, and records
// - Manage host overrides for DNS/DHCP
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/networking/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/verge-io/goVergeOS"
)

func main() {
	// Create client from environment variables
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(os.Getenv("VERGEOS_HOST")),
		vergeos.WithCredentials(os.Getenv("VERGEOS_USERNAME"), os.Getenv("VERGEOS_PASSWORD")),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Run examples
	fmt.Println("=== Networks ===")
	networkID := showNetworks(ctx, client)

	fmt.Println("\n=== Network Addresses ===")
	showAddresses(ctx, client, networkID)

	fmt.Println("\n=== DNS Views ===")
	viewID := showDNSViews(ctx, client)

	fmt.Println("\n=== DNS Zones ===")
	zoneID := showDNSZones(ctx, client, viewID)

	fmt.Println("\n=== DNS Records ===")
	showDNSRecords(ctx, client, zoneID)

	fmt.Println("\n=== Host Overrides ===")
	showHosts(ctx, client, networkID)
}

func showNetworks(ctx context.Context, client *vergeos.Client) int {
	// List networks to find one for demonstration
	networks, err := client.Networks.List(ctx, vergeos.WithLimit(10))
	if err != nil {
		log.Printf("Failed to list networks: %v", err)
		return 0
	}

	fmt.Printf("Found %d networks\n", len(networks))

	if len(networks) == 0 {
		return 0
	}

	fmt.Println("\nAvailable networks:")
	for _, n := range networks {
		fmt.Printf("  - [%d] %s (%s)\n", n.ID, n.Name, n.Network)
	}

	// Return first network ID for further examples
	return int(networks[0].ID)
}

func showAddresses(ctx context.Context, client *vergeos.Client, networkID int) {
	// List all addresses
	addresses, err := client.VNetAddresses.List(ctx, vergeos.WithLimit(20))
	if err != nil {
		log.Printf("Failed to list addresses: %v", err)
		return
	}

	fmt.Printf("Found %d total addresses\n", len(addresses))

	if len(addresses) == 0 {
		fmt.Println("No IP addresses configured")
		return
	}

	// Group by type
	byType := make(map[string]int)
	for _, a := range addresses {
		byType[a.Type]++
	}
	fmt.Println("\nAddresses by type:")
	for addrType, count := range byType {
		fmt.Printf("  %s: %d\n", addrType, count)
	}

	// Show addresses for specific network if provided
	if networkID > 0 {
		netAddrs, err := client.VNetAddresses.ListByNetwork(ctx, networkID)
		if err != nil {
			log.Printf("Failed to list addresses for network %d: %v", networkID, err)
		} else {
			fmt.Printf("\nAddresses in network %d: %d\n", networkID, len(netAddrs))
			limit := 10
			if len(netAddrs) < limit {
				limit = len(netAddrs)
			}
			for i := 0; i < limit; i++ {
				a := netAddrs[i]
				hostname := a.Hostname
				if hostname == "" {
					hostname = "(none)"
				}
				fmt.Printf("  - %s [%s] MAC: %s, Host: %s\n", a.IP, a.Type, a.MAC, hostname)
			}
			if len(netAddrs) > limit {
				fmt.Printf("  ... and %d more\n", len(netAddrs)-limit)
			}
		}

		// List only static addresses
		staticAddrs, err := client.VNetAddresses.ListByType(ctx, networkID, vergeos.AddressTypeStatic)
		if err != nil {
			log.Printf("Failed to list static addresses: %v", err)
		} else {
			fmt.Printf("\nStatic addresses in network %d: %d\n", networkID, len(staticAddrs))
		}
	}

	// Show first address details
	if len(addresses) > 0 {
		first := addresses[0]
		fmt.Printf("\nFirst address details:\n")
		fmt.Printf("  Key: %d\n", int(first.Key))
		fmt.Printf("  IP: %s\n", first.IP)
		fmt.Printf("  Type: %s\n", first.Type)
		fmt.Printf("  MAC: %s\n", first.MAC)
		fmt.Printf("  Hostname: %s\n", first.Hostname)
		fmt.Printf("  VNet: %d\n", int(first.VNet))
		fmt.Printf("  Owner: %s\n", first.Owner)
		fmt.Printf("  Vendor: %s\n", first.Vendor)
		if first.Expiration > 0 {
			fmt.Printf("  Expires: %s\n", time.Unix(first.Expiration, 0).Format(time.RFC3339))
		}
	}

	// Example: Create a static address (commented out to avoid modifying data)
	// addr, err := client.VNetAddresses.Create(ctx, &vergeos.VNetAddressCreateRequest{
	// 	VNet:     networkID,
	// 	Type:     vergeos.AddressTypeStatic,
	// 	IP:       "10.0.0.100",
	// 	Hostname: "my-server",
	// })
}

func showDNSViews(ctx context.Context, client *vergeos.Client) int {
	// List all DNS views
	views, err := client.VNetDNSViews.List(ctx)
	if err != nil {
		log.Printf("Failed to list DNS views: %v", err)
		return 0
	}

	fmt.Printf("Found %d DNS views\n", len(views))

	if len(views) == 0 {
		fmt.Println("No DNS views configured")
		return 0
	}

	fmt.Println("\nDNS views:")
	for _, v := range views {
		recursion := "no"
		if v.Recursion {
			recursion = "yes"
		}
		fmt.Printf("  - [%d] %s (VNet: %d, Recursion: %s)\n", v.Key, v.Name, v.VNet, recursion)
	}

	// Show first view details
	first := views[0]
	fmt.Printf("\nFirst view details:\n")
	fmt.Printf("  Key: %d\n", int(first.Key))
	fmt.Printf("  Name: %s\n", first.Name)
	fmt.Printf("  VNet: %d\n", int(first.VNet))
	fmt.Printf("  Recursion: %v\n", first.Recursion)
	fmt.Printf("  Match Clients: %s\n", first.MatchClients)
	fmt.Printf("  Match Destinations: %s\n", first.MatchDestinations)
	fmt.Printf("  Max Cache Size: %d bytes\n", first.MaxCacheSize)

	return int(first.Key)
}

func showDNSZones(ctx context.Context, client *vergeos.Client, viewID int) int {
	// List all DNS zones
	zones, err := client.VNetDNSZones.List(ctx)
	if err != nil {
		log.Printf("Failed to list DNS zones: %v", err)
		return 0
	}

	fmt.Printf("Found %d DNS zones\n", len(zones))

	if len(zones) == 0 {
		fmt.Println("No DNS zones configured")
		return 0
	}

	// Group by type
	byType := make(map[string]int)
	for _, z := range zones {
		byType[z.Type]++
	}
	fmt.Println("\nZones by type:")
	for zoneType, count := range byType {
		fmt.Printf("  %s: %d\n", zoneType, count)
	}

	fmt.Println("\nDNS zones:")
	for _, z := range zones {
		fmt.Printf("  - [%d] %s (%s, View: %d)\n", z.Key, z.Domain, z.Type, z.View)
	}

	// List zones for specific view if provided
	if viewID > 0 {
		viewZones, err := client.VNetDNSZones.ListByView(ctx, viewID)
		if err != nil {
			log.Printf("Failed to list zones for view %d: %v", viewID, err)
		} else {
			fmt.Printf("\nZones in view %d: %d\n", viewID, len(viewZones))
		}
	}

	// Show first zone details
	first := zones[0]
	fmt.Printf("\nFirst zone details:\n")
	fmt.Printf("  Key: %d\n", int(first.Key))
	fmt.Printf("  Domain: %s\n", first.Domain)
	fmt.Printf("  Type: %s\n", first.Type)
	fmt.Printf("  View: %d\n", int(first.View))
	fmt.Printf("  Nameserver: %s\n", first.Nameserver)
	fmt.Printf("  Email: %s\n", first.Email)
	fmt.Printf("  Serial Number: %d\n", first.SerialNumber)
	fmt.Printf("  Default TTL: %s\n", first.DefaultTTL)
	fmt.Printf("  Refresh: %s\n", first.RefreshInterval)
	fmt.Printf("  Retry: %s\n", first.RetryInterval)
	fmt.Printf("  Expiry: %s\n", first.ExpiryPeriod)
	fmt.Printf("  Negative TTL: %s\n", first.NegativeTTL)

	return int(first.Key)
}

func showDNSRecords(ctx context.Context, client *vergeos.Client, zoneID int) {
	// List all DNS records
	records, err := client.VNetDNSRecords.List(ctx, vergeos.WithLimit(50))
	if err != nil {
		log.Printf("Failed to list DNS records: %v", err)
		return
	}

	fmt.Printf("Found %d DNS records\n", len(records))

	if len(records) == 0 {
		fmt.Println("No DNS records configured")
		return
	}

	// Group by type
	byType := make(map[string]int)
	for _, r := range records {
		byType[r.Type]++
	}
	fmt.Println("\nRecords by type:")
	for recType, count := range byType {
		fmt.Printf("  %s: %d\n", recType, count)
	}

	// List records for specific zone if provided
	if zoneID > 0 {
		zoneRecords, err := client.VNetDNSRecords.ListByZone(ctx, zoneID)
		if err != nil {
			log.Printf("Failed to list records for zone %d: %v", zoneID, err)
		} else {
			fmt.Printf("\nRecords in zone %d:\n", zoneID)
			for _, r := range zoneRecords {
				host := r.Host
				if host == "" {
					host = "@"
				}
				extra := ""
				if r.Type == vergeos.DNSRecordTypeMX && r.MXPreference > 0 {
					extra = fmt.Sprintf(" (pref: %d)", r.MXPreference)
				}
				if r.Type == vergeos.DNSRecordTypeSRV {
					extra = fmt.Sprintf(" (weight: %d, port: %d)", r.Weight, r.Port)
				}
				fmt.Printf("  - %s %s -> %s%s\n", host, r.Type, r.Value, extra)
			}
		}

		// List only A records
		aRecords, err := client.VNetDNSRecords.ListByType(ctx, zoneID, vergeos.DNSRecordTypeA)
		if err != nil {
			log.Printf("Failed to list A records: %v", err)
		} else {
			fmt.Printf("\nA records in zone %d: %d\n", zoneID, len(aRecords))
		}
	}

	// Show first record details
	if len(records) > 0 {
		first := records[0]
		fmt.Printf("\nFirst record details:\n")
		fmt.Printf("  Key: %d\n", int(first.Key))
		fmt.Printf("  Zone: %d\n", int(first.Zone))
		fmt.Printf("  Host: %s\n", first.Host)
		fmt.Printf("  Type: %s\n", first.Type)
		fmt.Printf("  Value: %s\n", first.Value)
		fmt.Printf("  TTL: %s\n", first.TTL)
		if first.MXPreference > 0 {
			fmt.Printf("  MX Preference: %d\n", first.MXPreference)
		}
	}

	// Example: Create a DNS record (commented out to avoid modifying data)
	// record, err := client.VNetDNSRecords.Create(ctx, &vergeos.VNetDNSRecordCreateRequest{
	// 	Zone:  zoneID,
	// 	Host:  "www",
	// 	Type:  vergeos.DNSRecordTypeA,
	// 	Value: "192.168.1.100",
	// 	TTL:   ptr("1h"),
	// })
}

func showHosts(ctx context.Context, client *vergeos.Client, networkID int) {
	// List all host overrides
	hosts, err := client.VNetHosts.List(ctx)
	if err != nil {
		log.Printf("Failed to list hosts: %v", err)
		return
	}

	fmt.Printf("Found %d host overrides\n", len(hosts))

	if len(hosts) == 0 {
		fmt.Println("No host overrides configured")
		return
	}

	// Group by type
	byType := make(map[string]int)
	for _, h := range hosts {
		byType[h.Type]++
	}
	fmt.Println("\nHosts by type:")
	for hostType, count := range byType {
		fmt.Printf("  %s: %d\n", hostType, count)
	}

	fmt.Println("\nHost overrides:")
	for _, h := range hosts {
		fmt.Printf("  - [%d] %s -> %s (type: %s, VNet: %d)\n", h.Key, h.Host, h.IP, h.Type, h.VNet)
	}

	// List hosts for specific network if provided
	if networkID > 0 {
		netHosts, err := client.VNetHosts.ListByNetwork(ctx, networkID)
		if err != nil {
			log.Printf("Failed to list hosts for network %d: %v", networkID, err)
		} else {
			fmt.Printf("\nHosts in network %d: %d\n", networkID, len(netHosts))
		}
	}

	// Show first host details
	if len(hosts) > 0 {
		first := hosts[0]
		fmt.Printf("\nFirst host details:\n")
		fmt.Printf("  Key: %d\n", int(first.Key))
		fmt.Printf("  Host: %s\n", first.Host)
		fmt.Printf("  IP: %s\n", first.IP)
		fmt.Printf("  Type: %s\n", first.Type)
		fmt.Printf("  VNet: %d\n", int(first.VNet))
	}

	// Example: Create a host override (commented out to avoid modifying data)
	// host, err := client.VNetHosts.Create(ctx, &vergeos.VNetHostCreateRequest{
	// 	VNet: networkID,
	// 	Host: "printer.local",
	// 	IP:   "192.168.1.50",
	// })
}
