package mini_program

import "fmt"

// GetDailyVisitTrend 获取用户访问小程序日趋势
// POST /datacube/getweanalysisappiddailyvisittrend
func (c *Client) GetDailyVisitTrend(req *AnalysisDateRequest) (*VisitTrendResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappiddailyvisittrend?access_token=%s", c.GetAccessToken())
	result := &VisitTrendResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetWeeklyVisitTrend 获取用户访问小程序周趋势
// POST /datacube/getweanalysisappidweeklyvisittrend
func (c *Client) GetWeeklyVisitTrend(req *AnalysisDateRequest) (*VisitTrendResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidweeklyvisittrend?access_token=%s", c.GetAccessToken())
	result := &VisitTrendResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetMonthlyVisitTrend 获取用户访问小程序月趋势
// POST /datacube/getweanalysisappidmonthlyvisittrend
func (c *Client) GetMonthlyVisitTrend(req *AnalysisDateRequest) (*VisitTrendResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidmonthlyvisittrend?access_token=%s", c.GetAccessToken())
	result := &VisitTrendResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetDailyRetain 获取用户小程序访问日留存
// POST /datacube/getweanalysisappiddailyretaininfo
func (c *Client) GetDailyRetain(req *AnalysisDateRequest) (*UserRetainResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappiddailyretaininfo?access_token=%s", c.GetAccessToken())
	result := &UserRetainResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetWeeklyRetain 获取用户小程序访问周留存
// POST /datacube/getweanalysisappidweeklyretaininfo
func (c *Client) GetWeeklyRetain(req *AnalysisDateRequest) (*UserRetainResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidweeklyretaininfo?access_token=%s", c.GetAccessToken())
	result := &UserRetainResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetMonthlyRetain 获取用户小程序访问月留存
// POST /datacube/getweanalysisappidmonthlyretaininfo
func (c *Client) GetMonthlyRetain(req *AnalysisDateRequest) (*UserRetainResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidmonthlyretaininfo?access_token=%s", c.GetAccessToken())
	result := &UserRetainResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetVisitPage 获取小程序访问页面数据
// POST /datacube/getweanalysisappidvisitpage
func (c *Client) GetVisitPage(req *AnalysisDateRequest) (*VisitPageResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidvisitpage?access_token=%s", c.GetAccessToken())
	result := &VisitPageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetUserPortrait 获取小程序新增或活跃用户的画像分布数据
// POST /datacube/getweanalysisappiduserportrait
func (c *Client) GetUserPortrait(req *AnalysisDateRequest) (*UserPortraitResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappiduserportrait?access_token=%s", c.GetAccessToken())
	result := &UserPortraitResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetPerformanceData 小程序性能监控数据
// POST /wxaapi/log/get_performance
func (c *Client) GetPerformanceData(req *PerformanceQueryRequest) (*PerformanceDataResult, error) {
	path := fmt.Sprintf("/wxaapi/log/get_performance?access_token=%s", c.GetAccessToken())
	result := &PerformanceDataResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
