package isv

import (
	"context"
	"fmt"
)

// GetCheckinData retrieves check-in record data for the specified users and time range.
func (cc *CorpClient) GetCheckinData(ctx context.Context, req *GetCheckinDataReq) (*GetCheckinDataResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: GetCheckinData: req is required")
	}
	var resp GetCheckinDataResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckindata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCheckinOption retrieves the check-in rules for the specified users.
func (cc *CorpClient) GetCheckinOption(ctx context.Context, req *GetCheckinOptionReq) (*GetCheckinOptionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: GetCheckinOption: req is required")
	}
	var resp GetCheckinOptionResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckinoption", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCheckinDayData retrieves daily check-in report data for the specified users and time range.
func (cc *CorpClient) GetCheckinDayData(ctx context.Context, req *GetCheckinDayDataReq) (*GetCheckinDayDataResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: GetCheckinDayData: req is required")
	}
	var resp GetCheckinDayDataResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckin_daydata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCheckinMonthData retrieves monthly check-in report data for the specified users and time range.
func (cc *CorpClient) GetCheckinMonthData(ctx context.Context, req *GetCheckinMonthDataReq) (*GetCheckinMonthDataResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: GetCheckinMonthData: req is required")
	}
	var resp GetCheckinMonthDataResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckin_monthdata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
