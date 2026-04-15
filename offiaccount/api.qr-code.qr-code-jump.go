package offiaccount

import (
	"context"
	"fmt"
)

// GetQRCodeJump 获取已设置的二维码规则
// req: 获取二维码跳转规则请求参数
func (c *Client) GetQRCodeJump(ctx context.Context, req *GetQRCodeJumpRequest) (*GetQRCodeJumpResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/wxopen/qrcodejumpget?access_token=%s", token)

	var result GetQRCodeJumpResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddQRCodeJump 增加二维码规则
// req: 增加二维码规则请求参数
func (c *Client) AddQRCodeJump(ctx context.Context, req *AddQRCodeJumpRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/wxopen/qrcodejumpadd?access_token=%s", token)

	var result Resp
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PublishQRCodeJump 发布已设置的二维码规则
// prefix: 二维码规则
func (c *Client) PublishQRCodeJump(ctx context.Context, prefix string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/wxopen/qrcodejumppublish?access_token=%s", token)

	req := &PublishQRCodeJumpRequest{
		Prefix: prefix,
	}

	var result Resp
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteQRCodeJump 删除已设置的二维码规则
// req: 删除二维码规则请求参数
func (c *Client) DeleteQRCodeJump(ctx context.Context, req *DeleteQRCodeJumpRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/wxopen/qrcodejumpdelete?access_token=%s", token)

	var result Resp
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
