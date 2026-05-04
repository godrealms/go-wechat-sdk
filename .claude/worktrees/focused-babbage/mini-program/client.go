package mini_program

import (
	"context"
	"github.com/godrealms/go-wechat-sdk/core"
)

// Config holds mini-program configuration
type Config struct {
	core.BaseConfig
}

// Client is the WeChat Mini-Program client
type Client struct {
	*core.BaseClient
}

// NewClient creates a new mini-program client
// Token refresh uses POST to /cgi-bin/stable_token (different from offiaccount's GET /cgi-bin/token)
func NewClient(ctx context.Context, config *Config) *Client {
	base := core.NewBaseClient(ctx, &config.BaseConfig, "https://api.weixin.qq.com", "/cgi-bin/stable_token", "POST")
	return &Client{BaseClient: base}
}
