package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// Code2SessionResult is the result of jscode2session
type Code2SessionResult struct {
	core.Resp
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
}

// Watermark contains appid and timestamp for phone number verification
type Watermark struct {
	AppId     string `json:"appid"`
	Timestamp int64  `json:"timestamp"`
}

// PhoneInfo contains decrypted phone number details
type PhoneInfo struct {
	PhoneNumber     string    `json:"phoneNumber"`
	PurePhoneNumber string    `json:"purePhoneNumber"`
	CountryCode     string    `json:"countryCode"`
	Watermark       Watermark `json:"watermark"`
}

// PhoneNumberResult is the result of GetPhoneNumber
type PhoneNumberResult struct {
	core.Resp
	PhoneInfo PhoneInfo `json:"phone_info"`
}

// PaidUnionIdResult is the result of GetPaidUnionId
type PaidUnionIdResult struct {
	core.Resp
	Unionid string `json:"unionid"`
}
