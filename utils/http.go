package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTP is a thin JSON-over-HTTP client used by all SDK packages.
// It stores a base URL and optional default headers so callers never
// repeat boilerplate. Safe for concurrent use after construction.
type HTTP struct {
	BaseURL string
	Client  *http.Client
	Headers map[string]string
	Timeout time.Duration
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
	}

	for _, opt := range opts {
		opt(h)
	}

	h.Client = &http.Client{
		Timeout: h.Timeout,
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

// SetBaseURL replaces the base URL after construction. Not goroutine-safe;
// call only before the client is shared between goroutines.
func (h *HTTP) SetBaseURL(url string) {
	h.BaseURL = url
}

// do is the shared implementation for all HTTP verbs. It builds the full URL,
// marshals the request body (if any), sets headers, executes the request, and
// unmarshals the response into result.
func (h *HTTP) do(ctx context.Context, method, path string, body interface{}, query url.Values, result interface{}) error {
	if query != nil {
		path += "?" + query.Encode()
	}
	fullURL := h.BaseURL + path

	log.Println("method:", method)
	log.Println("fullURL:", fullURL)

	var bodyReader io.Reader
	var bodyJson string
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		bodyJson = string(jsonBody)
		log.Println("body:", bodyJson)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	log.Println("response:", string(respBody))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
		}
	}

	return nil
}

// Get sends a GET request to baseURL+path with the given query string
// and JSON-decodes the response body into result (if non-nil).
func (h *HTTP) Get(ctx context.Context, path string, query url.Values, result interface{}) error {
	return h.do(ctx, http.MethodGet, path, nil, query, result)
}

// Post sends a POST request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return h.do(ctx, http.MethodPost, path, body, nil, result)
}

// Put sends a PUT request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return h.do(ctx, http.MethodPut, path, body, nil, result)
}

// Patch sends a PATCH request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return h.do(ctx, http.MethodPatch, path, body, nil, result)
}

// Delete sends a DELETE request to baseURL+path and JSON-decodes
// the response into result (if non-nil).
func (h *HTTP) Delete(ctx context.Context, path string, result interface{}) error {
	return h.do(ctx, http.MethodDelete, path, nil, nil, result)
}

// PostForm sends an application/x-www-form-urlencoded POST to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) PostForm(ctx context.Context, path string, form url.Values, result interface{}) error {
	fullURL := h.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w:%s", err, string(respBody))
		}
	}

	return nil
}
