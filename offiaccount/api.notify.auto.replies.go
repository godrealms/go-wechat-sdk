package offiaccount

import (
	"context"
	"fmt"
)

// GetCurrentAutoReplyInfo 获取自动回复规则
func (c *Client) GetCurrentAutoReplyInfo(ctx context.Context) (*ReplyResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/get_current_autoreply_info?access_token=%s", token)
	result := &ReplyResp{}
	err = c.Https.Get(ctx, path, nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
