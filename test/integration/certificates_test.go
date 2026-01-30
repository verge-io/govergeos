//go:build integration

package integration

import (
	"context"
	"testing"

	vergeos "github.com/verge-io/govergeos"
)

// TestCertificatesList tests the Certificates service.
func TestCertificatesList(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Certificates service...")

	// List all certificates
	certs, err := client.Certificates.List(ctx)
	if err != nil {
		t.Fatalf("Certificates.List failed: %v", err)
	}

	t.Logf("Found %d certificates", len(certs))

	if len(certs) == 0 {
		t.Log("No certificates found - this is normal for fresh installations")
		return
	}

	// Log first certificate to verify field mapping
	first := certs[0]
	t.Logf("First certificate: Key=%d, Domain=%q, Type=%q, Valid=%v, Expires=%d",
		int(first.Key), first.Domain, first.Type, first.Valid, first.Expires)

	// Test Get by ID
	t.Run("Get", func(t *testing.T) {
		fetched, err := client.Certificates.Get(ctx, int(first.Key))
		if err != nil {
			t.Errorf("Certificates.Get(%d) failed: %v", int(first.Key), err)
		} else {
			t.Logf("Certificates.Get succeeded: Domain=%q, KeyType=%q, AutoCreated=%v",
				fetched.Domain, fetched.KeyType, fetched.AutoCreated)
		}
	})

	// Test ListValid
	t.Run("ListValid", func(t *testing.T) {
		validCerts, err := client.Certificates.ListValid(ctx)
		if err != nil {
			t.Errorf("Certificates.ListValid failed: %v", err)
		} else {
			t.Logf("Found %d valid certificates", len(validCerts))
		}
	})

	// Test ListByType
	t.Run("ListByType", func(t *testing.T) {
		if first.Type == "" {
			t.Skip("No certificate type available")
		}
		typeCerts, err := client.Certificates.ListByType(ctx, first.Type)
		if err != nil {
			t.Errorf("Certificates.ListByType failed: %v", err)
		} else {
			t.Logf("Found %d certificates of type %q", len(typeCerts), first.Type)
		}
	})

	// Test GetByDomain
	t.Run("GetByDomain", func(t *testing.T) {
		if first.Domain == "" {
			t.Skip("No certificate domain available")
		}
		byDomain, err := client.Certificates.GetByDomain(ctx, first.Domain)
		if err != nil {
			t.Errorf("Certificates.GetByDomain failed: %v", err)
		} else {
			t.Logf("GetByDomain succeeded: Key=%d", int(byDomain.Key))
		}
	})

	// Pretty print first certificate for field verification
	prettyPrint(t, "Sample Certificate", first)
}

// TestCertificatesCRUD tests certificate create/update/delete operations.
// This creates a self-signed certificate to avoid needing external domains.
func TestCertificatesCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Certificate CRUD with self-signed certificate...")

	// Create a self-signed certificate
	cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
		DomainName:  "sdk-test.local",
		Type:        vergeos.CertificateTypeSelfSigned,
		Description: "goVergeOS integration test certificate - safe to delete",
		KeyType:     vergeos.CertificateKeyTypeECDSA,
	})
	if err != nil {
		t.Fatalf("Certificates.Create (self-signed) failed: %v", err)
	}
	certID := int(cert.Key)
	t.Logf("Created certificate: [%d] Domain=%q Type=%q Valid=%v", certID, cert.Domain, cert.Type, cert.Valid)

	// Cleanup: delete certificate when done
	defer func() {
		t.Log("Cleaning up: deleting test certificate...")
		if err := client.Certificates.Delete(ctx, certID); err != nil {
			t.Logf("Warning: failed to delete test certificate: %v", err)
		} else {
			t.Log("Test certificate deleted successfully")
		}
	}()

	// Read
	cert, err = client.Certificates.Get(ctx, certID)
	if err != nil {
		t.Fatalf("Certificates.Get failed: %v", err)
	}
	t.Logf("Read certificate: [%d] Domain=%q Expires=%d", certID, cert.Domain, cert.Expires)

	// Get with keys
	certWithKeys, err := client.Certificates.GetWithKeys(ctx, certID)
	if err != nil {
		t.Errorf("Certificates.GetWithKeys failed: %v", err)
	} else {
		t.Logf("GetWithKeys: Public key length=%d, Private present=%v",
			len(certWithKeys.Public), certWithKeys.Private != "")
	}

	// Update
	newDesc := "Updated goVergeOS test certificate"
	cert, err = client.Certificates.Update(ctx, certID, &vergeos.CertificateUpdateRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Errorf("Certificates.Update failed: %v", err)
	} else {
		t.Logf("Updated certificate description to: %q", cert.Description)
	}

	// Verify deletion after cleanup runs
	t.Log("Certificate CRUD test completed")
}
