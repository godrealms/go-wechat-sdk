package offiaccount

import (
	"fmt"
	"net/url"
)

// DraftSwitch 草稿箱开关设置
// checkOnly: 仅检查状态时传true
func (c *Client) DraftSwitch(checkOnly bool) (*DraftSwitchResult, error) {
	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", c.GetAccessToken())
	if checkOnly {
		params.Add("checkonly", "1")
	}
	path := fmt.Sprintf("/cgi-bin/draft/switch?%s", params.Encode())

	var result DraftSwitchResult
	if checkOnly {
		err := c.Https.Get(c.ctx, path, nil, &result)
		if err != nil {
			return nil, err
		}
	} else {
		err := c.Https.Post(c.ctx, path, nil, &result)
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
}

// AddDraft 新增草稿
// articles: 图文素材集合
func (c *Client) AddDraft(articles []*DraftArticle) (*AddDraftResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/add?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"articles": articles,
	}

	// 发送请求
	var result AddDraftResult
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDraft 获取草稿详情
// mediaID: 要获取的草稿的media_id
func (c *Client) GetDraft(mediaID string) (*GetDraftResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/get?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 发送请求
	var result GetDraftResult
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteDraft 删除草稿
// mediaID: 要删除的草稿的media_id
func (c *Client) DeleteDraft(mediaID string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/delete?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateDraft 更新草稿
// req: 更新草稿请求参数
func (c *Client) UpdateDraft(req *UpdateDraftRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/update?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDraftCount 获取草稿总数
func (c *Client) GetDraftCount() (*DraftCountResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/count?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result DraftCountResult
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchGetDraft 批量获取草稿
// req: 批量获取草稿请求参数
func (c *Client) BatchGetDraft(req *BatchGetDraftRequest) (*BatchGetDraftResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/draft/batchget?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result BatchGetDraftResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
