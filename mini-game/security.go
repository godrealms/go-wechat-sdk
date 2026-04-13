package mini_game

import "context"

type MsgSecCheckReq struct {
	Content string `json:"content"`
	Version int    `json:"version"`
	Scene   int    `json:"scene"`
	OpenID  string `json:"openid"`
}
type SecCheckResult struct {
	Suggest string `json:"suggest"`
	Label   int    `json:"label"`
}
type MsgSecCheckResp struct {
	Result SecCheckResult `json:"result"`
}

func (c *Client) MsgSecCheck(ctx context.Context, req *MsgSecCheckReq) (*MsgSecCheckResp, error) {
	var resp MsgSecCheckResp
	if err := c.doPost(ctx, "/wxa/msg_sec_check", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
