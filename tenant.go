package vergeos

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// TenantService handles tenant operations.
type TenantService struct {
	client *Client
}

// List returns all tenants, with optional filtering and pagination.
func (s *TenantService) List(ctx context.Context, opts ...ListOption) ([]Tenant, error) {
	options := applyListOptions(opts)

	// Use tenant-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = tenantListFields
	}

	params := options.toQueryParams()

	var tenants []Tenant
	if err := s.client.get(ctx, "/tenants", params, &tenants); err != nil {
		return nil, err
	}

	return tenants, nil
}

// Get returns a single tenant by ID.
func (s *TenantService) Get(ctx context.Context, id int) (*Tenant, error) {
	params := url.Values{}
	params.Set("fields", tenantGetFields)

	var tenant Tenant
	endpoint := fmt.Sprintf("/tenants/%d", id)
	if err := s.client.get(ctx, endpoint, params, &tenant); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Tenant", ID: id}
		}
		return nil, err
	}

	return &tenant, nil
}

// GetByName returns a tenant by name.
func (s *TenantService) GetByName(ctx context.Context, name string) (*Tenant, error) {
	tenants, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(tenants) == 0 {
		return nil, &NotFoundError{Resource: "Tenant", ID: name}
	}
	if err := requireUniqueName("Tenant", name, tenants, func(t Tenant) any { return t.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("Tenant", name, tenants[0].Name, name); err != nil {
		return nil, err
	}
	// Get full details
	got, err := s.Get(ctx, int(tenants[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("Tenant", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a new tenant and returns the created tenant.
func (s *TenantService) Create(ctx context.Context, req *TenantCreateRequest) (*Tenant, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/tenants", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created tenant's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created tenant
	return s.Get(ctx, id)
}

// Update updates a tenant and returns the updated tenant.
func (s *TenantService) Update(ctx context.Context, id int, req *TenantUpdateRequest) (*Tenant, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/tenants/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Tenant", ID: id}
		}
		return nil, err
	}

	// Read back the updated tenant
	return s.Get(ctx, id)
}

// Delete deletes a tenant.
func (s *TenantService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/tenants/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "Tenant", ID: id}
		}
		return err
	}
	return nil
}

// PowerOn powers on a tenant.
func (s *TenantService) PowerOn(ctx context.Context, id int) error {
	return s.PowerOnWithNode(ctx, id, 0)
}

// PowerOnWithNode powers on a tenant on a specific preferred node.
// If preferredNode is 0, the system chooses the node.
func (s *TenantService) PowerOnWithNode(ctx context.Context, id int, preferredNode int) error {
	action := tenantAction{
		Tenant: id,
		Action: "poweron",
	}
	if preferredNode > 0 {
		action.Params = map[string]any{
			"preferred_node": preferredNode,
		}
	}

	if err := s.client.post(ctx, "/tenant_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power on tenant %d: %w", id, err)
	}
	return nil
}

// PowerOff powers off a tenant.
func (s *TenantService) PowerOff(ctx context.Context, id int) error {
	action := tenantAction{
		Tenant: id,
		Action: "poweroff",
	}

	if err := s.client.post(ctx, "/tenant_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power off tenant %d: %w", id, err)
	}
	return nil
}

// Reset resets a tenant (restart).
func (s *TenantService) Reset(ctx context.Context, id int) error {
	action := tenantAction{
		Tenant: id,
		Action: "reset",
	}

	if err := s.client.post(ctx, "/tenant_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to reset tenant %d: %w", id, err)
	}
	return nil
}

// Clone clones a tenant.
func (s *TenantService) Clone(ctx context.Context, id int, opts *TenantCloneOptions) error {
	if opts == nil {
		return &ValidationError{Message: "clone options are required"}
	}
	if opts.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required for clone"}
	}

	action := tenantAction{
		Tenant: id,
		Action: "clone",
		Params: map[string]any{
			"name":       opts.Name,
			"no_vnet":    opts.NoVNet,
			"no_storage": opts.NoStorage,
			"no_nodes":   opts.NoNodes,
		},
	}

	if err := s.client.post(ctx, "/tenant_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to clone tenant %d: %w", id, err)
	}
	return nil
}

// IsolateOn enables network isolation for a tenant.
func (s *TenantService) IsolateOn(ctx context.Context, id int) error {
	action := tenantAction{
		Tenant: id,
		Action: "isolateon",
	}

	if err := s.client.post(ctx, "/tenant_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to isolate tenant %d: %w", id, err)
	}
	return nil
}

// IsolateOff disables network isolation for a tenant.
func (s *TenantService) IsolateOff(ctx context.Context, id int) error {
	action := tenantAction{
		Tenant: id,
		Action: "isolateoff",
	}

	if err := s.client.post(ctx, "/tenant_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to un-isolate tenant %d: %w", id, err)
	}
	return nil
}

// Connect returns a client for a running tenant's own UI.
//
// The tenant's ui_address row is read through VNetAddresses. The new client
// uses that IP and the parent client's URL scheme. It keeps the parent's
// TLS, timeout, retry, rate limit, version, user agent, and power-wait
// settings. It does not keep the parent's username, password, or API key.
//
// opts supply the tenant credentials and may replace any inherited setting.
// WithCredentials or WithAPIKey is required. WithBaseURL replaces the
// address discovered here.
//
// A snapshot, a tenant that is not running, and a tenant with no UI address
// are refused before the tenant client is created.
func (s *TenantService) Connect(ctx context.Context, id int, opts ...ClientOption) (*Client, error) {
	if err := tenantClientAuth(opts); err != nil {
		return nil, err
	}

	tenant, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	baseURL, err := s.tenantUIBaseURL(ctx, tenant)
	if err != nil {
		return nil, err
	}

	inherited := s.client.tenantClientOptions(baseURL)
	return NewClient(append(inherited, opts...)...)
}

// ConnectByName returns a client for a running tenant's own UI, looked up by name.
// See Connect for the address and the settings the new client keeps.
func (s *TenantService) ConnectByName(ctx context.Context, name string, opts ...ClientOption) (*Client, error) {
	if err := tenantClientAuth(opts); err != nil {
		return nil, err
	}
	tenant, err := s.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	baseURL, err := s.tenantUIBaseURL(ctx, tenant)
	if err != nil {
		return nil, err
	}
	inherited := s.client.tenantClientOptions(baseURL)
	return NewClient(append(inherited, opts...)...)
}

// tenantUIBaseURL resolves the tenant UI from its ui_address row.
func (s *TenantService) tenantUIBaseURL(ctx context.Context, tenant *Tenant) (string, error) {
	if tenant == nil {
		return "", &ValidationError{Message: "tenant is required"}
	}
	if tenant.IsSnapshot {
		return "", &ValidationError{Message: fmt.Sprintf("cannot connect to tenant snapshot %q", tenant.Name)}
	}

	status, err := s.client.TenantStatus.Get(ctx, int(tenant.Key))
	if err != nil {
		return "", err
	}
	if !status.Running {
		return "", &ValidationError{Message: fmt.Sprintf("tenant %q is not running", tenant.Name)}
	}
	if tenant.UIAddress == 0 {
		return "", &ValidationError{Field: "ui_address", Message: fmt.Sprintf("tenant %q has no UI address configured", tenant.Name)}
	}

	address, err := s.client.VNetAddresses.Get(ctx, int(tenant.UIAddress))
	if err != nil {
		return "", err
	}
	return tenantUIBaseURL(s.client.baseURL, address.IP)
}

// tenantUIBaseURL builds the tenant client base URL from the parent scheme and the UI IP.
func tenantUIBaseURL(parentBaseURL, ip string) (string, error) {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return "", &ValidationError{Field: "ui_address", Message: "tenant UI address is not a valid IP"}
	}
	scheme := "https"
	if u, err := url.Parse(parentBaseURL); err == nil {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			scheme = strings.ToLower(u.Scheme)
		}
	}
	host := parsed.String()
	if parsed.To4() == nil {
		host = "[" + host + "]"
	}
	return scheme + "://" + host, nil
}

// tenantClientAuth reports whether opts carry tenant credentials.
// The check runs before any request so a missing password does not depend
// on the tenant being up.
func tenantClientAuth(opts []ClientOption) error {
	scratch := &Client{}
	for _, opt := range opts {
		if err := opt(scratch); err != nil {
			return fmt.Errorf("vergeos: failed to apply client option: %w", err)
		}
	}
	hasCredentials := scratch.username != "" && scratch.password != ""
	if !hasCredentials && scratch.apiKey == "" {
		return fmt.Errorf("vergeos: tenant client authentication required (use WithCredentials or WithAPIKey)")
	}
	return nil
}

// tenantClientOptions copies the parent settings a tenant UI client keeps.
// Credentials are not copied. The tenant has its own.
func (c *Client) tenantClientOptions(baseURL string) []ClientOption {
	if c == nil {
		return []ClientOption{WithBaseURL(baseURL)}
	}
	opts := []ClientOption{WithBaseURL(baseURL)}
	if c.userAgent != "" {
		opts = append(opts, WithUserAgent(c.userAgent))
	}
	if c.insecureTLS {
		opts = append(opts, WithInsecureTLS(true))
	}
	if c.timeoutSet {
		opts = append(opts, WithTimeout(c.timeout))
	}
	if c.retryConfigured {
		opts = append(opts, WithRetry(c.retryPolicy))
	}
	if c.rateLimit > 0 {
		opts = append(opts, WithRateLimit(c.rateLimit))
	}
	if c.skipVersionCheck {
		opts = append(opts, WithSkipVersionCheck())
	}
	if c.minimumMajorVersion > 0 {
		opts = append(opts, WithMinimumVersion(c.minimumMajorVersion))
	}
	if c.powerWaitTimeout > 0 || c.powerWaitInterval > 0 {
		opts = append(opts, WithPowerWait(c.powerWaitTimeout, c.powerWaitInterval))
	}
	return opts
}

// tenantAction represents a tenant action request.
type tenantAction struct {
	Tenant int            `json:"tenant"`
	Action string         `json:"action"`
	Params map[string]any `json:"params,omitempty"`
}

// TenantNodeService handles tenant node operations.
type TenantNodeService struct {
	client *Client
}

// List returns all tenant nodes, with optional filtering and pagination.
func (s *TenantNodeService) List(ctx context.Context, opts ...ListOption) ([]TenantNode, error) {
	options := applyListOptions(opts)

	// Use node-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = tenantNodeListFields
	}

	params := options.toQueryParams()

	var nodes []TenantNode
	if err := s.client.get(ctx, "/tenant_nodes", params, &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

// ListByTenant returns all nodes for a specific tenant.
func (s *TenantNodeService) ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantNode, error) {
	// Prepend tenant filter to any existing filters
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("tenant eq %d", tenantID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns a single tenant node by ID.
func (s *TenantNodeService) Get(ctx context.Context, id int) (*TenantNode, error) {
	params := url.Values{}
	params.Set("fields", tenantNodeGetFields)

	var node TenantNode
	endpoint := fmt.Sprintf("/tenant_nodes/%d", id)
	if err := s.client.get(ctx, endpoint, params, &node); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantNode", ID: id}
		}
		return nil, err
	}

	return &node, nil
}

// GetByName returns a tenant node by name within a specific tenant.
func (s *TenantNodeService) GetByName(ctx context.Context, tenantID int, name string) (*TenantNode, error) {
	nodes, err := s.List(ctx, WithFilter(fmt.Sprintf("tenant eq %d and name eq '%s'", tenantID, escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{Resource: "TenantNode", ID: name}
	}
	if err := requireUniqueName("TenantNode", name, nodes, func(n TenantNode) any { return n.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("TenantNode", name, nodes[0].Name, name); err != nil {
		return nil, err
	}
	// Get full details
	got, err := s.Get(ctx, int(nodes[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("TenantNode", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a new tenant node and returns the created node.
func (s *TenantNodeService) Create(ctx context.Context, req *TenantNodeCreateRequest) (*TenantNode, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Tenant <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	if req.CPUCores <= 0 {
		return nil, &ValidationError{Field: "cpu_cores", Message: "cpu_cores is required"}
	}
	if req.RAM < 2048 {
		return nil, &ValidationError{Field: "ram", Message: "ram must be at least 2048 MB"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/tenant_nodes", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created node's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created node
	return s.Get(ctx, id)
}

// Update updates a tenant node and returns the updated node.
func (s *TenantNodeService) Update(ctx context.Context, id int, req *TenantNodeUpdateRequest) (*TenantNode, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/tenant_nodes/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantNode", ID: id}
		}
		return nil, err
	}

	// Read back the updated node
	return s.Get(ctx, id)
}

// Delete deletes a tenant node.
func (s *TenantNodeService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/tenant_nodes/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "TenantNode", ID: id}
		}
		return err
	}
	return nil
}

// PowerOn powers on a tenant node.
func (s *TenantNodeService) PowerOn(ctx context.Context, id int) error {
	action := tenantNodeAction{
		TenantNode: id,
		Action:     "poweron",
	}

	if err := s.client.post(ctx, "/tenant_node_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power on tenant node %d: %w", id, err)
	}
	return nil
}

// PowerOff powers off a tenant node.
func (s *TenantNodeService) PowerOff(ctx context.Context, id int) error {
	action := tenantNodeAction{
		TenantNode: id,
		Action:     "poweroff",
	}

	if err := s.client.post(ctx, "/tenant_node_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to power off tenant node %d: %w", id, err)
	}
	return nil
}

// Reset resets a tenant node.
func (s *TenantNodeService) Reset(ctx context.Context, id int) error {
	action := tenantNodeAction{
		TenantNode: id,
		Action:     "reset",
	}

	if err := s.client.post(ctx, "/tenant_node_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to reset tenant node %d: %w", id, err)
	}
	return nil
}

// Kill forcefully terminates a tenant node.
func (s *TenantNodeService) Kill(ctx context.Context, id int) error {
	action := tenantNodeAction{
		TenantNode: id,
		Action:     "kill",
	}

	if err := s.client.post(ctx, "/tenant_node_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to kill tenant node %d: %w", id, err)
	}
	return nil
}

// Migrate migrates a tenant node to another host.
func (s *TenantNodeService) Migrate(ctx context.Context, id int, targetNode int) error {
	action := tenantNodeAction{
		TenantNode: id,
		Action:     "migrate",
	}
	if targetNode > 0 {
		action.Params = map[string]any{
			"node": targetNode,
		}
	}

	if err := s.client.post(ctx, "/tenant_node_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to migrate tenant node %d: %w", id, err)
	}
	return nil
}

// tenantNodeAction represents a tenant node action request.
type tenantNodeAction struct {
	TenantNode int            `json:"tenant_node"`
	Action     string         `json:"action"`
	Params     map[string]any `json:"params,omitempty"`
}

// TenantStatusService handles tenant status read operations.
// Each tenant has exactly one status record.
type TenantStatusService struct {
	client *Client
}

// List returns all tenant statuses.
func (s *TenantStatusService) List(ctx context.Context, opts ...ListOption) ([]TenantStatus, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = tenantStatusListFields
	}
	params := options.toQueryParams()

	var statuses []TenantStatus
	if err := s.client.get(ctx, "/tenant_status", params, &statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}

// Get returns the status for a specific tenant by its tenant key.
func (s *TenantStatusService) Get(ctx context.Context, tenantKey int) (*TenantStatus, error) {
	params := url.Values{}
	params.Set("fields", tenantStatusGetFields)
	params.Set("filter", fmt.Sprintf("tenant eq %d", tenantKey))

	var statuses []TenantStatus
	if err := s.client.get(ctx, "/tenant_status", params, &statuses); err != nil {
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, &NotFoundError{Resource: "TenantStatus", ID: tenantKey}
	}

	return &statuses[0], nil
}

// GetByKey returns the status by its direct row key.
func (s *TenantStatusService) GetByKey(ctx context.Context, key int) (*TenantStatus, error) {
	params := url.Values{}
	params.Set("fields", tenantStatusGetFields)

	var status TenantStatus
	endpoint := fmt.Sprintf("/tenant_status/%d", key)
	if err := s.client.get(ctx, endpoint, params, &status); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantStatus", ID: key}
		}
		return nil, err
	}

	return &status, nil
}

// TenantStatsHistoryShortService handles tenant short-term statistics history read operations.
// This service provides high-resolution historical metrics for tenant monitoring.
type TenantStatsHistoryShortService struct {
	client *Client
}

// List returns short-term historical stats for all tenants.
func (s *TenantStatsHistoryShortService) List(ctx context.Context, opts ...ListOption) ([]TenantStatsHistoryShort, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = tenantStatsHistoryShortFields
	}

	params := options.toQueryParams()

	var stats []TenantStatsHistoryShort
	if err := s.client.get(ctx, "/tenant_stats_history_short", params, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// ListByTenant returns short-term historical stats for a specific tenant.
func (s *TenantStatsHistoryShortService) ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantStatsHistoryShort, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("tenant eq %d", tenantID)))
	return s.List(ctx, opts...)
}

// GetLatest returns the most recent short-term stats record for a tenant.
func (s *TenantStatsHistoryShortService) GetLatest(ctx context.Context, tenantID int) (*TenantStatsHistoryShort, error) {
	stats, err := s.ListByTenant(ctx, tenantID, WithSort("-timestamp"), WithLimit(1))
	if err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, &NotFoundError{Resource: "TenantStatsHistoryShort", ID: fmt.Sprintf("tenant=%d", tenantID)}
	}

	return &stats[0], nil
}

// Get returns a single short-term stats record by ID.
func (s *TenantStatsHistoryShortService) Get(ctx context.Context, id int) (*TenantStatsHistoryShort, error) {
	params := url.Values{}
	params.Set("fields", tenantStatsHistoryShortFields)

	var stats TenantStatsHistoryShort
	endpoint := fmt.Sprintf("/tenant_stats_history_short/%d", id)
	if err := s.client.get(ctx, endpoint, params, &stats); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantStatsHistoryShort", ID: id}
		}
		return nil, err
	}

	return &stats, nil
}

// TenantStorageService handles tenant storage allocation operations.
type TenantStorageService struct {
	client *Client
}

// List returns all tenant storage allocations, with optional filtering and pagination.
func (s *TenantStorageService) List(ctx context.Context, opts ...ListOption) ([]TenantStorage, error) {
	options := applyListOptions(opts)

	// Use storage-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = tenantStorageListFields
	}

	params := options.toQueryParams()

	var storage []TenantStorage
	if err := s.client.get(ctx, "/tenant_storage", params, &storage); err != nil {
		return nil, err
	}

	return storage, nil
}

// ListByTenant returns all storage allocations for a specific tenant.
func (s *TenantStorageService) ListByTenant(ctx context.Context, tenantID int, opts ...ListOption) ([]TenantStorage, error) {
	// Prepend tenant filter to any existing filters
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("tenant eq %d", tenantID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns a single tenant storage allocation by ID.
func (s *TenantStorageService) Get(ctx context.Context, id int) (*TenantStorage, error) {
	params := url.Values{}
	params.Set("fields", tenantStorageGetFields)

	var storage TenantStorage
	endpoint := fmt.Sprintf("/tenant_storage/%d", id)
	if err := s.client.get(ctx, endpoint, params, &storage); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantStorage", ID: id}
		}
		return nil, err
	}

	return &storage, nil
}

// Create creates a new tenant storage allocation and returns the created allocation.
func (s *TenantStorageService) Create(ctx context.Context, req *TenantStorageCreateRequest) (*TenantStorage, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Tenant <= 0 {
		return nil, &ValidationError{Field: "tenant", Message: "tenant is required"}
	}
	if req.Tier <= 0 {
		return nil, &ValidationError{Field: "tier", Message: "tier is required"}
	}
	if req.Provisioned <= 0 {
		return nil, &ValidationError{Field: "provisioned", Message: "provisioned storage is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/tenant_storage", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created storage's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created storage
	return s.Get(ctx, id)
}

// Update updates a tenant storage allocation and returns the updated allocation.
func (s *TenantStorageService) Update(ctx context.Context, id int, req *TenantStorageUpdateRequest) (*TenantStorage, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/tenant_storage/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TenantStorage", ID: id}
		}
		return nil, err
	}

	// Read back the updated storage
	return s.Get(ctx, id)
}

// Delete deletes a tenant storage allocation.
func (s *TenantStorageService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/tenant_storage/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "TenantStorage", ID: id}
		}
		return err
	}
	return nil
}
