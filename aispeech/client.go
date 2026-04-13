// Package aispeech provides a client for the WeChat AI Speech (智能对话) API,
// covering automatic speech recognition (ASR), text-to-speech (TTS),
// natural language understanding (NLU), and dialog management.
// Create a Client with NewClient; token refresh is automatic.
package aispeech

import (
	"context"
	"fmt"
	"net/url"
	"sync"
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
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

// Client manages the aispeech API. Safe for concurrent use.
type Client struct {
	cfg         Config
	http        *utils.HTTP
	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time
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
	return c, nil
}

// HTTP returns the underlying HTTP client for making custom requests.
func (c *Client) HTTP() *utils.HTTP { return c.http }

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// AccessToken returns a valid access_token, refreshing 60 s before expiry.
// When a TokenSource is configured, the call is forwarded to it.
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
		return "", fmt.Errorf("aispeech: fetch token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", fmt.Errorf("aispeech: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("aispeech: empty access_token")
	}
	c.accessToken = out.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}
