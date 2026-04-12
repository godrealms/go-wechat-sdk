package pay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// ParseNotification 解析微信支付回调请求。
// 它会：
//  1. 从 HTTP header 取出 Wechatpay-* 签名字段；
//  2. 用本地缓存的平台证书验签（必要时自动拉取 /v3/certificates）；
//  3. 解析请求体为 *types.Notify；
//  4. 使用 APIv3 密钥解密 resource 字段，得到明文；
//  5. 如果 result != nil，就把明文 unmarshal 进去。
//
// 返回的第一个值总是原始 *types.Notify（即使 result 为 nil）。
//
// 典型用法：
//
//	var txn types.Transaction
//	notify, err := client.ParseNotification(ctx, r, &txn)
//	if err != nil { ... }
//	// notify.EventType / notify.Id ...
//	// txn 已经是解密后的交易结构
func (c *Client) ParseNotification(ctx context.Context, r *http.Request, result any) (*types.Notify, error) {
	if r == nil {
		return nil, fmt.Errorf("pay: nil *http.Request")
	}
	if r.Body == nil {
		return nil, fmt.Errorf("pay: nil request body")
	}
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("pay: read notify body: %w", err)
	}

	if err := c.verifyResponseSignature(ctx, r.Header, body); err != nil {
		return nil, fmt.Errorf("pay: notify signature invalid: %w", err)
	}

	notify := &types.Notify{}
	if err := json.Unmarshal(body, notify); err != nil {
		return nil, fmt.Errorf("pay: unmarshal notify body: %w", err)
	}

	if notify.Resource == nil {
		return notify, nil
	}

	plaintext, err := decryptNotifyResource(notify.Resource, c.apiV3Key)
	if err != nil {
		return nil, fmt.Errorf("pay: decrypt notify resource: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(plaintext, result); err != nil {
			return nil, fmt.Errorf("pay: unmarshal notify resource plaintext: %w", err)
		}
	}
	return notify, nil
}

// decryptNotifyResource 对 notify.Resource 按 AEAD_AES_256_GCM 解密，返回明文。
func decryptNotifyResource(res *types.Resource, apiV3Key string) ([]byte, error) {
	if res == nil {
		return nil, fmt.Errorf("resource is nil")
	}
	if res.Algorithm != "" && res.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("unsupported algorithm: %s", res.Algorithm)
	}
	return decryptAES256GCM(apiV3Key, res.Nonce, res.AssociatedData, res.Ciphertext)
}

// AckNotification 向微信支付回复成功 ACK 响应体。
// 微信支付要求 HTTP 2xx 且响应体为 {"code":"SUCCESS","message":"成功"}。
func AckNotification(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"code":"SUCCESS","message":"成功"}`))
}

// FailNotification 向微信支付回复失败 ACK。
func FailNotification(w http.ResponseWriter, message string) {
	if message == "" {
		message = "FAILED"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	body, _ := json.Marshal(map[string]string{"code": "FAIL", "message": message})
	_, _ = w.Write(body)
}

// ParseRefundNotify 解析退款回调通知。
// 它复用 ParseNotification 的签名验证与解密逻辑，将解密后的明文反序列化为 *types.RefundResp。
//
// 典型用法：
//
//	notify, refund, err := client.ParseRefundNotify(ctx, r)
//	if err != nil { ... }
//	// refund.RefundId, refund.Status ...
func (c *Client) ParseRefundNotify(ctx context.Context, r *http.Request) (*types.Notify, *types.RefundResp, error) {
	var refund types.RefundResp
	notify, err := c.ParseNotification(ctx, r, &refund)
	if err != nil {
		return nil, nil, err
	}
	return notify, &refund, nil
}
