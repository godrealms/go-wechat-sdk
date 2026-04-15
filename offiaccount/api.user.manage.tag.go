package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// GetTags 获取公众号已创建的标签列表
func (c *Client) GetTags(ctx context.Context) (*GetTagsResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := "/cgi-bin/tags/get"
	params := url.Values{"access_token": {token}}

	// 发送请求
	var result GetTagsResult
	if err := c.Https.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateTag 创建标签
// name: 标签名（30个字符以内）
func (c *Client) CreateTag(ctx context.Context, name string) (*CreateTagResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/create?access_token=%s", token)

	// 构造请求体
	req := &CreateTagRequest{
		Tag: &CreateTag{
			Name: name,
		},
	}

	// 发送请求
	var result CreateTagResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateTag 修改标签
// id: 标签ID
// name: 标签名
func (c *Client) UpdateTag(ctx context.Context, id int64, name string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/update?access_token=%s", token)

	// 构造请求体
	req := &UpdateTagRequest{
		Tag: &UpdateTag{
			Id:   id,
			Name: name,
		},
	}

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteTag 删除标签
// id: 标签ID
func (c *Client) DeleteTag(ctx context.Context, id int64) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/delete?access_token=%s", token)

	// 构造请求体
	req := &DeleteTagRequest{
		Tag: &DeleteTag{
			Id: id,
		},
	}

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTagFans 获取标签下粉丝列表
// req: 获取标签下粉丝列表请求参数
func (c *Client) GetTagFans(ctx context.Context, req *GetTagFansRequest) (*GetTagFansResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/user/tag/get?access_token=%s", token)

	// 发送请求
	var result GetTagFansResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchTagging 批量为用户打标签
// openidList: 粉丝openid列表，最多50个
// tagid: 标签id
func (c *Client) BatchTagging(ctx context.Context, openidList []string, tagid int64) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/members/batchtagging?access_token=%s", token)

	// 构造请求体
	req := &BatchTaggingRequest{
		OpenidList: openidList,
		Tagid:      tagid,
	}

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchUntagging 批量为用户取消标签
// openidList: 粉丝openid列表
// tagid: 标签id
func (c *Client) BatchUntagging(ctx context.Context, openidList []string, tagid int64) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/members/batchuntagging?access_token=%s", token)

	// 构造请求体
	req := &BatchUntaggingRequest{
		OpenidList: openidList,
		Tagid:      tagid,
	}

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTagidList 获取用户身上的标签列表
// openid: 用户openid
func (c *Client) GetTagidList(ctx context.Context, openid string) (*GetTagidListResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/tags/getidlist?access_token=%s", token)

	// 构造请求体
	req := &GetTagidListRequest{
		Openid: openid,
	}

	// 发送请求
	var result GetTagidListResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
