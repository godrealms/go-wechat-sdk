package offiaccount

import (
	"context"
	"fmt"
)

// GetInterfaceSummary 获取接口分析数据
// beginDate: 起始日期(YYYY-MM-DD格式)
// endDate: 结束日期(最大时间跨度30天)
func (c *Client) GetInterfaceSummary(ctx context.Context, beginDate, endDate string) (*GetInterfaceSummaryResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/datacube/getinterfacesummary?access_token=%s", token)

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetInterfaceSummaryResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInterfaceSummaryHour 获取接口分析分时数据
// beginDate: 起始日期(YYYY-MM-DD格式)
// endDate: 结束日期(最大时间跨度1天)
func (c *Client) GetInterfaceSummaryHour(ctx context.Context, beginDate, endDate string) (*GetInterfaceSummaryHourResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/datacube/getinterfacesummaryhour?access_token=%s", token)

	// 构造请求体
	req := &GetDataRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
	}

	// 发送请求
	var result GetInterfaceSummaryHourResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
