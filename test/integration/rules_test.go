//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/macstadium/govergeos"
)

// TestRulesList tests the VNetRules service list and get operations.
func TestRulesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing VNetRules service...")

	// List all firewall rules
	rules, err := client.VNetRules.List(ctx)
	if err != nil {
		t.Fatalf("VNetRules.List failed: %v", err)
	}

	t.Logf("Found %d firewall rules", len(rules))

	if len(rules) == 0 {
		t.Log("No firewall rules found")
		return
	}

	// Log first rule to verify field mapping
	first := rules[0]
	t.Logf("First rule: Key=%d, Name=%q, Protocol=%q, Direction=%q, Action=%q, Enabled=%v",
		int(first.Key), first.Name, first.Protocol, first.Direction, first.Action, first.Enabled)

	// Test Get
	t.Run("Get", func(t *testing.T) {
		if first.Key == 0 {
			t.Skip("No rule key available")
		}
		fetched, err := client.VNetRules.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetRules.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetRules.Get succeeded: Name=%q, SourceIP=%q, DestinationIP=%q",
				fetched.Name, fetched.SourceIP, fetched.DestinationIP)
		}
	})

	// Test ListByNetwork if we have a VNet ID
	t.Run("ListByNetwork", func(t *testing.T) {
		if first.VNet == 0 {
			t.Skip("No VNet ID available")
		}
		netRules, err := client.VNetRules.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetRules.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d rules in network %d", len(netRules), int(first.VNet))
		}
	})

	// Test GetByName within network
	t.Run("GetByName", func(t *testing.T) {
		if first.VNet == 0 || first.Name == "" {
			t.Skip("No VNet ID or name available")
		}
		byName, err := client.VNetRules.GetByName(ctx, int(first.VNet), first.Name)
		if err != nil {
			t.Errorf("VNetRules.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first rule for field verification
	prettyPrint(t, "Sample VNetRule", first)
}

// TestRuleAliasesList tests the VNetRuleAliases service list and get operations.
func TestRuleAliasesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing VNetRuleAliases service...")

	// List all rule aliases
	aliases, err := client.VNetRuleAliases.List(ctx)
	if err != nil {
		t.Fatalf("VNetRuleAliases.List failed: %v", err)
	}

	t.Logf("Found %d rule aliases", len(aliases))

	if len(aliases) == 0 {
		t.Log("No rule aliases found")
		return
	}

	// Log first alias to verify field mapping
	first := aliases[0]
	t.Logf("First alias: Key=%d, Name=%q, Value=%q, PublishingScope=%q",
		int(first.Key), first.Name, first.Value, first.PublishingScope)

	// Test Get
	t.Run("Get", func(t *testing.T) {
		if first.Key == 0 {
			t.Skip("No alias key available")
		}
		fetched, err := client.VNetRuleAliases.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetRuleAliases.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetRuleAliases.Get succeeded: Name=%q, Description=%q", fetched.Name, fetched.Description)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		if first.Name == "" {
			t.Skip("No alias name available")
		}
		byName, err := client.VNetRuleAliases.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("VNetRuleAliases.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	})

	// Pretty print first alias for field verification
	prettyPrint(t, "Sample VNetRuleAlias", first)
}

// TestRulesCRUD tests Create/Update/Delete operations for VNetRules and VNetRuleAliases.
// Creates a dedicated test network to avoid modifying production data.
func TestRulesCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test network for CRUD operations
	t.Log("Creating test network for CRUD tests...")
	testNetwork, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
		Name:        "sdk-rules-test-network",
		Description: "Temporary network for goVergeOS rules integration testing - safe to delete",
		Network:     "10.252.0.0/24",
		IPAddress:   "10.252.0.1",
		DHCPEnabled: ptr(false),
	})
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}
	networkID := int(testNetwork.ID)
	t.Logf("Created test network: %s (ID: %d)", testNetwork.Name, networkID)

	// Cleanup: delete test network when done
	defer func() {
		t.Log("Cleaning up: deleting test network...")
		if err := client.Networks.Delete(ctx, networkID); err != nil {
			t.Logf("Warning: failed to delete test network: %v", err)
		} else {
			t.Log("Test network deleted successfully")
		}
	}()

	// Test VNetRules CRUD
	t.Run("VNetRules", func(t *testing.T) {
		// Create
		rule, err := client.VNetRules.Create(ctx, &vergeos.VNetRuleCreateRequest{
			VNet:             networkID,
			Name:             "sdk-test-rule",
			Description:      "goVergeOS integration test rule - safe to delete",
			Protocol:         ptr("tcp"),
			Direction:        ptr("incoming"),
			DestinationPorts: ptr("8080"),
			Action:           ptr("accept"),
			Enabled:          ptr(true),
		})
		if err != nil {
			t.Fatalf("VNetRules.Create failed: %v", err)
		}
		ruleID := int(rule.Key)
		t.Logf("Created rule: [%d] %s (Protocol: %s, Ports: %s, Action: %s)",
			ruleID, rule.Name, rule.Protocol, rule.DestinationPorts, rule.Action)

		// Read
		rule, err = client.VNetRules.Get(ctx, ruleID)
		if err != nil {
			t.Fatalf("VNetRules.Get failed: %v", err)
		}
		t.Logf("Read rule: [%d] %s (Enabled: %v)", ruleID, rule.Name, rule.Enabled)

		// Update
		newDesc := "Updated goVergeOS test rule description"
		rule, err = client.VNetRules.Update(ctx, ruleID, &vergeos.VNetRuleUpdateRequest{
			Description:      &newDesc,
			DestinationPorts: ptr("8080,8443"),
		})
		if err != nil {
			t.Fatalf("VNetRules.Update failed: %v", err)
		}
		if rule.Description != newDesc {
			t.Errorf("Update verification failed: expected description %q, got %q", newDesc, rule.Description)
		} else {
			t.Logf("Updated rule: Description=%q, DestinationPorts=%q", rule.Description, rule.DestinationPorts)
		}

		// Test Disable
		err = client.VNetRules.Disable(ctx, ruleID, false)
		if err != nil {
			t.Errorf("VNetRules.Disable failed: %v", err)
		} else {
			rule, _ = client.VNetRules.Get(ctx, ruleID)
			if rule != nil && rule.Enabled {
				t.Error("Rule should be disabled")
			} else {
				t.Log("Disabled rule successfully")
			}
		}

		// Test Enable
		err = client.VNetRules.Enable(ctx, ruleID, false)
		if err != nil {
			t.Errorf("VNetRules.Enable failed: %v", err)
		} else {
			rule, _ = client.VNetRules.Get(ctx, ruleID)
			if rule != nil && !rule.Enabled {
				t.Error("Rule should be enabled")
			} else {
				t.Log("Enabled rule successfully")
			}
		}

		// Delete
		err = client.VNetRules.Delete(ctx, ruleID)
		if err != nil {
			t.Fatalf("VNetRules.Delete failed: %v", err)
		}
		t.Log("Deleted rule successfully")

		// Verify deletion
		_, err = client.VNetRules.Get(ctx, ruleID)
		if err == nil {
			t.Error("Expected error after deletion, but got none")
		} else if vergeos.IsNotFoundError(err) {
			t.Log("Verified: rule correctly deleted (NotFoundError)")
		} else {
			t.Logf("Got error after deletion: %v", err)
		}
	})

	// Test VNetRuleAliases CRUD
	t.Run("VNetRuleAliases", func(t *testing.T) {
		// Create
		alias, err := client.VNetRuleAliases.Create(ctx, &vergeos.VNetRuleAliasCreateRequest{
			Name:        "sdk-test-alias",
			Description: "goVergeOS integration test alias - safe to delete",
			Value:       "192.168.100.0/24,10.100.0.0/16",
		})
		if err != nil {
			t.Fatalf("VNetRuleAliases.Create failed: %v", err)
		}
		aliasID := int(alias.Key)
		t.Logf("Created alias: [%d] %s (Value: %s)", aliasID, alias.Name, alias.Value)

		// Read
		alias, err = client.VNetRuleAliases.Get(ctx, aliasID)
		if err != nil {
			t.Fatalf("VNetRuleAliases.Get failed: %v", err)
		}
		t.Logf("Read alias: [%d] %s", aliasID, alias.Name)

		// Update
		newValue := "192.168.100.0/24,10.100.0.0/16,172.16.0.0/12"
		alias, err = client.VNetRuleAliases.Update(ctx, aliasID, &vergeos.VNetRuleAliasUpdateRequest{
			Value: &newValue,
		})
		if err != nil {
			t.Fatalf("VNetRuleAliases.Update failed: %v", err)
		}
		if alias.Value != newValue {
			t.Errorf("Update verification failed: expected value %q, got %q", newValue, alias.Value)
		} else {
			t.Logf("Updated alias value to: %s", alias.Value)
		}

		// Delete
		err = client.VNetRuleAliases.Delete(ctx, aliasID)
		if err != nil {
			t.Fatalf("VNetRuleAliases.Delete failed: %v", err)
		}
		t.Log("Deleted alias successfully")

		// Verify deletion
		_, err = client.VNetRuleAliases.Get(ctx, aliasID)
		if err == nil {
			t.Error("Expected error after deletion, but got none")
		} else if vergeos.IsNotFoundError(err) {
			t.Log("Verified: alias correctly deleted (NotFoundError)")
		} else {
			t.Logf("Got error after deletion: %v", err)
		}
	})
}
