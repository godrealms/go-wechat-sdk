package offiaccount

import (
	"context"
	"fmt"
)

// GetProductCardInfo 获取商品卡片的DOM结构
// req: 获取商品卡片信息请求参数
func (c *Client) GetProductCardInfo(ctx context.Context, req *GetProductCardInfoRequest) (*GetProductCardInfoResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/channels/ec/service/product/getcardinfo?access_token=%s", token)

	// 发送请求
	var result GetProductCardInfoResult
	err = c.Https.Post(ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
