// Package channels provides a client for the WeChat Channels (视频号) server-side API.
package channels

import (
	"context"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Channels application credentials.
type Config struct {
	AppId     string
	AppSecret string
}

// Client is the WeChat Channels server-side API client. Safe for concurrent use.
type Client struct {
	cfg   Config
	http  *utils.HTTP
	cache *utils.TokenCache

	tokenSource TokenSource
}

// TokenSource is an injectable source of access tokens. When a Client is configured with a
// TokenSource, AccessToken() delegates to it instead of calling /cgi-bin/token directly.
//
// Aliased to utils.TokenSource so a single implementation (typically
// oplatform.AuthorizerClient) satisfies every WeChat-product Client.
type TokenSource = utils.TokenSource

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithHTTP replaces the default HTTP client, primarily used for testing.
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource injects an external access-token source (e.g. for Open Platform proxy calls).
// When set, AccessToken() will no longer call /cgi-bin/token.
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) { c.tokenSource = ts }
}

// NewClient constructs a Channels client.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.AppId == "" {
		return nil, fmt.Errorf("channels: AppId is required")
	}
	c := &Client{
		cfg:  cfg,
		http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	if c.tokenSource == nil && cfg.AppSecret == "" {
		return nil, fmt.Errorf("channels: AppSecret is required when no TokenSource is injected")
	}
	c.cache = utils.NewTokenCache("channels", c.fetchToken)
	return c, nil
}

// HTTP returns the underlying HTTP client, useful for calling custom WeChat API endpoints.
func (c *Client) HTTP() *utils.HTTP { return c.http }

// fetchToken delegates to utils.FetchAccessToken and is passed to TokenCache
// as the refresh callback. The closure exists so the TokenCache callback can
// produce a package-typed *APIError on errcode failures.
func (c *Client) fetchToken(ctx context.Context) (string, int64, error) {
	return utils.FetchAccessToken(ctx, c.http, c.cfg.AppId, c.cfg.AppSecret,
		func(code int, msg, path string) error {
			return &APIError{ErrCode: code, ErrMsg: msg, Path: path}
		})
}

// AccessToken returns a valid global access_token, refreshing it when fewer than
// 60 seconds remain before expiry. When a TokenSource is injected, delegates to it.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}
	return c.cache.Get(ctx)
}
