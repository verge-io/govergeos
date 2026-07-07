//go:build integration

package integration

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	vergeos "github.com/macstadium/govergeos"
)

var (
	// sharedClient is reused across all tests to avoid connection exhaustion.
	// Creating a new client for each test function causes ~50+ version check
	// requests plus connection churn, which can overwhelm the server.
	sharedClient     *vergeos.Client
	sharedClientOnce sync.Once
	sharedClientErr  error
)

// rateLimitedTransport wraps an http.RoundTripper and adds a delay between requests
// to avoid overwhelming the server with too many rapid requests.
type rateLimitedTransport struct {
	transport http.RoundTripper
	delay     time.Duration
	mu        sync.Mutex
	lastReq   time.Time
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	elapsed := time.Since(t.lastReq)
	if elapsed < t.delay {
		time.Sleep(t.delay - elapsed)
	}
	t.lastReq = time.Now()
	t.mu.Unlock()

	return t.transport.RoundTrip(req)
}

// setupTestClient returns a shared VergeOS client configured from environment variables.
// The client is created once and reused across all tests to reduce connection overhead.
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

	// Create client once and reuse - prevents connection exhaustion
	sharedClientOnce.Do(func() {
		// Create a rate-limited HTTP client to avoid overwhelming the server.
		// VergeOS has a default "Webserver max session API rate limit" of 50,
		// and rapid requests can cause "connection reset by peer" errors.
		baseTransport := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		}
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &rateLimitedTransport{
				transport: baseTransport,
				delay:     50 * time.Millisecond, // ~20 requests/sec max
			},
		}

		sharedClient, sharedClientErr = vergeos.NewClient(
			vergeos.WithEnvConfig(),
			vergeos.WithHTTPClient(httpClient),
		)
	})

	if sharedClientErr != nil {
		t.Fatalf("Failed to create client: %v", sharedClientErr)
	}

	return sharedClient
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

// ptr is a helper to create pointers to values for optional fields.
func ptr[T any](v T) *T {
	return &v
}
