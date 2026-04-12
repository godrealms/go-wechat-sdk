package isv

// ---------- get_permanent_code / get_auth_info 共用 ----------

// AuthCorpInfo 授权企业信息。
type AuthCorpInfo struct {
	CorpID            string `json:"corpid"`
	CorpName          string `json:"corp_name"`
	CorpType          string `json:"corp_type"`
	CorpSquareLogoURL string `json:"corp_square_logo_url"`
	CorpUserMax       int    `json:"corp_user_max"`
	CorpFullName      string `json:"corp_full_name"`
	VerifiedEndTime   int64  `json:"verified_end_time"`
	SubjectType       int    `json:"subject_type"`
	CorpWxqrcode      string `json:"corp_wxqrcode"`
	CorpScale         string `json:"corp_scale"`
	CorpIndustry      string `json:"corp_industry"`
	CorpSubIndustry   string `json:"corp_sub_industry"`
	Location          string `json:"location"`
}

// AgentPrivilege 应用可见范围等权限信息。
type AgentPrivilege struct {
	Level      int      `json:"level"`
	AllowParty []int    `json:"allow_party,omitempty"`
	AllowUser  []string `json:"allow_user,omitempty"`
	AllowTag   []int    `json:"allow_tag,omitempty"`
	ExtraParty []int    `json:"extra_party,omitempty"`
	ExtraUser  []string `json:"extra_user,omitempty"`
	ExtraTag   []int    `json:"extra_tag,omitempty"`
}

// SharedFromInfo 共享应用来源信息。
type SharedFromInfo struct {
	CorpID    string `json:"corpid"`
	ShareType int    `json:"share_type"`
}

// AuthAgent 授权应用信息。
type AuthAgent struct {
	AgentID       int             `json:"agentid"`
	Name          string          `json:"name"`
	RoundLogoURL  string          `json:"round_logo_url"`
	SquareLogoURL string          `json:"square_logo_url"`
	AppID         int             `json:"appid"`
	Privilege     AgentPrivilege  `json:"privilege,omitempty"`
	SharedFrom    *SharedFromInfo `json:"shared_from,omitempty"`
}

// AuthInfoAgent 授权应用列表。
type AuthInfoAgent struct {
	Agent []AuthAgent `json:"agent"`
}

// AuthUserInfo 授权管理员信息。
type AuthUserInfo struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
}

// PermanentCodeResp 是 service/get_permanent_code 的响应。
type PermanentCodeResp struct {
	AccessToken   string        `json:"access_token"`
	ExpiresIn     int           `json:"expires_in"`
	PermanentCode string        `json:"permanent_code"`
	AuthCorpInfo  AuthCorpInfo  `json:"auth_corp_info"`
	AuthInfo      AuthInfoAgent `json:"auth_info"`
	AuthUserInfo  AuthUserInfo  `json:"auth_user_info"`
}

// AuthInfoResp 是 service/get_auth_info 的响应。
type AuthInfoResp struct {
	AuthCorpInfo AuthCorpInfo  `json:"auth_corp_info"`
	AuthInfo     AuthInfoAgent `json:"auth_info"`
}

// ---------- get_admin_list ----------

type AdminInfo struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
	AuthType   int    `json:"auth_type"` // 0=普通管理员 1=超级管理员
}

type AdminListResp struct {
	Admin []AdminInfo `json:"admin"`
}
