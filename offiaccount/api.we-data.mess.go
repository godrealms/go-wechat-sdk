package offiaccount

// GetUpstreamMsg 获取消息发送概况数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，与endDate差值小于7天
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUpstreamMsg(beginDate, endDate string) (*GetUpstreamMsgResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsg"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgMonth 获取消息发送月数据
// beginDate: 起始日期(格式: yyyy-MM-dd)
// endDate: 结束日期(必须为同一天)
func (c *Client) GetUpstreamMsgMonth(beginDate, endDate string) (*GetUpstreamMsgResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsgmonth"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgDistWeek 获取消息发送分布周数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，跨度不超过15天
// endDate: 结束日期
func (c *Client) GetUpstreamMsgDistWeek(beginDate, endDate string) (*GetUpstreamMsgDistResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsgdistweek"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgDistResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgDistMonth 获取消息发送分布月数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，跨度不超过15天
// endDate: 结束日期
func (c *Client) GetUpstreamMsgDistMonth(beginDate, endDate string) (*GetUpstreamMsgDistResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsgdistmonth"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgDistResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgHour 获取消息发送分时数据
// beginDate: 起始日期(格式: yyyy-MM-dd)
// endDate: 结束日期(必须为同一天)
func (c *Client) GetUpstreamMsgHour(beginDate, endDate string) (*GetUpstreamMsgHourResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsghour"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgHourResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgWeek 获取消息发送周数据
// beginDate: 起始日期(格式: yyyy-MM-dd)
// endDate: 结束日期(必须为同一天)
func (c *Client) GetUpstreamMsgWeek(beginDate, endDate string) (*GetUpstreamMsgResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsgweek"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgDist 获取消息发送分布数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，跨度不超过15天
// endDate: 结束日期
func (c *Client) GetUpstreamMsgDist(beginDate, endDate string) (*GetUpstreamMsgDistResult, error) {
	// 构造请求URL
	path := "/datacube/getupstreammsgdist"

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUpstreamMsgDistResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
