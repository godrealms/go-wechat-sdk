// Package mini_program provides a client for the WeChat Mini Program
// (小程序) server-side API. The primary entry point is NewClient.
//
// 当前实现：
//   - Code2Session: 登录凭证校验（wx.login 换 openid/session_key）
//   - AccessToken:  获取/缓存全局 access_token
//   - DecryptUserData: 解密 wx.getUserInfo / wx.getPhoneNumber 返回的 encryptedData
//   - SendSubscribeMessage: 发送订阅消息
//
// 其它接口（数据分析、云开发、直播等）未逐个封装；通过 Client.HTTP 的 Get/Post 可直接调用，
// 也欢迎后续按同一模式扩展。
package mini_program

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Mini Program credentials from the WeChat developer console.
type Config struct {
	AppId     string
	AppSecret string
}

// Client is the Mini Program server-side client. It caches the access_token in-process
// and refreshes it automatically 60 s before expiry. Safe for concurrent use.
type Client struct {
	cfg  Config
	http *utils.HTTP

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithHTTP overrides the default HTTP client (primarily used in tests).
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource injects an external access token source (e.g. Open Platform delegated calls).
// When set, AccessToken() no longer calls /cgi-bin/token.
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) { c.tokenSource = ts }
}

// NewClient constructs a Mini Program client.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.AppId == "" {
		return nil, fmt.Errorf("mini_program: AppId is required")
	}
	c := &Client{
		cfg:  cfg,
		http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	if c.tokenSource == nil && cfg.AppSecret == "" {
		return nil, fmt.Errorf("mini_program: AppSecret is required when no TokenSource is injected")
	}
	return c, nil
}

// HTTP returns the underlying HTTP client, enabling callers to make custom API calls.
func (c *Client) HTTP() *utils.HTTP { return c.http }

// Code2SessionResp is the response from the wx.login code-to-session exchange.
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
		return nil, fmt.Errorf("mini_program: jsCode is required")
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
		return "", fmt.Errorf("mini_program: fetch token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", &APIError{ErrCode: out.ErrCode, ErrMsg: out.ErrMsg, Path: "/cgi-bin/token"}
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("mini_program: empty access_token")
	}
	c.accessToken = out.AccessToken
	ttl := out.ExpiresIn - 60
	if ttl < 60 {
		ttl = 60 // safety floor: never cache a token for less than 60s.
	}
	c.expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	return c.accessToken, nil
}

// SendSubscribeMessage sends a subscription template message to the user identified in body.
func (c *Client) SendSubscribeMessage(ctx context.Context, body any) error {
	if body == nil {
		return fmt.Errorf("mini_program: SendSubscribeMessage: body is required")
	}
	token, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {token}}
	out := struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}
	path := "/cgi-bin/message/subscribe/send?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &out); err != nil {
		return err
	}
	if out.ErrCode != 0 {
		return &APIError{ErrCode: out.ErrCode, ErrMsg: out.ErrMsg, Path: "/cgi-bin/message/subscribe/send"}
	}
	return nil
}
