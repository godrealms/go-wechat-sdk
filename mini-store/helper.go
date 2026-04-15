package mini_store

import (
	"context"
	"encoding/json"
	"net/url"
)

type baseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// doPost sends a POST JSON request with access_token, checks errcode,
// and unmarshals the response into out.
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	fullPath := path + "?" + q.Encode()
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
