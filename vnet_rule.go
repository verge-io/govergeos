package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// VNetRuleService handles network firewall rule operations.
type VNetRuleService struct {
	client *Client
}

// List returns all firewall rules, with optional filtering and pagination.
func (s *VNetRuleService) List(ctx context.Context, opts ...ListOption) ([]VNetRule, error) {
	options := applyListOptions(opts)

	// Use rule-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = vnetRuleListFields
	}

	params := options.toQueryParams()

	var rules []VNetRule
	if err := s.client.get(ctx, "/vnet_rules", params, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// ListByNetwork returns all firewall rules for a specific network.
func (s *VNetRuleService) ListByNetwork(ctx context.Context, vnetID int, opts ...ListOption) ([]VNetRule, error) {
	// Prepend network filter to any existing filters
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("vnet eq %d", vnetID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns a single firewall rule by ID.
func (s *VNetRuleService) Get(ctx context.Context, id int) (*VNetRule, error) {
	params := url.Values{}
	params.Set("fields", vnetRuleGetFields)

	var rule VNetRule
	endpoint := fmt.Sprintf("/vnet_rules/%d", id)
	if err := s.client.get(ctx, endpoint, params, &rule); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetRule", ID: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	return &rule, nil
}

// GetByName returns a firewall rule by name within a specific network.
func (s *VNetRuleService) GetByName(ctx context.Context, vnetID int, name string) (*VNetRule, error) {
	rules, err := s.List(ctx, WithFilter(fmt.Sprintf("vnet eq %d and name eq '%s'", vnetID, name)))
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, &NotFoundError{Resource: "VNetRule", ID: name}
	}
	// Get full details
	return s.Get(ctx, int(rules[0].Key))
}

// Create creates a new firewall rule and returns the created rule.
func (s *VNetRuleService) Create(ctx context.Context, req *VNetRuleCreateRequest) (*VNetRule, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.VNet <= 0 {
		return nil, &ValidationError{Field: "vnet", Message: "vnet is required"}
	}

	// Set defaults
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnet_rules", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created rule's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created rule
	return s.Get(ctx, id)
}

// Update updates a firewall rule and returns the updated rule.
func (s *VNetRuleService) Update(ctx context.Context, id int, req *VNetRuleUpdateRequest) (*VNetRule, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/vnet_rules/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetRule", ID: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	// Read back the updated rule
	return s.Get(ctx, id)
}

// Delete deletes a firewall rule.
func (s *VNetRuleService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vnet_rules/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VNetRule", ID: fmt.Sprintf("%d", id)}
		}
		return err
	}
	return nil
}

// Enable enables a firewall rule.
// If apply is true, firewall rules are automatically applied to the network.
// If forceApply is true, rules are applied even if there are other pending changes.
func (s *VNetRuleService) Enable(ctx context.Context, id int, apply bool, forceApply bool) error {
	action := vnetRuleAction{
		Rule:   id,
		Action: "enable",
		Params: vnetRuleActionParams{
			Apply:      apply,
			ForceApply: forceApply,
		},
	}

	if err := s.client.post(ctx, "/vnet_rule_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to enable rule %d: %w", id, err)
	}
	return nil
}

// Disable disables a firewall rule.
// If apply is true, firewall rules are automatically applied to the network.
// If forceApply is true, rules are applied even if there are other pending changes.
func (s *VNetRuleService) Disable(ctx context.Context, id int, apply bool, forceApply bool) error {
	action := vnetRuleAction{
		Rule:   id,
		Action: "disable",
		Params: vnetRuleActionParams{
			Apply:      apply,
			ForceApply: forceApply,
		},
	}

	if err := s.client.post(ctx, "/vnet_rule_actions", action, nil); err != nil {
		return fmt.Errorf("vergeos: failed to disable rule %d: %w", id, err)
	}
	return nil
}

// vnetRuleAction represents a rule action request.
type vnetRuleAction struct {
	Rule   int                  `json:"vnet_rule"`
	Action string               `json:"action"`
	Params vnetRuleActionParams `json:"params"`
}

// vnetRuleActionParams are the parameters for rule actions.
type vnetRuleActionParams struct {
	Apply      bool `json:"apply"`
	ForceApply bool `json:"force_apply"`
}

// VNetRuleAliasService handles network rule alias operations.
// Aliases allow you to define reusable address lists for firewall rules.
type VNetRuleAliasService struct {
	client *Client
}

// List returns all rule aliases, with optional filtering and pagination.
func (s *VNetRuleAliasService) List(ctx context.Context, opts ...ListOption) ([]VNetRuleAlias, error) {
	options := applyListOptions(opts)

	// Use alias-specific fields if not specified
	if options.Fields == "most" {
		options.Fields = vnetRuleAliasListFields
	}

	params := options.toQueryParams()

	var aliases []VNetRuleAlias
	if err := s.client.get(ctx, "/vnet_rule_aliases", params, &aliases); err != nil {
		return nil, err
	}

	return aliases, nil
}

// Get returns a single rule alias by ID.
func (s *VNetRuleAliasService) Get(ctx context.Context, id int) (*VNetRuleAlias, error) {
	params := url.Values{}
	params.Set("fields", vnetRuleAliasGetFields)

	var alias VNetRuleAlias
	endpoint := fmt.Sprintf("/vnet_rule_aliases/%d", id)
	if err := s.client.get(ctx, endpoint, params, &alias); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetRuleAlias", ID: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	return &alias, nil
}

// GetByName returns a rule alias by name.
func (s *VNetRuleAliasService) GetByName(ctx context.Context, name string) (*VNetRuleAlias, error) {
	aliases, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", name)))
	if err != nil {
		return nil, err
	}
	if len(aliases) == 0 {
		return nil, &NotFoundError{Resource: "VNetRuleAlias", ID: name}
	}
	// Get full details
	return s.Get(ctx, int(aliases[0].Key))
}

// Create creates a new rule alias and returns the created alias.
func (s *VNetRuleAliasService) Create(ctx context.Context, req *VNetRuleAliasCreateRequest) (*VNetRuleAlias, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.Value == "" {
		return nil, &ValidationError{Field: "value", Message: "value is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vnet_rule_aliases", req, &resp); err != nil {
		return nil, err
	}

	// Extract the created alias's ID
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}

	// Read back the created alias
	return s.Get(ctx, id)
}

// Update updates a rule alias and returns the updated alias.
func (s *VNetRuleAliasService) Update(ctx context.Context, id int, req *VNetRuleAliasUpdateRequest) (*VNetRuleAlias, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}

	endpoint := fmt.Sprintf("/vnet_rule_aliases/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "VNetRuleAlias", ID: fmt.Sprintf("%d", id)}
		}
		return nil, err
	}

	// Read back the updated alias
	return s.Get(ctx, id)
}

// Delete deletes a rule alias.
func (s *VNetRuleAliasService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/vnet_rule_aliases/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "VNetRuleAlias", ID: fmt.Sprintf("%d", id)}
		}
		return err
	}
	return nil
}
