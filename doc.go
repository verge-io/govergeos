// Package vergeos provides a Go client library for the VergeOS API.
//
// goVergeOS provides a convenient way to interact with VergeOS infrastructure
// from Go applications. It handles authentication, request building, and
// response parsing.
//
// # Quick Start
//
// Create a client and start making API calls:
//
//	client, err := vergeos.NewClient(
//	    vergeos.WithBaseURL("https://your-vergeos-host"),
//	    vergeos.WithCredentials("username", "password"),
//	    vergeos.WithInsecureTLS(true),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List all VMs
//	vms, err := client.VMs.List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a VM
//	vm, err := client.VMs.Create(ctx, &vergeos.VMCreateRequest{
//	    Name:     "my-vm",
//	    CPUCores: 4,
//	    RAM:      8192,
//	})
//
// # Environment Configuration
//
// Use WithEnvConfig() to configure the client from environment variables:
//
//	client, err := vergeos.NewClient(vergeos.WithEnvConfig())
//
// Environment variables:
//   - VERGEOS_HOST: Base URL (required)
//   - VERGEOS_USERNAME + VERGEOS_PASSWORD: Basic authentication
//   - VERGEOS_API_KEY: Bearer token authentication
//   - VERGEOS_VERIFY_SSL: TLS verification, "true" or "false" (default: "true")
//   - VERGEOS_TIMEOUT: Request timeout in seconds (default: "30")
//
// # Authentication
//
// The library supports HTTP Basic Authentication and API key authentication.
// For basic auth, MFA must be disabled for the user account.
//
// # Query Options
//
// List operations support filtering, sorting, and pagination:
//
//	vms, err := client.VMs.List(ctx,
//	    vergeos.WithFilter("enabled eq true"),
//	    vergeos.WithSort("name"),
//	    vergeos.WithLimit(50),
//	)
//
// # Error Handling
//
// The library provides typed errors and helper functions:
//
//	vm, err := client.VMs.Get(ctx, vmID)
//	if vergeos.IsNotFoundError(err) {
//	    // Handle not found
//	}
//	if vergeos.IsAuthError(err) {
//	    // Handle authentication failure
//	}
package vergeos
