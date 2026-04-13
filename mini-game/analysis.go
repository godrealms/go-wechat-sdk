package mini_game

import "context"

// AnalysisDateReq 数据分析日期请求。
type AnalysisDateReq struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
}

type DailySummaryItem struct {
	RefDate    string `json:"ref_date"`
	VisitTotal int64  `json:"visit_total"`
	SharePV    int64  `json:"share_pv"`
	ShareUV    int64  `json:"share_uv"`
}
type GetDailySummaryResp struct {
	List []DailySummaryItem `json:"list"`
}

type DailyRetainItem struct {
	DateKey string `json:"date_key"`
	Value   int    `json:"value"`
}
type GetDailyRetainResp struct {
	RefDate    string            `json:"ref_date"`
	VisitUVNew []DailyRetainItem `json:"visit_uv_new"`
	VisitUV    []DailyRetainItem `json:"visit_uv"`
}

func (c *Client) GetDailySummary(ctx context.Context, req *AnalysisDateReq) (*GetDailySummaryResp, error) {
	var resp GetDailySummaryResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailysummarytrend", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetDailyRetain(ctx context.Context, req *AnalysisDateReq) (*GetDailyRetainResp, error) {
	var resp GetDailyRetainResp
	if err := c.doPost(ctx, "/datacube/getweanalysisappiddailyretaininfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
