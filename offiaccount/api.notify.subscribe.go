package offiaccount

import (
	"context"
	"fmt"
)

// TemplateSubscribe 发送一次性订阅消息
func (c *Client) TemplateSubscribe(ctx context.Context, form *TemplateSubscribeReq) (err error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/cgi-bin/message/template/subscribe?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, form, result); err != nil {
		return err
	}
	return nil
}
