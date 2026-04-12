package isv

// ---------- service/get_login_info ----------

// LoginInfoResp 是 service/get_login_info 的响应。
// UserType: 1 = 企业管理员,2 = 企业成员,3 = 服务商(代开发)成员。
type LoginInfoResp struct {
	UserType int                 `json:"usertype"`
	UserInfo LoginInfoUser       `json:"user_info"`
	CorpInfo LoginInfoCorp       `json:"corp_info"`
	Agent    []LoginInfoAgent    `json:"agent"`
	AuthInfo LoginInfoPermission `json:"auth_info"`
}

// LoginInfoUser 登录者自身信息。
type LoginInfoUser struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
}

// LoginInfoCorp 登录者所属企业。
type LoginInfoCorp struct {
	CorpID string `json:"corpid"`
}

// LoginInfoAgent 第三方应用在该企业下的 agent。
type LoginInfoAgent struct {
	AgentID  int `json:"agentid"`
	AuthType int `json:"auth_type"` // 0 = 只使用,1 = 管理
}

// LoginInfoPermission —— 仅当 UserType=1(管理员)时非空,列出被管理的部门。
type LoginInfoPermission struct {
	Department []LoginInfoDepartment `json:"department"`
}

// LoginInfoDepartment 被管理的部门。
type LoginInfoDepartment struct {
	ID       int  `json:"id"`
	Writable bool `json:"writable"`
}

// ---------- service/get_register_code ----------

// GetRegisterCodeReq 是 service/get_register_code 的请求体。
// 所有字段均为可选,服务端按缺省值处理。
type GetRegisterCodeReq struct {
	TemplateID  string `json:"template_id,omitempty"`
	CorpName    string `json:"corp_name,omitempty"`
	AdminName   string `json:"admin_name,omitempty"`
	AdminMobile string `json:"admin_mobile,omitempty"`
	State       string `json:"state,omitempty"`
}

// RegisterCodeResp 是 service/get_register_code 的响应。
type RegisterCodeResp struct {
	RegisterCode string `json:"register_code"`
	ExpiresIn    int    `json:"expires_in"`
}

// ---------- service/get_registration_info ----------

// RegistrationInfoResp 是 service/get_registration_info 的响应。
// AuthInfo 复用子项目 1 的 AuthInfoAgent(字段布局一致)。
type RegistrationInfoResp struct {
	CorpInfo      RegistrationCorpInfo  `json:"corp_info"`
	AuthUserInfo  RegistrationAdminInfo `json:"auth_user_info"`
	ContactSync   RegistrationContact   `json:"contact_sync"`
	AuthInfo      AuthInfoAgent         `json:"auth_info"`
	PermanentCode string                `json:"permanent_code"`
}

// RegistrationCorpInfo 已注册企业的信息快照。
type RegistrationCorpInfo struct {
	CorpID            string `json:"corpid"`
	CorpName          string `json:"corp_name"`
	CorpType          string `json:"corp_type"`
	CorpSquareLogoURL string `json:"corp_square_logo_url"`
	CorpUserMax       int    `json:"corp_user_max"`
	SubjectType       int    `json:"subject_type"`
	VerifiedEndTime   int    `json:"verified_end_time"`
	CorpWxqrcode      string `json:"corp_wxqrcode"`
	CorpScale         string `json:"corp_scale"`
	CorpIndustry      string `json:"corp_industry"`
	CorpSubIndustry   string `json:"corp_sub_industry"`
}

// RegistrationAdminInfo 注册企业的初始管理员。
type RegistrationAdminInfo struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
}

// RegistrationContact —— 注册完成后返回的通讯录同步 token(单次有效)。
type RegistrationContact struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
