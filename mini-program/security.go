package mini_program

import "context"

// MsgSecCheckReq is the request body for the text content security check API.
type MsgSecCheckReq struct {
	Content   string `json:"content"`
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	OpenID    string `json:"openid"`
	Title     string `json:"title,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// MsgSecCheckResp is the response from the text content security check API.
type MsgSecCheckResp struct {
	TraceID string           `json:"trace_id"`
	Result  SecCheckResult   `json:"result"`
	Detail  []SecCheckDetail `json:"detail"`
}

// SecCheckResult holds the overall suggestion and label for a security check.
type SecCheckResult struct {
	Suggest string `json:"suggest"`
	Label   int    `json:"label"`
}

// SecCheckDetail holds per-strategy detail for a security check result.
type SecCheckDetail struct {
	Strategy string `json:"strategy"`
	ErrCode  int    `json:"errcode"`
	Suggest  string `json:"suggest"`
	Label    int    `json:"label"`
	Prob     int    `json:"prob"`
	Keyword  string `json:"keyword"`
}

// MediaCheckAsyncReq is the request body for the async media content security check API.
type MediaCheckAsyncReq struct {
	MediaURL  string `json:"media_url"`
	MediaType int    `json:"media_type"`
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	OpenID    string `json:"openid"`
}

// MediaCheckAsyncResp is the response from the async media content security check API.
type MediaCheckAsyncResp struct {
	TraceID string `json:"trace_id"`
}

// MsgSecCheck performs a synchronous text content security check.
func (c *Client) MsgSecCheck(ctx context.Context, req *MsgSecCheckReq) (*MsgSecCheckResp, error) {
	var resp MsgSecCheckResp
	if err := c.doPost(ctx, "/wxa/msg_sec_check", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MediaCheckAsync submits an image or audio file for asynchronous content security check.
// Results are delivered via the event callback notification.
func (c *Client) MediaCheckAsync(ctx context.Context, req *MediaCheckAsyncReq) (*MediaCheckAsyncResp, error) {
	var resp MediaCheckAsyncResp
	if err := c.doPost(ctx, "/wxa/media_check_async", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
