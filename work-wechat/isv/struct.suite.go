package isv

// SuiteAccessTokenResp 是 service/get_suite_token 的响应体。
type SuiteAccessTokenResp struct {
	SuiteAccessToken string `json:"suite_access_token"`
	ExpiresIn        int    `json:"expires_in"`
}
