package offiaccount

import (
	"context"
	"github.com/godrealms/go-wechat-sdk/core"
)

// AccessToken type alias for backward compatibility
type AccessToken = core.AccessToken

// Config holds offiaccount-specific configuration
type Config struct {
	core.BaseConfig
	Token          string `json:"token"`
	EncodingAESKey string `json:"encodingAESKey"`
}

// Client is the WeChat Official Account client
type Client struct {
	*core.BaseClient
	Token          string
	EncodingAESKey string
}

// NewClient creates a new offiaccount client
func NewClient(ctx context.Context, config *Config) *Client {
	base := core.NewBaseClient(ctx, &config.BaseConfig, "https://api.weixin.qq.com", "/cgi-bin/token", "GET")
	return &Client{
		BaseClient:     base,
		Token:          config.Token,
		EncodingAESKey: config.EncodingAESKey,
	}
}
