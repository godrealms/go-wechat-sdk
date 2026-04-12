package mini_program

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// baseResp 微信公共错误字段。
type baseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// doGet 发送 GET 请求，自动注入 access_token。
func (c *Client) doGet(ctx context.Context, path string, extra url.Values, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}
	return c.http.Get(ctx, path, q, out)
}

// doPost 发送 POST JSON，自动注入 access_token，检查 errcode。
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	fullPath := path + "?" + q.Encode()
	if out != nil {
		return c.http.Post(ctx, fullPath, body, out)
	}
	var resp baseResp
	if err := c.http.Post(ctx, fullPath, body, &resp); err != nil {
		return err
	}
	if resp.ErrCode != 0 {
		return fmt.Errorf("mini_program: %s errcode=%d errmsg=%s", path, resp.ErrCode, resp.ErrMsg)
	}
	return nil
}

// doPostRaw 发送 POST JSON，返回原始字节（用于二进制响应如图片）。
func (c *Client) doPostRaw(ctx context.Context, path string, body any) ([]byte, error) {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"access_token": {tok}}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	_, _, respBody, err := c.http.DoRequestWithRawResponse(ctx, http.MethodPost, path, q, raw, nil)
	if err != nil {
		return nil, err
	}
	if len(respBody) > 0 && respBody[0] == '{' {
		var resp baseResp
		if json.Unmarshal(respBody, &resp) == nil && resp.ErrCode != 0 {
			return nil, fmt.Errorf("mini_program: %s errcode=%d errmsg=%s", path, resp.ErrCode, resp.ErrMsg)
		}
	}
	return respBody, nil
}
