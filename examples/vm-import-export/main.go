// Example: list VM imports, their logs, and VM exports.
//
// Create, Wait, Run, and Delete are separate calls and are not made here.
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

	fmt.Println("=== VM imports ===")
	imports, err := client.VMImports.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list VM imports: %v", err)
	}
	for _, imp := range imports {
		fmt.Printf("- %s status=%s key=%s vm=%d\n", imp.Name, imp.Status, imp.Key, flexKey(imp.VM))
	}
	if len(imports) > 0 && imports[0].Key != "" {
		fmt.Printf("\n=== Logs for %s ===\n", imports[0].Name)
		logs, err := client.VMImports.Logs(ctx, imports[0].Key, vergeos.WithLimit(20))
		if err != nil {
			log.Fatalf("Failed to list import logs: %v", err)
		}
		for _, line := range logs {
			fmt.Printf("- %s %s\n", line.Level, line.Text)
		}
	}

	fmt.Println("\n=== VM exports ===")
	exports, err := client.VMExports.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list VM exports: %v", err)
	}
	for _, exp := range exports {
		fmt.Printf("- volume %s (%s) status=%s max=%d\n", exp.VolumeName, exp.Volume, exp.Status, exp.MaxExports)
	}
}

func flexKey(id *vergeos.FlexInt) int {
	if id == nil {
		return 0
	}
	return id.Int()
}
