---
title: Users
description: Manage users, groups, API keys, auth sources, OIDC applications, and resource-level permissions
tags: [user, group, member, api-key, auth-source, oidc, permission, access-control, authentication, authorization, rbac]
categories: [Users]
---

# Users

Manage users, groups, API keys, authentication sources, OIDC applications, and resource-level permissions.

`PhysicalAccess` is the Super Users membership argument. On VergeOS 26.2 it defaults to false. Setting it adds or removes that membership, which is what grants console, BMC, and hardware-level access. It is not a stored boolean of its own. Omitting it on create leaves the user out of the group.

```go
access := true
user, err := client.Users.Create(ctx, &vergeos.UserCreateRequest{
    Name: "console-admin",
    Password: "secret",
    PhysicalAccess: &access,
})
fmt.Println(user.PhysicalAccess)
```

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
fmt.Println(group.Identity) // identities-table key, used by Permissions

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

Manage API keys for programmatic access. Reads do not include `expires_type`. `Expires` is 0 when the key does not expire. Create and update still send `expires_type`.

```go
// List all API keys
keys, err := client.UserAPIKeys.List(ctx)

// List API keys for a specific user
keys, err := client.UserAPIKeys.ListByUser(ctx, userID)

// Create an API key. The secret is returned only by this call.
// VergeOS 26.2 sends it as private_key. Create returns an error when the response has no secret.
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

## Authentication Sources

External identity providers for single sign-on. `Update` reads the stored settings and merges the keys you send. A partial settings document does not delete the keys you left out, including `client_secret`. That secret is write-only: it is not returned, and printing settings redacts it.

```go
// List authentication sources
sources, err := client.AuthSources.List(ctx)

// Create an Azure AD source. The client secret is sent and not returned.
source, err := client.AuthSources.Create(ctx, &vergeos.AuthSourceCreateRequest{
    Name:   "Corporate Azure",
    Driver: vergeos.AuthSourceDriverAzure,
    Settings: vergeos.AuthSourceSettings{
        "tenant_id":     "tenant",
        "client_id":     "client",
        "client_secret": vergeos.NewWriteOnlySecret("secret"),
        "scope":         "openid profile email",
    },
    ButtonFAIcon: "bi-microsoft",
})

// Get by name
source, err = client.AuthSources.GetByName(ctx, "Corporate Azure")

// Change one setting. client_id, client_secret, and tenant_id stay.
source, err = client.AuthSources.Update(ctx, int(source.Key), &vergeos.AuthSourceUpdateRequest{
    Settings: vergeos.AuthSourceSettings{
        "scope": "openid profile email groups",
    },
})

// Delete an authentication source
err = client.AuthSources.Delete(ctx, int(source.Key))
```

---

## OIDC Applications

Applications that authenticate through VergeOS acting as an identity provider. The client secret is generated on create, returned once, and not kept on the application. Printing the secret redacts it.

```go
// List OIDC applications
apps, err := client.OIDCApplications.List(ctx)

// Create an application. secret.Value() is the generated client secret.
app, secret, err := client.OIDCApplications.Create(ctx, &vergeos.OIDCApplicationCreateRequest{
    Name:        "Tenant Portal",
    RedirectURI: "https://tenant.example.com/callback",
    Description: "OIDC for tenant authentication",
})
fmt.Println(secret)
_ = secret.Value()

// Get by name
app, err = client.OIDCApplications.GetByName(ctx, "Tenant Portal")

// Update redirect URLs. Separate multiple URLs with a newline.
app, err = client.OIDCApplications.Update(ctx, int(app.Key), &vergeos.OIDCApplicationUpdateRequest{
    RedirectURI: ptr("https://tenant.example.com/callback"),
})

// Delete an OIDC application
err = client.OIDCApplications.Delete(ctx, int(app.Key))
```

---

## Permissions

Manage resource-level access control. A permission is granted to an identity. Every user and every group has an identity key, returned as `User.Identity` and `Group.Identity` from the `identity` column. That key is a different number from the user or group row key. `Users.IdentityOf` and `Groups.IdentityOf` load it when you have the row key and not the record.

`Permissions.Create`, `Grant`, `GrantReadOnly`, `GrantFullAccess`, `Revoke`, `ListByIdentity`, and `GetByIdentityAndResource` take that `IdentityKey`. A user or group row key does not compile in that argument.

```go
user, err := client.Users.Get(ctx, userKey)
group, err := client.Groups.Get(ctx, groupKey)

// List all permissions
permissions, err := client.Permissions.List(ctx)

// List permissions for one identity
permissions, err = client.Permissions.ListByIdentity(ctx, user.Identity)

// List permissions for a specific resource type
vmPerms, err := client.Permissions.ListByTable(ctx, vergeos.PermissionTableVMs)

// List permissions for a specific resource instance
perms, err := client.Permissions.ListByResource(ctx, "vms", vmID)

// Get a permission by ID
perm, err := client.Permissions.Get(ctx, permID)

// Get a specific permission by identity and resource
perm, err = client.Permissions.GetByIdentityAndResource(ctx, user.Identity, "vms", vmID)

// Create a permission. Identity is the identities-table key.
perm, err = client.Permissions.Create(ctx, &vergeos.PermissionCreateRequest{
    Identity: user.Identity,
    Table:    "vms",
    Row:      vmID,
    List:     ptr(true),
    Read:     ptr(true),
    Modify:   ptr(true),
    Delete:   ptr(false),
})

// Update a permission
perm, err = client.Permissions.Update(ctx, permID, &vergeos.PermissionUpdateRequest{
    Delete: ptr(true), // Add delete permission
})

// Delete a permission
err = client.Permissions.Delete(ctx, permID)

// Convenience methods for common permission patterns

// From a row key, when the user or group record is not already loaded.
identity, err := client.Users.IdentityOf(ctx, userKey)
identity, err = client.Groups.IdentityOf(ctx, groupKey)

// Grant read-only access
perm, err = client.Permissions.GrantReadOnly(ctx, identity, "vms", vmID)

// Grant full access (list, read, create, modify, delete)
perm, err = client.Permissions.GrantFullAccess(ctx, group.Identity, "vms", vmID)

// Grant custom access
perm, err = client.Permissions.Grant(ctx, user.Identity, "vms", vmID,
    true,  // read
    true,  // modify
    false, // delete
)

// Revoke all access (deletes the permission if it exists)
err = client.Permissions.Revoke(ctx, user.Identity, "vms", vmID)
```

Common table names for permissions:
- `vergeos.PermissionTableVMs` ("vms")
- `vergeos.PermissionTableNetworks` ("vnets")
- `vergeos.PermissionTableVolumes` ("volumes")
- `vergeos.PermissionTableTenants` ("tenants")
- `vergeos.PermissionTableUsers` ("users")
- `vergeos.PermissionTableGroups` ("groups")
