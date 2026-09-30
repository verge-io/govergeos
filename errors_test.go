package vergeos

import (
	"errors"
	"fmt"
	"testing"
)

// --- Error message formatting ---

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 500, Endpoint: "/api/v4/vms", Message: "internal error"}
	want := "vergeos: API error 500 at /api/v4/vms: internal error"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{Resource: "VM", ID: 42}
	want := "vergeos: VM with ID 42 not found"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestNotFoundError_Error_StringID(t *testing.T) {
	err := &NotFoundError{Resource: "VolumeBrowserJob", ID: "sha1hash"}
	want := "vergeos: VolumeBrowserJob with ID sha1hash not found"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestAmbiguousNameError_Error(t *testing.T) {
	err := &AmbiguousNameError{Resource: "Tag", Name: "zzgo-dup", Keys: []any{FlexInt(5), FlexInt(6)}}
	want := `vergeos: Tag name "zzgo-dup" matches 2 objects with keys 5, 6`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestAmbiguousNameError_Error_StringKeys(t *testing.T) {
	err := &AmbiguousNameError{Resource: "Volume", Name: "data", Keys: []any{"abc", "def"}}
	want := `vergeos: Volume name "data" matches 2 objects with keys abc, def`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestIsAmbiguousNameError_Nil(t *testing.T) {
	if IsAmbiguousNameError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsAmbiguousNameError_Direct(t *testing.T) {
	err := &AmbiguousNameError{Resource: "Tag", Name: "zzgo-dup", Keys: []any{5, 6}}
	if !IsAmbiguousNameError(err) {
		t.Error("expected true for AmbiguousNameError")
	}
}

func TestIsAmbiguousNameError_Wrapped(t *testing.T) {
	inner := &AmbiguousNameError{Resource: "Group", Name: "admins", Keys: []any{1, 2}}
	err := fmt.Errorf("lookup: %w", inner)
	if !IsAmbiguousNameError(err) {
		t.Error("expected true for wrapped AmbiguousNameError")
	}
}

func TestIsAmbiguousNameError_OtherError(t *testing.T) {
	if IsAmbiguousNameError(&NotFoundError{Resource: "Tag", ID: "zzgo-dup"}) {
		t.Error("expected false for NotFoundError")
	}
}

func TestErrorsAs_AmbiguousNameError(t *testing.T) {
	err := fmt.Errorf("outer: %w", &AmbiguousNameError{Resource: "Tag", Name: "zzgo-dup", Keys: []any{5, 6}})
	var target *AmbiguousNameError
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match wrapped AmbiguousNameError")
	}
	if target.Resource != "Tag" || target.Name != "zzgo-dup" || len(target.Keys) != 2 || target.Keys[0] != 5 || target.Keys[1] != 6 {
		t.Fatalf("unexpected details: %+v", target)
	}
}

func TestAuthError_Error(t *testing.T) {
	err := &AuthError{Message: "invalid credentials"}
	want := "vergeos: authentication failed: invalid credentials"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestAuthError_Error_Empty(t *testing.T) {
	err := &AuthError{}
	want := "vergeos: authentication failed"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestPermissionError_Error(t *testing.T) {
	err := &PermissionError{APIError: APIError{StatusCode: 403, Endpoint: "/groups", Message: "Permission denied"}}
	want := "vergeos: permission denied: Permission denied"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestPermissionError_Error_Empty(t *testing.T) {
	err := &PermissionError{}
	want := "vergeos: permission denied"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestConflictError_Error(t *testing.T) {
	err := &ConflictError{APIError: APIError{StatusCode: 409, Endpoint: "/groups", Message: "name already exists"}}
	want := "vergeos: conflict: name already exists"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestConflictError_Error_Empty(t *testing.T) {
	err := &ConflictError{}
	want := "vergeos: conflict"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestValidationError_Error_WithField(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "is required"}
	want := "vergeos: validation error on field name: is required"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestValidationError_Error_WithoutField(t *testing.T) {
	err := &ValidationError{Message: "request is required"}
	want := "vergeos: validation error: request is required"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

// --- IsNotFoundError ---

func TestIsNotFoundError_Nil(t *testing.T) {
	if IsNotFoundError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsNotFoundError_Direct(t *testing.T) {
	err := &NotFoundError{Resource: "VM", ID: 1}
	if !IsNotFoundError(err) {
		t.Error("expected true for NotFoundError")
	}
}

func TestIsNotFoundError_Wrapped(t *testing.T) {
	inner := &NotFoundError{Resource: "VM", ID: 1}
	err := fmt.Errorf("wrapped: %w", inner)
	if !IsNotFoundError(err) {
		t.Error("expected true for wrapped NotFoundError")
	}
}

func TestIsNotFoundError_APIError404(t *testing.T) {
	err := &APIError{StatusCode: 404, Endpoint: "/vms/1", Message: "not found"}
	if !IsNotFoundError(err) {
		t.Error("expected true for 404 APIError")
	}
}

func TestIsNotFoundError_APIError500(t *testing.T) {
	err := &APIError{StatusCode: 500, Endpoint: "/vms", Message: "server error"}
	if IsNotFoundError(err) {
		t.Error("expected false for 500 APIError")
	}
}

func TestIsNotFoundError_OtherError(t *testing.T) {
	err := errors.New("some random error")
	if IsNotFoundError(err) {
		t.Error("expected false for generic error")
	}
}

// --- IsAuthError ---

func TestIsAuthError_Nil(t *testing.T) {
	if IsAuthError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsAuthError_Direct(t *testing.T) {
	err := &AuthError{Message: "bad creds"}
	if !IsAuthError(err) {
		t.Error("expected true for AuthError")
	}
}

func TestIsAuthError_Wrapped(t *testing.T) {
	inner := &AuthError{Message: "expired"}
	err := fmt.Errorf("wrapped: %w", inner)
	if !IsAuthError(err) {
		t.Error("expected true for wrapped AuthError")
	}
}

func TestIsAuthError_APIError401(t *testing.T) {
	err := &APIError{StatusCode: 401, Endpoint: "/login", Message: "unauthorized"}
	if !IsAuthError(err) {
		t.Error("expected true for 401 APIError")
	}
}

func TestIsAuthError_APIError403(t *testing.T) {
	err := &APIError{StatusCode: 403, Endpoint: "/admin", Message: "forbidden"}
	if IsAuthError(err) {
		t.Error("403 is a permission denial, not an authentication failure")
	}
}

func TestIsAuthError_PermissionError(t *testing.T) {
	err := &PermissionError{APIError: APIError{StatusCode: 403, Message: "Permission denied"}}
	if IsAuthError(err) {
		t.Error("PermissionError must not be reported as an authentication failure")
	}
}

func TestIsAuthError_APIError500(t *testing.T) {
	err := &APIError{StatusCode: 500, Endpoint: "/vms", Message: "server error"}
	if IsAuthError(err) {
		t.Error("expected false for 500 APIError")
	}
}

func TestIsAuthError_OtherError(t *testing.T) {
	err := errors.New("something else")
	if IsAuthError(err) {
		t.Error("expected false for generic error")
	}
}

// --- IsPermissionError ---

func TestIsPermissionError_Nil(t *testing.T) {
	if IsPermissionError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsPermissionError_Direct(t *testing.T) {
	err := &PermissionError{APIError: APIError{StatusCode: 403, Message: "Permission denied"}}
	if !IsPermissionError(err) {
		t.Error("expected true for PermissionError")
	}
}

func TestIsPermissionError_Wrapped(t *testing.T) {
	inner := &PermissionError{APIError: APIError{StatusCode: 403, Endpoint: "/groups", Message: "Permission denied"}}
	err := fmt.Errorf("create group: %w", inner)
	if !IsPermissionError(err) {
		t.Error("expected true for wrapped PermissionError")
	}
}

func TestIsPermissionError_APIError403(t *testing.T) {
	err := &APIError{StatusCode: 403, Endpoint: "/admin", Message: "forbidden"}
	if !IsPermissionError(err) {
		t.Error("expected true for 403 APIError")
	}
}

func TestIsPermissionError_APIError401(t *testing.T) {
	err := &APIError{StatusCode: 401, Endpoint: "/login", Message: "unauthorized"}
	if IsPermissionError(err) {
		t.Error("expected false for 401 APIError")
	}
}

func TestIsPermissionError_AuthError(t *testing.T) {
	err := &AuthError{Message: "Login required"}
	if IsPermissionError(err) {
		t.Error("expected false for AuthError")
	}
}

func TestIsPermissionError_OtherError(t *testing.T) {
	err := errors.New("something else")
	if IsPermissionError(err) {
		t.Error("expected false for generic error")
	}
}

// --- IsConflictError ---

func TestIsConflictError_Nil(t *testing.T) {
	if IsConflictError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsConflictError_Direct(t *testing.T) {
	err := &ConflictError{APIError: APIError{StatusCode: 409, Message: "name already exists"}}
	if !IsConflictError(err) {
		t.Error("expected true for ConflictError")
	}
}

func TestIsConflictError_Wrapped(t *testing.T) {
	inner := &ConflictError{APIError: APIError{StatusCode: 409, Endpoint: "/groups", Message: "name already exists"}}
	err := fmt.Errorf("create group: %w", inner)
	if !IsConflictError(err) {
		t.Error("expected true for wrapped ConflictError")
	}
}

func TestIsConflictError_APIError409(t *testing.T) {
	err := &APIError{StatusCode: 409, Endpoint: "/groups", Message: "name already exists"}
	if !IsConflictError(err) {
		t.Error("expected true for 409 APIError")
	}
}

func TestIsConflictError_APIError401(t *testing.T) {
	err := &APIError{StatusCode: 401, Endpoint: "/login", Message: "unauthorized"}
	if IsConflictError(err) {
		t.Error("expected false for 401 APIError")
	}
}

func TestIsConflictError_OtherError(t *testing.T) {
	err := errors.New("something else")
	if IsConflictError(err) {
		t.Error("expected false for generic error")
	}
}

// --- IsValidationError ---

func TestIsValidationError_Nil(t *testing.T) {
	if IsValidationError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsValidationError_Direct(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "required"}
	if !IsValidationError(err) {
		t.Error("expected true for ValidationError")
	}
}

func TestIsValidationError_Wrapped(t *testing.T) {
	inner := &ValidationError{Message: "missing"}
	err := fmt.Errorf("wrapped: %w", inner)
	if !IsValidationError(err) {
		t.Error("expected true for wrapped ValidationError")
	}
}

func TestIsValidationError_APIError400(t *testing.T) {
	err := &APIError{StatusCode: 400, Endpoint: "/vms", Message: "bad request"}
	if !IsValidationError(err) {
		t.Error("expected true for 400 APIError")
	}
}

func TestIsValidationError_APIError500(t *testing.T) {
	err := &APIError{StatusCode: 500, Endpoint: "/vms", Message: "server error"}
	if IsValidationError(err) {
		t.Error("expected false for 500 APIError")
	}
}

func TestIsValidationError_OtherError(t *testing.T) {
	err := errors.New("nope")
	if IsValidationError(err) {
		t.Error("expected false for generic error")
	}
}

// --- errors.As compatibility ---

func TestErrorsAs_APIError(t *testing.T) {
	err := &APIError{StatusCode: 422, Endpoint: "/test", Message: "unprocessable"}
	var target *APIError
	if !errors.As(err, &target) {
		t.Error("errors.As should match APIError")
	}
	if target.StatusCode != 422 {
		t.Errorf("expected status 422, got %d", target.StatusCode)
	}
}

func TestErrorsAs_NotFoundError(t *testing.T) {
	err := fmt.Errorf("outer: %w", &NotFoundError{Resource: "Network", ID: 7})
	var target *NotFoundError
	if !errors.As(err, &target) {
		t.Error("errors.As should match wrapped NotFoundError")
	}
	if target.Resource != "Network" {
		t.Errorf("expected resource 'Network', got %q", target.Resource)
	}
}

func TestErrorsAs_AuthError(t *testing.T) {
	err := fmt.Errorf("outer: %w", &AuthError{Message: "token expired"})
	var target *AuthError
	if !errors.As(err, &target) {
		t.Error("errors.As should match wrapped AuthError")
	}
}

func TestErrorsAs_PermissionError_APIError(t *testing.T) {
	err := fmt.Errorf("outer: %w", &PermissionError{APIError: APIError{
		StatusCode: 403,
		Endpoint:   "/groups",
		Message:    "Permission denied",
	}})
	var permErr *PermissionError
	if !errors.As(err, &permErr) {
		t.Fatal("errors.As should match wrapped PermissionError")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should match *APIError through PermissionError")
	}
	if apiErr.StatusCode != 403 || apiErr.Endpoint != "/groups" || apiErr.Message != "Permission denied" {
		t.Fatalf("APIError = %+v", apiErr)
	}
}

func TestErrorsAs_ConflictError_APIError(t *testing.T) {
	err := fmt.Errorf("outer: %w", &ConflictError{APIError: APIError{
		StatusCode: 409,
		Endpoint:   "/groups",
		Message:    "name already exists",
	}})
	var conflictErr *ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatal("errors.As should match wrapped ConflictError")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should match *APIError through ConflictError")
	}
	if apiErr.StatusCode != 409 || apiErr.Endpoint != "/groups" || apiErr.Message != "name already exists" {
		t.Fatalf("APIError = %+v", apiErr)
	}
}

func TestErrorsAs_ValidationError(t *testing.T) {
	err := fmt.Errorf("outer: %w", &ValidationError{Field: "ram", Message: "must be positive"})
	var target *ValidationError
	if !errors.As(err, &target) {
		t.Error("errors.As should match wrapped ValidationError")
	}
	if target.Field != "ram" {
		t.Errorf("expected field 'ram', got %q", target.Field)
	}
}
