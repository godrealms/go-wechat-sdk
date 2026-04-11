package oplatform

// 本文件汇总代小程序开发管理 (WxaAdmin) 所有子族的请求/响应 DTO。
// 各子族（account/category/domain/tester/code/release）的结构体
// 按顺序追加到下面的分隔注释段。

// ----- account -----

type WxaSetNicknameReq struct {
	Nickname     string `json:"nick_name"`
	IDCard       string `json:"id_card,omitempty"`
	License      string `json:"license,omitempty"`
	NamingOther1 string `json:"naming_other_stuff_1,omitempty"`
	NamingOther2 string `json:"naming_other_stuff_2,omitempty"`
	NamingOther3 string `json:"naming_other_stuff_3,omitempty"`
	NamingOther4 string `json:"naming_other_stuff_4,omitempty"`
	NamingOther5 string `json:"naming_other_stuff_5,omitempty"`
}

type WxaSetNicknameResp struct {
	Wording string `json:"wording,omitempty"`
	AuditID string `json:"audit_id,omitempty"`
}

type WxaQueryNicknameResp struct {
	Nickname   string `json:"nickname"`
	AuditStat  int    `json:"audit_stat"`
	FailReason string `json:"fail_reason,omitempty"`
	CreateTime int64  `json:"create_time"`
	AuditTime  int64  `json:"audit_time"`
}

type WxaCheckNicknameResp struct {
	HitCondition bool   `json:"hit_condition"`
	Wording      string `json:"wording,omitempty"`
}
