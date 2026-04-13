package pay

import (
	"context"
	"strconv"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// TransactionsApp places a new App-payment order and returns the prepay_id needed to invoke the WeChat Pay SDK on the client.
// 文档：https://pay.weixin.qq.com/doc/v3/merchant/4012791861
func (c *Client) TransactionsApp(ctx context.Context, order *types.Transactions) (*types.TransactionsAppResponse, error) {
	resp := &types.TransactionsAppResponse{}
	if err := c.postV3(ctx, "/v3/pay/transactions/app", order, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ModifyTransactionsApp calls TransactionsApp and assembles the signed APP payment parameters.
func (c *Client) ModifyTransactionsApp(ctx context.Context, order *types.Transactions) (*types.ModifyAppResponse, error) {
	resp, err := c.TransactionsApp(ctx, order)
	if err != nil {
		return nil, err
	}
	parameter := &types.ModifyAppResponse{
		AppId:        c.appid,
		PartnerId:    c.mchid,
		PrepayId:     resp.PrepayId,
		PackageValue: "Sign=WXPay",
		NonceStr:     utils.GenerateHashBasedString(32),
		TimeStamp:    strconv.FormatInt(time.Now().Unix(), 10),
	}
	if err := parameter.GenerateSignature(c.privateKey); err != nil {
		return nil, err
	}
	return parameter, nil
}
