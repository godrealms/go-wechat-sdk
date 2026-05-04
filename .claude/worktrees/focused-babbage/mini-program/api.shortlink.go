package mini_program

import "fmt"

// GenerateShortLink 获取小程序 Short Link
// POST /wxa/genwxashortlink
func (c *Client) GenerateShortLink(req *GenerateShortLinkRequest) (*GenerateShortLinkResult, error) {
	path := fmt.Sprintf("/wxa/genwxashortlink?access_token=%s", c.GetAccessToken())
	result := &GenerateShortLinkResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
