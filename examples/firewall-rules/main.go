// Example: Firewall Rules
//
// This example demonstrates how to work with network firewall rules
// in VergeOS. Firewall rules control traffic flow on virtual networks.
//
// Usage:
//
//	export VERGEOS_HOST=https://your-vergeos-host
//	export VERGEOS_USERNAME=admin
//	export VERGEOS_PASSWORD=yourpassword
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	vergeos "github.com/verge-io/goVergeOS"
)

func main() {
	// Get configuration from environment
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		log.Fatal("Please set VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD environment variables")
	}

	// Create client
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true), // For self-signed certificates
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// List all firewall rules
	fmt.Println("=== All Firewall Rules ===")
	rules, err := client.VNetRules.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list rules: %v", err)
	}

	fmt.Printf("Found %d firewall rules\n\n", len(rules))

	// Group rules by network
	networkRules := make(map[int][]vergeos.VNetRule)
	for _, rule := range rules {
		vnetID := rule.VNet.Int()
		networkRules[vnetID] = append(networkRules[vnetID], rule)
	}

	// Display rules grouped by network
	for vnetID, vnetRules := range networkRules {
		fmt.Printf("Network %d (%d rules):\n", vnetID, len(vnetRules))
		for _, rule := range vnetRules {
			status := "enabled"
			if !rule.Enabled {
				status = "disabled"
			}
			sysRule := ""
			if rule.SystemRule {
				sysRule = " [system]"
			}
			fmt.Printf("  [%d] %s (%s)%s\n", rule.Key.Int(), rule.Name, status, sysRule)
			fmt.Printf("      %s %s -> %s\n", rule.Direction, rule.Protocol, rule.Action)
			if rule.SourceIP != "" || rule.SourcePorts != "" {
				fmt.Printf("      Source: %s:%s\n", rule.SourceIP, rule.SourcePorts)
			}
			if rule.DestinationIP != "" || rule.DestinationPorts != "" {
				fmt.Printf("      Dest: %s:%s\n", rule.DestinationIP, rule.DestinationPorts)
			}
			if rule.TargetIP != "" || rule.TargetPorts != "" {
				fmt.Printf("      Target: %s:%s\n", rule.TargetIP, rule.TargetPorts)
			}
		}
		fmt.Println()
	}

	// Get rules for a specific network (using first network found)
	if len(rules) > 0 {
		vnetID := rules[0].VNet.Int()
		fmt.Printf("=== Rules for Network %d ===\n", vnetID)
		vnetRules, err := client.VNetRules.ListByNetwork(ctx, vnetID)
		if err != nil {
			log.Printf("Failed to get rules for network: %v", err)
		} else {
			fmt.Printf("Found %d rules for network %d\n", len(vnetRules), vnetID)
			for _, r := range vnetRules {
				fmt.Printf("  Order %d: %s (%s %s -> %s)\n",
					r.OrderID, r.Name, r.Direction, r.Protocol, r.Action)
			}
		}
	}

	// List rule aliases
	fmt.Println("\n=== Rule Aliases ===")
	aliases, err := client.VNetRuleAliases.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list aliases: %v", err)
	}

	if len(aliases) == 0 {
		fmt.Println("No rule aliases defined.")
		fmt.Println("Rule aliases allow you to define reusable address lists for firewall rules.")
	} else {
		for _, alias := range aliases {
			fmt.Printf("  %s: %s\n", alias.Name, alias.Value)
			if alias.Description != "" {
				fmt.Printf("    Description: %s\n", alias.Description)
			}
		}
	}

	// Demonstrate rule lookup by name
	if len(rules) > 0 {
		fmt.Println("\n=== Rule Lookup Example ===")
		rule := rules[0]
		vnetID := rule.VNet.Int()

		foundRule, err := client.VNetRules.GetByName(ctx, vnetID, rule.Name)
		if err != nil {
			log.Printf("Failed to get rule by name: %v", err)
		} else {
			fmt.Printf("Found rule '%s' (ID: %d) on network %d\n",
				foundRule.Name, foundRule.Key.Int(), foundRule.VNet.Int())
			fmt.Printf("  Protocol: %s, Direction: %s, Action: %s\n",
				foundRule.Protocol, foundRule.Direction, foundRule.Action)
		}
	}

	fmt.Println("\nDone!")
}
