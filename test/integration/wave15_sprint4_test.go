//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestWave15Sprint4 tests the Sprint 4 services:
// - Cluster CRUD (Create/Update/Delete)
// - Network Diagnostics and Statistics
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestWave15
func TestWave15Sprint4(t *testing.T) {
	client := setupTestClientWave15(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("ClusterCRUD", func(t *testing.T) {
		testClusterCRUD(t, ctx, client)
	})

	t.Run("NetworkDiagnosticsStatistics", func(t *testing.T) {
		testNetworkDiagnosticsStatistics(t, ctx, client)
	})
}

func testClusterCRUD(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Cluster CRUD operations...")

	// List existing clusters
	clusters, err := client.Clusters.List(ctx)
	if err != nil {
		t.Fatalf("Clusters.List failed: %v", err)
	}
	t.Logf("Found %d clusters", len(clusters))

	if len(clusters) == 0 {
		t.Log("No clusters available - skipping detailed tests")
		return
	}

	prettyPrintWave15(t, "First Cluster", clusters[0])

	// Test Get by ID
	first := clusters[0]
	t.Run("Get", func(t *testing.T) {
		cluster, err := client.Clusters.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("Clusters.Get failed: %v", err)
		}
		if int(cluster.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", cluster.Key, first.Key)
		}
	})

	// Test GetByName
	t.Run("GetByName", func(t *testing.T) {
		cluster, err := client.Clusters.GetByName(ctx, first.Name)
		if err != nil {
			t.Fatalf("Clusters.GetByName failed: %v", err)
		}
		if int(cluster.Key) != int(first.Key) {
			t.Errorf("Key mismatch: got %d, want %d", cluster.Key, first.Key)
		}
	})

	// Test GetStatus
	t.Run("GetStatus", func(t *testing.T) {
		status, err := client.Clusters.GetStatus(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("Clusters.GetStatus failed: %v", err)
		}
		prettyPrintWave15(t, "Cluster Status", status)
	})

	// Test Update - safely update the description and revert it
	t.Run("Update", func(t *testing.T) {
		clusterID := int(first.Key)
		originalDesc := first.Description

		// Update description
		newDesc := "SDK test update - " + time.Now().Format(time.RFC3339)
		updated, err := client.Clusters.Update(ctx, clusterID, &vergeos.ClusterUpdateRequest{
			Description: &newDesc,
		})
		if err != nil {
			t.Fatalf("Clusters.Update failed: %v", err)
		}
		if updated.Description != newDesc {
			t.Errorf("Description not updated: got %q, want %q", updated.Description, newDesc)
		}
		t.Logf("Updated cluster description to: %s", newDesc)

		// Revert to original description
		reverted, err := client.Clusters.Update(ctx, clusterID, &vergeos.ClusterUpdateRequest{
			Description: &originalDesc,
		})
		if err != nil {
			t.Fatalf("Failed to revert description: %v", err)
		}
		if reverted.Description != originalDesc {
			t.Errorf("Description not reverted: got %q, want %q", reverted.Description, originalDesc)
		}
		t.Logf("Reverted cluster description to: %s", originalDesc)
	})

	// Test Create/Delete - Skip on live systems unless explicitly enabled
	// This test is potentially disruptive as it creates/deletes a cluster
	if os.Getenv("VERGEOS_TEST_CREATE_CLUSTER") != "" {
		t.Run("CreateDelete", func(t *testing.T) {
			testClusterCreateDelete(t, ctx, client)
		})
	} else {
		t.Log("Skipping Create/Delete test - set VERGEOS_TEST_CREATE_CLUSTER=1 to enable")
	}
}

func testClusterCreateDelete(t *testing.T, ctx context.Context, client *vergeos.Client) {
	// Create a test cluster
	clusterName := "sdk-test-cluster-" + time.Now().Format("20060102-150405")
	req := &vergeos.ClusterCreateRequest{
		Name:        clusterName,
		Description: "Test cluster created by goVergeOS SDK integration tests",
	}

	t.Logf("Creating test cluster: %s", clusterName)
	cluster, err := client.Clusters.Create(ctx, req)
	if err != nil {
		t.Fatalf("Clusters.Create failed: %v", err)
	}
	t.Logf("Created cluster with ID: %d", cluster.Key)
	prettyPrintWave15(t, "Created Cluster", cluster)

	// Clean up - delete the cluster
	defer func() {
		t.Logf("Deleting test cluster: %d", cluster.Key)
		if err := client.Clusters.Delete(ctx, int(cluster.Key)); err != nil {
			t.Errorf("Failed to delete test cluster: %v", err)
		} else {
			t.Logf("Successfully deleted test cluster")
		}
	}()

	// Verify we can get the cluster
	retrieved, err := client.Clusters.Get(ctx, int(cluster.Key))
	if err != nil {
		t.Fatalf("Failed to retrieve created cluster: %v", err)
	}
	if retrieved.Name != clusterName {
		t.Errorf("Name mismatch: got %q, want %q", retrieved.Name, clusterName)
	}
}

func testNetworkDiagnosticsStatistics(t *testing.T, ctx context.Context, client *vergeos.Client) {
	t.Log("Testing Network Diagnostics and Statistics...")

	// List networks
	networks, err := client.Networks.List(ctx)
	if err != nil {
		t.Fatalf("Networks.List failed: %v", err)
	}
	t.Logf("Found %d networks", len(networks))

	if len(networks) == 0 {
		t.Log("No networks available - skipping diagnostics/statistics tests")
		return
	}

	// Find a running external/internal network for diagnostics (running networks have PowerState=true)
	// Prefer external networks as they have internet connectivity
	var runningNetwork *vergeos.Network
	for _, net := range networks {
		if net.PowerState && (net.Type == "external" || net.Type == "internal") {
			runningNetwork = &net
			if net.Type == "external" {
				break // Prefer external
			}
		}
	}

	if runningNetwork == nil {
		t.Log("No suitable running networks found - diagnostics require running networks")
		return
	}

	t.Logf("Using running network for tests: %s (ID: %d, Type: %s)", runningNetwork.Name, runningNetwork.ID, runningNetwork.Type)

	// Test GetDiagnostics (WhatsMyIP query)
	t.Run("GetDiagnostics", func(t *testing.T) {
		diagnostics, err := client.Networks.GetDiagnostics(ctx, int(runningNetwork.ID))
		if err != nil {
			t.Logf("Networks.GetDiagnostics returned error (may be expected if network is internal): %v", err)
			return
		}
		t.Logf("Query status: %s", diagnostics.Status)
		if diagnostics.Result != "" {
			t.Logf("Query result: %s", diagnostics.Result)
		}
	})

	// Test Ping (to a well-known address)
	t.Run("Ping", func(t *testing.T) {
		if runningNetwork.Type != "external" {
			t.Skip("Ping test requires external network")
		}
		result, err := client.Networks.Ping(ctx, int(runningNetwork.ID), "8.8.8.8", 3)
		if err != nil {
			t.Logf("Networks.Ping returned error: %v", err)
			return
		}
		t.Logf("Ping status: %s", result.Status)
		if result.Result != "" {
			t.Logf("Ping result:\n%s", result.Result)
		}
	})

	// Test DNS Lookup
	t.Run("DNSLookup", func(t *testing.T) {
		if runningNetwork.Type != "external" {
			t.Skip("DNS lookup test requires external network")
		}
		result, err := client.Networks.DNSLookup(ctx, int(runningNetwork.ID), "google.com")
		if err != nil {
			t.Logf("Networks.DNSLookup returned error: %v", err)
			return
		}
		t.Logf("DNS lookup status: %s", result.Status)
		if result.Result != "" {
			t.Logf("DNS lookup result:\n%s", result.Result)
		}
	})

	// Test GetStatistics
	t.Run("GetStatistics", func(t *testing.T) {
		statistics, err := client.Networks.GetStatistics(ctx, int(runningNetwork.ID))
		if err != nil {
			t.Logf("Networks.GetStatistics returned error (may be expected if monitoring not enabled): %v", err)
			return
		}
		t.Logf("Found %d statistics records", len(statistics))
		if len(statistics) > 0 {
			prettyPrintWave15(t, "Latest Statistics", statistics[0])
		}
	})

	// Test GetLatestStatistics
	t.Run("GetLatestStatistics", func(t *testing.T) {
		stats, err := client.Networks.GetLatestStatistics(ctx, int(runningNetwork.ID))
		if err != nil {
			t.Logf("Networks.GetLatestStatistics returned error: %v", err)
			return
		}
		if stats == nil {
			t.Log("No statistics available (monitoring may not be enabled on this network)")
			return
		}
		t.Logf("Latest stats - Quality: %d%%, Latency: %dms", stats.Quality, stats.LatencyUSAvg/1000)
	})

	// Test RunQuery with firewall rules display
	t.Run("ShowFirewallRules", func(t *testing.T) {
		result, err := client.Networks.RunQueryWait(ctx, &vergeos.NetworkQueryRequest{
			VNet:  int(runningNetwork.ID),
			Query: vergeos.NetworkQueryFirewall,
		})
		if err != nil {
			t.Logf("Networks.RunQueryWait (firewall) returned error: %v", err)
			return
		}
		t.Logf("Firewall rules query status: %s", result.Status)
		// Don't print full result as it can be very long
		if len(result.Result) > 500 {
			t.Logf("Firewall rules result (truncated): %s...", result.Result[:500])
		} else if result.Result != "" {
			t.Logf("Firewall rules result:\n%s", result.Result)
		}
	})
}

// setupTestClientWave15 creates a client from environment variables
func setupTestClientWave15(t *testing.T) *vergeos.Client {
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		t.Skip("Skipping integration test: VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD must be set")
	}

	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true),
		vergeos.WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

// prettyPrintWave15 logs a struct as formatted JSON for field verification
func prettyPrintWave15(t *testing.T, label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}
