package vergeos

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTenantUIBaseURL(t *testing.T) {
	got, err := tenantUIBaseURL("https://parent.example:8443", "203.0.113.50")
	if err != nil {
		t.Fatalf("tenantUIBaseURL failed: %v", err)
	}
	if got != "https://203.0.113.50" {
		t.Fatalf("base URL = %q", got)
	}

	got, err = tenantUIBaseURL("http://127.0.0.1:9", "2001:db8::1")
	if err != nil {
		t.Fatalf("tenantUIBaseURL failed: %v", err)
	}
	if got != "http://[2001:db8::1]" {
		t.Fatalf("base URL = %q", got)
	}

	if _, err := tenantUIBaseURL("https://parent.example", "not-an-ip"); !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantService_Connect_RequiresAuthBeforeLookup(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants/7": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("lookup ran before tenant credentials were supplied")
		},
	}))

	_, err := client.Tenants.Connect(context.Background(), 7)
	if err == nil || IsValidationError(err) {
		t.Fatalf("expected an authentication error, got %v", err)
	}
}

func TestTenantService_Connect_NotRunning(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 7, Name: "customer-a", UIAddress: 9})
		},
		"GET /api/v4/tenant_status": func(w http.ResponseWriter, r *http.Request) {
			if filter := r.URL.Query().Get("filter"); filter != "tenant eq 7" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TenantStatus{{Tenant: 7, Running: false, Status: "offline"}})
		},
		"GET /api/v4/vnet_addresses/9": func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("address lookup ran for a stopped tenant")
		},
	}))

	_, err := client.Tenants.Connect(context.Background(), 7, WithCredentials("admin", "secret"))
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestTenantService_Connect_SnapshotAndMissingAddress(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 7, Name: "customer-a", IsSnapshot: true, UIAddress: 9})
		},
	}))
	_, err := client.Tenants.Connect(context.Background(), 7, WithAPIKey("tenant-key"))
	if !IsValidationError(err) {
		t.Fatalf("snapshot: expected ValidationError, got %v", err)
	}

	client = newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tenants/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Tenant{Key: 7, Name: "customer-a"})
		},
		"GET /api/v4/tenant_status": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TenantStatus{{Tenant: 7, Running: true}})
		},
	}))
	_, err = client.Tenants.Connect(context.Background(), 7, WithAPIKey("tenant-key"))
	if !IsValidationError(err) {
		t.Fatalf("missing address: expected ValidationError, got %v", err)
	}
}

func TestTenantService_Connect_UsesUIAddress(t *testing.T) {
	var tenantHost string
	var tenantAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/tenants/7":
			jsonResponse(w, 200, Tenant{Key: 7, Name: "customer-a", UIAddress: 9})
		case "/api/v4/tenant_status":
			jsonResponse(w, 200, []TenantStatus{{Tenant: 7, Running: true, Status: "online"}})
		case "/api/v4/vnet_addresses/9":
			if fields := r.URL.Query().Get("fields"); fields != vnetAddressGetFields {
				t.Errorf("unexpected address fields: %s", fields)
			}
			jsonResponse(w, 200, VNetAddress{Key: 9, IP: "203.0.113.50", Type: AddressTypeVirtual})
		case "/version.json":
			tenantHost = r.Host
			jsonResponse(w, 200, map[string]string{"version": "4.2.0"})
		case "/api/v4/clusters":
			user, pass, _ := r.BasicAuth()
			tenantAuth = user + ":" + pass
			if r.Header.Get("Authorization") == "Bearer parent-key" {
				t.Errorf("parent API key was sent to the tenant")
			}
			jsonResponse(w, 200, []map[string]int{{"$key": 1}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	parent := &Client{
		baseURL:             server.URL,
		username:            "parent",
		password:            "parent-secret",
		apiKey:              "parent-key",
		httpClient:          server.Client(),
		userAgent:           "govergeos-test",
		insecureTLS:         true,
		timeout:             12 * time.Second,
		timeoutSet:          true,
		retryConfigured:     true,
		retryPolicy:         RetryPolicy{MaxAttempts: 1},
		rateLimit:           20 * time.Millisecond,
		skipVersionCheck:    true,
		minimumMajorVersion: 27,
		powerWaitTimeout:    90 * time.Second,
		powerWaitInterval:   2 * time.Second,
	}
	initServices(parent)

	redirect := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, network, server.Listener.Addr().String())
		},
	}
	child, err := parent.Tenants.Connect(context.Background(), 7,
		WithCredentials("tenant-admin", "tenant-secret"),
		WithHTTPClient(&http.Client{Transport: redirect}),
	)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if child.baseURL != "http://203.0.113.50" {
		t.Fatalf("base URL = %q", child.baseURL)
	}
	if tenantHost != "203.0.113.50" {
		t.Fatalf("tenant host = %q", tenantHost)
	}
	if tenantAuth != "tenant-admin:tenant-secret" {
		t.Fatalf("tenant auth = %q", tenantAuth)
	}
	if child.apiKey != "" || child.username != "tenant-admin" {
		t.Fatalf("child kept the wrong credentials: user=%q key=%q", child.username, child.apiKey)
	}
	if !child.insecureTLS || !child.skipVersionCheck || child.serverVersion != "4.2.0" {
		t.Fatalf("version/tls settings = insecure %v skip %v version %q", child.insecureTLS, child.skipVersionCheck, child.serverVersion)
	}
	if child.minimumMajorVersion != 27 || child.userAgent != "govergeos-test" {
		t.Fatalf("inherited identity = major %d ua %q", child.minimumMajorVersion, child.userAgent)
	}
	if child.httpClient.Timeout != 12*time.Second || child.rateLimit != 20*time.Millisecond {
		t.Fatalf("timeout %s rate %s", child.httpClient.Timeout, child.rateLimit)
	}
	if !child.retryConfigured || child.retryPolicy.MaxAttempts != 1 {
		t.Fatalf("retry = configured %v policy %+v", child.retryConfigured, child.retryPolicy)
	}
	if child.powerWaitTimeout != 90*time.Second || child.powerWaitInterval != 2*time.Second {
		t.Fatalf("power wait = %s / %s", child.powerWaitTimeout, child.powerWaitInterval)
	}
}

func TestTenantService_ConnectByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/tenants":
			if filter := r.URL.Query().Get("filter"); filter != "name eq 'customer-a'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Tenant{{Key: 7, Name: "customer-a"}})
		case "/api/v4/tenants/7":
			jsonResponse(w, 200, Tenant{Key: 7, Name: "customer-a", UIAddress: 9})
		case "/api/v4/tenant_status":
			jsonResponse(w, 200, []TenantStatus{{Tenant: 7, Running: true}})
		case "/api/v4/vnet_addresses/9":
			jsonResponse(w, 200, VNetAddress{Key: 9, IP: "203.0.113.50"})
		case "/version.json":
			jsonResponse(w, 200, map[string]string{"version": "26.1.0"})
		case "/api/v4/clusters":
			if user, _, ok := r.BasicAuth(); !ok || user != "tenant-admin" {
				t.Errorf("unexpected basic auth")
			}
			jsonResponse(w, 200, []map[string]int{{"$key": 1}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	parent := &Client{
		baseURL:    server.URL,
		username:   "parent",
		password:   "parent-secret",
		httpClient: server.Client(),
		userAgent:  "govergeos-test",
	}
	initServices(parent)

	redirect := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, network, server.Listener.Addr().String())
		},
	}
	child, err := parent.Tenants.ConnectByName(context.Background(), "customer-a",
		WithCredentials("tenant-admin", "tenant-secret"),
		WithHTTPClient(&http.Client{Transport: redirect}),
	)
	if err != nil {
		t.Fatalf("ConnectByName failed: %v", err)
	}
	if child.baseURL != "http://203.0.113.50" || child.serverVersion != "26.1.0" {
		t.Fatalf("child = url %q version %q", child.baseURL, child.serverVersion)
	}
}
