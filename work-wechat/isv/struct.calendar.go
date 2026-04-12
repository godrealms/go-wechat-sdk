package isv

// Calendar 日历对象。
type Calendar struct {
	CalID       string          `json:"cal_id,omitempty"`
	Organizer   string          `json:"organizer"`
	Summary     string          `json:"summary"`
	Color       string          `json:"color,omitempty"`
	Description string          `json:"description,omitempty"`
	Shares      []CalendarShare `json:"shares,omitempty"`
}

// CalendarShare 日历共享对象。
type CalendarShare struct {
	UserID string `json:"userid"`
}

// CreateCalendarReq 创建日历请求。
type CreateCalendarReq struct {
	Calendar Calendar `json:"calendar"`
}

// CreateCalendarResp 创建日历响应。
type CreateCalendarResp struct {
	CalID string `json:"cal_id"`
}

// UpdateCalendarReq 更新日历请求。
type UpdateCalendarReq struct {
	Calendar Calendar `json:"calendar"`
}

// GetCalendarReq 获取日历详情请求。
type GetCalendarReq struct {
	CalIDList []string `json:"cal_id_list"`
}

// GetCalendarResp 获取日历详情响应。
type GetCalendarResp struct {
	CalendarList []Calendar `json:"calendar_list"`
}

// DeleteCalendarReq 删除日历请求。
type DeleteCalendarReq struct {
	CalID string `json:"cal_id"`
}

// Schedule 日程对象。
type Schedule struct {
	ScheduleID  string             `json:"schedule_id,omitempty"`
	Organizer   string             `json:"organizer"`
	Summary     string             `json:"summary"`
	Description string             `json:"description,omitempty"`
	StartTime   int64              `json:"start_time"`
	EndTime     int64              `json:"end_time"`
	Location    string             `json:"location,omitempty"`
	CalID       string             `json:"cal_id,omitempty"`
	Attendees   []ScheduleAttendee `json:"attendees,omitempty"`
	Reminders   *ScheduleReminder  `json:"reminders,omitempty"`
}

// ScheduleAttendee 日程参与人。
type ScheduleAttendee struct {
	UserID string `json:"userid"`
}

// ScheduleReminder 日程提醒。
type ScheduleReminder struct {
	IsRemind     int `json:"is_remind"`
	RemindBefore *int `json:"remind_before_event_secs,omitempty"`
}

// CreateScheduleReq 创建日程请求。
type CreateScheduleReq struct {
	Schedule Schedule `json:"schedule"`
}

// CreateScheduleResp 创建日程响应。
type CreateScheduleResp struct {
	ScheduleID string `json:"schedule_id"`
}

// UpdateScheduleReq 更新日程请求。
type UpdateScheduleReq struct {
	Schedule Schedule `json:"schedule"`
}

// GetScheduleReq 获取日程详情请求。
type GetScheduleReq struct {
	ScheduleIDList []string `json:"schedule_id_list"`
}

// GetScheduleResp 获取日程详情响应。
type GetScheduleResp struct {
	ScheduleList []Schedule `json:"schedule_list"`
}

// DeleteScheduleReq 删除日程请求。
type DeleteScheduleReq struct {
	ScheduleID string `json:"schedule_id"`
}

// GetScheduleByCalendarReq 获取日历下日程列表请求。
type GetScheduleByCalendarReq struct {
	CalID  string `json:"cal_id"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// GetScheduleByCalendarResp 获取日历下日程列表响应。
type GetScheduleByCalendarResp struct {
	ScheduleList []Schedule `json:"schedule_list"`
}
