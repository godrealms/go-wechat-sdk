package oplatform

import "context"

// ModifyServerDomain 设置/增加/删除服务器域名。
// /wxa/modify_domain
func (w *WxaAdminClient) ModifyServerDomain(ctx context.Context, req *WxaModifyServerDomainReq) (*WxaServerDomainResp, error) {
	var resp WxaServerDomainResp
	if err := w.doPost(ctx, "/wxa/modify_domain", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetWebviewDomain 设置业务域名。
// /wxa/setwebviewdomain
func (w *WxaAdminClient) SetWebviewDomain(ctx context.Context, req *WxaSetWebviewDomainReq) error {
	return w.doPost(ctx, "/wxa/setwebviewdomain", req, nil)
}

// GetDomainConfirmFile 获取业务域名校验文件。
// /wxa/get_webview_confirmfile
func (w *WxaAdminClient) GetDomainConfirmFile(ctx context.Context) (*WxaDomainConfirmFile, error) {
	var resp WxaDomainConfirmFile
	if err := w.doPost(ctx, "/wxa/get_webview_confirmfile", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ModifyDomainDirectly 快速配置小程序服务器域名。
// /wxa/modify_domain_directly
func (w *WxaAdminClient) ModifyDomainDirectly(ctx context.Context, req *WxaModifyDomainDirectlyReq) (*WxaServerDomainResp, error) {
	var resp WxaServerDomainResp
	if err := w.doPost(ctx, "/wxa/modify_domain_directly", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
