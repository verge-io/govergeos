// Example: DR Sites - Remote Site Management
//
// This example demonstrates how to:
// - List and manage remote sites for disaster recovery
// - View incoming and outgoing sync configurations
// - Query site status and synchronization state
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass go run ./examples/dr-sites/
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
	fmt.Println("=== Remote Sites ===")
	showSites(ctx, client)

	fmt.Println("\n=== Incoming Syncs ===")
	showIncomingSyncs(ctx, client)

	fmt.Println("\n=== Outgoing Syncs ===")
	showOutgoingSyncs(ctx, client)
}

func showSites(ctx context.Context, client *vergeos.Client) {
	// List all remote sites
	sites, err := client.Sites.List(ctx)
	if err != nil {
		log.Printf("Failed to list sites: %v", err)
		return
	}

	fmt.Printf("Found %d remote sites\n", len(sites))

	if len(sites) == 0 {
		fmt.Println("No remote sites configured - this system is standalone")
		return
	}

	// Group by status
	byStatus := make(map[string]int)
	for _, s := range sites {
		byStatus[s.Status]++
	}
	fmt.Println("\nSites by status:")
	for status, count := range byStatus {
		fmt.Printf("  %s: %d\n", status, count)
	}

	// Show site details
	fmt.Println("\nSite details:")
	for _, site := range sites {
		fmt.Printf("\n  Site: %s (ID: %d)\n", site.Name, int(site.Key))
		fmt.Printf("    URL: %s\n", site.URL)
		fmt.Printf("    Status: %s\n", site.Status)
		if site.StatusInfo != "" {
			fmt.Printf("    Status Info: %s\n", site.StatusInfo)
		}
		fmt.Printf("    Auth Status: %s\n", site.AuthenticationStatus)
		fmt.Printf("    Enabled: %v\n", site.Enabled)
		if site.Domain != "" {
			fmt.Printf("    Domain: %s\n", site.Domain)
		}
		if site.City != "" {
			fmt.Printf("    Location: %s, %s\n", site.City, site.Country)
		}

		// Show sync configuration
		fmt.Printf("    Cloud Snapshots: %s\n", site.ConfigCloudSnapshots)
		fmt.Printf("    Statistics: %s\n", site.ConfigStatistics)
		fmt.Printf("    Management: %s\n", site.ConfigManagement)

		// Show sync status
		fmt.Printf("    Incoming Syncs Enabled: %v\n", site.IncomingSyncsEnabled)
		fmt.Printf("    Outgoing Syncs Enabled: %v\n", site.OutgoingSyncsEnabled)

		if site.Created > 0 {
			fmt.Printf("    Created: %s\n", time.Unix(site.Created, 0).Format(time.RFC3339))
		}
	}

	// Get a specific site by name (if we have any)
	if len(sites) > 0 {
		site, err := client.Sites.GetByName(ctx, sites[0].Name)
		if err != nil {
			log.Printf("Failed to get site by name: %v", err)
		} else {
			fmt.Printf("\nRetrieved site by name: %s (Key: %d)\n", site.Name, int(site.Key))
		}
	}

	// Example: Refresh site connection (commented out to avoid side effects)
	// if len(sites) > 0 {
	// 	siteID := int(sites[0].Key)
	// 	if err := client.Sites.Refresh(ctx, siteID); err != nil {
	// 		log.Printf("Failed to refresh site: %v", err)
	// 	} else {
	// 		fmt.Printf("Refreshed site %d\n", siteID)
	// 	}
	// }
}

func showIncomingSyncs(ctx context.Context, client *vergeos.Client) {
	// List all incoming syncs
	syncs, err := client.SiteSyncsIncoming.List(ctx)
	if err != nil {
		log.Printf("Failed to list incoming syncs: %v", err)
		return
	}

	fmt.Printf("Found %d incoming syncs\n", len(syncs))

	if len(syncs) == 0 {
		fmt.Println("No incoming syncs configured")
		return
	}

	// Group by status
	byStatus := make(map[string]int)
	enabledCount := 0
	for _, s := range syncs {
		byStatus[s.Status]++
		if s.Enabled {
			enabledCount++
		}
	}
	fmt.Printf("\nEnabled: %d / %d\n", enabledCount, len(syncs))
	fmt.Println("By status:")
	for status, count := range byStatus {
		fmt.Printf("  %s: %d\n", status, count)
	}

	// Show sync details
	fmt.Println("\nIncoming sync details:")
	for _, sync := range syncs {
		status := "disabled"
		if sync.Enabled {
			status = "enabled"
		}
		fmt.Printf("  - %s [%s] (Site: %d, Status: %s)\n",
			sync.Name, status, int(sync.Site), sync.Status)
		if sync.StatusInfo != "" {
			fmt.Printf("      Info: %s\n", sync.StatusInfo)
		}
		fmt.Printf("      State: %s\n", sync.State)
		if sync.LastSync > 0 {
			fmt.Printf("      Last Sync: %s\n", time.Unix(sync.LastSync, 0).Format(time.RFC3339))
		}
	}
}

func showOutgoingSyncs(ctx context.Context, client *vergeos.Client) {
	// List all outgoing syncs
	syncs, err := client.SiteSyncsOutgoing.List(ctx)
	if err != nil {
		log.Printf("Failed to list outgoing syncs: %v", err)
		return
	}

	fmt.Printf("Found %d outgoing syncs\n", len(syncs))

	if len(syncs) == 0 {
		fmt.Println("No outgoing syncs configured")
		return
	}

	// Group by status
	byStatus := make(map[string]int)
	enabledCount := 0
	for _, s := range syncs {
		byStatus[s.Status]++
		if s.Enabled {
			enabledCount++
		}
	}
	fmt.Printf("\nEnabled: %d / %d\n", enabledCount, len(syncs))
	fmt.Println("By status:")
	for status, count := range byStatus {
		fmt.Printf("  %s: %d\n", status, count)
	}

	// Show sync details
	fmt.Println("\nOutgoing sync details:")
	for _, sync := range syncs {
		status := "disabled"
		if sync.Enabled {
			status = "enabled"
		}
		fmt.Printf("  - %s [%s] (Site: %d, Status: %s)\n",
			sync.Name, status, int(sync.Site), sync.Status)
		if sync.StatusInfo != "" {
			fmt.Printf("      Info: %s\n", sync.StatusInfo)
		}
		fmt.Printf("      State: %s, Destination Tier: %s\n", sync.State, sync.DestinationTier)
		if sync.SendThrottle > 0 {
			fmt.Printf("      Send Throttle: %d KB/s\n", sync.SendThrottle)
		}
		if sync.LastRun > 0 {
			fmt.Printf("      Last Run: %s\n", time.Unix(sync.LastRun, 0).Format(time.RFC3339))
		}
	}

	// Example: Throttle an outgoing sync (commented out to avoid side effects)
	// if len(syncs) > 0 {
	// 	syncID := int(syncs[0].Key)
	// 	throttleKBps := 10000 // 10 MB/s
	// 	if err := client.SiteSyncsOutgoing.Throttle(ctx, syncID, throttleKBps); err != nil {
	// 		log.Printf("Failed to throttle sync: %v", err)
	// 	} else {
	// 		fmt.Printf("Set throttle on sync %d to %d KB/s\n", syncID, throttleKBps)
	// 	}
	// }
}
