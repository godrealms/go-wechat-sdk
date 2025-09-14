package offiaccount

// GetUserTitleUrl 获取添加发票链接
// req: 获取添加发票链接请求参数
func (c *Client) GetUserTitleUrl(req *GetUserTitleUrlRequest) (*GetUserTitleUrlResult, error) {
	// 构造请求URL
	path := "/card/invoice/biz/getusertitleurl"

	// 发送请求
	var result GetUserTitleUrlResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSelectTitleUrl 获取选择发票抬头链接
// req: 获取选择发票抬头链接请求参数
func (c *Client) GetSelectTitleUrl(req *GetSelectTitleUrlRequest) (*GetSelectTitleUrlResult, error) {
	// 构造请求URL
	path := "/card/invoice/biz/getselecttitleurl"

	// 发送请求
	var result GetSelectTitleUrlResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ScanTitle 扫描二维码获取抬头
// req: 扫描二维码获取抬头请求参数
func (c *Client) ScanTitle(req *ScanTitleRequest) (*ScanTitleResult, error) {
	// 构造请求URL
	path := "/card/invoice/scantitle"

	// 发送请求
	var result ScanTitleResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
