package isv

import (
	"context"
	"net/url"
	"strconv"
)

// OAuth2 scope 常量。企业微信仅接受这两个值。
const (
	ScopeBase        = "snsapi_base"        // 静默授权,只能拿到 UserId
	ScopePrivateInfo = "snsapi_privateinfo" // 手动授权,可拿到 user_ticket 换敏感详情
)

// OAuth2Option 配置 OAuth2URL 的可选参数。
type OAuth2Option func(*oauth2Params)

type oauth2Params struct {
	scope    string
	agentID  int
	hasAgent bool
}

// WithOAuth2Scope 覆盖默认的 scope。默认 "snsapi_privateinfo"。
// 可选值:snsapi_base / snsapi_privateinfo。
func WithOAuth2Scope(scope string) OAuth2Option {
	return func(p *oauth2Params) { p.scope = scope }
}

// WithOAuth2AgentID 设置 agentid query 参数。
// 仅当 scope=snsapi_privateinfo 时必填,调用方负责正确性。
func WithOAuth2AgentID(agentID int) OAuth2Option {
	return func(p *oauth2Params) {
		p.agentID = agentID
		p.hasAgent = true
	}
}

// OAuth2URL 构造企业微信第三方网页授权 URL。
// 调用方把返回值塞到 302 Location header 即可。
// redirectURI:用户同意授权后企业微信回跳的 URL,必须在服务商后台白名单内。
// state:调用方自定义的防 CSRF 值,回跳时原样带回。
func (c *Client) OAuth2URL(redirectURI, state string, opts ...OAuth2Option) string {
	p := &oauth2Params{scope: ScopePrivateInfo}
	for _, opt := range opts {
		opt(p)
	}
	q := url.Values{}
	q.Set("appid", c.cfg.SuiteID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", p.scope)
	q.Set("state", state)
	if p.hasAgent {
		q.Set("agentid", strconv.Itoa(p.agentID))
	}
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + q.Encode() + "#wechat_redirect"
}

// GetUserInfo3rd 用回调返回的 auth_code 换取成员身份(UserId / user_ticket / open_userid)。
// 接口:GET /cgi-bin/service/auth/getuserinfo3rd?code=<authCode>
// 使用 provider_access_token。
// 返回的 UserTicket 可继续调用 GetUserDetail3rd 换取敏感详情。
func (c *Client) GetUserInfo3rd(ctx context.Context, authCode string) (*UserInfo3rdResp, error) {
	if err := requireNonEmpty("GetUserInfo3rd", "authCode", authCode); err != nil {
		return nil, err
	}
	extra := url.Values{"code": {authCode}}
	var resp UserInfo3rdResp
	if err := c.providerDoGet(ctx, "/cgi-bin/service/auth/getuserinfo3rd", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserDetail3rd 用 user_ticket 换取成员的敏感详情(姓名/邮箱/头像/手机号)。
// 接口:POST /cgi-bin/service/auth/getuserdetail3rd,body 为 {"user_ticket": "..."}。
// 注意:此接口对敏感字段有调用者备案要求,调用前请确认合规。
func (c *Client) GetUserDetail3rd(ctx context.Context, userTicket string) (*UserDetail3rdResp, error) {
	if err := requireNonEmpty("GetUserDetail3rd", "userTicket", userTicket); err != nil {
		return nil, err
	}
	body := map[string]string{"user_ticket": userTicket}
	var resp UserDetail3rdResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/auth/getuserdetail3rd", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
