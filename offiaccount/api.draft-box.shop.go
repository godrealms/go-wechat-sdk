package offiaccount

import "fmt"

// GetProductCardInfo 获取商品卡片的DOM结构
// req: 获取商品卡片信息请求参数
func (c *Client) GetProductCardInfo(req *GetProductCardInfoRequest) (*GetProductCardInfoResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/channels/ec/service/product/getcardinfo?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result GetProductCardInfoResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
