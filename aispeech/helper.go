package aispeech

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// baseResp holds the common WeChat error fields present in every API response.
type baseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// doPost sends a POST JSON request to path with access_token in the query,
// always checks errcode before unmarshalling into out.
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
		return fmt.Errorf("aispeech: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}
