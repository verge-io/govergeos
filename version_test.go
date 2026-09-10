package vergeos

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		wantErr bool
	}{
		{"26.0.0", 26, false},
		{"v26.1.2", 26, false},
		{"26.0.0-beta1", 26, false},
		{"4.2.0", 4, false},
		{"4", 4, false},
		{"v4.2.0", 4, false},
		{"26.1.3-rc1", 26, false},
		{"", 0, true},
		{"v", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			major, err := parseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseVersion(%q) expected error, got %d", tt.input, major)
				}
				return
			}
			if err != nil {
				t.Errorf("parseVersion(%q) unexpected error: %v", tt.input, err)
			}
			if major != tt.major {
				t.Errorf("parseVersion(%q) = %d, want %d", tt.input, major, tt.major)
			}
		})
	}
}

func TestIsVersionAtLeast(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{version: "26.0.2.2", want: false},
		{version: "26.1.4", want: false},
		{version: "26.1.5", want: true},
		{version: "v26.1.5", want: true},
		{version: "26.1.5-beta1", want: false},
		{version: "26.1.5+build", want: true},
		{version: "26.2.0-dev", want: true},
		{version: "26.2", want: true},
		{version: "27", want: true},
		{version: "invalid", want: false},
		{version: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := isVersionAtLeast(tt.version, 26, 1, 5); got != tt.want {
				t.Errorf("isVersionAtLeast(%q, 26, 1, 5) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestCheckServerVersionRecordsVersion(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /version.json": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, http.StatusOK, versionResponse{Version: "26.1.5"})
		},
	}))

	if err := client.checkServerVersion(context.Background()); err != nil {
		t.Fatalf("checkServerVersion failed: %v", err)
	}
	if client.serverVersion != "26.1.5" {
		t.Fatalf("serverVersion = %q, want 26.1.5", client.serverVersion)
	}
}

func TestIsUnsupportedVersionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "UnsupportedVersionError",
			err:  &UnsupportedVersionError{ServerVersion: "4.2.0", Required: 26},
			want: true,
		},
		{
			name: "wrapped UnsupportedVersionError",
			err:  errors.New("wrapped: " + (&UnsupportedVersionError{ServerVersion: "4.2.0", Required: 26}).Error()),
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "APIError",
			err:  &APIError{StatusCode: 500, Message: "server error"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnsupportedVersionError(tt.err); got != tt.want {
				t.Errorf("IsUnsupportedVersionError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnsupportedVersionError_Error(t *testing.T) {
	err := &UnsupportedVersionError{
		ServerVersion: "4.2.0",
		Required:      26,
	}
	expected := "unsupported server version 4.2.0: this SDK requires VergeOS 26.x"
	if got := err.Error(); got != expected {
		t.Errorf("UnsupportedVersionError.Error() = %q, want %q", got, expected)
	}
}
