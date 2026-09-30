package vergeos

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// defaultRetryMaxAttempts is the total number of tries, including the first.
	defaultRetryMaxAttempts = 3
	// defaultRetryInitialBackoff is the base delay before the second attempt.
	defaultRetryInitialBackoff = 100 * time.Millisecond
	// defaultRetryMaxBackoff caps a single delay.
	defaultRetryMaxBackoff = 2 * time.Second
)

// RetryPolicy controls retries of idempotent API requests.
//
// GET, PUT, and DELETE are retried when the connection is reset, when the
// peer closes it before a response (including EOF), when the client times
// out before any response, or when the server returns 429, 502, or 503.
// POST is not retried. Creates and actions are not idempotent, and a POST
// that fails with no response is returned to the caller.
//
// HTTP 401 is never retried. A repeated failed login locks the account.
//
// The zero value, and a value passed to WithRetry with MaxAttempts left at
// 0, uses these defaults: 3 attempts, 100ms initial backoff, doubling up
// to 2s, with equal jitter. MaxAttempts of 1 disables retries. The
// client's Timeout is the budget for the whole call, including sleeps and
// later attempts, so a timeout that has already cancelled the request is
// not retried.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts, including the first.
	// Zero uses the default of 3. 1 disables retries.
	MaxAttempts int
	// InitialBackoff is the base delay before the second attempt.
	// Zero uses the default of 100ms. Each later delay doubles until
	// MaxBackoff, then equal jitter is applied.
	InitialBackoff time.Duration
	// MaxBackoff caps one delay. Zero uses the default of 2s.
	// A value below InitialBackoff is raised to InitialBackoff.
	MaxBackoff time.Duration
}

// WithRetry sets the retry policy.
//
// The default policy is already enabled. Pass RetryPolicy{MaxAttempts: 1}
// to disable retries. A zero MaxAttempts, InitialBackoff, or MaxBackoff
// keeps that field's default.
func WithRetry(policy RetryPolicy) ClientOption {
	return func(c *Client) error {
		c.retryPolicy = policy
		c.retryConfigured = true
		return nil
	}
}

// WithRateLimit spaces the start of each request by at least interval.
//
// The spacing applies to every attempt, including retries. The client does
// not rate-limit unless this option is set. A zero or negative interval
// leaves spacing off.
//
// VergeOS drops connections when a session exceeds its webserver API rate
// limit (50 requests by default) instead of returning 429. An interval
// such as 50ms keeps a burst under that limit.
func WithRateLimit(interval time.Duration) ClientOption {
	return func(c *Client) error {
		if interval < 0 {
			interval = 0
		}
		c.rateLimit = interval
		return nil
	}
}

func (p RetryPolicy) normalized() RetryPolicy {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = defaultRetryMaxAttempts
	}
	if p.InitialBackoff <= 0 {
		p.InitialBackoff = defaultRetryInitialBackoff
	}
	if p.MaxBackoff <= 0 {
		p.MaxBackoff = defaultRetryMaxBackoff
	}
	if p.MaxBackoff < p.InitialBackoff {
		p.MaxBackoff = p.InitialBackoff
	}
	return p
}

func defaultRetryPolicy() RetryPolicy {
	return (RetryPolicy{}).normalized()
}

// applyTransportPolicy installs retry and optional rate limiting.
//
// The caller's *http.Client is copied first so its Transport field is not
// replaced. The copy shares the underlying RoundTripper, which keeps the
// connection pool, then the copy's Transport is wrapped.
func (c *Client) applyTransportPolicy() {
	if c.transportReady || c.httpClient == nil {
		return
	}
	c.transportReady = true

	cloned := *c.httpClient
	c.httpClient = &cloned

	policy := defaultRetryPolicy()
	if c.retryConfigured {
		policy = c.retryPolicy.normalized()
	}

	base := c.httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	if c.rateLimit > 0 {
		base = &rateLimitTransport{base: base, interval: c.rateLimit}
	}
	c.httpClient.Transport = &retryTransport{base: base, policy: policy}
}

type retryTransport struct {
	base   http.RoundTripper
	policy RetryPolicy
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	attempts := t.policy.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	if !retryableMethod(req.Method) {
		attempts = 1
	}

	var resp *http.Response
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if attempt > 1 {
			if sleepErr := sleepContext(req.Context(), t.policy.delayBeforeRetry(attempt-1)); sleepErr != nil {
				return nil, sleepErr
			}
		}

		attemptReq, reqErr := requestAttempt(req, attempt)
		if reqErr != nil {
			if err != nil {
				return nil, err
			}
			if resp != nil {
				return resp, nil
			}
			return nil, reqErr
		}

		resp, err = t.base.RoundTrip(attemptReq)
		if !shouldRetry(req, resp, err, attempt, attempts) {
			return resp, err
		}
		closeResponse(resp)
		resp = nil
	}
	return resp, err
}

func (t *retryTransport) CloseIdleConnections() {
	closeIdle(t.base)
}

type rateLimitTransport struct {
	base     http.RoundTripper
	interval time.Duration
	mu       sync.Mutex
	next     time.Time
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.wait(req.Context()); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

func (t *rateLimitTransport) CloseIdleConnections() {
	closeIdle(t.base)
}

func (t *rateLimitTransport) wait(ctx context.Context) error {
	t.mu.Lock()
	now := time.Now()
	start := t.next
	if !start.After(now) {
		start = now
	}
	t.next = start.Add(t.interval)
	t.mu.Unlock()

	delay := time.Until(start)
	if delay <= 0 {
		return nil
	}
	return sleepContext(ctx, delay)
}

func retryableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// retryableStatus reports statuses that are safe to repeat for an
// idempotent method. 401 is absent on purpose: another attempt with the
// same rejected credentials counts toward account lockout.
func retryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable:
		return true
	default:
		return false
	}
}

// retryableTransportError reports a failure that happened before a
// response was returned. Connection resets and EOFs are the two ways a
// VergeOS node drops a session that has exceeded its API rate limit.
func retryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "server closed idle connection")
}

func shouldRetry(req *http.Request, resp *http.Response, err error, attempt, attempts int) bool {
	if attempt >= attempts || !retryableMethod(req.Method) || !canReplay(req) {
		return false
	}
	if req.Context().Err() != nil {
		return false
	}
	if err != nil {
		return retryableTransportError(err)
	}
	return resp != nil && retryableStatus(resp.StatusCode)
}

func canReplay(req *http.Request) bool {
	if req.Body == nil || req.Body == http.NoBody {
		return true
	}
	return req.GetBody != nil
}

func requestAttempt(req *http.Request, attempt int) (*http.Request, error) {
	if attempt == 1 {
		return req, nil
	}
	clone := req.Clone(req.Context())
	if req.Body == nil || req.Body == http.NoBody {
		return clone, nil
	}
	if req.GetBody == nil {
		return nil, errors.New("vergeos: request body cannot be replayed")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	clone.Body = body
	return clone, nil
}

func (p RetryPolicy) delayBeforeRetry(failedAttempt int) time.Duration {
	if failedAttempt < 1 {
		failedAttempt = 1
	}
	backoff := p.InitialBackoff
	for i := 1; i < failedAttempt; i++ {
		if backoff > p.MaxBackoff/2 {
			backoff = p.MaxBackoff
			break
		}
		backoff *= 2
	}
	if backoff > p.MaxBackoff {
		backoff = p.MaxBackoff
	}
	return jitterDuration(backoff)
}

func jitterDuration(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	half := d / 2
	span := int64(d - half)
	if span <= 0 {
		return half
	}
	return half + time.Duration(rand.Int63n(span+1))
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func closeResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	_ = resp.Body.Close()
}

type closeIdler interface {
	CloseIdleConnections()
}

func closeIdle(rt http.RoundTripper) {
	if c, ok := rt.(closeIdler); ok {
		c.CloseIdleConnections()
	}
}
