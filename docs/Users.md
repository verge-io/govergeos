---
title: Users
description: Manage users, groups, API keys, and resource-level permissions
tags: [user, group, member, api-key, permission, access-control, authentication, authorization, rbac]
categories: [Users]
---

# Users

Manage users, groups, API keys, and resource-level permissions.

## User Management

```go
// List users
users, err := client.Users.List(ctx)

// Create a user
user, err := client.Users.Create(ctx, &vergeos.UserCreateRequest{
    Name:     "newuser",
    Password: "securepassword",
})
```

---

## Groups

```go
// List groups
groups, err := client.Groups.List(ctx)

// Get a group by ID
group, err := client.Groups.Get(ctx, groupID)

// Get a group by name
group, err := client.Groups.GetByName(ctx, "developers")

// Create a group
group, err := client.Groups.Create(ctx, &vergeos.GroupCreateRequest{
    Name:        "developers",
    Description: "Development team",
})

// Update a group
group, err := client.Groups.Update(ctx, groupID, &vergeos.GroupUpdateRequest{
    Description: ptr("Updated description"),
})

// Delete a group
err = client.Groups.Delete(ctx, groupID)

// Manage group members
member, err := client.Members.Add(ctx, groupID, "username")
```

---

## User API Keys

Manage API keys for programmatic access.

```go
// List all API keys
keys, err := client.UserAPIKeys.List(ctx)

// List API keys for a specific user
keys, err := client.UserAPIKeys.ListByUser(ctx, userID)

// Create an API key (token only returned on creation!)
key, token, err := client.UserAPIKeys.Create(ctx, &vergeos.UserAPIKeyCreateRequest{
    User:        userID,
    Name:        "automation-key",
    Description: "Key for CI/CD pipeline",
    ExpiresType: vergeos.APIKeyExpiresDate,
    Expires:     ptr(time.Now().AddDate(1, 0, 0).Unix()), // 1 year
})
fmt.Printf("Save this token (shown only once): %s\n", token)

// Create a non-expiring key with IP restrictions
key, token, err := client.UserAPIKeys.Create(ctx, &vergeos.UserAPIKeyCreateRequest{
    User:        userID,
    Name:        "restricted-key",
    ExpiresType: vergeos.APIKeyExpiresNever,
    IPAllowList: "10.0.0.0/8,192.168.1.0/24",
})

// Update an API key
key, err := client.UserAPIKeys.Update(ctx, keyID, &vergeos.UserAPIKeyUpdateRequest{
    Description: ptr("Updated description"),
    IPDenyList:  ptr("192.168.1.100"),
})

// Delete an API key
err = client.UserAPIKeys.Delete(ctx, keyID)
```

---

## Permissions

Manage resource-level access control. Permissions grant identities (users/groups) access to specific resources.

```go
// List all permissions
permissions, err := client.Permissions.List(ctx)

// List permissions for a specific identity (user or group)
permissions, err := client.Permissions.ListByIdentity(ctx, identityID)

// List permissions for a specific resource type
vmPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableVMs)

// List permissions for a specific resource instance
perms, err := client.Permissions.ListByResource(ctx, "vms", vmID)

// Get a permission by ID
perm, err := client.Permissions.Get(ctx, permID)

// Get a specific permission by identity and resource
perm, err := client.Permissions.GetByIdentityAndResource(ctx, identityID, "vms", vmID)

// Create a permission
perm, err := client.Permissions.Create(ctx, &vergeos.PermissionCreateRequest{
    Identity: userID,
    Table:    "vms",
    Row:      vmID,
    List:     ptr(true),
    Read:     ptr(true),
    Modify:   ptr(true),
    Delete:   ptr(false),
})

// Update a permission
perm, err := client.Permissions.Update(ctx, permID, &vergeos.PermissionUpdateRequest{
    Delete: ptr(true), // Add delete permission
})

// Delete a permission
err = client.Permissions.Delete(ctx, permID)

// Convenience methods for common permission patterns

// Grant read-only access
perm, err := client.Permissions.GrantReadOnly(ctx, userID, "vms", vmID)

// Grant full access (list, read, create, modify, delete)
perm, err := client.Permissions.GrantFullAccess(ctx, userID, "vms", vmID)

// Grant custom access
perm, err := client.Permissions.Grant(ctx, userID, "vms", vmID,
    true,  // read
    true,  // modify
    false, // delete
)

// Revoke all access (deletes the permission if it exists)
err = client.Permissions.Revoke(ctx, userID, "vms", vmID)
```

Common table names for permissions:
- `vergeos.PermissionTableVMs` ("vms")
- `vergeos.PermissionTableNetworks` ("vnets")
- `vergeos.PermissionTableVolumes` ("volumes")
- `vergeos.PermissionTableTenants` ("tenants")
- `vergeos.PermissionTableUsers` ("users")
- `vergeos.PermissionTableGroups` ("groups")
