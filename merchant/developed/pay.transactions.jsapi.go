package pay

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// TransactionsJsapi places a new JSAPI-payment order and returns the prepay_id needed to invoke WeChat Pay in a Mini Program or official account page.
// 文档：https://pay.weixin.qq.com/doc/v3/merchant/4012791862
func (c *Client) TransactionsJsapi(ctx context.Context, order *types.Transactions) (*types.TransactionsJsapiResp, error) {
	resp := &types.TransactionsJsapiResp{}
	if err := c.postV3(ctx, "/v3/pay/transactions/jsapi", order, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ModifyTransactionsJsapi 在 TransactionsJsapi 之后，组装并签名小程序/JSAPI 调起支付参数。
func (c *Client) ModifyTransactionsJsapi(ctx context.Context, order *types.Transactions) (*types.TransactionsJsapi, error) {
	resp, err := c.TransactionsJsapi(ctx, order)
	if err != nil {
		return nil, err
	}
	parameter := &types.TransactionsJsapi{
		TimeStamp: strconv.FormatInt(time.Now().Unix(), 10),
		NonceStr:  utils.GenerateHashBasedString(32),
		Package:   fmt.Sprintf("prepay_id=%s", resp.PrepayId),
		SignType:  "RSA",
	}
	if err := parameter.GenerateSignature(c.appid, c.privateKey); err != nil {
		return nil, err
	}
	return parameter, nil
}
