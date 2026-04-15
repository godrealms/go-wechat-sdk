package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// GetUserInfo retrieves the basic profile of a follower identified by openid.
// lang selects the language of the returned country/region field: zh_CN (simplified Chinese),
// zh_TW (traditional Chinese), or en (English); defaults to zh_CN when empty.
func (c *Client) GetUserInfo(ctx context.Context, openid string, lang string) (*UserInfo, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL和查询参数
	path := "/cgi-bin/user/info"
	params := url.Values{"access_token": {token}}
	params.Add("openid", openid)

	if lang != "" {
		params.Add("lang", lang)
	}

	// 发送请求
	var result UserInfo
	if err := c.doGet(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchGetUserInfo 批量获取用户基本信息
// req: 批量获取用户基本信息请求参数
func (c *Client) BatchGetUserInfo(ctx context.Context, req *BatchGetUserInfoRequest) (*BatchGetUserInfoResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/user/info/batchget?access_token=%s", token)

	// 发送请求
	var result BatchGetUserInfoResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateRemark 设置用户备注名
// openid: 用户的标识，对当前公众号唯一
// remark: 备注名
func (c *Client) UpdateRemark(ctx context.Context, openid string, remark string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/user/info/updateremark?access_token=%s", token)

	// 构造请求体
	req := map[string]string{
		"openid": openid,
		"remark": remark,
	}

	// 发送请求
	var result Resp
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFans 获取关注用户列表
// nextOpenid: 上一批列表的最后一个OPENID，不填默认从头开始拉取
func (c *Client) GetFans(ctx context.Context, nextOpenid string) (*GetFansResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL和查询参数
	path := "/cgi-bin/user/get"
	params := url.Values{"access_token": {token}}

	if nextOpenid != "" {
		params.Add("next_openid", nextOpenid)
	}

	// 发送请求
	var result GetFansResult
	if err := c.doGet(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBlacklist 获取公众号的黑名单列表
// beginOpenid: 起始OpenID，为空时从开头拉取
func (c *Client) GetBlacklist(ctx context.Context, beginOpenid string) (*GetBlacklistResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/members/getblacklist?access_token=%s", token)

	// 构造请求体
	req := map[string]string{
		"begin_openid": beginOpenid,
	}

	// 发送请求
	var result GetBlacklistResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchBlacklist 拉黑用户
// openidList: 需要拉黑的openid列表，一次最多拉黑20个用户
func (c *Client) BatchBlacklist(ctx context.Context, openidList []string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/members/batchblacklist?access_token=%s", token)

	// 构造请求体
	req := map[string]interface{}{
		"openid_list": openidList,
	}

	// 发送请求
	var result Resp
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchUnblacklist 取消拉黑用户
// openidList: 需要取消拉黑的openid列表，一次最多取消20个用户
func (c *Client) BatchUnblacklist(ctx context.Context, openidList []string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/members/batchunblacklist?access_token=%s", token)

	// 构造请求体
	req := map[string]interface{}{
		"openid_list": openidList,
	}

	// 发送请求
	var result Resp
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
