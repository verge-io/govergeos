//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/goVergeOS"
)

// TestWave1StorageFirewall tests the Wave 1 services (Volumes, VNet Rules, Rule Aliases)
// against a live VergeOS API to verify field mappings are correct.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave1
func TestWave1StorageFirewall(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("Volumes", func(t *testing.T) {
		testVolumes(t, ctx, client)
	})

	t.Run("VNetRules", func(t *testing.T) {
		testVNetRules(t, ctx, client)
	})

	t.Run("VNetRuleAliases", func(t *testing.T) {
		testVNetRuleAliases(t, ctx, client)
	})
}

func testVolumes(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Volumes service...")

	// List all volumes
	volumes, err := client.Volumes.List(ctx)
	if err != nil {
		t.Fatalf("Volumes.List failed: %v", err)
	}

	t.Logf("Found %d volumes", len(volumes))

	if len(volumes) == 0 {
		t.Log("No volumes found - skipping Get tests (requires NAS service)")
		return
	}

	// Log first volume to verify field mapping
	first := volumes[0]
	t.Logf("First volume: Key=%q, Name=%q, Enabled=%v, FSType=%q, Service=%d",
		first.Key, first.Name, first.Enabled, first.FSType, int(first.Service))

	// Verify Key is a non-empty string (volumes use SHA1 hash string keys)
	if first.Key == "" {
		t.Error("Volume.Key is empty - expected SHA1 hash string")
	}

	// Test Get by string key
	if first.Key != "" {
		fetched, err := client.Volumes.Get(ctx, first.Key)
		if err != nil {
			t.Errorf("Volumes.Get(%q) failed: %v", first.Key, err)
		} else {
			t.Logf("Volumes.Get succeeded: Key=%q, Name=%q, MaxSize=%d", fetched.Key, fetched.Name, fetched.MaxSize)
		}
	}

	// Test ListByService if we have a service ID
	if first.Service > 0 {
		serviceVolumes, err := client.Volumes.ListByService(ctx, int(first.Service))
		if err != nil {
			t.Errorf("Volumes.ListByService failed: %v", err)
		} else {
			t.Logf("Found %d volumes in service %d", len(serviceVolumes), int(first.Service))
		}

		// Test GetByName within service
		if first.Name != "" {
			byName, err := client.Volumes.GetByName(ctx, int(first.Service), first.Name)
			if err != nil {
				t.Errorf("Volumes.GetByName failed: %v", err)
			} else {
				t.Logf("GetByName succeeded: Key=%q", byName.Key)
			}
		}
	}

	// Pretty print first volume for field verification
	prettyPrint(t, "Sample Volume", first)
}

func testVNetRules(t *testing.T, ctx context.Context, client *vergeos.Client) {
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
	if first.Key > 0 {
		fetched, err := client.VNetRules.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetRules.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetRules.Get succeeded: Name=%q, SourceIP=%q, DestinationIP=%q",
				fetched.Name, fetched.SourceIP, fetched.DestinationIP)
		}
	}

	// Test ListByNetwork if we have a VNet ID
	if first.VNet > 0 {
		netRules, err := client.VNetRules.ListByNetwork(ctx, int(first.VNet))
		if err != nil {
			t.Errorf("VNetRules.ListByNetwork failed: %v", err)
		} else {
			t.Logf("Found %d rules in network %d", len(netRules), int(first.VNet))
		}

		// Test GetByName within network
		if first.Name != "" {
			byName, err := client.VNetRules.GetByName(ctx, int(first.VNet), first.Name)
			if err != nil {
				t.Errorf("VNetRules.GetByName failed: %v", err)
			} else {
				t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
			}
		}
	}

	// Pretty print first rule for field verification
	prettyPrint(t, "Sample VNetRule", first)
}

func testVNetRuleAliases(t *testing.T, ctx context.Context, client *vergeos.Client) {
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
	if first.Key > 0 {
		fetched, err := client.VNetRuleAliases.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("VNetRuleAliases.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("VNetRuleAliases.Get succeeded: Name=%q, Description=%q", fetched.Name, fetched.Description)
		}
	}

	// Test GetByName
	if first.Name != "" {
		byName, err := client.VNetRuleAliases.GetByName(ctx, first.Name)
		if err != nil {
			t.Errorf("VNetRuleAliases.GetByName failed: %v", err)
		} else {
			t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
		}
	}

	// Pretty print first alias for field verification
	prettyPrint(t, "Sample VNetRuleAlias", first)
}

// TestWave1FirewallCRUD tests Create/Update/Delete operations for Wave 1 firewall services.
// This test creates a dedicated test network to avoid modifying production data.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave1FirewallCRUD
func TestWave1FirewallCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test network for CRUD operations
	t.Log("Creating test network for CRUD tests...")
	testNetwork, err := client.Networks.Create(ctx, &vergeos.NetworkCreateRequest{
		Name:        "sdk-wave1-test-network",
		Description: "Temporary network for Wave 1 goVergeOS integration testing - safe to delete",
		Network:     "10.251.0.0/24",
		IPAddress:   "10.251.0.1",
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

	// Run CRUD subtests
	t.Run("VNetRulesCRUD", func(t *testing.T) {
		testVNetRulesCRUD(t, ctx, client, networkID)
	})

	t.Run("VNetRuleAliasesCRUD", func(t *testing.T) {
		testVNetRuleAliasesCRUD(t, ctx, client)
	})
}

func testVNetRulesCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, networkID int) {
	t.Log("Testing VNetRules CRUD...")

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

	// Test Enable/Disable (these update the enabled field via PUT)
	err = client.VNetRules.Disable(ctx, ruleID, false)
	if err != nil {
		t.Errorf("VNetRules.Disable failed: %v", err)
	} else {
		t.Log("Disabled rule successfully")
	}

	// Verify disabled
	rule, err = client.VNetRules.Get(ctx, ruleID)
	if err != nil {
		t.Errorf("Failed to verify disabled state: %v", err)
	} else if rule.Enabled {
		t.Error("Rule should be disabled but Enabled=true")
	} else {
		t.Log("Verified rule is disabled")
	}

	err = client.VNetRules.Enable(ctx, ruleID, false)
	if err != nil {
		t.Errorf("VNetRules.Enable failed: %v", err)
	} else {
		t.Log("Enabled rule successfully")
	}

	// Verify enabled
	rule, err = client.VNetRules.Get(ctx, ruleID)
	if err != nil {
		t.Errorf("Failed to verify enabled state: %v", err)
	} else if !rule.Enabled {
		t.Error("Rule should be enabled but Enabled=false")
	} else {
		t.Log("Verified rule is enabled")
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
	} else if !vergeos.IsNotFoundError(err) {
		t.Logf("Got expected error after deletion: %v", err)
	} else {
		t.Log("Verified: rule correctly deleted (NotFoundError)")
	}
}

func testVNetRuleAliasesCRUD(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing VNetRuleAliases CRUD...")

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
	} else if !vergeos.IsNotFoundError(err) {
		t.Logf("Got expected error after deletion: %v", err)
	} else {
		t.Log("Verified: alias correctly deleted (NotFoundError)")
	}
}

// ptr is defined in wave5_networking_test.go
