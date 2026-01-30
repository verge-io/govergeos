package vergeos

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// clearEnvVars clears all VERGEOS_* environment variables
func clearEnvVars() {
	_ = os.Unsetenv("VERGEOS_HOST")
	_ = os.Unsetenv("VERGEOS_USERNAME")
	_ = os.Unsetenv("VERGEOS_PASSWORD")
	_ = os.Unsetenv("VERGEOS_API_KEY")
	_ = os.Unsetenv("VERGEOS_VERIFY_SSL")
	_ = os.Unsetenv("VERGEOS_TIMEOUT")
}

// TestWithEnvConfigBasicAuth tests that basic auth credentials are read from env vars
func TestWithEnvConfigBasicAuth(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://example.com")
	_ = os.Setenv("VERGEOS_USERNAME", "testuser")
	_ = os.Setenv("VERGEOS_PASSWORD", "testpass")

	// Create a bare client and apply the option
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://example.com")
	}
	if c.username != "testuser" {
		t.Errorf("username = %q, want %q", c.username, "testuser")
	}
	if c.password != "testpass" {
		t.Errorf("password = %q, want %q", c.password, "testpass")
	}
}

// TestWithEnvConfigAPIKey tests that API key is read from env vars
func TestWithEnvConfigAPIKey(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://example.com")
	_ = os.Setenv("VERGEOS_API_KEY", "test-api-key-123")

	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	if c.apiKey != "test-api-key-123" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "test-api-key-123")
	}
	// Should not have username/password when using API key
	if c.username != "" || c.password != "" {
		t.Errorf("expected empty username/password with API key auth")
	}
}

// TestWithEnvConfigBasicAuthTakesPrecedence tests that username/password is preferred over API key
func TestWithEnvConfigBasicAuthTakesPrecedence(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://example.com")
	_ = os.Setenv("VERGEOS_USERNAME", "testuser")
	_ = os.Setenv("VERGEOS_PASSWORD", "testpass")
	_ = os.Setenv("VERGEOS_API_KEY", "should-not-use-this")

	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	if c.username != "testuser" || c.password != "testpass" {
		t.Errorf("expected username/password to be set")
	}
	if c.apiKey != "" {
		t.Errorf("apiKey = %q, expected empty when username/password provided", c.apiKey)
	}
}

// TestWithEnvConfigHostTrailingSlash tests that trailing slash is removed from host
func TestWithEnvConfigHostTrailingSlash(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://example.com/")
	_ = os.Setenv("VERGEOS_USERNAME", "user")
	_ = os.Setenv("VERGEOS_PASSWORD", "pass")

	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want %q (trailing slash should be removed)", c.baseURL, "https://example.com")
	}
}

// TestWithEnvConfigVerifySSL tests VERGEOS_VERIFY_SSL parsing
func TestWithEnvConfigVerifySSL(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		insecure bool // true if TLS verification should be disabled
	}{
		{"empty (default secure)", "", false},
		{"true", "true", false},
		{"TRUE", "TRUE", false},
		{"1", "1", false},
		{"yes", "yes", false},
		{"false", "false", true},
		{"FALSE", "FALSE", true},
		{"0", "0", true},
		{"False", "False", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			defer clearEnvVars()

			_ = os.Setenv("VERGEOS_HOST", "https://example.com")
			_ = os.Setenv("VERGEOS_USERNAME", "user")
			_ = os.Setenv("VERGEOS_PASSWORD", "pass")
			if tt.value != "" {
				_ = os.Setenv("VERGEOS_VERIFY_SSL", tt.value)
			}

			c := &Client{
				httpClient: &http.Client{
					Timeout: defaultTimeout,
					Transport: &http.Transport{
						MaxIdleConns:        100,
						MaxIdleConnsPerHost: 20,
						IdleConnTimeout:     90 * time.Second,
					},
				},
			}

			opt := WithEnvConfig()
			if err := opt(c); err != nil {
				t.Fatalf("WithEnvConfig() returned error: %v", err)
			}

			// Check if transport was modified for insecure TLS
			transport, ok := c.httpClient.Transport.(*http.Transport)
			if !ok {
				t.Fatal("expected http.Transport")
			}

			gotInsecure := transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify
			if gotInsecure != tt.insecure {
				t.Errorf("InsecureSkipVerify = %v, want %v", gotInsecure, tt.insecure)
			}
		})
	}
}

// TestWithEnvConfigTimeout tests VERGEOS_TIMEOUT parsing
func TestWithEnvConfigTimeout(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{"default (30s)", "", 30 * time.Second, false},
		{"60 seconds", "60", 60 * time.Second, false},
		{"120 seconds", "120", 120 * time.Second, false},
		{"invalid", "invalid", 0, true},
		{"negative", "-1", -1 * time.Second, false}, // strconv.Atoi accepts negative
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			defer clearEnvVars()

			_ = os.Setenv("VERGEOS_HOST", "https://example.com")
			_ = os.Setenv("VERGEOS_USERNAME", "user")
			_ = os.Setenv("VERGEOS_PASSWORD", "pass")
			if tt.value != "" {
				_ = os.Setenv("VERGEOS_TIMEOUT", tt.value)
			}

			c := &Client{
				httpClient: &http.Client{Timeout: defaultTimeout},
			}

			opt := WithEnvConfig()
			err := opt(c)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "VERGEOS_TIMEOUT") {
					t.Errorf("error = %q, expected to mention VERGEOS_TIMEOUT", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("WithEnvConfig() returned error: %v", err)
			}

			if c.httpClient.Timeout != tt.want {
				t.Errorf("Timeout = %v, want %v", c.httpClient.Timeout, tt.want)
			}
		})
	}
}

// TestWithEnvConfigDoesNotOverrideExplicit tests that env vars don't override already-set values
func TestWithEnvConfigDoesNotOverrideExplicit(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://env-host.com")
	_ = os.Setenv("VERGEOS_USERNAME", "env-user")
	_ = os.Setenv("VERGEOS_PASSWORD", "env-pass")

	c := &Client{
		baseURL:    "https://explicit-host.com",
		username:   "explicit-user",
		password:   "explicit-pass",
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	// Explicit values should not be overridden
	if c.baseURL != "https://explicit-host.com" {
		t.Errorf("baseURL = %q, want %q (should not override explicit)", c.baseURL, "https://explicit-host.com")
	}
	if c.username != "explicit-user" {
		t.Errorf("username = %q, want %q (should not override explicit)", c.username, "explicit-user")
	}
	if c.password != "explicit-pass" {
		t.Errorf("password = %q, want %q (should not override explicit)", c.password, "explicit-pass")
	}
}

// TestWithEnvConfigDoesNotOverrideAPIKey tests that env vars don't override already-set API key
func TestWithEnvConfigDoesNotOverrideAPIKey(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://env-host.com")
	_ = os.Setenv("VERGEOS_API_KEY", "env-api-key")

	c := &Client{
		apiKey:     "explicit-api-key",
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	if c.apiKey != "explicit-api-key" {
		t.Errorf("apiKey = %q, want %q (should not override explicit)", c.apiKey, "explicit-api-key")
	}
}

// TestNewClientWithEnvConfig tests full NewClient flow with env vars
// Note: This will fail with version check error since we're not connecting to a real server
func TestNewClientWithEnvConfigValidation(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	// Test missing host
	t.Run("missing host", func(t *testing.T) {
		_ = os.Setenv("VERGEOS_USERNAME", "user")
		_ = os.Setenv("VERGEOS_PASSWORD", "pass")
		defer clearEnvVars()

		_, err := NewClient(WithEnvConfig())
		if err == nil {
			t.Fatal("expected error for missing host")
		}
		if !strings.Contains(err.Error(), "base URL is required") {
			t.Errorf("error = %q, expected to mention base URL", err.Error())
		}
	})

	// Test missing auth
	t.Run("missing auth", func(t *testing.T) {
		clearEnvVars()
		_ = os.Setenv("VERGEOS_HOST", "https://example.com")

		_, err := NewClient(WithEnvConfig())
		if err == nil {
			t.Fatal("expected error for missing auth")
		}
		if !strings.Contains(err.Error(), "authentication required") {
			t.Errorf("error = %q, expected to mention authentication", err.Error())
		}
	})
}
