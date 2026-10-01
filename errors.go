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
	// Key is the resource id when it is a string, such as a VM import key.
	// ID stays 0 in that case.
	Key    string
	Action string
}

// Error implements the error interface.
func (e *TimeoutError) Error() string {
	id := any(e.ID)
	if e.Key != "" {
		id = e.Key
	}
	return fmt.Sprintf("vergeos: timeout waiting for %s %v to %s", e.Resource, id, e.Action)
}

// VMImportFailedError is returned when a VM import has already failed.
// Wait returns it as soon as the import row reports the failure. The
// caller does not sit through the rest of the timeout.
type VMImportFailedError struct {
	Key          string
	Name         string
	Status       string
	StatusInfo   string
	FailedDrives int
	LogLines     []string
}

// Error implements the error interface.
func (e *VMImportFailedError) Error() string {
	status := e.Status
	if status == "" {
		status = "error"
	}
	msg := fmt.Sprintf("vergeos: VM import %s failed with status %s", e.Key, status)
	if e.FailedDrives == 1 {
		msg += " (1 failed drive)"
	} else if e.FailedDrives > 0 {
		msg += fmt.Sprintf(" (%d failed drives)", e.FailedDrives)
	}
	if e.StatusInfo != "" {
		msg += ": " + e.StatusInfo
	}
	if len(e.LogLines) > 0 {
		msg += ": " + strings.Join(e.LogLines, "; ")
	}
	return msg
}

// IsVMImportFailed reports whether err is a VMImportFailedError.
func IsVMImportFailed(err error) bool {
	if err == nil {
		return false
	}
	var failed *VMImportFailedError
	return errors.As(err, &failed)
}

// VMImportInProgressError is returned by DeleteByName when several import
// rows share a name and at least one of them is still importing.
// DeleteByName does not delete any of them in that case.
type VMImportInProgressError struct {
	Name string
	Keys []string
}

// Error implements the error interface.
func (e *VMImportInProgressError) Error() string {
	if len(e.Keys) == 1 {
		return fmt.Sprintf("vergeos: VM import %s named %q is still importing", e.Keys[0], e.Name)
	}
	return fmt.Sprintf("vergeos: %d VM import records named %q are still importing (keys: %s)", len(e.Keys), e.Name, strings.Join(e.Keys, ", "))
}

// IsVMImportInProgress reports whether err is a VMImportInProgressError.
func IsVMImportInProgress(err error) bool {
	if err == nil {
		return false
	}
	var inProgress *VMImportInProgressError
	return errors.As(err, &inProgress)
}

// VMExportFailedError is returned when a VM export has already failed.
// Wait returns it as soon as the export row reports status error, or when
// the newest statistics row records errors.
type VMExportFailedError struct {
	ID              int
	Status          string
	StatusInfo      string
	Errors          int
	VirtualMachines int
	FileName        string
}

// Error implements the error interface.
func (e *VMExportFailedError) Error() string {
	if e.Errors > 0 {
		errWord := "errors"
		if e.Errors == 1 {
			errWord = "error"
		}
		msg := fmt.Sprintf("vergeos: VM export %d finished with %d %s", e.ID, e.Errors, errWord)
		if e.VirtualMachines == 1 {
			msg += " across 1 VM"
		} else if e.VirtualMachines > 1 {
			msg += fmt.Sprintf(" across %d VMs", e.VirtualMachines)
		}
		if e.FileName != "" {
			msg += " in " + e.FileName
		}
		return msg
	}
	msg := fmt.Sprintf("vergeos: VM export %d failed with status %s", e.ID, e.Status)
	if e.StatusInfo != "" {
		msg += ": " + e.StatusInfo
	}
	return msg
}

// IsVMExportFailed reports whether err is a VMExportFailedError.
func IsVMExportFailed(err error) bool {
	if err == nil {
		return false
	}
	var failed *VMExportFailedError
	return errors.As(err, &failed)
}

// IsTimeoutError returns true if the error is a TimeoutError.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

// IsNotFoundError reports whether err means the requested resource does not exist.
//
// That is a NotFoundError: a read or delete of that resource, or a name lookup
// that matched nothing. A 404 APIError is not enough. VergeOS returns 404 when
// a create or update names a related row that is missing, and the row being
// written can still exist. The APIError carries the platform message.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var notFound *NotFoundError
	return errors.As(err, &notFound)
}

// statusNotFound reports an HTTP 404, or an error already rewritten as
// NotFoundError. Get and Delete of one row use it so a missing id stays
// NotFoundError. Update does not: a 404 on a field write can name a missing
// related row while that row still exists.
func statusNotFound(err error) bool {
	if err == nil {
		return false
	}
	if IsNotFoundError(err) {
		return true
	}
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
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
