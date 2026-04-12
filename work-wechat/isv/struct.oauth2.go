package isv

// UserInfo3rdResp 是 service/auth/getuserinfo3rd 的响应。
// 企业成员 / 非企业成员 / 外部联系人返回的字段不同,全部用 omitempty 填进同一个结构体。
// 注意:官方字段名大小写混用,必须原样映射。
type UserInfo3rdResp struct {
	CorpID         string `json:"CorpId"`
	UserID         string `json:"UserId"`          // 企业成员
	DeviceID       string `json:"DeviceId"`
	UserTicket     string `json:"user_ticket"`     // 企业成员才返回,用于后续换详情
	ExpiresIn      int    `json:"expires_in"`      // user_ticket 有效期(秒)
	OpenUserID     string `json:"open_userid"`     // 跨服务商匿名 id
	OpenID         string `json:"OpenId"`          // 非企业成员时的微信 openid
	ExternalUserID string `json:"external_userid"` // 外部联系人
}

// UserDetail3rdResp 是 service/auth/getuserdetail3rd 的响应。
// 注意:此接口对敏感字段有调用者备案要求,调用前请确认合规。
type UserDetail3rdResp struct {
	CorpID  string `json:"corpid"`
	UserID  string `json:"userid"`
	Gender  string `json:"gender"` // 1 男 / 2 女
	Avatar  string `json:"avatar"`
	QRCode  string `json:"qr_code"`
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
	BizMail string `json:"biz_mail"`
	Address string `json:"address"`
}
