//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestTagCategories tests the TagCategory service against a live VergeOS API.
func TestTagCategories(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		categories, err := client.TagCategories.List(ctx)
		if err != nil {
			t.Fatalf("TagCategories.List failed: %v", err)
		}
		t.Logf("Found %d tag categories", len(categories))

		if len(categories) == 0 {
			t.Log("No tag categories found")
			return
		}

		first := categories[0]
		t.Logf("First category: Key=%d, Name=%q, SingleTagSelection=%v, TaggableVMs=%v",
			int(first.Key), first.Name, first.SingleTagSelection, first.TaggableVMs)
		prettyPrint(t, "Sample TagCategory", first)
	})

	t.Run("Get", func(t *testing.T) {
		categories, err := client.TagCategories.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(categories) == 0 {
			t.Skip("No tag categories available")
		}

		first := categories[0]
		fetched, err := client.TagCategories.Get(ctx, int(first.Key))
		if err != nil {
			t.Fatalf("TagCategories.Get(%d) failed: %v", int(first.Key), err)
		}
		t.Logf("TagCategories.Get succeeded: Name=%q", fetched.Name)
	})

	t.Run("GetByName", func(t *testing.T) {
		categories, err := client.TagCategories.List(ctx, vergeos.WithLimit(1))
		if err != nil || len(categories) == 0 {
			t.Skip("No tag categories available")
		}

		first := categories[0]
		byName, err := client.TagCategories.GetByName(ctx, first.Name)
		if err != nil {
			t.Fatalf("TagCategories.GetByName failed: %v", err)
		}
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	})
}

// TestTagCategoriesCRUD tests Create/Update/Delete operations for TagCategories and Tags.
func TestTagCategoriesCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test tag category
	t.Log("Creating test tag category...")
	categoryName := "sdk-test-category-" + time.Now().Format("20060102-150405")
	trueVal := true
	category, err := client.TagCategories.Create(ctx, &vergeos.TagCategoryCreateRequest{
		Name:        categoryName,
		Description: "goVergeOS integration test category - safe to delete",
		TaggableVMs: &trueVal,
	})
	if err != nil {
		t.Fatalf("TagCategories.Create failed: %v", err)
	}
	categoryID := int(category.Key)
	t.Logf("Created category: [%d] %s", categoryID, category.Name)

	defer func() {
		t.Log("Cleaning up: deleting test tag category...")
		if err := client.TagCategories.Delete(ctx, categoryID); err != nil {
			t.Logf("Warning: failed to delete test category: %v", err)
		} else {
			t.Log("Test category deleted successfully")
		}
	}()

	// Update category
	newDesc := "Updated goVergeOS test category description"
	category, err = client.TagCategories.Update(ctx, categoryID, &vergeos.TagCategoryUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("TagCategories.Update failed: %v", err)
	}
	if category.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, category.Description)
	} else {
		t.Logf("Updated category description: %q", category.Description)
	}

	// Test Tags CRUD within this category
	t.Run("TagsCRUD", func(t *testing.T) {
		testTagsCRUD(t, ctx, client, categoryID)
	})
}

func testTagsCRUD(t *testing.T, ctx context.Context, client *vergeos.Client, categoryID int) {
	t.Log("Testing Tags CRUD...")

	// Create a tag
	tagName := "sdk-test-tag-" + time.Now().Format("150405")
	tag, err := client.Tags.Create(ctx, &vergeos.TagCreateRequest{
		Category:    categoryID,
		Name:        tagName,
		Description: "goVergeOS integration test tag - safe to delete",
	})
	if err != nil {
		t.Fatalf("Tags.Create failed: %v", err)
	}
	tagID := int(tag.Key)
	t.Logf("Created tag: [%d] %s (CategoryDisplay: %q)", tagID, tag.Name, tag.CategoryDisplay)

	// Read
	tag, err = client.Tags.Get(ctx, tagID)
	if err != nil {
		t.Fatalf("Tags.Get failed: %v", err)
	}
	t.Logf("Read tag: [%d] %s (Description: %q)", tagID, tag.Name, tag.Description)

	// Update
	newDesc := "Updated goVergeOS test tag description"
	tag, err = client.Tags.Update(ctx, tagID, &vergeos.TagUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Tags.Update failed: %v", err)
	}
	if tag.Description != newDesc {
		t.Errorf("Update verification failed: expected description %q, got %q", newDesc, tag.Description)
	} else {
		t.Logf("Updated tag description: %q", tag.Description)
	}

	// Test GetByName
	byName, err := client.Tags.GetByName(ctx, tagName)
	if err != nil {
		t.Errorf("Tags.GetByName failed: %v", err)
	} else {
		t.Logf("GetByName succeeded: Key=%d", int(byName.Key))
	}

	// Test ListByCategory
	tags, err := client.Tags.ListByCategory(ctx, categoryID)
	if err != nil {
		t.Errorf("Tags.ListByCategory failed: %v", err)
	} else {
		t.Logf("Found %d tags in category %d", len(tags), categoryID)
	}

	// Delete
	err = client.Tags.Delete(ctx, tagID)
	if err != nil {
		t.Fatalf("Tags.Delete failed: %v", err)
	}
	t.Log("Tag deleted successfully")

	// Verify deletion
	_, err = client.Tags.Get(ctx, tagID)
	if err == nil {
		t.Error("Expected error after deletion, but got none")
	} else if vergeos.IsNotFoundError(err) {
		t.Log("Verified: tag correctly deleted (NotFoundError)")
	} else {
		t.Logf("Got error after deletion: %v", err)
	}
}

// TestTags tests the Tags service against a live VergeOS API.
func TestTags(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		tags, err := client.Tags.List(ctx)
		if err != nil {
			t.Fatalf("Tags.List failed: %v", err)
		}
		t.Logf("Found %d tags", len(tags))

		if len(tags) > 0 {
			first := tags[0]
			t.Logf("First tag: Key=%d, Name=%q, Category=%d", int(first.Key), first.Name, int(first.Category))
		}
	})
}
