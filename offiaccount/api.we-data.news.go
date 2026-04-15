package offiaccount

import (
	"context"
	"fmt"
)

// GetArticleSummary 获取图文群发每日数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetArticleSummary(ctx context.Context, beginDate, endDate string) (*GetArticleSummaryResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getarticlesummary?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetArticleSummaryResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserReadHour 获取图文统计分时数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserReadHour(ctx context.Context, beginDate, endDate string) (*GetUserReadHourResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getuserreadhour?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUserReadHourResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserShareHour 获取图文分享转发分时数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserShareHour(ctx context.Context, beginDate, endDate string) (*GetUserShareHourResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getusersharehour?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUserShareHourResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserRead 获取图文统计数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserRead(ctx context.Context, beginDate, endDate string) (*GetUserReadResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getuserread?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUserReadResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetArticleTotal 获取图文群发总数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetArticleTotal(ctx context.Context, beginDate, endDate string) (*GetArticleTotalResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getarticletotal?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetArticleTotalResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserShare 获取图文分享转发数据
// beginDate: 起始日期(YYYY-MM-DD)
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUserShare(ctx context.Context, beginDate, endDate string) (*GetUserShareResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getusershare?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUserShareResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
