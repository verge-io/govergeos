package vergeos

import (
	"context"
	"fmt"
)

// TenantRecipeService reads tenant recipes and their questions.
type TenantRecipeService struct {
	client *Client
}

// List returns tenant recipes.
func (s *TenantRecipeService) List(ctx context.Context, opts ...ListOption) ([]TenantRecipe, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = tenantRecipeFields
	}

	var recipes []TenantRecipe
	if err := s.client.get(ctx, "/tenant_recipes", options.toQueryParams(), &recipes); err != nil {
		return nil, err
	}
	return recipes, nil
}

// ListByCatalog returns recipes in the catalog with this hex key.
func (s *TenantRecipeService) ListByCatalog(ctx context.Context, catalogID string, opts ...ListOption) ([]TenantRecipe, error) {
	if catalogID == "" {
		return nil, &ValidationError{Field: "catalog", Message: "catalog id is required"}
	}
	filter := WithFilter(fmt.Sprintf("catalog eq '%s'", escapeFilterValue(catalogID)))
	return s.List(ctx, append([]ListOption{filter}, opts...)...)
}

// Get returns the tenant recipe with this hex key.
func (s *TenantRecipeService) Get(ctx context.Context, id string) (*TenantRecipe, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	return getHexResource[TenantRecipe](ctx, s.client, "/tenant_recipes", "TenantRecipe", id, tenantRecipeFields, func(row TenantRecipe) (string, string) {
		return row.Key, row.ID
	})
}

// GetByName returns the tenant recipe with this exact name.
// The same name in two catalogs is an AmbiguousNameError. ListByCatalog
// narrows the lookup to one catalog.
func (s *TenantRecipeService) GetByName(ctx context.Context, name string) (*TenantRecipe, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "TenantRecipe", ID: name}
	}
	if err := requireUniqueName("TenantRecipe", name, rows, func(row TenantRecipe) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("TenantRecipe", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, tenantRecipeKey(rows[0]))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("TenantRecipe", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Questions returns the questions for a tenant recipe, in form order.
//
// Each question's Type is what Deploy checks answers against. Question names
// are the answer keys.
func (s *TenantRecipeService) Questions(ctx context.Context, recipeID string, opts ...ListOption) ([]RecipeQuestion, error) {
	if recipeID == "" {
		return nil, &ValidationError{Field: "recipe", Message: "recipe id is required"}
	}
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = recipeQuestionFields
	}
	ref := fmt.Sprintf("recipe eq '%s'", escapeFilterValue("tenant_recipes/"+recipeID))
	if options.Filter != "" {
		options.Filter = ref + " and " + options.Filter
	} else {
		options.Filter = ref
	}
	if options.Sort == "" {
		options.Sort = "orderid"
	}

	var questions []RecipeQuestion
	if err := s.client.get(ctx, "/recipe_questions", options.toQueryParams(), &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

func tenantRecipeKey(row TenantRecipe) string {
	if row.Key != "" {
		return row.Key
	}
	return row.ID
}
