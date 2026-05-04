package bill

import wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"

// Client is the WeChat Pay bill client
type Client struct {
	*wechat.Client
}

// NewClient creates a bill client wrapping an existing developed Client
func NewClient(c *wechat.Client) *Client {
	return &Client{Client: c}
}
