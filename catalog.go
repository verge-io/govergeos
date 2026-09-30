package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// CatalogService reads recipe catalogs.
type CatalogService struct {
	client *Client
}

// List returns catalogs.
func (s *CatalogService) List(ctx context.Context, opts ...ListOption) ([]Catalog, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = catalogFields
	}

	var catalogs []Catalog
	if err := s.client.get(ctx, "/catalogs", options.toQueryParams(), &catalogs); err != nil {
		return nil, err
	}
	return catalogs, nil
}

// Get returns the catalog with this hex key.
func (s *CatalogService) Get(ctx context.Context, id string) (*Catalog, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "id is required"}
	}
	return getHexResource[Catalog](ctx, s.client, "/catalogs", "Catalog", id, catalogFields, func(row Catalog) (string, string) {
		return row.Key, row.ID
	})
}

// GetByName returns the catalog with this exact name.
// Two catalogs with the same name is an AmbiguousNameError.
func (s *CatalogService) GetByName(ctx context.Context, name string) (*Catalog, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "Catalog", ID: name}
	}
	if err := requireUniqueName("Catalog", name, rows, func(row Catalog) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("Catalog", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, catalogKey(rows[0]))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("Catalog", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

func catalogKey(row Catalog) string {
	if row.Key != "" {
		return row.Key
	}
	return row.ID
}

// getHexResource loads one row whose id or $key is id.
// Catalogs and VM recipes are addressed by a 40-character hex string. The
// working read is a filtered list on id, which is the same string as $key.
func getHexResource[T any](ctx context.Context, client *Client, endpoint, resource, id, fields string, ids func(T) (string, string)) (*T, error) {
	params := url.Values{}
	params.Set("fields", fields)
	params.Set("filter", fmt.Sprintf("id eq '%s'", escapeFilterValue(id)))
	params.Set("limit", "2")

	var rows []T
	if err := client.get(ctx, endpoint, params, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: resource, ID: id}
	}
	if len(rows) > 1 {
		keys := make([]any, len(rows))
		for i := range rows {
			key, rowID := ids(rows[i])
			if key == "" {
				key = rowID
			}
			keys[i] = key
		}
		return nil, &AmbiguousNameError{Resource: resource, Name: id, Keys: keys}
	}
	key, rowID := ids(rows[0])
	if key != id && rowID != id {
		return nil, &NotFoundError{Resource: resource, ID: id}
	}
	return &rows[0], nil
}
