package offiaccount

import "fmt"

// OpenArticleComment 打开已群发文章评论
// req: 打开已群发文章评论请求参数
func (c *Client) OpenArticleComment(req *OpenArticleCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/open?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CloseArticleComment 关闭已群发文章评论
// req: 关闭已群发文章评论请求参数
func (c *Client) CloseArticleComment(req *CloseArticleCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/close?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListComment 查看指定文章的评论数据
// req: 查看指定文章的评论数据请求参数
func (c *Client) ListComment(req *ListCommentRequest) (*ListCommentResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/list?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result ListCommentResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ElectComment 评论标记精选
// req: 评论标记精选请求参数
func (c *Client) ElectComment(req *ElectCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/markelect?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UnElectComment 评论取消精选
// req: 评论取消精选请求参数
func (c *Client) UnElectComment(req *UnElectCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/unmarkelect?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteComment 删除评论
// req: 删除评论请求参数
func (c *Client) DeleteComment(req *DeleteCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/delete?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ReplyComment 回复评论
// req: 回复评论请求参数
func (c *Client) ReplyComment(req *ReplyCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/reply/add?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteReplyComment 删除回复
// req: 删除回复请求参数
func (c *Client) DeleteReplyComment(req *DeleteReplyCommentRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/comment/reply/delete?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
