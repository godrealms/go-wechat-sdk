package oplatform

import "context"

// BindTester 绑定体验者。
// /wxa/bind_tester
func (w *WxaAdminClient) BindTester(ctx context.Context, wechatID string) (*WxaBindTesterResp, error) {
	body := map[string]string{"wechatid": wechatID}
	var resp WxaBindTesterResp
	if err := w.doPost(ctx, "/wxa/bind_tester", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnbindTester 解绑体验者。
// 提供 wechatID 或 userStr 任意一个（二选一）；两者都空时微信会报错。
// /wxa/unbind_tester
func (w *WxaAdminClient) UnbindTester(ctx context.Context, wechatID, userStr string) error {
	body := map[string]string{}
	if wechatID != "" {
		body["wechatid"] = wechatID
	}
	if userStr != "" {
		body["userstr"] = userStr
	}
	return w.doPost(ctx, "/wxa/unbind_tester", body, nil)
}

// ListTesters 获取体验者列表。
// /wxa/memberauth
func (w *WxaAdminClient) ListTesters(ctx context.Context) (*WxaListTestersResp, error) {
	body := map[string]string{"action": "get_experiencer"}
	var resp WxaListTestersResp
	if err := w.doPost(ctx, "/wxa/memberauth", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
