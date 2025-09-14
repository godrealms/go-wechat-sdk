package offiaccount

import "fmt"

// CreateKFSession 创建会话
// kfAccount: 完整客服账号
// openID: 粉丝的openid
func (c *Client) CreateKFSession(kfAccount, openID string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/create?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"kf_account": kfAccount,
		"openid":     openID,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CloseKFSession 关闭会话
// kfAccount: 完整客服账号
// openID: 粉丝的openid
func (c *Client) CloseKFSession(kfAccount, openID string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/close?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"kf_account": kfAccount,
		"openid":     openID,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKFCustomerSession 获取客户会话状态
// openID: 粉丝的openid
func (c *Client) GetKFCustomerSession(openID string) (*KFCustomerSessionInfo, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/getsession?access_token=%s&openid=%s", c.GetAccessToken(), openID)

	// 发送请求
	var result KFCustomerSessionInfo
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKFSessionList 获取客服会话列表
// kfAccount: 完整客服账号
func (c *Client) GetKFSessionList(kfAccount string) (*KFSessionListResp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/getsessionlist?access_token=%s&kf_account=%s", c.GetAccessToken(), kfAccount)

	// 发送请求
	var result KFSessionListResp
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWaitCaseList 获取未接入会话列表
func (c *Client) GetWaitCaseList() (*WaitCaseListResp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/customservice/kfsession/getwaitcase?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result WaitCaseListResp
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
