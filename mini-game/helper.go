package mini_game

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/godrealms/go-wechat-sdk/utils"
)

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
		return fmt.Errorf("mini-game: %s: marshal request: %w", path, err)
	}
	_, _, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, http.MethodPost, path, q, raw, nil,
	)
	if err != nil {
		return err
	}
	return decodeEnvelope(path, respBody, out)
}

// decodeEnvelope delegates to the shared utils.DecodeEnvelope, producing a
// package-local *APIError on non-zero errcodes.
func decodeEnvelope(path string, respBody []byte, out any) error {
	return utils.DecodeEnvelope("mini-game", path, respBody, out, func(code int, msg, p string) error {
		return &APIError{ErrCode: code, ErrMsg: msg, Path: p}
	})
}
