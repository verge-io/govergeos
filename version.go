package vergeos

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// RequiredMajorVersion is the VergeOS major version this SDK requires.
const RequiredMajorVersion = 26

// UnsupportedVersionError is returned when the server version is not supported.
type UnsupportedVersionError struct {
	ServerVersion string
	Required      int
}

func (e *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("unsupported server version %s: this SDK requires VergeOS %d.x",
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

// checkServerVersion fetches /version.json and validates the server is v26.
// Called during NewClient() - returns error if version check fails.
func (c *Client) checkServerVersion(ctx context.Context) error {
	var resp versionResponse
	if err := c.getAbsolute(ctx, "/version.json", nil, &resp); err != nil {
		return fmt.Errorf("failed to check server version: %w", err)
	}

	major, err := parseVersion(resp.Version)
	if err != nil {
		return fmt.Errorf("failed to parse server version %q: %w", resp.Version, err)
	}
	if major != RequiredMajorVersion {
		return &UnsupportedVersionError{
			ServerVersion: resp.Version,
			Required:      RequiredMajorVersion,
		}
	}
	return nil
}
