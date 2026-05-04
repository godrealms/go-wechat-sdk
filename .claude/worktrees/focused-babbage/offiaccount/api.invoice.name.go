package offiaccount

import "fmt"

// GetUserTitleUrl 获取添加发票链接
// req: 获取添加发票链接请求参数
func (c *Client) GetUserTitleUrl(req *GetUserTitleUrlRequest) (*GetUserTitleUrlResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/biz/getusertitleurl?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result GetUserTitleUrlResult
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSelectTitleUrl 获取选择发票抬头链接
// req: 获取选择发票抬头链接请求参数
func (c *Client) GetSelectTitleUrl(req *GetSelectTitleUrlRequest) (*GetSelectTitleUrlResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/biz/getselecttitleurl?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result GetSelectTitleUrlResult
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ScanTitle 扫描二维码获取抬头
// req: 扫描二维码获取抬头请求参数
func (c *Client) ScanTitle(req *ScanTitleRequest) (*ScanTitleResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/scantitle?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result ScanTitleResult
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
