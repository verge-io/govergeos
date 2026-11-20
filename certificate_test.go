package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCertificateService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, "domainname") {
				t.Errorf("fields %q do not request domainname", fields)
			}
			jsonResponse(w, 200, []Certificate{
				{Key: 1, Domain: "example.com", Type: "letsencrypt", Valid: true},
				{Key: 2, Domain: "test.com", Type: "manual", Valid: true},
			})
		},
	}))

	certs, err := client.Certificates.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(certs) != 2 {
		t.Fatalf("expected 2 certificates, got %d", len(certs))
	}
	if certs[0].Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", certs[0].Domain)
	}
}

func TestCertificateService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, "domainname") {
				t.Errorf("fields %q do not request domainname", fields)
			}
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", DomainName: "example.com", Type: "letsencrypt", Valid: true})
		},
	}))

	cert, err := client.Certificates.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cert.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", cert.Domain)
	}
	if cert.DomainName != "example.com" {
		t.Errorf("expected domainname 'example.com', got %q", cert.DomainName)
	}
	if cert.Type != "letsencrypt" {
		t.Errorf("expected type 'letsencrypt', got %q", cert.Type)
	}
}

func TestCertificateService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Certificates.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCertificateService_GetByDomain(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "domain eq 'example.com'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Certificate{{Key: 1, Domain: "example.com"}})
		},
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", Type: "letsencrypt"})
		},
	}))

	cert, err := client.Certificates.GetByDomain(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("GetByDomain failed: %v", err)
	}
	if cert.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", cert.Domain)
	}
}

func TestCertificateService_GetByDomain_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Certificate{})
		},
	}))

	_, err := client.Certificates.GetByDomain(context.Background(), "nonexistent.com")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCertificateService_GetWithKeys(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if !strings.Contains(fields, "domainname") {
				t.Errorf("fields %q do not request domainname", fields)
			}
			if !strings.Contains(fields, "public") || !strings.Contains(fields, "private") || !strings.Contains(fields, "chain") {
				t.Errorf("fields %q do not request the key material", fields)
			}
			jsonResponse(w, 200, Certificate{
				Key:     1,
				Domain:  "example.com",
				Public:  "-----BEGIN CERTIFICATE-----\nMIIB...",
				Private: "-----BEGIN PRIVATE KEY-----\nMIIE...",
				Chain:   "-----BEGIN CERTIFICATE-----\nMIIC...",
			})
		},
	}))

	cert, err := client.Certificates.GetWithKeys(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetWithKeys failed: %v", err)
	}
	if cert.Public == "" {
		t.Error("expected non-empty public key")
	}
	if cert.Private == "" {
		t.Error("expected non-empty private key")
	}
}

func TestCertificateService_GetWithKeys_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Certificates.GetWithKeys(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCertificateService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read body: %v", err)
				return
			}
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Errorf("decode body: %v", err)
				return
			}
			if _, ok := body["domainname"]; ok {
				t.Errorf("create body sent domainname (%#v), which VergeOS ignores", body["domainname"])
			}
			if body["domainlist"] != "example.com" {
				t.Errorf("domainlist = %#v, want example.com", body["domainlist"])
			}
			if body["type"] != "self_signed" {
				t.Errorf("type = %#v, want self_signed", body["type"])
			}
			if body["key_type"] != "ecdsa" {
				t.Errorf("key_type = %#v, want ecdsa", body["key_type"])
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", DomainName: "example.com", Type: "self_signed"})
		},
	}))

	req := &CertificateCreateRequest{
		DomainName: "example.com",
		Type:       "self_signed",
		KeyType:    CertificateKeyTypeECDSA,
	}
	cert, err := client.Certificates.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if req.DomainList != "" {
		t.Errorf("Create changed the caller's DomainList to %q", req.DomainList)
	}
	if cert.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", cert.Domain)
	}
	if cert.DomainName != "example.com" {
		t.Errorf("expected domainname 'example.com', got %q", cert.DomainName)
	}
}

func TestCertificateCreateRequest_MarshalJSON(t *testing.T) {
	contact := 7
	agree := true
	tests := []struct {
		name       string
		req        CertificateCreateRequest
		domainlist string
		omitList   bool
	}{
		{
			name: "domain name only",
			req: CertificateCreateRequest{
				DomainName: "internal.bugtest.example.com",
				Type:       CertificateTypeSelfSigned,
				KeyType:    CertificateKeyTypeECDSA,
			},
			domainlist: "internal.bugtest.example.com",
		},
		{
			name: "primary already first",
			req: CertificateCreateRequest{
				DomainName: "bugtest-both.example.com",
				DomainList: "bugtest-both.example.com,alt.bugtest-both.example.com",
				Type:       CertificateTypeSelfSigned,
			},
			domainlist: "bugtest-both.example.com,alt.bugtest-both.example.com",
		},
		{
			name: "primary added in front of extra names",
			req: CertificateCreateRequest{
				DomainName: "primary.example.com",
				DomainList: "alt.example.com",
			},
			domainlist: "primary.example.com,alt.example.com",
		},
		{
			name: "primary moved to the front",
			req: CertificateCreateRequest{
				DomainName: "primary.example.com",
				DomainList: "alt.example.com, primary.example.com, extra.example.com, primary.example.com",
			},
			domainlist: "primary.example.com,alt.example.com,extra.example.com",
		},
		{
			name: "trimmed primary",
			req: CertificateCreateRequest{
				DomainName: "  internal.local  ",
			},
			domainlist: "internal.local",
		},
		{
			name: "empty parts dropped",
			req: CertificateCreateRequest{
				DomainName: "primary.example.com",
				DomainList: "alt.example.com,, ,extra.example.com",
			},
			domainlist: "primary.example.com,alt.example.com,extra.example.com",
		},
		{
			name: "list only keeps the caller's text",
			req: CertificateCreateRequest{
				DomainList: "example.com, www.example.com",
				Type:       CertificateTypeLetsEncrypt,
				Contact:    &contact,
				AgreeTOS:   &agree,
			},
			domainlist: "example.com, www.example.com",
		},
		{
			name: "blank primary leaves the list",
			req: CertificateCreateRequest{
				DomainName: "   ",
				DomainList: "example.com",
			},
			domainlist: "example.com",
		},
		{
			name:     "no names",
			req:      CertificateCreateRequest{Type: CertificateTypeManual, Public: "pem", Private: "key"},
			omitList: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalList := tt.req.DomainList
			raw, err := json.Marshal(&tt.req)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if tt.req.DomainList != originalList {
				t.Errorf("marshal changed DomainList from %q to %q", originalList, tt.req.DomainList)
			}
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if _, ok := body["domainname"]; ok {
				t.Errorf("body sent domainname (%#v), which VergeOS ignores", body["domainname"])
			}
			got, ok := body["domainlist"].(string)
			if tt.omitList {
				if ok {
					t.Errorf("domainlist = %q, want it omitted", got)
				}
			} else if !ok || got != tt.domainlist {
				t.Errorf("domainlist = %#v, want %q", body["domainlist"], tt.domainlist)
			}
			if tt.req.Type != "" && body["type"] != tt.req.Type {
				t.Errorf("type = %#v, want %q", body["type"], tt.req.Type)
			}
			if tt.req.KeyType != "" && body["key_type"] != tt.req.KeyType {
				t.Errorf("key_type = %#v, want %q", body["key_type"], tt.req.KeyType)
			}
			if tt.req.Contact != nil && body["contact"] != float64(*tt.req.Contact) {
				t.Errorf("contact = %#v, want %d", body["contact"], *tt.req.Contact)
			}
			if tt.req.AgreeTOS != nil && body["agree_tos"] != *tt.req.AgreeTOS {
				t.Errorf("agree_tos = %#v, want %v", body["agree_tos"], *tt.req.AgreeTOS)
			}
			if tt.req.Public != "" && body["public"] != tt.req.Public {
				t.Errorf("public = %#v, want %q", body["public"], tt.req.Public)
			}
		})
	}
}

func TestCertificateCreateRequest_UnmarshalDomainName(t *testing.T) {
	var req CertificateCreateRequest
	raw := []byte(`{"domainname":"internal.local","type":"self_signed","key_type":"ecdsa"}`)
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.DomainName != "internal.local" {
		t.Fatalf("domain name = %q, want internal.local", req.DomainName)
	}
	out, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(out, &body); err != nil {
		t.Fatalf("decode marshaled body: %v", err)
	}
	if _, ok := body["domainname"]; ok {
		t.Errorf("remarshaled body sent domainname (%#v), which VergeOS ignores", body["domainname"])
	}
	if body["domainlist"] != "internal.local" {
		t.Errorf("domainlist = %#v, want internal.local", body["domainlist"])
	}
}

func TestCertificate_UnmarshalDomainName(t *testing.T) {
	var cert Certificate
	raw := []byte(`{"$key":4,"domain":"","domainname":"verge-api","domainlist":"verge-api,yb-api,localhost","type":"self_signed"}`)
	if err := json.Unmarshal(raw, &cert); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if cert.Domain != "" {
		t.Errorf("domain = %q, want empty", cert.Domain)
	}
	if cert.DomainName != "verge-api" {
		t.Errorf("domainname = %q, want verge-api", cert.DomainName)
	}
	if cert.DomainList != "verge-api,yb-api,localhost" {
		t.Errorf("domainlist = %q", cert.DomainList)
	}
}

func TestCertificateService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Certificates.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCertificateService_Update(t *testing.T) {
	newDesc := "updated cert"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			var req CertificateUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Description == nil || *req.Description != newDesc {
				t.Errorf("expected description %q, got %v", newDesc, req.Description)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", Description: newDesc})
		},
	}))

	cert, err := client.Certificates.Update(context.Background(), 1, &CertificateUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if cert.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, cert.Description)
	}
}

func TestCertificateService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Certificates.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCertificateService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/certificates/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newDesc := "updated"
	_, err := client.Certificates.Update(context.Background(), 999, &CertificateUpdateRequest{Description: &newDesc})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	apiErr, ok := err.(*APIError)
	if IsNotFoundError(err) || !ok || apiErr.StatusCode != 404 {
		t.Errorf("expected API 404, not a missing resource, got %T: %v", err, err)
	}
}

func TestCertificateService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Certificates.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestCertificateService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/certificates/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	err := client.Certificates.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCertificateService_Renew(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			var req CertificateUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Renew == nil || !*req.Renew {
				t.Error("expected renew=true")
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", Type: "letsencrypt", Valid: true})
		},
	}))

	cert, err := client.Certificates.Renew(context.Background(), 1)
	if err != nil {
		t.Fatalf("Renew failed: %v", err)
	}
	if !cert.Valid {
		t.Error("expected cert to be valid after renewal")
	}
}

func TestCertificateService_ListExpiring(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter == "" {
				t.Error("expected filter for expiring certs")
			}
			jsonResponse(w, 200, []Certificate{{Key: 1, Domain: "example.com", Valid: true}})
		},
	}))

	certs, err := client.Certificates.ListExpiring(context.Background(), 30)
	if err != nil {
		t.Fatalf("ListExpiring failed: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(certs))
	}
}

func TestCertificateService_ListValid(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "valid eq true" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Certificate{
				{Key: 1, Domain: "example.com", Valid: true},
			})
		},
	}))

	certs, err := client.Certificates.ListValid(context.Background())
	if err != nil {
		t.Fatalf("ListValid failed: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(certs))
	}
}

func TestCertificateService_ListByType(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "type eq 'letsencrypt'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Certificate{{Key: 1, Domain: "example.com", Type: "letsencrypt"}})
		},
	}))

	certs, err := client.Certificates.ListByType(context.Background(), "letsencrypt")
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(certs))
	}
	if certs[0].Type != "letsencrypt" {
		t.Errorf("expected type 'letsencrypt', got %q", certs[0].Type)
	}
}
