package vergeos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOIDCApplication_DiscardsClientSecret(t *testing.T) {
	payload := []byte(`{"$key":4,"name":"Tenant Portal","client_id":"cid","client_secret":"` + probeSecret + `","enabled":true}`)
	var app OIDCApplication
	if err := json.Unmarshal(payload, &app); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if app.ClientSecret.Value() != "" {
		t.Fatal("OIDC application kept the client secret")
	}
	if app.ClientID != "cid" || app.Name != "Tenant Portal" {
		t.Fatalf("app = %+v", app)
	}
	encoded, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	assertNoClientSecret(t, "marshaled OIDC application", string(encoded))
	if strings.Contains(string(encoded), clientSecretField) {
		t.Fatal("marshaled OIDC application included client_secret")
	}
	assertNoClientSecret(t, "formatted OIDC application", fmt.Sprintf("%v %+v %#v %s", app, app, app, app))
}

func TestOIDCApplicationService_ListAndGet(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/oidc_applications": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if strings.Contains(fields, clientSecretField) {
				t.Errorf("list fields included client_secret: %s", fields)
			}
			jsonResponse(w, 200, []map[string]any{
				{"$key": 1, "name": "Tenant Portal", "client_id": "cid", "client_secret": probeSecret, "enabled": true},
			})
		},
		"GET /api/v4/oidc_applications/1": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if strings.Contains(fields, clientSecretField) {
				t.Errorf("get fields included client_secret: %s", fields)
			}
			if !strings.Contains(fields, "well_known_configuration") {
				t.Errorf("get fields = %s", fields)
			}
			jsonResponse(w, 200, map[string]any{
				"$key":                     1,
				"name":                     "Tenant Portal",
				"client_id":                "cid",
				"client_secret":            probeSecret,
				"enabled":                  true,
				"well_known_configuration": "https://verge.example/.well-known/openid-configuration",
			})
		},
	}))

	apps, err := client.OIDCApplications.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(apps) != 1 || apps[0].ClientSecret.Value() != "" {
		t.Fatal("List kept a client secret")
	}
	assertNoClientSecret(t, "listed application", fmt.Sprintf("%+v", apps[0]))

	app, err := client.OIDCApplications.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if app.WellKnownConfiguration == "" || app.ClientID != "cid" {
		t.Fatalf("app = %+v", app)
	}
	if app.ClientSecret.Value() != "" {
		t.Fatal("Get kept a client secret")
	}
	assertNoClientSecret(t, "Get result", fmt.Sprintf("%+v", app))
}

func TestOIDCApplicationService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/oidc_applications/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))
	_, err := client.OIDCApplications.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestOIDCApplicationService_GetByName(t *testing.T) {
	const name = `portal{x}`
	wantFilter := "name eq '" + escapeFilterValue(name) + "'"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/oidc_applications": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != wantFilter {
				t.Errorf("filter = %q, want %q", got, wantFilter)
			}
			jsonResponse(w, 200, []OIDCApplication{{Key: 2, Name: name, ClientID: "cid"}})
		},
		"GET /api/v4/oidc_applications/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, OIDCApplication{Key: 2, Name: name, ClientID: "cid"})
		},
	}))

	app, err := client.OIDCApplications.GetByName(context.Background(), name)
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if app.Name != name || app.Key != 2 {
		t.Fatalf("app = %+v", app)
	}
}

func TestOIDCApplicationService_GetByName_Ambiguous(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/oidc_applications": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []OIDCApplication{
				{Key: 1, Name: "shared"},
				{Key: 2, Name: "shared"},
			})
		},
	}))
	_, err := client.OIDCApplications.GetByName(context.Background(), "shared")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("expected AmbiguousNameError, got %v", err)
	}
}

func TestOIDCApplicationService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/oidc_applications": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assertNoClientSecret(t, "create request", string(body))
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode create: %v", err)
			}
			if payload["name"] != "Tenant Portal" {
				t.Fatalf("name = %#v", payload["name"])
			}
			if payload["enabled"] != true || payload["scope_profile"] != true || payload["scope_email"] != true || payload["scope_groups"] != true {
				t.Fatalf("defaults = %#v", payload)
			}
			if payload["force_auth_source"] != nil || payload["map_user"] != nil {
				t.Fatalf("unset row references = %#v %#v", payload["force_auth_source"], payload["map_user"])
			}
			if payload["redirect_uri"] != "https://tenant.example.com/callback" {
				t.Fatalf("redirect_uri = %#v", payload["redirect_uri"])
			}
			jsonResponse(w, 200, map[string]any{"$key": 4, "response": map[string]any{"client_secret": "from-post"}})
		},
		"GET /api/v4/oidc_applications/4": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, clientSecretField) {
				t.Errorf("create read did not request client_secret: %s", fields)
			}
			jsonResponse(w, 200, map[string]any{
				"$key":          4,
				"name":          "Tenant Portal",
				"client_id":     "generated-id",
				"client_secret": probeSecret,
				"enabled":       true,
			})
		},
	}))

	req := &OIDCApplicationCreateRequest{
		Name:        "Tenant Portal",
		RedirectURI: "https://tenant.example.com/callback",
		Description: "OIDC for tenant authentication",
	}
	app, secret, err := client.OIDCApplications.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if req.Enabled != nil || req.ScopeProfile != nil {
		t.Fatal("Create changed the caller's request")
	}
	if secret.Value() != probeSecret {
		t.Fatal("Create did not return the generated client secret")
	}
	assertNoClientSecret(t, "formatted secret", fmt.Sprintf("%v %+v %#v %s", secret, secret, secret, secret))
	if app.ClientID != "generated-id" || app.ClientSecret.Value() != "" {
		t.Fatal("Create stored the client secret on the application")
	}
	encoded, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Marshal app: %v", err)
	}
	assertNoClientSecret(t, "marshaled created application", string(encoded))
	assertNoClientSecret(t, "formatted created application", fmt.Sprintf("%+v", app))
}

func TestOIDCApplicationService_Create_SecretFromPostWhenGetOmitsIt(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/oidc_applications": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 5, "client_secret": probeSecret})
		},
		"GET /api/v4/oidc_applications/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 5, "name": "Portal", "client_id": "cid"})
		},
	}))

	app, secret, err := client.OIDCApplications.Create(context.Background(), &OIDCApplicationCreateRequest{Name: "Portal"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if app.Name != "Portal" || secret.Value() != probeSecret {
		t.Fatal("Create dropped the secret returned by POST")
	}
	if app.ClientSecret.Value() != "" {
		t.Fatal("application retained the POST secret")
	}
}

func TestOIDCApplicationService_Create_ExplicitFalseEnabled(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/oidc_applications": func(w http.ResponseWriter, r *http.Request) {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if payload["enabled"] != false {
				t.Fatalf("enabled = %#v", payload["enabled"])
			}
			authSource, _ := payload["force_auth_source"].(float64)
			if authSource != 8 {
				t.Fatalf("force_auth_source = %#v", payload["force_auth_source"])
			}
			jsonResponse(w, 200, apiResponse{Key: float64(6)})
		},
		"GET /api/v4/oidc_applications/6": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, OIDCApplication{Key: 6, Name: "disabled", Enabled: false})
		},
	}))

	enabled := false
	authSource := 8
	app, _, err := client.OIDCApplications.Create(context.Background(), &OIDCApplicationCreateRequest{
		Name:            "disabled",
		Enabled:         &enabled,
		ForceAuthSource: &authSource,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if app.Key != 6 {
		t.Fatalf("key = %d", app.Key)
	}
}

func TestOIDCApplicationService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	if _, _, err := client.OIDCApplications.Create(context.Background(), nil); !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}
	if _, _, err := client.OIDCApplications.Create(context.Background(), &OIDCApplicationCreateRequest{}); !IsValidationError(err) {
		t.Fatalf("missing name: %v", err)
	}
}

func TestOIDCApplicationService_UpdateAndDelete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/oidc_applications/1": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assertNoClientSecret(t, "update request", string(body))
			var req OIDCApplicationUpdateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if req.RedirectURI == nil || *req.RedirectURI != "https://new.example/callback" {
				t.Fatalf("redirect = %#v", req.RedirectURI)
			}
			if strings.Contains(string(body), clientSecretField) {
				t.Fatal("update request included client_secret")
			}
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/oidc_applications/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, OIDCApplication{Key: 1, Name: "Tenant Portal", RedirectURI: "https://new.example/callback"})
		},
		"DELETE /api/v4/oidc_applications/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		"DELETE /api/v4/oidc_applications/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	redirect := "https://new.example/callback"
	app, err := client.OIDCApplications.Update(context.Background(), 1, &OIDCApplicationUpdateRequest{RedirectURI: &redirect})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if app.RedirectURI != redirect {
		t.Fatalf("redirect = %q", app.RedirectURI)
	}
	if _, err := client.OIDCApplications.Update(context.Background(), 1, nil); !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}
	if err := client.OIDCApplications.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := client.OIDCApplications.Delete(context.Background(), 9); !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
