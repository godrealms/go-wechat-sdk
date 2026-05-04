package mini_program

import "fmt"

// GenerateUrlLink 获取小程序 URL Link
// POST /wxa/generate_urllink
func (c *Client) GenerateUrlLink(req *GenerateUrlLinkRequest) (*GenerateUrlLinkResult, error) {
	path := fmt.Sprintf("/wxa/generate_urllink?access_token=%s", c.GetAccessToken())
	result := &GenerateUrlLinkResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// QueryUrlLink 查询小程序 URL Link
// POST /wxa/query_urllink
func (c *Client) QueryUrlLink(urlLink string) (*QueryUrlLinkResult, error) {
	path := fmt.Sprintf("/wxa/query_urllink?access_token=%s", c.GetAccessToken())
	body := map[string]string{"url_link": urlLink}
	result := &QueryUrlLinkResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
