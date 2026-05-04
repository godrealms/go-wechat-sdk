// Package xiaowei provides a client for the WeChat Xiaowei (小微商户)
// micro-merchant API. Create a Client with NewClient; token refresh is automatic.
// Xiaowei serves individual sellers who operate without a business license.
package xiaowei

import (
	"context"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Xiaowei app credentials.
type Config struct {
	AppId     string
	AppSecret string
}

// TokenSource is an injectable access_token provider. Aliased to
// utils.TokenSource so a single implementation works across every
// WeChat-product Client.
type TokenSource = utils.TokenSource

// Client manages the Xiaowei API. Safe for concurrent use.
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
func WithTokenSource(ts TokenSource) Option { return func(c *Client) { c.tokenSource = ts } }

// NewClient constructs a Xiaowei client. Returns an error if AppId is
// empty or if AppSecret is empty and no TokenSource is provided.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.AppId == "" {
		return nil, fmt.Errorf("xiaowei: AppId is required")
	}
	c := &Client{
		cfg:  cfg,
		http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(30*time.Second)),
	}
	for _, o := range opts {
		o(c)
	}
	if c.tokenSource == nil && cfg.AppSecret == "" {
		return nil, fmt.Errorf("xiaowei: AppSecret is required when no TokenSource is injected")
	}
	c.cache = utils.NewTokenCache("xiaowei", c.fetchToken)
	return c, nil
}

// HTTP returns the underlying HTTP client for custom requests.
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
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}
	return c.cache.Get(ctx)
}
