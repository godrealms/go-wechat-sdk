package oplatform

import (
	"context"
	"fmt"
	"net/url"
)

// PreAuthCode 调用 /cgi-bin/component/api_create_preauthcode 换预授权码。
// 预授权码 TTL ~10 分钟，调用方应每次使用前重新调用本方法。
func (c *Client) PreAuthCode(ctx context.Context) (string, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{"component_appid": c.cfg.ComponentAppID}

	var resp preAuthCodeResp
	path := "/cgi-bin/component/api_create_preauthcode?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return "", fmt.Errorf("oplatform: api_create_preauthcode: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return "", err
	}
	if resp.PreAuthCode == "" {
		return "", fmt.Errorf("oplatform: empty pre_auth_code")
	}
	return resp.PreAuthCode, nil
}

// AuthorizeURL 构造 PC 版本的引导授权跳转 URL。
//
//	preAuthCode - 来自 PreAuthCode()
//	redirectURI - 授权完成后回调地址
//	authType    - 1=公众号 2=小程序 3=全部；当 bizAppid 非空时被忽略
//	bizAppid    - 指定授权方 appid；不指定传 ""
func (c *Client) AuthorizeURL(preAuthCode, redirectURI string, authType int, bizAppid string) string {
	q := url.Values{
		"component_appid": {c.cfg.ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {redirectURI},
	}
	if bizAppid != "" {
		q.Set("biz_appid", bizAppid)
	} else {
		q.Set("auth_type", fmt.Sprintf("%d", authType))
	}
	return "https://mp.weixin.qq.com/cgi-bin/componentloginpage?" + q.Encode()
}

// MobileAuthorizeURL 构造移动端（扫码）授权跳转 URL。
func (c *Client) MobileAuthorizeURL(preAuthCode, redirectURI string, authType int, bizAppid string) string {
	q := url.Values{
		"action":          {"bindcomponent"},
		"no_scan":         {"1"},
		"component_appid": {c.cfg.ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {redirectURI},
	}
	if bizAppid != "" {
		q.Set("biz_appid", bizAppid)
	} else {
		q.Set("auth_type", fmt.Sprintf("%d", authType))
	}
	return "https://mp.weixin.qq.com/safe/bindcomponent?" + q.Encode() + "#wechat_redirect"
}
