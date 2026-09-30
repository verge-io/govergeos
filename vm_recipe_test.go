package vergeos

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestVMRecipeService_ListGetQuestions(t *testing.T) {
	const id = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	const catalogID = "cccccccccccccccccccccccccccccccccccccccc"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_recipes": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			switch filter {
			case "":
				jsonResponse(w, 200, []VMRecipe{{
					Key: id, ID: id, Name: "Debian 12 (Bookworm)", Catalog: catalogID, CatalogName: "Linux", Downloaded: true,
				}})
			case "id eq '" + id + "'":
				jsonResponse(w, 200, []VMRecipe{{
					Key: id, ID: id, Name: "Debian 12 (Bookworm)", Version: "12",
				}})
			case "name eq 'Debian 12 (Bookworm)'":
				jsonResponse(w, 200, []VMRecipe{{
					Key: id, ID: id, Name: "Debian 12 (Bookworm)",
				}})
			case "catalog eq '" + catalogID + "'":
				jsonResponse(w, 200, []VMRecipe{{
					Key: id, ID: id, Name: "Debian 12 (Bookworm)", Catalog: catalogID,
				}})
			default:
				t.Errorf("filter = %q", filter)
				jsonResponse(w, 200, []VMRecipe{})
			}
		},
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			want := "recipe eq 'vm_recipes/" + id + "'"
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
				{Key: 1, Name: "HOSTNAME", Type: "string", Required: true, OrderID: 1},
				{Key: 2, Name: "SELECT_CREATE_UEFI", Type: "bool", OrderID: 2},
				{Key: 3, Name: "YB_DRIVE_OS_SIZE", Type: "disksize", OrderID: 3},
			})
		},
	}))

	recipes, err := client.VMRecipes.List(context.Background())
	if err != nil || len(recipes) != 1 || !recipes[0].Downloaded {
		t.Fatalf("list = %+v %v", recipes, err)
	}
	got, err := client.VMRecipes.Get(context.Background(), id)
	if err != nil || got.Version != "12" {
		t.Fatalf("get = %+v %v", got, err)
	}
	byName, err := client.VMRecipes.GetByName(context.Background(), "Debian 12 (Bookworm)")
	if err != nil || byName.Key != id {
		t.Fatalf("by name = %+v %v", byName, err)
	}
	inCatalog, err := client.VMRecipes.ListByCatalog(context.Background(), catalogID)
	if err != nil || len(inCatalog) != 1 {
		t.Fatalf("by catalog = %+v %v", inCatalog, err)
	}

	questions, err := client.VMRecipes.Questions(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) != 3 || questions[1].Type != "bool" || questions[2].Type != "disksize" {
		t.Fatalf("questions = %+v", questions)
	}
}

func TestVMRecipeService_GetByNameAmbiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_recipes": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMRecipe{
				{Key: "a", Name: "Ubuntu"},
				{Key: "b", Name: "Ubuntu"},
			})
		},
	}))
	_, err := client.VMRecipes.GetByName(context.Background(), "Ubuntu")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("ambiguous: %v", err)
	}
	_, err = client.VMRecipes.Questions(context.Background(), "")
	if !IsValidationError(err) {
		t.Fatalf("empty recipe: %v", err)
	}
}
