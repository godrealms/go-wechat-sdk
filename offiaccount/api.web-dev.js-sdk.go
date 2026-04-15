package offiaccount

import (
	"context"
	"net/url"
)

// GetJSApiTicket 获取jsapi_ticket
// 用于调用 js-sdk 的临时票据，有效期为7200 秒，通过access_token 来获取
func (c *Client) GetJSApiTicket(ctx context.Context) (*Ticket, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := "/cgi-bin/ticket/getticket"
	params := url.Values{
		"access_token": {token},
		"type":         {"jsapi"},
	}

	// 发送请求
	var result Ticket
	if err := c.doGet(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWxCardTicket 获取微信卡券ticket
// 用于调用微信卡券的临时票据，有效期为7200 秒，通过access_token 来获取
func (c *Client) GetWxCardTicket(ctx context.Context) (*Ticket, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := "/cgi-bin/ticket/getticket"
	params := url.Values{
		"access_token": {token},
		"type":         {"wx_card"},
	}

	// 发送请求
	var result Ticket
	if err := c.doGet(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
