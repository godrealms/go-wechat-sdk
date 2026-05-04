package pay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Merchant transfer to Change APIs: https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter4_3_1.shtml
//
// This file provides a thin wrapper around the WeChat Pay merchant transfer API.
// Request and response bodies use map[string]any, consistent with profit_sharing.go.

// CreateTransferBatch 发起商家转账。body 必须包含非空 out_batch_no——这是
// 微信支付为这个端点设计的幂等键,缺失或重复会导致重复打款风险。SDK 在
// 提交前做最小校验,具体的 batch_no 生成/持久化策略仍由调用方负责。
func (c *Client) CreateTransferBatch(ctx context.Context, body any) (map[string]any, error) {
	if err := requireIdempotencyKey(body, "out_batch_no"); err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/transfer/batches", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryTransferBatch 查询转账批次（通过微信批次单号）。
func (c *Client) QueryTransferBatch(ctx context.Context, batchId string, needQueryDetail bool, offset, limit int) (map[string]any, error) {
	if batchId == "" {
		return nil, fmt.Errorf("pay: batchId is required")
	}
	urlPath := fmt.Sprintf("/v3/transfer/batches/batch-id/%s", batchId)
	query := url.Values{}
	if needQueryDetail {
		query.Set("need_query_detail", "true")
	} else {
		query.Set("need_query_detail", "false")
	}
	query.Set("offset", fmt.Sprintf("%d", offset))
	query.Set("limit", fmt.Sprintf("%d", limit))
	result := map[string]any{}
	if err := c.doV3(ctx, http.MethodGet, urlPath, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryTransferDetail 查询转账明细（通过微信明细单号）。
func (c *Client) QueryTransferDetail(ctx context.Context, batchId, detailId string) (map[string]any, error) {
	if batchId == "" || detailId == "" {
		return nil, fmt.Errorf("pay: batchId and detailId are required")
	}
	urlPath := fmt.Sprintf("/v3/transfer/batches/batch-id/%s/details/detail-id/%s", batchId, detailId)
	result := map[string]any{}
	if err := c.getV3(ctx, urlPath, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
