package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTagCategoryService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_categories": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TagCategory{
				{Key: 1, Name: "environment", TaggableVMs: true},
				{Key: 2, Name: "department", TaggableUsers: true},
			})
		},
	}))

	categories, err := client.TagCategories.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}
	if categories[0].Name != "environment" {
		t.Errorf("expected name 'environment', got %q", categories[0].Name)
	}
}

func TestTagCategoryService_List_WithFilter(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_categories": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter parameter")
			}
			jsonResponse(w, 200, []TagCategory{{Key: 1, Name: "environment"}})
		},
	}))

	categories, err := client.TagCategories.List(context.Background(), WithFilter("name eq 'environment'"))
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(categories))
	}
}

func TestTagCategoryService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_categories/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagCategory{Key: 1, Name: "environment", TaggableVMs: true, TaggableVNets: true})
		},
	}))

	cat, err := client.TagCategories.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cat.Name != "environment" {
		t.Errorf("expected name 'environment', got %q", cat.Name)
	}
	if !cat.TaggableVMs {
		t.Error("expected TaggableVMs to be true")
	}
}

func TestTagCategoryService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_categories/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TagCategories.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTagCategoryService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_categories": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TagCategory{{Key: 1, Name: "environment"}})
		},
	}))

	cat, err := client.TagCategories.GetByName(context.Background(), "environment")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if cat.Name != "environment" {
		t.Errorf("expected name 'environment', got %q", cat.Name)
	}
}

func TestTagCategoryService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_categories": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TagCategory{})
		},
	}))

	_, err := client.TagCategories.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTagCategoryService_Create(t *testing.T) {
	taggableVMs := true
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tag_categories": func(w http.ResponseWriter, r *http.Request) {
			var req TagCategoryCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "environment" {
				t.Errorf("expected name 'environment', got %q", req.Name)
			}
			jsonResponse(w, 200, map[string]interface{}{"$key": 1})
		},
		"GET /api/v4/tag_categories/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagCategory{Key: 1, Name: "environment", TaggableVMs: true})
		},
	}))

	cat, err := client.TagCategories.Create(context.Background(), &TagCategoryCreateRequest{
		Name:        "environment",
		TaggableVMs: &taggableVMs,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if cat.Name != "environment" {
		t.Errorf("expected name 'environment', got %q", cat.Name)
	}
}

func TestTagCategoryService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagCategories.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagCategoryService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagCategories.Create(context.Background(), &TagCategoryCreateRequest{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagCategoryService_Update(t *testing.T) {
	newName := "env"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tag_categories/1": func(w http.ResponseWriter, r *http.Request) {
			var req TagCategoryUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tag_categories/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagCategory{Key: 1, Name: newName})
		},
	}))

	cat, err := client.TagCategories.Update(context.Background(), 1, &TagCategoryUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if cat.Name != newName {
		t.Errorf("expected name %q, got %q", newName, cat.Name)
	}
}

func TestTagCategoryService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagCategories.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagCategoryService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tag_categories/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "updated"
	_, err := client.TagCategories.Update(context.Background(), 999, &TagCategoryUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTagCategoryService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tag_categories/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TagCategories.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTagCategoryService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tag_categories/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.TagCategories.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}
