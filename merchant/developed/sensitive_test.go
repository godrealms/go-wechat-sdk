package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptSensitiveField_RoundTrip(t *testing.T) {
	priv, cert := newTestKeyAndCert(t)

	ciphertext, err := EncryptSensitiveField(cert, "张三")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("empty ciphertext")
	}

	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	// 用对应的私钥解密，验证 SDK 走的确实是微信支付 v3 规定的 RSA-OAEP(SHA-1)。
	// 必须用 sha1 解密——若这里改回 sha256 能通过，恰恰说明加密端用错了算法。
	plain, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, raw, nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plain) != "张三" {
		t.Errorf("round trip mismatch: %q", string(plain))
	}
}

// TestEncryptSensitiveField_UsesSHA1 pins the OAEP hash to SHA-1, the algorithm
// WeChat Pay v3 mandates (RSA/ECB/OAEPWithSHA-1AndMGF1Padding). A SHA-256
// round-trip would "pass" against itself while producing ciphertext WeChat's
// servers cannot decrypt, so beyond the SHA-1 round trip we also assert that
// SHA-256 decryption FAILS — a guard against silently regressing to the wrong
// hash.
func TestEncryptSensitiveField_UsesSHA1(t *testing.T) {
	priv, cert := newTestKeyAndCert(t)

	ciphertext, err := EncryptSensitiveField(cert, "李四")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}

	// SHA-1 (the mandated algorithm) must decrypt cleanly.
	if _, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, raw, nil); err != nil {
		t.Fatalf("SHA-1 OAEP decrypt failed — encryption is not using SHA-1: %v", err)
	}
	// SHA-256 must NOT decrypt it; if it does, the encrypt side regressed to the
	// wrong hash and WeChat's servers would reject the field.
	if _, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, raw, nil); err == nil {
		t.Fatal("SHA-256 OAEP decrypted the ciphertext — encryption regressed to SHA-256")
	}
}

func TestEncryptSensitiveField_NilCert(t *testing.T) {
	if _, err := EncryptSensitiveField(nil, "x"); err == nil {
		t.Fatal("expected error for nil cert")
	}
}

func TestEncryptSensitiveField_NonRSACert(t *testing.T) {
	// 伪造一个 PublicKey 非 *rsa.PublicKey 的证书：直接清空原证书的 PublicKey 字段
	_, cert := newTestKeyAndCert(t)
	cert.PublicKey = "not-rsa"
	_, err := EncryptSensitiveField(cert, "x")
	if err == nil || !strings.Contains(err.Error(), "not RSA") {
		t.Fatalf("expected non-RSA error, got %v", err)
	}
}

func TestClient_PlatformCertForEncrypt_FromCache(t *testing.T) {
	client, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	// newClientWithFakeServer 已经调用 AddPlatformCertificate，缓存里有一张。
	cert, serial, err := client.PlatformCertForEncrypt(context.Background())
	if err != nil {
		t.Fatalf("PlatformCertForEncrypt failed: %v", err)
	}
	if cert == nil {
		t.Fatal("cert is nil")
	}
	if serial == "" {
		t.Fatal("serial is empty")
	}
}
