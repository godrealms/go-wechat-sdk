// Package channels provides a client for the WeChat Channels (视频号) server-side API.
package channels

import (
	"context"
	"fmt"
	"net/url"
	"sync"
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
	cfg  Config
	http *utils.HTTP

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}

// TokenSource is an injectable source of access tokens. When a Client is configured with a
// TokenSource, AccessToken() delegates to it instead of calling /cgi-bin/token directly.
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

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
	return c, nil
}

// HTTP returns the underlying HTTP client, useful for calling custom WeChat API endpoints.
func (c *Client) HTTP() *utils.HTTP { return c.http }

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// AccessToken returns a valid global access_token, refreshing it when fewer than
// 60 seconds remain before expiry. When a TokenSource is injected, delegates to it.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		t := c.accessToken
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return c.accessToken, nil
	}
	q := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.cfg.AppId},
		"secret":     {c.cfg.AppSecret},
	}
	out := &accessTokenResp{}
	if err := c.http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
		return "", fmt.Errorf("channels: fetch token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", fmt.Errorf("channels: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("channels: empty access_token")
	}
	c.accessToken = out.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}
