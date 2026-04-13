// Package oplatform provides a client for the WeChat Open Platform (开放平台)
// third-party platform API. Use it to manage component access tokens and
// delegate API calls on behalf of authorized official accounts.
package oplatform

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

// Config holds the Open Platform component credentials.
type Config struct {
	ComponentAppID     string // 第三方平台 appid
	ComponentAppSecret string // 第三方平台 secret
	Token              string // 回调签名 Token
	EncodingAESKey     string // 43 字符 AESKey
}

// Client is the WeChat Open Platform third-party platform client. Safe for concurrent use.
type Client struct {
	cfg    Config
	http   *utils.HTTP
	store  Store
	crypto *wxcrypto.MsgCrypto

	componentMu sync.Mutex // 保护 component access_token 刷新
	authMu      sync.Map   // map[string]*sync.Mutex per-authorizer 刷新锁
}

// Option 构造可选项。
type Option func(*Client)

// WithStore 注入自定义 Store 实现（默认 MemoryStore）。
func WithStore(s Store) Option {
	return func(c *Client) {
		if s != nil {
			c.store = s
		}
	}
}

// WithHTTP 注入自定义 utils.HTTP 客户端（测试常用）。
func WithHTTP(h *utils.HTTP) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// NewClient constructs an Open Platform client. No network requests are made during construction.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.ComponentAppID == "" {
		return nil, fmt.Errorf("oplatform: ComponentAppID is required")
	}
	if cfg.ComponentAppSecret == "" {
		return nil, fmt.Errorf("oplatform: ComponentAppSecret is required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("oplatform: Token is required")
	}
	crypto, err := wxcrypto.New(cfg.Token, cfg.EncodingAESKey, cfg.ComponentAppID)
	if err != nil {
		return nil, fmt.Errorf("oplatform: init crypto: %w", err)
	}
	c := &Client{
		cfg:    cfg,
		http:   utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
		store:  NewMemoryStore(),
		crypto: crypto,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Store 暴露底层 Store，便于外部检查/管理（例如清理失效 authorizer）。
func (c *Client) Store() Store { return c.store }

// HTTP 暴露底层 HTTP 客户端。
func (c *Client) HTTP() *utils.HTTP { return c.http }

// ComponentAppID 返回第三方平台 appid。
func (c *Client) ComponentAppID() string { return c.cfg.ComponentAppID }

// checkWeixinErr 如果 errcode != 0 则返回 *WeixinError，否则 nil。
func checkWeixinErr(errcode int, errmsg string) error {
	if errcode == 0 {
		return nil
	}
	return &WeixinError{ErrCode: errcode, ErrMsg: errmsg}
}

// touchContext 保证 context 非 nil。
func touchContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
