package developed

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// 商家转账到零钱接口：https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter4_3_1.shtml
//
// 本文件提供商家转账 API 的薄封装，请求体/响应体采用 map[string]any，
// 与 profit_sharing.go 保持一致。

// CreateTransferBatch 发起商家转账。
func (c *Client) CreateTransferBatch(ctx context.Context, body any) (map[string]any, error) {
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
