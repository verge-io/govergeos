//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

const (
	// certificateDomainWait bounds how long TestCertificatesCRUD polls for
	// domain and domainname. VergeOS copies them after Create returns.
	certificateDomainWait = 30 * time.Second
	certificateDomainPoll = 200 * time.Millisecond
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
// Domain and domainname can still be empty on the create response. The
// certificates loop copies them after it issues the certificate, so the
// test polls Get before asserting them.
func TestCertificatesCRUD(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	t.Log("Testing Certificate CRUD with self-signed certificate...")

	const domain = "sdk-test.local"

	// Create a self-signed certificate
	cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
		DomainName:  domain,
		Type:        vergeos.CertificateTypeSelfSigned,
		Description: "goVergeOS integration test certificate - safe to delete",
		KeyType:     vergeos.CertificateKeyTypeECDSA,
	})
	if err != nil {
		t.Fatalf("Certificates.Create (self-signed) failed: %v", err)
	}
	certID := int(cert.Key)
	if certID == 0 {
		t.Fatal("Certificates.Create returned a certificate with no key")
	}
	t.Logf("Created certificate: [%d] Domain=%q DomainName=%q Type=%q Valid=%v", certID, cert.Domain, cert.DomainName, cert.Type, cert.Valid)

	// Delete the row this test created. Cleanup uses its own context so a
	// timed-out wait does not cancel the delete.
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		t.Log("Cleaning up: deleting test certificate...")
		if err := client.Certificates.Delete(cleanupCtx, certID); err != nil && !vergeos.IsNotFoundError(err) {
			t.Errorf("failed to delete test certificate %d: %v", certID, err)
		} else {
			t.Log("Test certificate deleted successfully")
		}
	}()

	waitCtx, waitCancel := context.WithTimeout(ctx, certificateDomainWait)
	defer waitCancel()
	issued, err := waitForCertificateDomain(waitCtx, client, certID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Certificate domain populated: Domain=%q DomainName=%q", issued.Domain, issued.DomainName)
	if issued.Domain != domain {
		t.Errorf("certificate domain = %q, want %s", issued.Domain, domain)
	}
	if issued.DomainName != domain {
		t.Errorf("certificate domainname = %q, want %s", issued.DomainName, domain)
	}
	byName, err := client.Certificates.GetByDomain(ctx, domain)
	if err != nil {
		t.Errorf("Certificates.GetByDomain(%s) failed: %v", domain, err)
	} else if int(byName.Key) != certID {
		t.Errorf("GetByDomain key = %d, want %d", int(byName.Key), certID)
	}

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

// waitForCertificateDomain polls Get until domain and domainname are set.
// VergeOS returns the new row before the certificates loop issues the
// certificate and copies those columns.
func waitForCertificateDomain(ctx context.Context, client *vergeos.Client, id int) (*vergeos.Certificate, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, certificateDomainWait)
		defer cancel()
	}

	var last *vergeos.Certificate
	for {
		if err := ctx.Err(); err != nil {
			return nil, certificateDomainTimeout(id, last, err)
		}

		cert, err := client.Certificates.Get(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return nil, certificateDomainTimeout(id, last, ctx.Err())
			}
			return nil, fmt.Errorf("Certificates.Get(%d) while waiting for domain and domainname: %w", id, err)
		}
		last = cert
		if cert.Domain != "" && cert.DomainName != "" {
			return cert, nil
		}

		delay := certificateDomainPoll
		if deadline, ok := ctx.Deadline(); ok {
			if remaining := time.Until(deadline); remaining < delay {
				delay = remaining
			}
		}
		if delay <= 0 {
			err := ctx.Err()
			if err == nil {
				err = context.DeadlineExceeded
			}
			return nil, certificateDomainTimeout(id, last, err)
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, certificateDomainTimeout(id, last, ctx.Err())
		case <-timer.C:
		}
	}
}

func certificateDomainTimeout(id int, cert *vergeos.Certificate, err error) error {
	domain, domainName := "", ""
	if cert != nil {
		domain = cert.Domain
		domainName = cert.DomainName
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("timed out after %s waiting for certificate %d domain and domainname to be populated (domain=%q domainname=%q)", certificateDomainWait, id, domain, domainName)
	}
	return fmt.Errorf("stopped waiting for certificate %d domain and domainname to be populated (domain=%q domainname=%q): %w", id, domain, domainName, err)
}
