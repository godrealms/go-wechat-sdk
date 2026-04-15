package offiaccount

import (
	"context"
	"fmt"
)

// GetUserTitleUrl 获取添加发票链接
// req: 获取添加发票链接请求参数
func (c *Client) GetUserTitleUrl(ctx context.Context, req *GetUserTitleUrlRequest) (*GetUserTitleUrlResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/biz/getusertitleurl?access_token=%s", token)

	var result GetUserTitleUrlResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSelectTitleUrl 获取选择发票抬头链接
// req: 获取选择发票抬头链接请求参数
func (c *Client) GetSelectTitleUrl(ctx context.Context, req *GetSelectTitleUrlRequest) (*GetSelectTitleUrlResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/biz/getselecttitleurl?access_token=%s", token)

	var result GetSelectTitleUrlResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ScanTitle 扫描二维码获取抬头
// req: 扫描二维码获取抬头请求参数
func (c *Client) ScanTitle(ctx context.Context, req *ScanTitleRequest) (*ScanTitleResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/scantitle?access_token=%s", token)

	var result ScanTitleResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
