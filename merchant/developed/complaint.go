package pay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Consumer complaint APIs: https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter10_2_11.shtml
//
// This file provides a thin wrapper around the WeChat Pay complaint management API.

// ListComplaints returns a paginated list of consumer complaint orders.
// beginDate and endDate must be in yyyy-MM-dd format.
func (c *Client) ListComplaints(ctx context.Context, beginDate, endDate string, offset, limit int) (map[string]any, error) {
	if beginDate == "" || endDate == "" {
		return nil, fmt.Errorf("pay: beginDate and endDate are required")
	}
	query := url.Values{
		"begin_date": {beginDate},
		"end_date":   {endDate},
		"offset":     {fmt.Sprintf("%d", offset)},
		"limit":      {fmt.Sprintf("%d", limit)},
	}
	result := map[string]any{}
	if err := c.doV3(ctx, http.MethodGet, "/v3/merchant-service/complaints-v2", query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetComplaint returns the details of a single consumer complaint order.
func (c *Client) GetComplaint(ctx context.Context, complaintId string) (map[string]any, error) {
	if complaintId == "" {
		return nil, fmt.Errorf("pay: complaintId is required")
	}
	urlPath := fmt.Sprintf("/v3/merchant-service/complaints-v2/%s", complaintId)
	result := map[string]any{}
	if err := c.getV3(ctx, urlPath, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResponseComplaint 回复用户。
func (c *Client) ResponseComplaint(ctx context.Context, complaintId string, body any) error {
	if complaintId == "" {
		return fmt.Errorf("pay: complaintId is required")
	}
	urlPath := fmt.Sprintf("/v3/merchant-service/complaints-v2/%s/response", complaintId)
	return c.postV3(ctx, urlPath, body, nil)
}

// CompleteComplaint 反馈处理完成。
func (c *Client) CompleteComplaint(ctx context.Context, complaintId string) error {
	if complaintId == "" {
		return fmt.Errorf("pay: complaintId is required")
	}
	urlPath := fmt.Sprintf("/v3/merchant-service/complaints-v2/%s/complete", complaintId)
	return c.postV3(ctx, urlPath, map[string]string{}, nil)
}
