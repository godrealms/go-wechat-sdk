package isv

import "context"

// GetApprovalTemplate retrieves the details of an approval template by its template ID.
func (cc *CorpClient) GetApprovalTemplate(ctx context.Context, templateID string) (*ApprovalTemplateResp, error) {
	body := &GetApprovalTemplateReq{TemplateID: templateID}
	var resp ApprovalTemplateResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/gettemplatedetail", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplyEvent submits an approval application.
func (cc *CorpClient) ApplyEvent(ctx context.Context, req *ApplyEventReq) (*ApplyEventResp, error) {
	var resp ApplyEventResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/applyevent", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalDetail retrieves the details of an approval application by its sp_no.
func (cc *CorpClient) GetApprovalDetail(ctx context.Context, spNo string) (*ApprovalDetailResp, error) {
	body := &GetApprovalDetailReq{SpNo: spNo}
	var resp ApprovalDetailResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovaldetail", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalData retrieves a batch of approval record numbers matching the given filters.
func (cc *CorpClient) GetApprovalData(ctx context.Context, req *GetApprovalDataReq) (*GetApprovalDataResp, error) {
	var resp GetApprovalDataResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovalinfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
