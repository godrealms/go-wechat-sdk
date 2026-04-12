package isv

import (
	"net/url"
	"strconv"
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
	p := &oauth2Params{scope: "snsapi_privateinfo"}
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
