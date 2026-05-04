package utils

import (
	"context"
	"fmt"
	"net/url"
)

// TokenSource is an injectable access_token provider. Any package's Client
// that exposes a TokenSource type aliases this so a single implementation
// (typically oplatform.AuthorizerClient or a custom Redis-backed source)
// can serve every WeChat product line.
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

// Invalidator is an optional capability for TokenSource implementations
// that support explicit cache eviction. When a Client receives a
// 40001/40014/42001/42007 from WeChat it calls Invalidate so the next
// AccessToken call re-fetches a fresh token.
//
// utils.TokenCache implements this interface; the open-platform
// AuthorizerClient and any third-party TokenSource that wants 40001
// self-heal support should also implement it.
type Invalidator interface {
	Invalidate()
}

// AccessTokenResp is the shape of WeChat's /cgi-bin/token response
// (success + error variants). Exported because it is part of the
// FetchAccessToken contract; callers normally do not construct it.
type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// FetchAccessToken issues the standard /cgi-bin/token HTTP call to WeChat's
// global access_token endpoint and returns (token, expires_in_seconds, err).
//
// newAPIError converts a non-zero ErrCode into a package-typed error so
// callers can inspect it with errors.As against their own *APIError type:
//
//	return utils.FetchAccessToken(ctx, c.http, c.cfg.AppId, c.cfg.AppSecret,
//	    func(code int, msg, path string) error {
//	        return &APIError{ErrCode: code, ErrMsg: msg, Path: path}
//	    })
//
// The function returns an empty-token error if WeChat replies with HTTP 200
// but no access_token field — this is rare but observed during platform
// incidents and would otherwise poison the TokenCache.
func FetchAccessToken(
	ctx context.Context,
	http *HTTP,
	appID, appSecret string,
	newAPIError func(code int, msg, path string) error,
) (string, int64, error) {
	q := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {appID},
		"secret":     {appSecret},
	}
	out := &AccessTokenResp{}
	if err := http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
		return "", 0, fmt.Errorf("fetch access_token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", 0, newAPIError(out.ErrCode, out.ErrMsg, "/cgi-bin/token")
	}
	if out.AccessToken == "" {
		return "", 0, fmt.Errorf("/cgi-bin/token returned empty access_token")
	}
	return out.AccessToken, out.ExpiresIn, nil
}
