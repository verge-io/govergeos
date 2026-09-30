package vergeos

import (
	"context"
	"fmt"
)

// VMRecipeService reads VM recipes and their questions.
type VMRecipeService struct {
	client *Client
}

// List returns VM recipes.
func (s *VMRecipeService) List(ctx context.Context, opts ...ListOption) ([]VMRecipe, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vmRecipeFields
	}

	var recipes []VMRecipe
	if err := s.client.get(ctx, "/vm_recipes", options.toQueryParams(), &recipes); err != nil {
		return nil, err
	}
	return recipes, nil
}

// ListByCatalog returns recipes in the catalog with this hex key.
func (s *VMRecipeService) ListByCatalog(ctx context.Context, catalogID string, opts ...ListOption) ([]VMRecipe, error) {
	if catalogID == "" {
		return nil, &ValidationError{Field: "catalog", Message: "catalog id is required"}
	}
	filter := WithFilter(fmt.Sprintf("catalog eq '%s'", escapeFilterValue(catalogID)))
	return s.List(ctx, append([]ListOption{filter}, opts...)...)
}

// Get returns the VM recipe with this hex key.
func (s *VMRecipeService) Get(ctx context.Context, id string) (*VMRecipe, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	return getHexResource[VMRecipe](ctx, s.client, "/vm_recipes", "VMRecipe", id, vmRecipeFields, func(row VMRecipe) (string, string) {
		return row.Key, row.ID
	})
}

// GetByName returns the VM recipe with this exact name.
// The same name in two catalogs is an AmbiguousNameError. ListByCatalog
// narrows the lookup to one catalog.
func (s *VMRecipeService) GetByName(ctx context.Context, name string) (*VMRecipe, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "VMRecipe", ID: name}
	}
	if err := requireUniqueName("VMRecipe", name, rows, func(row VMRecipe) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("VMRecipe", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, vmRecipeKey(rows[0]))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VMRecipe", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Questions returns the questions for a VM recipe, in form order.
//
// Each question's Type is what Deploy checks answers against. Question names
// are the answer keys.
func (s *VMRecipeService) Questions(ctx context.Context, recipeID string, opts ...ListOption) ([]RecipeQuestion, error) {
	if recipeID == "" {
		return nil, &ValidationError{Field: "recipe", Message: "recipe id is required"}
	}
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = recipeQuestionFields
	}
	ref := fmt.Sprintf("recipe eq '%s'", escapeFilterValue("vm_recipes/"+recipeID))
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

func vmRecipeKey(row VMRecipe) string {
	if row.Key != "" {
		return row.Key
	}
	return row.ID
}
