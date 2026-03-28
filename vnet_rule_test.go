package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// --- VNetRuleService tests ---

func TestVNetRuleService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rules": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetRule{
				{Key: 1, Name: "allow-ssh", VNet: 10, Action: "accept"},
				{Key: 2, Name: "block-all", VNet: 10, Action: "drop"},
			})
		},
	}))

	rules, err := client.VNetRules.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Name != "allow-ssh" {
		t.Errorf("expected name 'allow-ssh', got %q", rules[0].Name)
	}
}

func TestVNetRuleService_ListByNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rules": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "vnet eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []VNetRule{{Key: 1, VNet: 10, Name: "allow-ssh"}})
		},
	}))

	rules, err := client.VNetRules.ListByNetwork(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByNetwork failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}

func TestVNetRuleService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 1, Name: "allow-ssh", Action: "accept", Enabled: true})
		},
	}))

	rule, err := client.VNetRules.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if rule.Name != "allow-ssh" {
		t.Errorf("expected name 'allow-ssh', got %q", rule.Name)
	}
	if !rule.Enabled {
		t.Error("expected rule to be enabled")
	}
}

func TestVNetRuleService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rules/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetRules.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rules": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetRule{{Key: 1, VNet: 10, Name: "allow-ssh"}})
		},
		"GET /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 1, VNet: 10, Name: "allow-ssh", Action: "accept"})
		},
	}))

	rule, err := client.VNetRules.GetByName(context.Background(), 10, "allow-ssh")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if rule.Action != "accept" {
		t.Errorf("expected action 'accept', got %q", rule.Action)
	}
}

func TestVNetRuleService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rules": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetRule{})
		},
	}))

	_, err := client.VNetRules.GetByName(context.Background(), 10, "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_rules": func(w http.ResponseWriter, r *http.Request) {
			var req VNetRuleCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "allow-http" {
				t.Errorf("expected name 'allow-http', got %q", req.Name)
			}
			if req.VNet != 10 {
				t.Errorf("expected vnet 10, got %d", req.VNet)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 5, "status": "ok"})
		},
		"GET /api/v4/vnet_rules/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 5, VNet: 10, Name: "allow-http", Action: "accept", Enabled: true})
		},
	}))

	rule, err := client.VNetRules.Create(context.Background(), &VNetRuleCreateRequest{
		Name: "allow-http",
		VNet: 10,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(rule.Key) != 5 {
		t.Errorf("expected key 5, got %d", int(rule.Key))
	}
}

func TestVNetRuleService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRules.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRules.Create(context.Background(), &VNetRuleCreateRequest{VNet: 10})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_Create_MissingVNet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRules.Create(context.Background(), &VNetRuleCreateRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing vnet")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_Update(t *testing.T) {
	newName := "allow-https"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetRuleUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 1, Name: newName})
		},
	}))

	rule, err := client.VNetRules.Update(context.Background(), 1, &VNetRuleUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if rule.Name != newName {
		t.Errorf("expected name %q, got %q", newName, rule.Name)
	}
}

func TestVNetRuleService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRules.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetRules.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetRuleService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_rules/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetRules.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetRuleService_Enable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetRuleUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to be true")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 1, VNet: 10, Enabled: true})
		},
	}))

	err := client.VNetRules.Enable(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
}

func TestVNetRuleService_Enable_WithApply(t *testing.T) {
	applyCalled := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 1, VNet: 10, Enabled: true})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			applyCalled = true
			w.WriteHeader(200)
		},
	}))

	err := client.VNetRules.Enable(context.Background(), 1, true)
	if err != nil {
		t.Fatalf("Enable with apply failed: %v", err)
	}
	if !applyCalled {
		t.Error("expected ApplyRules to be called")
	}
}

func TestVNetRuleService_Disable(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetRuleUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled == nil || *req.Enabled {
				t.Error("expected enabled to be false")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_rules/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRule{Key: 1, VNet: 10, Enabled: false})
		},
	}))

	err := client.VNetRules.Disable(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
}

// --- VNetRuleAliasService tests ---

func TestVNetRuleAliasService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rule_aliases": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetRuleAlias{
				{Key: 1, Name: "internal-nets", Value: "10.0.0.0/8,172.16.0.0/12"},
				{Key: 2, Name: "dns-servers", Value: "8.8.8.8,8.8.4.4"},
			})
		},
	}))

	aliases, err := client.VNetRuleAliases.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(aliases))
	}
	if aliases[0].Name != "internal-nets" {
		t.Errorf("expected name 'internal-nets', got %q", aliases[0].Name)
	}
}

func TestVNetRuleAliasService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rule_aliases/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRuleAlias{Key: 1, Name: "internal-nets", Value: "10.0.0.0/8"})
		},
	}))

	alias, err := client.VNetRuleAliases.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if alias.Name != "internal-nets" {
		t.Errorf("expected name 'internal-nets', got %q", alias.Name)
	}
}

func TestVNetRuleAliasService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rule_aliases/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.VNetRuleAliases.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetRuleAliasService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rule_aliases": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetRuleAlias{{Key: 1, Name: "internal-nets"}})
		},
		"GET /api/v4/vnet_rule_aliases/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRuleAlias{Key: 1, Name: "internal-nets", Value: "10.0.0.0/8"})
		},
	}))

	alias, err := client.VNetRuleAliases.GetByName(context.Background(), "internal-nets")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if alias.Value != "10.0.0.0/8" {
		t.Errorf("expected value '10.0.0.0/8', got %q", alias.Value)
	}
}

func TestVNetRuleAliasService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_rule_aliases": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetRuleAlias{})
		},
	}))

	_, err := client.VNetRuleAliases.GetByName(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestVNetRuleAliasService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_rule_aliases": func(w http.ResponseWriter, r *http.Request) {
			var req VNetRuleAliasCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "web-servers" {
				t.Errorf("expected name 'web-servers', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 3, "status": "ok"})
		},
		"GET /api/v4/vnet_rule_aliases/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRuleAlias{Key: 3, Name: "web-servers", Value: "192.168.1.10,192.168.1.11"})
		},
	}))

	alias, err := client.VNetRuleAliases.Create(context.Background(), &VNetRuleAliasCreateRequest{
		Name:  "web-servers",
		Value: "192.168.1.10,192.168.1.11",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if int(alias.Key) != 3 {
		t.Errorf("expected key 3, got %d", int(alias.Key))
	}
}

func TestVNetRuleAliasService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRuleAliases.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleAliasService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRuleAliases.Create(context.Background(), &VNetRuleAliasCreateRequest{Value: "10.0.0.0/8"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleAliasService_Create_MissingValue(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRuleAliases.Create(context.Background(), &VNetRuleAliasCreateRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleAliasService_Update(t *testing.T) {
	newValue := "10.0.0.0/8,172.16.0.0/12"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_rule_aliases/1": func(w http.ResponseWriter, r *http.Request) {
			var req VNetRuleAliasUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Value == nil || *req.Value != newValue {
				t.Errorf("expected value %q, got %v", newValue, req.Value)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/vnet_rule_aliases/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetRuleAlias{Key: 1, Name: "internal-nets", Value: newValue})
		},
	}))

	alias, err := client.VNetRuleAliases.Update(context.Background(), 1, &VNetRuleAliasUpdateRequest{Value: &newValue})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if alias.Value != newValue {
		t.Errorf("expected value %q, got %q", newValue, alias.Value)
	}
}

func TestVNetRuleAliasService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.VNetRuleAliases.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestVNetRuleAliasService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_rule_aliases/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.VNetRuleAliases.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVNetRuleAliasService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/vnet_rule_aliases/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.VNetRuleAliases.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
