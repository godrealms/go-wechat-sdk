package mini_program

import "context"

// MsgSecCheckReq 文本内容安全检测请求。
type MsgSecCheckReq struct {
	Content   string `json:"content"`
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	OpenID    string `json:"openid"`
	Title     string `json:"title,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// MsgSecCheckResp 文本内容安全检测响应。
type MsgSecCheckResp struct {
	TraceID string           `json:"trace_id"`
	Result  SecCheckResult   `json:"result"`
	Detail  []SecCheckDetail `json:"detail"`
}

// SecCheckResult 安全检测结果。
type SecCheckResult struct {
	Suggest string `json:"suggest"`
	Label   int    `json:"label"`
}

// SecCheckDetail 安全检测详情。
type SecCheckDetail struct {
	Strategy string `json:"strategy"`
	ErrCode  int    `json:"errcode"`
	Suggest  string `json:"suggest"`
	Label    int    `json:"label"`
	Prob     int    `json:"prob"`
	Keyword  string `json:"keyword"`
}

// MediaCheckAsyncReq 媒体内容异步安全检测请求。
type MediaCheckAsyncReq struct {
	MediaURL  string `json:"media_url"`
	MediaType int    `json:"media_type"`
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	OpenID    string `json:"openid"`
}

// MediaCheckAsyncResp 媒体内容异步安全检测响应。
type MediaCheckAsyncResp struct {
	TraceID string `json:"trace_id"`
}

// MsgSecCheck 文本内容安全检测。
func (c *Client) MsgSecCheck(ctx context.Context, req *MsgSecCheckReq) (*MsgSecCheckResp, error) {
	var resp MsgSecCheckResp
	if err := c.doPost(ctx, "/wxa/msg_sec_check", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MediaCheckAsync 图片/音频异步内容安全检测。结果通过回调通知推送。
func (c *Client) MediaCheckAsync(ctx context.Context, req *MediaCheckAsyncReq) (*MediaCheckAsyncResp, error) {
	var resp MediaCheckAsyncResp
	if err := c.doPost(ctx, "/wxa/media_check_async", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
