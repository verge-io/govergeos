package vergeos

// Catalog publishing scopes. A catalog's scope is chosen when it is created
// and describes who can see its recipes.
const (
	CatalogScopePrivate = "private"
	CatalogScopeGlobal  = "global"
	CatalogScopeTenant  = "tenant"
	CatalogScopeNone    = "none"
)

// Catalog is a collection of recipes inside a catalog repository.
//
// Catalog keys are 40-character hex strings. The same value is returned as
// both Key and ID.
type Catalog struct {
	// Key is the catalog's primary key.
	Key string `json:"$key,omitempty"`
	// ID is the same hex string as Key.
	ID string `json:"id,omitempty"`
	// Repository is the parent catalog repository key.
	Repository FlexInt `json:"repository,omitempty"`
	// RepositoryName is the parent repository's name.
	RepositoryName string `json:"repository_name,omitempty"`
	// Name is the catalog name.
	Name string `json:"name,omitempty"`
	// Description is the catalog description.
	Description string `json:"description,omitempty"`
	// PublishingScope is who can see the catalog (private, global, tenant, none).
	PublishingScope string `json:"publishing_scope,omitempty"`
	// Enabled reports whether the catalog is enabled.
	Enabled bool `json:"enabled,omitempty"`
	// Created is the creation time in microseconds.
	Created int64 `json:"created,omitempty"`
}

const catalogFields = "$key,id,repository,repository#name as repository_name,name,description,publishing_scope,enabled,created"
