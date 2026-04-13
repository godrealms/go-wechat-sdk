package pay

import (
	"context"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// TransactionsH5 places a new H5-payment order and returns the redirect URL for WeChat Pay in a mobile browser.
// 文档：https://pay.weixin.qq.com/doc/v3/merchant/4012791863
func (c *Client) TransactionsH5(ctx context.Context, order *types.Transactions) (*types.TransactionsH5Resp, error) {
	resp := &types.TransactionsH5Resp{}
	if err := c.postV3(ctx, "/v3/pay/transactions/h5", order, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
