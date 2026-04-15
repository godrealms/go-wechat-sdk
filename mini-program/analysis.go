package mini_program

import (
	"context"
	"fmt"
)

// requireDateRange validates that both begin_date and end_date are non-empty.
// Format is NOT validated client-side (WeChat accepts YYYYMMDD; server is
// authoritative on the exact format and lookback window).
func requireDateRange(method, beginDate, endDate string) error {
	if beginDate == "" || endDate == "" {
		return fmt.Errorf("mini_program: %s: beginDate and endDate are required", method)
	}
	return nil
}

// AnalysisDateReq is the common date-range request for data-analysis endpoints.
type AnalysisDateReq struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
}

// DailySummaryItem holds a single entry from the daily-summary trend response.
type DailySummaryItem struct {
	RefDate    string `json:"ref_date"`
	VisitTotal int    `json:"visit_total"`
	SharePV    int    `json:"share_pv"`
	ShareUV    int    `json:"share_uv"`
}

// GetDailySummaryResp is the response for the daily-summary trend API.
type GetDailySummaryResp struct {
	List []DailySummaryItem `json:"list"`
}

// VisitPageItem holds a single page's visit metrics.
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

// GetVisitPageResp is the response for the visit-page data API.
type GetVisitPageResp struct {
	RefDate string          `json:"ref_date"`
	List    []VisitPageItem `json:"list"`
}

// DailyVisitTrendItem holds a single day's visit trend metrics.
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

// GetDailyVisitTrendResp is the response for the daily visit trend API.
type GetDailyVisitTrendResp struct {
	List []DailyVisitTrendItem `json:"list"`
}

// GetDailySummary returns daily summary trend data for the given date range.
func (c *Client) GetDailySummary(ctx context.Context, beginDate, endDate string) (*GetDailySummaryResp, error) {
	if err := requireDateRange("GetDailySummary", beginDate, endDate); err != nil {
		return nil, err
	}
	body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
	var resp GetDailySummaryResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailysummarytrend", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVisitPage returns per-page visit data for the given date range.
func (c *Client) GetVisitPage(ctx context.Context, beginDate, endDate string) (*GetVisitPageResp, error) {
	if err := requireDateRange("GetVisitPage", beginDate, endDate); err != nil {
		return nil, err
	}
	body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
	var resp GetVisitPageResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappidvisitpage", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDailyVisitTrend returns daily visit trend data for the given date range.
func (c *Client) GetDailyVisitTrend(ctx context.Context, beginDate, endDate string) (*GetDailyVisitTrendResp, error) {
	if err := requireDateRange("GetDailyVisitTrend", beginDate, endDate); err != nil {
		return nil, err
	}
	body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
	var resp GetDailyVisitTrendResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailyvisittrend", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
