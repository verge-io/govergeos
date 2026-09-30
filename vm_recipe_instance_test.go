package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const recipeKey = "dddddddddddddddddddddddddddddddddddddddd"

func recipeQuestionFixture() []RecipeQuestion {
	return []RecipeQuestion{
		{Name: "HOSTNAME", Type: "string", Required: true},
		{Name: "SELECT_CREATE_UEFI", Type: "bool"},
		{Name: "YB_DRIVE_OS_SIZE", Type: "disksize"},
		{Name: "YB_NIC_ETH0", Type: "network"},
	}
}

func TestVMRecipeInstanceService_DeployValidatesBeforeSend(t *testing.T) {
	var posted []byte
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, recipeQuestionFixture())
		},
		"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("fields") != "$key,name" {
				t.Errorf("network fields = %q", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []Network{{ID: 12, Name: "Internal"}})
		},
		"POST /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
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
		"GET /api/v4/vm_recipe_instances/9": func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("fields"), "answers") {
				t.Errorf("get fields = %q", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, VMRecipeInstance{Key: 9, Recipe: recipeKey, Name: "web-01", VM: 44})
		},
		"GET /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Query().Get("fields"), "answers") {
				t.Errorf("list fields included answers: %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, []VMRecipeInstance{{Key: 9, Name: "web-01", Recipe: recipeKey}})
		},
	}))

	auto := true
	instance, err := client.VMRecipeInstances.Deploy(context.Background(), &VMRecipeDeployRequest{
		Recipe:     recipeKey,
		Name:       "web-01",
		AutoUpdate: &auto,
		Answers: RecipeAnswers{
			"HOSTNAME":           "web-01",
			"SELECT_CREATE_UEFI": "yes",
			"YB_DRIVE_OS_SIZE":   int64(RecipeDiskSize50GB),
			"YB_NIC_ETH0":        "Internal",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if int(instance.Key) != 9 || int(instance.VM) != 44 {
		t.Fatalf("instance = %+v", instance)
	}

	var body map[string]any
	if err := json.Unmarshal(posted, &body); err != nil {
		t.Fatal(err)
	}
	if body["recipe"] != recipeKey || body["name"] != "web-01" || body["auto_update"] != true {
		t.Fatalf("body = %#v", body)
	}
	if _, ok := body["simulate"]; ok {
		t.Fatal("deploy JSON included simulate")
	}
	answers := body["answers"].(map[string]any)
	if answers["SELECT_CREATE_UEFI"] != true {
		t.Fatalf("uefi = %#v", answers["SELECT_CREATE_UEFI"])
	}
	if answers["YB_NIC_ETH0"] != float64(12) {
		t.Fatalf("network = %#v", answers["YB_NIC_ETH0"])
	}
	if answers["YB_DRIVE_OS_SIZE"] != float64(RecipeDiskSize50GB) {
		t.Fatalf("disk = %#v", answers["YB_DRIVE_OS_SIZE"])
	}

	listed, err := client.VMRecipeInstances.List(context.Background())
	if err != nil || len(listed) != 1 || listed[0].Answers != nil {
		t.Fatalf("list = %+v %v", listed, err)
	}
}

func TestVMRecipeInstanceService_DeployRefusesTraps(t *testing.T) {
	posted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, recipeQuestionFixture())
		},
		"POST /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			posted = true
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
	}))

	_, err := client.VMRecipeInstances.Deploy(context.Background(), &VMRecipeDeployRequest{
		Recipe: recipeKey,
		Name:   "uefi-test",
		Answers: RecipeAnswers{
			"HOSTNAME":           "uefi-test",
			"SELECT_CREATE_UEFI": "enabled",
			"YB_DRIVE_OS_SIZE":   int64(RecipeDiskSize50GB),
		},
	})
	if !IsValidationError(err) || !strings.Contains(err.Error(), "not a recognized boolean") || !strings.Contains(err.Error(), "enabled") {
		t.Fatalf("bool: %v", err)
	}
	if posted {
		t.Fatal("unrecognized bool was sent")
	}

	_, err = client.VMRecipeInstances.Deploy(context.Background(), &VMRecipeDeployRequest{
		Recipe: recipeKey,
		Name:   "disk-test",
		Answers: RecipeAnswers{
			"HOSTNAME":         "disk-test",
			"YB_DRIVE_OS_SIZE": 50,
		},
	})
	if !IsValidationError(err) || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("disk: %v", err)
	}
	if posted {
		t.Fatal("50-byte disk size was sent")
	}
}

func TestVMRecipeInstanceService_PreviewIsNotDeploy(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, recipeQuestionFixture())
		},
		"POST /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["simulate"] != true {
				t.Fatalf("simulate = %#v", body["simulate"])
			}
			jsonResponse(w, 405, map[string]any{
				"err": "Simulation complete",
				"response": map[string]any{
					"cloudinit_files": []map[string]string{{"name": "user-data", "contents": "#cloud-config"}},
					"logs":            []string{"rendered user-data"},
					"answers":         map[string]any{"HOSTNAME": "preview", "YB_VM_KEY": "123"},
				},
			})
		},
		"GET /api/v4/vm_recipe_instances/123": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("preview looked up an instance")
		},
		"GET /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("preview listed instances")
		},
	}))

	preview, err := client.VMRecipeInstances.Preview(context.Background(), &VMRecipeDeployRequest{
		Recipe:  recipeKey,
		Name:    "preview",
		Answers: RecipeAnswers{"HOSTNAME": "preview", "YB_DRIVE_OS_SIZE": 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.CloudInitFiles[0].Contents != "#cloud-config" || preview.Answers["YB_VM_KEY"] != "123" {
		t.Fatalf("preview = %#v", preview)
	}
}

func TestVMRecipeInstanceService_PreviewDoesNotReturnACreate(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []RecipeQuestion{})
		},
		"POST /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 201, map[string]any{"$key": 5, "response": map[string]any{"vm": 8}})
		},
		"GET /api/v4/vm_recipe_instances/5": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("persisted preview was read back as an instance")
		},
	}))

	preview, err := client.VMRecipeInstances.Preview(context.Background(), &VMRecipeDeployRequest{
		Recipe: recipeKey,
		Name:   "should-not-exist",
	})
	if preview != nil || !IsRecipePreviewPersistedError(err) {
		t.Fatalf("preview = %#v err = %v", preview, err)
	}
	if !strings.Contains(err.Error(), "should-not-exist") {
		t.Fatalf("error = %v", err)
	}
}

func TestVMRecipeInstanceService_PreviewLogFailure(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/recipe_questions": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []RecipeQuestion{{Name: "HOSTNAME", Type: "string", Required: true}})
		},
		"POST /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 405, map[string]any{
				"err": "Simulation complete",
				"response": map[string]any{
					"cloudinit_files": []any{},
					"logs":            []string{"Error executing API command: OS drive was not created"},
					"answers":         map[string]any{},
				},
			})
		},
	}))

	preview, err := client.VMRecipeInstances.Preview(context.Background(), &VMRecipeDeployRequest{
		Recipe:  recipeKey,
		Name:    "tier-test",
		Answers: RecipeAnswers{"HOSTNAME": "tier-test"},
	})
	if preview != nil || !IsRecipePreviewFailedError(err) {
		t.Fatalf("preview = %#v err = %v", preview, err)
	}
	if !strings.Contains(err.Error(), "OS drive was not created") {
		t.Fatalf("error = %v", err)
	}
}

func TestVMRecipeInstanceService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_recipe_instances": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "recipe eq '"+recipeKey+"'" {
				jsonResponse(w, 200, []VMRecipeInstance{{Key: 3, Name: "web-01", Recipe: recipeKey}})
				return
			}
			if filter != "name eq 'web-01'" {
				t.Errorf("filter = %q", filter)
			}
			jsonResponse(w, 200, []VMRecipeInstance{{Key: 3, Name: "web-01"}})
		},
		"GET /api/v4/vm_recipe_instances/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMRecipeInstance{Key: 3, Name: "web-01", VM: 8})
		},
	}))

	got, err := client.VMRecipeInstances.GetByName(context.Background(), "web-01")
	if err != nil || int(got.VM) != 8 {
		t.Fatalf("get = %+v %v", got, err)
	}
	rows, err := client.VMRecipeInstances.ListByRecipe(context.Background(), recipeKey)
	if err != nil || len(rows) != 1 {
		t.Fatalf("by recipe = %+v %v", rows, err)
	}
	_, err = client.VMRecipeInstances.Deploy(context.Background(), nil)
	if !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}

	missing := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_recipe_instances/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err = missing.VMRecipeInstances.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("missing instance: %v", err)
	}
	_, err = missing.VMRecipeInstances.Deploy(context.Background(), &VMRecipeDeployRequest{Recipe: recipeKey})
	if !IsValidationError(err) {
		t.Fatalf("missing name: %v", err)
	}
}
