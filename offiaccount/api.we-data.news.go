package offiaccount

// GetArticleSummary 获取图文群发每日数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetArticleSummary(beginDate, endDate string) (*GetArticleSummaryResult, error) {
	// 构造请求URL
	path := "/datacube/getarticlesummary"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetArticleSummaryResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserReadHour 获取图文统计分时数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserReadHour(beginDate, endDate string) (*GetUserReadHourResult, error) {
	// 构造请求URL
	path := "/datacube/getuserreadhour"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserReadHourResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserShareHour 获取图文分享转发分时数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserShareHour(beginDate, endDate string) (*GetUserShareHourResult, error) {
	// 构造请求URL
	path := "/datacube/getusersharehour"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserShareHourResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserRead 获取图文统计数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserRead(beginDate, endDate string) (*GetUserReadResult, error) {
	// 构造请求URL
	path := "/datacube/getuserread"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserReadResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetArticleTotal 获取图文群发总数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetArticleTotal(beginDate, endDate string) (*GetArticleTotalResult, error) {
	// 构造请求URL
	path := "/datacube/getarticletotal"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetArticleTotalResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserShare 获取图文分享转发数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserShare(beginDate, endDate string) (*GetUserShareResult, error) {
	// 构造请求URL
	path := "/datacube/getusershare"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserShareResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
