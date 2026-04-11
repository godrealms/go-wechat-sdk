package oplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// WxaAdminClient 代小程序开发管理客户端。
//
// 所有方法都以 AuthorizerClient.AccessToken() 作为 token 源，
// 通过共享的 doPost / doGet / doGetRaw 辅助统一处理 errcode 折叠和
// 两次 JSON 解码（先提取 errcode，再反序列化到 typed struct）。
//
// WxaAdminClient 本身无状态，线程安全；可在多 goroutine 共享。
type WxaAdminClient struct {
	auth *AuthorizerClient
}

// WxaAdmin 从 AuthorizerClient 构造开发管理客户端。
// 构造不做 I/O。
func (a *AuthorizerClient) WxaAdmin() *WxaAdminClient {
	return &WxaAdminClient{auth: a}
}

// doPost 通用 POST 辅助。
func (w *WxaAdminClient) doPost(ctx context.Context, path string, body, out any) error {
	ctx = touchContext(ctx)
	token, err := w.auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	fullPath := path + "?access_token=" + url.QueryEscape(token)

	var raw json.RawMessage
	if err := w.auth.c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}

// doGet 通用 GET 辅助：access_token 合并到 caller 传入的 query values。
func (w *WxaAdminClient) doGet(ctx context.Context, path string, q url.Values, out any) error {
	ctx = touchContext(ctx)
	token, err := w.auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	if q == nil {
		q = url.Values{}
	}
	q.Set("access_token", token)

	var raw json.RawMessage
	if err := w.auth.c.http.Get(ctx, path, q, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}

// doGetRaw 用于响应体是二进制而非 JSON 的接口（目前只有 GetQrcode）。
func (w *WxaAdminClient) doGetRaw(ctx context.Context, path string, q url.Values) ([]byte, string, error) {
	ctx = touchContext(ctx)
	token, err := w.auth.AccessToken(ctx)
	if err != nil {
		return nil, "", err
	}
	if q == nil {
		q = url.Values{}
	}
	q.Set("access_token", token)

	_, header, body, err := w.auth.c.http.DoRequestWithRawResponse(
		ctx, http.MethodGet, path, q, nil, nil,
	)
	if err != nil {
		return nil, "", fmt.Errorf("oplatform: %s: %w", path, err)
	}
	// 微信在出错时仍可能返回 JSON (errcode!=0)，检测一下
	if len(body) > 0 && body[0] == '{' {
		var base struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if json.Unmarshal(body, &base) == nil && base.ErrCode != 0 {
			return nil, "", &WeixinError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg}
		}
	}
	return body, header.Get("Content-Type"), nil
}

// decodeRaw 折叠 errcode 检查 + typed out 反序列化。
func decodeRaw(path string, raw json.RawMessage, out any) error {
	var base struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.Unmarshal(raw, &base)
	if err := checkWeixinErr(base.ErrCode, base.ErrMsg); err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("oplatform: %s decode: %w", path, err)
		}
	}
	return nil
}
