package mini_program

import "context"

// GetPhoneNumberReq 获取用户手机号请求。
type GetPhoneNumberReq struct {
	Code string `json:"code"`
}

// PhoneInfo 手机号信息。
type PhoneInfo struct {
	PhoneNumber     string    `json:"phoneNumber"`
	PurePhoneNumber string    `json:"purePhoneNumber"`
	CountryCode     string    `json:"countryCode"`
	Watermark       Watermark `json:"watermark"`
}

// Watermark 数据水印。
type Watermark struct {
	AppID     string `json:"appid"`
	Timestamp int64  `json:"timestamp"`
}

// GetPhoneNumberResp 获取用户手机号响应。
type GetPhoneNumberResp struct {
	PhoneInfo PhoneInfo `json:"phone_info"`
}

// GetPhoneNumber 获取用户手机号（用前端 getPhoneNumber 返回的 code 换取）。
func (c *Client) GetPhoneNumber(ctx context.Context, code string) (*GetPhoneNumberResp, error) {
	body := &GetPhoneNumberReq{Code: code}
	var resp GetPhoneNumberResp
	if err := c.doPost(ctx, "/wxa/business/getuserphonenumber", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
