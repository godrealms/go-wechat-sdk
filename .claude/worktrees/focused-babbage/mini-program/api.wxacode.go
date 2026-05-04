package mini_program

import "fmt"

// GetQRCode 获取小程序码 (适用于需要的码数量较少的业务场景，总共生成的码不超过10万个)
// POST /wxa/getwxacode — returns raw PNG bytes
func (c *Client) GetQRCode(req *QRCodeRequest) ([]byte, error) {
	path := fmt.Sprintf("/wxa/getwxacode?access_token=%s", c.GetAccessToken())
	return c.Https.PostBinary(c.Ctx, path, req)
}

// GetUnlimited 获取不限制的小程序码 (适用于需要的码数量极多或者无法预知总数量的业务场景)
// POST /wxa/getwxacodeunlimit — returns raw PNG bytes
func (c *Client) GetUnlimited(req *UnlimitedQRCodeRequest) ([]byte, error) {
	path := fmt.Sprintf("/wxa/getwxacodeunlimit?access_token=%s", c.GetAccessToken())
	return c.Https.PostBinary(c.Ctx, path, req)
}

// CreateQRCode 获取小程序二维码 (适用于需要的码数量较少的业务场景，生成的码永久有效)
// POST /cgi-bin/wxaapp/createwxaqrcode — returns raw PNG bytes
func (c *Client) CreateQRCode(req *CreateQRCodeRequest) ([]byte, error) {
	path := fmt.Sprintf("/cgi-bin/wxaapp/createwxaqrcode?access_token=%s", c.GetAccessToken())
	return c.Https.PostBinary(c.Ctx, path, req)
}
