package vergeos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDefaultUserAgentMatchesRelease(t *testing.T) {
	if defaultUserAgent != "govergeos/0.4.0" {
		t.Fatalf("defaultUserAgent = %q, want govergeos/0.4.0", defaultUserAgent)
	}
}

// clearEnvVars clears all VERGEOS_* environment variables
func clearEnvVars() {
	_ = os.Unsetenv("VERGEOS_HOST")
	_ = os.Unsetenv("VERGEOS_USERNAME")
	_ = os.Unsetenv("VERGEOS_PASSWORD")
	_ = os.Unsetenv("VERGEOS_API_KEY")
	_ = os.Unsetenv("VERGEOS_VERIFY_SSL")
	_ = os.Unsetenv("VERGEOS_INSECURE")
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

// TestWithEnvConfigAPIKeyTakesPrecedence tests that an API key is used when
// username and password are also set.
func TestWithEnvConfigAPIKeyTakesPrecedence(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "https://example.com")
	_ = os.Setenv("VERGEOS_USERNAME", "testuser")
	_ = os.Setenv("VERGEOS_PASSWORD", "testpass")
	_ = os.Setenv("VERGEOS_API_KEY", "preferred-key")

	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
	}

	opt := WithEnvConfig()
	if err := opt(c); err != nil {
		t.Fatalf("WithEnvConfig() returned error: %v", err)
	}

	if c.apiKey != "preferred-key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "preferred-key")
	}
	if c.username != "" || c.password != "" {
		t.Errorf("username/password = %q/%q, want empty when an API key is set", c.username, c.password)
	}
}

// TestWithEnvConfigCredentialCombinations covers each way the auth
// environment variables can be combined.
func TestWithEnvConfigCredentialCombinations(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		apiKey   string
		wantUser string
		wantPass string
		wantKey  string
	}{
		{name: "username and password", username: "user", password: "pass", wantUser: "user", wantPass: "pass"},
		{name: "api key only", apiKey: "key", wantKey: "key"},
		{name: "both, api key wins", username: "user", password: "pass", apiKey: "key", wantKey: "key"},
		{name: "api key and username without password", username: "user", apiKey: "key", wantKey: "key"},
		{name: "api key and password without username", password: "pass", apiKey: "key", wantKey: "key"},
		{name: "username without password", username: "user"},
		{name: "password without username", password: "pass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			defer clearEnvVars()

			_ = os.Setenv("VERGEOS_HOST", "https://example.com")
			if tt.username != "" {
				_ = os.Setenv("VERGEOS_USERNAME", tt.username)
			}
			if tt.password != "" {
				_ = os.Setenv("VERGEOS_PASSWORD", tt.password)
			}
			if tt.apiKey != "" {
				_ = os.Setenv("VERGEOS_API_KEY", tt.apiKey)
			}

			c := &Client{httpClient: &http.Client{Timeout: defaultTimeout}}
			if err := WithEnvConfig()(c); err != nil {
				t.Fatalf("WithEnvConfig() returned error: %v", err)
			}
			if c.username != tt.wantUser || c.password != tt.wantPass || c.apiKey != tt.wantKey {
				t.Fatalf("got user=%q pass=%q key=%q, want user=%q pass=%q key=%q",
					c.username, c.password, c.apiKey, tt.wantUser, tt.wantPass, tt.wantKey)
			}
		})
	}
}

// TestWithEnvConfigHostScheme covers hosts with and without a scheme.
func TestWithEnvConfigHostScheme(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		want    string
		wantErr []string
	}{
		{name: "https url", host: "https://example.com", want: "https://example.com"},
		{name: "https url trailing slash", host: "https://example.com/", want: "https://example.com"},
		{name: "http url", host: "http://example.com", want: "http://example.com"},
		{name: "http url with port", host: "http://example.com:8000", want: "http://example.com:8000"},
		{name: "https url with port", host: "https://10.1.2.3:8443/", want: "https://10.1.2.3:8443"},
		{name: "uppercase scheme", host: "HTTPS://example.com", want: "https://example.com"},
		{name: "mixed case http", host: "HtTp://example.com", want: "http://example.com"},
		{name: "bare hostname", host: "vergeos.example.com", want: "https://vergeos.example.com"},
		{name: "bare hostname trailing slash", host: "vergeos.example.com/", want: "https://vergeos.example.com"},
		{name: "bare ip", host: "10.0.0.5", want: "https://10.0.0.5"},
		{name: "bare host and port", host: "vergeos.example.com:8443", want: "https://vergeos.example.com:8443"},
		{name: "bare ipv4 and port", host: "10.0.0.5:8443", want: "https://10.0.0.5:8443"},
		{name: "bare ipv6", host: "[2001:db8::1]", want: "https://[2001:db8::1]"},
		{name: "bare ipv6 and port", host: "[2001:db8::1]:8443", want: "https://[2001:db8::1]:8443"},
		{name: "surrounding space", host: "  example.com  ", want: "https://example.com"},
		{name: "ftp scheme", host: "ftp://example.com", wantErr: []string{"VERGEOS_HOST", "unsupported protocol scheme", "ftp"}},
		{name: "file scheme", host: "file:///tmp/x", wantErr: []string{"VERGEOS_HOST", "unsupported protocol scheme", "file"}},
		{name: "ssh scheme", host: "ssh://example.com", wantErr: []string{"VERGEOS_HOST", "unsupported protocol scheme", "ssh"}},
		{name: "scheme only", host: "https://", wantErr: []string{"VERGEOS_HOST", "missing host"}},
		{name: "whitespace only", host: "   ", wantErr: []string{"VERGEOS_HOST", "missing host"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			defer clearEnvVars()

			_ = os.Setenv("VERGEOS_HOST", tt.host)
			_ = os.Setenv("VERGEOS_API_KEY", "key")

			c := &Client{
				httpClient: &http.Client{
					Timeout:   defaultTimeout,
					Transport: &http.Transport{},
				},
			}
			err := WithEnvConfig()(c)
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatal("expected error")
				}
				for _, part := range tt.wantErr {
					if !strings.Contains(err.Error(), part) {
						t.Errorf("error = %q, want it to contain %q", err.Error(), part)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("WithEnvConfig() returned error: %v", err)
			}
			if c.baseURL != tt.want {
				t.Errorf("baseURL = %q, want %q", c.baseURL, tt.want)
			}
		})
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

			transport := configuredTransport(t, c)
			gotInsecure := transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify
			if gotInsecure != tt.insecure {
				t.Errorf("InsecureSkipVerify = %v, want %v", gotInsecure, tt.insecure)
			}
		})
	}
}

// TestWithEnvConfigTLSCombinations covers VERGEOS_VERIFY_SSL and VERGEOS_INSECURE
// alone, in agreement, and in contradiction.
func TestWithEnvConfigTLSCombinations(t *testing.T) {
	tests := []struct {
		name      string
		verifySSL string
		insecure  string
		setVerify bool
		setInsec  bool
		wantSkip  bool
		wantErr   []string
	}{
		{name: "neither set", wantSkip: false},
		{name: "verify true", setVerify: true, verifySSL: "true", wantSkip: false},
		{name: "verify TRUE", setVerify: true, verifySSL: "TRUE", wantSkip: false},
		{name: "verify 1", setVerify: true, verifySSL: "1", wantSkip: false},
		{name: "verify yes", setVerify: true, verifySSL: "yes", wantSkip: false},
		{name: "verify false", setVerify: true, verifySSL: "false", wantSkip: true},
		{name: "verify FALSE", setVerify: true, verifySSL: "FALSE", wantSkip: true},
		{name: "verify 0", setVerify: true, verifySSL: "0", wantSkip: true},
		{name: "verify False", setVerify: true, verifySSL: "False", wantSkip: true},
		{name: "insecure true", setInsec: true, insecure: "true", wantSkip: true},
		{name: "insecure TRUE", setInsec: true, insecure: "TRUE", wantSkip: true},
		{name: "insecure yes", setInsec: true, insecure: "yes", wantSkip: true},
		{name: "insecure on", setInsec: true, insecure: "on", wantSkip: true},
		{name: "insecure 1", setInsec: true, insecure: "1", wantSkip: true},
		{name: "insecure y", setInsec: true, insecure: "y", wantSkip: true},
		{name: "insecure t", setInsec: true, insecure: "t", wantSkip: true},
		{name: "insecure false", setInsec: true, insecure: "false", wantSkip: false},
		{name: "insecure FALSE", setInsec: true, insecure: "FALSE", wantSkip: false},
		{name: "insecure no", setInsec: true, insecure: "no", wantSkip: false},
		{name: "insecure off", setInsec: true, insecure: "off", wantSkip: false},
		{name: "insecure 0", setInsec: true, insecure: "0", wantSkip: false},
		{name: "insecure n", setInsec: true, insecure: "n", wantSkip: false},
		{name: "insecure f", setInsec: true, insecure: "f", wantSkip: false},
		{name: "insecure with spaces", setInsec: true, insecure: "  true  ", wantSkip: true},
		{name: "agree skip", setVerify: true, verifySSL: "false", setInsec: true, insecure: "true", wantSkip: true},
		{name: "agree skip numeric", setVerify: true, verifySSL: "0", setInsec: true, insecure: "1", wantSkip: true},
		{name: "agree verify", setVerify: true, verifySSL: "true", setInsec: true, insecure: "false", wantSkip: false},
		{name: "agree verify numeric", setVerify: true, verifySSL: "1", setInsec: true, insecure: "0", wantSkip: false},
		{
			name: "contradict both true", setVerify: true, verifySSL: "true", setInsec: true, insecure: "true",
			wantErr: []string{"VERGEOS_VERIFY_SSL", "VERGEOS_INSECURE", "contradict"},
		},
		{
			name: "contradict both false", setVerify: true, verifySSL: "false", setInsec: true, insecure: "false",
			wantErr: []string{"VERGEOS_VERIFY_SSL", "VERGEOS_INSECURE", "contradict"},
		},
		{
			name: "contradict verify on and insecure yes", setVerify: true, verifySSL: "1", setInsec: true, insecure: "yes",
			wantErr: []string{"VERGEOS_VERIFY_SSL", "VERGEOS_INSECURE", "contradict"},
		},
		{
			name: "contradict verify off and insecure no", setVerify: true, verifySSL: "0", setInsec: true, insecure: "no",
			wantErr: []string{"VERGEOS_VERIFY_SSL", "VERGEOS_INSECURE", "contradict"},
		},
		{
			name: "invalid insecure", setInsec: true, insecure: "maybe",
			wantErr: []string{"VERGEOS_INSECURE", "maybe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			defer clearEnvVars()

			_ = os.Setenv("VERGEOS_HOST", "https://example.com")
			_ = os.Setenv("VERGEOS_API_KEY", "key")
			if tt.setVerify {
				_ = os.Setenv("VERGEOS_VERIFY_SSL", tt.verifySSL)
			}
			if tt.setInsec {
				_ = os.Setenv("VERGEOS_INSECURE", tt.insecure)
			}

			c := &Client{
				httpClient: &http.Client{
					Timeout:   defaultTimeout,
					Transport: &http.Transport{},
				},
			}
			err := WithEnvConfig()(c)
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatal("expected error")
				}
				for _, part := range tt.wantErr {
					if !strings.Contains(err.Error(), part) {
						t.Errorf("error = %q, want it to contain %q", err.Error(), part)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("WithEnvConfig() returned error: %v", err)
			}

			transport := configuredTransport(t, c)
			gotSkip := transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify
			if gotSkip != tt.wantSkip {
				t.Errorf("InsecureSkipVerify = %v, want %v", gotSkip, tt.wantSkip)
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
			if err := c.applyTransportPolicy(); err != nil {
				t.Fatalf("apply transport: %v", err)
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

// startupRequest is one request observed during NewClient.
type startupRequest struct {
	path          string
	authorization string
	limit         string
	fields        string
	username      string
	password      string
	hasBasicAuth  bool
}

// newStartupServer serves the two requests NewClient makes: a public
// version document, then one authenticated clusters read.
func newStartupServer(t *testing.T, version string, clusterStatus int, clusterBody string) (*httptest.Server, *[]startupRequest) {
	t.Helper()
	return newStartupServerTLS(t, false, version, clusterStatus, clusterBody)
}

// newStartupServerTLS is newStartupServer over HTTP or HTTPS.
func newStartupServerTLS(t *testing.T, useTLS bool, version string, clusterStatus int, clusterBody string) (*httptest.Server, *[]startupRequest) {
	t.Helper()

	seen := &[]startupRequest{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, hasBasicAuth := r.BasicAuth()
		*seen = append(*seen, startupRequest{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			limit:         r.URL.Query().Get("limit"),
			fields:        r.URL.Query().Get("fields"),
			username:      username,
			password:      password,
			hasBasicAuth:  hasBasicAuth,
		})

		switch r.URL.Path {
		case "/version.json":
			jsonResponse(w, http.StatusOK, versionResponse{Version: version})
		case apiBasePath + credentialCheckEndpoint:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(clusterStatus)
			_, _ = w.Write([]byte(clusterBody))
		default:
			http.NotFound(w, r)
		}
	})
	var server *httptest.Server
	if useTLS {
		server = httptest.NewTLSServer(handler)
	} else {
		server = httptest.NewServer(handler)
	}
	t.Cleanup(server.Close)
	return server, seen
}

func credentialChecks(seen []startupRequest) []startupRequest {
	var checks []startupRequest
	for _, req := range seen {
		if req.path == apiBasePath+credentialCheckEndpoint {
			checks = append(checks, req)
		}
	}
	return checks
}

func TestNewClientWrongPasswordReturnsAuthError(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusUnauthorized, `{"err":"Login required"}`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "wrong-password"),
		WithHTTPClient(server.Client()),
	)
	if client != nil {
		t.Fatal("NewClient returned a client for a wrong password")
	}
	if !IsAuthError(err) {
		t.Fatalf("NewClient error = %v, want AuthError", err)
	}

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("NewClient error type = %T, want *AuthError", err)
	}
	if authErr.Message != "Login required" {
		t.Fatalf("AuthError.Message = %q, want %q", authErr.Message, "Login required")
	}

	if len(*seen) != 2 {
		t.Fatalf("NewClient made %d requests, want version check plus one credential check: %#v", len(*seen), *seen)
	}
	versionReq := (*seen)[0]
	if versionReq.path != "/version.json" {
		t.Fatalf("first request path = %q, want /version.json", versionReq.path)
	}
	if versionReq.authorization != "" || versionReq.hasBasicAuth {
		t.Fatalf("version check sent credentials: %#v", versionReq)
	}

	checks := credentialChecks(*seen)
	if len(checks) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(checks))
	}
	check := checks[0]
	if !check.hasBasicAuth || check.username != "admin" || check.password != "wrong-password" {
		t.Fatalf("credential check basic auth = %q:%q (present %v), want admin:wrong-password", check.username, check.password, check.hasBasicAuth)
	}
	if check.limit != "1" {
		t.Fatalf("credential check limit = %q, want 1", check.limit)
	}
	if check.fields != "$key" {
		t.Fatalf("credential check fields = %q, want $key", check.fields)
	}
}

func TestNewClientInvalidAPIKeyReturnsAuthError(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusUnauthorized, `{"err":"Login required"}`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithAPIKey("bad-key"),
		WithHTTPClient(server.Client()),
	)
	if client != nil {
		t.Fatal("NewClient returned a client for a rejected API key")
	}
	if !IsAuthError(err) {
		t.Fatalf("NewClient error = %v, want AuthError", err)
	}

	checks := credentialChecks(*seen)
	if len(checks) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(checks))
	}
	if checks[0].authorization != "Bearer bad-key" {
		t.Fatalf("Authorization = %q, want %q", checks[0].authorization, "Bearer bad-key")
	}
	if checks[0].hasBasicAuth {
		t.Fatal("API key client also sent basic auth")
	}
	if (*seen)[0].authorization != "" {
		t.Fatalf("version check sent Authorization %q", (*seen)[0].authorization)
	}
}

func TestNewClientAPIKeyWinsOverCredentials(t *testing.T) {
	orders := []struct {
		name string
		opts []ClientOption
	}{
		{
			name: "credentials then api key",
			opts: []ClientOption{
				WithCredentials("admin", "correct-password"),
				WithAPIKey("preferred-key"),
			},
		},
		{
			name: "api key then credentials",
			opts: []ClientOption{
				WithAPIKey("preferred-key"),
				WithCredentials("admin", "correct-password"),
			},
		},
	}

	for _, order := range orders {
		t.Run(order.name, func(t *testing.T) {
			server, seen := newStartupServer(t, "26.1.8", http.StatusOK, `[]`)
			opts := []ClientOption{
				WithBaseURL(server.URL),
				WithHTTPClient(server.Client()),
			}
			opts = append(opts, order.opts...)

			client, err := NewClient(opts...)
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}
			if client == nil {
				t.Fatal("NewClient returned a nil client")
			}

			checks := credentialChecks(*seen)
			if len(checks) != 1 {
				t.Fatalf("credential checks = %d, want exactly one", len(checks))
			}
			if checks[0].authorization != "Bearer preferred-key" {
				t.Fatalf("Authorization = %q, want %q", checks[0].authorization, "Bearer preferred-key")
			}
			if checks[0].hasBasicAuth {
				t.Fatal("basic auth was sent alongside the API key")
			}
		})
	}
}

func TestNewClientEnvAPIKeyWinsOverPassword(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	server, seen := newStartupServer(t, "26.1.8", http.StatusOK, `[]`)
	_ = os.Setenv("VERGEOS_HOST", server.URL)
	_ = os.Setenv("VERGEOS_USERNAME", "admin")
	_ = os.Setenv("VERGEOS_PASSWORD", "secret")
	_ = os.Setenv("VERGEOS_API_KEY", "preferred-key")

	client, err := NewClient(
		WithEnvConfig(),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned a nil client")
	}

	checks := credentialChecks(*seen)
	if len(checks) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(checks))
	}
	if checks[0].authorization != "Bearer preferred-key" {
		t.Fatalf("Authorization = %q, want %q", checks[0].authorization, "Bearer preferred-key")
	}
	if checks[0].hasBasicAuth {
		t.Fatal("basic auth was sent alongside the API key")
	}
}

// TestNewClientBareHostKeepsExplicitCredentials matches the option order
// used by the version-enforcement integration test: credentials first, then
// WithEnvConfig, so a bare VERGEOS_HOST becomes https and an API key in the
// environment does not replace the username and password.
func TestNewClientBareHostKeepsExplicitCredentials(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	server, seen := newStartupServerTLS(t, true, "26.1.8", http.StatusOK, `[]`)
	bareHost := strings.TrimPrefix(server.URL, "https://")
	if strings.Contains(bareHost, "://") {
		t.Fatalf("test host %q still has a scheme", bareHost)
	}
	_ = os.Setenv("VERGEOS_HOST", bareHost)
	_ = os.Setenv("VERGEOS_API_KEY", "env-key-must-not-win")

	client, err := NewClient(
		WithCredentials("admin", "correct-password"),
		WithEnvConfig(),
		WithInsecureTLS(true),
		WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.baseURL != "https://"+bareHost {
		t.Fatalf("baseURL = %q, want %q", client.baseURL, "https://"+bareHost)
	}
	if client.serverVersion != "26.1.8" {
		t.Fatalf("serverVersion = %q, want 26.1.8", client.serverVersion)
	}

	checks := credentialChecks(*seen)
	if len(checks) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(checks))
	}
	if !checks[0].hasBasicAuth || checks[0].username != "admin" || checks[0].password != "correct-password" {
		t.Fatalf("credential check = %#v, want the explicit username and password", checks[0])
	}
	if strings.HasPrefix(checks[0].authorization, "Bearer ") {
		t.Fatalf("Authorization = %q, want basic auth from the explicit credentials", checks[0].authorization)
	}
}

func TestNewClientEnvRejectsUnsupportedScheme(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	_ = os.Setenv("VERGEOS_HOST", "ftp://example.com")
	_ = os.Setenv("VERGEOS_API_KEY", "key")

	client, err := NewClient(WithEnvConfig())
	if client != nil {
		t.Fatal("NewClient returned a client for an unsupported host scheme")
	}
	if err == nil {
		t.Fatal("expected error")
	}
	for _, part := range []string{"VERGEOS_HOST", "unsupported protocol scheme", "ftp"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error = %q, want it to contain %q", err.Error(), part)
		}
	}
}

func TestNewClientSucceedsWhenCredentialsAccepted(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusOK, `[]`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("NewClient returned error for accepted credentials: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned a nil client")
	}
	if client.serverVersion != "26.1.8" {
		t.Fatalf("serverVersion = %q, want 26.1.8", client.serverVersion)
	}
	if client.VMs == nil || client.Clusters == nil {
		t.Fatal("NewClient returned a client without services")
	}

	checks := credentialChecks(*seen)
	if len(checks) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(checks))
	}
	if !checks[0].hasBasicAuth || checks[0].username != "admin" || checks[0].password != "correct-password" {
		t.Fatalf("credential check did not present the supplied password: %#v", checks[0])
	}
	if (*seen)[0].path != "/version.json" || (*seen)[0].authorization != "" {
		t.Fatalf("version check = %#v, want an unauthenticated /version.json request", (*seen)[0])
	}
}

func TestNewClientSkipsCredentialCheckOnUnsupportedVersion(t *testing.T) {
	server, seen := newStartupServer(t, "4.2.0", http.StatusOK, `[]`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(server.Client()),
	)
	if client != nil {
		t.Fatal("NewClient returned a client for an unsupported version")
	}
	if !IsUnsupportedVersionError(err) {
		t.Fatalf("NewClient error = %v, want UnsupportedVersionError", err)
	}
	if len(credentialChecks(*seen)) != 0 {
		t.Fatalf("credential check ran for an unsupported version: %#v", *seen)
	}
	if len(*seen) != 1 || (*seen)[0].path != "/version.json" || (*seen)[0].authorization != "" {
		t.Fatalf("startup requests = %#v, want one unauthenticated version check", *seen)
	}
}

func TestNewClientAcceptsLaterMajor(t *testing.T) {
	server, seen := newStartupServer(t, "27.0.0", http.StatusOK, `[]`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("NewClient rejected VergeOS 27: %v", err)
	}
	if client.serverVersion != "27.0.0" {
		t.Fatalf("serverVersion = %q, want 27.0.0", client.serverVersion)
	}
	if len(credentialChecks(*seen)) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(credentialChecks(*seen)))
	}
}

func TestNewClientMinimumVersionOverride(t *testing.T) {
	t.Run("rejects below the override", func(t *testing.T) {
		server, seen := newStartupServer(t, "26.1.8", http.StatusOK, `[]`)

		client, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithHTTPClient(server.Client()),
			WithMinimumVersion(27),
		)
		if client != nil {
			t.Fatal("NewClient returned a client below the configured minimum")
		}
		var unsupported *UnsupportedVersionError
		if !errors.As(err, &unsupported) {
			t.Fatalf("error = %v, want UnsupportedVersionError", err)
		}
		if unsupported.Required != 27 {
			t.Fatalf("Required = %d, want 27", unsupported.Required)
		}
		if got := err.Error(); got != "unsupported server version 26.1.8: this SDK requires VergeOS 27.0 or later" {
			t.Fatalf("error = %q", got)
		}
		if len(credentialChecks(*seen)) != 0 {
			t.Fatalf("credential check ran for a version below the override: %#v", *seen)
		}
	})

	t.Run("accepts the override and later", func(t *testing.T) {
		server, seen := newStartupServer(t, "28.0.0", http.StatusOK, `[]`)

		client, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithHTTPClient(server.Client()),
			WithMinimumVersion(27),
		)
		if err != nil {
			t.Fatalf("NewClient rejected a major above the override: %v", err)
		}
		if client.serverVersion != "28.0.0" {
			t.Fatalf("serverVersion = %q, want 28.0.0", client.serverVersion)
		}
		if len(credentialChecks(*seen)) != 1 {
			t.Fatalf("credential checks = %d, want exactly one", len(credentialChecks(*seen)))
		}
	})

	t.Run("rejects a non-positive major", func(t *testing.T) {
		_, err := NewClient(
			WithBaseURL("https://verge.example"),
			WithCredentials("admin", "correct-password"),
			WithMinimumVersion(0),
		)
		if err == nil {
			t.Fatal("expected error for minimum version 0")
		}
		if !strings.Contains(err.Error(), "minimum version must be >= 1") {
			t.Fatalf("error = %q, want minimum version validation", err.Error())
		}
	})
}

func TestNewClientSkipVersionCheck(t *testing.T) {
	t.Run("allows an older major and still checks credentials", func(t *testing.T) {
		server, seen := newStartupServer(t, "4.2.0", http.StatusOK, `[]`)

		client, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithHTTPClient(server.Client()),
			WithMinimumVersion(27),
			WithSkipVersionCheck(),
		)
		if err != nil {
			t.Fatalf("WithSkipVersionCheck rejected 4.2.0: %v", err)
		}
		if client.serverVersion != "4.2.0" {
			t.Fatalf("serverVersion = %q, want 4.2.0", client.serverVersion)
		}
		checks := credentialChecks(*seen)
		if len(checks) != 1 {
			t.Fatalf("credential checks = %d, want exactly one", len(checks))
		}
		if !checks[0].hasBasicAuth || checks[0].username != "admin" || checks[0].password != "correct-password" {
			t.Fatalf("credential check = %#v", checks[0])
		}
		if (*seen)[0].path != "/version.json" || (*seen)[0].authorization != "" {
			t.Fatalf("version check = %#v, want an unauthenticated /version.json request", (*seen)[0])
		}
	})

	t.Run("still rejects bad credentials", func(t *testing.T) {
		server, seen := newStartupServer(t, "27.0.0", http.StatusUnauthorized, `{"err":"Login required"}`)

		client, err := NewClient(
			WithBaseURL(server.URL),
			WithAPIKey("bad-key"),
			WithHTTPClient(server.Client()),
			WithSkipVersionCheck(),
		)
		if client != nil {
			t.Fatal("NewClient returned a client for a rejected API key")
		}
		if !IsAuthError(err) {
			t.Fatalf("error = %v, want AuthError", err)
		}
		if len(*seen) != 2 {
			t.Fatalf("NewClient made %d requests, want version check plus one credential check", len(*seen))
		}
	})
}

func TestNewClientForbiddenIsPermissionError(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusForbidden, `{"err":"Permission denied"}`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithAPIKey("limited-key"),
		WithHTTPClient(server.Client()),
	)
	if client != nil {
		t.Fatal("NewClient returned a client when the credential check was forbidden")
	}
	if err == nil {
		t.Fatal("NewClient returned a nil error when the credential check was forbidden")
	}
	if IsAuthError(err) {
		t.Fatalf("permission denial reported as authentication failure: %v", err)
	}
	if !IsPermissionError(err) {
		t.Fatalf("NewClient error = %v, want PermissionError", err)
	}
	if err.Error() != "vergeos: permission denied: Permission denied" {
		t.Fatalf("error = %q, want permission denied with the platform message", err.Error())
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As *APIError failed for %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", apiErr.StatusCode)
	}
	if len(credentialChecks(*seen)) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(credentialChecks(*seen)))
	}
}

func TestClientStatusErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		body           string
		wantError      string
		wantAuth       bool
		wantPermission bool
		wantConflict   bool
		wantNotFound   bool
		wantType       string
		wantAPIStatus  int // 0 when errors.As *APIError should fail
		wantEndpoint   string
		wantMessage    string
	}{
		{
			name:          "401 login required",
			status:        http.StatusUnauthorized,
			body:          `{"err":"Login required"}`,
			wantError:     "vergeos: authentication failed: Login required",
			wantAuth:      true,
			wantType:      "*vergeos.AuthError",
			wantMessage:   "Login required",
			wantAPIStatus: 0,
		},
		{
			name:           "403 permission denied",
			status:         http.StatusForbidden,
			body:           `{"err":"Permission denied"}`,
			wantError:      "vergeos: permission denied: Permission denied",
			wantPermission: true,
			wantType:       "*vergeos.PermissionError",
			wantAPIStatus:  http.StatusForbidden,
			wantEndpoint:   "/groups",
			wantMessage:    "Permission denied",
		},
		{
			name:          "409 name taken",
			status:        http.StatusConflict,
			body:          `{"err":"name already exists"}`,
			wantError:     "vergeos: conflict: name already exists",
			wantConflict:  true,
			wantType:      "*vergeos.ConflictError",
			wantAPIStatus: http.StatusConflict,
			wantEndpoint:  "/groups",
			wantMessage:   "name already exists",
		},
		{
			name:          "404 plain APIError",
			status:        http.StatusNotFound,
			body:          `{"err":"not found"}`,
			wantError:     "vergeos: API error 404 at /groups: not found",
			wantNotFound:  false,
			wantType:      "*vergeos.APIError",
			wantAPIStatus: http.StatusNotFound,
			wantEndpoint:  "/groups",
			wantMessage:   "not found",
		},
		{
			name:          "500 plain APIError",
			status:        http.StatusInternalServerError,
			body:          `{"err":"unavailable"}`,
			wantError:     "vergeos: API error 500 at /groups: unavailable",
			wantType:      "*vergeos.APIError",
			wantAPIStatus: http.StatusInternalServerError,
			wantEndpoint:  "/groups",
			wantMessage:   "unavailable",
		},
		{
			name:           "403 plain text body",
			status:         http.StatusForbidden,
			body:           "Permission denied",
			wantError:      "vergeos: permission denied: Permission denied",
			wantPermission: true,
			wantType:       "*vergeos.PermissionError",
			wantAPIStatus:  http.StatusForbidden,
			wantEndpoint:   "/groups",
			wantMessage:    "Permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.status)
					_, _ = w.Write([]byte(tt.body))
				},
			}))

			_, err := client.Groups.List(context.Background())
			if err == nil {
				t.Fatal("expected an error")
			}
			if err.Error() != tt.wantError {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantError)
			}
			if got := errorTypeName(err); got != tt.wantType {
				t.Fatalf("type = %s, want %s", got, tt.wantType)
			}
			if IsAuthError(err) != tt.wantAuth {
				t.Fatalf("IsAuthError = %v, want %v", IsAuthError(err), tt.wantAuth)
			}
			if IsPermissionError(err) != tt.wantPermission {
				t.Fatalf("IsPermissionError = %v, want %v", IsPermissionError(err), tt.wantPermission)
			}
			if IsConflictError(err) != tt.wantConflict {
				t.Fatalf("IsConflictError = %v, want %v", IsConflictError(err), tt.wantConflict)
			}
			if IsNotFoundError(err) != tt.wantNotFound {
				t.Fatalf("IsNotFoundError = %v, want %v", IsNotFoundError(err), tt.wantNotFound)
			}

			var apiErr *APIError
			gotAPI := errors.As(err, &apiErr)
			if tt.wantAPIStatus == 0 {
				if gotAPI {
					t.Fatalf("errors.As *APIError matched %T, want no match", err)
				}
				return
			}
			if !gotAPI {
				t.Fatalf("errors.As *APIError failed for %T", err)
			}
			if apiErr.StatusCode != tt.wantAPIStatus || apiErr.Endpoint != tt.wantEndpoint || apiErr.Message != tt.wantMessage {
				t.Fatalf("APIError = %+v", apiErr)
			}
		})
	}
}

func TestGetAbsoluteStatusErrorMapping(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version.json" {
			t.Fatalf("path = %q, want /version.json", r.URL.Path)
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Permission denied"))
	}))

	err := client.getAbsolute(context.Background(), "/version.json", nil, &versionResponse{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if IsAuthError(err) {
		t.Fatalf("403 reported as authentication failure: %v", err)
	}
	if !IsPermissionError(err) {
		t.Fatalf("error = %v, want PermissionError", err)
	}
	if err.Error() != "vergeos: permission denied: Permission denied" {
		t.Fatalf("error = %q", err.Error())
	}
}

func errorTypeName(err error) string {
	switch err.(type) {
	case *AuthError:
		return "*vergeos.AuthError"
	case *PermissionError:
		return "*vergeos.PermissionError"
	case *ConflictError:
		return "*vergeos.ConflictError"
	case *APIError:
		return "*vergeos.APIError"
	default:
		return ""
	}
}

func TestNewClientCredentialProbeFailureIsNotAuthError(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusInternalServerError, `{"err":"unavailable"}`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(server.Client()),
	)
	if client != nil {
		t.Fatal("NewClient returned a client when the credential check failed")
	}
	if err == nil {
		t.Fatal("NewClient returned a nil error when the credential check failed")
	}
	if IsAuthError(err) {
		t.Fatalf("server error reported as AuthError: %v", err)
	}
	if len(credentialChecks(*seen)) != 1 {
		t.Fatalf("credential checks = %d, want exactly one", len(credentialChecks(*seen)))
	}
}
