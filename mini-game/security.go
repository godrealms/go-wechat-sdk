package mini_game

import "context"

// MsgSecCheckReq holds the parameters for submitting text content to the security check API.
type MsgSecCheckReq struct {
	Content string `json:"content"`
	Version int    `json:"version"`
	Scene   int    `json:"scene"`
	OpenID  string `json:"openid"`
}

// SecCheckResult contains the suggestion and risk label returned by the security check.
type SecCheckResult struct {
	Suggest string `json:"suggest"`
	Label   int    `json:"label"`
}

// MsgSecCheckResp is the response returned by MsgSecCheck.
type MsgSecCheckResp struct {
	Result SecCheckResult `json:"result"`
}

// MsgSecCheck submits text content to WeChat's security check API and returns the moderation result.
func (c *Client) MsgSecCheck(ctx context.Context, req *MsgSecCheckReq) (*MsgSecCheckResp, error) {
	var resp MsgSecCheckResp
	if err := c.doPost(ctx, "/wxa/msg_sec_check", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
