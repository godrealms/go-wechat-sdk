package mini_program

import "fmt"

// GenerateScheme 获取小程序 scheme 码
// POST /wxa/generatescheme
func (c *Client) GenerateScheme(req *GenerateSchemeRequest) (*GenerateSchemeResult, error) {
	path := fmt.Sprintf("/wxa/generatescheme?access_token=%s", c.GetAccessToken())
	result := &GenerateSchemeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// QueryScheme 查询小程序 scheme 码
// POST /wxa/queryscheme
func (c *Client) QueryScheme(scheme string) (*QuerySchemeResult, error) {
	path := fmt.Sprintf("/wxa/queryscheme?access_token=%s", c.GetAccessToken())
	body := map[string]string{"scheme": scheme}
	result := &QuerySchemeResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
