package vergeos

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// RequiredMajorVersion is the oldest VergeOS major this SDK accepts by default.
// Newer majors are accepted. WithMinimumVersion overrides this floor, and
// WithSkipVersionCheck disables the comparison.
const RequiredMajorVersion = 26

// UnsupportedVersionError is returned when the server is older than the
// minimum major this client accepts.
type UnsupportedVersionError struct {
	ServerVersion string
	Required      int
}

func (e *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("unsupported server version %s: this SDK requires VergeOS %d.0 or later",
		e.ServerVersion, e.Required)
}

// IsUnsupportedVersionError returns true if err is an UnsupportedVersionError.
func IsUnsupportedVersionError(err error) bool {
	var unsupportedErr *UnsupportedVersionError
	return errors.As(err, &unsupportedErr)
}

// versionResponse matches the /version.json response format.
type versionResponse struct {
	Version string `json:"version"`
}

// parseVersion extracts major version from version string.
// Handles formats: "26.0.0", "v26.0.0", "26.0.0-beta1"
func parseVersion(v string) (int, error) {
	v = strings.TrimPrefix(v, "v")
	// Handle dash-suffixed versions like "26.0.0-beta1"
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("empty version string")
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	return major, nil
}

// isVersionAtLeast reports whether a VergeOS version meets a minimum version.
// Missing minor or patch components are treated as zero.
func isVersionAtLeast(v string, requiredMajor, requiredMinor, requiredPatch int) bool {
	v = strings.TrimPrefix(v, "v")
	if idx := strings.Index(v, "+"); idx != -1 {
		v = v[:idx]
	}
	prerelease := false
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
		prerelease = true
	}

	parts := strings.Split(v, ".")
	current := [3]int{}
	for i := 0; i < len(current) && i < len(parts); i++ {
		if parts[i] == "" {
			return false
		}
		value, err := strconv.Atoi(parts[i])
		if err != nil {
			return false
		}
		current[i] = value
	}

	required := [3]int{requiredMajor, requiredMinor, requiredPatch}
	for i := range current {
		if current[i] != required[i] {
			return current[i] > required[i]
		}
	}
	return !prerelease
}

// checkServerVersion fetches /version.json and rejects majors older than the
// client's minimum. The default minimum is RequiredMajorVersion. Newer majors
// are accepted. WithSkipVersionCheck still records the version and does not
// reject it. Called during NewClient().
//
// Credentials are not sent. /version.json does not require them, and sending
// a bad password here would be a second login attempt.
func (c *Client) checkServerVersion(ctx context.Context) error {
	var resp versionResponse
	if err := c.getAbsolute(ctx, "/version.json", nil, &resp); err != nil {
		return fmt.Errorf("failed to check server version: %w", err)
	}

	if c.skipVersionCheck {
		c.serverVersion = resp.Version
		return nil
	}

	major, err := parseVersion(resp.Version)
	if err != nil {
		return fmt.Errorf("failed to parse server version %q: %w", resp.Version, err)
	}
	minimum := c.minimumMajor()
	if major < minimum {
		return &UnsupportedVersionError{
			ServerVersion: resp.Version,
			Required:      minimum,
		}
	}
	c.serverVersion = resp.Version
	return nil
}

// minimumMajor returns the oldest major this client will accept.
func (c *Client) minimumMajor() int {
	if c.minimumMajorVersion > 0 {
		return c.minimumMajorVersion
	}
	return RequiredMajorVersion
}

// VergeOS 26.2.0 is the first release that has the 26.2 field set.
// The minimum supported major stays RequiredMajorVersion (26.0).
const (
	vergeOS262Major = 26
	vergeOS262Minor = 2
	vergeOS262Patch = 0
)

// atLeast262 reports whether the version recorded at NewClient is 26.2.0 or later.
// An empty version is not. A 26.2 prerelease is not.
func (c *Client) atLeast262() bool {
	if c == nil {
		return false
	}
	return isVersionAtLeast(c.serverVersion, vergeOS262Major, vergeOS262Minor, vergeOS262Patch)
}

// require262 returns an error when the recorded version is older than 26.2.
// An empty version is allowed so a test double can call the method.
func (c *Client) require262(feature string) error {
	if c == nil || c.serverVersion == "" || c.atLeast262() {
		return nil
	}
	return fmt.Errorf("vergeos: %s requires VergeOS 26.2 or later (server is %s)", feature, c.serverVersion)
}

// appendVersionFields adds extra when include is set.
func appendVersionFields(base, extra string, include bool) string {
	if !include || extra == "" {
		return base
	}
	if base == "" {
		return extra
	}
	return base + "," + extra
}
