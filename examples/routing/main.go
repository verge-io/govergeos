// Example: list dynamic routing configuration.
//
// BGP, OSPF, and EIGRP rows hang off one vnet_bgp record per network.
// This example only reads. Create, update, and delete return
// RoutingRestartStatus. Pass WithRestartNetwork when that change should
// restart the network in the same call.
//
// Usage:
//
//	export VERGEOS_HOST=your-vergeos-host
//	export VERGEOS_USERNAME=admin
//	export VERGEOS_PASSWORD=yourpassword
//	export VERGEOS_INSECURE=true
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"

	vergeos "github.com/verge-io/govergeos"
)

func main() {
	client, err := vergeos.NewClient(vergeos.WithEnvConfig())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	networks, err := client.Networks.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list networks: %v", err)
	}
	if len(networks) == 0 {
		fmt.Println("No networks found.")
		return
	}

	for _, network := range networks {
		fmt.Printf("\n=== %s (Key %d, need_restart %v) ===\n", network.Name, network.Key, network.NeedRestart)
		cfg, err := client.VNetBGP.GetByNetwork(ctx, network.Key.Int())
		if vergeos.IsNotFoundError(err) {
			fmt.Println("No vnet_bgp row.")
			continue
		}
		if err != nil {
			log.Printf("Failed to read vnet_bgp for %s: %v", network.Name, err)
			continue
		}
		fmt.Printf("vnet_bgp key %d\n", cfg.Key)

		routers, err := client.VNetBGPRouters.ListByBGP(ctx, cfg.Key.Int())
		if err != nil {
			log.Printf("Failed to list BGP routers: %v", err)
		} else {
			fmt.Printf("BGP routers: %d\n", len(routers))
			for _, router := range routers {
				fmt.Printf("  - ASN %d (Key %d)\n", router.ASN, router.Key)
			}
		}

		ospf, err := client.VNetOSPFCommands.ListByBGP(ctx, cfg.Key.Int())
		if err != nil {
			log.Printf("Failed to list OSPF commands: %v", err)
		} else {
			fmt.Printf("OSPF commands: %d\n", len(ospf))
			for _, cmd := range ospf {
				fmt.Printf("  - %s %s\n", cmd.Command, cmd.Params)
			}
		}

		eigrp, err := client.VNetEIGRPRouters.ListByBGP(ctx, cfg.Key.Int())
		if err != nil {
			log.Printf("Failed to list EIGRP routers: %v", err)
		} else {
			fmt.Printf("EIGRP routers: %d\n", len(eigrp))
			for _, router := range eigrp {
				fmt.Printf("  - ASN %d (Key %d)\n", router.ASN, router.Key)
			}
		}
	}
}
