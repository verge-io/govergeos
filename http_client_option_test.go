package vergeos

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// configuredTransport applies the recorded HTTP settings and returns the
// *http.Transport underneath retry and rate limiting.
func configuredTransport(t *testing.T, c *Client) *http.Transport {
	t.Helper()
	if err := c.applyTransportPolicy(); err != nil {
		t.Fatalf("apply transport: %v", err)
	}
	tr, ok := unwrapTransport(c.httpClient.Transport).(*http.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http.Transport", unwrapTransport(c.httpClient.Transport))
	}
	return tr
}

func unwrapTransport(rt http.RoundTripper) http.RoundTripper {
	for {
		switch next := rt.(type) {
		case *retryTransport:
			rt = next.base
		case *rateLimitTransport:
			rt = next.base
		default:
			return rt
		}
	}
}

func permutations(items []string) [][]string {
	if len(items) == 0 {
		return [][]string{{}}
	}
	var out [][]string
	for i, item := range items {
		rest := append(append([]string{}, items[:i]...), items[i+1:]...)
		for _, tail := range permutations(rest) {
			out = append(out, append([]string{item}, tail...))
		}
	}
	return out
}

type foreignTransport struct {
	calls int
}

func (f *foreignTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	body := `[]`
	if req.URL.Path == "/version.json" {
		body = `{"version":"26.1.8"}`
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func fmtBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func TestHTTPClientOptionOrder(t *testing.T) {
	server, _ := newStartupServer(t, "26.1.8", http.StatusOK, `[]`)
	serverTransport := server.Client().Transport.(*http.Transport)
	defTLS := http.DefaultTransport.(*http.Transport).TLSClientConfig

	const (
		callerTimeout   = 45 * time.Second
		envTimeout      = 12 * time.Second
		explicitTimeout = 8 * time.Second
		rateInterval    = 5 * time.Millisecond
	)

	newCaller := func() *http.Client {
		transport := serverTransport.Clone()
		transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		return &http.Client{
			Transport: transport,
			Timeout:   callerTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	assertBuilt := func(t *testing.T, caller *http.Client, opts []ClientOption, wantTimeout time.Duration, wantSkip bool, wantRate time.Duration) {
		t.Helper()
		beforeTimeout := caller.Timeout
		beforeTransport := caller.Transport
		var beforeCfg *tls.Config
		if tr, ok := beforeTransport.(*http.Transport); ok {
			beforeCfg = tr.TLSClientConfig
		}

		client, err := NewClient(opts...)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if caller.Timeout != beforeTimeout || caller.Transport != beforeTransport {
			t.Fatalf("caller client changed: timeout %s transport %p", caller.Timeout, caller.Transport)
		}
		if tr, ok := caller.Transport.(*http.Transport); ok {
			if tr.TLSClientConfig != beforeCfg {
				t.Fatal("caller TLS config was replaced")
			}
			if beforeCfg != nil && (beforeCfg.InsecureSkipVerify || beforeCfg.MinVersion != tls.VersionTLS12) {
				t.Fatal("caller TLS config was modified")
			}
		}
		if http.DefaultTransport.(*http.Transport).TLSClientConfig != defTLS {
			t.Fatal("http.DefaultTransport TLS config changed")
		}
		if client.httpClient == caller {
			t.Fatal("NewClient reused the caller's http.Client")
		}
		if client.httpClient.Timeout != wantTimeout {
			t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, wantTimeout)
		}

		inner := unwrapTransport(client.httpClient.Transport)
		tr, ok := inner.(*http.Transport)
		if !ok {
			t.Fatalf("transport = %T, want *http.Transport", inner)
		}
		gotSkip := tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify
		if gotSkip != wantSkip {
			t.Fatalf("InsecureSkipVerify = %v, want %v", gotSkip, wantSkip)
		}
		if wantSkip && beforeTransport != nil && tr == beforeTransport {
			t.Fatal("insecure TLS reused the caller's transport")
		}
		if !wantSkip && beforeTransport != nil && tr != beforeTransport {
			t.Fatal("transport was cloned when TLS did not change")
		}
		if wantSkip && beforeCfg != nil && tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
			t.Fatal("clone dropped MinVersion")
		}
		if beforeTransport == nil && tr == http.DefaultTransport {
			t.Fatal("built client is using http.DefaultTransport")
		}

		retry, ok := client.httpClient.Transport.(*retryTransport)
		if !ok {
			t.Fatalf("outer transport = %T", client.httpClient.Transport)
		}
		if wantRate > 0 {
			limit, ok := retry.base.(*rateLimitTransport)
			if !ok || limit.interval != wantRate || limit.base != tr {
				t.Fatalf("rate limit = %#v, want interval %s wrapping the TLS transport", retry.base, wantRate)
			}
		} else if _, limited := retry.base.(*rateLimitTransport); limited {
			t.Fatal("rate limit installed")
		}
		if caller.CheckRedirect != nil {
			req, reqErr := http.NewRequest(http.MethodGet, server.URL, nil)
			if reqErr != nil {
				t.Fatal(reqErr)
			}
			if err := client.httpClient.CheckRedirect(req, nil); !errors.Is(err, http.ErrUseLastResponse) {
				t.Fatalf("CheckRedirect = %v", err)
			}
		}
	}

	optsFor := func(order []string, caller *http.Client, insecure bool) []ClientOption {
		opts := []ClientOption{
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
		}
		for _, name := range order {
			switch name {
			case "http":
				opts = append(opts, WithHTTPClient(caller))
			case "timeout":
				opts = append(opts, WithTimeout(explicitTimeout))
			case "insecure":
				opts = append(opts, WithInsecureTLS(insecure))
			case "env":
				opts = append(opts, WithEnvConfig())
			case "rate":
				opts = append(opts, WithRateLimit(rateInterval))
			default:
				t.Fatalf("unknown option %q", name)
			}
		}
		return opts
	}

	runOrders := func(t *testing.T, names []string, env func(), insecure bool, wantTimeout time.Duration, wantSkip bool, wantRate time.Duration, callerFn func() *http.Client) {
		t.Helper()
		for _, order := range permutations(names) {
			order := append([]string(nil), order...)
			t.Run(strings.Join(order, " then "), func(t *testing.T) {
				clearEnvVars()
				t.Cleanup(clearEnvVars)
				if env != nil {
					env()
				}
				caller := callerFn()
				assertBuilt(t, caller, optsFor(order, caller, insecure), wantTimeout, wantSkip, wantRate)
			})
		}
	}

	t.Run("timeout and insecure survive every order with WithHTTPClient", func(t *testing.T) {
		runOrders(t, []string{"http", "insecure", "timeout"}, nil, true, explicitTimeout, true, 0, newCaller)
	})

	t.Run("insecure survives both orders", func(t *testing.T) {
		runOrders(t, []string{"http", "insecure"}, nil, true, callerTimeout, true, 0, newCaller)
	})

	t.Run("timeout survives both orders", func(t *testing.T) {
		runOrders(t, []string{"http", "timeout"}, nil, false, explicitTimeout, false, 0, newCaller)
	})

	t.Run("zero timeout is applied", func(t *testing.T) {
		clearEnvVars()
		t.Cleanup(clearEnvVars)
		caller := newCaller()
		for _, order := range [][]string{{"http", "timeout"}, {"timeout", "http"}} {
			order := append([]string(nil), order...)
			t.Run(strings.Join(order, " then "), func(t *testing.T) {
				opts := []ClientOption{
					WithBaseURL(server.URL),
					WithCredentials("admin", "correct-password"),
				}
				for _, name := range order {
					if name == "http" {
						opts = append(opts, WithHTTPClient(caller))
					} else {
						opts = append(opts, WithTimeout(0))
					}
				}
				assertBuilt(t, caller, opts, 0, false, 0)
			})
		}
	})

	t.Run("rate limit is layered on the copy", func(t *testing.T) {
		runOrders(t, []string{"http", "insecure", "rate"}, nil, true, callerTimeout, true, rateInterval, newCaller)
	})

	envSkip := func() {
		_ = os.Setenv("VERGEOS_VERIFY_SSL", "false")
		_ = os.Setenv("VERGEOS_TIMEOUT", "12")
	}
	t.Run("env TLS and timeout survive both orders", func(t *testing.T) {
		runOrders(t, []string{"http", "env"}, envSkip, false, envTimeout, true, 0, newCaller)
	})

	t.Run("VERGEOS_INSECURE survives both orders", func(t *testing.T) {
		runOrders(t, []string{"http", "env"}, func() {
			_ = os.Setenv("VERGEOS_INSECURE", "true")
			_ = os.Setenv("VERGEOS_TIMEOUT", "12")
		}, false, envTimeout, true, 0, newCaller)
	})

	t.Run("later timeout wins between env and WithTimeout", func(t *testing.T) {
		for _, order := range permutations([]string{"http", "env", "timeout"}) {
			order := append([]string(nil), order...)
			t.Run(strings.Join(order, " then "), func(t *testing.T) {
				clearEnvVars()
				t.Cleanup(clearEnvVars)
				envSkip()
				want := envTimeout
				for _, name := range order {
					if name == "timeout" {
						want = explicitTimeout
					}
					if name == "env" {
						want = envTimeout
					}
				}
				caller := newCaller()
				assertBuilt(t, caller, optsFor(order, caller, false), want, true, 0)
			})
		}
	})

	t.Run("later insecure flag wins between env and WithInsecureTLS", func(t *testing.T) {
		for _, insecure := range []bool{true, false} {
			insecure := insecure
			for _, order := range permutations([]string{"http", "env", "insecure"}) {
				order := append([]string(nil), order...)
				t.Run(fmtBool(insecure)+" "+strings.Join(order, " then "), func(t *testing.T) {
					clearEnvVars()
					t.Cleanup(clearEnvVars)
					envSkip()
					wantSkip := true
					for _, name := range order {
						if name == "insecure" {
							wantSkip = insecure
						}
						if name == "env" {
							wantSkip = true
						}
					}
					caller := newCaller()
					assertBuilt(t, caller, optsFor(order, caller, insecure), envTimeout, wantSkip, 0)
				})
			}
		}
	})

	t.Run("verify SSL true does not clear WithInsecureTLS", func(t *testing.T) {
		runOrders(t, []string{"http", "env", "insecure"}, func() {
			_ = os.Setenv("VERGEOS_VERIFY_SSL", "true")
		}, true, callerTimeout, true, 0, newCaller)
	})

	t.Run("empty client keeps env TLS and timeout in both orders", func(t *testing.T) {
		runOrders(t, []string{"http", "env"}, envSkip, false, envTimeout, true, 0, func() *http.Client {
			return &http.Client{}
		})
	})

	t.Run("WithEnvConfig alone applies timeout and TLS", func(t *testing.T) {
		clearEnvVars()
		t.Cleanup(clearEnvVars)
		envSkip()
		client, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithEnvConfig(),
		)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if client.httpClient.Timeout != envTimeout {
			t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, envTimeout)
		}
		tr, ok := unwrapTransport(client.httpClient.Transport).(*http.Transport)
		if !ok || tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
			t.Fatal("VERGEOS_VERIFY_SSL=false was not applied")
		}
		if http.DefaultTransport.(*http.Transport).TLSClientConfig != defTLS {
			t.Fatal("http.DefaultTransport TLS config changed")
		}
	})

	t.Run("existing skip on a custom transport stays when WithInsecureTLS is false", func(t *testing.T) {
		caller := newCaller()
		caller.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
		client, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithHTTPClient(caller),
			WithInsecureTLS(false),
		)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if !caller.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify {
			t.Fatal("caller InsecureSkipVerify was cleared")
		}
		if unwrapTransport(client.httpClient.Transport) != caller.Transport {
			t.Fatal("WithInsecureTLS(false) replaced the caller's transport")
		}
	})

	t.Run("same caller can build two clients without being modified", func(t *testing.T) {
		clearEnvVars()
		t.Cleanup(clearEnvVars)
		caller := newCaller()
		first, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithHTTPClient(caller),
			WithTimeout(explicitTimeout),
			WithInsecureTLS(true),
		)
		if err != nil {
			t.Fatal(err)
		}
		second, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithInsecureTLS(true),
			WithHTTPClient(caller),
			WithTimeout(envTimeout),
		)
		if err != nil {
			t.Fatal(err)
		}
		if caller.Timeout != callerTimeout || caller.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify {
			t.Fatal("shared caller was modified")
		}
		if first.httpClient.Timeout != explicitTimeout || second.httpClient.Timeout != envTimeout {
			t.Fatalf("timeouts = %s and %s", first.httpClient.Timeout, second.httpClient.Timeout)
		}
		if first.httpClient == second.httpClient || first.httpClient == caller || second.httpClient == caller {
			t.Fatal("clients were not copied")
		}
	})
}

func TestHTTPClientInsecureTLSRequiresTransport(t *testing.T) {
	for _, order := range permutations([]string{"http", "insecure", "timeout"}) {
		order := append([]string(nil), order...)
		t.Run(strings.Join(order, " then "), func(t *testing.T) {
			rt := &foreignTransport{}
			caller := &http.Client{Transport: rt, Timeout: time.Second}
			opts := []ClientOption{
				WithBaseURL("https://verge.example"),
				WithCredentials("admin", "correct-password"),
			}
			for _, name := range order {
				switch name {
				case "http":
					opts = append(opts, WithHTTPClient(caller))
				case "insecure":
					opts = append(opts, WithInsecureTLS(true))
				case "timeout":
					opts = append(opts, WithTimeout(3*time.Second))
				}
			}
			client, err := NewClient(opts...)
			if client != nil {
				t.Fatal("NewClient returned a client")
			}
			if err == nil || !strings.Contains(err.Error(), "insecure TLS") || !strings.Contains(err.Error(), "foreignTransport") {
				t.Fatalf("error = %v", err)
			}
			if rt.calls != 0 {
				t.Fatalf("transport calls = %d, want 0", rt.calls)
			}
			if caller.Timeout != time.Second || caller.Transport != rt {
				t.Fatal("caller client was modified")
			}
		})
	}

	for _, order := range permutations([]string{"http", "env"}) {
		order := append([]string(nil), order...)
		t.Run("env "+strings.Join(order, " then "), func(t *testing.T) {
			clearEnvVars()
			t.Cleanup(clearEnvVars)
			_ = os.Setenv("VERGEOS_VERIFY_SSL", "false")
			_ = os.Setenv("VERGEOS_TIMEOUT", "12")
			rt := &foreignTransport{}
			caller := &http.Client{Transport: rt, Timeout: 2 * time.Second}
			opts := []ClientOption{
				WithBaseURL("https://verge.example"),
				WithAPIKey("key"),
			}
			for _, name := range order {
				if name == "http" {
					opts = append(opts, WithHTTPClient(caller))
				} else {
					opts = append(opts, WithEnvConfig())
				}
			}
			_, err := NewClient(opts...)
			if err == nil || !strings.Contains(err.Error(), "insecure TLS") {
				t.Fatalf("error = %v", err)
			}
			if rt.calls != 0 {
				t.Fatalf("transport calls = %d, want 0", rt.calls)
			}
			if caller.Timeout != 2*time.Second || caller.Transport != rt {
				t.Fatal("caller client was modified")
			}
		})
	}

	t.Run("VERGEOS_INSECURE with a foreign transport", func(t *testing.T) {
		for _, order := range [][]string{{"http", "env"}, {"env", "http"}} {
			order := append([]string(nil), order...)
			t.Run(strings.Join(order, " then "), func(t *testing.T) {
				clearEnvVars()
				t.Cleanup(clearEnvVars)
				_ = os.Setenv("VERGEOS_INSECURE", "yes")
				rt := &foreignTransport{}
				caller := &http.Client{Transport: rt}
				opts := []ClientOption{
					WithBaseURL("https://verge.example"),
					WithAPIKey("key"),
				}
				for _, name := range order {
					if name == "http" {
						opts = append(opts, WithHTTPClient(caller))
					} else {
						opts = append(opts, WithEnvConfig())
					}
				}
				_, err := NewClient(opts...)
				if err == nil || !strings.Contains(err.Error(), "insecure TLS") {
					t.Fatalf("error = %v", err)
				}
				if rt.calls != 0 || caller.Transport != rt {
					t.Fatal("caller transport was used or replaced")
				}
			})
		}
	})

	t.Run("false after env keeps a foreign transport", func(t *testing.T) {
		clearEnvVars()
		t.Cleanup(clearEnvVars)
		_ = os.Setenv("VERGEOS_VERIFY_SSL", "false")
		rt := &foreignTransport{}
		caller := &http.Client{Transport: rt, Timeout: 4 * time.Second}
		client, err := NewClient(
			WithBaseURL("https://verge.example"),
			WithCredentials("admin", "correct-password"),
			WithHTTPClient(caller),
			WithEnvConfig(),
			WithInsecureTLS(false),
		)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if unwrapTransport(client.httpClient.Transport) != rt {
			t.Fatal("foreign transport was replaced")
		}
		if caller.Transport != rt || caller.Timeout != 4*time.Second {
			t.Fatal("caller client was modified")
		}
		if rt.calls != 2 {
			t.Fatalf("transport calls = %d, want 2", rt.calls)
		}
	})

	t.Run("env after false still rejects a foreign transport", func(t *testing.T) {
		clearEnvVars()
		t.Cleanup(clearEnvVars)
		_ = os.Setenv("VERGEOS_INSECURE", "1")
		rt := &foreignTransport{}
		caller := &http.Client{Transport: rt}
		_, err := NewClient(
			WithBaseURL("https://verge.example"),
			WithAPIKey("key"),
			WithInsecureTLS(false),
			WithHTTPClient(caller),
			WithEnvConfig(),
		)
		if err == nil || !strings.Contains(err.Error(), "insecure TLS") {
			t.Fatalf("error = %v", err)
		}
		if rt.calls != 0 || caller.Transport != rt {
			t.Fatal("caller transport was used or replaced")
		}
	})

	t.Run("WithInsecureTLS false allows a foreign transport", func(t *testing.T) {
		for _, order := range permutations([]string{"http", "insecure", "timeout", "rate"}) {
			order := append([]string(nil), order...)
			t.Run(strings.Join(order, " then "), func(t *testing.T) {
				rt := &foreignTransport{}
				caller := &http.Client{Transport: rt, Timeout: 9 * time.Second}
				opts := []ClientOption{
					WithBaseURL("https://verge.example"),
					WithCredentials("admin", "correct-password"),
				}
				wantTimeout := 9 * time.Second
				for _, name := range order {
					switch name {
					case "http":
						opts = append(opts, WithHTTPClient(caller))
					case "insecure":
						opts = append(opts, WithInsecureTLS(false))
					case "timeout":
						opts = append(opts, WithTimeout(6*time.Second))
						wantTimeout = 6 * time.Second
					case "rate":
						opts = append(opts, WithRateLimit(time.Millisecond))
					}
				}
				client, err := NewClient(opts...)
				if err != nil {
					t.Fatalf("NewClient: %v", err)
				}
				if client.httpClient.Timeout != wantTimeout {
					t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, wantTimeout)
				}
				inner := unwrapTransport(client.httpClient.Transport)
				if inner != rt {
					t.Fatalf("inner = %T, want the caller transport", inner)
				}
				retry := client.httpClient.Transport.(*retryTransport)
				limit, ok := retry.base.(*rateLimitTransport)
				if !ok || limit.base != rt || limit.interval != time.Millisecond {
					t.Fatal("rate limit was not layered on the caller transport")
				}
				if caller.Timeout != 9*time.Second || caller.Transport != rt {
					t.Fatal("caller client was modified")
				}
			})
		}
	})
}

func TestWithHTTPClientNil(t *testing.T) {
	_, err := NewClient(
		WithBaseURL("https://verge.example"),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(nil),
	)
	if err == nil || !strings.Contains(err.Error(), "http client is nil") {
		t.Fatalf("error = %v", err)
	}
}

// TestClonedTransportKeepsHTTP2Disabled checks that an empty TLSNextProto
// map still disables HTTP/2 after the transport is copied for insecure TLS.
// A nil map is what lets net/http enable HTTP/2 on the first request.
func TestClonedTransportKeepsHTTP2Disabled(t *testing.T) {
	server, _ := newStartupServer(t, "26.1.8", http.StatusOK, `[]`)
	const (
		callerTimeout   = 30 * time.Second
		explicitTimeout = 8 * time.Second
	)

	for _, order := range permutations([]string{"http", "insecure", "timeout"}) {
		order := append([]string(nil), order...)
		t.Run(strings.Join(order, " then "), func(t *testing.T) {
			disabled := map[string]func(string, *tls.Conn) http.RoundTripper{}
			base := &http.Transport{TLSNextProto: disabled}
			caller := &http.Client{Transport: base, Timeout: callerTimeout}
			opts := []ClientOption{
				WithBaseURL(server.URL),
				WithCredentials("admin", "correct-password"),
			}
			for _, name := range order {
				switch name {
				case "http":
					opts = append(opts, WithHTTPClient(caller))
				case "insecure":
					opts = append(opts, WithInsecureTLS(true))
				case "timeout":
					opts = append(opts, WithTimeout(explicitTimeout))
				}
			}

			client, err := NewClient(opts...)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			if caller.Transport != base || caller.Timeout != callerTimeout {
				t.Fatal("caller client was modified")
			}
			if base.TLSNextProto == nil || len(base.TLSNextProto) != 0 {
				t.Fatalf("caller TLSNextProto = %v, want an empty map", base.TLSNextProto)
			}

			got, ok := unwrapTransport(client.httpClient.Transport).(*http.Transport)
			if !ok {
				t.Fatalf("transport = %T", unwrapTransport(client.httpClient.Transport))
			}
			if got == base {
				t.Fatal("insecure TLS reused the caller's transport")
			}
			if got.TLSNextProto == nil {
				t.Fatal("clone lost TLSNextProto, so the first request can re-enable HTTP/2")
			}
			if _, h2 := got.TLSNextProto["h2"]; h2 || len(got.TLSNextProto) != 0 {
				t.Fatalf("clone TLSNextProto = %#v, want an empty map with HTTP/2 disabled", got.TLSNextProto)
			}
			if got.TLSClientConfig == nil || !got.TLSClientConfig.InsecureSkipVerify {
				t.Fatal("InsecureSkipVerify was not set on the clone")
			}
			if client.httpClient.Timeout != explicitTimeout {
				t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, explicitTimeout)
			}

			got.TLSNextProto["h2"] = nil
			if _, shared := base.TLSNextProto["h2"]; shared {
				t.Fatal("clone shares the caller's TLSNextProto map")
			}
		})
	}

	t.Run("copies TLSNextProto entries", func(t *testing.T) {
		var called bool
		next := map[string]func(string, *tls.Conn) http.RoundTripper{
			"custom": func(string, *tls.Conn) http.RoundTripper {
				called = true
				return nil
			},
		}
		base := &http.Transport{TLSNextProto: next}
		caller := &http.Client{Transport: base, Timeout: callerTimeout}
		client, err := NewClient(
			WithBaseURL(server.URL),
			WithCredentials("admin", "correct-password"),
			WithTimeout(explicitTimeout),
			WithHTTPClient(caller),
			WithInsecureTLS(true),
		)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		got := unwrapTransport(client.httpClient.Transport).(*http.Transport)
		fn := got.TLSNextProto["custom"]
		if fn == nil || len(got.TLSNextProto) != 1 {
			t.Fatalf("clone TLSNextProto = %#v, want the custom entry", got.TLSNextProto)
		}
		fn("", nil)
		if !called {
			t.Fatal("cloned TLSNextProto entry was not the caller's function")
		}
		if _, ok := base.TLSNextProto["h2"]; ok || len(base.TLSNextProto) != 1 {
			t.Fatal("caller TLSNextProto was modified")
		}
	})
}
