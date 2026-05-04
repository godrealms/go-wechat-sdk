package mini_program

import "fmt"

// SendSubscribeMessage 发送订阅消息
// POST /cgi-bin/message/subscribe/send (access_token in URL)
func (c *Client) SendSubscribeMessage(req *SendSubscribeMessageRequest) error {
	path := fmt.Sprintf("/cgi-bin/message/subscribe/send?access_token=%s", c.GetAccessToken())
	result := &SendSubscribeMessageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	return result.GetError()
}
