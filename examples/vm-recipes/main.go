// Example: list VM recipes from a catalog.
//
// Deploy is a separate call and is not made here. Preview is the dry run;
// it is also not called here, because it still needs a recipe whose required
// answers you have.
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

	fmt.Println("=== Catalogs ===")
	catalogs, err := client.Catalogs.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list catalogs: %v", err)
	}
	for _, catalog := range catalogs {
		fmt.Printf("- %s (key %s, scope %s)\n", catalog.Name, catalog.Key, catalog.PublishingScope)
	}

	fmt.Println("\n=== VM recipes ===")
	recipes, err := client.VMRecipes.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list VM recipes: %v", err)
	}
	for _, recipe := range recipes {
		fmt.Printf("- %s v%s (catalog %s, downloaded %v)\n", recipe.Name, recipe.Version, recipe.CatalogName, recipe.Downloaded)
	}
	if len(recipes) == 0 {
		return
	}

	fmt.Printf("\n=== Questions for %s ===\n", recipes[0].Name)
	questions, err := client.VMRecipes.Questions(ctx, recipes[0].Key)
	if err != nil {
		log.Fatalf("Failed to list questions: %v", err)
	}
	for _, question := range questions {
		fmt.Printf("- %s type=%s required=%v\n", question.Name, question.Type, question.Required)
	}
}
