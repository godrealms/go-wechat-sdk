package pay

import (
	"context"
	"fmt"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// TransactionsClose closes an existing order. The WeChat Pay API requires a {"mchid":"..."} body
// and returns 204 No Content on success.
// 微信支付要求 body 是 {"mchid":"..."}，且关闭成功返回 204 No Content。
func (c *Client) TransactionsClose(ctx context.Context, outTradeNo string) error {
	if outTradeNo == "" {
		return fmt.Errorf("pay: outTradeNo is required")
	}
	body := &types.MchID{Mchid: c.mchid}
	urlPath := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s/close", outTradeNo)
	return c.postV3(ctx, urlPath, body, nil)
}
