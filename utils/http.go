package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Logger is an optional debug-logging interface. By default the SDK does not print any
// request/response content to avoid leaking sensitive data (OpenIDs, order numbers, amounts)
// in production. Callers may inject their own implementation via WithLogger.
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
func WithLogger(logger Logger) Option {
	return func(h *HTTP) {
		if logger != nil {
			h.Logger = logger
		}
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

	h.Logger.Debugf("http %s %s body=%s", method, fullURL, string(body))

	resp, err := h.Client.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, fmt.Errorf("read response body failed: %w", err)
	}

	h.Logger.Debugf("http %s %s status=%d response=%s",
		method, fullURL, resp.StatusCode, string(respBody))

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
