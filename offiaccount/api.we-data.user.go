package offiaccount

import (
	"context"
	"fmt"
)

// GetUserSummary 获取用户增减数据
// beginDate: 起始日期(格式yyyy-MM-dd)
// endDate: 结束日期(最大跨度7天)
func (c *Client) GetUserSummary(ctx context.Context, beginDate, endDate string) (*GetUserSummaryResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/datacube/getusersummary?access_token=%s", token)

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserSummaryResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCumulate 获取累计用户数据
// beginDate: 起始日期(格式yyyy-MM-dd)
// endDate: 结束日期(最大跨度7天)
func (c *Client) GetUserCumulate(ctx context.Context, beginDate, endDate string) (*GetUserCumulateResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/datacube/getusercumulate?access_token=%s", token)

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetUserCumulateResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
