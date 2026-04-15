package pay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// doV3 是所有微信支付 V3 接口的统一调用入口。
//
// 该方法保证：
//  1. 请求体只 marshal 一次，签名 body 与发送 body 完全一致；
//  2. Authorization 头通过 per-request headers 注入，不污染共享 client；
//  3. 在解析 result 之前，会通过 verifyResponseSignature 验证响应签名；
//  4. 把 ctx 一路透传，调用方可通过 ctx 控制超时/取消。
//
// 当 body 为 nil（如 GET）时，签名串中的 body 段为空字符串，与微信文档一致。
func (c *Client) doV3(
	ctx context.Context,
	method, urlPath string,
	query url.Values,
	body any,
	result any,
) error {
	return c.doV3WithHeaders(ctx, method, urlPath, query, body, nil, result)
}

// doV3WithHeaders 与 doV3 相同，但允许调用方追加额外请求头（例如
// 敏感信息加密接口所需的 Wechatpay-Serial）。SDK 自己管理的 Accept /
// Authorization / User-Agent / Content-Type 仍然由内部生成并覆盖调用方；
// 其余头部会被合并写入。
func (c *Client) doV3WithHeaders(
	ctx context.Context,
	method, urlPath string,
	query url.Values,
	body any,
	extraHeaders http.Header,
	result any,
) error {
	if err := c.validateForRequest(); err != nil {
		return err
	}

	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
	}

	// 签名时使用的 path 必须包含 query string（与实际请求 URL 一致）
	signPath := urlPath
	if len(query) > 0 {
		signPath = signPath + "?" + query.Encode()
	}

	nonce := utils.GenerateHashBasedString(32)
	ts := time.Now().Unix()

	auth, err := c.authorizationHeader(method, signPath, string(raw), nonce, ts)
	if err != nil {
		return err
	}

	headers := http.Header{}
	// 先拷贝调用方自定义头，避免被后续的 SDK 管理头覆盖（对调用方头保持
	// "次优先级" 的语义）。
	for k, vs := range extraHeaders {
		for _, v := range vs {
			headers.Add(k, v)
		}
	}
	// SDK 管理的头最后写，保证调用方无法覆盖签名等关键头。
	headers.Set("Accept", "application/json")
	headers.Set("Authorization", auth)
	headers.Set("User-Agent", "go-wechat-sdk/1.x")
	if raw != nil {
		headers.Set("Content-Type", "application/json")
	}

	statusCode, respHeader, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, method, urlPath, query, raw, headers,
	)
	if err != nil {
		var httpErr *utils.HTTPError
		if errors.As(err, &httpErr) && len(httpErr.Body) > 0 {
			var env struct {
				Code    string          `json:"code"`
				Message string          `json:"message"`
				Detail  json.RawMessage `json:"detail,omitempty"`
			}
			if jerr := json.Unmarshal([]byte(httpErr.Body), &env); jerr == nil && env.Code != "" {
				return &V3Error{
					HTTPStatus: httpErr.StatusCode,
					Code:       env.Code,
					Message:    env.Message,
					Detail:     env.Detail,
					Path:       urlPath,
				}
			}
		}
		return err
	}

	// 部分接口（关闭订单等）成功时返回 204 No Content + 空 body，
	// 这种情况无须验签也无须反序列化。
	if statusCode == http.StatusNoContent || len(respBody) == 0 {
		return nil
	}

	if err := c.verifyResponseSignature(ctx, respHeader, respBody); err != nil {
		return fmt.Errorf("verify response signature failed: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response failed: %w: %s", err, string(respBody))
		}
	}
	return nil
}

// getV3 / postV3 是 doV3 的便捷封装。
func (c *Client) getV3(ctx context.Context, urlPath string, query url.Values, result any) error {
	return c.doV3(ctx, http.MethodGet, urlPath, query, nil, result)
}

func (c *Client) postV3(ctx context.Context, urlPath string, body any, result any) error {
	return c.doV3(ctx, http.MethodPost, urlPath, nil, body, result)
}

// PostV3Raw 暴露给其他包（比如 merchant/service）以便复用核心签名/验签逻辑。
// 对普通业务代码来说推荐用更具体的方法（TransactionsJsapi 等）。
func (c *Client) PostV3Raw(ctx context.Context, urlPath string, body any, result any) error {
	return c.postV3(ctx, urlPath, body, result)
}

// GetV3Raw 暴露给其他包以便复用核心签名/验签逻辑。
func (c *Client) GetV3Raw(ctx context.Context, urlPath string, query url.Values, result any) error {
	return c.getV3(ctx, urlPath, query, result)
}

// DoV3 是一个更灵活的转发入口，允许调用方自定义 HTTP 方法、query、
// 请求体和额外请求头。签名、验签、平台证书拉取等逻辑仍由 SDK 处理。
//
// 典型使用场景：
//   - 需要 PUT/PATCH 等 PostV3Raw/GetV3Raw 未覆盖的方法；
//   - 需要通过 Wechatpay-Serial 头告知服务端敏感字段使用的平台证书序列号；
//   - 需要自定义 Idempotency-Key 等业务头。
func (c *Client) DoV3(
	ctx context.Context,
	method, urlPath string,
	query url.Values,
	body any,
	extraHeaders http.Header,
	result any,
) error {
	return c.doV3WithHeaders(ctx, method, urlPath, query, body, extraHeaders, result)
}
