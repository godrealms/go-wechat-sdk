package pay

import (
	"context"
	"encoding/json"
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

	headers := http.Header{
		"Accept":        []string{"application/json"},
		"Authorization": []string{auth},
		"User-Agent":    []string{"go-wechat-sdk/1.x"},
	}
	if raw != nil {
		headers.Set("Content-Type", "application/json")
	}

	statusCode, respHeader, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, method, urlPath, query, raw, headers,
	)
	if err != nil {
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
