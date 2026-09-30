package vergeos

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func fastRetry(maxAttempts int) ClientOption {
	if maxAttempts <= 0 {
		maxAttempts = defaultRetryMaxAttempts
	}
	return WithRetry(RetryPolicy{
		MaxAttempts:    maxAttempts,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
	})
}

func newPolicyClient(t *testing.T, base http.RoundTripper, opts ...ClientOption) *Client {
	t.Helper()
	c := &Client{
		baseURL:  "https://vergeos.example.test",
		username: "user",
		password: "pass",
		httpClient: &http.Client{
			Transport: base,
			Timeout:   5 * time.Second,
		},
		userAgent: "govergeos-test",
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			t.Fatalf("option: %v", err)
		}
	}
	c.applyTransportPolicy()
	return c
}

type roundTripOutcome struct {
	status int
	err    error
	body   string
}

type scriptedTransport struct {
	mu       sync.Mutex
	calls    int
	methods  []string
	bodies   []string
	paths    []string
	outcomes []roundTripOutcome
}

func (s *scriptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.methods = append(s.methods, req.Method)
	s.bodies = append(s.bodies, string(body))
	if req.URL != nil {
		s.paths = append(s.paths, req.URL.Path)
	}
	outcome := s.outcomes[len(s.outcomes)-1]
	if s.calls-1 < len(s.outcomes) {
		outcome = s.outcomes[s.calls-1]
	}
	if outcome.err != nil {
		return nil, outcome.err
	}
	status := outcome.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(outcome.body)),
		Request:    req,
	}, nil
}

func (s *scriptedTransport) snapshot() (calls int, methods, bodies, paths []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls, append([]string(nil), s.methods...), append([]string(nil), s.bodies...), append([]string(nil), s.paths...)
}

func resetError() error {
	return &url.Error{
		Op:  "Put",
		URL: "https://vergeos.example.test/api/v4/cloud_snapshots/4",
		Err: &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET},
	}
}

func timeoutNetError() error {
	return &url.Error{
		Op:  "Get",
		URL: "https://vergeos.example.test/api/v4/clusters",
		Err: &net.OpError{Op: "dial", Net: "tcp", Err: timeoutErr{}},
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func okBody() roundTripOutcome {
	return roundTripOutcome{status: http.StatusOK, body: "[]"}
}

func TestRetryIdempotentAndNotIdempotent(t *testing.T) {
	resetThenOK := []roundTripOutcome{{err: resetError()}, okBody()}
	timeoutThenOK := []roundTripOutcome{{err: timeoutNetError()}, okBody()}
	eofThenOK := []roundTripOutcome{{err: io.EOF}, okBody()}
	unexpectedEOFThenOK := []roundTripOutcome{{err: io.ErrUnexpectedEOF}, okBody()}
	textResetThenOK := []roundTripOutcome{{
		err: errors.New("read tcp 127.0.0.1:1->127.0.0.1:443: read: connection reset by peer"),
	}, okBody()}
	statusThenOK := func(status int) []roundTripOutcome {
		return []roundTripOutcome{{status: status, body: `{"err":"unavailable"}`}, okBody()}
	}
	unauthorized := []roundTripOutcome{{status: http.StatusUnauthorized, body: `{"err":"Login required"}`}}

	tests := []struct {
		name        string
		method      string
		outcomes    []roundTripOutcome
		maxAttempts int
		wantCalls   int
		wantErr     bool
		wantAuth    bool
		wantStatus  int
		wantReset   bool
		wantBody    string
	}{
		{name: "GET connection reset", method: http.MethodGet, outcomes: resetThenOK, wantCalls: 2},
		{name: "PUT connection reset", method: http.MethodPut, outcomes: resetThenOK, wantCalls: 2, wantBody: `"description":"updated"`},
		{name: "DELETE connection reset", method: http.MethodDelete, outcomes: resetThenOK, wantCalls: 2},
		{name: "GET timeout before response", method: http.MethodGet, outcomes: timeoutThenOK, wantCalls: 2},
		{name: "PUT timeout before response", method: http.MethodPut, outcomes: timeoutThenOK, wantCalls: 2, wantBody: `"description":"updated"`},
		{name: "DELETE timeout before response", method: http.MethodDelete, outcomes: timeoutThenOK, wantCalls: 2},
		{name: "GET EOF", method: http.MethodGet, outcomes: eofThenOK, wantCalls: 2},
		{name: "PUT unexpected EOF", method: http.MethodPut, outcomes: unexpectedEOFThenOK, wantCalls: 2, wantBody: `"description":"updated"`},
		{name: "GET reset text", method: http.MethodGet, outcomes: textResetThenOK, wantCalls: 2},
		{name: "GET 429", method: http.MethodGet, outcomes: statusThenOK(http.StatusTooManyRequests), wantCalls: 2},
		{name: "PUT 429", method: http.MethodPut, outcomes: statusThenOK(http.StatusTooManyRequests), wantCalls: 2, wantBody: `"description":"updated"`},
		{name: "DELETE 429", method: http.MethodDelete, outcomes: statusThenOK(http.StatusTooManyRequests), wantCalls: 2},
		{name: "GET 502", method: http.MethodGet, outcomes: statusThenOK(http.StatusBadGateway), wantCalls: 2},
		{name: "PUT 502", method: http.MethodPut, outcomes: statusThenOK(http.StatusBadGateway), wantCalls: 2, wantBody: `"description":"updated"`},
		{name: "DELETE 502", method: http.MethodDelete, outcomes: statusThenOK(http.StatusBadGateway), wantCalls: 2},
		{name: "GET 503", method: http.MethodGet, outcomes: statusThenOK(http.StatusServiceUnavailable), wantCalls: 2},
		{name: "PUT 503", method: http.MethodPut, outcomes: statusThenOK(http.StatusServiceUnavailable), wantCalls: 2, wantBody: `"description":"updated"`},
		{name: "DELETE 503", method: http.MethodDelete, outcomes: statusThenOK(http.StatusServiceUnavailable), wantCalls: 2},
		{name: "GET 401", method: http.MethodGet, outcomes: unauthorized, wantCalls: 1, wantErr: true, wantAuth: true},
		{name: "PUT 401", method: http.MethodPut, outcomes: unauthorized, wantCalls: 1, wantErr: true, wantAuth: true, wantBody: `"description":"updated"`},
		{name: "DELETE 401", method: http.MethodDelete, outcomes: unauthorized, wantCalls: 1, wantErr: true, wantAuth: true},
		{name: "GET 500", method: http.MethodGet, outcomes: []roundTripOutcome{{status: http.StatusInternalServerError, body: `{"err":"unavailable"}`}}, wantCalls: 1, wantErr: true, wantStatus: http.StatusInternalServerError},
		{name: "GET 400", method: http.MethodGet, outcomes: []roundTripOutcome{{status: http.StatusBadRequest, body: `{"err":"bad"}}`}}, wantCalls: 1, wantErr: true, wantStatus: http.StatusBadRequest},
		{name: "GET 403", method: http.MethodGet, outcomes: []roundTripOutcome{{status: http.StatusForbidden, body: `{"err":"denied"}`}}, wantCalls: 1, wantErr: true, wantStatus: http.StatusForbidden},
		{name: "GET 404", method: http.MethodGet, outcomes: []roundTripOutcome{{status: http.StatusNotFound, body: `{"err":"missing"}`}}, wantCalls: 1, wantErr: true, wantStatus: http.StatusNotFound},
		{name: "GET connection refused", method: http.MethodGet, outcomes: []roundTripOutcome{{err: &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}}}, wantCalls: 1, wantErr: true},
		{name: "GET lookup failure", method: http.MethodGet, outcomes: []roundTripOutcome{{err: errors.New("dial tcp: no such host")}}, wantCalls: 1, wantErr: true},
		{name: "GET deadline exceeded", method: http.MethodGet, outcomes: []roundTripOutcome{{err: &url.Error{Op: "Get", URL: "https://vergeos.example.test", Err: context.DeadlineExceeded}}}, wantCalls: 1, wantErr: true},
		{name: "POST connection reset", method: http.MethodPost, outcomes: []roundTripOutcome{{err: resetError()}, okBody()}, wantCalls: 1, wantErr: true, wantReset: true, wantBody: `"name":"new"`},
		{name: "POST timeout", method: http.MethodPost, outcomes: []roundTripOutcome{{err: timeoutNetError()}, okBody()}, wantCalls: 1, wantErr: true, wantBody: `"name":"new"`},
		{name: "POST EOF", method: http.MethodPost, outcomes: []roundTripOutcome{{err: io.EOF}, okBody()}, wantCalls: 1, wantErr: true, wantBody: `"name":"new"`},
		{name: "POST 429", method: http.MethodPost, outcomes: statusThenOK(http.StatusTooManyRequests), wantCalls: 1, wantErr: true, wantStatus: http.StatusTooManyRequests, wantBody: `"name":"new"`},
		{name: "POST 502", method: http.MethodPost, outcomes: statusThenOK(http.StatusBadGateway), wantCalls: 1, wantErr: true, wantStatus: http.StatusBadGateway, wantBody: `"name":"new"`},
		{name: "POST 503", method: http.MethodPost, outcomes: statusThenOK(http.StatusServiceUnavailable), wantCalls: 1, wantErr: true, wantStatus: http.StatusServiceUnavailable, wantBody: `"name":"new"`},
		{name: "POST 401", method: http.MethodPost, outcomes: unauthorized, wantCalls: 1, wantErr: true, wantAuth: true, wantBody: `"name":"new"`},
		{name: "POST success", method: http.MethodPost, outcomes: []roundTripOutcome{okBody()}, wantCalls: 1, wantBody: `"name":"new"`},
		{name: "GET reset disabled", method: http.MethodGet, outcomes: resetThenOK, maxAttempts: 1, wantCalls: 1, wantErr: true, wantReset: true},
		{name: "GET 503 disabled", method: http.MethodGet, outcomes: statusThenOK(http.StatusServiceUnavailable), maxAttempts: 1, wantCalls: 1, wantErr: true, wantStatus: http.StatusServiceUnavailable},
		{
			name:      "GET reset exhausted",
			method:    http.MethodGet,
			outcomes:  []roundTripOutcome{{err: resetError()}, {err: resetError()}, {err: resetError()}},
			wantCalls: 3,
			wantErr:   true,
			wantReset: true,
		},
		{
			name:       "GET 503 exhausted",
			method:     http.MethodGet,
			outcomes:   []roundTripOutcome{{status: http.StatusServiceUnavailable, body: `{"err":"unavailable"}`}},
			wantCalls:  3,
			wantErr:    true,
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &scriptedTransport{outcomes: tt.outcomes}
			maxAttempts := tt.maxAttempts
			if maxAttempts == 0 {
				maxAttempts = 3
			}
			client := newPolicyClient(t, rt, fastRetry(maxAttempts))

			err := callMethod(client, tt.method)
			calls, methods, bodies, _ := rt.snapshot()
			if calls != tt.wantCalls {
				t.Fatalf("calls = %d, want %d", calls, tt.wantCalls)
			}
			for _, method := range methods {
				if method != tt.method {
					t.Fatalf("method = %s, want %s", method, tt.method)
				}
			}
			if tt.wantBody != "" {
				if len(bodies) != tt.wantCalls {
					t.Fatalf("bodies = %d, want %d", len(bodies), tt.wantCalls)
				}
				for i, body := range bodies {
					if !strings.Contains(body, tt.wantBody) {
						t.Fatalf("attempt %d body = %s, want it to contain %s", i+1, body, tt.wantBody)
					}
				}
			}
			if tt.method == http.MethodGet || tt.method == http.MethodDelete {
				for _, body := range bodies {
					if body != "" {
						t.Fatalf("body = %q, want empty", body)
					}
				}
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if IsAuthError(err) != tt.wantAuth {
				t.Fatalf("IsAuthError = %v, want %v (%v)", IsAuthError(err), tt.wantAuth, err)
			}
			if tt.wantReset && !errors.Is(err, syscall.ECONNRESET) && (err == nil || !strings.Contains(err.Error(), "connection reset")) {
				t.Fatalf("error = %v, want connection reset", err)
			}
			if tt.wantStatus != 0 {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("error = %v, want APIError", err)
				}
				if apiErr.StatusCode != tt.wantStatus {
					t.Fatalf("status = %d, want %d", apiErr.StatusCode, tt.wantStatus)
				}
			}
		})
	}
}

func callMethod(c *Client, method string) error {
	ctx := context.Background()
	switch method {
	case http.MethodGet:
		return c.get(ctx, "/groups", nil, nil)
	case http.MethodPut:
		return c.put(ctx, "/groups/1", map[string]string{"description": "updated"}, nil)
	case http.MethodPost:
		return c.post(ctx, "/groups", map[string]string{"name": "new"}, nil)
	case http.MethodDelete:
		return c.delete(ctx, "/groups/1")
	default:
		return errors.New("unsupported method")
	}
}

func TestRetryZeroMaxAttemptsUsesDefaultCap(t *testing.T) {
	rt := &scriptedTransport{outcomes: []roundTripOutcome{{err: resetError()}}}
	client := newPolicyClient(t, rt, WithRetry(RetryPolicy{
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
	}))

	err := client.get(context.Background(), "/groups", nil, nil)
	calls, _, _, _ := rt.snapshot()
	if calls != defaultRetryMaxAttempts {
		t.Fatalf("calls = %d, want default %d", calls, defaultRetryMaxAttempts)
	}
	if !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("error = %v, want connection reset", err)
	}
}

func TestDefaultRetryCapAndBackoff(t *testing.T) {
	rt := &scriptedTransport{outcomes: []roundTripOutcome{{status: http.StatusServiceUnavailable, body: `{"err":"unavailable"}`}}}
	client := newPolicyClient(t, rt)

	start := time.Now()
	err := client.get(context.Background(), "/groups", nil, nil)
	elapsed := time.Since(start)

	calls, _, _, _ := rt.snapshot()
	if calls != defaultRetryMaxAttempts {
		t.Fatalf("calls = %d, want %d", calls, defaultRetryMaxAttempts)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("error = %v, want APIError 503", err)
	}
	// Two sleeps, each at least half of the exponential delay: 50ms + 100ms.
	if elapsed < 100*time.Millisecond {
		t.Fatalf("elapsed = %s, want at least 100ms of backoff", elapsed)
	}
}

func TestRetryPolicyNormalization(t *testing.T) {
	p := (RetryPolicy{}).normalized()
	if p.MaxAttempts != defaultRetryMaxAttempts || p.InitialBackoff != defaultRetryInitialBackoff || p.MaxBackoff != defaultRetryMaxBackoff {
		t.Fatalf("zero policy = %+v", p)
	}
	disabled := (RetryPolicy{MaxAttempts: 1}).normalized()
	if disabled.MaxAttempts != 1 {
		t.Fatalf("MaxAttempts = %d, want 1", disabled.MaxAttempts)
	}
	raised := (RetryPolicy{InitialBackoff: 5 * time.Second, MaxBackoff: time.Millisecond}).normalized()
	if raised.MaxBackoff != 5*time.Second {
		t.Fatalf("MaxBackoff = %s, want 5s", raised.MaxBackoff)
	}

	for i := 0; i < 40; i++ {
		first := p.delayBeforeRetry(1)
		if first < p.InitialBackoff/2 || first > p.InitialBackoff {
			t.Fatalf("first delay = %s, want [%s, %s]", first, p.InitialBackoff/2, p.InitialBackoff)
		}
		capped := p.delayBeforeRetry(8)
		if capped < p.MaxBackoff/2 || capped > p.MaxBackoff {
			t.Fatalf("capped delay = %s, want [%s, %s]", capped, p.MaxBackoff/2, p.MaxBackoff)
		}
	}
}

func TestRetryStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rt := &scriptedTransport{outcomes: []roundTripOutcome{{err: resetError()}, okBody()}}
	client := newPolicyClient(t, rt, WithRetry(RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 5 * time.Second,
		MaxBackoff:     5 * time.Second,
	}))

	go func() {
		for {
			calls, _, _, _ := rt.snapshot()
			if calls >= 1 {
				break
			}
			time.Sleep(time.Millisecond)
		}
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := client.get(ctx, "/groups", nil, nil)
	elapsed := time.Since(start)
	calls, _, _, _ := rt.snapshot()
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("elapsed = %s, want the backoff to be interrupted", elapsed)
	}
}

func TestRetryDoesNotStartWhenContextAlreadyCanceled(t *testing.T) {
	rt := &scriptedTransport{outcomes: []roundTripOutcome{okBody()}}
	client := newPolicyClient(t, rt, fastRetry(3))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.get(ctx, "/groups", nil, nil)
	calls, _, _, _ := rt.snapshot()
	if calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRetrySkipsUnreplayableBody(t *testing.T) {
	rt := &scriptedTransport{outcomes: []roundTripOutcome{
		{status: http.StatusBadGateway, body: `{"err":"bad gateway"}`},
		okBody(),
	}}
	client := newPolicyClient(t, rt, fastRetry(3))
	req, err := http.NewRequest(http.MethodPut, client.baseURL+"/api/v4/groups/1", io.NopCloser(strings.NewReader(`{"name":"x"}`)))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetBody != nil {
		t.Fatal("expected a body without GetBody")
	}
	resp, err := client.httpClient.Transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
	calls, _, _, _ := rt.snapshot()
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestCloudSnapshotUpdateRetriesConnectionReset(t *testing.T) {
	rt := &scriptedTransport{outcomes: []roundTripOutcome{
		{err: resetError()},
		{status: http.StatusOK},
		{status: http.StatusOK, body: `{"$key":4,"name":"snap","description":"updated"}`},
	}}
	client := newPolicyClient(t, rt, fastRetry(3))
	initServices(client)

	description := "updated"
	snap, err := client.CloudSnapshots.Update(context.Background(), 4, &CloudSnapshotUpdateRequest{
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if snap.Name != "snap" || snap.Description != "updated" {
		t.Fatalf("snapshot = %+v", snap)
	}
	calls, methods, bodies, paths := rt.snapshot()
	if calls != 3 {
		t.Fatalf("calls = %d, want PUT, PUT, GET", calls)
	}
	if methods[0] != http.MethodPut || methods[1] != http.MethodPut || methods[2] != http.MethodGet {
		t.Fatalf("methods = %v", methods)
	}
	if !strings.Contains(bodies[0], `"description":"updated"`) || bodies[0] != bodies[1] {
		t.Fatalf("PUT bodies = %q and %q", bodies[0], bodies[1])
	}
	for _, path := range paths[:2] {
		if path != "/api/v4/cloud_snapshots/4" {
			t.Fatalf("PUT path = %s", path)
		}
	}
}

func TestCloudSnapshotCreateDoesNotRetryConnectionReset(t *testing.T) {
	rt := &scriptedTransport{outcomes: []roundTripOutcome{{err: resetError()}, okBody()}}
	client := newPolicyClient(t, rt, fastRetry(4))
	initServices(client)

	_, err := client.CloudSnapshots.Create(context.Background(), &CloudSnapshotCreateRequest{Name: "snap"})
	if !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("error = %v, want connection reset", err)
	}
	calls, methods, _, paths := rt.snapshot()
	if calls != 1 || methods[0] != http.MethodPost {
		t.Fatalf("calls = %d methods = %v, want one POST", calls, methods)
	}
	if !strings.Contains(paths[0], "/cloud_snapshots") {
		t.Fatalf("path = %s", paths[0])
	}
}

func TestGetAbsoluteRetriesServiceUnavailableButNot401(t *testing.T) {
	t.Run("503 then 200", func(t *testing.T) {
		rt := &scriptedTransport{outcomes: []roundTripOutcome{
			{status: http.StatusServiceUnavailable, body: "unavailable"},
			{status: http.StatusOK, body: `{"version":"26.1.8"}`},
		}}
		client := newPolicyClient(t, rt, fastRetry(3))
		var resp versionResponse
		if err := client.getAbsolute(context.Background(), "/version.json", nil, &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Version != "26.1.8" {
			t.Fatalf("version = %q", resp.Version)
		}
		calls, methods, _, paths := rt.snapshot()
		if calls != 2 || methods[0] != http.MethodGet || paths[0] != "/version.json" {
			t.Fatalf("calls=%d methods=%v paths=%v", calls, methods, paths)
		}
	})

	t.Run("401 once", func(t *testing.T) {
		rt := &scriptedTransport{outcomes: []roundTripOutcome{{
			status: http.StatusUnauthorized,
			body:   `{"err":"Login required"}`,
		}}}
		client := newPolicyClient(t, rt, fastRetry(5))
		err := client.getAbsolute(context.Background(), "/version.json", nil, &versionResponse{})
		if !IsAuthError(err) {
			t.Fatalf("error = %v, want AuthError", err)
		}
		calls, _, _, _ := rt.snapshot()
		if calls != 1 {
			t.Fatalf("calls = %d, want 1", calls)
		}
	})
}

func TestRetryRealConnectionReset(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var calls atomic.Int32
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		for {
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				return
			}
			n := calls.Add(1)
			if n == 1 {
				if tc, ok := conn.(*net.TCPConn); ok {
					_ = tc.SetLinger(0)
				}
				_ = conn.Close()
				continue
			}
			_, _ = io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\n[]")
			_ = conn.Close()
		}
	}()

	client := newPolicyClient(t, &http.Transport{DisableKeepAlives: true}, fastRetry(3))
	client.baseURL = "http://" + ln.Addr().String()

	err = client.get(context.Background(), "/groups", nil, nil)
	_ = ln.Close()
	<-serveDone
	if err != nil {
		t.Fatalf("GET after connection reset: %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("accepted connections = %d, want 2", got)
	}
}

func TestRateLimitSpacesRequests(t *testing.T) {
	const interval = 80 * time.Millisecond
	rec := &recordingTransport{}
	client := newPolicyClient(t, rec, WithRateLimit(interval), fastRetry(1))

	start := time.Now()
	if err := client.get(context.Background(), "/groups", nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := client.get(context.Background(), "/groups", nil, nil); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < interval-15*time.Millisecond {
		t.Fatalf("elapsed = %s, want at least %s", elapsed, interval)
	}
	rec.mu.Lock()
	starts := append([]time.Time(nil), rec.starts...)
	rec.mu.Unlock()
	if len(starts) != 2 {
		t.Fatalf("calls = %d, want 2", len(starts))
	}
	if gap := starts[1].Sub(starts[0]); gap < interval-15*time.Millisecond {
		t.Fatalf("gap = %s, want at least %s", gap, interval)
	}
}

func TestRateLimitSpacesConcurrentRequests(t *testing.T) {
	const interval = 40 * time.Millisecond
	const n = 4
	rec := &recordingTransport{}
	client := newPolicyClient(t, rec, WithRateLimit(interval), fastRetry(1))

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if err := client.get(context.Background(), "/groups", nil, nil); err != nil {
				t.Errorf("get: %v", err)
			}
		}()
	}
	wg.Wait()

	rec.mu.Lock()
	starts := append([]time.Time(nil), rec.starts...)
	rec.mu.Unlock()
	if len(starts) != n {
		t.Fatalf("calls = %d, want %d", len(starts), n)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i].Before(starts[j]) })
	for i := 1; i < len(starts); i++ {
		if gap := starts[i].Sub(starts[i-1]); gap < interval-15*time.Millisecond {
			t.Fatalf("gap %d = %s, want at least %s", i, gap, interval)
		}
	}
}

func TestRateLimitHonorsCanceledContext(t *testing.T) {
	rec := &recordingTransport{}
	client := newPolicyClient(t, rec, WithRateLimit(5*time.Second), fastRetry(1))
	if err := client.get(context.Background(), "/groups", nil, nil); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := client.get(ctx, "/groups", nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("canceled request waited %s", time.Since(start))
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.starts) != 1 {
		t.Fatalf("calls = %d, want 1", len(rec.starts))
	}
}

func TestRateLimitNegativeIntervalIsOff(t *testing.T) {
	client := newPolicyClient(t, &http.Transport{}, WithRateLimit(-1*time.Second), fastRetry(2))
	retry, ok := client.httpClient.Transport.(*retryTransport)
	if !ok {
		t.Fatalf("transport = %T", client.httpClient.Transport)
	}
	if _, limited := retry.base.(*rateLimitTransport); limited {
		t.Fatal("negative interval installed a rate limit")
	}
}

type recordingTransport struct {
	mu     sync.Mutex
	starts []time.Time
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.starts = append(r.starts, time.Now())
	r.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("[]")),
		Request:    req,
	}, nil
}

func TestTransportPolicyPreservesCallerAndTLS(t *testing.T) {
	base := &http.Transport{}
	caller := &http.Client{Transport: base, Timeout: time.Second}
	client := &Client{
		baseURL:    "https://vergeos.example.test",
		username:   "user",
		password:   "pass",
		httpClient: caller,
		userAgent:  "govergeos-test",
	}
	if err := WithInsecureTLS(true)(client); err != nil {
		t.Fatal(err)
	}
	if err := WithRateLimit(time.Millisecond)(client); err != nil {
		t.Fatal(err)
	}
	if err := WithRetry(RetryPolicy{MaxAttempts: 2})(client); err != nil {
		t.Fatal(err)
	}
	client.applyTransportPolicy()

	if caller.Transport == client.httpClient.Transport {
		t.Fatal("applyTransportPolicy replaced the caller's Transport")
	}
	if client.httpClient == caller {
		t.Fatal("applyTransportPolicy reused the caller's client")
	}
	if client.httpClient.Timeout != time.Second {
		t.Fatalf("timeout = %s", client.httpClient.Timeout)
	}

	retry, ok := client.httpClient.Transport.(*retryTransport)
	if !ok {
		t.Fatalf("transport = %T", client.httpClient.Transport)
	}
	if retry.policy.MaxAttempts != 2 {
		t.Fatalf("MaxAttempts = %d, want 2", retry.policy.MaxAttempts)
	}
	limit, ok := retry.base.(*rateLimitTransport)
	if !ok {
		t.Fatalf("inner = %T", retry.base)
	}
	got, ok := limit.base.(*http.Transport)
	if !ok {
		t.Fatalf("base = %T", limit.base)
	}
	if got != caller.Transport {
		t.Fatal("rate limit is not wrapping the caller's transport")
	}
	if got == base {
		t.Fatal("WithInsecureTLS did not install its own transport")
	}
	if got.TLSClientConfig == nil || !got.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify was not preserved")
	}
	client.httpClient.CloseIdleConnections()
}

func TestNewClientRetriesTransientStartupGets(t *testing.T) {
	var versionCalls, credCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version.json":
			versionCalls++
			if versionCalls == 1 {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			jsonResponse(w, http.StatusOK, versionResponse{Version: "26.1.8"})
		case apiBasePath + credentialCheckEndpoint:
			credCalls++
			if credCalls == 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(`{"err":"bad gateway"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(server.Client()),
		fastRetry(3),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.serverVersion != "26.1.8" {
		t.Fatalf("serverVersion = %q", client.serverVersion)
	}
	if versionCalls != 2 {
		t.Fatalf("version requests = %d, want 2", versionCalls)
	}
	if credCalls != 2 {
		t.Fatalf("credential requests = %d, want 2", credCalls)
	}
	if server.Client().Transport == client.httpClient.Transport {
		t.Fatal("NewClient replaced the httptest client's Transport")
	}
}

func TestNewClientDoesNotRetry401(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusUnauthorized, `{"err":"Login required"}`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "wrong-password"),
		WithHTTPClient(server.Client()),
		WithRetry(RetryPolicy{MaxAttempts: 5, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond}),
	)
	if client != nil {
		t.Fatal("NewClient returned a client for a rejected password")
	}
	if !IsAuthError(err) {
		t.Fatalf("error = %v, want AuthError", err)
	}
	if len(credentialChecks(*seen)) != 1 {
		t.Fatalf("credential checks = %d, want 1: %#v", len(credentialChecks(*seen)), *seen)
	}
}

func TestNewClientDisabledRetryDoesNotRepeat503(t *testing.T) {
	server, seen := newStartupServer(t, "26.1.8", http.StatusServiceUnavailable, `{"err":"unavailable"}`)

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithCredentials("admin", "correct-password"),
		WithHTTPClient(server.Client()),
		WithRetry(RetryPolicy{MaxAttempts: 1}),
	)
	if client != nil {
		t.Fatal("NewClient returned a client")
	}
	if IsAuthError(err) {
		t.Fatalf("503 reported as AuthError: %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("error = %v, want APIError 503", err)
	}
	if len(credentialChecks(*seen)) != 1 {
		t.Fatalf("credential checks = %d, want 1", len(credentialChecks(*seen)))
	}
}
