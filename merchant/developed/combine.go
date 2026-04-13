package pay

import (
	"context"
	"fmt"
)

// 合单支付接口：https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_3_1.shtml
//
// 本文件只封装合单 JSAPI/APP/H5/Native 下单、查询、关单。
// 参数采用 map[string]any / 任意 JSON，便于跟随微信文档灵活调整。

// CombineTransactionsJsapi 合单 JSAPI 下单。
func (c *Client) CombineTransactionsJsapi(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/combine-transactions/jsapi", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CombineTransactionsApp 合单 APP 下单。
func (c *Client) CombineTransactionsApp(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/combine-transactions/app", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CombineTransactionsH5 合单 H5 下单。
func (c *Client) CombineTransactionsH5(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/combine-transactions/h5", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CombineTransactionsNative 合单 Native 下单。
func (c *Client) CombineTransactionsNative(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/combine-transactions/native", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryCombineOrder 合单订单查询（按合单订单号）。
func (c *Client) QueryCombineOrder(ctx context.Context, combineOutTradeNo string) (map[string]any, error) {
	if combineOutTradeNo == "" {
		return nil, fmt.Errorf("pay: combineOutTradeNo is required")
	}
	urlPath := fmt.Sprintf("/v3/combine-transactions/out-trade-no/%s", combineOutTradeNo)
	result := map[string]any{}
	if err := c.getV3(ctx, urlPath, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CloseCombineOrder 合单订单关闭。
func (c *Client) CloseCombineOrder(ctx context.Context, combineOutTradeNo string, body any) error {
	if combineOutTradeNo == "" {
		return fmt.Errorf("pay: combineOutTradeNo is required")
	}
	urlPath := fmt.Sprintf("/v3/combine-transactions/out-trade-no/%s/close", combineOutTradeNo)
	return c.postV3(ctx, urlPath, body, nil)
}

// RefundsCombine 合单退款（实际上合单退款还是走普通 /v3/refund/domestic/refunds，
// 这里只提供一个语义别名；直接使用 Client.Refunds 也可）。
func (c *Client) RefundsCombine(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/refund/domestic/refunds", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

