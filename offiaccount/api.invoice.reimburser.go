package offiaccount

// GetInvoiceInfo 查询报销发票信息
// req: 查询报销发票信息请求参数
func (c *Client) GetInvoiceInfo(req *GetInvoiceInfoRequest) (*GetInvoiceInfoResult, error) {
	// 构造请求URL
	path := "/card/invoice/reimburse/getinvoiceinfo"

	// 发送请求
	var result GetInvoiceInfoResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateInvoiceReimburseStatus 报销方更新发票状态
// req: 报销方更新发票状态请求参数
func (c *Client) UpdateInvoiceReimburseStatus(req *UpdateInvoiceReimburseStatusRequest) (*Resp, error) {
	// 构造请求URL
	path := "/card/invoice/reimburse/updateinvoicestatus"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateInvoiceReimburseStatusBatch 批量更新报销发票状态
// req: 批量更新报销发票状态请求参数
func (c *Client) UpdateInvoiceReimburseStatusBatch(req *UpdateInvoiceReimburseStatusBatchRequest) (*Resp, error) {
	// 构造请求URL
	path := "/card/invoice/reimburse/updatestatusbatch"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInvoiceBatch 批量获取报销发票信息
// req: 批量获取报销发票信息请求参数
func (c *Client) GetInvoiceBatch(req *GetInvoiceBatchRequest) (*GetInvoiceBatchResult, error) {
	// 构造请求URL
	path := "/card/invoice/reimburse/getinvoicebatch"

	// 发送请求
	var result GetInvoiceBatchResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
