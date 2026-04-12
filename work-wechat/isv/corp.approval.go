package isv

import "context"

// GetApprovalTemplate 获取审批模板详情。
func (cc *CorpClient) GetApprovalTemplate(ctx context.Context, templateID string) (*ApprovalTemplateResp, error) {
	body := &GetApprovalTemplateReq{TemplateID: templateID}
	var resp ApprovalTemplateResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/gettemplatedetail", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplyEvent 提交审批申请。
func (cc *CorpClient) ApplyEvent(ctx context.Context, req *ApplyEventReq) (*ApplyEventResp, error) {
	var resp ApplyEventResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/applyevent", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalDetail 获取审批申请详情。
func (cc *CorpClient) GetApprovalDetail(ctx context.Context, spNo string) (*ApprovalDetailResp, error) {
	body := &GetApprovalDetailReq{SpNo: spNo}
	var resp ApprovalDetailResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovaldetail", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalData 批量获取审批单号。
func (cc *CorpClient) GetApprovalData(ctx context.Context, req *GetApprovalDataReq) (*GetApprovalDataResp, error) {
	var resp GetApprovalDataResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovalinfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
