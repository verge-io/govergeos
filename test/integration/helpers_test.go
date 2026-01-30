//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// setupTestClient creates a VergeOS client configured from environment variables.
// Uses WithEnvConfig() for clean, composable configuration.
//
// Required environment variables:
//   - VERGEOS_HOST: Base URL for the VergeOS API
//   - VERGEOS_USERNAME + VERGEOS_PASSWORD: Basic authentication
//     OR
//   - VERGEOS_API_KEY: Bearer token authentication
//
// Optional environment variables:
//   - VERGEOS_VERIFY_SSL: Verify TLS certificates (default: "true")
//   - VERGEOS_TIMEOUT: Request timeout in seconds (default: "30")
//
// For integration tests, VERGEOS_VERIFY_SSL defaults to "false" if not set,
// since most test environments use self-signed certificates.
func setupTestClient(t *testing.T) *vergeos.Client {
	t.Helper()

	// Check if required env var is set
	if os.Getenv("VERGEOS_HOST") == "" {
		t.Skip("Skipping: VERGEOS_HOST not set")
	}

	// Default VERGEOS_VERIFY_SSL to false for integration tests
	// (most test environments use self-signed certs)
	if os.Getenv("VERGEOS_VERIFY_SSL") == "" {
		os.Setenv("VERGEOS_VERIFY_SSL", "false")
	}

	client, err := vergeos.NewClient(vergeos.WithEnvConfig())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

// prettyPrint outputs a value as formatted JSON for test logging.
func prettyPrint(t *testing.T, label string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s: (failed to marshal: %v)", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(data))
}
