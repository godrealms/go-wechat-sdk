package pay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// 分账接口：https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter8_1_1.shtml
//
// 本文件提供常用分账 API 的薄封装，请求体/响应体采用 map[string]any，
// 既能应对微信不定期新增字段，也避免被 types 限死。
// 如需强类型，可自行在 types 目录下定义结构体传入。

// ProfitSharingOrder 创建分账单。body 必须包含非空 out_order_no 作为幂等键。
// https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter8_1_1.shtml
func (c *Client) ProfitSharingOrder(ctx context.Context, body any) (map[string]any, error) {
	if err := requireIdempotencyKey(body, "out_order_no"); err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/profitsharing/orders", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingQueryOrder 查询分账结果。
func (c *Client) ProfitSharingQueryOrder(ctx context.Context, outOrderNo, transactionId string) (map[string]any, error) {
	if outOrderNo == "" || transactionId == "" {
		return nil, fmt.Errorf("pay: outOrderNo and transactionId are required")
	}
	urlPath := fmt.Sprintf("/v3/profitsharing/orders/%s", outOrderNo)
	query := url.Values{"transaction_id": {transactionId}}
	result := map[string]any{}
	if err := c.doV3(ctx, http.MethodGet, urlPath, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingReturn 请求分账回退。body 必须包含非空 out_return_no。
func (c *Client) ProfitSharingReturn(ctx context.Context, body any) (map[string]any, error) {
	if err := requireIdempotencyKey(body, "out_return_no"); err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/profitsharing/return-orders", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingQueryReturn 查询分账回退。
func (c *Client) ProfitSharingQueryReturn(ctx context.Context, outReturnNo, outOrderNo string) (map[string]any, error) {
	if outReturnNo == "" || outOrderNo == "" {
		return nil, fmt.Errorf("pay: outReturnNo and outOrderNo are required")
	}
	urlPath := fmt.Sprintf("/v3/profitsharing/return-orders/%s", outReturnNo)
	query := url.Values{"out_order_no": {outOrderNo}}
	result := map[string]any{}
	if err := c.doV3(ctx, http.MethodGet, urlPath, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingUnfreeze 解冻剩余资金。body 必须包含非空 out_order_no。
func (c *Client) ProfitSharingUnfreeze(ctx context.Context, body any) (map[string]any, error) {
	if err := requireIdempotencyKey(body, "out_order_no"); err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/profitsharing/orders/unfreeze", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingMerchantAmounts 查询剩余待分金额。
func (c *Client) ProfitSharingMerchantAmounts(ctx context.Context, transactionId string) (map[string]any, error) {
	if transactionId == "" {
		return nil, fmt.Errorf("pay: transactionId is required")
	}
	urlPath := fmt.Sprintf("/v3/profitsharing/transactions/%s/amounts", transactionId)
	result := map[string]any{}
	if err := c.getV3(ctx, urlPath, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingAddReceiver 添加分账接收方。
func (c *Client) ProfitSharingAddReceiver(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/profitsharing/receivers/add", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingDeleteReceiver 删除分账接收方。
func (c *Client) ProfitSharingDeleteReceiver(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/profitsharing/receivers/delete", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingBills 申请分账账单（按日）。
func (c *Client) ProfitSharingBills(ctx context.Context, billDate, tarType string) (map[string]any, error) {
	if billDate == "" {
		return nil, fmt.Errorf("pay: billDate is required")
	}
	query := url.Values{"bill_date": {billDate}}
	if tarType != "" {
		query.Set("tar_type", tarType)
	}
	result := map[string]any{}
	if err := c.getV3(ctx, "/v3/profitsharing/bills", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
