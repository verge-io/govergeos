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
	if !IsAuthError(err) {
		t.Error("expected true for 403 APIError")
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
