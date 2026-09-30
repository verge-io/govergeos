package vergeos

import (
	"encoding/json"
	"fmt"
)

// Auth source driver constants. The driver is chosen at creation and cannot
// be changed.
const (
	// AuthSourceDriverAzure is Azure Active Directory.
	AuthSourceDriverAzure = "azure"
	// AuthSourceDriverGoogle is Google OAuth.
	AuthSourceDriverGoogle = "google"
	// AuthSourceDriverGitLab is GitLab OpenID.
	AuthSourceDriverGitLab = "gitlab"
	// AuthSourceDriverOkta is Okta.
	AuthSourceDriverOkta = "okta"
	// AuthSourceDriverOpenID is a generic OpenID Connect provider.
	AuthSourceDriverOpenID = "openid"
	// AuthSourceDriverOpenIDWellKnown is OpenID Connect with discovery.
	AuthSourceDriverOpenIDWellKnown = "openid-well-known"
	// AuthSourceDriverOAuth2 is a generic OAuth2 provider.
	AuthSourceDriverOAuth2 = "oauth2"
	// AuthSourceDriverVergeIO is a parent VergeOS system.
	AuthSourceDriverVergeIO = "verge.io"
)

// AuthSource is an external identity provider used for SSO.
//
// Settings holds the driver-specific document (client_id, endpoints, scope,
// and so on). client_secret is write-only: decoding a response removes it,
// and formatting a value that still has it prints "[redacted]".
type AuthSource struct {
	// Key is the auth source ID.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the display name shown on the login button.
	Name string `json:"name,omitempty"`
	// Driver is the provider type. See AuthSourceDriver* constants.
	// The driver cannot be changed after creation.
	Driver string `json:"driver,omitempty"`
	// Settings is the driver-specific configuration. client_secret is not
	// present on values returned by the service.
	Settings AuthSourceSettings `json:"settings,omitempty"`
	// Menu reports whether the source is shown in the login dropdown.
	Menu bool `json:"menu,omitempty"`
	// Debug reports whether provider debug logging is enabled.
	Debug bool `json:"debug,omitempty"`
	// DebugTS is the debug-mode timestamp (microseconds).
	DebugTS int64 `json:"debug_ts,omitempty"`
	// ButtonBackgroundColor is the login button background (CSS color).
	ButtonBackgroundColor string `json:"button_background_color,omitempty"`
	// ButtonColor is the login button text color (CSS color).
	ButtonColor string `json:"button_color,omitempty"`
	// ButtonFAIcon is the login button icon class (for example "bi-google").
	ButtonFAIcon string `json:"button_fa_icon,omitempty"`
	// IconColor is the login button icon color (CSS color).
	IconColor string `json:"icon_color,omitempty"`
}

// AuthSourceSettings is the driver-specific settings document.
//
// The API stores this as one JSON object. client_secret inside it is
// write-only. UnmarshalJSON removes that key so a decoded auth source
// does not hold the secret. String, GoString, and Format redact it when
// a request value still has the key. MarshalJSON keeps the default map
// encoding, so a request sends the secret.
//
// A plain string or a WriteOnlySecret may be stored under client_secret.
// Both are sent on the wire and neither is printed.
type AuthSourceSettings map[string]any

// UnmarshalJSON decodes a settings object, or a string containing one,
// and drops client_secret.
func (s *AuthSourceSettings) UnmarshalJSON(data []byte) error {
	obj, err := decodeSettingsObject(data)
	if err != nil {
		return err
	}
	if obj == nil {
		*s = nil
		return nil
	}
	delete(obj, clientSecretField)
	*s = obj
	return nil
}

// String prints the settings with client_secret redacted.
func (s AuthSourceSettings) String() string {
	return fmt.Sprint(s.redactedCopy())
}

// GoString prints the settings with client_secret redacted.
func (s AuthSourceSettings) GoString() string {
	return fmt.Sprintf("%#v", s.redactedCopy())
}

// Format prints the settings with client_secret redacted.
func (s AuthSourceSettings) Format(f fmt.State, verb rune) {
	formatRedacted(f, verb, s.redactedCopy())
}

// String prints the auth source with client_secret redacted.
func (a AuthSource) String() string {
	type plain AuthSource
	return fmt.Sprintf("%v", plain(a))
}

// GoString prints the auth source with client_secret redacted.
func (a AuthSource) GoString() string {
	type plain AuthSource
	return fmt.Sprintf("%#v", plain(a))
}

// Format prints the auth source with client_secret redacted.
func (a AuthSource) Format(f fmt.State, verb rune) {
	type plain AuthSource
	formatRedacted(f, verb, plain(a))
}

// String prints the create request with client_secret redacted.
func (r AuthSourceCreateRequest) String() string {
	type plain AuthSourceCreateRequest
	return fmt.Sprintf("%v", plain(r))
}

// GoString prints the create request with client_secret redacted.
func (r AuthSourceCreateRequest) GoString() string {
	type plain AuthSourceCreateRequest
	return fmt.Sprintf("%#v", plain(r))
}

// Format prints the create request with client_secret redacted.
func (r AuthSourceCreateRequest) Format(f fmt.State, verb rune) {
	type plain AuthSourceCreateRequest
	formatRedacted(f, verb, plain(r))
}

// String prints the update request with client_secret redacted.
func (r AuthSourceUpdateRequest) String() string {
	type plain AuthSourceUpdateRequest
	return fmt.Sprintf("%v", plain(r))
}

// GoString prints the update request with client_secret redacted.
func (r AuthSourceUpdateRequest) GoString() string {
	type plain AuthSourceUpdateRequest
	return fmt.Sprintf("%#v", plain(r))
}

// Format prints the update request with client_secret redacted.
func (r AuthSourceUpdateRequest) Format(f fmt.State, verb rune) {
	type plain AuthSourceUpdateRequest
	formatRedacted(f, verb, plain(r))
}

// redactedCopy returns a map[string]any so formatting cannot recurse into
// AuthSourceSettings.Format. client_secret is replaced with a marker.
func (s AuthSourceSettings) redactedCopy() map[string]any {
	out := make(map[string]any, len(s))
	for k, v := range s {
		if k == clientSecretField && secretPresent(v) {
			out[k] = redactedSecret
			continue
		}
		out[k] = v
	}
	return out
}

// secretPresent reports whether v carries a client secret that formatting
// must hide. An empty string and a nil value are not secrets.
func secretPresent(v any) bool {
	if v == nil {
		return false
	}
	switch s := v.(type) {
	case string:
		return s != ""
	case WriteOnlySecret:
		return s.Value() != ""
	default:
		return true
	}
}

// decodeSettingsObject parses a settings document without removing
// client_secret. data may be a JSON object or a JSON string containing
// one. A null or empty document returns a nil map.
func decodeSettingsObject(data []byte) (map[string]any, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err == nil {
		return obj, nil
	}
	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil || encoded == "" {
		return nil, errSettingsShape
	}
	if err := json.Unmarshal([]byte(encoded), &obj); err != nil {
		return nil, errSettingsShape
	}
	return obj, nil
}

// AuthSourceCreateRequest is the request body for creating an auth source.
type AuthSourceCreateRequest struct {
	// Name is the display name shown on the login button (required).
	Name string `json:"name"`
	// Driver is the provider type (required). See AuthSourceDriver* constants.
	// The driver cannot be changed after creation.
	Driver string `json:"driver"`
	// Settings is the driver-specific configuration. client_secret is
	// write-only: it is sent, and it is not returned on the created source.
	Settings AuthSourceSettings `json:"settings,omitempty"`
	// Menu shows the source in the login dropdown instead of as a button.
	Menu bool `json:"menu,omitempty"`
	// Debug enables provider debug logging.
	Debug bool `json:"debug,omitempty"`
	// ButtonBackgroundColor is the login button background (CSS color).
	ButtonBackgroundColor string `json:"button_background_color,omitempty"`
	// ButtonColor is the login button text color (CSS color).
	ButtonColor string `json:"button_color,omitempty"`
	// ButtonFAIcon is the login button icon class (for example "bi-google").
	ButtonFAIcon string `json:"button_fa_icon,omitempty"`
	// IconColor is the login button icon color (CSS color).
	IconColor string `json:"icon_color,omitempty"`
}

// AuthSourceUpdateRequest is the request body for updating an auth source.
//
// The driver cannot be changed. Settings is merged with the stored document
// before the request is sent. See AuthSourceService.Update.
type AuthSourceUpdateRequest struct {
	// Name is the display name.
	Name *string `json:"name,omitempty"`
	// Settings replaces individual keys in the stored document. Omitted keys
	// are kept. client_secret is write-only and is not returned afterwards.
	Settings AuthSourceSettings `json:"settings,omitempty"`
	// Menu shows the source in the login dropdown.
	Menu *bool `json:"menu,omitempty"`
	// Debug enables provider debug logging. The platform turns debug off
	// on its own after about an hour.
	Debug *bool `json:"debug,omitempty"`
	// ButtonBackgroundColor is the login button background (CSS color).
	ButtonBackgroundColor *string `json:"button_background_color,omitempty"`
	// ButtonColor is the login button text color (CSS color).
	ButtonColor *string `json:"button_color,omitempty"`
	// ButtonFAIcon is the login button icon class.
	ButtonFAIcon *string `json:"button_fa_icon,omitempty"`
	// IconColor is the login button icon color (CSS color).
	IconColor *string `json:"icon_color,omitempty"`
}

// authSourceListFields are the fields requested when listing auth sources.
// settings is omitted so a list does not transfer client_secret.
const authSourceListFields = "$key,name,driver,menu,debug,debug_ts,button_background_color,button_color,button_fa_icon,icon_color"

// authSourceGetFields are the fields requested when getting one auth source.
// settings is included. Decoding still drops client_secret.
const authSourceGetFields = authSourceListFields + ",settings"

// authSourceSettingsFields is the projection used to read the stored
// settings document before an update merges into it.
const authSourceSettingsFields = "$key,settings"
