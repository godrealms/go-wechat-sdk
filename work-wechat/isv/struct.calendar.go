package isv

// Calendar represents a calendar object.
type Calendar struct {
	CalID       string          `json:"cal_id,omitempty"`
	Organizer   string          `json:"organizer"`
	Summary     string          `json:"summary"`
	Color       string          `json:"color,omitempty"`
	Description string          `json:"description,omitempty"`
	Shares      []CalendarShare `json:"shares,omitempty"`
}

// CalendarShare represents a user with whom a calendar is shared.
type CalendarShare struct {
	UserID string `json:"userid"`
}

// CreateCalendarReq is the request for CreateCalendar.
type CreateCalendarReq struct {
	Calendar Calendar `json:"calendar"`
}

// CreateCalendarResp is the response from CreateCalendar.
type CreateCalendarResp struct {
	CalID string `json:"cal_id"`
}

// UpdateCalendarReq is the request for UpdateCalendar.
type UpdateCalendarReq struct {
	Calendar Calendar `json:"calendar"`
}

// GetCalendarReq is the request for GetCalendar.
type GetCalendarReq struct {
	CalIDList []string `json:"cal_id_list"`
}

// GetCalendarResp is the response from GetCalendar.
type GetCalendarResp struct {
	CalendarList []Calendar `json:"calendar_list"`
}

// DeleteCalendarReq is the request for DeleteCalendar.
type DeleteCalendarReq struct {
	CalID string `json:"cal_id"`
}

// Schedule represents a schedule event.
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

// ScheduleAttendee represents an attendee of a schedule event.
type ScheduleAttendee struct {
	UserID string `json:"userid"`
}

// ScheduleReminder holds the reminder configuration for a schedule event.
type ScheduleReminder struct {
	IsRemind     int  `json:"is_remind"`
	RemindBefore *int `json:"remind_before_event_secs,omitempty"`
}

// CreateScheduleReq is the request for CreateSchedule.
type CreateScheduleReq struct {
	Schedule Schedule `json:"schedule"`
}

// CreateScheduleResp is the response from CreateSchedule.
type CreateScheduleResp struct {
	ScheduleID string `json:"schedule_id"`
}

// UpdateScheduleReq is the request for UpdateSchedule.
type UpdateScheduleReq struct {
	Schedule Schedule `json:"schedule"`
}

// GetScheduleReq is the request for GetSchedule.
type GetScheduleReq struct {
	ScheduleIDList []string `json:"schedule_id_list"`
}

// GetScheduleResp is the response from GetSchedule.
type GetScheduleResp struct {
	ScheduleList []Schedule `json:"schedule_list"`
}

// DeleteScheduleReq is the request for DeleteSchedule.
type DeleteScheduleReq struct {
	ScheduleID string `json:"schedule_id"`
}

// GetScheduleByCalendarReq is the request for GetScheduleByCalendar.
type GetScheduleByCalendarReq struct {
	CalID  string `json:"cal_id"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// GetScheduleByCalendarResp is the response from GetScheduleByCalendar.
type GetScheduleByCalendarResp struct {
	ScheduleList []Schedule `json:"schedule_list"`
}
