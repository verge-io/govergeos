package vergeos

// TenantRecipe is a template that deploys a tenant (virtual data center).
//
// Recipe keys are 40-character hex strings, the same value as ID.
type TenantRecipe struct {
	// Key is the recipe's primary key.
	Key string `json:"$key,omitempty"`
	// ID is the same hex string as Key.
	ID string `json:"id,omitempty"`
	// Name is the recipe name.
	Name string `json:"name,omitempty"`
	// Description is the recipe description.
	Description string `json:"description,omitempty"`
	// Icon is the UI icon name.
	Icon string `json:"icon,omitempty"`
	// Version is the recipe version string.
	Version string `json:"version,omitempty"`
	// Build is the recipe build number.
	Build FlexInt `json:"build,omitempty"`
	// Catalog is the parent catalog key.
	Catalog string `json:"catalog,omitempty"`
	// CatalogDisplay is the parent catalog's display label.
	CatalogDisplay string `json:"catalog_display,omitempty"`
	// CatalogName is the parent catalog's name.
	CatalogName string `json:"catalog_name,omitempty"`
	// CatalogRepository is the parent catalog's repository key.
	CatalogRepository FlexInt `json:"catalog_repository,omitempty"`
	// Downloaded reports whether the recipe has been downloaded and can be deployed.
	// A list filter for this flag is "downloaded eq 1" or "downloaded eq 0".
	Downloaded bool `json:"downloaded,omitempty"`
	// UpdateAvailable reports whether the catalog has a newer build.
	UpdateAvailable bool `json:"update_available,omitempty"`
	// NeedsRepublish reports whether a local recipe has unpublished edits.
	NeedsRepublish bool `json:"needs_republish,omitempty"`
	// PreserveCerts reports whether SSL certificates are copied from the base tenant.
	PreserveCerts bool `json:"preserve_certs,omitempty"`
	// Tenant is the golden-image tenant key for a local recipe. It is nil when
	// the recipe has no local tenant.
	Tenant *FlexInt `json:"tenant,omitempty"`
	// TenantSnapshot is the golden-image tenant snapshot key.
	TenantSnapshot *FlexInt `json:"tenant_snapshot,omitempty"`
	// Creator is the user who created the recipe.
	Creator string `json:"creator,omitempty"`
}

// TenantRecipeDeployRequest is the input to Deploy.
type TenantRecipeDeployRequest struct {
	// Recipe is the recipe key (40-character hex).
	Recipe string `json:"recipe"`
	// Name is the name of the tenant to create.
	Name string `json:"name"`
	// Answers are checked against the recipe's questions before the request
	// is sent. Nil is an empty set. Keys the recipe does not define are refused
	// when the recipe publishes questions.
	Answers RecipeAnswers `json:"answers,omitempty"`
}

// TenantRecipeInstance is one tenant deployed from a recipe.
type TenantRecipeInstance struct {
	// Key is the instance row id.
	Key FlexInt `json:"$key,omitempty"`
	// Recipe is the recipe key this instance was deployed from.
	Recipe string `json:"recipe,omitempty"`
	// RecipeName is that recipe's name.
	RecipeName string `json:"recipe_name,omitempty"`
	// Tenant is the tenant that was created.
	Tenant FlexInt `json:"tenant,omitempty"`
	// TenantName is that tenant's name.
	TenantName string `json:"tenant_name,omitempty"`
	// Name is the instance name.
	Name string `json:"name,omitempty"`
	// Version is the recipe version at deploy time.
	Version string `json:"version,omitempty"`
	// Build is the recipe build at deploy time.
	Build FlexInt `json:"build,omitempty"`
	// Created is the creation time in microseconds.
	Created int64 `json:"created,omitempty"`
	// Modified is the last modification time in microseconds.
	Modified int64 `json:"modified,omitempty"`
	// Answers is the stored answer document. Get includes it. List does not,
	// because the document can contain credentials.
	Answers RecipeAnswerSet `json:"answers,omitempty"`
}

const (
	tenantRecipeFields = "$key,id,name,description,icon,version,build,catalog," +
		"catalog#$display as catalog_display,catalog#name as catalog_name," +
		"catalog#repository as catalog_repository,downloaded,update_available," +
		"needs_republish,preserve_certs,tenant,tenant_snapshot,creator"

	tenantRecipeInstanceListFields = "$key,recipe,recipe#name as recipe_name,tenant,tenant#name as tenant_name,name,version,build,created,modified"

	tenantRecipeInstanceGetFields = tenantRecipeInstanceListFields + ",answers"
)
