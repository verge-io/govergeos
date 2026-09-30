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

const probeSecret = "probe-secret"

func assertNoClientSecret(t *testing.T, label, text string) {
	t.Helper()
	if strings.Contains(text, probeSecret) {
		t.Fatalf("%s included the client secret", label)
	}
}

func TestWriteOnlySecret_RedactsAndDiscards(t *testing.T) {
	secret := NewWriteOnlySecret(probeSecret)
	printed := fmt.Sprintf("%v %+v %#v %s %q", secret, secret, secret, secret, secret)
	assertNoClientSecret(t, "formatted WriteOnlySecret", printed)
	if !strings.Contains(printed, redactedSecret) {
		t.Fatal("formatted WriteOnlySecret did not redact")
	}
	if secret.Value() != probeSecret {
		t.Fatal("Value did not return the secret")
	}

	encoded, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(encoded) != `"`+probeSecret+`"` {
		t.Fatal("MarshalJSON did not write the secret")
	}

	var decoded WriteOnlySecret
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Value() != "" {
		t.Fatal("UnmarshalJSON retained a client secret")
	}
	remarshaled, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("remarshal: %v", err)
	}
	assertNoClientSecret(t, "remarshaled WriteOnlySecret", string(remarshaled))
}

func TestAuthSourceSettings_RedactsSecretAndDecodesWithoutIt(t *testing.T) {
	settings := AuthSourceSettings{
		"client_id":     "probe-client",
		"client_secret": probeSecret,
		"scope":         "openid profile email",
	}
	printed := fmt.Sprintf("%v %+v %#v %s", settings, settings, settings, settings)
	assertNoClientSecret(t, "formatted settings", printed)
	if !strings.Contains(printed, "probe-client") {
		t.Fatalf("formatted settings dropped client_id: %s", printed)
	}

	req := &AuthSourceCreateRequest{
		Name:     "probe",
		Driver:   AuthSourceDriverOpenID,
		Settings: settings,
	}
	printed = fmt.Sprintf("%v %+v %#v %s", req, req, req, req)
	assertNoClientSecret(t, "formatted create request", printed)

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal request: %v", err)
	}
	if !strings.Contains(string(body), probeSecret) {
		t.Fatal("marshaled create request omitted the client secret")
	}

	var decoded AuthSource
	payload := []byte(`{"$key":1,"name":"probe","driver":"openid","settings":{"client_id":"probe-client","client_secret":"` + probeSecret + `","scope":"openid","debug":false}}`)
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal source: %v", err)
	}
	if _, ok := decoded.Settings[clientSecretField]; ok {
		t.Fatal("decoded settings kept client_secret")
	}
	if decoded.Settings["client_id"] != "probe-client" {
		t.Fatalf("client_id = %#v", decoded.Settings["client_id"])
	}
	if decoded.Settings["debug"] != false {
		t.Fatalf("debug = %#v", decoded.Settings["debug"])
	}
	remarshaled, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("remarshal source: %v", err)
	}
	assertNoClientSecret(t, "remarshaled auth source", string(remarshaled))
	printed = fmt.Sprintf("%+v", decoded)
	assertNoClientSecret(t, "formatted auth source", printed)

	var fromString AuthSourceSettings
	if err := json.Unmarshal([]byte(`"{\"client_id\":\"probe-client\",\"client_secret\":\"`+probeSecret+`\"}"`), &fromString); err != nil {
		t.Fatalf("Unmarshal string settings: %v", err)
	}
	if _, ok := fromString[clientSecretField]; ok {
		t.Fatal("string settings kept client_secret")
	}
	if fromString["client_id"] != "probe-client" {
		t.Fatalf("string settings client_id = %#v", fromString["client_id"])
	}
}

func TestAuthSourceService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if strings.Contains(fields, "settings") || strings.Contains(fields, clientSecretField) {
				t.Errorf("list fields included settings or client_secret: %s", fields)
			}
			jsonResponse(w, 200, []AuthSource{
				{Key: 1, Name: "Azure AD", Driver: AuthSourceDriverAzure},
				{Key: 2, Name: "Google", Driver: AuthSourceDriverGoogle},
			})
		},
	}))

	sources, err := client.AuthSources.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sources) != 2 || sources[0].Name != "Azure AD" || sources[1].Driver != AuthSourceDriverGoogle {
		t.Fatalf("sources = %+v", sources)
	}
}

func TestAuthSourceService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("fields"), "settings") {
				t.Errorf("get fields = %s", r.URL.Query().Get("fields"))
			}
			jsonResponse(w, 200, map[string]any{
				"$key":   1,
				"name":   "probe",
				"driver": AuthSourceDriverOpenID,
				"settings": map[string]any{
					"client_id":     "probe-client",
					"client_secret": probeSecret,
					"scope":         "openid profile email",
				},
			})
		},
	}))

	source, err := client.AuthSources.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if source.Name != "probe" || source.Driver != AuthSourceDriverOpenID {
		t.Fatalf("source = %+v", source)
	}
	if _, ok := source.Settings[clientSecretField]; ok {
		t.Fatal("Get kept client_secret")
	}
	if source.Settings["client_id"] != "probe-client" {
		t.Fatalf("client_id = %#v", source.Settings["client_id"])
	}
	printed := fmt.Sprintf("%+v", source)
	assertNoClientSecret(t, "Get result", printed)
}

func TestAuthSourceService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.AuthSources.Get(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestAuthSourceService_GetByName(t *testing.T) {
	const name = `o'brien{x}`
	wantFilter := "name eq '" + escapeFilterValue(name) + "'"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("filter"); got != wantFilter {
				t.Errorf("filter = %q, want %q", got, wantFilter)
			}
			jsonResponse(w, 200, []AuthSource{{Key: 3, Name: name, Driver: AuthSourceDriverOkta}})
		},
		"GET /api/v4/auth_sources/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, AuthSource{Key: 3, Name: name, Driver: AuthSourceDriverOkta})
		},
	}))

	source, err := client.AuthSources.GetByName(context.Background(), name)
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if source.Name != name || source.Key != 3 {
		t.Fatalf("source = %+v", source)
	}
}

func TestAuthSourceService_GetByName_NotFoundAmbiguousAndMismatch(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
			"GET /api/v4/auth_sources": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []AuthSource{})
			},
		}))
		_, err := client.AuthSources.GetByName(context.Background(), "missing")
		if !IsNotFoundError(err) {
			t.Fatalf("expected NotFoundError, got %v", err)
		}
	})

	t.Run("ambiguous", func(t *testing.T) {
		client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
			"GET /api/v4/auth_sources": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []AuthSource{
					{Key: 1, Name: "shared"},
					{Key: 2, Name: "shared"},
				})
			},
		}))
		_, err := client.AuthSources.GetByName(context.Background(), "shared")
		if !IsAmbiguousNameError(err) {
			t.Fatalf("expected AmbiguousNameError, got %v", err)
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
			"GET /api/v4/auth_sources": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []AuthSource{{Key: 1, Name: "other"}})
			},
		}))
		_, err := client.AuthSources.GetByName(context.Background(), "wanted")
		if !IsNotFoundError(err) {
			t.Fatalf("expected NotFoundError, got %v", err)
		}
	})
}

func TestAuthSourceService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/auth_sources": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), probeSecret) {
				t.Fatal("create request omitted the client secret")
			}
			var req AuthSourceCreateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode create: %v", err)
			}
			if req.Name != "Corporate Azure" || req.Driver != AuthSourceDriverAzure {
				t.Fatalf("create body name=%q driver=%q", req.Name, req.Driver)
			}
			if req.ButtonFAIcon != "bi-microsoft" {
				t.Fatalf("icon = %q", req.ButtonFAIcon)
			}
			// Decoding the request drops client_secret. The raw body check
			// above is what proves it was sent.
			jsonResponse(w, 200, apiResponse{Key: float64(7)})
		},
		"GET /api/v4/auth_sources/7": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{
				"$key":   7,
				"name":   "Corporate Azure",
				"driver": AuthSourceDriverAzure,
				"settings": map[string]any{
					"client_id":     "client",
					"client_secret": probeSecret,
					"tenant_id":     "tenant",
				},
			})
		},
	}))

	source, err := client.AuthSources.Create(context.Background(), &AuthSourceCreateRequest{
		Name:   "Corporate Azure",
		Driver: AuthSourceDriverAzure,
		Settings: AuthSourceSettings{
			"tenant_id":     "tenant",
			"client_id":     "client",
			"client_secret": NewWriteOnlySecret(probeSecret),
		},
		ButtonFAIcon: "bi-microsoft",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if source.Key != 7 || source.Settings["client_id"] != "client" {
		t.Fatalf("source = %+v", source)
	}
	if _, ok := source.Settings[clientSecretField]; ok {
		t.Fatal("Create result kept client_secret")
	}
	assertNoClientSecret(t, "Create result", fmt.Sprintf("%+v", source))
}

func TestAuthSourceService_Create_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	ctx := context.Background()

	if _, err := client.AuthSources.Create(ctx, nil); !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}
	if _, err := client.AuthSources.Create(ctx, &AuthSourceCreateRequest{Driver: AuthSourceDriverGoogle}); !IsValidationError(err) {
		t.Fatalf("missing name: %v", err)
	}
	if _, err := client.AuthSources.Create(ctx, &AuthSourceCreateRequest{Name: "Google"}); !IsValidationError(err) {
		t.Fatalf("missing driver: %v", err)
	}
}

func TestAuthSourceService_Update_MergesSettings(t *testing.T) {
	var puts int
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields == authSourceSettingsFields {
				jsonResponse(w, 200, map[string]any{
					"$key": 1,
					"settings": map[string]any{
						"client_id":              "probe-client",
						"client_secret":          probeSecret,
						"scope":                  "openid profile email",
						"authorization_endpoint": "https://idp.example/auth",
						"debug":                  false,
					},
				})
				return
			}
			jsonResponse(w, 200, map[string]any{
				"$key":   1,
				"name":   "probe",
				"driver": AuthSourceDriverOpenID,
				"settings": map[string]any{
					"client_id":              "probe-client",
					"client_secret":          probeSecret,
					"scope":                  "openid",
					"authorization_endpoint": "https://idp.example/auth",
					"debug":                  false,
				},
			})
		},
		"PUT /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			puts++
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), probeSecret) {
				t.Fatal("merged settings omitted the stored client secret")
			}
			var req struct {
				Name     string         `json:"name"`
				Settings map[string]any `json:"settings"`
			}
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode put: %v", err)
			}
			if req.Name != "renamed" {
				t.Fatalf("name = %q", req.Name)
			}
			settings := req.Settings
			if settings["scope"] != "openid" || settings["client_id"] != "probe-client" {
				t.Fatalf("merged scope/client_id = %#v / %#v", settings["scope"], settings["client_id"])
			}
			if settings["authorization_endpoint"] != "https://idp.example/auth" {
				t.Fatal("merged settings dropped authorization_endpoint")
			}
			if settings["debug"] != false {
				t.Fatal("merged settings dropped the server-injected debug key")
			}
			w.WriteHeader(http.StatusOK)
		},
	}))

	patch := AuthSourceSettings{"scope": "openid"}
	name := "renamed"
	source, err := client.AuthSources.Update(context.Background(), 1, &AuthSourceUpdateRequest{
		Name:     &name,
		Settings: patch,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if puts != 1 {
		t.Fatalf("puts = %d", puts)
	}
	if _, ok := patch[clientSecretField]; ok {
		t.Fatal("Update wrote the stored secret into the caller's settings")
	}
	if _, ok := source.Settings[clientSecretField]; ok {
		t.Fatal("Update result kept client_secret")
	}
	if source.Settings["client_id"] != "probe-client" || source.Settings["scope"] != "openid" {
		t.Fatalf("result settings = %#v", map[string]any(source.Settings))
	}
	assertNoClientSecret(t, "Update result", fmt.Sprintf("%+v", source))
}

func TestAuthSourceService_Update_StringEncodedSettings(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("fields") == authSourceSettingsFields {
				jsonResponse(w, 200, map[string]any{
					"$key":     1,
					"settings": `{"client_id":"probe-client","client_secret":"` + probeSecret + `"}`,
				})
				return
			}
			jsonResponse(w, 200, AuthSource{Key: 1, Name: "probe", Driver: AuthSourceDriverOpenID})
		},
		"PUT /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), probeSecret) || !strings.Contains(string(body), "probe-client") {
				t.Fatal("string-encoded settings were not merged into the PUT")
			}
			if strings.Contains(string(body), `\"client_secret\"`) {
				t.Fatal("settings were sent as an encoded string")
			}
			w.WriteHeader(http.StatusOK)
		},
	}))

	_, err := client.AuthSources.Update(context.Background(), 1, &AuthSourceUpdateRequest{
		Settings: AuthSourceSettings{"scope": "openid"},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestAuthSourceService_Update_WithoutSettingsDoesNotReadThem(t *testing.T) {
	var settingsReads int
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("fields") == authSourceSettingsFields {
				settingsReads++
			}
			jsonResponse(w, 200, AuthSource{Key: 1, Name: "renamed"})
		},
		"PUT /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "settings") {
				t.Fatalf("name-only update sent settings: %s", body)
			}
			w.WriteHeader(http.StatusOK)
		},
	}))

	name := "renamed"
	source, err := client.AuthSources.Update(context.Background(), 1, &AuthSourceUpdateRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if settingsReads != 0 {
		t.Fatalf("settings reads = %d", settingsReads)
	}
	if source.Name != "renamed" {
		t.Fatalf("name = %q", source.Name)
	}
}

func TestAuthSourceService_Update_InvalidSettings(t *testing.T) {
	var puts int
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 1, "settings": []any{"nope"}})
		},
		"PUT /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			puts++
			w.WriteHeader(http.StatusOK)
		},
	}))

	_, err := client.AuthSources.Update(context.Background(), 1, &AuthSourceUpdateRequest{
		Settings: AuthSourceSettings{"scope": "openid"},
	})
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if puts != 0 {
		t.Fatal("invalid settings were written")
	}
	if strings.Contains(err.Error(), probeSecret) {
		t.Fatal("validation error included a client secret")
	}
}

func TestAuthSourceService_Update_NilAndNotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/auth_sources/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	if _, err := client.AuthSources.Update(context.Background(), 1, nil); !IsValidationError(err) {
		t.Fatalf("nil request: %v", err)
	}
	_, err := client.AuthSources.Update(context.Background(), 9, &AuthSourceUpdateRequest{
		Settings: AuthSourceSettings{"scope": "openid"},
	})
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestAuthSourceService_Delete(t *testing.T) {
	deleted := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/auth_sources/1": func(w http.ResponseWriter, r *http.Request) {
			deleted = true
			w.WriteHeader(http.StatusOK)
		},
		"DELETE /api/v4/auth_sources/9": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	if err := client.AuthSources.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Fatal("expected DELETE")
	}
	if err := client.AuthSources.Delete(context.Background(), 9); !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
