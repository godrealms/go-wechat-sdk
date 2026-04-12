package isv

import (
	"context"
	"net/http"
	"net/url"
)

// doPost 发送 JSON POST 到 parent.baseURL + path,query 自动注入 corp_access_token。
// 企业微信 corp 接口的 query key 是 access_token（不是 corp_access_token）。
func (cc *CorpClient) doPost(ctx context.Context, path string, body, out interface{}) error {
	tok, err := cc.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	return cc.parent.doPostRaw(ctx, path, q, body, out)
}

// doGet 发送 GET 到 parent.baseURL + path，query 自动注入 corp_access_token。
func (cc *CorpClient) doGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
	tok, err := cc.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}
	return cc.parent.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}
