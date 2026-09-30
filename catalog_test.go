package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestCatalogService_ListAndGet(t *testing.T) {
	const id = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/catalogs": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got == "" || got == "most" {
				t.Errorf("fields = %q", got)
			}
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				jsonResponse(w, 200, []Catalog{{
					Key: id, ID: id, Name: "Verge.io Recipes", PublishingScope: CatalogScopeGlobal,
				}})
				return
			}
			if filter != "id eq '"+id+"'" && filter != "name eq 'Verge.io Recipes'" {
				t.Errorf("filter = %q", filter)
			}
			jsonResponse(w, 200, []Catalog{{
				Key: id, ID: id, Name: "Verge.io Recipes", Repository: 2, RepositoryName: "Marketplace",
			}})
		},
	}))

	catalogs, err := client.Catalogs.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogs) != 1 || catalogs[0].Name != "Verge.io Recipes" {
		t.Fatalf("catalogs = %+v", catalogs)
	}

	got, err := client.Catalogs.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if got.RepositoryName != "Marketplace" || int(got.Repository) != 2 {
		t.Fatalf("get = %+v", got)
	}

	byName, err := client.Catalogs.GetByName(context.Background(), "Verge.io Recipes")
	if err != nil || byName.Key != id {
		t.Fatalf("by name = %+v %v", byName, err)
	}
}

func TestCatalogService_GetByNameAmbiguousAndMissing(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/catalogs": func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("filter") {
			case "name eq 'Shared'":
				jsonResponse(w, 200, []Catalog{
					{Key: "a", Name: "Shared"},
					{Key: "b", Name: "Shared"},
				})
			case "name eq 'Missing'":
				jsonResponse(w, 200, []Catalog{})
			default:
				t.Errorf("filter = %q", r.URL.Query().Get("filter"))
				jsonResponse(w, 200, []Catalog{})
			}
		},
	}))

	_, err := client.Catalogs.GetByName(context.Background(), "Shared")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("ambiguous: %v", err)
	}
	_, err = client.Catalogs.GetByName(context.Background(), "Missing")
	if !IsNotFoundError(err) {
		t.Fatalf("missing: %v", err)
	}
	_, err = client.Catalogs.Get(context.Background(), "")
	if !IsValidationError(err) {
		t.Fatalf("empty id: %v", err)
	}
}
