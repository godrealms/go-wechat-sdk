package pay

import (
	"context"
	"fmt"
	"net/url"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// QueryTransactionId queries a WeChat Pay order by the WeChat transaction ID.
// 通过微信支付订单号查询订单。
func (c *Client) QueryTransactionId(ctx context.Context, transactionId string) (*types.QueryResponse, error) {
	if transactionId == "" {
		return nil, fmt.Errorf("pay: transactionId is required")
	}
	urlPath := fmt.Sprintf("/v3/pay/transactions/id/%s", transactionId)
	resp := &types.QueryResponse{}
	if err := c.getV3(ctx, urlPath, url.Values{"mchid": {c.mchid}}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// QueryOutTradeNo queries a WeChat Pay order by the merchant's out_trade_no.
// 通过商户订单号查询订单。
func (c *Client) QueryOutTradeNo(ctx context.Context, outTradeNo string) (*types.QueryResponse, error) {
	if outTradeNo == "" {
		return nil, fmt.Errorf("pay: outTradeNo is required")
	}
	urlPath := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s", outTradeNo)
	resp := &types.QueryResponse{}
	if err := c.getV3(ctx, urlPath, url.Values{"mchid": {c.mchid}}, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
