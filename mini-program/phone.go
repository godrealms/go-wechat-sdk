package mini_program

import (
	"context"
	"fmt"
)

// GetPhoneNumberReq is the request body for the get-phone-number API.
type GetPhoneNumberReq struct {
	Code string `json:"code"`
}

// PhoneInfo holds the decrypted phone number returned by the WeChat API.
type PhoneInfo struct {
	PhoneNumber     string    `json:"phoneNumber"`
	PurePhoneNumber string    `json:"purePhoneNumber"`
	CountryCode     string    `json:"countryCode"`
	Watermark       Watermark `json:"watermark"`
}

// Watermark is the data watermark embedded in WeChat sensitive data responses.
type Watermark struct {
	AppID     string `json:"appid"`
	Timestamp int64  `json:"timestamp"`
}

// GetPhoneNumberResp is the response from the get-phone-number API.
type GetPhoneNumberResp struct {
	PhoneInfo PhoneInfo `json:"phone_info"`
}

// GetPhoneNumber exchanges the code returned by the front-end getPhoneNumber callback for the user's phone number.
func (c *Client) GetPhoneNumber(ctx context.Context, code string) (*GetPhoneNumberResp, error) {
	if code == "" {
		return nil, fmt.Errorf("mini_program: GetPhoneNumber: code is required")
	}
	body := &GetPhoneNumberReq{Code: code}
	var resp GetPhoneNumberResp
	if err := c.doPost(ctx, "/wxa/business/getuserphonenumber", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
