package types_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// generateTestRSAKey generates a 2048-bit RSA key pair for testing.
func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return key
}

// encodePublicKeyPEM encodes an RSA public key to PEM format.
func encodePublicKeyPEM(t *testing.T, pub *rsa.PublicKey) string {
	t.Helper()
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}
	return string(pem.EncodeToMemory(block))
}

// encryptAES256GCM encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext.
func encryptAES256GCM(t *testing.T, key, nonce, associatedData, plaintext []byte) string {
	t.Helper()
	c, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create AES cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		t.Fatalf("failed to create GCM: %v", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, associatedData)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// --- TransactionsJsapi.GenerateSignature ---

func TestTransactionsJsapi_GenerateSignature_Success(t *testing.T) {
	key := generateTestRSAKey(t)
	j := &types.TransactionsJsapi{
		TimeStamp: "1609459200",
		NonceStr:  "abc123",
		Package:   "prepay_id=test_prepay_id",
		SignType:   "RSA",
	}
	err := j.GenerateSignature("wx_app_id", key)
	if err != nil {
		t.Fatalf("GenerateSignature() returned error: %v", err)
	}
	if j.PaySign == "" {
		t.Error("expected PaySign to be set after GenerateSignature")
	}
}

func TestTransactionsJsapi_GenerateSignature_NilKey(t *testing.T) {
	j := &types.TransactionsJsapi{
		TimeStamp: "1609459200",
		NonceStr:  "abc123",
		Package:   "prepay_id=test_prepay_id",
	}
	err := j.GenerateSignature("wx_app_id", nil)
	if err == nil {
		t.Error("expected error when private key is nil")
	}
}

// --- ModifyAppResponse.GenerateSignature ---

func TestModifyAppResponse_GenerateSignature_Success(t *testing.T) {
	key := generateTestRSAKey(t)
	r := &types.ModifyAppResponse{
		AppId:        "wx_app_id",
		PrepayId:     "test_prepay_id",
		PackageValue: "Sign=WXPay",
		NonceStr:     "abc123",
		TimeStamp:    "1609459200",
	}
	err := r.GenerateSignature(key)
	if err != nil {
		t.Fatalf("GenerateSignature() returned error: %v", err)
	}
	if r.Sign == "" {
		t.Error("expected Sign to be set after GenerateSignature")
	}
}

func TestModifyAppResponse_GenerateSignature_NilKey(t *testing.T) {
	r := &types.ModifyAppResponse{
		AppId:     "wx_app_id",
		PrepayId:  "test_prepay_id",
		NonceStr:  "abc123",
		TimeStamp: "1609459200",
	}
	err := r.GenerateSignature(nil)
	if err == nil {
		t.Error("expected error when private key is nil")
	}
}

// --- Notify.VerifySignature ---

func TestNotify_VerifySignature_InvalidPublicKey(t *testing.T) {
	n := &types.Notify{}
	// Pass an invalid PEM block — should return false (block == nil path)
	ok := n.VerifySignature("ts", "nonce", "sig", "body", "not-a-pem")
	if ok {
		t.Error("expected VerifySignature to return false for invalid PEM")
	}
}

func TestNotify_VerifySignature_InvalidBase64Signature(t *testing.T) {
	key := generateTestRSAKey(t)
	pubPEM := encodePublicKeyPEM(t, &key.PublicKey)
	n := &types.Notify{}
	// Valid PEM, invalid base64 signature
	ok := n.VerifySignature("ts", "nonce", "!!!not-base64!!!", "body", pubPEM)
	if ok {
		t.Error("expected VerifySignature to return false for invalid base64 signature")
	}
}

// --- Notify.DecryptAES256GCM and DecryptResource ---

func TestNotify_DecryptAES256GCM_Success(t *testing.T) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		t.Fatalf("failed to generate AES key: %v", err)
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("failed to generate nonce: %v", err)
	}
	associatedData := []byte("test-associated-data")
	plaintext := []byte("hello world")

	ciphertext := encryptAES256GCM(t, aesKey, nonce, associatedData, plaintext)

	n := &types.Notify{
		Resource: &types.Resource{
			Ciphertext:     ciphertext,
			Nonce:          string(nonce),
			AssociatedData: string(associatedData),
		},
	}
	got, err := n.DecryptAES256GCM(string(aesKey))
	if err != nil {
		t.Fatalf("DecryptAES256GCM() returned error: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Errorf("DecryptAES256GCM() = %q, want %q", got, plaintext)
	}
}

func TestNotify_DecryptAES256GCM_InvalidCiphertext(t *testing.T) {
	aesKey := make([]byte, 32)
	rand.Read(aesKey) //nolint:errcheck

	n := &types.Notify{
		Resource: &types.Resource{
			Ciphertext:     "!!!invalid-base64!!!",
			Nonce:          "123456789012",
			AssociatedData: "",
		},
	}
	_, err := n.DecryptAES256GCM(string(aesKey))
	if err == nil {
		t.Error("expected error for invalid base64 ciphertext")
	}
}

func TestNotify_DecryptAES256GCM_InvalidKeyLength(t *testing.T) {
	// AES requires key of 16, 24, or 32 bytes; use a 5-byte key
	n := &types.Notify{
		Resource: &types.Resource{
			Ciphertext:     base64.StdEncoding.EncodeToString([]byte("data")),
			Nonce:          "123456789012",
			AssociatedData: "",
		},
	}
	_, err := n.DecryptAES256GCM("short")
	if err == nil {
		t.Error("expected error for invalid AES key length")
	}
}

func TestNotify_DecryptResource_Success(t *testing.T) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		t.Fatalf("failed to generate AES key: %v", err)
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("failed to generate nonce: %v", err)
	}
	associatedData := []byte("test")
	// Valid Transaction JSON as plaintext
	plaintext := []byte(`{"transaction_id":"tx001","mchid":"1234567890","out_trade_no":"order001","trade_state":"SUCCESS"}`)

	ciphertext := encryptAES256GCM(t, aesKey, nonce, associatedData, plaintext)

	n := &types.Notify{
		Resource: &types.Resource{
			Ciphertext:     ciphertext,
			Nonce:          string(nonce),
			AssociatedData: string(associatedData),
		},
	}
	tx, err := n.DecryptResource(string(aesKey))
	if err != nil {
		t.Fatalf("DecryptResource() returned error: %v", err)
	}
	if tx.TransactionId != "tx001" {
		t.Errorf("expected TransactionId=tx001, got %q", tx.TransactionId)
	}
}

func TestNotify_DecryptResource_InvalidJSON(t *testing.T) {
	aesKey := make([]byte, 32)
	rand.Read(aesKey) //nolint:errcheck
	nonce := make([]byte, 12)
	rand.Read(nonce) //nolint:errcheck

	// Plaintext is not valid JSON
	plaintext := []byte("not-json")
	ciphertext := encryptAES256GCM(t, aesKey, nonce, nil, plaintext)

	n := &types.Notify{
		Resource: &types.Resource{
			Ciphertext:     ciphertext,
			Nonce:          string(nonce),
			AssociatedData: "",
		},
	}
	_, err := n.DecryptResource(string(aesKey))
	if err == nil {
		t.Error("expected error when decrypted data is not valid JSON")
	}
}
