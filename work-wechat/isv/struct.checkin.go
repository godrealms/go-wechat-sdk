package isv

// GetCheckinDataReq is the request for GetCheckinData.
type GetCheckinDataReq struct {
	OpenCheckinDataType int      `json:"opencheckindatatype"`
	StartTime           int64    `json:"starttime"`
	EndTime             int64    `json:"endtime"`
	UserIDList          []string `json:"useridlist"`
}

// CheckinData represents a single check-in record.
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

// GetCheckinDataResp is the response from GetCheckinData.
type GetCheckinDataResp struct {
	CheckinData []CheckinData `json:"checkindata"`
}

// GetCheckinOptionReq is the request for GetCheckinOption.
type GetCheckinOptionReq struct {
	DateTime   int64    `json:"datetime"`
	UserIDList []string `json:"useridlist"`
}

// CheckinOption holds the check-in rule for a single user.
type CheckinOption struct {
	UserID string       `json:"userid"`
	Group  CheckinGroup `json:"group"`
}

// CheckinGroup holds the check-in rule group information.
type CheckinGroup struct {
	GroupID   int    `json:"groupid"`
	GroupName string `json:"groupname"`
	GroupType int    `json:"grouptype"`
}

// GetCheckinOptionResp is the response from GetCheckinOption.
type GetCheckinOptionResp struct {
	Info []CheckinOption `json:"info"`
}

// GetCheckinDayDataReq is the request for GetCheckinDayData.
type GetCheckinDayDataReq struct {
	StartTime  int64    `json:"starttime"`
	EndTime    int64    `json:"endtime"`
	UserIDList []string `json:"useridlist"`
}

// CheckinDayData holds the daily check-in report for a single user.
type CheckinDayData struct {
	BaseInfo    CheckinDayBase    `json:"base_info"`
	SummaryInfo CheckinDaySummary `json:"summary_info"`
}

// CheckinDayBase holds the basic identification fields of a daily check-in report entry.
type CheckinDayBase struct {
	Date   int64  `json:"date"`
	Name   string `json:"name"`
	NameEx string `json:"name_ex"`
	AcctID string `json:"acctid"`
}

// CheckinDaySummary holds the statistical summary of a daily check-in report.
type CheckinDaySummary struct {
	CheckinCount    int `json:"checkin_count"`
	RegularWorkSec  int `json:"regular_work_sec"`
	StandardWorkSec int `json:"standard_work_sec"`
}

// GetCheckinDayDataResp is the response from GetCheckinDayData.
type GetCheckinDayDataResp struct {
	Datas []CheckinDayData `json:"datas"`
}

// GetCheckinMonthDataReq is the request for GetCheckinMonthData.
type GetCheckinMonthDataReq struct {
	StartTime  int64    `json:"starttime"`
	EndTime    int64    `json:"endtime"`
	UserIDList []string `json:"useridlist"`
}

// CheckinMonthData holds the monthly check-in report for a single user.
type CheckinMonthData struct {
	BaseInfo    CheckinDayBase      `json:"base_info"`
	SummaryInfo CheckinMonthSummary `json:"summary_info"`
}

// CheckinMonthSummary holds the statistical summary of a monthly check-in report.
type CheckinMonthSummary struct {
	WorkDays       int `json:"work_days"`
	RegularWorkSec int `json:"regular_work_sec"`
	ExceptDays     int `json:"except_days"`
}

// GetCheckinMonthDataResp is the response from GetCheckinMonthData.
type GetCheckinMonthDataResp struct {
	Datas []CheckinMonthData `json:"datas"`
}
