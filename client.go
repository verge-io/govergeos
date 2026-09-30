package vergeos

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultTimeout is the default HTTP request timeout.
	defaultTimeout = 30 * time.Second
	// apiBasePath is the base path for the VergeOS API.
	apiBasePath = "/api/v4"
	// defaultUserAgent is the default User-Agent header.
	defaultUserAgent = "govergeos/0.3.1"
	// maxResponseSize is the maximum response body size (100 MB).
	maxResponseSize = 100 << 20
	// credentialCheckEndpoint is a small table present on every VergeOS system.
	// A limit=1 read is enough to learn whether the credentials were accepted.
	credentialCheckEndpoint = "/clusters"
)

// Client is the VergeOS API client.
type Client struct {
	// baseURL is the base URL for the VergeOS API.
	baseURL string
	// username is the API username.
	username string
	// password is the API password.
	password string
	// apiKey is the API key for Bearer token authentication.
	// When set, requests use it instead of username and password.
	apiKey string
	// httpClient is the HTTP client used for requests.
	// WithHTTPClient stores the base here. Timeout, TLS, and transport
	// wrapping are applied to a copy after every option has run.
	httpClient *http.Client
	// timeoutSet reports that WithTimeout or VERGEOS_TIMEOUT selected timeout.
	timeoutSet bool
	// timeout is applied to a copy of httpClient when timeoutSet is true.
	timeout time.Duration
	// insecureTLS skips certificate verification when true. WithInsecureTLS
	// and the TLS environment variables record it. A later option replaces
	// an earlier one. The change is made on a cloned *http.Transport.
	insecureTLS bool
	// userAgent is the User-Agent header sent with requests.
	userAgent string
	// serverVersion is the version reported by /version.json during client initialization.
	serverVersion string
	// minimumMajorVersion overrides RequiredMajorVersion when greater than zero.
	minimumMajorVersion int
	// skipVersionCheck records serverVersion without rejecting the major.
	skipVersionCheck bool
	// retryPolicy is used when WithRetry was set. Otherwise the default policy applies.
	retryPolicy RetryPolicy
	// retryConfigured reports that WithRetry was applied.
	retryConfigured bool
	// rateLimit is the minimum time between request starts. Zero leaves spacing off.
	rateLimit time.Duration
	// powerWaitTimeout is the default budget for a VM power wait.
	// Zero uses defaultPowerWaitTimeout (150s).
	powerWaitTimeout time.Duration
	// powerWaitInterval is how often a VM power wait reads power state.
	// Zero uses defaultPowerWaitInterval (5s).
	powerWaitInterval time.Duration
	// transportReady stops applyTransportPolicy from wrapping the transport twice.
	transportReady bool

	// Services for interacting with different API resources.
	// All services implement their corresponding interfaces for mock testing.
	VMs                      VMServiceInterface
	VMSnapshots              VMSnapshotServiceInterface
	VMNICs                   VMNICServiceInterface
	VMDrives                 VMDriveServiceInterface
	VMDevices                VMDeviceServiceInterface
	VMImports                VMImportServiceInterface
	VMImportLogs             VMImportLogServiceInterface
	VMExports                VMExportServiceInterface
	Catalogs                 CatalogServiceInterface
	VMRecipes                VMRecipeServiceInterface
	VMRecipeInstances        VMRecipeInstanceServiceInterface
	Networks                 NetworkServiceInterface
	Users                    UserServiceInterface
	Members                  MemberServiceInterface
	CloudInitFiles           CloudInitServiceInterface
	Clusters                 ClusterServiceInterface
	Nodes                    NodeServiceInterface
	VGPUProfiles             VGPUProfileServiceInterface
	NodeGPUs                 NodeGPUServiceInterface
	NodeGPUStats             NodeGPUStatsServiceInterface
	NodeGPUInstances         NodeGPUInstanceServiceInterface
	NodeVGPUDevices          NodeVGPUDeviceServiceInterface
	NodeHostGPUDevices       NodeHostGPUDeviceServiceInterface
	NodeVGPUProfiles         NodeVGPUProfileServiceInterface
	NodeMemory               NodeMemoryServiceInterface
	NodeLLDPNeighbors        NodeLLDPNeighborServiceInterface
	Groups                   GroupServiceInterface
	Files                    FileServiceInterface
	ResourceGroups           ResourceGroupServiceInterface
	Settings                 SettingsServiceInterface
	System                   SystemServiceInterface
	Schema                   SchemaServiceInterface
	Tags                     TagServiceInterface
	TagCategories            TagCategoryServiceInterface
	TagMembers               TagMemberServiceInterface
	Volumes                  VolumeServiceInterface
	VNetRules                VNetRuleServiceInterface
	VNetRuleAliases          VNetRuleAliasServiceInterface
	Tenants                  TenantServiceInterface
	TenantNodes              TenantNodeServiceInterface
	TenantStorage            TenantStorageServiceInterface
	TenantStatus             TenantStatusServiceInterface
	TenantStatsHistoryShort  TenantStatsHistoryShortServiceInterface
	TenantSnapshots          TenantSnapshotServiceInterface
	TenantLayer2Networks     TenantLayer2NetworkServiceInterface
	TenantNetworkBlocks      TenantNetworkBlockServiceInterface
	TenantExternalIPs        TenantExternalIPServiceInterface
	SnapshotProfiles         SnapshotProfileServiceInterface
	SnapshotProfilePeriods   SnapshotProfilePeriodServiceInterface
	Alarms                   AlarmServiceInterface
	AlarmTypes               AlarmTypeServiceInterface
	Tasks                    TaskServiceInterface
	TaskSchedules            TaskScheduleServiceInterface
	TaskScheduleTriggers     TaskScheduleTriggerServiceInterface
	TaskEvents               TaskEventServiceInterface
	TaskScripts              TaskScriptServiceInterface
	VNetAddresses            VNetAddressServiceInterface
	VNetProxies              VNetProxyServiceInterface
	VNetProxyTenants         VNetProxyTenantServiceInterface
	VNetDNSViews             VNetDNSViewServiceInterface
	VNetDNSZones             VNetDNSZoneServiceInterface
	VNetDNSRecords           VNetDNSRecordServiceInterface
	VNetHosts                VNetHostServiceInterface
	VNetBGP                  VNetBGPServiceInterface
	VNetBGPRouters           VNetBGPRouterServiceInterface
	VNetBGPRouterCommands    VNetBGPRouterCommandServiceInterface
	VNetBGPInterfaces        VNetBGPInterfaceServiceInterface
	VNetBGPInterfaceCommands VNetBGPInterfaceCommandServiceInterface
	VNetBGPRouteMaps         VNetBGPRouteMapServiceInterface
	VNetBGPRouteMapCommands  VNetBGPRouteMapCommandServiceInterface
	VNetBGPIPCommands        VNetBGPIPCommandServiceInterface
	VNetOSPFCommands         VNetOSPFCommandServiceInterface
	VNetEIGRPRouters         VNetEIGRPRouterServiceInterface
	VNetEIGRPRouterCommands  VNetEIGRPRouterCommandServiceInterface
	VNetWireGuards           VNetWireGuardServiceInterface
	VNetWireGuardPeers       VNetWireGuardPeerServiceInterface
	VNetWireGuardPeerStatus  VNetWireGuardPeerStatusServiceInterface
	Certificates             CertificateServiceInterface
	VNetIPSecs               VNetIPSecServiceInterface
	VNetIPSecPhase1s         VNetIPSecPhase1ServiceInterface
	VNetIPSecPhase2s         VNetIPSecPhase2ServiceInterface
	VNetIPSecConnections     VNetIPSecConnectionServiceInterface
	Sites                    SiteServiceInterface
	SiteSyncsIncoming        SiteSyncIncomingServiceInterface
	SiteSyncsOutgoing        SiteSyncOutgoingServiceInterface
	SiteSyncProfilePeriods   SiteSyncProfilePeriodServiceInterface
	CloudSnapshots           CloudSnapshotServiceInterface
	CloudSnapshotVMs         CloudSnapshotVMServiceInterface
	CloudSnapshotTenants     CloudSnapshotTenantServiceInterface
	VolumeCIFSShares         VolumeCIFSShareServiceInterface
	VolumeNFSShares          VolumeNFSShareServiceInterface
	VolumeBrowser            VolumeBrowserServiceInterface
	WebhookURLs              WebhookURLServiceInterface
	Webhooks                 WebhookServiceInterface
	UserAPIKeys              UserAPIKeyServiceInterface
	AuthSources              AuthSourceServiceInterface
	OIDCApplications         OIDCApplicationServiceInterface
	NASServices              NASServiceServiceInterface
	NASServiceUsers          NASServiceUserServiceInterface
	VolumeSyncs              VolumeSyncServiceInterface
	VolumeSnapshots          VolumeSnapshotServiceInterface
	Permissions              PermissionServiceInterface
	Logs                     LogServiceInterface
	StorageTiers             StorageTierServiceInterface
	ClusterTiers             ClusterTierServiceInterface
	MachineDrivePhys         MachineDrivePhysServiceInterface
	ClusterStatsHistory      ClusterStatsHistoryServiceInterface
	MachineStatus            MachineStatusServiceInterface
	MachineStats             MachineStatsServiceInterface
	MachineDriveStats        MachineDriveStatsServiceInterface
	MachineNICs              MachineNICServiceInterface
	UpdateSettings           UpdateSettingsServiceInterface
	UpdateBranches           UpdateBranchServiceInterface
	UpdateSourcePackages     UpdateSourcePackageServiceInterface
	Billing                  BillingServiceInterface
	NASServiceAntivirus      NASServiceAntivirusServiceInterface
	SharedObjects            SharedObjectServiceInterface
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client) error

// WithBaseURL sets the base URL for the VergeOS API.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		// Remove trailing slash if present
		c.baseURL = strings.TrimSuffix(baseURL, "/")
		return nil
	}
}

// WithCredentials sets the username and password for API authentication.
// This uses HTTP Basic Authentication.
//
// Requests use an API key instead when one is also configured, whether
// that key came from WithAPIKey or from VERGEOS_API_KEY via WithEnvConfig.
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) error {
		c.username = username
		c.password = password
		return nil
	}
}

// WithAPIKey sets the API key for Bearer token authentication.
//
// The API key is used even when username and password are also set.
// WithEnvConfig applies the same rule: VERGEOS_API_KEY wins over
// VERGEOS_USERNAME and VERGEOS_PASSWORD.
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) error {
		c.apiKey = apiKey
		return nil
	}
}

// defaultTransport returns the transport used when the caller does not
// supply one.
func defaultTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
}

// defaultHTTPClient is the client used when WithHTTPClient is not set.
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: defaultTransport(),
	}
}

// cloneHTTPTransport copies the exported fields of t.
//
// http.Transport.Clone also enables HTTP/2 on t, and that writes a TLS
// config into the caller's transport. Copying the fields leaves t as it was.
// A nil TLSNextProto lets the copy configure HTTP/2 on its first request.
// A non-nil map, including an empty one, is the documented way to disable
// HTTP/2, so the copy gets its own map instead of nil.
func cloneHTTPTransport(t *http.Transport) *http.Transport {
	cloned := &http.Transport{
		Proxy:                  t.Proxy,
		OnProxyConnectResponse: t.OnProxyConnectResponse,
		DialContext:            t.DialContext,
		//lint:ignore SA1019 caller transports may still set Dial; DialContext wins when both are set
		Dial: t.Dial,
		//lint:ignore SA1019 caller transports may still set DialTLS; DialTLSContext wins when both are set
		DialTLS:                t.DialTLS,
		DialTLSContext:         t.DialTLSContext,
		TLSHandshakeTimeout:    t.TLSHandshakeTimeout,
		DisableKeepAlives:      t.DisableKeepAlives,
		DisableCompression:     t.DisableCompression,
		MaxIdleConns:           t.MaxIdleConns,
		MaxIdleConnsPerHost:    t.MaxIdleConnsPerHost,
		MaxConnsPerHost:        t.MaxConnsPerHost,
		IdleConnTimeout:        t.IdleConnTimeout,
		ResponseHeaderTimeout:  t.ResponseHeaderTimeout,
		ExpectContinueTimeout:  t.ExpectContinueTimeout,
		ProxyConnectHeader:     t.ProxyConnectHeader.Clone(),
		GetProxyConnectHeader:  t.GetProxyConnectHeader,
		MaxResponseHeaderBytes: t.MaxResponseHeaderBytes,
		ForceAttemptHTTP2:      t.ForceAttemptHTTP2,
		WriteBufferSize:        t.WriteBufferSize,
		ReadBufferSize:         t.ReadBufferSize,
	}
	if t.TLSClientConfig != nil {
		cloned.TLSClientConfig = t.TLSClientConfig.Clone()
	}
	if t.TLSNextProto != nil {
		next := make(map[string]func(string, *tls.Conn) http.RoundTripper, len(t.TLSNextProto))
		for name, fn := range t.TLSNextProto {
			next[name] = fn
		}
		cloned.TLSNextProto = next
	}
	return cloned
}

// transportWithInsecureSkipVerify returns a transport that skips certificate
// verification. A nil transport gets a new default. An *http.Transport is
// cloned and the flag is set on the clone. Any other type is an error:
// replacing it would drop the caller's transport.
func transportWithInsecureSkipVerify(rt http.RoundTripper) (http.RoundTripper, error) {
	if rt == nil {
		t := defaultTransport()
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		return t, nil
	}
	t, ok := rt.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("insecure TLS cannot be applied to transport %T (WithInsecureTLS and VERGEOS_INSECURE require *http.Transport)", rt)
	}
	cloned := cloneHTTPTransport(t)
	if cloned.TLSClientConfig == nil {
		cloned.TLSClientConfig = &tls.Config{}
	}
	cloned.TLSClientConfig.InsecureSkipVerify = true
	return cloned, nil
}

// prepareHTTPClient copies the base client and applies the recorded timeout
// and TLS settings. The client stored by WithHTTPClient, and its transport,
// are not modified.
func (c *Client) prepareHTTPClient() error {
	if c.httpClient == nil {
		return fmt.Errorf("http client is nil")
	}
	cloned := *c.httpClient
	if c.timeoutSet {
		cloned.Timeout = c.timeout
	}
	if c.insecureTLS {
		tr, err := transportWithInsecureSkipVerify(cloned.Transport)
		if err != nil {
			return err
		}
		cloned.Transport = tr
	}
	c.httpClient = &cloned
	return nil
}

// WithInsecureTLS configures whether to skip TLS certificate verification.
// This is useful for self-signed certificates.
//
// The choice is recorded and applied to a copy of the base client after
// every option has run. true skips verification. false does not, and it
// replaces an earlier true from WithInsecureTLS or from the TLS environment
// variables. The TLS change is made on a clone of *http.Transport so the
// caller's transport is left alone. A transport of any other type cannot
// take that change, and NewClient returns an error instead of dropping it.
// false does not rewrite TLS settings that were already on a custom transport.
func WithInsecureTLS(insecure bool) ClientOption {
	return func(c *Client) error {
		c.insecureTLS = insecure
		return nil
	}
}

// WithTimeout sets the HTTP request timeout.
//
// The timeout is applied to a copy of the base client after every option
// has run, including when WithHTTPClient is also set. A later WithTimeout
// or VERGEOS_TIMEOUT replaces an earlier one. The *http.Client passed to
// WithHTTPClient is not modified.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		c.timeout = timeout
		c.timeoutSet = true
		return nil
	}
}

// WithPowerWait sets the default timeout and poll interval for VM power
// waits. PowerOn, PowerOff, and Kill use these values. A zero duration
// leaves that setting at its default (150 seconds, polled every 5 seconds).
// A negative duration is rejected.
//
// VMPowerOffOptions.Timeout and PollInterval override these defaults for
// one shutdown. A context deadline can still end a wait sooner. A longer
// timeout extends the wait past 150 seconds.
func WithPowerWait(timeout, pollInterval time.Duration) ClientOption {
	return func(c *Client) error {
		if timeout < 0 {
			return fmt.Errorf("power wait timeout must be >= 0, got %s", timeout)
		}
		if pollInterval < 0 {
			return fmt.Errorf("power wait poll interval must be >= 0, got %s", pollInterval)
		}
		if timeout > 0 {
			c.powerWaitTimeout = timeout
		}
		if pollInterval > 0 {
			c.powerWaitInterval = pollInterval
		}
		return nil
	}
}

// vmPowerWait returns the timeout and poll interval for a VM power wait.
func (c *Client) vmPowerWait() (time.Duration, time.Duration) {
	timeout := defaultPowerWaitTimeout
	interval := defaultPowerWaitInterval
	if c != nil && c.powerWaitTimeout > 0 {
		timeout = c.powerWaitTimeout
	}
	if c != nil && c.powerWaitInterval > 0 {
		interval = c.powerWaitInterval
	}
	return timeout, interval
}

// WithHTTPClient sets the base HTTP client.
//
// NewClient copies this client after every option has run. Timeout, TLS,
// retry, and rate limit settings are applied to the copy. CheckRedirect,
// the cookie jar, and a timeout that no other option set are kept from the
// base. The *http.Client value passed in is not modified, and neither is
// its transport.
//
// Transport stays shared when TLS settings do not need to change, so the
// connection pool stays shared. WithInsecureTLS(true), or a TLS environment
// variable that skips verification, clones an *http.Transport and sets
// InsecureSkipVerify on the clone. Any other transport type cannot take
// that change, and NewClient returns an error.
//
// Option order does not decide whether timeout, TLS, or rate limit settings
// survive WithHTTPClient. A later WithTimeout or WithInsecureTLS still
// replaces an earlier one.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("http client is nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// WithMinimumVersion sets the oldest VergeOS major NewClient will accept.
// The default is RequiredMajorVersion (26). A server older than major is
// rejected with UnsupportedVersionError. Newer majors are still accepted.
//
// WithSkipVersionCheck disables this comparison. major must be at least 1.
func WithMinimumVersion(major int) ClientOption {
	return func(c *Client) error {
		if major < 1 {
			return fmt.Errorf("minimum version must be >= 1, got %d", major)
		}
		c.minimumMajorVersion = major
		return nil
	}
}

// WithSkipVersionCheck connects even when the server major is outside the
// range NewClient would otherwise accept. The version is still read and
// stored, so feature gates that use it keep working, and the credential
// check still runs.
//
// The default check accepts VergeOS 26 and every later major. Use this
// option to keep a deployment running if a later release narrows that
// range, or to talk to a server this release does not claim to support.
func WithSkipVersionCheck() ClientOption {
	return func(c *Client) error {
		c.skipVersionCheck = true
		return nil
	}
}

// WithEnvConfig configures the client from environment variables.
// This option should typically be applied first, allowing subsequent options
// to override specific values.
//
// TLS and timeout values are recorded here and applied to a copy of the base
// HTTP client after every option has run. WithHTTPClient does not drop them.
// A later WithTimeout or WithInsecureTLS replaces the value read from the
// environment. A later WithEnvConfig replaces an earlier WithTimeout, and it
// replaces an earlier WithInsecureTLS(false) when the environment asks to
// skip verification.
//
// Environment variables:
//   - VERGEOS_HOST: Host or base URL. A value with no scheme is treated as
//     https. http and https are accepted; any other scheme is an error that
//     names VERGEOS_HOST. Required if the base URL was not set with WithBaseURL.
//   - VERGEOS_USERNAME + VERGEOS_PASSWORD: Basic authentication
//   - VERGEOS_API_KEY: Bearer token authentication. Used when set, including
//     when username and password are also set. WithAPIKey follows the same rule.
//   - VERGEOS_VERIFY_SSL: Verify TLS certificates, "true" or "false" (default: "true").
//     "false" and "0" skip verification.
//   - VERGEOS_INSECURE: Skip TLS verification when "true" (also "yes", "on", or "1").
//     VERGEOS_INSECURE=true is the same as VERGEOS_VERIFY_SSL=false. If both are
//     set and they disagree, WithEnvConfig returns an error.
//   - VERGEOS_TIMEOUT: Request timeout in seconds (default: "30")
//
// Example:
//
//	export VERGEOS_HOST=vergeos.example.com
//	export VERGEOS_API_KEY=secret
//	export VERGEOS_INSECURE=true
//
//	// Simple usage
//	client, err := vergeos.NewClient(vergeos.WithEnvConfig())
//
//	// With explicit override
//	client, err := vergeos.NewClient(
//	    vergeos.WithEnvConfig(),
//	    vergeos.WithTimeout(60*time.Second),
//	)
func WithEnvConfig() ClientOption {
	return func(c *Client) error {
		// Base URL (only if not already set)
		if c.baseURL == "" {
			if host := os.Getenv("VERGEOS_HOST"); host != "" {
				normalized, err := normalizeEnvHost(host)
				if err != nil {
					return err
				}
				c.baseURL = normalized
			}
		}

		// Authentication (only if not already set).
		// An API key wins when both it and username/password are set.
		hasAuth := (c.username != "" && c.password != "") || c.apiKey != ""
		if !hasAuth {
			username := os.Getenv("VERGEOS_USERNAME")
			password := os.Getenv("VERGEOS_PASSWORD")
			apiKey := os.Getenv("VERGEOS_API_KEY")

			if apiKey != "" {
				c.apiKey = apiKey
			} else if username != "" && password != "" {
				c.username = username
				c.password = password
			}
		}

		skipVerify, err := envSkipTLSVerify()
		if err != nil {
			return err
		}
		if skipVerify {
			c.insecureTLS = true
		}

		// Timeout (only if set in environment). A later WithTimeout replaces it.
		if timeoutStr := os.Getenv("VERGEOS_TIMEOUT"); timeoutStr != "" {
			timeout, err := strconv.Atoi(timeoutStr)
			if err != nil {
				return fmt.Errorf("invalid VERGEOS_TIMEOUT value %q: %w", timeoutStr, err)
			}
			c.timeout = time.Duration(timeout) * time.Second
			c.timeoutSet = true
		}

		return nil
	}
}

// normalizeEnvHost returns an http or https base URL for VERGEOS_HOST.
// A value with no scheme is treated as https.
func normalizeEnvHost(raw string) (string, error) {
	host := strings.TrimSpace(raw)
	if host == "" {
		return "", fmt.Errorf("invalid VERGEOS_HOST %q: missing host", raw)
	}
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	u, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("invalid VERGEOS_HOST %q: %w", raw, err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("invalid VERGEOS_HOST %q: unsupported protocol scheme %q (use http or https)", raw, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid VERGEOS_HOST %q: missing host", raw)
	}
	u.Scheme = scheme
	return strings.TrimSuffix(u.String(), "/"), nil
}

// envSkipTLSVerify reports whether the TLS environment variables ask the
// client to skip certificate verification.
//
// VERGEOS_INSECURE=true is an alias for VERGEOS_VERIFY_SSL=false. When both
// variables are set they must agree. An empty value is treated as unset.
// VERGEOS_VERIFY_SSL disables verification only for "false" and "0".
// VERGEOS_INSECURE accepts the same boolean words Ansible does: true/false,
// yes/no, on/off, y/n, t/f, and 1/0.
func envSkipTLSVerify() (bool, error) {
	verifyRaw := strings.TrimSpace(os.Getenv("VERGEOS_VERIFY_SSL"))
	insecureRaw := strings.TrimSpace(os.Getenv("VERGEOS_INSECURE"))

	verifySet := verifyRaw != ""
	insecureSet := insecureRaw != ""

	verifySkips := false
	if verifySet {
		verifySkips = strings.EqualFold(verifyRaw, "false") || verifyRaw == "0"
	}

	insecureSkips := false
	if insecureSet {
		switch strings.ToLower(insecureRaw) {
		case "1", "t", "true", "y", "yes", "on":
			insecureSkips = true
		case "0", "f", "false", "n", "no", "off":
			insecureSkips = false
		default:
			return false, fmt.Errorf("invalid VERGEOS_INSECURE value %q: expected true or false", insecureRaw)
		}
	}

	if verifySet && insecureSet && verifySkips != insecureSkips {
		return false, fmt.Errorf("VERGEOS_VERIFY_SSL=%q and VERGEOS_INSECURE=%q contradict each other (VERGEOS_INSECURE=true is the same as VERGEOS_VERIFY_SSL=false)", verifyRaw, insecureRaw)
	}
	if insecureSet {
		return insecureSkips, nil
	}
	return verifySkips, nil
}

// NewClient creates a new VergeOS API client.
//
// Client creation checks that the server is VergeOS 26 or later and that
// the supplied credentials are accepted. /version.json does not require
// authentication, so a wrong password is caught by one follow-up read of
// a small table. A 401 from that read is not retried: VergeOS locks an
// account after a small number of failed logins, and retrying a bad
// password locks the account for every client that shares it, including
// API keys. A dropped connection on that GET is retried like any other
// idempotent request.
//
// A failed login is returned as an AuthError. A credential that is
// accepted but not allowed to read the check endpoint is a PermissionError.
// A server older than the minimum major is returned as an
// UnsupportedVersionError. WithMinimumVersion changes that floor.
// WithSkipVersionCheck records the version and does not reject it.
func NewClient(opts ...ClientOption) (*Client, error) {
	// Create client with defaults. Options record HTTP settings. The client
	// those settings apply to is built once, after every option has run.
	c := &Client{
		httpClient: defaultHTTPClient(),
		userAgent:  defaultUserAgent,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("vergeos: failed to apply client option: %w", err)
		}
	}

	// Validate required fields
	if c.baseURL == "" {
		return nil, fmt.Errorf("vergeos: base URL is required (use WithBaseURL)")
	}
	hasCredentials := c.username != "" && c.password != ""
	hasAPIKey := c.apiKey != ""
	if !hasCredentials && !hasAPIKey {
		return nil, fmt.Errorf("vergeos: authentication required (use WithCredentials or WithAPIKey)")
	}

	// Build the HTTP client from the recorded options, then wrap it so
	// startup checks use the same retry and rate limit policy as later calls.
	if err := c.applyTransportPolicy(); err != nil {
		return nil, fmt.Errorf("vergeos: %w", err)
	}

	// Initialize services
	c.VMs = &VMService{client: c}
	c.VMSnapshots = &VMSnapshotService{client: c}
	c.VMNICs = &VMNICService{client: c}
	c.VMDrives = &VMDriveService{client: c}
	c.VMDevices = &VMDeviceService{client: c}
	c.VMImports = &VMImportService{client: c}
	c.VMImportLogs = &VMImportLogService{client: c}
	c.VMExports = &VMExportService{client: c}
	c.Catalogs = &CatalogService{client: c}
	c.VMRecipes = &VMRecipeService{client: c}
	c.VMRecipeInstances = &VMRecipeInstanceService{client: c}
	c.Networks = &NetworkService{client: c}
	c.Users = &UserService{client: c}
	c.Members = &MemberService{client: c}
	c.CloudInitFiles = &CloudInitService{client: c}
	c.Clusters = &ClusterService{client: c}
	c.Nodes = &NodeService{client: c}
	c.VGPUProfiles = &VGPUProfileService{client: c}
	c.NodeGPUs = &NodeGPUService{client: c}
	c.NodeGPUStats = &NodeGPUStatsService{client: c}
	c.NodeGPUInstances = &NodeGPUInstanceService{client: c}
	c.NodeVGPUDevices = &NodeVGPUDeviceService{client: c}
	c.NodeHostGPUDevices = &NodeHostGPUDeviceService{client: c}
	c.NodeVGPUProfiles = &NodeVGPUProfileService{client: c}
	c.NodeMemory = &NodeMemoryService{client: c}
	c.NodeLLDPNeighbors = &NodeLLDPNeighborService{client: c}
	c.Groups = &GroupService{client: c}
	c.Files = &FileService{client: c}
	c.ResourceGroups = &ResourceGroupService{client: c}
	c.Settings = &SettingsService{client: c}
	c.System = &SystemService{client: c}
	c.Schema = &SchemaService{client: c}
	c.Tags = &TagService{client: c}
	c.TagCategories = &TagCategoryService{client: c}
	c.TagMembers = &TagMemberService{client: c}
	c.Volumes = &VolumeService{client: c}
	c.VNetRules = &VNetRuleService{client: c}
	c.VNetRuleAliases = &VNetRuleAliasService{client: c}
	c.Tenants = &TenantService{client: c}
	c.TenantNodes = &TenantNodeService{client: c}
	c.TenantStorage = &TenantStorageService{client: c}
	c.TenantStatus = &TenantStatusService{client: c}
	c.TenantStatsHistoryShort = &TenantStatsHistoryShortService{client: c}
	c.TenantSnapshots = &TenantSnapshotService{client: c}
	c.TenantLayer2Networks = &TenantLayer2NetworkService{client: c}
	c.TenantNetworkBlocks = &TenantNetworkBlockService{client: c}
	c.TenantExternalIPs = &TenantExternalIPService{client: c}
	c.SnapshotProfiles = &SnapshotProfileService{client: c}
	c.SnapshotProfilePeriods = &SnapshotProfilePeriodService{client: c}
	c.Alarms = &AlarmService{client: c}
	c.AlarmTypes = &AlarmTypeService{client: c}
	c.Tasks = &TaskService{client: c}
	c.TaskSchedules = &TaskScheduleService{client: c}
	c.TaskScheduleTriggers = &TaskScheduleTriggerService{client: c}
	c.TaskEvents = &TaskEventService{client: c}
	c.TaskScripts = &TaskScriptService{client: c}
	c.VNetAddresses = &VNetAddressService{client: c}
	c.VNetProxies = &VNetProxyService{client: c}
	c.VNetProxyTenants = &VNetProxyTenantService{client: c}
	c.VNetDNSViews = &VNetDNSViewService{client: c}
	c.VNetDNSZones = &VNetDNSZoneService{client: c}
	c.VNetDNSRecords = &VNetDNSRecordService{client: c}
	c.VNetHosts = &VNetHostService{client: c}
	c.VNetBGP = &VNetBGPService{client: c}
	c.VNetBGPRouters = &VNetBGPRouterService{client: c}
	c.VNetBGPRouterCommands = &VNetBGPRouterCommandService{client: c}
	c.VNetBGPInterfaces = &VNetBGPInterfaceService{client: c}
	c.VNetBGPInterfaceCommands = &VNetBGPInterfaceCommandService{client: c}
	c.VNetBGPRouteMaps = &VNetBGPRouteMapService{client: c}
	c.VNetBGPRouteMapCommands = &VNetBGPRouteMapCommandService{client: c}
	c.VNetBGPIPCommands = &VNetBGPIPCommandService{client: c}
	c.VNetOSPFCommands = &VNetOSPFCommandService{client: c}
	c.VNetEIGRPRouters = &VNetEIGRPRouterService{client: c}
	c.VNetEIGRPRouterCommands = &VNetEIGRPRouterCommandService{client: c}
	c.VNetWireGuards = &VNetWireGuardService{client: c}
	c.VNetWireGuardPeers = &VNetWireGuardPeerService{client: c}
	c.VNetWireGuardPeerStatus = &VNetWireGuardPeerStatusService{client: c}
	c.Certificates = &CertificateService{client: c}
	c.VNetIPSecs = &VNetIPSecService{client: c}
	c.VNetIPSecPhase1s = &VNetIPSecPhase1Service{client: c}
	c.VNetIPSecPhase2s = &VNetIPSecPhase2Service{client: c}
	c.VNetIPSecConnections = &VNetIPSecConnectionService{client: c}
	c.Sites = &SiteService{client: c}
	c.SiteSyncsIncoming = &SiteSyncIncomingService{client: c}
	c.SiteSyncsOutgoing = &SiteSyncOutgoingService{client: c}
	c.SiteSyncProfilePeriods = &SiteSyncProfilePeriodService{client: c}
	c.CloudSnapshots = &CloudSnapshotService{client: c}
	c.CloudSnapshotVMs = &CloudSnapshotVMService{client: c}
	c.CloudSnapshotTenants = &CloudSnapshotTenantService{client: c}
	c.VolumeCIFSShares = &VolumeCIFSShareService{client: c}
	c.VolumeNFSShares = &VolumeNFSShareService{client: c}
	c.VolumeBrowser = &VolumeBrowserService{client: c}
	c.WebhookURLs = &WebhookURLService{client: c}
	c.Webhooks = &WebhookService{client: c}
	c.UserAPIKeys = &UserAPIKeyService{client: c}
	c.AuthSources = &AuthSourceService{client: c}
	c.OIDCApplications = &OIDCApplicationService{client: c}
	c.NASServices = &NASServiceService{client: c}
	c.NASServiceUsers = &NASServiceUserService{client: c}
	c.VolumeSyncs = &VolumeSyncService{client: c}
	c.VolumeSnapshots = &VolumeSnapshotService{client: c}
	c.Permissions = &PermissionService{client: c}
	c.Logs = &LogService{client: c}
	c.StorageTiers = &StorageTierService{client: c}
	c.ClusterTiers = &ClusterTierService{client: c}
	c.MachineDrivePhys = &MachineDrivePhysService{client: c}
	c.ClusterStatsHistory = &ClusterStatsHistoryService{client: c}
	c.MachineStatus = &MachineStatusService{client: c}
	c.MachineStats = &MachineStatsService{client: c}
	c.MachineDriveStats = &MachineDriveStatsService{client: c}
	c.MachineNICs = &MachineNICService{client: c}
	c.UpdateSettings = &UpdateSettingsService{client: c}
	c.UpdateBranches = &UpdateBranchService{client: c}
	c.UpdateSourcePackages = &UpdateSourcePackageService{client: c}
	c.Billing = &BillingService{client: c}
	c.NASServiceAntivirus = &NASServiceAntivirusService{client: c}
	c.SharedObjects = &SharedObjectService{client: c}

	// Validate server version before returning client.
	// /version.json is public, so this does not authenticate.
	if err := c.checkServerVersion(context.Background()); err != nil {
		return nil, err
	}

	// One authenticated read. A 401 is returned as-is. Repeating a
	// rejected password counts toward account lockout.
	if err := c.checkCredentials(context.Background()); err != nil {
		return nil, err
	}

	return c, nil
}

// checkCredentials performs one authenticated request.
//
// A 401 is not retried. The platform locks an account after a
// configurable number of failed logins (five on the system this was
// measured against). Repeating a rejected password locks that account
// for every client sharing it, including its API keys. A connection
// reset before a response is retried, because the server did not
// reject the credentials.
func (c *Client) checkCredentials(ctx context.Context) error {
	params := url.Values{}
	params.Set("limit", "1")
	params.Set("fields", "$key")
	return c.get(ctx, credentialCheckEndpoint, params, nil)
}

// apiResponse represents the standard VergeOS API response structure.
type apiResponse struct {
	Key      any    `json:"$key,omitempty"`
	Response any    `json:"response,omitempty"`
	Err      string `json:"err,omitempty"`
}

// request performs an HTTP request to the VergeOS API.
func (c *Client) request(ctx context.Context, method, endpoint string, body any, params url.Values) (*http.Response, error) {
	// Build URL
	u := fmt.Sprintf("%s%s%s", c.baseURL, apiBasePath, endpoint)
	if len(params) > 0 {
		u = fmt.Sprintf("%s?%s", u, params.Encode())
	}

	// Build request body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("vergeos: failed to marshal request body: %w", err)
		}
		// *bytes.Reader makes http.NewRequest record GetBody, which is
		// what a retry uses to send this payload again.
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("vergeos: failed to create request: %w", err)
	}

	// An API key takes precedence over username and password.
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-JSON-Non-Compact", "1")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vergeos: request failed: %w", err)
	}

	return resp, nil
}

// do performs an HTTP request and decodes the response.
func (c *Client) do(ctx context.Context, method, endpoint string, body any, params url.Values, result any) error {
	resp, err := c.request(ctx, method, endpoint, body, params)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body (bounded to prevent OOM from misbehaving servers)
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("vergeos: failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to parse error message from response
		var apiResp apiResponse
		errMsg := string(respBody)
		if json.Unmarshal(respBody, &apiResp) == nil && apiResp.Err != "" {
			errMsg = apiResp.Err
		}

		return apiStatusError(resp.StatusCode, endpoint, errMsg)
	}

	// Decode response if result is provided
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("vergeos: failed to decode response: %w", err)
		}
	}

	return nil
}

// get performs a GET request.
func (c *Client) get(ctx context.Context, endpoint string, params url.Values, result any) error {
	return c.do(ctx, http.MethodGet, endpoint, nil, params, result)
}

// post performs a POST request and returns the created resource's key.
func (c *Client) post(ctx context.Context, endpoint string, body any, result any) error {
	return c.do(ctx, http.MethodPost, endpoint, body, nil, result)
}

// postStatus performs a POST and returns the HTTP status and body.
//
// Unlike post, a non-2xx status is not turned into an error. Recipe preview
// completes as HTTP 405 with the report in the body. Callers decide what a
// status means. POST is not retried.
func (c *Client) postStatus(ctx context.Context, endpoint string, body any) (int, []byte, error) {
	resp, err := c.request(ctx, http.MethodPost, endpoint, body, nil)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("vergeos: failed to read response body: %w", err)
	}
	return resp.StatusCode, respBody, nil
}

// put performs a PUT request.
func (c *Client) put(ctx context.Context, endpoint string, body any, result any) error {
	return c.do(ctx, http.MethodPut, endpoint, body, nil, result)
}

// putRaw performs a PUT and returns the response body unchanged.
// An empty body is a nil slice.
func (c *Client) putRaw(ctx context.Context, endpoint string, body any) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.put(ctx, endpoint, body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, endpoint string) error {
	return c.do(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

// getAbsolute performs a GET request to an absolute path (not under /api/v4/).
// This is used for endpoints like /version.json that are outside the API path
// and do not require authentication. Credentials are omitted so a bad password
// is presented only by checkCredentials, and only once.
func (c *Client) getAbsolute(ctx context.Context, path string, params url.Values, result any) error {
	// Build URL with absolute path
	u := c.baseURL + path
	if len(params) > 0 {
		u = fmt.Sprintf("%s?%s", u, params.Encode())
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("vergeos: failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vergeos: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		return apiStatusError(resp.StatusCode, path, string(body))
	}

	// Decode response
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("vergeos: failed to decode response: %w", err)
		}
	}

	return nil
}

// apiStatusError maps a non-success HTTP status to an SDK error.
// 401 is an authentication failure, 403 is a permission denial, and 409
// is a conflict. Other statuses, including 404, are a plain APIError.
// IsNotFoundError already checks the status code.
func apiStatusError(status int, endpoint, message string) error {
	switch status {
	case http.StatusUnauthorized:
		return &AuthError{Message: message}
	case http.StatusForbidden:
		return &PermissionError{APIError: APIError{
			StatusCode: status,
			Endpoint:   endpoint,
			Message:    message,
		}}
	case http.StatusConflict:
		return &ConflictError{APIError: APIError{
			StatusCode: status,
			Endpoint:   endpoint,
			Message:    message,
		}}
	default:
		return &APIError{
			StatusCode: status,
			Endpoint:   endpoint,
			Message:    message,
		}
	}
}

// getKey extracts the key from an API response.
func getKey(resp apiResponse) (int, error) {
	if resp.Key == nil {
		return 0, fmt.Errorf("vergeos: response missing $key field")
	}

	switch v := resp.Key.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		var id int
		if _, err := fmt.Sscanf(v, "%d", &id); err != nil {
			return 0, fmt.Errorf("vergeos: invalid $key value: %v", v)
		}
		return id, nil
	default:
		return 0, fmt.Errorf("vergeos: unexpected $key type: %T", v)
	}
}
