// Package xiaowei provides a client for the WeChat Xiaowei (小微商户)
// micro-merchant API. Create a Client with NewClient; token refresh is automatic.
// Xiaowei serves individual sellers who operate without a business license.
package xiaowei

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Xiaowei app credentials.
type Config struct {
	AppId     string
	AppSecret string
}

// TokenSource is an injectable access_token provider.
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

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

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// fetchToken issues the /cgi-bin/token HTTP call. It is passed to TokenCache
// as the refresh callback and is not called directly.
func (c *Client) fetchToken(ctx context.Context) (string, int64, error) {
	q := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.cfg.AppId},
		"secret":     {c.cfg.AppSecret},
	}
	out := &accessTokenResp{}
	if err := c.http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
		return "", 0, fmt.Errorf("xiaowei: fetch token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", 0, &APIError{ErrCode: out.ErrCode, ErrMsg: out.ErrMsg, Path: "/cgi-bin/token"}
	}
	return out.AccessToken, out.ExpiresIn, nil
}

// AccessToken returns a valid access_token, refreshing 60 s before expiry.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}
	return c.cache.Get(ctx)
}
