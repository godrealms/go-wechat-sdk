package isv

// GetCheckinDataReq 获取打卡记录请求。
type GetCheckinDataReq struct {
	OpenCheckinDataType int      `json:"opencheckindatatype"`
	StartTime           int64    `json:"starttime"`
	EndTime             int64    `json:"endtime"`
	UserIDList          []string `json:"useridlist"`
}

// CheckinData 打卡记录。
type CheckinData struct {
	UserID         string `json:"userid"`
	GroupName      string `json:"groupname"`
	CheckinType    string `json:"checkin_type"`
	CheckinTime    int64  `json:"checkin_time"`
	ExceptionType  string `json:"exception_type"`
	LocationTitle  string `json:"location_title"`
	LocationDetail string `json:"location_detail"`
	Notes          string `json:"notes"`
}

// GetCheckinDataResp 获取打卡记录响应。
type GetCheckinDataResp struct {
	CheckinData []CheckinData `json:"checkindata"`
}

// GetCheckinOptionReq 获取打卡规则请求。
type GetCheckinOptionReq struct {
	DateTime   int64    `json:"datetime"`
	UserIDList []string `json:"useridlist"`
}

// CheckinOption 打卡规则。
type CheckinOption struct {
	UserID string       `json:"userid"`
	Group  CheckinGroup `json:"group"`
}

// CheckinGroup 打卡规则组。
type CheckinGroup struct {
	GroupID   int    `json:"groupid"`
	GroupName string `json:"groupname"`
	GroupType int    `json:"grouptype"`
}

// GetCheckinOptionResp 获取打卡规则响应。
type GetCheckinOptionResp struct {
	Info []CheckinOption `json:"info"`
}

// GetCheckinDayDataReq 获取打卡日报请求。
type GetCheckinDayDataReq struct {
	StartTime  int64    `json:"starttime"`
	EndTime    int64    `json:"endtime"`
	UserIDList []string `json:"useridlist"`
}

// CheckinDayData 打卡日报数据。
type CheckinDayData struct {
	BaseInfo    CheckinDayBase    `json:"base_info"`
	SummaryInfo CheckinDaySummary `json:"summary_info"`
}

// CheckinDayBase 日报基础信息。
type CheckinDayBase struct {
	Date   int64  `json:"date"`
	Name   string `json:"name"`
	NameEx string `json:"name_ex"`
	AcctID string `json:"acctid"`
}

// CheckinDaySummary 日报统计。
type CheckinDaySummary struct {
	CheckinCount    int `json:"checkin_count"`
	RegularWorkSec  int `json:"regular_work_sec"`
	StandardWorkSec int `json:"standard_work_sec"`
}

// GetCheckinDayDataResp 获取打卡日报响应。
type GetCheckinDayDataResp struct {
	Datas []CheckinDayData `json:"datas"`
}

// GetCheckinMonthDataReq 获取打卡月报请求。
type GetCheckinMonthDataReq struct {
	StartTime  int64    `json:"starttime"`
	EndTime    int64    `json:"endtime"`
	UserIDList []string `json:"useridlist"`
}

// CheckinMonthData 打卡月报数据。
type CheckinMonthData struct {
	BaseInfo    CheckinDayBase      `json:"base_info"`
	SummaryInfo CheckinMonthSummary `json:"summary_info"`
}

// CheckinMonthSummary 月报统计。
type CheckinMonthSummary struct {
	WorkDays       int `json:"work_days"`
	RegularWorkSec int `json:"regular_work_sec"`
	ExceptDays     int `json:"except_days"`
}

// GetCheckinMonthDataResp 获取打卡月报响应。
type GetCheckinMonthDataResp struct {
	Datas []CheckinMonthData `json:"datas"`
}
