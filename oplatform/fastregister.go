package oplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// FastRegisterClient 提供开放平台代注册小程序相关的 component 级接口。
//
// 所有方法都以 Client.ComponentAccessToken() 作为 token 源（而不是 authorizer
// 级别的 token），因为快速注册流程发生在 authorizer 关系建立之前。
//
// FastRegisterClient 无状态，线程安全，可在多 goroutine 共享。
type FastRegisterClient struct {
	c *Client
}

// FastRegister 从 Client 构造 FastRegisterClient。构造不做 I/O。
func (c *Client) FastRegister() *FastRegisterClient {
	return &FastRegisterClient{c: c}
}

// doPost 通用 POST 辅助：
//   - 从 Client 取 component_access_token 并以 query 参数形式拼接
//   - 正确处理 path 中已经带有 ?action=xxx 的情况（使用 & 分隔而非 ?）
//   - 复用包级 decodeRaw 进行两段式 JSON 解码（errcode 折叠 + typed unmarshal）
func (f *FastRegisterClient) doPost(ctx context.Context, path string, body, out any) error {
	ctx = touchContext(ctx)
	token, err := f.c.ComponentAccessToken(ctx)
	if err != nil {
		return err
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	fullPath := path + sep + "component_access_token=" + url.QueryEscape(token)

	var raw json.RawMessage
	if err := f.c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}
