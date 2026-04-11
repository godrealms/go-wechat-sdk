package oplatform

import (
	"context"
	"fmt"
	"net/url"
)

// GetSubscribeCategory 获取小程序账号所属类目。
// GET /wxaapi/newtmpl/getcategory
func (w *WxaAdminClient) GetSubscribeCategory(ctx context.Context) (*WxaSubscribeCategoryResp, error) {
	var resp WxaSubscribeCategoryResp
	if err := w.doGet(ctx, "/wxaapi/newtmpl/getcategory", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPubTemplateTitles 获取模板库标题列表。
// ids 是用 "-" 连接的类目 ID 列表，例如 "2-3-5"。
// GET /wxaapi/newtmpl/getpubtemplatetitles
func (w *WxaAdminClient) GetPubTemplateTitles(ctx context.Context, ids string, start, limit int) (*WxaPubTemplateTitles, error) {
	q := url.Values{
		"ids":   {ids},
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	}
	var resp WxaPubTemplateTitles
	if err := w.doGet(ctx, "/wxaapi/newtmpl/getpubtemplatetitles", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPubTemplateKeywords 获取模板库标题下的关键词列表。
// GET /wxaapi/newtmpl/getpubtemplatekeywords
func (w *WxaAdminClient) GetPubTemplateKeywords(ctx context.Context, tid string) (*WxaPubTemplateKeywords, error) {
	q := url.Values{"tid": {tid}}
	var resp WxaPubTemplateKeywords
	if err := w.doGet(ctx, "/wxaapi/newtmpl/getpubtemplatekeywords", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddSubscribeTemplate 组合关键词添加私有模板。
// POST /wxaapi/newtmpl/addtemplate
func (w *WxaAdminClient) AddSubscribeTemplate(ctx context.Context, req *WxaAddSubscribeTemplateReq) (*WxaAddSubscribeTemplateResp, error) {
	var resp WxaAddSubscribeTemplateResp
	if err := w.doPost(ctx, "/wxaapi/newtmpl/addtemplate", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSubscribeTemplate 删除私有模板。
// POST /wxaapi/newtmpl/deltemplate
func (w *WxaAdminClient) DeleteSubscribeTemplate(ctx context.Context, priTmplID string) error {
	body := map[string]string{"priTmplId": priTmplID}
	return w.doPost(ctx, "/wxaapi/newtmpl/deltemplate", body, nil)
}

// ListSubscribeTemplates 获取账号下已添加的私有模板列表。
// GET /wxaapi/newtmpl/gettemplate
func (w *WxaAdminClient) ListSubscribeTemplates(ctx context.Context) (*WxaSubscribeTemplateList, error) {
	var resp WxaSubscribeTemplateList
	if err := w.doGet(ctx, "/wxaapi/newtmpl/gettemplate", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendSubscribeMessage 发送订阅消息。
// POST /cgi-bin/message/subscribe/send
func (w *WxaAdminClient) SendSubscribeMessage(ctx context.Context, req *WxaSendSubscribeReq) error {
	return w.doPost(ctx, "/cgi-bin/message/subscribe/send", req, nil)
}
