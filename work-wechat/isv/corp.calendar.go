package isv

import "context"

// CreateCalendar 创建日历。
func (cc *CorpClient) CreateCalendar(ctx context.Context, req *CreateCalendarReq) (*CreateCalendarResp, error) {
	var resp CreateCalendarResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/calendar/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCalendar 更新日历。
func (cc *CorpClient) UpdateCalendar(ctx context.Context, req *UpdateCalendarReq) error {
	return cc.doPost(ctx, "/cgi-bin/oa/calendar/update", req, nil)
}

// GetCalendar 获取日历详情。
func (cc *CorpClient) GetCalendar(ctx context.Context, calIDs []string) (*GetCalendarResp, error) {
	body := &GetCalendarReq{CalIDList: calIDs}
	var resp GetCalendarResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/calendar/get", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCalendar 删除日历。
func (cc *CorpClient) DeleteCalendar(ctx context.Context, calID string) error {
	body := &DeleteCalendarReq{CalID: calID}
	return cc.doPost(ctx, "/cgi-bin/oa/calendar/del", body, nil)
}

// CreateSchedule 创建日程。
func (cc *CorpClient) CreateSchedule(ctx context.Context, req *CreateScheduleReq) (*CreateScheduleResp, error) {
	var resp CreateScheduleResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSchedule 更新日程。
func (cc *CorpClient) UpdateSchedule(ctx context.Context, req *UpdateScheduleReq) error {
	return cc.doPost(ctx, "/cgi-bin/oa/schedule/update", req, nil)
}

// GetSchedule 获取日程详情。
func (cc *CorpClient) GetSchedule(ctx context.Context, scheduleIDs []string) (*GetScheduleResp, error) {
	body := &GetScheduleReq{ScheduleIDList: scheduleIDs}
	var resp GetScheduleResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/get", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSchedule 删除日程。
func (cc *CorpClient) DeleteSchedule(ctx context.Context, scheduleID string) error {
	body := &DeleteScheduleReq{ScheduleID: scheduleID}
	return cc.doPost(ctx, "/cgi-bin/oa/schedule/del", body, nil)
}

// GetScheduleByCalendar 获取日历下的日程列表。
func (cc *CorpClient) GetScheduleByCalendar(ctx context.Context, req *GetScheduleByCalendarReq) (*GetScheduleByCalendarResp, error) {
	var resp GetScheduleByCalendarResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/get_by_calendar", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
