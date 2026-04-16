package isv

import "context"

// GetLoginInfo 用服务商管理端 OAuth 回跳返回的 auth_code 换取登录身份。
// 使用 provider_access_token,不使用 suite_access_token。
func (c *Client) GetLoginInfo(ctx context.Context, authCode string) (*LoginInfoResp, error) {
	if err := requireNonEmpty("GetLoginInfo", "authCode", authCode); err != nil {
		return nil, err
	}
	body := map[string]string{"auth_code": authCode}
	var resp LoginInfoResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_login_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRegisterCode 生成注册企业微信的 register_code(邀请链接的核心参数)。
// 所有字段都是可选 —— 服务端按缺省值处理。
func (c *Client) GetRegisterCode(ctx context.Context, req *GetRegisterCodeReq) (*RegisterCodeResp, error) {
	if req == nil {
		req = &GetRegisterCodeReq{}
	}
	var resp RegisterCodeResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_register_code", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRegistrationInfo 查询 register_code 对应的注册进度,
// 成功后返回已注册企业的 corpid / 管理员 / 永久授权码 / 通讯录同步 token。
func (c *Client) GetRegistrationInfo(ctx context.Context, registerCode string) (*RegistrationInfoResp, error) {
	if err := requireNonEmpty("GetRegistrationInfo", "registerCode", registerCode); err != nil {
		return nil, err
	}
	body := map[string]string{"register_code": registerCode}
	var resp RegistrationInfoResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_registration_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
