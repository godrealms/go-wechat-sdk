package oplatform

// 本文件汇总快速注册 (FastRegister) 子项目的请求/响应 DTO。
// 和 WxaAdmin 的 DTO 分开，因为它们不属于 WxaAdmin 家族。

// ----- enterprise -----

type FastRegEnterpriseReq struct {
	Name               string `json:"name"`
	Code               string `json:"code"`
	CodeType           int    `json:"code_type"`
	LegalPersonaWechat string `json:"legal_persona_wechat"`
	LegalPersonaName   string `json:"legal_persona_name"`
	ComponentPhone     string `json:"component_phone"`
}

type FastRegEnterpriseResp struct{}

type FastRegEnterpriseStatus struct {
	Status          int    `json:"status"`
	AuthCode        string `json:"auth_code,omitempty"`
	AuthorizerAppid string `json:"authorizer_appid,omitempty"`
	IsWxVerify      bool   `json:"is_wx_verify,omitempty"`
	IsLinkMp        bool   `json:"is_link_mp,omitempty"`
}

// ----- personal -----

type FastRegPersonalReq struct {
	IDName         string `json:"idname"`
	WxUser         string `json:"wxuser"`
	ComponentPhone string `json:"component_phone,omitempty"`
}

type FastRegPersonalResp struct {
	TaskID string `json:"taskid"`
}

type FastRegPersonalStatus struct {
	Status            int    `json:"status"`
	AppID             string `json:"appid,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// ----- beta (复用主体试用版) -----

type FastRegBetaReq struct {
	Name               string `json:"name"`
	Code               string `json:"code"`
	CodeType           int    `json:"code_type"`
	LegalPersonaWechat string `json:"legal_persona_wechat"`
	LegalPersonaName   string `json:"legal_persona_name"`
	ComponentPhone     string `json:"component_phone"`
}

type FastRegBetaResp struct {
	UniqueID string `json:"unique_id"`
}

type FastRegBetaStatus struct {
	Status            int    `json:"status"`
	AppID             string `json:"appid,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// ----- admin rebind -----

type RebindAdminQrcode struct {
	TaskID    string `json:"taskid"`
	QrcodeURL string `json:"qrcode_url"`
}
