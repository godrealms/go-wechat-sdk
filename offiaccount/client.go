package offiaccount

import (
	"context"
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
	ctx         context.Context
	Config      *Config
	Https       *utils.HTTP
	accessToken *AccessToken
	tokenMutex  sync.RWMutex
}

// NewClient 创建客户端
func NewClient(ctx context.Context, config *Config) *Client {
	return &Client{
		ctx:    ctx,
		Config: config,
		Https:  utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
}

// getAccessToken 获取access_token，自动刷新
func (c *Client) getAccessToken() string {
	c.tokenMutex.RLock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		token := c.accessToken.AccessToken
		c.tokenMutex.RUnlock()
		return token
	}
	c.tokenMutex.RUnlock()

	// 需要刷新token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// 双重检查
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		return c.accessToken.AccessToken
	}

	// 刷新access_token
	token, err := c.refreshAccessToken()
	if err != nil {
		return ""
	}
	c.accessToken = token
	return c.accessToken.AccessToken
}

// refreshAccessToken 刷新access_token
func (c *Client) refreshAccessToken() (*AccessToken, error) {
	query := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.Config.AppId},
		"secret":     {c.Config.AppSecret},
	}
	result := &AccessToken{}
	err := c.Https.Get(c.ctx, "/cgi-bin/token", query, result)
	if err != nil {
		return nil, err
	}
	// 提前10秒过期，避免临界点问题
	result.ExpiresIn = time.Now().Unix() + result.ExpiresIn - 10
	return result, nil
}
