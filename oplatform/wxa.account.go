package oplatform

import "context"

// SetNickname 设置小程序名称。
// /wxa/setnickname
func (w *WxaAdminClient) SetNickname(ctx context.Context, req *WxaSetNicknameReq) (*WxaSetNicknameResp, error) {
	var resp WxaSetNicknameResp
	if err := w.doPost(ctx, "/wxa/setnickname", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryNickname 查询改名审核状态。
// /wxa/api_wxa_querynickname
func (w *WxaAdminClient) QueryNickname(ctx context.Context, auditID string) (*WxaQueryNicknameResp, error) {
	body := map[string]string{"audit_id": auditID}
	var resp WxaQueryNicknameResp
	if err := w.doPost(ctx, "/wxa/api_wxa_querynickname", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CheckNickname 名称合法性预检。
// /cgi-bin/wxverify/checkwxverifynickname
func (w *WxaAdminClient) CheckNickname(ctx context.Context, nickname string) (*WxaCheckNicknameResp, error) {
	body := map[string]string{"nick_name": nickname}
	var resp WxaCheckNicknameResp
	if err := w.doPost(ctx, "/cgi-bin/wxverify/checkwxverifynickname", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ModifyHeadImage 修改头像。头像区域固定为整张图 (0,0)-(1,1)。
// /cgi-bin/account/modifyheadimage
func (w *WxaAdminClient) ModifyHeadImage(ctx context.Context, mediaID string) error {
	body := map[string]any{
		"head_img_media_id": mediaID,
		"x1":                0.0,
		"y1":                0.0,
		"x2":                1.0,
		"y2":                1.0,
	}
	return w.doPost(ctx, "/cgi-bin/account/modifyheadimage", body, nil)
}

// ModifySignature 修改功能介绍。
// /cgi-bin/account/modifysignature
func (w *WxaAdminClient) ModifySignature(ctx context.Context, signature string) error {
	body := map[string]string{"signature": signature}
	return w.doPost(ctx, "/cgi-bin/account/modifysignature", body, nil)
}
