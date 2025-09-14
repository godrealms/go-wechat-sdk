package offiaccount

// GetInterfaceSummary 获取接口分析数据
// beginDate: 起始日期(YYYY-MM-DD格式)
// endDate: 结束日期(最大时间跨度30天)
func (c *Client) GetInterfaceSummary(beginDate, endDate string) (*GetInterfaceSummaryResult, error) {
	// 构造请求URL
	path := "/datacube/getinterfacesummary"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetInterfaceSummaryResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInterfaceSummaryHour 获取接口分析分时数据
// beginDate: 起始日期(YYYY-MM-DD格式)
// endDate: 结束日期(最大时间跨度1天)
func (c *Client) GetInterfaceSummaryHour(beginDate, endDate string) (*GetInterfaceSummaryHourResult, error) {
	// 构造请求URL
	path := "/datacube/getinterfacesummaryhour"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetInterfaceSummaryHourResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
