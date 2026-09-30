//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestCatalogsAndVMRecipes lists catalogs, VM recipes, and the questions on
// the first recipe. It does not deploy.
func TestCatalogsAndVMRecipes(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	catalogs, err := client.Catalogs.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list catalogs: %v", err)
	}
	t.Logf("Found %d catalog(s)", len(catalogs))
	for _, catalog := range catalogs {
		t.Logf("Catalog: Key=%s Name=%q Scope=%q", catalog.Key, catalog.Name, catalog.PublishingScope)
	}

	recipes, err := client.VMRecipes.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list VM recipes: %v", err)
	}
	t.Logf("Found %d VM recipe(s)", len(recipes))
	if len(recipes) == 0 {
		return
	}

	recipe := recipes[0]
	t.Logf("Recipe: Key=%s Name=%q Version=%q Downloaded=%v", recipe.Key, recipe.Name, recipe.Version, recipe.Downloaded)
	if recipe.Key != "" {
		got, err := client.VMRecipes.Get(ctx, recipe.Key)
		if err != nil {
			t.Fatalf("Failed to get recipe %s: %v", recipe.Key, err)
		}
		if got.Name != recipe.Name {
			t.Fatalf("Get name = %q, list name = %q", got.Name, recipe.Name)
		}
	}

	questions, err := client.VMRecipes.Questions(ctx, recipe.Key)
	if err != nil {
		t.Fatalf("Failed to list questions: %v", err)
	}
	t.Logf("Found %d question(s) on %s", len(questions), recipe.Name)
	for _, question := range questions {
		// Name and type only. Defaults and answers can hold a guest password.
		t.Logf("Question: Name=%q Type=%s Required=%v", question.Name, question.Type, question.Required)
	}

	instances, err := client.VMRecipeInstances.List(ctx, vergeos.WithLimit(5))
	if err != nil {
		t.Fatalf("Failed to list recipe instances: %v", err)
	}
	t.Logf("Found %d recipe instance(s) in the first page", len(instances))
}
