---
title: Tags
description: Organize resources with tags, tag categories, and tag member assignments
tags: [tag, tag-category, tag-member, label, categorize, organize, assign]
categories: [Tags]
---

# Tags

Organize and categorize resources with tags. Requires VergeOS v26+.

## Tag Management

```go
// List all tags
tags, err := client.Tags.List(ctx)

// Get a tag by ID
tag, err := client.Tags.Get(ctx, tagID)

// Get a tag by name
tag, err := client.Tags.GetByName(ctx, "production")

// List tags in a specific category
tags, err := client.Tags.ListByCategory(ctx, categoryID)

// Create a tag
tag, err := client.Tags.Create(ctx, &vergeos.TagCreateRequest{
    Category:    categoryID,
    Name:        "production",
    Description: "Production environment",
})

// Update a tag
newDesc := "Production servers"
tag, err := client.Tags.Update(ctx, tagID, &vergeos.TagUpdateRequest{
    Description: &newDesc,
})

// Delete a tag
err = client.Tags.Delete(ctx, tagID)
```

---

## Tag Categories

Tag categories organize tags and define which resource types can be tagged.

```go
// List all tag categories
categories, err := client.TagCategories.List(ctx)

// Get a category by ID
category, err := client.TagCategories.Get(ctx, categoryID)

// Get a category by name
category, err := client.TagCategories.GetByName(ctx, "Environment")

// Create a tag category
trueVal := true
category, err := client.TagCategories.Create(ctx, &vergeos.TagCategoryCreateRequest{
    Name:               "Environment",
    Description:        "Environment classification",
    SingleTagSelection: &trueVal,  // Only one tag from this category per resource
    TaggableVMs:        &trueVal,
    TaggableVNets:      &trueVal,
    TaggableVolumes:    &trueVal,
})

// Update a category
newDesc := "Updated description"
category, err := client.TagCategories.Update(ctx, categoryID, &vergeos.TagCategoryUpdateRequest{
    Description: &newDesc,
})

// Delete a category (also deletes all tags in it)
err = client.TagCategories.Delete(ctx, categoryID)
```

---

## Tag Members

Tag members represent tag assignments to resources (VMs, networks, etc.).

```go
// List all tag assignments
members, err := client.TagMembers.List(ctx)

// List tags assigned to a specific VM
members, err := client.TagMembers.ListByMember(ctx, "vms/123")

// List all resources with a specific tag
members, err := client.TagMembers.ListByTag(ctx, tagID)

// Assign a tag to a VM
member, err := client.TagMembers.Assign(ctx, tagID, "vms/123")

// Remove a tag from a VM
err = client.TagMembers.Unassign(ctx, tagID, "vms/123")

// Or use Create/Delete directly
member, err := client.TagMembers.Create(ctx, &vergeos.TagMemberCreateRequest{
    Tag:    tagID,
    Member: "vms/123",  // format: "object_type/object_id"
})
err = client.TagMembers.Delete(ctx, memberID)
```
