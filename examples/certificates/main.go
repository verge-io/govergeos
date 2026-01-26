// Example: SSL/TLS Certificate Management
//
// This example demonstrates how to manage SSL/TLS certificates in VergeOS.
// Certificates can be Let's Encrypt (ACME), manually uploaded, or self-signed.
//
// Usage:
//
//	export VERGEOS_HOST=https://your-vergeos-host
//	export VERGEOS_USERNAME=admin
//	export VERGEOS_PASSWORD=yourpassword
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vergeos "github.com/verge-io/vergeos-go-sdk"
)

func main() {
	// Get configuration from environment
	host := os.Getenv("VERGEOS_HOST")
	username := os.Getenv("VERGEOS_USERNAME")
	password := os.Getenv("VERGEOS_PASSWORD")

	if host == "" || username == "" || password == "" {
		log.Fatal("Please set VERGEOS_HOST, VERGEOS_USERNAME, and VERGEOS_PASSWORD environment variables")
	}

	// Create client
	client, err := vergeos.NewClient(
		vergeos.WithBaseURL(host),
		vergeos.WithCredentials(username, password),
		vergeos.WithInsecureTLS(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// =========================================================================
	// List all certificates
	// =========================================================================
	fmt.Println("=== Listing All Certificates ===")
	certs, err := client.Certificates.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list certificates: %v", err)
	}

	fmt.Printf("Found %d certificate(s)\n", len(certs))
	for _, cert := range certs {
		expires := "N/A"
		if cert.Expires > 0 {
			t := time.Unix(cert.Expires, 0)
			expires = t.Format("2006-01-02")
		}

		validStatus := "INVALID"
		if cert.Valid {
			validStatus = "VALID"
		}

		fmt.Printf("  - %s (ID: %d)\n", cert.Domain, cert.Key)
		fmt.Printf("    Type: %s, Key: %s, Status: %s\n", cert.Type, cert.KeyType, validStatus)
		fmt.Printf("    Expires: %s\n", expires)
		if cert.DomainList != "" {
			fmt.Printf("    SANs: %s\n", cert.DomainList)
		}
	}

	// =========================================================================
	// List valid certificates only
	// =========================================================================
	fmt.Println("\n=== Valid Certificates ===")
	validCerts, err := client.Certificates.ListValid(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to list valid certificates: %v\n", err)
	} else {
		fmt.Printf("Found %d valid certificate(s)\n", len(validCerts))
		for _, cert := range validCerts {
			daysUntilExpiry := int64(0)
			if cert.Expires > 0 {
				daysUntilExpiry = (cert.Expires - time.Now().Unix()) / 86400
			}
			fmt.Printf("  - %s (%s) - expires in %d days\n", cert.Domain, cert.Type, daysUntilExpiry)
		}
	}

	// =========================================================================
	// List certificates by type
	// =========================================================================
	fmt.Println("\n=== Certificates by Type ===")

	// Let's Encrypt certificates
	leCerts, err := client.Certificates.ListByType(ctx, vergeos.CertificateTypeLetsEncrypt)
	if err != nil {
		fmt.Printf("Warning: Failed to list Let's Encrypt certificates: %v\n", err)
	} else {
		fmt.Printf("Let's Encrypt: %d certificate(s)\n", len(leCerts))
		for _, cert := range leCerts {
			fmt.Printf("  - %s (auto-renew)\n", cert.Domain)
		}
	}

	// Self-signed certificates
	ssCerts, err := client.Certificates.ListByType(ctx, vergeos.CertificateTypeSelfSigned)
	if err != nil {
		fmt.Printf("Warning: Failed to list self-signed certificates: %v\n", err)
	} else {
		fmt.Printf("Self-Signed: %d certificate(s)\n", len(ssCerts))
		for _, cert := range ssCerts {
			fmt.Printf("  - %s\n", cert.Domain)
		}
	}

	// Manual certificates
	manualCerts, err := client.Certificates.ListByType(ctx, vergeos.CertificateTypeManual)
	if err != nil {
		fmt.Printf("Warning: Failed to list manual certificates: %v\n", err)
	} else {
		fmt.Printf("Manual: %d certificate(s)\n", len(manualCerts))
		for _, cert := range manualCerts {
			fmt.Printf("  - %s\n", cert.Domain)
		}
	}

	// =========================================================================
	// Get certificate details
	// =========================================================================
	if len(certs) > 0 {
		cert := certs[0]
		fmt.Println("\n=== Certificate Details ===")
		certDetail, err := client.Certificates.Get(ctx, int(cert.Key))
		if err != nil {
			fmt.Printf("Warning: Failed to get certificate details: %v\n", err)
		} else {
			fmt.Printf("Domain: %s\n", certDetail.Domain)
			fmt.Printf("Type: %s\n", certDetail.Type)
			fmt.Printf("Key Type: %s\n", certDetail.KeyType)
			if certDetail.RSAKeySize != "" {
				fmt.Printf("RSA Key Size: %s\n", certDetail.RSAKeySize)
			}
			fmt.Printf("Valid: %v\n", certDetail.Valid)
			fmt.Printf("Auto-Created: %v\n", certDetail.AutoCreated)
			if certDetail.DomainList != "" {
				fmt.Printf("Additional Domains (SANs): %s\n", certDetail.DomainList)
			}
			if certDetail.ACMEServer != "" {
				fmt.Printf("ACME Server: %s\n", certDetail.ACMEServer)
			}
			if certDetail.Expires > 0 {
				expires := time.Unix(certDetail.Expires, 0)
				fmt.Printf("Expires: %s\n", expires.Format("2006-01-02 15:04:05"))
			}
			if certDetail.Created > 0 {
				created := time.Unix(certDetail.Created, 0)
				fmt.Printf("Created: %s\n", created.Format("2006-01-02 15:04:05"))
			}
		}

		// Get certificate by domain name
		fmt.Println("\n=== Lookup by Domain ===")
		foundCert, err := client.Certificates.GetByDomain(ctx, cert.Domain)
		if err != nil {
			fmt.Printf("Warning: Failed to find certificate by domain: %v\n", err)
		} else {
			fmt.Printf("Found certificate for '%s' (ID: %d)\n", foundCert.Domain, foundCert.Key)
		}
	}

	// =========================================================================
	// Certificate expiration check
	// =========================================================================
	fmt.Println("\n=== Expiration Status ===")
	now := time.Now().Unix()
	warningDays := int64(30)
	criticalDays := int64(7)

	for _, cert := range certs {
		if cert.Expires == 0 {
			continue
		}

		daysUntilExpiry := (cert.Expires - now) / 86400
		status := "OK"

		if daysUntilExpiry < 0 {
			status = "EXPIRED"
		} else if daysUntilExpiry < criticalDays {
			status = "CRITICAL"
		} else if daysUntilExpiry < warningDays {
			status = "WARNING"
		}

		if status != "OK" {
			fmt.Printf("  [%s] %s - expires in %d days\n", status, cert.Domain, daysUntilExpiry)
		}
	}

	fmt.Println("\n=== Certificate Operations Reference ===")
	fmt.Println("Listing:")
	fmt.Println("  - client.Certificates.List(ctx)                    - All certificates")
	fmt.Println("  - client.Certificates.ListValid(ctx)               - Valid certificates only")
	fmt.Println("  - client.Certificates.ListByType(ctx, type)        - By type")
	fmt.Println("  - client.Certificates.Get(ctx, id)                 - Get by ID")
	fmt.Println("  - client.Certificates.GetByDomain(ctx, domain)     - Get by domain name")
	fmt.Println("  - client.Certificates.GetWithKeys(ctx, id)         - Get with PEM keys")
	fmt.Println("\nCreation:")
	fmt.Println("  - Let's Encrypt: Type=letsencrypt, DomainList, Contact, AgreeTOS")
	fmt.Println("  - Self-Signed: Type=self_signed, DomainName, KeyType")
	fmt.Println("  - Manual: Type=manual, DomainName, Public, Private, Chain")
	fmt.Println("\nManagement:")
	fmt.Println("  - client.Certificates.Create(ctx, req)   - Create certificate")
	fmt.Println("  - client.Certificates.Update(ctx, id, req) - Update certificate")
	fmt.Println("  - client.Certificates.Delete(ctx, id)    - Delete certificate")
	fmt.Println("  - client.Certificates.Renew(ctx, id)     - Renew Let's Encrypt cert")

	fmt.Println("\n=== Certificate Type Constants ===")
	fmt.Printf("  vergeos.CertificateTypeLetsEncrypt = %q\n", vergeos.CertificateTypeLetsEncrypt)
	fmt.Printf("  vergeos.CertificateTypeSelfSigned  = %q\n", vergeos.CertificateTypeSelfSigned)
	fmt.Printf("  vergeos.CertificateTypeManual      = %q\n", vergeos.CertificateTypeManual)
	fmt.Printf("  vergeos.CertificateKeyTypeECDSA    = %q\n", vergeos.CertificateKeyTypeECDSA)
	fmt.Printf("  vergeos.CertificateKeyTypeRSA      = %q\n", vergeos.CertificateKeyTypeRSA)

	fmt.Println("\n=== Done ===")
}
