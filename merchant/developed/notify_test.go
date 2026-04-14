package pay

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// encryptAES256GCM 反向生成一段 ciphertext（测试用）。
func encryptAES256GCM(t *testing.T, key, nonce, associatedData, plaintext string) string {
	t.Helper()
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		t.Fatal(err)
	}
	ct := gcm.Seal(nil, []byte(nonce), []byte(plaintext), []byte(associatedData))
	return base64.StdEncoding.EncodeToString(ct)
}

// helper: build a Client with one platform cert already trusted.
func buildClient(t *testing.T) (*Client, *rsa.PrivateKey) {
	t.Helper()
	key, cert := newTestKeyAndCert(t)
	c, err := NewClient(Config{
		Appid:             "wxtest",
		Mchid:             "1900000001",
		CertificateNumber: "TESTSERIAL",
		APIv3Key:          "01234567890123456789012345678901", // 32 bytes
		PrivateKey:        key,
		Certificate:       cert,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.AddPlatformCertificate(cert)
	return c, key
}

func TestParseNotification_Success(t *testing.T) {
	c, privKey := buildClient(t)
	apiV3 := "01234567890123456789012345678901"

	// 构造一个典型 Transaction plaintext
	plaintextJSON := `{"transaction_id":"4200001234","out_trade_no":"ord-1","mchid":"1900000001","trade_state":"SUCCESS","amount":{"total":100}}`
	nonce := "1234567890ab"
	aad := "transaction"
	ciphertext := encryptAES256GCM(t, apiV3, nonce, aad, plaintextJSON)

	notifyBody, _ := json.Marshal(map[string]any{
		"id":         "EV-001",
		"event_type": "TRANSACTION.SUCCESS",
		"resource": map[string]any{
			"algorithm":       "AEAD_AES_256_GCM",
			"ciphertext":      ciphertext,
			"associated_data": aad,
			"nonce":           nonce,
			"original_type":   "transaction",
		},
	})

	// 用 privKey 签名响应头 (与平台证书同一 key)
	// 使用当前时间以满足 ±5 分钟重放窗口 (audit C5).
	ts := fmt.Sprintf("%d", time.Now().Unix())
	respNonce := "abcdefghij"
	source := fmt.Sprintf("%s\n%s\n%s\n", ts, respNonce, string(notifyBody))
	sigB64, err := utils.SignSHA256WithRSA(source, privKey)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewReader(notifyBody))
	req.Header.Set("Wechatpay-Timestamp", ts)
	req.Header.Set("Wechatpay-Nonce", respNonce)
	req.Header.Set("Wechatpay-Signature", sigB64)
	// Serial 必须匹配缓存里的平台证书
	req.Header.Set("Wechatpay-Serial", utils.GetCertificateSerialNumber(*c.CertificateVal()))

	type txn struct {
		TransactionId string `json:"transaction_id"`
		OutTradeNo    string `json:"out_trade_no"`
		TradeState    string `json:"trade_state"`
	}
	var parsed txn
	notify, err := c.ParseNotification(context.Background(), req, &parsed)
	if err != nil {
		t.Fatal(err)
	}
	if notify.Id != "EV-001" {
		t.Errorf("notify.Id: %s", notify.Id)
	}
	if parsed.TransactionId != "4200001234" || parsed.OutTradeNo != "ord-1" || parsed.TradeState != "SUCCESS" {
		t.Errorf("decrypted = %+v", parsed)
	}
}

func TestParseNotification_BadSignature(t *testing.T) {
	c, _ := buildClient(t)
	body := []byte(`{"id":"x","event_type":"TRANSACTION.SUCCESS"}`)
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewReader(body))
	req.Header.Set("Wechatpay-Timestamp", "1700000000")
	req.Header.Set("Wechatpay-Nonce", "abc")
	req.Header.Set("Wechatpay-Signature", base64.StdEncoding.EncodeToString([]byte("wrong")))
	req.Header.Set("Wechatpay-Serial", utils.GetCertificateSerialNumber(*c.CertificateVal()))
	if _, err := c.ParseNotification(context.Background(), req, nil); err == nil {
		t.Error("expected signature error")
	}
}

func TestAckNotification(t *testing.T) {
	w := httptest.NewRecorder()
	AckNotification(w)
	if w.Code != 200 {
		t.Errorf("code = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "SUCCESS") {
		t.Errorf("body = %s", w.Body.String())
	}
}
