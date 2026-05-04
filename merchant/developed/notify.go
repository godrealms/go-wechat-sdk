package pay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// MaxNotifyBodySize caps the bytes read from a notify callback request body.
// WeChat Pay notify payloads are well under 100 KiB in practice; the 1 MiB cap
// gives ~10× headroom while preventing memory exhaustion from a misconfigured
// or hostile sender. Callers needing a different limit can read and validate
// the body themselves before invoking the lower-level decryptNotifyResource.
const MaxNotifyBodySize int64 = 1 << 20

// ParseNotification decodes and verifies an inbound WeChat Pay callback notification from r. It decrypts the resource field using AES-256-GCM with the configured APIv3 key.
//
// The request body is read with a MaxNotifyBodySize cap; a body that would
// exceed the cap is rejected with "notify body exceeds N bytes" rather than
// being silently truncated.
func (c *Client) ParseNotification(ctx context.Context, r *http.Request, result any) (*types.Notify, error) {
	if r == nil {
		return nil, fmt.Errorf("pay: nil *http.Request")
	}
	if r.Body == nil {
		return nil, fmt.Errorf("pay: nil request body")
	}
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxNotifyBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("pay: read notify body: %w", err)
	}
	if int64(len(body)) > MaxNotifyBodySize {
		return nil, fmt.Errorf("pay: notify body exceeds %d bytes", MaxNotifyBodySize)
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

// decryptNotifyResource decrypts notify.Resource using AEAD_AES_256_GCM and
// returns the plaintext.
func decryptNotifyResource(res *types.Resource, apiV3Key string) ([]byte, error) {
	if res == nil {
		return nil, fmt.Errorf("resource is nil")
	}
	if res.Algorithm != "" && res.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("unsupported algorithm: %s", res.Algorithm)
	}
	return decryptAES256GCM(apiV3Key, res.Nonce, res.AssociatedData, res.Ciphertext)
}

// AckNotification writes a success ACK response body required by WeChat Pay.
// WeChat Pay expects HTTP 2xx with body {"code":"SUCCESS","message":"成功"}.
func AckNotification(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"code":"SUCCESS","message":"成功"}`))
}

// FailNotification writes a failure ACK response to WeChat Pay with the given message.
func FailNotification(w http.ResponseWriter, message string) {
	if message == "" {
		message = "FAILED"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	body, _ := json.Marshal(map[string]string{"code": "FAIL", "message": message})
	_, _ = w.Write(body)
}

// ParseRefundNotify parses a WeChat Pay refund callback notification. It
// reuses ParseNotification's signature verification and decryption, then
// deserializes the plaintext into *types.RefundResp.
//
// Typical usage:
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
