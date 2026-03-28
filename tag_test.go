package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestTagService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tags": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tag{
				{Key: 1, Name: "production"},
				{Key: 2, Name: "staging"},
			})
		},
	}))

	tags, err := client.Tags.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
}

func TestTagService_ListWithOptions(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tags": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("limit") != "5" {
				t.Errorf("expected limit=5, got %s", r.URL.Query().Get("limit"))
			}
			if r.URL.Query().Get("sort") != "-name" {
				t.Errorf("expected sort=-name, got %s", r.URL.Query().Get("sort"))
			}
			jsonResponse(w, 200, []Tag{{Key: 1, Name: "z-tag"}})
		},
	}))

	tags, err := client.Tags.List(context.Background(), WithLimit(5), WithSort("-name"))
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
}

func TestTagService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tags/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tag{Key: 1, Name: "production"})
		},
	}))

	tag, err := client.Tags.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if tag.Name != "production" {
		t.Errorf("expected name 'production', got %q", tag.Name)
	}
}

func TestTagService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tags": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tag{{Key: 1, Name: "production"}})
		},
	}))

	tag, err := client.Tags.GetByName(context.Background(), "production")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if tag.Name != "production" {
		t.Errorf("expected name 'production', got %q", tag.Name)
	}
}

func TestTagService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tags": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tag{})
		},
	}))

	_, err := client.Tags.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestTagService_ListByCategory(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tags": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Tag{{Key: 1, Name: "production"}})
		},
	}))

	tags, err := client.Tags.ListByCategory(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByCategory failed: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
}

func TestTagService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tags": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, apiResponse{Key: float64(1)})
		},
		"GET /api/v4/tags/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tag{Key: 1, Name: "new-tag", Category: 5})
		},
	}))

	tag, err := client.Tags.Create(context.Background(), &TagCreateRequest{Name: "new-tag", Category: 5})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if tag.Name != "new-tag" {
		t.Errorf("expected name 'new-tag', got %q", tag.Name)
	}
}

func TestTagService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tags.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTagService_Create_MissingCategory(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tags.Create(context.Background(), &TagCreateRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestTagService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tags.Create(context.Background(), &TagCreateRequest{Category: 5})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTagService_Update(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tags/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/tags/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tag{Key: 1, Name: "updated"})
		},
	}))

	newName := "updated"
	tag, err := client.Tags.Update(context.Background(), 1, &TagUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if tag.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", tag.Name)
	}
}

func TestTagService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Tags.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTagService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tags/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Tags.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
