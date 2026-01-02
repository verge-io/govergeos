package vergeos

// User represents a VergeOS user.
type User struct {
	// ID is the unique identifier for the user.
	ID FlexInt `json:"$key,omitempty"`
	// AuthSource is the authentication source ID.
	AuthSource int `json:"auth_source,omitempty"`
	// Name is the username.
	Name string `json:"name"`
	// RemoteName is the remote username (for external auth).
	RemoteName string `json:"remote_name,omitempty"`
	// Enabled indicates whether the user is enabled.
	Enabled bool `json:"enabled"`
	// DisplayName is the user's display name.
	DisplayName string `json:"displayname,omitempty"`
	// Email is the user's email address.
	Email string `json:"email,omitempty"`
	// Type is the user type.
	Type string `json:"type,omitempty"`
	// ChangePassword indicates whether the user must change their password.
	ChangePassword bool `json:"change_password,omitempty"`
}

// UserCreateRequest is the request body for creating a user.
type UserCreateRequest struct {
	// AuthSource is the authentication source ID.
	AuthSource int `json:"auth_source,omitempty"`
	// Name is the username (required).
	Name string `json:"name"`
	// RemoteName is the remote username (for external auth).
	RemoteName string `json:"remote_name,omitempty"`
	// Enabled indicates whether the user is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// DisplayName is the user's display name.
	DisplayName string `json:"displayname,omitempty"`
	// Email is the user's email address.
	Email string `json:"email,omitempty"`
	// Type is the user type.
	Type string `json:"type,omitempty"`
	// Password is the user's password.
	Password string `json:"password,omitempty"`
	// ChangePassword indicates whether the user must change their password.
	ChangePassword *bool `json:"change_password,omitempty"`
}

// UserUpdateRequest is the request body for updating a user.
// Note: auth_source, type, password, and change_password cannot be updated after creation.
type UserUpdateRequest struct {
	// Name is the username.
	Name *string `json:"name,omitempty"`
	// RemoteName is the remote username.
	RemoteName *string `json:"remote_name,omitempty"`
	// Enabled indicates whether the user is enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// DisplayName is the user's display name.
	DisplayName *string `json:"displayname,omitempty"`
	// Email is the user's email address.
	Email *string `json:"email,omitempty"`
}

// userListFields are the fields to request when listing users.
const userListFields = "$key,auth_source,name,remote_name,enabled,displayname,email,type,change_password"
