package isv

// CorpTokenResp 是 service/get_corp_token 的响应。
type CorpTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
