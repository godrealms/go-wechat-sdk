package offiaccount

import "fmt"

// GetUserSummary 获取用户增减数据
// beginDate: 起始日期(格式yyyy-MM-dd)
// endDate: 结束日期(最大跨度7天)
func (c *Client) GetUserSummary(beginDate, endDate string) (*GetUserSummaryResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/datacube/getusersummary?access_token=%s", c.GetAccessToken())

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserSummaryResult
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCumulate 获取累计用户数据
// beginDate: 起始日期(格式yyyy-MM-dd)
// endDate: 结束日期(最大跨度7天)
func (c *Client) GetUserCumulate(beginDate, endDate string) (*GetUserCumulateResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/datacube/getusercumulate?access_token=%s", c.GetAccessToken())

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserCumulateResult
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
