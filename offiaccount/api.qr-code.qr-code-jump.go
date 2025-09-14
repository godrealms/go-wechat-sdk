package offiaccount

// GetQRCodeJump 获取已设置的二维码规则
// req: 获取二维码跳转规则请求参数
func (c *Client) GetQRCodeJump(req *GetQRCodeJumpRequest) (*GetQRCodeJumpResult, error) {
	// 构造请求URL
	path := "/cgi-bin/wxopen/qrcodejumpget"

	// 发送请求
	var result GetQRCodeJumpResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AddQRCodeJump 增加二维码规则
// req: 增加二维码规则请求参数
func (c *Client) AddQRCodeJump(req *AddQRCodeJumpRequest) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/wxopen/qrcodejumpadd"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// PublishQRCodeJump 发布已设置的二维码规则
// prefix: 二维码规则
func (c *Client) PublishQRCodeJump(prefix string) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/wxopen/qrcodejumppublish"

	// 构造请求体
	req := &PublishQRCodeJumpRequest{
		Prefix: prefix,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteQRCodeJump 删除已设置的二维码规则
// req: 删除二维码规则请求参数
func (c *Client) DeleteQRCodeJump(req *DeleteQRCodeJumpRequest) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/wxopen/qrcodejumpdelete"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
