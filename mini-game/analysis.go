package mini_game

import "context"

// AnalysisDateReq specifies the date range for Mini Game data analysis queries.
type AnalysisDateReq struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
}

// DailySummaryItem holds the aggregated visit and share metrics for a single day.
type DailySummaryItem struct {
	RefDate    string `json:"ref_date"`
	VisitTotal int64  `json:"visit_total"`
	SharePV    int64  `json:"share_pv"`
	ShareUV    int64  `json:"share_uv"`
}

// GetDailySummaryResp is the response returned by GetDailySummary.
type GetDailySummaryResp struct {
	List []DailySummaryItem `json:"list"`
}

// DailyRetainItem represents a single retention data point keyed by date offset.
type DailyRetainItem struct {
	DateKey string `json:"date_key"`
	Value   int    `json:"value"`
}

// GetDailyRetainResp is the response returned by GetDailyRetain.
type GetDailyRetainResp struct {
	RefDate    string            `json:"ref_date"`
	VisitUVNew []DailyRetainItem `json:"visit_uv_new"`
	VisitUV    []DailyRetainItem `json:"visit_uv"`
}

// GetDailySummary retrieves the daily visit summary trend for the Mini Game within the given date range.
func (c *Client) GetDailySummary(ctx context.Context, req *AnalysisDateReq) (*GetDailySummaryResp, error) {
	var resp GetDailySummaryResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailysummarytrend", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDailyRetain retrieves the daily user retention data for the Mini Game within the given date range.
func (c *Client) GetDailyRetain(ctx context.Context, req *AnalysisDateReq) (*GetDailyRetainResp, error) {
	var resp GetDailyRetainResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailyretaininfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
