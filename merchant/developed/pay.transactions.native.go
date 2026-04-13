package pay

import (
	"context"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// TransactionsNative places a new Native-payment order and returns the code_url for generating a QR code.
// 文档：https://pay.weixin.qq.com/doc/v3/merchant/4012791864
func (c *Client) TransactionsNative(ctx context.Context, order *types.Transactions) (*types.TransactionsNativeResp, error) {
	resp := &types.TransactionsNativeResp{}
	if err := c.postV3(ctx, "/v3/pay/transactions/native", order, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
