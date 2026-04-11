package oplatform

// AuthorizationInfo 对应 /cgi-bin/component/api_query_auth 的 authorization_info。
type AuthorizationInfo struct {
	AuthorizerAppID        string     `json:"authorizer_appid"`
	AuthorizerAccessToken  string     `json:"authorizer_access_token"`
	ExpiresIn              int64      `json:"expires_in"`
	AuthorizerRefreshToken string     `json:"authorizer_refresh_token"`
	FuncInfo               []FuncInfo `json:"func_info"`
}

type FuncInfo struct {
	FuncscopeCategory FuncscopeCategory `json:"funcscope_category"`
}

type FuncscopeCategory struct {
	ID int `json:"id"`
}

type queryAuthResp struct {
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
	ErrCode           int               `json:"errcode,omitempty"`
	ErrMsg            string            `json:"errmsg,omitempty"`
}

// authorizer_access_token 刷新响应
type authorizerTokenResp struct {
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
	ExpiresIn              int64  `json:"expires_in"`
	ErrCode                int    `json:"errcode,omitempty"`
	ErrMsg                 string `json:"errmsg,omitempty"`
}

// AuthorizerInfo / Option 查询结构体
type AuthorizerInfo struct {
	NickName        string `json:"nick_name"`
	HeadImg         string `json:"head_img"`
	ServiceTypeInfo struct {
		ID int `json:"id"`
	} `json:"service_type_info"`
	VerifyTypeInfo struct {
		ID int `json:"id"`
	} `json:"verify_type_info"`
	UserName        string `json:"user_name"`
	PrincipalName   string `json:"principal_name"`
	Alias           string `json:"alias"`
	BusinessInfo    any    `json:"business_info,omitempty"`
	QrcodeURL       string `json:"qrcode_url"`
	Signature       string `json:"signature"`
	MiniProgramInfo any    `json:"MiniProgramInfo,omitempty"`
}

type AuthorizerInfoResp struct {
	AuthorizerInfo    AuthorizerInfo    `json:"authorizer_info"`
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
	ErrCode           int               `json:"errcode,omitempty"`
	ErrMsg            string            `json:"errmsg,omitempty"`
}

type AuthorizerOption struct {
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
	ErrCode         int    `json:"errcode,omitempty"`
	ErrMsg          string `json:"errmsg,omitempty"`
}

type AuthorizerList struct {
	TotalCount int `json:"total_count"`
	List       []struct {
		AuthorizerAppID string `json:"authorizer_appid"`
		RefreshToken    string `json:"refresh_token"`
		AuthTime        int64  `json:"auth_time"`
	} `json:"list"`
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
}
