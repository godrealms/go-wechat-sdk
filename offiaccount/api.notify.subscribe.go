package offiaccount

import (
	"errors"
	"fmt"
)

// TemplateSubscribe 发送一次性订阅消息
func (c *Client) TemplateSubscribe(form *TemplateSubscribeReq) (err error) {
	path := fmt.Sprintf("/cgi-bin/message/template/subscribe?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err = c.Https.Post(c.ctx, path, form, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}
