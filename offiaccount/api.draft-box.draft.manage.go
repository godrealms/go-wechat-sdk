package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// DraftSwitch 草稿箱开关设置
//
// 无论是否查询，官方接口都要求 POST；checkonly=1 只是 query 参数，用于仅查询状态不做切换。
// https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Temporary_MP_Switch.html
//
// checkOnly: true 时仅查询当前开关状态，不做切换
func (c *Client) DraftSwitch(ctx context.Context, checkOnly bool) (*DraftSwitchResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	if checkOnly {
		params.Add("checkonly", "1")
	}
	path := fmt.Sprintf("/cgi-bin/draft/switch?%s", params.Encode())

	var result DraftSwitchResult
	if err = c.doPost(ctx, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddDraft 新增草稿
// articles: 图文素材集合
func (c *Client) AddDraft(ctx context.Context, articles []*DraftArticle) (*AddDraftResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/add?access_token=%s", token)

	// 构造请求体
	body := map[string]interface{}{
		"articles": articles,
	}

	// 发送请求
	var result AddDraftResult
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDraft 获取草稿详情
// mediaID: 要获取的草稿的media_id
func (c *Client) GetDraft(ctx context.Context, mediaID string) (*GetDraftResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/get?access_token=%s", token)

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 发送请求
	var result GetDraftResult
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteDraft 删除草稿
// mediaID: 要删除的草稿的media_id
func (c *Client) DeleteDraft(ctx context.Context, mediaID string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/delete?access_token=%s", token)

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateDraft 更新草稿
// req: 更新草稿请求参数
func (c *Client) UpdateDraft(ctx context.Context, req *UpdateDraftRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/update?access_token=%s", token)

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDraftCount 获取草稿总数
func (c *Client) GetDraftCount(ctx context.Context) (*DraftCountResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/count?access_token=%s", token)

	// 发送请求
	var result DraftCountResult
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchGetDraft 批量获取草稿
// req: 批量获取草稿请求参数
func (c *Client) BatchGetDraft(ctx context.Context, req *BatchGetDraftRequest) (*BatchGetDraftResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/batchget?access_token=%s", token)

	// 发送请求
	var result BatchGetDraftResult
	err = c.doPost(ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
