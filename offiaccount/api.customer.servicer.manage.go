package offiaccount

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// UploadKFHeadImg 设置客服头像
// kfAccount: 完整客服账号，格式为：账号前缀@公众号微信号
// filename: 头像文件名
// reader: 头像文件内容读取器
func (c *Client) UploadKFHeadImg(kfAccount string, filename string, reader io.Reader) (*Resp, error) {
	// 获取access_token
	token := c.GetAccessToken()
	if token == "" {
		return nil, fmt.Errorf("get access token failed")
	}

	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("kf_account", kfAccount)

	path := fmt.Sprintf("/customservice/kfaccount/uploadheadimg?%s", params.Encode())

	// 创建multipart表单
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file field: %v", err)
	}

	// 复制文件内容
	_, err = io.Copy(part, reader)
	if err != nil {
		return nil, fmt.Errorf("copying file contents failed: %v", err)
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("closing writer failed: %v", err)
	}

	// 构建完整URL
	fullURL := c.Https.BaseURL + path

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", fullURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("the http request was created failed: %v", err)
	}

	// 设置Content-Type
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.Https.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending http request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result Resp
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
	}

	return &result, nil
}

// AddKFAccount 添加客服账号
// kfAccount: 完整客服账号，格式为：账号前缀@公众号微信号
// nickname: 客服昵称，最长16个字
func (c *Client) AddKFAccount(kfAccount, nickname string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/add?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"kf_account": kfAccount,
		"nickname":   nickname,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateKFAccount 修改客服账号
// kfAccount: 完整客服账号，格式为：账号前缀@公众号微信号
// nickname: 客服昵称，最长16个字
func (c *Client) UpdateKFAccount(kfAccount, nickname string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/update?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"kf_account": kfAccount,
		"nickname":   nickname,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DelKFAccount 删除客服账号
func (c *Client) DelKFAccount(kfAccount string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/del?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"kf_account": kfAccount,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// InviteKFWorker 邀请绑定客服账号
// kfAccount: 完整客服帐号，格式为：帐号前缀@公众号微信号
// inviteWX: 接收绑定邀请的客服微信号
func (c *Client) InviteKFWorker(kfAccount, inviteWX string) (*Resp, error) {
	// 获取access_token
	token := c.GetAccessToken()
	if token == "" {
		return nil, fmt.Errorf("get access token failed")
	}

	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfaccount/inviteworker?access_token=%s", token)

	// 构造请求体
	body := map[string]interface{}{
		"kf_account": kfAccount,
		"invite_wx":  inviteWX,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetOnlineKFList 获取在线客服列表
func (c *Client) GetOnlineKFList() (*KFOnlineListResp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/customservice/getonlinekflist?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result KFOnlineListResp
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKFList 获取所有客服账号
func (c *Client) GetKFList() (*KFListResp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/customservice/getkflist?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result KFListResp
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
