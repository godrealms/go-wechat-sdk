package isv

import (
	"context"
	"net/http"
	"net/url"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// doPost 发送 JSON POST 到 parent.baseURL + path,query 自动注入 corp_access_token。
// 企业微信 corp 接口的 query key 是 access_token（不是 corp_access_token）。
func (cc *CorpClient) doPost(ctx context.Context, path string, body, out any) error {
	return utils.DoWithTokenRetry(cc, func() error {
		tok, err := cc.AccessToken(ctx)
		if err != nil {
			return err
		}
		q := url.Values{"access_token": {tok}}
		return cc.parent.doPostRaw(ctx, path, q, body, out)
	})
}

// doGet 发送 GET 到 parent.baseURL + path，query 自动注入 corp_access_token。
func (cc *CorpClient) doGet(ctx context.Context, path string, extra url.Values, out any) error {
	return utils.DoWithTokenRetry(cc, func() error {
		tok, err := cc.AccessToken(ctx)
		if err != nil {
			return err
		}
		q := url.Values{"access_token": {tok}}
		for k, vs := range extra {
			q[k] = vs
		}
		return cc.parent.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
	})
}

// doPostExtra is like doPost but merges extra query params alongside access_token.
func (cc *CorpClient) doPostExtra(ctx context.Context, path string, extra url.Values, body, out any) error {
	return utils.DoWithTokenRetry(cc, func() error {
		tok, err := cc.AccessToken(ctx)
		if err != nil {
			return err
		}
		q := url.Values{"access_token": {tok}}
		for k, vs := range extra {
			q[k] = vs
		}
		return cc.parent.doPostRaw(ctx, path, q, body, out)
	})
}

// Invalidate marks the cached corp_access_token stale so the next AccessToken
// call re-fetches it via permanent_code (writing the fresh token back to the
// shared Store, which benefits every instance reading that Store). doPost,
// doGet, and doPostExtra call it automatically on a token-expired response
// (see utils.IsTokenExpired). Implements utils.Invalidator.
//
// Unlike the flat product packages — whose token cache is process-local — the
// corp token is Store-backed, so invalidation is a per-handle flag rather than
// an eviction: the refresh itself repopulates the Store for everyone.
func (cc *CorpClient) Invalidate() {
	cc.forceRefresh.Store(true)
}
