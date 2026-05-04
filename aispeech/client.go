// Package aispeech provides a client for the WeChat AI Speech (智能对话) API,
// covering automatic speech recognition (ASR), text-to-speech (TTS),
// natural language understanding (NLU), and dialog management.
// Create a Client with NewClient; token refresh is automatic.
package aispeech

import (
	"context"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the aispeech app credentials.
type Config struct {
	AppId     string
	AppSecret string
}

// TokenSource is an injectable access_token provider. Configure it via
// WithTokenSource to delegate token management (e.g. open-platform flows)
// without calling /cgi-bin/token.
//
// Aliased to utils.TokenSource so a single implementation works across
// every WeChat-product Client.
type TokenSource = utils.TokenSource

// Client manages the aispeech API. Safe for concurrent use.
type Client struct {
	cfg   Config
	http  *utils.HTTP
	cache *utils.TokenCache

	tokenSource TokenSource
}

// Option is a functional option applied during NewClient.
type Option func(*Client)

// WithHTTP injects a custom HTTP client, primarily for testing.
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource injects an external access_token provider.
// When set, AccessToken() delegates to it without calling /cgi-bin/token.
func WithTokenSource(ts TokenSource) Option { return func(c *Client) { c.tokenSource = ts } }

// NewClient constructs an aispeech client. Returns an error if AppId is
// empty or if AppSecret is empty and no TokenSource is provided.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.AppId == "" {
		return nil, fmt.Errorf("aispeech: AppId is required")
	}
	c := &Client{
		cfg:  cfg,
		http: utils.NewHTTP("https://openai.weixin.qq.com", utils.WithTimeout(30*time.Second)),
	}
	for _, o := range opts {
		o(c)
	}
	if c.tokenSource == nil && cfg.AppSecret == "" {
		return nil, fmt.Errorf("aispeech: AppSecret is required when no TokenSource is injected")
	}
	c.cache = utils.NewTokenCache("aispeech", c.fetchToken)
	return c, nil
}

// HTTP returns the underlying HTTP client for making custom requests.
func (c *Client) HTTP() *utils.HTTP { return c.http }

// fetchToken delegates to utils.FetchAccessToken; the closure produces
// a package-typed *APIError on errcode failures.
func (c *Client) fetchToken(ctx context.Context) (string, int64, error) {
	return utils.FetchAccessToken(ctx, c.http, c.cfg.AppId, c.cfg.AppSecret,
		func(code int, msg, path string) error {
			return &APIError{ErrCode: code, ErrMsg: msg, Path: path}
		})
}

// AccessToken returns a valid access_token, refreshing 60 s before expiry.
// When a TokenSource is configured, the call is forwarded to it.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}
	return c.cache.Get(ctx)
}
