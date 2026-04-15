package offiaccount

import (
	"context"
	"fmt"
)

// GetInvoiceInfo 查询报销发票信息
// req: 查询报销发票信息请求参数
func (c *Client) GetInvoiceInfo(ctx context.Context, req *GetInvoiceInfoRequest) (*GetInvoiceInfoResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/reimburse/getinvoiceinfo?access_token=%s", token)

	var result GetInvoiceInfoResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateInvoiceReimburseStatus 报销方更新发票状态
// req: 报销方更新发票状态请求参数
func (c *Client) UpdateInvoiceReimburseStatus(ctx context.Context, req *UpdateInvoiceReimburseStatusRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/reimburse/updateinvoicestatus?access_token=%s", token)

	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateInvoiceReimburseStatusBatch 批量更新报销发票状态
// req: 批量更新报销发票状态请求参数
func (c *Client) UpdateInvoiceReimburseStatusBatch(ctx context.Context, req *UpdateInvoiceReimburseStatusBatchRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/reimburse/updatestatusbatch?access_token=%s", token)

	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInvoiceBatch 批量获取报销发票信息
// req: 批量获取报销发票信息请求参数
func (c *Client) GetInvoiceBatch(ctx context.Context, req *GetInvoiceBatchRequest) (*GetInvoiceBatchResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/reimburse/getinvoicebatch?access_token=%s", token)

	var result GetInvoiceBatchResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
