package offiaccount

import (
	"context"
	"fmt"
)

// GetKFMsgList 获取聊天记录
// req: 获取聊天记录请求参数
func (c *Client) GetKFMsgList(ctx context.Context, req *KFGetMsgListRequest) (*MsgListResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/customservice/msgrecord/getmsglist?access_token=%s", token)

	var result MsgListResp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SetKFTyping 设置客服输入状态
// req: 客服输入状态请求参数
func (c *Client) SetKFTyping(ctx context.Context, req *KFTypingRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/custom/typing?access_token=%s", token)

	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SendKFMessage 发送客服消息
// msg: 客服消息
func (c *Client) SendKFMessage(ctx context.Context, msg *KFMessage) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/custom/send?access_token=%s", token)

	var result Resp
	if err := c.Https.Post(ctx, path, msg, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
