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
	"strings"
	"time"
)

const (
	// defaultTimeout is the default HTTP request timeout.
	defaultTimeout = 30 * time.Second
	// apiBasePath is the base path for the VergeOS API.
	apiBasePath = "/api/v4"
	// defaultUserAgent is the default User-Agent header.
	defaultUserAgent = "vergeos-go-sdk/1.0"
)

// Client is the VergeOS API client.
type Client struct {
	// baseURL is the base URL for the VergeOS API.
	baseURL string
	// username is the API username.
	username string
	// password is the API password.
	password string
	// httpClient is the HTTP client used for requests.
	httpClient *http.Client
	// userAgent is the User-Agent header sent with requests.
	userAgent string

	// Services for interacting with different API resources.
	// All services implement their corresponding interfaces for mock testing.
	VMs                     VMServiceInterface
	VMNICs                  VMNICServiceInterface
	VMDrives                VMDriveServiceInterface
	VMDevices               VMDeviceServiceInterface
	Networks                NetworkServiceInterface
	Users                   UserServiceInterface
	Members                 MemberServiceInterface
	CloudInitFiles          CloudInitServiceInterface
	Clusters                ClusterServiceInterface
	Nodes                   NodeServiceInterface
	Groups                  GroupServiceInterface
	MediaSources            MediaSourceServiceInterface
	ResourceGroups          ResourceGroupServiceInterface
	Settings                SettingsServiceInterface
	System                  SystemServiceInterface
	Schema                  SchemaServiceInterface
	Tags                    TagServiceInterface
	TagMembers              TagMemberServiceInterface
	Volumes                 VolumeServiceInterface
	VNetRules               VNetRuleServiceInterface
	VNetRuleAliases         VNetRuleAliasServiceInterface
	Tenants                 TenantServiceInterface
	TenantNodes             TenantNodeServiceInterface
	TenantStorage           TenantStorageServiceInterface
	SnapshotProfiles        SnapshotProfileServiceInterface
	SnapshotProfilePeriods  SnapshotProfilePeriodServiceInterface
	Alarms                  AlarmServiceInterface
	AlarmTypes              AlarmTypeServiceInterface
	Tasks                   TaskServiceInterface
	VNetAddresses           VNetAddressServiceInterface
	VNetDNSViews            VNetDNSViewServiceInterface
	VNetDNSZones            VNetDNSZoneServiceInterface
	VNetDNSRecords          VNetDNSRecordServiceInterface
	VNetHosts               VNetHostServiceInterface
	VNetWireGuards          VNetWireGuardServiceInterface
	VNetWireGuardPeers      VNetWireGuardPeerServiceInterface
	VNetWireGuardPeerStatus VNetWireGuardPeerStatusServiceInterface
	Certificates            CertificateServiceInterface
	VNetIPSecs              VNetIPSecServiceInterface
	VNetIPSecPhase1s        VNetIPSecPhase1ServiceInterface
	VNetIPSecPhase2s        VNetIPSecPhase2ServiceInterface
	VNetIPSecConnections    VNetIPSecConnectionServiceInterface
	Sites                   SiteServiceInterface
	SiteSyncsIncoming       SiteSyncIncomingServiceInterface
	SiteSyncsOutgoing       SiteSyncOutgoingServiceInterface
	CloudSnapshots          CloudSnapshotServiceInterface
	CloudSnapshotVMs        CloudSnapshotVMServiceInterface
	CloudSnapshotTenants    CloudSnapshotTenantServiceInterface
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
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) error {
		c.username = username
		c.password = password
		return nil
	}
}

// WithInsecureTLS configures whether to skip TLS certificate verification.
// This is useful for self-signed certificates.
func WithInsecureTLS(insecure bool) ClientOption {
	return func(c *Client) error {
		if insecure {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			}
			c.httpClient.Transport = transport
		}
		return nil
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
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

// NewClient creates a new VergeOS API client.
func NewClient(opts ...ClientOption) (*Client, error) {
	// Create client with defaults
	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		userAgent: defaultUserAgent,
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
	if c.username == "" || c.password == "" {
		return nil, fmt.Errorf("vergeos: credentials are required (use WithCredentials)")
	}

	// Initialize services
	c.VMs = &VMService{client: c}
	c.VMNICs = &VMNICService{client: c}
	c.VMDrives = &VMDriveService{client: c}
	c.VMDevices = &VMDeviceService{client: c}
	c.Networks = &NetworkService{client: c}
	c.Users = &UserService{client: c}
	c.Members = &MemberService{client: c}
	c.CloudInitFiles = &CloudInitService{client: c}
	c.Clusters = &ClusterService{client: c}
	c.Nodes = &NodeService{client: c}
	c.Groups = &GroupService{client: c}
	c.MediaSources = &MediaSourceService{client: c}
	c.ResourceGroups = &ResourceGroupService{client: c}
	c.Settings = &SettingsService{client: c}
	c.System = &SystemService{client: c}
	c.Schema = &SchemaService{client: c}
	c.Tags = &TagService{client: c}
	c.TagMembers = &TagMemberService{client: c}
	c.Volumes = &VolumeService{client: c}
	c.VNetRules = &VNetRuleService{client: c}
	c.VNetRuleAliases = &VNetRuleAliasService{client: c}
	c.Tenants = &TenantService{client: c}
	c.TenantNodes = &TenantNodeService{client: c}
	c.TenantStorage = &TenantStorageService{client: c}
	c.SnapshotProfiles = &SnapshotProfileService{client: c}
	c.SnapshotProfilePeriods = &SnapshotProfilePeriodService{client: c}
	c.Alarms = &AlarmService{client: c}
	c.AlarmTypes = &AlarmTypeService{client: c}
	c.Tasks = &TaskService{client: c}
	c.VNetAddresses = &VNetAddressService{client: c}
	c.VNetDNSViews = &VNetDNSViewService{client: c}
	c.VNetDNSZones = &VNetDNSZoneService{client: c}
	c.VNetDNSRecords = &VNetDNSRecordService{client: c}
	c.VNetHosts = &VNetHostService{client: c}
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
	c.CloudSnapshots = &CloudSnapshotService{client: c}
	c.CloudSnapshotVMs = &CloudSnapshotVMService{client: c}
	c.CloudSnapshotTenants = &CloudSnapshotTenantService{client: c}

	return c, nil
}

// apiResponse represents the standard VergeOS API response structure.
type apiResponse struct {
	Key      interface{} `json:"$key,omitempty"`
	Response interface{} `json:"response,omitempty"`
	Err      string      `json:"err,omitempty"`
}

// request performs an HTTP request to the VergeOS API.
func (c *Client) request(ctx context.Context, method, endpoint string, body interface{}, params url.Values) (*http.Response, error) {
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
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("vergeos: failed to create request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(c.username, c.password)
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
func (c *Client) do(ctx context.Context, method, endpoint string, body interface{}, params url.Values, result interface{}) error {
	resp, err := c.request(ctx, method, endpoint, body, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("vergeos: failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		// Try to parse error message from response
		var apiResp apiResponse
		errMsg := string(respBody)
		if json.Unmarshal(respBody, &apiResp) == nil && apiResp.Err != "" {
			errMsg = apiResp.Err
		}

		// Return appropriate error type
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return &AuthError{Message: errMsg}
		}
		if resp.StatusCode == 404 {
			return &APIError{
				StatusCode: resp.StatusCode,
				Endpoint:   endpoint,
				Message:    errMsg,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   endpoint,
			Message:    errMsg,
		}
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
func (c *Client) get(ctx context.Context, endpoint string, params url.Values, result interface{}) error {
	return c.do(ctx, http.MethodGet, endpoint, nil, params, result)
}

// post performs a POST request and returns the created resource's key.
func (c *Client) post(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	return c.do(ctx, http.MethodPost, endpoint, body, nil, result)
}

// put performs a PUT request.
func (c *Client) put(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	return c.do(ctx, http.MethodPut, endpoint, body, nil, result)
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, endpoint string) error {
	return c.do(ctx, http.MethodDelete, endpoint, nil, nil, nil)
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
