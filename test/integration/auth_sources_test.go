//go:build integration

package integration

import (
	"context"
	"testing"
	"time"
)

// TestAuthSourcesList lists authentication sources without printing settings.
// settings can contain a client secret.
func TestAuthSourcesList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sources, err := client.AuthSources.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list auth sources: %v", err)
	}
	t.Logf("Found %d auth source(s)", len(sources))
	for _, source := range sources {
		if _, ok := source.Settings["client_secret"]; ok {
			t.Fatal("list result included client_secret")
		}
		t.Logf("Auth source: Key=%d Name=%q Driver=%q", int(source.Key), source.Name, source.Driver)
	}
}

// TestOIDCApplicationsList lists OIDC applications without printing client secrets.
func TestOIDCApplicationsList(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apps, err := client.OIDCApplications.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list OIDC applications: %v", err)
	}
	t.Logf("Found %d OIDC application(s)", len(apps))
	for _, app := range apps {
		if app.ClientSecret.Value() != "" {
			t.Fatal("list result included a client secret")
		}
		t.Logf("OIDC application: Key=%d Name=%q ClientID=%q Enabled=%v", int(app.Key), app.Name, app.ClientID, app.Enabled)
	}
}
