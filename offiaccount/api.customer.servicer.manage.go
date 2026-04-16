package offiaccount

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// UploadKFHeadImg 设置客服头像。
// kfAccount: 完整客服账号，格式为：账号前缀@公众号微信号
// filename: 头像文件名
// reader: 头像文件内容读取器
func (c *Client) UploadKFHeadImg(ctx context.Context, kfAccount string, filename string, reader io.Reader) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: UploadKFHeadImg: read file: %w", err)
	}
	params := url.Values{}
	params.Set("access_token", token)
	params.Set("kf_account", kfAccount)
	path := "/customservice/kfaccount/uploadheadimg?" + params.Encode()
	var result Resp
	if err = c.doPostMultipartFile(ctx, path, "media", filename, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddKFAccount 添加客服账号
// kfAccount: 完整客服账号，格式为：账号前缀@公众号微信号
// nickname: 客服昵称，最长16个字
func (c *Client) AddKFAccount(ctx context.Context, kfAccount, nickname string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/add?access_token=%s", token)

	// 构造请求体
	body := map[string]any{
		"kf_account": kfAccount,
		"nickname":   nickname,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateKFAccount 修改客服账号
// kfAccount: 完整客服账号，格式为：账号前缀@公众号微信号
// nickname: 客服昵称，最长16个字
func (c *Client) UpdateKFAccount(ctx context.Context, kfAccount, nickname string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/update?access_token=%s", token)

	// 构造请求体
	body := map[string]any{
		"kf_account": kfAccount,
		"nickname":   nickname,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DelKFAccount 删除客服账号
func (c *Client) DelKFAccount(ctx context.Context, kfAccount string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/del?access_token=%s", token)

	// 构造请求体
	body := map[string]any{
		"kf_account": kfAccount,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// InviteKFWorker 邀请绑定客服账号
// kfAccount: 完整客服帐号，格式为：帐号前缀@公众号微信号
// inviteWX: 接收绑定邀请的客服微信号
func (c *Client) InviteKFWorker(ctx context.Context, kfAccount, inviteWX string) (*Resp, error) {
	// 获取access_token
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}

	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/inviteworker?access_token=%s", token)

	// 构造请求体
	body := map[string]any{
		"kf_account": kfAccount,
		"invite_wx":  inviteWX,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetOnlineKFList 获取在线客服列表
func (c *Client) GetOnlineKFList(ctx context.Context) (*KFOnlineListResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/customservice/getonlinekflist?access_token=%s", token)

	// 发送请求
	var result KFOnlineListResp
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKFList 获取所有客服账号
func (c *Client) GetKFList(ctx context.Context) (*KFListResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/customservice/getkflist?access_token=%s", token)

	// 发送请求
	var result KFListResp
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
