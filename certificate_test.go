package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCertificateService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/certificates": func(w http.ResponseWriter, r *http.Request) {
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
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", Type: "letsencrypt", Valid: true})
		},
	}))

	cert, err := client.Certificates.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cert.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", cert.Domain)
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
			if fields == "" {
				t.Error("expected fields parameter")
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
			var req CertificateCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.DomainName != "example.com" {
				t.Errorf("expected domain 'example.com', got %q", req.DomainName)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1, "status": "OK"})
		},
		"GET /api/v4/certificates/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Certificate{Key: 1, Domain: "example.com", Type: "self_signed"})
		},
	}))

	cert, err := client.Certificates.Create(context.Background(), &CertificateCreateRequest{
		DomainName: "example.com",
		Type:       "self_signed",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if cert.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", cert.Domain)
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
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
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
