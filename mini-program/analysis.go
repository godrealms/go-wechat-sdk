package mini_program

import "context"

// AnalysisDateReq 数据分析通用日期请求。
type AnalysisDateReq struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
}

// DailySummaryItem 概况趋势数据项。
type DailySummaryItem struct {
	RefDate    string `json:"ref_date"`
	VisitTotal int    `json:"visit_total"`
	SharePV    int    `json:"share_pv"`
	ShareUV    int    `json:"share_uv"`
}

// GetDailySummaryResp 概况趋势响应。
type GetDailySummaryResp struct {
	List []DailySummaryItem `json:"list"`
}

// VisitPageItem 页面访问数据项。
type VisitPageItem struct {
	PagePath       string  `json:"page_path"`
	PageVisitPV    int     `json:"page_visit_pv"`
	PageVisitUV    int     `json:"page_visit_uv"`
	PageStaytimePV float64 `json:"page_staytime_pv"`
	EntryPagePV    int     `json:"entrypage_pv"`
	ExitPagePV     int     `json:"exitpage_pv"`
	PageSharePV    int     `json:"page_share_pv"`
	PageShareUV    int     `json:"page_share_uv"`
}

// GetVisitPageResp 页面访问数据响应。
type GetVisitPageResp struct {
	RefDate string          `json:"ref_date"`
	List    []VisitPageItem `json:"list"`
}

// DailyVisitTrendItem 日访问趋势数据项。
type DailyVisitTrendItem struct {
	RefDate         string  `json:"ref_date"`
	SessionCnt      int     `json:"session_cnt"`
	VisitPV         int     `json:"visit_pv"`
	VisitUV         int     `json:"visit_uv"`
	VisitUVNew      int     `json:"visit_uv_new"`
	StayTimeUV      float64 `json:"stay_time_uv"`
	StayTimeSession float64 `json:"stay_time_session"`
	VisitDepth      float64 `json:"visit_depth"`
}

// GetDailyVisitTrendResp 日访问趋势响应。
type GetDailyVisitTrendResp struct {
	List []DailyVisitTrendItem `json:"list"`
}

// GetDailySummary 获取概况趋势数据。
func (c *Client) GetDailySummary(ctx context.Context, beginDate, endDate string) (*GetDailySummaryResp, error) {
	body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
	var resp GetDailySummaryResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailysummarytrend", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVisitPage 获取页面访问数据。
func (c *Client) GetVisitPage(ctx context.Context, beginDate, endDate string) (*GetVisitPageResp, error) {
	body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
	var resp GetVisitPageResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappidvisitpage", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDailyVisitTrend 获取日访问趋势。
func (c *Client) GetDailyVisitTrend(ctx context.Context, beginDate, endDate string) (*GetDailyVisitTrendResp, error) {
	body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
	var resp GetDailyVisitTrendResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailyvisittrend", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
