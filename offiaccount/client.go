// Package offiaccount provides a client for the WeChat Official Account (公众号) server-side API.
// Create a Client with NewClient, then call any of the API methods.
// Token refresh is automatic and concurrency-safe.
package offiaccount

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Official Account credentials from the WeChat developer console.
type Config struct {
	AppId          string `json:"appId"`
	AppSecret      string `json:"appSecret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encodingAESKey"`
}

// Client manages access token lifecycle and provides methods for the WeChat Official Account (公众号) server-side API.
// It is safe for concurrent use.
type Client struct {
	Config *Config
	Https  *utils.HTTP

	tokenMutex  sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}

// Option is a functional configuration applied to a Client during NewClient.
type Option func(*Client)

// WithTokenSource configures the client to obtain access tokens from src instead of calling /cgi-bin/token directly.
// Use this for open-platform component-on-behalf-of flows.
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) { c.tokenSource = ts }
}

// WithHTTPClient replaces the client's default HTTP transport.
// Useful for testing with a mock server or for setting a custom base URL.
func WithHTTPClient(h *utils.HTTP) Option {
	return func(c *Client) {
		if h != nil {
			c.Https = h
		}
	}
}

// NewClient creates an Official Account client.
//
// The ctx parameter is retained for API compatibility but is no longer
// stored on the Client — every method that touches the network takes its
// own ctx, which is the Go-idiomatic shape. Pass anything, including nil.
// Call WithTokenSource to delegate token management to an open-platform
// authorizer.
func NewClient(ctx context.Context, config *Config, opts ...Option) *Client {
	_ = ctx // retained for signature compatibility; not stored
	c := &Client{
		Config: config,
		Https:  utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Invalidate clears the cached access_token, forcing the next AccessTokenE
// call to fetch a fresh one. Use this when WeChat returns a 40001/40014/42001
// error indicating the cached token is no longer valid.
//
// When a TokenSource is configured and implements the Invalidator interface,
// the call is delegated to it. Otherwise, only the Client's internal cache
// is cleared.
//
// doGet and doPost call this automatically on a token-expired response and
// retry the request once with a fresh token, so most callers never need to
// invoke it directly.
func (c *Client) Invalidate() {
	if c.tokenSource != nil {
		if inv, ok := c.tokenSource.(Invalidator); ok {
			inv.Invalidate()
		}
		return
	}
	c.tokenMutex.Lock()
	c.accessToken = ""
	c.expiresAt = time.Time{}
	c.tokenMutex.Unlock()
}

// AccessTokenE returns a valid access_token, propagating any fetch error to the caller.
// It uses an in-process read-write-locked cache; the token is refreshed 60 s before expiry.
// When a TokenSource is configured the call is delegated to it without touching /cgi-bin/token.
func (c *Client) AccessTokenE(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}

	c.tokenMutex.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		token := c.accessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return c.accessToken, nil
	}

	token, err := c.refreshAccessToken(ctx)
	if err != nil {
		return "", err
	}
	c.accessToken = token.AccessToken
	// Clamp TTL with a 60s floor so a malformed/hostile upstream returning
	// expires_in<=60 cannot poison the cache into a permanent miss and
	// trigger a token-refresh storm. Same guard as mini-program/channels (audit C7).
	ttl := token.ExpiresIn - 60
	if ttl < 60 {
		ttl = 60
	}
	c.expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	return c.accessToken, nil
}

// refreshAccessToken 调用微信服务器刷新 access_token。
func (c *Client) refreshAccessToken(ctx context.Context) (*AccessToken, error) {
	if c.Config == nil || c.Config.AppId == "" || c.Config.AppSecret == "" {
		return nil, fmt.Errorf("offiaccount: AppId and AppSecret are required")
	}
	query := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.Config.AppId},
		"secret":     {c.Config.AppSecret},
	}
	result := &AccessToken{}
	if err := c.Https.Get(ctx, "/cgi-bin/token", query, result); err != nil {
		return nil, fmt.Errorf("offiaccount: fetch access_token failed: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("offiaccount: empty access_token returned")
	}
	return result, nil
}
