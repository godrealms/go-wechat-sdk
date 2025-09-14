package offiaccount

import (
	"fmt"
)

// GetCurrentAutoReplyInfo 获取自动回复规则
func (c *Client) GetCurrentAutoReplyInfo() (*ReplyResp, error) {
	path := fmt.Sprintf("/cgi-bin/get_current_autoreply_info?access_token=%s", c.GetAccessToken())
	result := &ReplyResp{}
	err := c.Https.Get(c.ctx, path, nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
