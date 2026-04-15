// Package mini_game provides a client for the WeChat Mini Game (小游戏) server-side API.
//
// 当前实现：
//   - Code2Session: 登录凭证校验（wx.login 换 openid/session_key）
//   - AccessToken:  获取/缓存全局 access_token
//
// 其它接口通过 Client.HTTP 的 Get/Post 可直接调用，
// 也欢迎后续按同一模式扩展。
package mini_game

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Mini Game credentials from the WeChat developer console.
type Config struct {
	AppId     string
	AppSecret string
}

// Client is the Mini Game server-side API client. Safe for concurrent use.
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
// A typical use case is Open Platform proxy calls.
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

// NewClient constructs a Mini Game client.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.AppId == "" {
		return nil, fmt.Errorf("mini_game: AppId is required")
	}
	c := &Client{
		cfg:  cfg,
		http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	if c.tokenSource == nil && cfg.AppSecret == "" {
		return nil, fmt.Errorf("mini_game: AppSecret is required when no TokenSource is injected")
	}
	return c, nil
}

// HTTP returns the underlying HTTP client, useful for calling custom WeChat API endpoints.
func (c *Client) HTTP() *utils.HTTP { return c.http }

// Code2SessionResp holds the response from the WeChat login credential validation endpoint.
type Code2SessionResp struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode,omitempty"`
	ErrMsg     string `json:"errmsg,omitempty"`
}

// Code2Session exchanges the js_code obtained from wx.login for the user's openid,
// session_key, and (if applicable) unionid.
func (c *Client) Code2Session(ctx context.Context, jsCode string) (*Code2SessionResp, error) {
	if jsCode == "" {
		return nil, fmt.Errorf("mini_game: jsCode is required")
	}
	q := url.Values{
		"appid":      {c.cfg.AppId},
		"secret":     {c.cfg.AppSecret},
		"js_code":    {jsCode},
		"grant_type": {"authorization_code"},
	}
	out := &Code2SessionResp{}
	if err := c.http.Get(ctx, "/sns/jscode2session", q, out); err != nil {
		return nil, err
	}
	if out.ErrCode != 0 {
		return nil, &APIError{ErrCode: out.ErrCode, ErrMsg: out.ErrMsg, Path: "/sns/jscode2session"}
	}
	return out, nil
}

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
		return "", fmt.Errorf("mini_game: fetch token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", &APIError{ErrCode: out.ErrCode, ErrMsg: out.ErrMsg, Path: "/cgi-bin/token"}
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("mini_game: empty access_token")
	}
	// Clamp TTL: WeChat normally returns 7200s; a hostile or malformed upstream
	// could return a small or zero value which would leave expiresAt in the
	// past and cause a refresh storm. Floor at 60s.
	ttl := out.ExpiresIn - 60
	if ttl < 60 {
		ttl = 60
	}
	c.accessToken = out.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	return c.accessToken, nil
}
