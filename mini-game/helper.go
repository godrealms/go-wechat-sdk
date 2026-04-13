package mini_game

import (
	"context"
	"encoding/json"
	"net/url"
)

// baseResp 微信公共错误字段。
type baseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// doPost 发送 POST JSON，自动注入 access_token，始终检查 errcode。
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	fullPath := path + "?" + q.Encode()

	// 始终先解码到 json.RawMessage，确保 errcode 检查不被跳过。
	var raw json.RawMessage
	if err := c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return err
	}
	var base baseResp
	_ = json.Unmarshal(raw, &base)
	if base.ErrCode != 0 {
		return &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}
