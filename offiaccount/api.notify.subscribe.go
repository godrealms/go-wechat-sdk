package offiaccount

import (
	"context"
	"errors"
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
	err = c.Https.Post(ctx, path, form, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}
