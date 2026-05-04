package transfer

import wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"

// Client is the WeChat Pay merchant transfer client
type Client struct {
	*wechat.Client
}

// NewClient creates a transfer client wrapping an existing developed Client
func NewClient(c *wechat.Client) *Client {
	return &Client{Client: c}
}
