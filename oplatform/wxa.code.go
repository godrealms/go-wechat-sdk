package oplatform

import (
	"context"
	"net/url"
)

// Commit 上传代码。
// /wxa/commit
func (w *WxaAdminClient) Commit(ctx context.Context, req *WxaCommitReq) error {
	return w.doPost(ctx, "/wxa/commit", req, nil)
}

// GetPage 获取已上传代码的页面列表。
// /wxa/get_page
func (w *WxaAdminClient) GetPage(ctx context.Context) (*WxaGetPageResp, error) {
	var resp WxaGetPageResp
	if err := w.doGet(ctx, "/wxa/get_page", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetQrcode 获取体验版二维码（返回二进制图片）。
// /wxa/get_qrcode
func (w *WxaAdminClient) GetQrcode(ctx context.Context, path string) ([]byte, string, error) {
	q := url.Values{}
	if path != "" {
		q.Set("path", path)
	}
	return w.doGetRaw(ctx, "/wxa/get_qrcode", q)
}

// GetCodeCategory 获取代码草稿可选类目。
// /wxa/get_category
func (w *WxaAdminClient) GetCodeCategory(ctx context.Context) (*WxaGetCodeCategoryResp, error) {
	var resp WxaGetCodeCategoryResp
	if err := w.doGet(ctx, "/wxa/get_category", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
