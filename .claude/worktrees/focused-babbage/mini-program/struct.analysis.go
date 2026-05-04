package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// DailyVisitTrendItem represents one day of visit trend data
type DailyVisitTrendItem struct {
	RefDate         string  `json:"ref_date"`
	SessionCnt      int64   `json:"session_cnt"`
	VisitPv         int64   `json:"visit_pv"`
	VisitUv         int64   `json:"visit_uv"`
	VisitUvNew      int64   `json:"visit_uv_new"`
	StayTimeSession float64 `json:"stay_time_session"`
	VisitDepth      float64 `json:"visit_depth"`
}

// VisitTrendResult is the result of GetDailyVisitTrend / GetWeeklyVisitTrend / GetMonthlyVisitTrend
type VisitTrendResult struct {
	core.Resp
	List []*DailyVisitTrendItem `json:"list"`
}

// DailyRetainInfo contains new and active user retain info
type DailyRetainInfo struct {
	RefDate    string `json:"ref_date"`
	VisitUvNew int64  `json:"visit_uv_new"`
	VisitUv    int64  `json:"visit_uv"`
}

// UserRetainResult is the result of GetDailyRetain / GetWeeklyRetain / GetMonthlyRetain
type UserRetainResult struct {
	core.Resp
	RefDate    string             `json:"ref_date"`
	VisitUvNew []*DailyRetainInfo `json:"visit_uv_new"`
	VisitUv    []*DailyRetainInfo `json:"visit_uv"`
}

// VisitPageItem represents one page's visit statistics
type VisitPageItem struct {
	PagePath       string  `json:"page_path"`
	PageVisitPv    int64   `json:"page_visit_pv"`
	PageVisitUv    int64   `json:"page_visit_uv"`
	PageStayTimeUv float64 `json:"page_staytime_uv"`
	EntrypagePv    int64   `json:"entrypage_pv"`
	ExitpagePv     int64   `json:"exitpage_pv"`
	PageSharePv    int64   `json:"page_share_pv"`
	PageShareUv    int64   `json:"page_share_uv"`
}

// VisitPageResult is the result of GetVisitPage
type VisitPageResult struct {
	core.Resp
	List []*VisitPageItem `json:"list"`
}

// AnalysisDateRequest is the common date range request for analysis APIs
type AnalysisDateRequest struct {
	BeginDate string `json:"begin_date"` // format: 20170313
	EndDate   string `json:"end_date"`   // format: 20170313
}

// UserPortraitItem represents user attribute distribution
type UserPortraitItem struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// UserPortraitResult is the result of GetUserPortrait
type UserPortraitResult struct {
	core.Resp
	RefDate string              `json:"ref_date"`
	VisitUv *UserPortraitDetail `json:"visit_uv"`
	ShareUv *UserPortraitDetail `json:"share_uv"`
}

// UserPortraitDetail contains demographic breakdown
type UserPortraitDetail struct {
	Province  []*UserPortraitItem `json:"province"`
	City      []*UserPortraitItem `json:"city"`
	Genders   []*UserPortraitItem `json:"genders"`
	Platforms []*UserPortraitItem `json:"platforms"`
	Devices   []*UserPortraitItem `json:"devices"`
	Ages      []*UserPortraitItem `json:"ages"`
}

// PerformanceQueryRequest is the request for GetPerformanceData
type PerformanceQueryRequest struct {
	CommonQuery *PerformanceCommonQuery `json:"commonQuery"`
	Queries     []*PerformanceQuery     `json:"queries"`
}

// PerformanceCommonQuery contains common query parameters
type PerformanceCommonQuery struct {
	AppId string `json:"appid"`
}

// PerformanceQuery specifies a single metric to query
type PerformanceQuery struct {
	Metric    string `json:"metric"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

// PerformanceDataResult is the result of GetPerformanceData
type PerformanceDataResult struct {
	core.Resp
	Data []*PerformanceDataItem `json:"data"`
}

// PerformanceDataItem represents one metric's data
type PerformanceDataItem struct {
	Metric string                  `json:"metric"`
	Data   []*PerformanceDataPoint `json:"data"`
}

// PerformanceDataPoint is a single time-series data point
type PerformanceDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}
