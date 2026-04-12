package isv

import "context"

// GetCheckinData 获取打卡记录数据。
func (cc *CorpClient) GetCheckinData(ctx context.Context, req *GetCheckinDataReq) (*GetCheckinDataResp, error) {
	var resp GetCheckinDataResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckindata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCheckinOption 获取打卡规则。
func (cc *CorpClient) GetCheckinOption(ctx context.Context, req *GetCheckinOptionReq) (*GetCheckinOptionResp, error) {
	var resp GetCheckinOptionResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckinoption", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCheckinDayData 获取打卡日报数据。
func (cc *CorpClient) GetCheckinDayData(ctx context.Context, req *GetCheckinDayDataReq) (*GetCheckinDayDataResp, error) {
	var resp GetCheckinDayDataResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckin_daydata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCheckinMonthData 获取打卡月报数据。
func (cc *CorpClient) GetCheckinMonthData(ctx context.Context, req *GetCheckinMonthDataReq) (*GetCheckinMonthDataResp, error) {
	var resp GetCheckinMonthDataResp
	if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckin_monthdata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
