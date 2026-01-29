//go:build integration

package integration

import (
	"os"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestVersionEnforcement tests the mandatory version check in NewClient (ADR-016).
// This test verifies that the SDK correctly validates the server is running VergeOS 26.x.
//
// Run with:
//
//	VERGEOS_HOST=https://your-host VERGEOS_USERNAME=user VERGEOS_PASSWORD=pass \
//	  go test -tags=integration -v ./test/integration/ -run TestVersion
func TestVersionEnforcement(t *testing.T) {
	t.Run("NewClient_SucceedsOnV26Server", func(t *testing.T) {
		host := os.Getenv("VERGEOS_HOST")
		username := os.Getenv("VERGEOS_USERNAME")
		password := os.Getenv("VERGEOS_PASSWORD")

		if host == "" || username == "" || password == "" {
			t.Skip("Skipping integration test: VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD must be set")
		}

		// NewClient should succeed because the server is running v26
		client, err := vergeos.NewClient(
			vergeos.WithBaseURL(host),
			vergeos.WithCredentials(username, password),
			vergeos.WithInsecureTLS(true),
			vergeos.WithTimeout(30*time.Second),
		)
		if err != nil {
			if vergeos.IsUnsupportedVersionError(err) {
				t.Fatalf("NewClient returned UnsupportedVersionError - server is not v26: %v", err)
			}
			t.Fatalf("NewClient failed with unexpected error: %v", err)
		}

		if client == nil {
			t.Fatal("NewClient returned nil client without error")
		}

		t.Log("NewClient succeeded - server version check passed (v26 confirmed)")

		// Verify client is functional by making a simple API call
		version, err := client.System.GetVersion(nil)
		if err != nil {
			t.Fatalf("System.GetVersion failed after successful client creation: %v", err)
		}

		t.Logf("Server version: %s", version)

		// Verify the version is not empty
		if version == "" {
			t.Error("System.GetVersion returned empty version")
		}
	})

	t.Run("IsUnsupportedVersionError_Helper", func(t *testing.T) {
		// Test that the helper correctly identifies UnsupportedVersionError
		err := &vergeos.UnsupportedVersionError{
			ServerVersion: "4.2.0",
			Required:      26,
		}

		if !vergeos.IsUnsupportedVersionError(err) {
			t.Error("IsUnsupportedVersionError should return true for UnsupportedVersionError")
		}

		// Test that other errors don't match
		otherErr := &vergeos.APIError{StatusCode: 500, Message: "server error"}
		if vergeos.IsUnsupportedVersionError(otherErr) {
			t.Error("IsUnsupportedVersionError should return false for APIError")
		}

		t.Log("IsUnsupportedVersionError helper works correctly")
	})

	t.Run("UnsupportedVersionError_Message", func(t *testing.T) {
		err := &vergeos.UnsupportedVersionError{
			ServerVersion: "4.2.0",
			Required:      26,
		}

		expected := "unsupported server version 4.2.0: this SDK requires VergeOS 26.x"
		if err.Error() != expected {
			t.Errorf("UnsupportedVersionError.Error() = %q, want %q", err.Error(), expected)
		}

		t.Logf("Error message format verified: %s", err.Error())
	})

	t.Run("NewClient_FailsWithInvalidHost", func(t *testing.T) {
		// Test that NewClient fails gracefully when it can't reach the server
		_, err := vergeos.NewClient(
			vergeos.WithBaseURL("https://invalid-host-that-does-not-exist.local"),
			vergeos.WithCredentials("user", "pass"),
			vergeos.WithInsecureTLS(true),
			vergeos.WithTimeout(5*time.Second),
		)

		if err == nil {
			t.Fatal("NewClient should fail when server is unreachable")
		}

		// The error should NOT be an UnsupportedVersionError (it's a network error)
		if vergeos.IsUnsupportedVersionError(err) {
			t.Error("Network errors should not be reported as UnsupportedVersionError")
		}

		t.Logf("NewClient correctly failed with unreachable host: %v", err)
	})
}
