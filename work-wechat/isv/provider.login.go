package isv

import "context"

// GetLoginInfo 用服务商管理端 OAuth 回跳返回的 auth_code 换取登录身份。
// 使用 provider_access_token,不使用 suite_access_token。
func (c *Client) GetLoginInfo(ctx context.Context, authCode string) (*LoginInfoResp, error) {
	body := map[string]string{"auth_code": authCode}
	var resp LoginInfoResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_login_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
