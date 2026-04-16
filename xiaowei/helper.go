package xiaowei

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

// doPost 发送 POST JSON，自动注入 access_token，始终检查 errcode。
//
// 出现非零 errcode 时返回 *APIError；当响应不是合法 JSON envelope 时
// 返回带原因的错误，避免"静默 unmarshal"把损坏响应当成功。
//
// 实现注：刻意不走 c.http.Post，而是用 DoRequestWithRawResponse + 自己
// 反序列化，让所有 JSON 解码错误都在本文件统一格式化。
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}

	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("xiaowei: %s: marshal request: %w", path, err)
	}
	_, _, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, http.MethodPost, path, q, raw, nil,
	)
	if err != nil {
		return err
	}
	return decodeEnvelope(path, respBody, out)
}

// decodeEnvelope is the shared error-aware unmarshal step.
func decodeEnvelope(path string, respBody []byte, out any) error {
	var base baseResp
	if err := json.Unmarshal(respBody, &base); err != nil {
		return fmt.Errorf("xiaowei: %s: decode envelope: %w (body snippet: %s)",
			path, err, snippet(respBody))
	}
	if base.ErrCode != 0 {
		return &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
	}
	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("xiaowei: %s: decode result: %w", path, err)
		}
	}
	return nil
}

// snippet returns at most the first 200 bytes of body as a string, for use
// in error messages.
func snippet(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}
