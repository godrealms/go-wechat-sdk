package offiaccount

// GetJSApiTicket 获取jsapi_ticket
// 用于调用 js-sdk 的临时票据，有效期为7200 秒，通过access_token 来获取
func (c *Client) GetJSApiTicket() (*Ticket, error) {
	// 构造请求URL
	path := "/cgi-bin/ticket/getticket?type=jsapi"

	// 发送请求
	var result Ticket
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWxCardTicket 获取微信卡券ticket
// 用于调用微信卡券的临时票据，有效期为7200 秒，通过access_token 来获取
func (c *Client) GetWxCardTicket() (*Ticket, error) {
	// 构造请求URL
	path := "/cgi-bin/ticket/getticket?type=wx_card"

	// 发送请求
	var result Ticket
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
