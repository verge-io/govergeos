package vergeos

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ParentFirewallStatus reports the parent network's firewall apply state
// after a tenant network block or external IP change.
//
// VergeOS leaves need_fw_apply set on that parent network and does not apply
// the rules itself. Pending is that flag. Pass WithApplyParentFirewall to
// apply the rules in the same call; Applied is then true when that apply
// succeeded.
type ParentFirewallStatus struct {
	// NetworkID is the parent network (vnet) the change belongs to.
	NetworkID int
	// Applied is true when this call applied the parent network's rules.
	Applied bool
	// Pending is true when the parent network still has need_fw_apply set.
	Pending bool
}

// ParentFirewallOption configures the firewall follow-up for a tenant
// network block or external IP change.
type ParentFirewallOption func(*parentFirewallOptions)

type parentFirewallOptions struct {
	apply bool
}

// WithApplyParentFirewall applies the parent network's firewall rules after
// the change. Without this option the call still returns ParentFirewallStatus
// so a pending need_fw_apply is visible.
func WithApplyParentFirewall() ParentFirewallOption {
	return func(opts *parentFirewallOptions) {
		opts.apply = true
	}
}

func applyParentFirewallOptions(opts []ParentFirewallOption) parentFirewallOptions {
	var cfg parentFirewallOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// parentFirewallFollowUp reads need_fw_apply on the parent network and, when
// requested, applies its rules first.
//
// A nil error means the status was read. Pending still reflects the network
// after a successful apply, so a flag that remains set stays visible.
func parentFirewallFollowUp(ctx context.Context, client *Client, networkID int, opts []ParentFirewallOption) (*ParentFirewallStatus, error) {
	status := &ParentFirewallStatus{NetworkID: networkID}
	if networkID <= 0 {
		return status, nil
	}

	cfg := applyParentFirewallOptions(opts)
	var applyErr error
	if cfg.apply {
		applyErr = client.Networks.ApplyRules(ctx, networkID)
		if applyErr == nil {
			status.Applied = true
		}
	}

	network, err := client.Networks.Get(ctx, networkID)
	if err != nil {
		if applyErr != nil {
			return status, fmt.Errorf("failed to apply rules to network %d: %w", networkID, applyErr)
		}
		return status, fmt.Errorf("failed to read firewall status for network %d: %w", networkID, err)
	}
	status.Pending = network.NeedFWApply
	if applyErr != nil {
		return status, fmt.Errorf("failed to apply rules to network %d: %w", networkID, applyErr)
	}
	return status, nil
}

const tenantOwnerPrefix = "tenants/"

// tenantOwnerRef is the VergeOS owner path for a tenant.
func tenantOwnerRef(tenantID int) string {
	return tenantOwnerPrefix + strconv.Itoa(tenantID)
}

// tenantKeyFromOwner parses a tenant ID from an owner path of the form
// "tenants/{id}". It returns 0 when owner is not a tenant reference.
func tenantKeyFromOwner(owner string) int {
	rest, ok := strings.CutPrefix(owner, tenantOwnerPrefix)
	if !ok || rest == "" {
		return 0
	}
	id, err := strconv.Atoi(rest)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}
