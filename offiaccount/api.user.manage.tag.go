package offiaccount

// GetTags 获取公众号已创建的标签列表
func (c *Client) GetTags() (*GetTagsResult, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/get"

	// 发送请求
	var result GetTagsResult
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateTag 创建标签
// name: 标签名（30个字符以内）
func (c *Client) CreateTag(name string) (*CreateTagResult, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/create"

	// 构造请求体
	req := &CreateTagRequest{
		Tag: &CreateTag{
			Name: name,
		},
	}

	// 发送请求
	var result CreateTagResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateTag 修改标签
// id: 标签ID
// name: 标签名
func (c *Client) UpdateTag(id int64, name string) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/update"

	// 构造请求体
	req := &UpdateTagRequest{
		Tag: &UpdateTag{
			Id:   id,
			Name: name,
		},
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteTag 删除标签
// id: 标签ID
func (c *Client) DeleteTag(id int64) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/delete"

	// 构造请求体
	req := &DeleteTagRequest{
		Tag: &DeleteTag{
			Id: id,
		},
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTagFans 获取标签下粉丝列表
// req: 获取标签下粉丝列表请求参数
func (c *Client) GetTagFans(req *GetTagFansRequest) (*GetTagFansResult, error) {
	// 构造请求URL
	path := "/cgi-bin/user/tag/get"

	// 发送请求
	var result GetTagFansResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchTagging 批量为用户打标签
// openidList: 粉丝openid列表，最多50个
// tagid: 标签id
func (c *Client) BatchTagging(openidList []string, tagid int64) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/members/batchtagging"

	// 构造请求体
	req := &BatchTaggingRequest{
		OpenidList: openidList,
		Tagid:      tagid,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchUntagging 批量为用户取消标签
// openidList: 粉丝openid列表
// tagid: 标签id
func (c *Client) BatchUntagging(openidList []string, tagid int64) (*Resp, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/members/batchuntagging"

	// 构造请求体
	req := &BatchUntaggingRequest{
		OpenidList: openidList,
		Tagid:      tagid,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTagidList 获取用户身上的标签列表
// openid: 用户openid
func (c *Client) GetTagidList(openid string) (*GetTagidListResult, error) {
	// 构造请求URL
	path := "/cgi-bin/tags/getidlist"

	// 构造请求体
	req := &GetTagidListRequest{
		Openid: openid,
	}

	// 发送请求
	var result GetTagidListResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
