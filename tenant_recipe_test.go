package vergeos

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestTenantRecipeService_ListGetQuestions(t *testing.T) {
	const id = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	const catalogID = "cccccccccccccccccccccccccccccccccccccccc"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_recipes": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			switch filter {
			case "":
				jsonResponse(w, 200, []TenantRecipe{{
					Key: id, ID: id, Name: "30-Day Trial (POC)", Catalog: catalogID, CatalogName: "Tenants", Downloaded: true,
				}})
			case "id eq '" + id + "'":
				jsonResponse(w, 200, []TenantRecipe{{
					Key: id, ID: id, Name: "30-Day Trial (POC)", Version: "1.0.0", PreserveCerts: true,
				}})
			case "name eq '30-Day Trial (POC)'":
				jsonResponse(w, 200, []TenantRecipe{{
					Key: id, ID: id, Name: "30-Day Trial (POC)",
				}})
			case "catalog eq '" + catalogID + "'":
				jsonResponse(w, 200, []TenantRecipe{{
					Key: id, ID: id, Name: "30-Day Trial (POC)", Catalog: catalogID,
				}})
			default:
				t.Errorf("filter = %q", filter)
				jsonResponse(w, 200, []TenantRecipe{})
			}
		},
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			want := "recipe eq 'tenant_recipes/" + id + "'"
			if !strings.Contains(filter, want) {
				t.Errorf("filter = %q", filter)
			}
			if r.URL.Query().Get("sort") != "orderid" {
				t.Errorf("sort = %q", r.URL.Query().Get("sort"))
			}
			if !strings.Contains(r.URL.Query().Get("fields"), "type") {
				t.Errorf("fields = %q", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []RecipeQuestion{
				{Key: 1, Name: "YB_USER_NAME", Type: "string", Required: true, OrderID: 1},
				{Key: 2, Name: "YB_EXPOSE_CLOUD_SNAPSHOTS", Type: "bool", OrderID: 2},
				{Key: 3, Name: "YB_NODE_1_RAM", Type: "num", OrderID: 3},
			})
		},
	}))

	recipes, err := client.TenantRecipes.List(context.Background())
	if err != nil || len(recipes) != 1 || !recipes[0].Downloaded {
		t.Fatalf("list = %+v %v", recipes, err)
	}
	got, err := client.TenantRecipes.Get(context.Background(), id)
	if err != nil || got.Version != "1.0.0" || !got.PreserveCerts {
		t.Fatalf("get = %+v %v", got, err)
	}
	byName, err := client.TenantRecipes.GetByName(context.Background(), "30-Day Trial (POC)")
	if err != nil || byName.Key != id {
		t.Fatalf("by name = %+v %v", byName, err)
	}
	inCatalog, err := client.TenantRecipes.ListByCatalog(context.Background(), catalogID)
	if err != nil || len(inCatalog) != 1 {
		t.Fatalf("by catalog = %+v %v", inCatalog, err)
	}

	questions, err := client.TenantRecipes.Questions(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) != 3 || questions[1].Type != "bool" || questions[2].Type != "num" {
		t.Fatalf("questions = %+v", questions)
	}
}

func TestTenantRecipeService_GetByNameAmbiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_recipes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantRecipe{
				{Key: "a", Name: "Standard"},
				{Key: "b", Name: "Standard"},
			})
		},
	}))
	_, err := client.TenantRecipes.GetByName(context.Background(), "Standard")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("ambiguous: %v", err)
	}
	_, err = client.TenantRecipes.Questions(context.Background(), "")
	if !IsValidationError(err) {
		t.Fatalf("empty recipe: %v", err)
	}
	_, err = client.TenantRecipes.ListByCatalog(context.Background(), "")
	if !IsValidationError(err) {
		t.Fatalf("empty catalog: %v", err)
	}
}
