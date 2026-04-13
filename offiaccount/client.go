package offiaccount

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

type Config struct {
	AppId          string `json:"appId"`
	AppSecret      string `json:"appSecret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encodingAESKey"`
}

// Client 微信公众号
type Client struct {
	ctx    context.Context
	Config *Config
	Https  *utils.HTTP

	tokenMutex  sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}

// Option 构造可选项。
type Option func(*Client)

// WithTokenSource 注入外部 access_token 来源（例如开放平台代调用）。
// 设置后 AccessTokenE 不再调用 /cgi-bin/token。
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) { c.tokenSource = ts }
}

// WithHTTPClient 允许替换底层 HTTP 客户端（主要用于测试注入）。
func WithHTTPClient(h *utils.HTTP) Option {
	return func(c *Client) {
		if h != nil {
			c.Https = h
		}
	}
}

// NewClient 创建客户端
func NewClient(ctx context.Context, config *Config, opts ...Option) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	c := &Client{
		ctx:    ctx,
		Config: config,
		Https:  utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// AccessTokenE 显式获取 access_token，错误会被传递给调用方。
// 如果注入了 TokenSource，则委托给它；否则走自有 /cgi-bin/token 流程。
func (c *Client) AccessTokenE(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = c.ctx
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
	c.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}

// getAccessToken 旧版内部入口（保留以兼容现有代码）。
//
// Deprecated: 使用 AccessTokenE 来正确处理错误。
func (c *Client) getAccessToken() string {
	token, _ := c.AccessTokenE(c.ctx)
	return token
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
