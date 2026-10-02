package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const tenantRecipeHexKey = "dddddddddddddddddddddddddddddddddddddddd"

func tenantRecipeQuestionFixture() []RecipeQuestion {
	return []RecipeQuestion{
		{Name: "YB_USER_NAME", Type: "string", Required: true},
		{Name: "YB_EXPOSE_CLOUD_SNAPSHOTS", Type: "bool"},
		{Name: "YB_NIC_ETH0", Type: "network"},
	}
}

func TestTenantRecipeInstanceService_DeployValidatesBeforeSend(t *testing.T) {
	var posted []byte
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			want := "recipe eq 'tenant_recipes/" + tenantRecipeHexKey + "'"
			if !strings.Contains(filter, want) {
				t.Errorf("filter = %q", filter)
			}
			jsonResponse(w, 200, tenantRecipeQuestionFixture())
		},
		"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("fields") != "$key,name" {
				t.Errorf("network fields = %q", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []Network{{Key: 12, Name: "Internal"}})
		},
		"POST /api/v4/tenant_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			posted = body
			if strings.Contains(string(body), "simulate") {
				t.Errorf("deploy body contained simulate: %s", body)
			}
			jsonResponse(w, 200, map[string]any{"$key": 9})
		},
		"GET /api/v4/tenant_recipe_instances/9": func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("fields"), "answers") {
				t.Errorf("get fields = %q", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, TenantRecipeInstance{Key: 9, Recipe: tenantRecipeHexKey, Name: "customer-a", Tenant: 44})
		},
		"GET /api/v4/tenant_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Query().Get("fields"), "answers") {
				t.Errorf("list fields included answers: %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []TenantRecipeInstance{{Key: 9, Name: "customer-a", Recipe: tenantRecipeHexKey}})
		},
	}))

	instance, err := client.TenantRecipeInstances.Deploy(context.Background(), &TenantRecipeDeployRequest{
		Recipe: tenantRecipeHexKey,
		Name:   "customer-a",
		Answers: RecipeAnswers{
			"YB_USER_NAME":              "admin",
			"YB_EXPOSE_CLOUD_SNAPSHOTS": "yes",
			"YB_NIC_ETH0":               "Internal",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if int(instance.Key) != 9 || int(instance.Tenant) != 44 {
		t.Fatalf("instance = %+v", instance)
	}

	var body map[string]any
	if err := json.Unmarshal(posted, &body); err != nil {
		t.Fatal(err)
	}
	if body["recipe"] != tenantRecipeHexKey || body["name"] != "customer-a" {
		t.Fatalf("body = %#v", body)
	}
	if _, ok := body["simulate"]; ok {
		t.Fatal("deploy JSON included simulate")
	}
	answers := body["answers"].(map[string]any)
	if answers["YB_EXPOSE_CLOUD_SNAPSHOTS"] != true {
		t.Fatalf("bool = %#v", answers["YB_EXPOSE_CLOUD_SNAPSHOTS"])
	}
	if answers["YB_NIC_ETH0"] != float64(12) {
		t.Fatalf("network = %#v", answers["YB_NIC_ETH0"])
	}

	listed, err := client.TenantRecipeInstances.List(context.Background())
	if err != nil || len(listed) != 1 || listed[0].Answers != nil {
		t.Fatalf("list = %+v %v", listed, err)
	}
}

func TestTenantRecipeInstanceService_DeployRefusesTraps(t *testing.T) {
	posted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, tenantRecipeQuestionFixture())
		},
		"POST /api/v4/tenant_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			posted = true
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
	}))

	_, err := client.TenantRecipeInstances.Deploy(context.Background(), &TenantRecipeDeployRequest{
		Recipe: tenantRecipeHexKey,
		Name:   "bool-test",
		Answers: RecipeAnswers{
			"YB_USER_NAME":              "admin",
			"YB_EXPOSE_CLOUD_SNAPSHOTS": "enabled",
		},
	})
	if !IsValidationError(err) || !strings.Contains(err.Error(), "not a recognized boolean") || !strings.Contains(err.Error(), "enabled") {
		t.Fatalf("bool: %v", err)
	}
	if posted {
		t.Fatal("unrecognized bool was sent")
	}
}

func TestTenantRecipeInstanceService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "recipe eq '"+tenantRecipeHexKey+"'" {
				jsonResponse(w, 200, []TenantRecipeInstance{{Key: 3, Name: "customer-a", Recipe: tenantRecipeHexKey}})
				return
			}
			if filter != "name eq 'customer-a'" {
				t.Errorf("filter = %q", filter)
			}
			jsonResponse(w, 200, []TenantRecipeInstance{{Key: 3, Name: "customer-a"}})
		},
		"GET /api/v4/tenant_recipe_instances/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TenantRecipeInstance{Key: 3, Name: "customer-a", Tenant: 8})
		},
	}))

	got, err := client.TenantRecipeInstances.GetByName(context.Background(), "customer-a")
	if err != nil || int(got.Tenant) != 8 {
		t.Fatalf("get = %+v %v", got, err)
	}
	rows, err := client.TenantRecipeInstances.ListByRecipe(context.Background(), tenantRecipeHexKey)
	if err != nil || len(rows) != 1 {
		t.Fatalf("by recipe = %+v %v", rows, err)
	}
	_, err = client.TenantRecipeInstances.Deploy(context.Background(), nil)
	if !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}

	missing := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenant_recipe_instances/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err = missing.TenantRecipeInstances.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("missing instance: %v", err)
	}
	_, err = missing.TenantRecipeInstances.Deploy(context.Background(), &TenantRecipeDeployRequest{Recipe: tenantRecipeHexKey})
	if !IsValidationError(err) {
		t.Fatalf("missing name: %v", err)
	}
}
