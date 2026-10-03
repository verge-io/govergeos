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
//   - VERGEOS_HOST: Host or base URL (required). A host with no scheme is treated as https.
//   - VERGEOS_USERNAME + VERGEOS_PASSWORD: Basic authentication
//   - VERGEOS_API_KEY: Bearer token authentication. Used when set, even if username and password are also set.
//   - VERGEOS_VERIFY_SSL: TLS verification, "true" or "false" (default: "true")
//   - VERGEOS_INSECURE: Skip TLS verification when "true". Same as VERGEOS_VERIFY_SSL=false. An error if the two disagree.
//   - VERGEOS_TIMEOUT: Request timeout in seconds (default: "30")
//
// # Authentication
//
// The library supports HTTP Basic Authentication and API key authentication.
// When both are configured, the API key is used. For basic auth, MFA must be
// disabled for the user account.
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
// # Retries
//
// GET, PUT, and DELETE are retried on a connection reset, a timeout before
// any response, and HTTP 429, 502, and 503. The default is 3 attempts with
// exponential backoff and jitter. POST is not retried. HTTP 401 is never
// retried, because a repeated failed login locks the account.
// WithRetry changes or disables the policy. WithRateLimit spaces request
// starts and is off unless set.
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
//	    // Authentication failed (HTTP 401)
//	}
//	if vergeos.IsPermissionError(err) {
//	    // Authenticated, but not allowed (HTTP 403)
//	}
//	if vergeos.IsConflictError(err) {
//	    // Conflicts with existing state (HTTP 409)
//	}
//
// A 404 from create or update is an APIError carrying the platform message.
// IsNotFoundError is false for that error, because VergeOS also returns 404
// when a related row named in the request is missing. Get and Delete of a
// missing id are NotFoundError.
package vergeos
