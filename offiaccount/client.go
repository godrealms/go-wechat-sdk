package offiaccount

import (
	"context"
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
	AccessToken *AccessToken
}

// NewClient 创建客户端
func NewClient(ctx context.Context, config *Config) *Client {
	return &Client{
		ctx:    ctx,
		Config: config,
		Https:  utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
}
