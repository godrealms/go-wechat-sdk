package isv

import (
	"context"
	"fmt"
)

// GetApprovalTemplate retrieves the details of an approval template by its template ID.
func (cc *CorpClient) GetApprovalTemplate(ctx context.Context, templateID string) (*ApprovalTemplateResp, error) {
	if err := requireNonEmpty("GetApprovalTemplate", "templateID", templateID); err != nil {
		return nil, err
	}
	body := &GetApprovalTemplateReq{TemplateID: templateID}
	var resp ApprovalTemplateResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/gettemplatedetail", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplyEvent submits an approval application.
func (cc *CorpClient) ApplyEvent(ctx context.Context, req *ApplyEventReq) (*ApplyEventResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: ApplyEvent: req is required")
	}
	var resp ApplyEventResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/applyevent", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalDetail retrieves the details of an approval application by its sp_no.
func (cc *CorpClient) GetApprovalDetail(ctx context.Context, spNo string) (*ApprovalDetailResp, error) {
	if err := requireNonEmpty("GetApprovalDetail", "spNo", spNo); err != nil {
		return nil, err
	}
	body := &GetApprovalDetailReq{SpNo: spNo}
	var resp ApprovalDetailResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovaldetail", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetApprovalData retrieves a batch of approval record numbers matching the given filters.
func (cc *CorpClient) GetApprovalData(ctx context.Context, req *GetApprovalDataReq) (*GetApprovalDataResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: GetApprovalData: req is required")
	}
	var resp GetApprovalDataResp
	if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovalinfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
