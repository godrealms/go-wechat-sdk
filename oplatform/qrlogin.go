package oplatform

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// QRLoginClient 提供微信开放平台"网站应用微信登录"能力 (snsapi_login)。
// 与 component Client 完全独立：无 Store 依赖、无 token 缓存，每次换 token
// 都直接调微信接口。
type QRLoginClient struct {
	appID     string
	appSecret string
	http      *utils.HTTP
}

type QRLoginOption func(*QRLoginClient)

// WithQRLoginHTTP 注入自定义 HTTP（测试常用）。
func WithQRLoginHTTP(h *utils.HTTP) QRLoginOption {
	return func(q *QRLoginClient) {
		if h != nil {
			q.http = h
		}
	}
}

// NewQRLoginClient 构造一个 QR Login 客户端。
func NewQRLoginClient(appID, appSecret string, opts ...QRLoginOption) *QRLoginClient {
	q := &QRLoginClient{
		appID:     appID,
		appSecret: appSecret,
		http:      utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(q)
	}
	return q
}

// AuthorizeURL 构造开放平台扫码登录跳转 URL。
//
//	scope - snsapi_login / snsapi_base / snsapi_userinfo
//	state - CSRF 防护 token
func (q *QRLoginClient) AuthorizeURL(redirectURI, scope, state string) string {
	v := url.Values{
		"appid":         {q.appID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {scope},
		"state":         {state},
	}
	return "https://open.weixin.qq.com/connect/qrconnect?" + v.Encode() + "#wechat_redirect"
}

// Code2Token 用扫码登录 code 换取 access_token。
func (q *QRLoginClient) Code2Token(ctx context.Context, code string) (*QRLoginToken, error) {
	ctx = touchContext(ctx)
	if code == "" {
		return nil, fmt.Errorf("oplatform: code is required")
	}
	v := url.Values{
		"appid":      {q.appID},
		"secret":     {q.appSecret},
		"code":       {code},
		"grant_type": {"authorization_code"},
	}
	out := &QRLoginToken{}
	if err := q.http.Get(ctx, "/sns/oauth2/access_token", v, out); err != nil {
		return nil, fmt.Errorf("oplatform: qrlogin code2token: %w", err)
	}
	if err := checkWeixinErr(out.ErrCode, out.ErrMsg); err != nil {
		return nil, err
	}
	return out, nil
}

// RefreshToken 用 refresh_token 换取新的 access_token。
func (q *QRLoginClient) RefreshToken(ctx context.Context, refreshToken string) (*QRLoginToken, error) {
	ctx = touchContext(ctx)
	if refreshToken == "" {
		return nil, fmt.Errorf("oplatform: refresh_token is required")
	}
	v := url.Values{
		"appid":         {q.appID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	out := &QRLoginToken{}
	if err := q.http.Get(ctx, "/sns/oauth2/refresh_token", v, out); err != nil {
		return nil, fmt.Errorf("oplatform: qrlogin refresh: %w", err)
	}
	if err := checkWeixinErr(out.ErrCode, out.ErrMsg); err != nil {
		return nil, err
	}
	return out, nil
}

// UserInfo 拉取用户信息。仅 snsapi_userinfo scope 可用。
func (q *QRLoginClient) UserInfo(ctx context.Context, accessToken, openID string) (*QRLoginUserInfo, error) {
	ctx = touchContext(ctx)
	v := url.Values{
		"access_token": {accessToken},
		"openid":       {openID},
		"lang":         {"zh_CN"},
	}
	out := &QRLoginUserInfo{}
	if err := q.http.Get(ctx, "/sns/userinfo", v, out); err != nil {
		return nil, fmt.Errorf("oplatform: qrlogin userinfo: %w", err)
	}
	if err := checkWeixinErr(out.ErrCode, out.ErrMsg); err != nil {
		return nil, err
	}
	return out, nil
}

// Auth 检查 access_token 是否有效。
func (q *QRLoginClient) Auth(ctx context.Context, accessToken, openID string) error {
	ctx = touchContext(ctx)
	v := url.Values{"access_token": {accessToken}, "openid": {openID}}
	out := &qrloginAuthResp{}
	if err := q.http.Get(ctx, "/sns/auth", v, out); err != nil {
		return fmt.Errorf("oplatform: qrlogin auth: %w", err)
	}
	return checkWeixinErr(out.ErrCode, out.ErrMsg)
}
