package offiaccount

import "net/url"

// GetUserInfo 获取用户基本信息
// openid: 普通用户的标识，对当前公众号唯一
// lang: 返回国家地区语言版本，zh_CN 简体，zh_TW 繁体，en 英语，默认为zh_CN
func (c *Client) GetUserInfo(openid string, lang string) (*UserInfo, error) {
	// 构造请求URL和查询参数
	path := "/cgi-bin/user/info"
	params := url.Values{}
	params.Add("openid", openid)

	if lang != "" {
		params.Add("lang", lang)
	}

	// 发送请求
	var result UserInfo
	err := c.Https.Get(c.ctx, path, params, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchGetUserInfo 批量获取用户基本信息
// req: 批量获取用户基本信息请求参数
func (c *Client) BatchGetUserInfo(req *BatchGetUserInfoRequest) (*BatchGetUserInfoResult, error) {
	// 构造请求URL
	path := "/cgi-bin/user/info/batchget"

	// 发送请求
	var result BatchGetUserInfoResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateRemark 设置用户备注名
// openid: 用户的标识，对当前公众号唯一
// remark: 备注名
func (c *Client) UpdateRemark(openid string, remark string) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/user/info/updateremark"

	// 构造请求体
	req := map[string]string{
		"openid": openid,
		"remark": remark,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFans 获取关注用户列表
// nextOpenid: 上一批列表的最后一个OPENID，不填默认从头开始拉取
func (c *Client) GetFans(nextOpenid string) (*GetFansResult, error) {
	// 构造请求URL和查询参数
	path := "/cgi-bin/user/get"
	params := url.Values{}

	if nextOpenid != "" {
		params.Add("next_openid", nextOpenid)
	}

	// 发送请求
	var result GetFansResult
	err := c.Https.Get(c.ctx, path, params, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBlacklist 获取公众号的黑名单列表
// beginOpenid: 起始OpenID，为空时从开头拉取
func (c *Client) GetBlacklist(beginOpenid string) (*GetBlacklistResult, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/members/getblacklist"

	// 构造请求体
	req := map[string]string{
		"begin_openid": beginOpenid,
	}

	// 发送请求
	var result GetBlacklistResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchBlacklist 拉黑用户
// openidList: 需要拉黑的openid列表，一次最多拉黑20个用户
func (c *Client) BatchBlacklist(openidList []string) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/members/batchblacklist"

	// 构造请求体
	req := map[string]interface{}{
		"openid_list": openidList,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchUnblacklist 取消拉黑用户
// openidList: 需要取消拉黑的openid列表，一次最多取消20个用户
func (c *Client) BatchUnblacklist(openidList []string) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/members/batchunblacklist"

	// 构造请求体
	req := map[string]interface{}{
		"openid_list": openidList,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
