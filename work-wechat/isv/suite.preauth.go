package isv

import (
	"context"
	"net/url"
)

// GetPreAuthCode 拉取预授权码。
func (c *Client) GetPreAuthCode(ctx context.Context) (*PreAuthCodeResp, error) {
	var resp PreAuthCodeResp
	if err := c.doGet(ctx, "/cgi-bin/service/get_pre_auth_code", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSessionInfo 为指定 pre_auth_code 绑定授权会话配置。
func (c *Client) SetSessionInfo(ctx context.Context, preAuthCode string, info *SessionInfo) error {
	body := map[string]interface{}{
		"pre_auth_code": preAuthCode,
		"session_info":  info,
	}
	return c.doPost(ctx, "/cgi-bin/service/set_session_info", body, nil)
}

// AuthorizeURL 拼接企业管理员扫码授权的 PC 跳转 URL(不发起 HTTP)。
func (c *Client) AuthorizeURL(preAuthCode, redirectURI, state string) string {
	q := url.Values{
		"suite_id":      {c.cfg.SuiteID},
		"pre_auth_code": {preAuthCode},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	return "https://open.work.weixin.qq.com/3rdapp/install?" + q.Encode()
}
