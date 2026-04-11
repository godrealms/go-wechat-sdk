package oplatform

import "context"

// SubmitAudit 提交审核。
// /wxa/submit_audit
func (w *WxaAdminClient) SubmitAudit(ctx context.Context, req *WxaSubmitAuditReq) (*WxaSubmitAuditResp, error) {
	var resp WxaSubmitAuditResp
	if err := w.doPost(ctx, "/wxa/submit_audit", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuditStatus 查询指定版本审核状态。
// /wxa/get_auditstatus
func (w *WxaAdminClient) GetAuditStatus(ctx context.Context, auditID int64) (*WxaAuditStatus, error) {
	body := map[string]int64{"auditid": auditID}
	var resp WxaAuditStatus
	if err := w.doPost(ctx, "/wxa/get_auditstatus", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLatestAuditStatus 查询最新一次审核状态。
// /wxa/get_latest_auditstatus
func (w *WxaAdminClient) GetLatestAuditStatus(ctx context.Context) (*WxaAuditStatus, error) {
	var resp WxaAuditStatus
	if err := w.doGet(ctx, "/wxa/get_latest_auditstatus", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UndoCodeAudit 撤回代码审核。
// /wxa/undocodeaudit
func (w *WxaAdminClient) UndoCodeAudit(ctx context.Context) error {
	return w.doGet(ctx, "/wxa/undocodeaudit", nil, nil)
}

// SpeedupAudit 加急审核。
// /wxa/speedupaudit
func (w *WxaAdminClient) SpeedupAudit(ctx context.Context, auditID int64) error {
	body := map[string]int64{"auditid": auditID}
	return w.doPost(ctx, "/wxa/speedupaudit", body, nil)
}

// Release 发布已通过审核的版本。
// /wxa/release
func (w *WxaAdminClient) Release(ctx context.Context) error {
	return w.doPost(ctx, "/wxa/release", struct{}{}, nil)
}

// RevertCodeRelease 版本回退。
// /wxa/revertcoderelease
func (w *WxaAdminClient) RevertCodeRelease(ctx context.Context) error {
	return w.doGet(ctx, "/wxa/revertcoderelease", nil, nil)
}

// ChangeVisitStatus 修改可见状态。action = "open" | "close"
// /wxa/change_visitstatus
func (w *WxaAdminClient) ChangeVisitStatus(ctx context.Context, action string) error {
	body := map[string]string{"action": action}
	return w.doPost(ctx, "/wxa/change_visitstatus", body, nil)
}

// GetSupportVersion 查询小程序支持版本信息。
// /cgi-bin/wxopen/getweappsupportversion
func (w *WxaAdminClient) GetSupportVersion(ctx context.Context) (*WxaSupportVersionResp, error) {
	var resp WxaSupportVersionResp
	if err := w.doPost(ctx, "/cgi-bin/wxopen/getweappsupportversion", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSupportVersion 设置小程序最低支持版本。
// /cgi-bin/wxopen/setweappsupportversion
func (w *WxaAdminClient) SetSupportVersion(ctx context.Context, version string) error {
	body := map[string]string{"version": version}
	return w.doPost(ctx, "/cgi-bin/wxopen/setweappsupportversion", body, nil)
}
