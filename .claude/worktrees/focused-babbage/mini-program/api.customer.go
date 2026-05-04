package mini_program

import (
	"fmt"
	"net/url"
)

// SendCustomerMessage 发送客服消息给用户
// POST /cgi-bin/message/custom/send (access_token in URL)
func (c *Client) SendCustomerMessage(req *SendCustomerMessageRequest) error {
	path := fmt.Sprintf("/cgi-bin/message/custom/send?access_token=%s", c.GetAccessToken())
	result := &SendCustomerMessageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	return result.GetError()
}

// SetTyping 下发客服当前输入状态给用户
// POST /cgi-bin/message/custom/typing (access_token in URL)
func (c *Client) SetTyping(toUser string, command TypingStatus) error {
	path := fmt.Sprintf("/cgi-bin/message/custom/typing?access_token=%s", c.GetAccessToken())
	req := &SetTypingRequest{ToUser: toUser, Command: command}
	result := &SendCustomerMessageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	return result.GetError()
}

// GetKfAccountList 获取客服账号列表
// GET /cgi-bin/customservice/getkfaccountlist
func (c *Client) GetKfAccountList() (*KfAccountListResult, error) {
	query := c.TokenQuery(url.Values{})
	result := &KfAccountListResult{}
	err := c.Https.Get(c.Ctx, "/cgi-bin/customservice/getkfaccountlist", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
