// Package channels 提供微信视频号 API 的常用入口。
package channels

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config 视频号配置。
type Config struct {
	AppId     string
	AppSecret string
}

// Client 视频号服务端客户端。并发安全。
type Client struct {
	cfg  Config
	http *utils.HTTP

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}

// TokenSource 是 access_token 的可注入来源。
// 当 Client 配置了 TokenSource 时，AccessToken() 会直接委托给它，
// 不再调用 /cgi-bin/token。
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
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

// NewClient 构造客户端。
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

// HTTP 返回底层 HTTP 客户端，便于扩展自定义接口。
func (c *Client) HTTP() *utils.HTTP { return c.http }

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// AccessToken 获取全局 access_token（带进程内缓存，提前 60 秒过期）。
// 注入 TokenSource 时直接委托。
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
