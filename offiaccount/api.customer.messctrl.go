package offiaccount

import (
	"context"
	"fmt"
)

// CreateKFSession 创建会话
// kfAccount: 完整客服账号
// openID: 粉丝的openid
func (c *Client) CreateKFSession(ctx context.Context, kfAccount, openID string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/create?access_token=%s", token)

	// 构造请求体
	body := map[string]any{
		"kf_account": kfAccount,
		"openid":     openID,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CloseKFSession 关闭会话
// kfAccount: 完整客服账号
// openID: 粉丝的openid
func (c *Client) CloseKFSession(ctx context.Context, kfAccount, openID string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/close?access_token=%s", token)

	// 构造请求体
	body := map[string]any{
		"kf_account": kfAccount,
		"openid":     openID,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKFCustomerSession 获取客户会话状态
// openID: 粉丝的openid
func (c *Client) GetKFCustomerSession(ctx context.Context, openID string) (*KFCustomerSessionInfo, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/getsession?access_token=%s&openid=%s", token, openID)

	// 发送请求
	var result KFCustomerSessionInfo
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKFSessionList 获取客服会话列表
// kfAccount: 完整客服账号
func (c *Client) GetKFSessionList(ctx context.Context, kfAccount string) (*KFSessionListResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/getsessionlist?access_token=%s&kf_account=%s", token, kfAccount)

	// 发送请求
	var result KFSessionListResp
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWaitCaseList 获取未接入会话列表
func (c *Client) GetWaitCaseList(ctx context.Context) (*WaitCaseListResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/getwaitcase?access_token=%s", token)

	// 发送请求
	var result WaitCaseListResp
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
