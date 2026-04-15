package offiaccount

import (
	"context"
	"fmt"
)

// GetUpstreamMsg 获取消息发送概况数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，与endDate差值小于7天
// endDate: 结束日期(最大值为昨日)
func (c *Client) GetUpstreamMsg(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsg?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgMonth 获取消息发送月数据
// beginDate: 起始日期(格式: yyyy-MM-dd)
// endDate: 结束日期(必须为同一天)
func (c *Client) GetUpstreamMsgMonth(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsgmonth?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgDistWeek 获取消息发送分布周数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，跨度不超过15天
// endDate: 结束日期
func (c *Client) GetUpstreamMsgDistWeek(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgDistResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsgdistweek?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgDistResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgDistMonth 获取消息发送分布月数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，跨度不超过15天
// endDate: 结束日期
func (c *Client) GetUpstreamMsgDistMonth(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgDistResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsgdistmonth?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgDistResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgHour 获取消息发送分时数据
// beginDate: 起始日期(格式: yyyy-MM-dd)
// endDate: 结束日期(必须为同一天)
func (c *Client) GetUpstreamMsgHour(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgHourResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsghour?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgHourResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgWeek 获取消息发送周数据
// beginDate: 起始日期(格式: yyyy-MM-dd)
// endDate: 结束日期(必须为同一天)
func (c *Client) GetUpstreamMsgWeek(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsgweek?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUpstreamMsgDist 获取消息发送分布数据
// beginDate: 起始日期(格式: yyyy-MM-dd)，跨度不超过15天
// endDate: 结束日期
func (c *Client) GetUpstreamMsgDist(ctx context.Context, beginDate, endDate string) (*GetUpstreamMsgDistResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/datacube/getupstreammsgdist?access_token=%s", token)

	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	var result GetUpstreamMsgDistResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
