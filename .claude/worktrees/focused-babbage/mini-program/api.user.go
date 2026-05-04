package mini_program

import (
	"fmt"
	"net/url"
)

// GetPhoneNumber 获取手机号
// POST /wxa/business/getuserphonenumber (access_token in URL)
func (c *Client) GetPhoneNumber(code string) (*PhoneNumberResult, error) {
	path := fmt.Sprintf("/wxa/business/getuserphonenumber?access_token=%s", c.GetAccessToken())
	body := map[string]string{"code": code}
	result := &PhoneNumberResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetPaidUnionId 用户支付完成后获取 UnionId (需要先有支付记录)
// GET /wxa/getpaidunionid
func (c *Client) GetPaidUnionId(openid, transactionId string) (*PaidUnionIdResult, error) {
	query := c.TokenQuery(url.Values{
		"openid":         {openid},
		"transaction_id": {transactionId},
	})
	result := &PaidUnionIdResult{}
	err := c.Https.Get(c.Ctx, "/wxa/getpaidunionid", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
