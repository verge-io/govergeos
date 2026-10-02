package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// TenantRecipeInstanceService deploys tenants from recipes and reads those deployments.
type TenantRecipeInstanceService struct {
	client *Client
}

// List returns tenant recipe instances.
// The default field set does not include answers.
func (s *TenantRecipeInstanceService) List(ctx context.Context, opts ...ListOption) ([]TenantRecipeInstance, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = tenantRecipeInstanceListFields
	}

	var instances []TenantRecipeInstance
	if err := s.client.get(ctx, "/tenant_recipe_instances", options.toQueryParams(), &instances); err != nil {
		return nil, err
	}
	return instances, nil
}

// ListByRecipe returns instances deployed from the recipe with this hex key.
func (s *TenantRecipeInstanceService) ListByRecipe(ctx context.Context, recipeID string, opts ...ListOption) ([]TenantRecipeInstance, error) {
	if recipeID == "" {
		return nil, &ValidationError{Field: "recipe", Message: "recipe id is required"}
	}
	filter := WithFilter(fmt.Sprintf("recipe eq '%s'", escapeFilterValue(recipeID)))
	return s.List(ctx, append([]ListOption{filter}, opts...)...)
}

// Get returns one recipe instance, including its stored answers.
// Answers may contain credentials.
func (s *TenantRecipeInstanceService) Get(ctx context.Context, id int) (*TenantRecipeInstance, error) {
	params := url.Values{}
	params.Set("fields", tenantRecipeInstanceGetFields)
	var instance TenantRecipeInstance
	endpoint := fmt.Sprintf("/tenant_recipe_instances/%d", id)
	if err := s.client.get(ctx, endpoint, params, &instance); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TenantRecipeInstance", ID: id}
		}
		return nil, err
	}
	return &instance, nil
}

// GetByName returns the recipe instance with this exact name.
func (s *TenantRecipeInstanceService) GetByName(ctx context.Context, name string) (*TenantRecipeInstance, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "TenantRecipeInstance", ID: name}
	}
	if err := requireUniqueName("TenantRecipeInstance", name, rows, func(row TenantRecipeInstance) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("TenantRecipeInstance", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(rows[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("TenantRecipeInstance", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Deploy creates a tenant from a recipe.
//
// Answers are checked against the recipe's questions before the request is
// sent. A bool answer has to be a boolean or one of true, yes, on, 1, false,
// no, off, 0, or the integers 0 and 1. Anything else is refused, because the
// platform would otherwise store it as false. A disksize answer is a byte
// count. Zero keeps the recipe default. Any other value under 1 MB is refused.
// Network names are resolved to vnet keys. __new_internal__ is sent as itself.
func (s *TenantRecipeInstanceService) Deploy(ctx context.Context, req *TenantRecipeDeployRequest) (*TenantRecipeInstance, error) {
	body, err := s.instanceBody(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/tenant_recipe_instances", body, &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return s.GetByName(ctx, req.Name)
	}
	return s.Get(ctx, id)
}

// instanceBody builds the POST document for Deploy.
func (s *TenantRecipeInstanceService) instanceBody(ctx context.Context, req *TenantRecipeDeployRequest) (map[string]any, error) {
	if req == nil {
		return nil, &ValidationError{Message: "request is required"}
	}
	if req.Recipe == "" {
		return nil, &ValidationError{Field: "recipe", Message: "recipe is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	questions, err := s.client.TenantRecipes.Questions(ctx, req.Recipe)
	if err != nil {
		return nil, err
	}
	var networks []Network
	if recipeAnswersNeedNetworks(questions, req.Answers) {
		networks, err = s.client.Networks.List(ctx, WithFields("$key,name"))
		if err != nil {
			return nil, err
		}
	}
	resolved, err := resolveRecipeAnswers(questions, req.Answers, networks)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"recipe": req.Recipe,
		"name":   req.Name,
	}
	if len(resolved) > 0 {
		body["answers"] = resolved
	}
	return body, nil
}
