package mini_program

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// doGet 发送 GET 请求，自动注入 access_token，始终检查 errcode。
//
// 与 c.http.Get 直接调用的区别：那个调用会把响应直接 json.Unmarshal 到 out，
// 而 WeChat 的错误响应是 `{"errcode":N,"errmsg":"..."}` —— 这种结构对几乎所有
// 业务 out 类型都"成功 unmarshal"成零值，于是错误被静默吞掉。本函数走
// DoRequestWithRawResponse + 自己反序列化，强制走一次 errcode 检查。
func (c *Client) doGet(ctx context.Context, path string, extra url.Values, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}
	_, _, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, http.MethodGet, path, q, nil, nil,
	)
	if err != nil {
		return err
	}
	return decodeEnvelope(path, respBody, out)
}

// doPost 发送 POST JSON，自动注入 access_token，始终检查 errcode。
//
// 出现非零 errcode 时返回 *APIError；当响应不是合法 JSON envelope 时返回
// 带原因的错误，避免"静默 unmarshal"把损坏响应当成功。
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
		return fmt.Errorf("mini-program: %s: marshal request: %w", path, err)
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
	return utils.DecodeEnvelope("mini-program", path, respBody, out, func(code int, msg, p string) error {
		return &APIError{ErrCode: code, ErrMsg: msg, Path: p}
	})
}

// doPostRaw 发送 POST JSON，返回原始字节（用于二进制响应如图片）。
//
// 微信对二进制接口（QR 码、图片）返回二进制 body；如果服务端报错，会改回
// JSON envelope `{"errcode":N,...}`。我们用首字节判别 JSON 而非 unmarshal
// 整个 body，避免把恰好以 '{' 起头的二进制数据当成 JSON 解析失败。
//
// 与 doPost 的非对称性：doPost 要求整个响应都是合法 JSON，否则 fail loud；
// doPostRaw 则必须容忍非 JSON 响应（就是二进制成功返回）。代价是：如果代理
// 返回一段以 '{' 起头但不完整的 JSON（例如截断到 `{"errc`），我们会把它当
// 成二进制原样返回，调用方直到尝试解码图片才会发现异常。这是"二进制/JSON
// 同端点"设计的固有代价，不是 bug。
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
		var base utils.BaseResp
		if json.Unmarshal(respBody, &base) == nil && base.ErrCode != 0 {
			return nil, &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
		}
	}
	return respBody, nil
}
