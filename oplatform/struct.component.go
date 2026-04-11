package oplatform

// component_access_token 响应
type componentTokenResp struct {
	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int64  `json:"expires_in"`
	ErrCode              int    `json:"errcode,omitempty"`
	ErrMsg               string `json:"errmsg,omitempty"`
}

// pre_auth_code 响应
type preAuthCodeResp struct {
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// ComponentNotify 回调解析结果。
type ComponentNotify struct {
	AppID      string // ComponentAppID
	CreateTime int64
	InfoType   string // component_verify_ticket / authorized / updateauthorized / unauthorized

	// component_verify_ticket 时填
	ComponentVerifyTicket string

	// authorized / updateauthorized / unauthorized 时填
	AuthorizerAppID              string
	AuthorizationCode            string
	AuthorizationCodeExpiredTime int64
	PreAuthCode                  string

	// 原始明文 XML
	Raw []byte
}
