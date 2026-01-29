// Diagnostic tool: API Field Type Analysis
//
// This tool analyzes the actual JSON types returned by the VergeOS API
// to help debug type mapping issues (string vs number, null handling, etc.)
//
// Usage:
//
//	export VERGEOS_HOST=https://your-vergeos-host
//	export VERGEOS_USERNAME=your-username
//	export VERGEOS_PASSWORD=your-password
//	go run ./test/integration/type-analysis/
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	vergeos "github.com/verge-io/govergeos"
)

var (
	apiHost     string
	apiUser     string
	apiPassword string
)

// RawCluster captures cluster data with interface{} types for analysis
type RawCluster struct {
	Key         interface{} `json:"$key"`
	ID          interface{} `json:"id"`
	System      interface{} `json:"system"`
	Name        interface{} `json:"name"`
	Description interface{} `json:"description"`
	Enabled     interface{} `json:"enabled"`
	Storage     interface{} `json:"storage"`
	Compute     interface{} `json:"compute"`
}

// RawVM captures VM data with interface{} types for analysis
type RawVM struct {
	Key             interface{} `json:"$key"`
	Name            interface{} `json:"name"`
	Cluster         interface{} `json:"cluster"`
	PreferredNode   interface{} `json:"preferred_node"`
	SnapshotProfile interface{} `json:"snapshot_profile"`
	HAGroup         interface{} `json:"ha_group"`
	Machine         interface{} `json:"machine"`
}

// RawNetwork captures network data with interface{} types for analysis
type RawNetwork struct {
	Key      interface{} `json:"$key"`
	Name     interface{} `json:"name"`
	Type     interface{} `json:"type"`
	Layer2ID interface{} `json:"layer2_id"`
	MTU      interface{} `json:"mtu"`
}

func main() {
	apiHost = os.Getenv("VERGEOS_HOST")
	apiUser = os.Getenv("VERGEOS_USERNAME")
	apiPassword = os.Getenv("VERGEOS_PASSWORD")

	if apiHost == "" || apiUser == "" || apiPassword == "" {
		log.Fatal("Please set VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD environment variables")
	}

	// Verify goVergeOS still works
	_, err := vergeos.NewClient(
		vergeos.WithBaseURL(apiHost),
		vergeos.WithCredentials(apiUser, apiPassword),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== API Field Type Analysis ===")
	fmt.Println("Testing actual return types from live VergeOS API")
	fmt.Printf("Host: %s\n", apiHost)
	fmt.Println()

	// Test Clusters
	fmt.Println("--- CLUSTERS ---")
	testClusters(ctx)

	// Test VMs
	fmt.Println("\n--- VMS (sample of 5) ---")
	testVMs(ctx)

	// Test Networks
	fmt.Println("\n--- NETWORKS (sample of 5) ---")
	testNetworks(ctx)

	// Test Settings
	fmt.Println("\n--- SETTINGS (sample of 3) ---")
	testSettings(ctx)

	// Test Resource Groups
	fmt.Println("\n--- RESOURCE GROUPS ---")
	testResourceGroups(ctx)

	fmt.Println("\n=== Analysis Complete ===")
}

func testClusters(ctx context.Context) {
	resp, err := makeRawRequest("/clusters?fields=$key,id,system,name,enabled,storage,compute")
	if err != nil {
		log.Printf("Clusters request failed: %v", err)
		return
	}

	var clusters []RawCluster
	if err := json.Unmarshal(resp, &clusters); err != nil {
		log.Printf("Clusters unmarshal failed: %v", err)
		return
	}

	if len(clusters) == 0 {
		fmt.Println("No clusters found")
		return
	}

	for _, c := range clusters {
		fmt.Printf("Cluster: %v\n", c.Name)
		fmt.Printf("  $key:    %v (type: %s)\n", c.Key, getType(c.Key))
		fmt.Printf("  id:      %v (type: %s)\n", c.ID, getType(c.ID))
		fmt.Printf("  system:  %v (type: %s) ** KEY FIELD **\n", c.System, getType(c.System))
		fmt.Printf("  enabled: %v (type: %s)\n", c.Enabled, getType(c.Enabled))
	}
}

func testVMs(ctx context.Context) {
	resp, err := makeRawRequest("/machines?fields=$key,name,cluster,preferred_node,snapshot_profile,ha_group,machine&limit=5")
	if err != nil {
		log.Printf("VMs request failed: %v", err)
		return
	}

	var vms []RawVM
	if err := json.Unmarshal(resp, &vms); err != nil {
		log.Printf("VMs unmarshal failed: %v", err)
		return
	}

	for _, vm := range vms {
		fmt.Printf("VM: %v\n", vm.Name)
		fmt.Printf("  $key:             %v (type: %s)\n", vm.Key, getType(vm.Key))
		fmt.Printf("  cluster:          %v (type: %s)\n", vm.Cluster, getType(vm.Cluster))
		fmt.Printf("  preferred_node:   %v (type: %s)\n", vm.PreferredNode, getType(vm.PreferredNode))
		fmt.Printf("  snapshot_profile: %v (type: %s)\n", vm.SnapshotProfile, getType(vm.SnapshotProfile))
		fmt.Printf("  ha_group:         %v (type: %s)\n", vm.HAGroup, getType(vm.HAGroup))
		fmt.Printf("  machine:          %v (type: %s)\n", vm.Machine, getType(vm.Machine))
	}
}

func testNetworks(ctx context.Context) {
	resp, err := makeRawRequest("/vnets?fields=$key,name,type,layer2_id,mtu&limit=5")
	if err != nil {
		log.Printf("Networks request failed: %v", err)
		return
	}

	var networks []RawNetwork
	if err := json.Unmarshal(resp, &networks); err != nil {
		log.Printf("Networks unmarshal failed: %v", err)
		return
	}

	for _, net := range networks {
		fmt.Printf("Network: %v\n", net.Name)
		fmt.Printf("  $key:      %v (type: %s)\n", net.Key, getType(net.Key))
		fmt.Printf("  type:      %v (type: %s)\n", net.Type, getType(net.Type))
		fmt.Printf("  layer2_id: %v (type: %s)\n", net.Layer2ID, getType(net.Layer2ID))
		fmt.Printf("  mtu:       %v (type: %s)\n", net.MTU, getType(net.MTU))
	}
}

func testSettings(ctx context.Context) {
	resp, err := makeRawRequest("/settings?fields=$key,key,value&limit=3")
	if err != nil {
		log.Printf("Settings request failed: %v", err)
		return
	}

	var settings []map[string]interface{}
	if err := json.Unmarshal(resp, &settings); err != nil {
		log.Printf("Settings unmarshal failed: %v", err)
		return
	}

	for _, s := range settings {
		fmt.Printf("Setting: %v\n", s["key"])
		fmt.Printf("  $key:  %v (type: %s)\n", s["$key"], getType(s["$key"]))
		fmt.Printf("  key:   %v (type: %s)\n", s["key"], getType(s["key"]))
		fmt.Printf("  value: %v (type: %s)\n", s["value"], getType(s["value"]))
	}
}

func testResourceGroups(ctx context.Context) {
	resp, err := makeRawRequest("/resource_groups?fields=$key,name,type,enabled")
	if err != nil {
		log.Printf("Resource groups request failed: %v", err)
		return
	}

	var groups []map[string]interface{}
	if err := json.Unmarshal(resp, &groups); err != nil {
		log.Printf("Resource groups unmarshal failed: %v", err)
		return
	}

	if len(groups) == 0 {
		fmt.Println("No resource groups found")
		return
	}

	for _, g := range groups {
		fmt.Printf("Resource Group: %v\n", g["name"])
		fmt.Printf("  $key:    %v (type: %s) ** UUID expected **\n", g["$key"], getType(g["$key"]))
		fmt.Printf("  type:    %v (type: %s)\n", g["type"], getType(g["type"]))
		fmt.Printf("  enabled: %v (type: %s)\n", g["enabled"], getType(g["enabled"]))
	}
}

// makeRawRequest makes a raw API request and returns the response body
func makeRawRequest(endpoint string) ([]byte, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	url := apiHost + "/api/v4" + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiUser, apiPassword)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-JSON-Non-Compact", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func getType(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "bool"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}
