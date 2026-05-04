package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultMaxResponseSize is the default 10 MiB cap applied to response bodies
// when MaxResponseSize is zero. A response that would exceed this limit is
// rejected with an explicit error rather than being silently truncated.
const DefaultMaxResponseSize int64 = 10 << 20

// redactedQueryKeys lists URL query parameter names whose values must be
// replaced with "***" before logging. WeChat protocols put credentials in
// the URL for several endpoints (OAuth code exchange, OA clear_quota, etc.)
// — without redaction these land verbatim in any HTTP debug log sink.
//
// Match is case-insensitive (RedactURL lower-cases the key before lookup).
var redactedQueryKeys = map[string]struct{}{
	"secret":              {},
	"appsecret":           {},
	"app_secret":          {},
	"component_appsecret": {},
	"suite_secret":        {},
	"code":                {},      // OAuth authorization code (single-use but credential-equivalent)
	"access_token":        {},      // short-lived but still a bearer credential
	"refresh_token":       {},
}

// RedactedValue is the placeholder substituted for sensitive query parameter
// values by RedactURL. Letters-only so URL encoding leaves it untouched and
// the log line stays human-readable.
const RedactedValue = "REDACTED"

// RedactURL returns rawURL with the values of any sensitive query parameters
// (secret, appsecret, code, access_token, refresh_token, etc.) replaced with
// RedactedValue. Use this when including a WeChat URL in user-visible logs
// or error messages.
//
// If rawURL cannot be parsed it is returned unchanged so the function is
// safe to apply blindly. The replacement is value-only — the key, parameter
// order, and other params are preserved as written.
func RedactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	if len(q) == 0 {
		return rawURL
	}
	changed := false
	for k := range q {
		if _, ok := redactedQueryKeys[strings.ToLower(k)]; ok {
			q.Set(k, RedactedValue)
			changed = true
		}
	}
	if !changed {
		return rawURL
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// Logger is an optional debug-logging interface. By default the SDK does not print any
// request/response content to avoid leaking sensitive data (OpenIDs, order numbers, amounts)
// in production. Callers may inject their own implementation via WithLogger — but see the
// PII warning on WithLogger before wiring a non-nop sink.
type Logger interface {
	Debugf(format string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Debugf(string, ...any) {}

// HTTP is a thin JSON-over-HTTP client used by all SDK packages.
// It stores a base URL and optional default headers so callers never
// repeat boilerplate. Safe for concurrent use after construction.
type HTTP struct {
	BaseURL string
	Client  *http.Client
	// Headers are the client-level default request headers. They are not written by the SDK
	// after NewHTTP returns, so concurrent use is safe. Per-request overrides go through the
	// headers parameter of DoRequest.
	Headers map[string]string
	Timeout time.Duration
	Logger  Logger
	// MaxResponseSize caps the number of bytes the SDK will read from a response body.
	// Zero means "use DefaultMaxResponseSize (10 MiB)". A negative value disables the cap.
	// A response that would exceed the cap is rejected with an explicit error to avoid
	// OOM from a misconfigured or hostile server.
	MaxResponseSize int64
}

// Option is a functional configuration applied to HTTP during NewHTTP.
type Option func(*HTTP)

// NewHTTP constructs an HTTP client rooted at baseURL.
// All relative paths passed to Get/Post/etc. are joined to baseURL.
// Default timeout is 30 s; override with WithTimeout.
func NewHTTP(baseURL string, opts ...Option) *HTTP {
	h := &HTTP{
		BaseURL: baseURL,
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
		Logger:  nopLogger{},
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.Client == nil {
		h.Client = &http.Client{Timeout: h.Timeout}
	}

	return h
}

// WithTimeout overrides the default 30-second per-request deadline.
func WithTimeout(timeout time.Duration) Option {
	return func(h *HTTP) {
		h.Timeout = timeout
	}
}

// WithHeaders merges the provided key-value pairs into the client's
// default header set. Existing keys are overwritten.
func WithHeaders(headers map[string]string) Option {
	return func(h *HTTP) {
		for k, v := range headers {
			h.Headers[k] = v
		}
	}
}

// WithLogger injects a debug logger. Pass nil to use the no-op default.
//
// Credential redaction: the SDK runs every URL through RedactURL before
// emitting it, replacing the values of secret / appsecret / code /
// access_token / refresh_token / etc. with "***". This protects against the
// most common credential-in-URL leaks (OAuth code-exchange endpoint, OA
// clear_quota, suite secret).
//
// PII warning: request and response BODIES are still logged verbatim. They
// can contain OpenIDs, payment amounts, encrypted bank data, and other
// sensitive fields — especially on refund, profit-sharing, and transfer
// endpoints. Audit your logger sink before enabling on a production
// workload.
func WithLogger(logger Logger) Option {
	return func(h *HTTP) {
		if logger != nil {
			h.Logger = logger
		}
	}
}

// WithMaxResponseSize overrides the default 10 MiB response-body cap.
// Pass a positive value in bytes to set a custom limit. Pass a negative
// value to disable the cap entirely (only do this for trusted endpoints
// that legitimately stream large payloads). Zero is treated as "use the
// default".
func WithMaxResponseSize(n int64) Option {
	return func(h *HTTP) {
		h.MaxResponseSize = n
	}
}

// WithHTTPClient injects a custom http.Client (e.g. with a proxy or custom connection pool).
// When provided, the Timeout option has no effect; set the timeout on the client directly.
func WithHTTPClient(c *http.Client) Option {
	return func(h *HTTP) {
		if c != nil {
			h.Client = c
		}
	}
}

// SetBaseURL replaces the base URL after construction. Not goroutine-safe;
// call only before the client is shared between goroutines.
func (h *HTTP) SetBaseURL(u string) {
	h.BaseURL = u
}

// readLimitedBody reads up to limit+1 bytes from body and returns an error if
// the result exceeds limit, so callers see an explicit "exceeds N bytes" error
// rather than silently-truncated JSON. limit==0 uses DefaultMaxResponseSize;
// limit<0 disables the cap.
func readLimitedBody(body io.Reader, limit int64) ([]byte, error) {
	if limit == 0 {
		limit = DefaultMaxResponseSize
	}
	if limit < 0 {
		buf, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("read response body failed: %w", err)
		}
		return buf, nil
	}
	buf, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}
	if int64(len(buf)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return buf, nil
}

// buildURL merges base + path + query into the final URL, handling cases where path
// already contains a query string.
func (h *HTTP) buildURL(path string, query url.Values) (string, error) {
	full := h.BaseURL + path
	if len(query) == 0 {
		return full, nil
	}
	u, err := url.Parse(full)
	if err != nil {
		return "", fmt.Errorf("parse url failed: %w", err)
	}
	q := u.Query()
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// HTTPError carries the status code, raw body, and response headers of a failed response.
type HTTPError struct {
	StatusCode int
	Body       string
	Header     http.Header
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("request failed with status code %d: %s", e.StatusCode, e.Body)
}

// DoRequest sends a single request. A nil body means no request body.
// headers are merged with the client defaults (per-request values take precedence).
// The caller is responsible for marshalling the body so that the signed body and the
// sent body are identical.
func (h *HTTP) DoRequest(
	ctx context.Context,
	method, path string,
	query url.Values,
	body []byte,
	headers http.Header,
	result any,
) error {
	_, _, respBody, err := h.doRaw(ctx, method, path, query, body, headers)
	if err != nil {
		return err
	}
	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
		}
	}
	return nil
}

// DoRequestWithRawResponse is like DoRequest but returns the raw response body and headers,
// primarily for scenarios that need to verify the response signature (e.g. WeChat Pay v3).
func (h *HTTP) DoRequestWithRawResponse(
	ctx context.Context,
	method, path string,
	query url.Values,
	body []byte,
	headers http.Header,
) (statusCode int, respHeader http.Header, respBody []byte, err error) {
	return h.doRaw(ctx, method, path, query, body, headers)
}

func (h *HTTP) doRaw(
	ctx context.Context,
	method, path string,
	query url.Values,
	body []byte,
	headers http.Header,
) (int, http.Header, []byte, error) {
	fullURL, err := h.buildURL(path, query)
	if err != nil {
		return 0, nil, nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}
	for k, vs := range headers {
		req.Header[k] = vs
	}
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	loggedURL := RedactURL(fullURL)
	h.Logger.Debugf("http %s %s body=%s", method, loggedURL, string(body))

	resp, err := h.Client.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := readLimitedBody(resp.Body, h.MaxResponseSize)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, err
	}

	h.Logger.Debugf("http %s %s status=%d response=%s",
		method, loggedURL, resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, resp.Header, respBody, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			Header:     resp.Header,
		}
	}
	return resp.StatusCode, resp.Header, respBody, nil
}

// do is the shared implementation for the convenience methods (Get/Post/Put/Patch/Delete).
// It auto-marshals the body to JSON.
func (h *HTTP) do(
	ctx context.Context,
	method, path string,
	body any,
	query url.Values,
	result any,
) error {
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
	}
	return h.DoRequest(ctx, method, path, query, raw, nil, result)
}

// Get sends a GET request to baseURL+path with the given query string
// and JSON-decodes the response body into result (if non-nil).
func (h *HTTP) Get(ctx context.Context, path string, query url.Values, result any) error {
	return h.do(ctx, http.MethodGet, path, nil, query, result)
}

// Post sends a POST request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Post(ctx context.Context, path string, body any, result any) error {
	return h.do(ctx, http.MethodPost, path, body, nil, result)
}

// Put sends a PUT request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Put(ctx context.Context, path string, body any, result any) error {
	return h.do(ctx, http.MethodPut, path, body, nil, result)
}

// Patch sends a PATCH request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Patch(ctx context.Context, path string, body any, result any) error {
	return h.do(ctx, http.MethodPatch, path, body, nil, result)
}

// Delete sends a DELETE request to baseURL+path and JSON-decodes
// the response into result (if non-nil).
func (h *HTTP) Delete(ctx context.Context, path string, result any) error {
	return h.do(ctx, http.MethodDelete, path, nil, nil, result)
}

// PostForm sends an application/x-www-form-urlencoded POST to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) PostForm(ctx context.Context, path string, form url.Values, result any) error {
	return h.DoRequest(ctx, http.MethodPost, path, nil,
		[]byte(form.Encode()),
		http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		result)
}
