---
title: Certificates
description: Manage SSL/TLS certificates including Let's Encrypt, manual, and self-signed
tags: [certificate, ssl, tls, lets-encrypt, self-signed, pem, renewal, https]
categories: [Certificates]
---

# Certificates

Manage SSL/TLS certificates including Let's Encrypt, manual, and self-signed.

```go
// List all certificates
certs, err := client.Certificates.List(ctx)

// List valid certificates
validCerts, err := client.Certificates.ListValid(ctx)

// Get a certificate by domain
cert, err := client.Certificates.GetByDomain(ctx, "example.com")

// Create a Let's Encrypt certificate
cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
    DomainList: "example.com,www.example.com",
    Type:       vergeos.CertificateTypeLetsEncrypt,
    Contact:    ptr(adminUserID),
    AgreeTOS:   ptr(true),
})

// Create a self-signed certificate
cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
    DomainName: "internal.local",
    Type:       vergeos.CertificateTypeSelfSigned,
    KeyType:    vergeos.CertificateKeyTypeECDSA,
})

// Upload a manual certificate
cert, err := client.Certificates.Create(ctx, &vergeos.CertificateCreateRequest{
    DomainName: "secure.example.com",
    Type:       vergeos.CertificateTypeManual,
    Public:     publicKeyPEM,
    Private:    privateKeyPEM,
    Chain:      chainPEM,
})

// Get certificate with keys (for export)
certWithKeys, err := client.Certificates.GetWithKeys(ctx, certID)
fmt.Println(certWithKeys.Public)   // PEM public key
fmt.Println(certWithKeys.Private)  // PEM private key

// Renew a Let's Encrypt certificate
cert, err = client.Certificates.Renew(ctx, certID)

// Delete a certificate
err = client.Certificates.Delete(ctx, certID)
```
