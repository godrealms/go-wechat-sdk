package offiaccount

import "fmt"

// GetKFMsgList 获取聊天记录
// req: 获取聊天记录请求参数
func (c *Client) GetKFMsgList(req *KFGetMsgListRequest) (*MsgListResp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/msgrecord/getmsglist?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result MsgListResp
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SetKFTyping 设置客服输入状态
// req: 客服输入状态请求参数
func (c *Client) SetKFTyping(req *KFTypingRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/message/custom/typing?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.Ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SendKFMessage 发送客服消息
// msg: 客服消息
func (c *Client) SendKFMessage(msg *KFMessage) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/message/custom/send?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.Ctx, path, msg, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
