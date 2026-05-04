package isv

import (
	"context"
	"fmt"
)

// CreateCalendar creates a new calendar.
func (cc *CorpClient) CreateCalendar(ctx context.Context, req *CreateCalendarReq) (*CreateCalendarResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: CreateCalendar: req is required")
	}
	var resp CreateCalendarResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/calendar/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCalendar updates an existing calendar.
func (cc *CorpClient) UpdateCalendar(ctx context.Context, req *UpdateCalendarReq) error {
	if req == nil {
		return fmt.Errorf("isv: UpdateCalendar: req is required")
	}
	return cc.doPost(ctx, "/cgi-bin/oa/calendar/update", req, nil)
}

// GetCalendar retrieves the details of one or more calendars by their IDs.
func (cc *CorpClient) GetCalendar(ctx context.Context, calIDs []string) (*GetCalendarResp, error) {
	if len(calIDs) == 0 {
		return nil, fmt.Errorf("isv: GetCalendar: calIDs is required")
	}
	body := &GetCalendarReq{CalIDList: calIDs}
	var resp GetCalendarResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/calendar/get", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCalendar deletes a calendar by its ID.
func (cc *CorpClient) DeleteCalendar(ctx context.Context, calID string) error {
	if err := requireNonEmpty("DeleteCalendar", "calID", calID); err != nil {
		return err
	}
	body := &DeleteCalendarReq{CalID: calID}
	return cc.doPost(ctx, "/cgi-bin/oa/calendar/del", body, nil)
}

// CreateSchedule creates a new schedule event.
func (cc *CorpClient) CreateSchedule(ctx context.Context, req *CreateScheduleReq) (*CreateScheduleResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: CreateSchedule: req is required")
	}
	var resp CreateScheduleResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSchedule updates an existing schedule event.
func (cc *CorpClient) UpdateSchedule(ctx context.Context, req *UpdateScheduleReq) error {
	if req == nil {
		return fmt.Errorf("isv: UpdateSchedule: req is required")
	}
	return cc.doPost(ctx, "/cgi-bin/oa/schedule/update", req, nil)
}

// GetSchedule retrieves the details of one or more schedule events by their IDs.
func (cc *CorpClient) GetSchedule(ctx context.Context, scheduleIDs []string) (*GetScheduleResp, error) {
	if len(scheduleIDs) == 0 {
		return nil, fmt.Errorf("isv: GetSchedule: scheduleIDs is required")
	}
	body := &GetScheduleReq{ScheduleIDList: scheduleIDs}
	var resp GetScheduleResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/get", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSchedule deletes a schedule event by its ID.
func (cc *CorpClient) DeleteSchedule(ctx context.Context, scheduleID string) error {
	if err := requireNonEmpty("DeleteSchedule", "scheduleID", scheduleID); err != nil {
		return err
	}
	body := &DeleteScheduleReq{ScheduleID: scheduleID}
	return cc.doPost(ctx, "/cgi-bin/oa/schedule/del", body, nil)
}

// GetScheduleByCalendar retrieves the list of schedule events under a specific calendar.
func (cc *CorpClient) GetScheduleByCalendar(ctx context.Context, req *GetScheduleByCalendarReq) (*GetScheduleByCalendarResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: GetScheduleByCalendar: req is required")
	}
	var resp GetScheduleByCalendarResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/get_by_calendar", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
