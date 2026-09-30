package vergeos

import (
	"errors"
	"fmt"
	"strings"
)

// APIError represents an error returned by the VergeOS API.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API.
	StatusCode int
	// Endpoint is the API endpoint that was called.
	Endpoint string
	// Message is the error message from the API.
	Message string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("vergeos: API error %d at %s: %s", e.StatusCode, e.Endpoint, e.Message)
}

// NotFoundError is returned when a resource is not found.
type NotFoundError struct {
	Resource string
	ID       any
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("vergeos: %s with ID %v not found", e.Resource, e.ID)
}

// AmbiguousNameError is returned when a name lookup matches more than one row.
// Keys lists those resource keys in list order. Callers that want one of the
// matches should use List and choose.
type AmbiguousNameError struct {
	Resource string
	Name     string
	Keys     []any
}

// Error implements the error interface.
func (e *AmbiguousNameError) Error() string {
	return fmt.Sprintf("vergeos: %s name %q matches %d objects with keys %s", e.Resource, e.Name, len(e.Keys), formatNameKeys(e.Keys))
}

func formatNameKeys(keys []any) string {
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = fmt.Sprint(key)
	}
	return strings.Join(parts, ", ")
}

// IsAmbiguousNameError returns true if err is an AmbiguousNameError.
func IsAmbiguousNameError(err error) bool {
	if err == nil {
		return false
	}
	var ambiguous *AmbiguousNameError
	return errors.As(err, &ambiguous)
}

// AuthError is returned when authentication fails (HTTP 401).
type AuthError struct {
	Message string
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("vergeos: authentication failed: %s", e.Message)
	}
	return "vergeos: authentication failed"
}

// PermissionError is returned when the caller is authenticated but the
// platform denies the request (HTTP 403).
//
// It unwraps to *APIError so errors.As can recover the status code,
// endpoint, and platform message.
type PermissionError struct {
	APIError
}

// Error implements the error interface.
func (e *PermissionError) Error() string {
	if e != nil && e.Message != "" {
		return fmt.Sprintf("vergeos: permission denied: %s", e.Message)
	}
	return "vergeos: permission denied"
}

// Unwrap exposes the underlying APIError.
func (e *PermissionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.APIError
}

// ConflictError is returned when the request conflicts with current state
// (HTTP 409), such as creating a resource whose name is already taken.
//
// It unwraps to *APIError so errors.As can recover the status code,
// endpoint, and platform message.
type ConflictError struct {
	APIError
}

// Error implements the error interface.
func (e *ConflictError) Error() string {
	if e != nil && e.Message != "" {
		return fmt.Sprintf("vergeos: conflict: %s", e.Message)
	}
	return "vergeos: conflict"
}

// Unwrap exposes the underlying APIError.
func (e *ConflictError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.APIError
}

// ValidationError is returned when request validation fails.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("vergeos: validation error on field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("vergeos: validation error: %s", e.Message)
}

// TimeoutError is returned when a polling operation exceeds its retry limit.
type TimeoutError struct {
	Resource string
	ID       int
	Action   string
}

// Error implements the error interface.
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("vergeos: timeout waiting for %s %d to %s", e.Resource, e.ID, e.Action)
}

// IsTimeoutError returns true if the error is a TimeoutError.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

// IsNotFoundError returns true if the error is a NotFoundError or a 404 API error.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var notFound *NotFoundError
	if errors.As(err, &notFound) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsAuthError returns true if the error is an AuthError or a 401 API error.
// A 403 is a permission denial; use IsPermissionError for that.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsPermissionError returns true if the error is a PermissionError or a 403 API error.
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}
	var permErr *PermissionError
	if errors.As(err, &permErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 403
	}
	return false
}

// IsConflictError returns true if the error is a ConflictError or a 409 API error.
func IsConflictError(err error) bool {
	if err == nil {
		return false
	}
	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

// RecipePreviewPersistedError means a recipe preview was accepted as a real
// deploy. The platform returned a successful HTTP status for a POST that was
// supposed to simulate. A VM with Name may exist. Preview does not return it.
type RecipePreviewPersistedError struct {
	Name string
}

// Error implements the error interface.
func (e *RecipePreviewPersistedError) Error() string {
	if e == nil {
		return "vergeos: recipe preview was accepted as a deploy"
	}
	return fmt.Sprintf("vergeos: recipe preview for %q was accepted as a deploy; a VM with that name may exist, and this call did not return it", e.Name)
}

// IsRecipePreviewPersistedError returns true if err is a RecipePreviewPersistedError.
func IsRecipePreviewPersistedError(err error) bool {
	if err == nil {
		return false
	}
	var persisted *RecipePreviewPersistedError
	return errors.As(err, &persisted)
}

// RecipePreviewFailedError means the platform finished a preview and its log
// recorded a failed step. Lines are those log lines. Preview is the report,
// which may contain guest credentials. A nil result from Preview is not a
// successful dry run.
type RecipePreviewFailedError struct {
	Lines   []string
	Preview *VMRecipePreview
}

// Error implements the error interface.
func (e *RecipePreviewFailedError) Error() string {
	if e == nil || len(e.Lines) == 0 {
		return "vergeos: recipe preview reported a failed step"
	}
	return fmt.Sprintf("vergeos: recipe preview reported a failed step: %s", strings.Join(e.Lines, "; "))
}

// IsRecipePreviewFailedError returns true if err is a RecipePreviewFailedError.
func IsRecipePreviewFailedError(err error) bool {
	if err == nil {
		return false
	}
	var failed *RecipePreviewFailedError
	return errors.As(err, &failed)
}

// IsValidationError returns true if the error is a ValidationError or a 400 API error.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 400
	}
	return false
}
