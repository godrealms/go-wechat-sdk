package isv

// SuiteAccessTokenResp 是 service/get_suite_token 的响应体。
type SuiteAccessTokenResp struct {
	SuiteAccessToken string `json:"suite_access_token"`
	ExpiresIn        int    `json:"expires_in"`
}

// PreAuthCodeResp 是 service/get_pre_auth_code 的响应。
type PreAuthCodeResp struct {
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

// SessionInfo 是 service/set_session_info 的 session_info 字段。
type SessionInfo struct {
	AppID    []int `json:"appid,omitempty"`     // 限定授权的应用 ID 列表
	AuthType int   `json:"auth_type,omitempty"` // 0=管理员授权,1=成员授权
}
