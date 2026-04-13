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

// Option 构造可选项。
type Option func(*Client)

// WithHTTP 用自定义 HTTP 客户端（主要用于测试）。
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource 注入外部 access_token 来源（例如开放平台代调用）。
// 设置后 AccessToken() 不再调用 /cgi-bin/token。
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

// HTTP 返回底层 HTTP 客户端，便于扩展自定义接口。
func (c *Client) HTTP() *utils.HTTP { return c.http }

// Code2SessionResp 是 wx.login 换 session 的响应。
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
		return nil, fmt.Errorf("mini_program: code2session errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
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
		return "", fmt.Errorf("mini_program: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("mini_program: empty access_token")
	}
	c.accessToken = out.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}

// SendSubscribeMessage sends a subscription template message to the user identified in body.
func (c *Client) SendSubscribeMessage(ctx context.Context, body any) error {
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
		return fmt.Errorf("mini_program: subscribe send errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	return nil
}
